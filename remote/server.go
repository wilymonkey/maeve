package remote

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/proto"
	"github.com/wilymonkey/maeve/utils"
	"github.com/zeebo/blake3"
	"google.golang.org/grpc"
)

type commServer struct {
	proto.UnimplementedCommsServer
}

func RunServer() {
	cfg := conf.GetConf()
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.ServerPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	proto.RegisterCommsServer(grpcServer, &commServer{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}

func (c *commServer) GetDBVersion(
	ctx context.Context,
	req *proto.GetDBVersionRequest,
) (*proto.GetDBVersionResponse, error) {
	exists, err := local.PathExists(db.DBPath(req.Node))
	if err != nil {
		return nil, err
	}
	if !exists {
		return &proto.GetDBVersionResponse{}, nil
	}

	conn, err := db.OpenRead(req.Node)
	if err != nil {
		return nil, err
	}
	defer utils.Cleanup(&err, conn.Close)

	version, err := db.GetVersion(conn)
	if err != nil {
		return nil, err
	}
	return &proto.GetDBVersionResponse{
		Version: version,
	}, nil
}

func (c *commServer) BackupDir(
	ctx context.Context,
	req *proto.BackupDirRequest,
) (*proto.BackupDirResponse, error) {
	return &proto.BackupDirResponse{
		Path: conf.RemoteTempDir(req.Node),
	}, nil
}

func (c *commServer) VerifyFile(
	ctx context.Context,
	req *proto.VerifyFileRequest,
) (*proto.VerifyFileResponse, error) {
	nodeDir := conf.RemoteTempDir(req.Node)
	path := filepath.Join(nodeDir, req.Filemeta.RelPath)
	isGood, err := local.Validate(req.Filemeta, path, blake3.New())
	if err != nil {
		return nil, err
	}
	return &proto.VerifyFileResponse{IsGood: isGood}, nil
}

func (c *commServer) ConformToDB(
	ctx context.Context,
	req *proto.ConformToDBRequest,
) (*proto.ConformToDBResponse, error) {
	nodeDir := conf.RemoteTempDir(req.Node)
	conn, err := db.OpenRead(nodeDir)
	if err != nil {
		return nil, err
	}
	defer utils.Cleanup(&err, conn.Close)

	metas, err := db.GetLatestMeta(conn)
	if err != nil {
		return nil, err
	}

	missing, err := GetMissing(nodeDir, metas)
	if err != nil {
		return nil, err
	}

	ids, err := db.IdsFromMetas(conn, missing)
	if err != nil {
		return nil, err
	}

	return &proto.ConformToDBResponse{
		MissingLinkIds: ids,
	}, err
}

func (c *commServer) Finalise(
	ctx context.Context,
	req *proto.FinaliseRequest,
) (*proto.Empty, error) {
	tempDir := conf.RemoteTempDir(req.Node)
	dbSource := db.DBPath(tempDir)
	nodeDir := filepath.Dir(tempDir)
	snapSource := filepath.Join(tempDir, req.Snapshot)

	utils.Assertf(local.PathOk(snapSource), "path %q should exist", snapSource)
	utils.Assertf(local.PathOk(dbSource), "path %q should exist", dbSource)

	snapTarget := filepath.Join(nodeDir, filepath.Base(snapSource))
	if err := os.Rename(snapSource, snapTarget); err != nil {
		return nil, fmt.Errorf("moving dir: %v", err)
	}

	dbTarget := filepath.Join(nodeDir, filepath.Base(dbSource))
	if err := os.Rename(dbSource, dbTarget); err != nil {
		return nil, fmt.Errorf("moving database: %v", err)
	}

	if err := os.RemoveAll(tempDir); err != nil {
		return nil, fmt.Errorf("deleting temp dir: %v", err)
	}

	return &proto.Empty{}, nil
}
