package remote

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
	"github.com/zeebo/blake3"
)

type PushStatus struct {
	Meta     *local.FileMeta
	SentSize int64
}

func newPushStatus(meta *local.FileMeta) *PushStatus {
	return &PushStatus{Meta: meta}
}

func (n *NodeConn) PushLinks(
	metas []*local.FileMeta,
	ctx context.Context,
	progChan chan<- *PushStatus,
) error {
	const maxRetries = 3
	defer close(progChan)

	if n.backupDir == "" {
		if err := n.addBackupDir(); err != nil {
			return err
		}
	}

	if n.sftpClient == nil {
		if err := n.addSFTP(); err != nil {
			return err
		}
	}

	verifiedPush := func(m *local.FileMeta) error {
		if err := n.pushFile(m, ctx, progChan); err != nil {
			return err
		}
		isGood, err := n.VerifyFile(m)
		if err != nil {
			return err
		}
		if !isGood {
			err = fmt.Errorf("remote file hash did not match local hash")
			return help.WrapErr(err, "verifying sent file")
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
			n.addSFTP() // Renew.
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeConn) PushDB(path string, ctx context.Context) error {
	if n.backupDir == "" {
		if err := n.addBackupDir(); err != nil {
			return err
		}
	}

	if n.sftpClient == nil {
		if err := n.addSFTP(); err != nil {
			return err
		}
	}

	meta, err := local.GenFileMeta(path, conf.MyNode(), blake3.New())
	if err != nil {
		return err
	}

	progChan := make(chan *PushStatus, 10)
	go func() {
		for range progChan {
		}
	}()

	if err := n.pushFile(&meta, ctx, progChan); err != nil {
		return err
	}

	return nil
}

func (n *NodeConn) pushFile(
	meta *local.FileMeta,
	ctx context.Context,
	progChan chan<- *PushStatus,
) error {
	utils.Assert(n.backupDir != "", "Backup Dir has already been attained")

	status := newPushStatus(meta)
	source := filepath.Join(conf.MyNode(), meta.RelPath)
	target := filepath.Join(n.backupDir, meta.RelPath)

	localFile, err := os.Open(source)
	if err != nil {
		return help.WrapErr(err, "opening file")
	}
	defer utils.Cleanup(&err, localFile.Close)

	remoteFile, err := n.sftpClient.Create(target)
	if err != nil {
		if !strings.Contains(err.Error(), "does not exist") {
			return help.WrapErr(err, "creating remote file")
		}

		parentDir := filepath.Dir(target)
		if err = n.sftpClient.MkdirAll(parentDir); err != nil {
			return help.WrapErr(err, "creating parent dir for remote file")
		}
		remoteFile, err = n.sftpClient.Create(target)
		if err != nil {
			return help.WrapErr(err, "creating remote file AGAIN")
		}
	}
	defer utils.Cleanup(&err, remoteFile.Close)

	pw := newProgWriter(
		remoteFile, meta.Size, ctx,
		func(written int64) {
			status.SentSize = written
			progChan <- status
		},
	)
	if _, err := io.Copy(pw, localFile); err != nil {
		return help.WrapErr(err, "pushing file")
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
	lastRefill time.Time
}

func newProgWriter(
	writer io.Writer,
	total int64,
	ctx context.Context,
	onProgress func(written int64),

) *progWriter {
	rateKB := conf.GetConf().MaxUpload * 1024

	return &progWriter{
		writer:     writer,
		total:      total,
		ctx:        ctx,
		onProgress: onProgress,
		rate:       rateKB,
		burstSize:  rateKB / 2,
		tokens:     rateKB,
		lastRefill: time.Now(), // TODO: Make the time a ticker instead of manually refilling.
	}
}

func (pw *progWriter) Write(p []byte) (int, error) {
	if err := pw.ctx.Err(); err != nil {
		return 0, err
	}
	if pw.rate < 1 {
		n, err := pw.writer.Write(p)
		pw.written += int64(n)
		pw.onProgress(pw.written)
		return n, err
	}

	var writtenNow int
	for len(p) > 0 {
		pw.refillTokens()
		toWrite := min(int64(len(p)), pw.tokens)
		if toWrite > 0 {
			n, err := pw.writer.Write(p[:toWrite])
			if err != nil {
				return int(pw.written), help.WrapErr(err, "send byte array")
			}
			p = p[n:]
			writtenNow += n

			bytes := int64(n)
			pw.tokens -= bytes
			pw.written += bytes
			pw.onProgress(pw.written)
		}

		if len(p) > 0 {
			utils.Sleep(10) // Allow tokens to refill
		}
	}
	return writtenNow, nil
}

func (pw *progWriter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(pw.lastRefill)
	newTokens := int64(float64(pw.rate) * elapsed.Seconds())
	if newTokens > 0 {
		pw.tokens += newTokens
		if pw.tokens > pw.burstSize {
			pw.tokens = pw.burstSize
		}
		pw.lastRefill = now
	}
}
