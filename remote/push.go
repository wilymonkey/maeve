package remote

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/proto"
	"github.com/wilymonkey/maeve/utils"
	"github.com/zeebo/blake3"
	"google.golang.org/grpc"
)

type PushStatus struct {
	Meta     *local.FileMeta
	SentSize int64
}

func newPushStatus(meta *local.FileMeta) *PushStatus {
	return &PushStatus{Meta: meta}
}

func (c *Comms) PushLinks(
	metas []*local.FileMeta,
	ctx context.Context,
	progChan chan<- *PushStatus,
) error {
	const maxRetries = 3
	defer close(progChan)

	if c.backupDir == "" {
		if err := c.addBackupDir(ctx); err != nil {
			return err
		}
	}

	verifiedPush := func(m *local.FileMeta) error {
		if err := c.pushFile(m, ctx, progChan); err != nil {
			return err
		}
		resp, err := c.Client.VerifyFile(
			ctx,
			&proto.VerifyFileRequest{Filemeta: m.AsProto()},
		)
		if err != nil {
			return err
		}
		if !resp.IsGood {
			return fmt.Errorf("remote file hash did not match local hash")
		}
		return nil
	}

	for _, m := range metas {
		var err error
		for range maxRetries {
			if err = verifiedPush(m); err == nil {
				break // Break on sucess.
			}
			utils.Sleep(1000)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Comms) PushDB(ctx context.Context, path string) error {
	meta, err := local.GenFileMeta(path, conf.MyNode(), blake3.New())
	if err != nil {
		return err
	}

	progChan := make(chan *PushStatus, 10)
	go func() {
		for range progChan {
		}
	}()

	if err := c.pushFile(&meta, ctx, progChan); err != nil {
		return err
	}

	return nil
}

func (c *Comms) pushFile(
	meta *local.FileMeta,
	ctx context.Context,
	progChan chan<- *PushStatus,
) error {
	utils.Assert(c.backupDir != "", "Backup Dir has already been attained")

	status := newPushStatus(meta)
	source := filepath.Join(conf.MyNode(), meta.RelPath)

	localFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer utils.Cleanup(&err, localFile.Close)

	stream, err := c.Client.Upload(ctx)
	if err != nil {
		return fmt.Errorf("acquiring client stream: %w", err)
	}

	err = stream.Send(&proto.UploadRequest{
		Data: &proto.UploadRequest_Filemeta{
			Filemeta: meta.AsProto(),
		},
	})
	if err != nil {
		return fmt.Errorf("sending metadata: %w", err)
	}

	pw := newProgWriter(
		&gRPCWriter{stream: stream},
		meta.Size,
		ctx,
		func(written int64) {
			status.SentSize = written
			progChan <- status
		},
	)
	if _, err := io.Copy(pw, localFile); err != nil {
		return fmt.Errorf("pushing file: %w", err)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("closing file upload: %w", err)
	}

	return err
}

type progWriter struct {
	writer     io.Writer
	total      int64
	written    int64
	onProgress func(written int64)
	ctx        context.Context
	rate       int64
	burstSize  int64
	tokens     int64
	ticker     *time.Ticker
}

func newProgWriter(
	writer io.Writer,
	total int64,
	ctx context.Context,
	onProgress func(written int64),

) *progWriter {
	rateKB := conf.GetConf().MaxUpload * 1024

	pw := &progWriter{
		writer:     writer,
		total:      total,
		ctx:        ctx,
		onProgress: onProgress,
		rate:       rateKB,
		burstSize:  rateKB / 2,
		tokens:     rateKB / 2, // start half full
	}

	pw.ticker = time.NewTicker(100 * time.Millisecond)
	go pw.refillTokens()
	return pw
}

func (pw *progWriter) Write(p []byte) (int, error) {
	if err := pw.ctx.Err(); err != nil {
		return 0, err
	}

	// No rate limit.
	if pw.rate < 1 {
		n, err := pw.writer.Write(p)
		pw.written += int64(n)
		pw.onProgress(pw.written)
		return n, err
	}

	var writtenNow int
	for len(p) > 0 {

		// wait a tiny bit if no tokens
		if pw.tokens < 1 {
			select {
			case <-pw.ctx.Done():
				return writtenNow, pw.ctx.Err()
			case <-time.After(10 * time.Millisecond):
				continue
			}
		}

		toWrite := min(int64(len(p)), pw.tokens)
		n, err := pw.writer.Write(p[:toWrite])
		if err != nil {
			return int(pw.written), fmt.Errorf("send byte array: %w", err)
		}
		p = p[n:]
		writtenNow += int(n)

		bytes := int64(n)
		pw.tokens -= bytes
		pw.written += bytes
		pw.onProgress(pw.written)
	}
	return writtenNow, nil
}

func (pw *progWriter) refillTokens() {
	for {
		select {
		case <-pw.ctx.Done():
			pw.ticker.Stop()
			return
		case <-pw.ticker.C:
			pw.tokens += pw.rate / 10 // 100ms interval
			if pw.tokens > pw.burstSize {
				pw.tokens = pw.burstSize
			}
		}
	}
}

type gRPCWriter struct {
	stream grpc.ClientStreamingClient[proto.UploadRequest, proto.Empty]
}

func (gw *gRPCWriter) Write(p []byte) (int, error) {
	err := gw.stream.Send(&proto.UploadRequest{
		Data: &proto.UploadRequest_Chunk{Chunk: p},
	})
	return len(p), err
}
