package remote

import (
	"io"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
	"github.com/zeebo/blake3"
)

type sshPipeConn struct {
	reader io.Reader
	writer io.WriteCloser
}

func (c *sshPipeConn) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (c *sshPipeConn) Write(b []byte) (n int, err error) {
	return c.writer.Write(b)
}

func (c *sshPipeConn) Close() error {
	return c.writer.Close()
}

func (c *sshPipeConn) LocalAddr() net.Addr                { return nil }
func (c *sshPipeConn) RemoteAddr() net.Addr               { return nil }
func (c *sshPipeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *sshPipeConn) SetWriteDeadline(t time.Time) error { return nil }
func (c *sshPipeConn) SetDeadline(t time.Time) error      { return nil }

func RunServer() error {
	rpcFuncs := new(RPCFuncs)
	if err := rpc.Register(rpcFuncs); err != nil {
		return help.WrapError(err, "register rpc functions")
	}

	conn := &sshPipeConn{reader: os.Stdin, writer: os.Stdout}
	rpc.ServeConn(conn)
	return nil
}

type RPCFuncs int

type GetDBVersionArgs struct {
	Node string
}
type GetDBVersionReply struct {
	Version *db.DBVersion
}

func (h *RPCFuncs) GetDBVersion(args *GetDBVersionArgs, reply *GetDBVersionReply) error {
	exists, err := local.PathExists(db.DBPath(args.Node))
	if err != nil {
		return err
	}
	if !exists {
		reply.Version = &db.DBVersion{}
		return nil
	}

	conn, err := db.OpenRead(args.Node)
	if err != nil {
		return err
	}
	defer utils.Cleanup(&err, conn.Close)

	reply.Version, err = db.GetVersion(conn)
	if err != nil {
		return err
	}
	return err
}

func (n *NodeConn) GetDBVersion() (*db.DBVersion, error) {
	args := &GetDBVersionArgs{Node: conf.MyName()}
	var reply GetDBVersionReply
	if err := n.rpcClient.Call("RPCFuncs.GetDBVersion", args, &reply); err != nil {
		return reply.Version, help.DecodeErr(err)
	}
	return reply.Version, nil
}

type BackupDirArgs struct {
	Node string
}
type BackupDirReply struct {
	Path string
}

func (h *RPCFuncs) BackupDir(args *BackupDirArgs, reply *BackupDirReply) error {
	reply.Path = conf.RemoteTempDir(args.Node)
	return nil
}

func (n *NodeConn) addBackupDir() error {
	args := &BackupDirArgs{Node: conf.MyName()}
	var reply BackupDirReply
	if err := n.rpcClient.Call("RPCFuncs.BackupDir", args, &reply); err != nil {
		return help.DecodeErr(err)
	}
	n.backupDir = reply.Path
	return nil
}

type VerifyFileArgs struct {
	Hash *local.FileMeta
	Node string
}
type VerifyFileReply struct {
	IsGood bool
}

func (h *RPCFuncs) VerifyFile(args *VerifyFileArgs, reply *VerifyFileReply) error {
	nodeDir := conf.RemoteTempDir(args.Node)
	path := filepath.Join(nodeDir, args.Hash.RelPath)
	isGood, err := args.Hash.Validate(path, blake3.New())
	if err != nil {
		return err
	}
	reply.IsGood = isGood
	return nil
}

func (n *NodeConn) VerifyFile(hash *local.FileMeta) (bool, error) {
	args := &VerifyFileArgs{
		Hash: hash,
		Node: conf.MyName(),
	}
	var reply VerifyFileReply
	if err := n.rpcClient.Call("RPCFuncs.VerifyFile", args, &reply); err != nil {
		return false, help.DecodeErr(err)
	}
	return reply.IsGood, nil
}

type ConformToDBArgs struct {
	Node string
}
type ConformToDBReply struct {
	MissingLinkIds []int64
}

func (h *RPCFuncs) ConformToDB(args *ConformToDBArgs, reply *ConformToDBReply) error {
	nodeDir := conf.RemoteTempDir(args.Node)
	conn, err := db.OpenRead(nodeDir)
	if err != nil {
		return err
	}
	defer utils.Cleanup(&err, conn.Close)

	metas, err := db.GetLatestMeta(conn)
	if err != nil {
		return err
	}

	missing, err := GetMissing(nodeDir, metas)
	if err != nil {
		return err
	}

	ids, err := db.IdsFromMetas(conn, missing)
	if err != nil {
		return err
	}

	reply.MissingLinkIds = ids
	return err
}

func (n *NodeConn) ConformToDB() ([]int64, error) {
	args := &ConformToDBArgs{
		Node: conf.MyName(),
	}
	var reply ConformToDBReply
	if err := n.rpcClient.Call("RPCFuncs.ConformToDB", args, &reply); err != nil {
		return nil, help.DecodeErr(err)
	}
	return reply.MissingLinkIds, nil
}
