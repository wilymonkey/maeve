package remote

import (
	"fmt"
	"log"
	"net"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/proto"
	"google.golang.org/grpc"
)

type commServer struct {
	proto.UnimplementedNodeServiceServer
}

func newServer() *commServer {
	return &commServer{}
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
	proto.RegisterNodeServiceServer(grpcServer, newServer())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
