package remote

import (
	"context"
	"fmt"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/proto"
	"google.golang.org/grpc"
)

type Comms struct {
	conn      *grpc.ClientConn
	Client    proto.CommsClient
	backupDir string
}

func NewComms(addr string) (*Comms, error) {
	var opts []grpc.DialOption
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("create grpc client: %w", err)
	}
	return &Comms{
		conn:   conn,
		Client: proto.NewCommsClient(conn),
	}, nil
}

func (c *Comms) Close() {
	c.Close()
}

func (c *Comms) addBackupDir(ctx context.Context) error {
	resp, err := c.Client.BackupDir(
		ctx,
		&proto.BackupDirRequest{Node: conf.MyName()},
	)
	if err != nil {
		return err
	}
	c.backupDir = resp.Path
	return nil
}
