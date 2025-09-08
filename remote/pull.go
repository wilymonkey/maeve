package remote

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/proto"
	"github.com/wilymonkey/maeve/utils"
	"google.golang.org/grpc"
)

func (c *Comms) PullDB(ctx context.Context) error {
	progChan := make(chan *PushStatus, 10)
	go func() {
		for range progChan {
		}
	}()

	dbPath := db.DBPath(conf.MyNode())
	if err := c.pullFile(dbPath, ctx, progChan); err != nil {
		return err
	}

	return nil
}

func (c *Comms) pullFile(
	relpath string,
	ctx context.Context,
	progChan chan<- *PushStatus,
) error {
	target := filepath.Join(conf.MyNode(), relpath)
	localFile, err := os.Open(target)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer utils.Cleanup(&err, localFile.Close)

	stream, err := c.Client.Download(
		ctx,
		&proto.DownloadRequest{
			Node:    conf.MyName(),
			RelPath: relpath,
		},
	)
	if err != nil {
		return fmt.Errorf("acquiring download stream: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("getting file metadata: %w", err)
	}

	metaResp, ok := resp.Data.(*proto.DownloadResponse_Filemeta)
	utils.Assert(ok, "response from server should be filemeta")

	meta := local.FromProto(metaResp.Filemeta)
	status := newPushStatus(meta)
	pw := newProgWriter(
		localFile,
		meta.Size,
		ctx,
		func(written int64) {
			status.SentSize = written
			progChan <- status
		},
	)
	if _, err := io.Copy(pw, newGRPCReader(stream)); err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}

	return err
}

type gRPCReader struct {
	stream grpc.ServerStreamingClient[proto.DownloadResponse]
	buf    []byte
}

func newGRPCReader(
	stream grpc.ServerStreamingClient[proto.DownloadResponse],
) *gRPCReader {
	return &gRPCReader{
		stream: stream,
		buf:    make([]byte, 0),
	}
}

func (r *gRPCReader) Read(p []byte) (int, error) {
	// If we have leftover data in the buffer, use it first
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}

	resp, err := r.stream.Recv()
	if err != nil {
		if err == io.EOF {
			return 0, io.EOF
		}
		return 0, err
	}

	chunk, ok := resp.Data.(*proto.DownloadResponse_Chunk)
	utils.Assert(ok, "response from server should be data chunk")

	n := copy(p, chunk.Chunk)
	// Save leftover for next Read call
	if n < len(chunk.Chunk) {
		r.buf = chunk.Chunk[n:]
	}

	return n, nil
}
