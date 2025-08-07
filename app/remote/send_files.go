package remote

import (
	"context"
	"io"
	netRPC "net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"github.com/wilymonkey/maeve/conf"
	hs "github.com/wilymonkey/maeve/hashsums"
	"github.com/wilymonkey/maeve/rpc"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
)

type SendStatus struct {
	Hash      hs.FileHash
	Curr      int64
	Total     int64
	Verifying bool
	IsGood    bool
}

func NewDoneStatus(hash hs.FileHash) (SendStatus, error) {
	path := hash.AbsPath(conf.GetConf().MyNode())
	localFile, err := os.Open(path)
	if err != nil {
		return SendStatus{}, utils.WrapErr(err)
	}
	defer localFile.Close()
	fileInfo, err := localFile.Stat()
	if err != nil {
		return SendStatus{}, utils.WrapErr(err)
	}
	size := fileInfo.Size()
	return SendStatus{
		Hash:      hash,
		Curr:      size,
		Total:     size,
		Verifying: true,
		IsGood:    true,
	}, nil
}

func SendFiles(
	hashes []hs.FileHash,
	sftpClient *sftp.Client,
	rpcClient *netRPC.Client,
	ctx context.Context,
	progChan chan SendStatus,
) error {
	eGrp, ctx := errgroup.WithContext(ctx)

	remoteDir, err := rpc.TempLocation(rpcClient)
	if err != nil {
		return utils.WrapErr(err)
	}

	downChan := make(chan hs.FileHash, 2)
	verifyChan := make(chan SendStatus, 1)

	eGrp.Go(func() error {
		defer close(verifyChan)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case hash, ok := <-downChan:
				if !ok {
					return nil
				}
				status, err := pushFile(sftpClient, hash, remoteDir, ctx, progChan)
				if err != nil {
					return utils.WrapErr(err)
				}
				verifyChan <- status
			}
		}
	})

	eGrp.Go(func() error {
		defer close(downChan)

		i := 0
		if i < len(hashes) {
			downChan <- hashes[i]
			i++
		}
		if i < len(hashes) {
			downChan <- hashes[i]
			i++
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case status, ok := <-verifyChan:
				if !ok {
					return nil
				}

				status.Verifying = true
				progChan <- status

				isGood, err := rpc.VerifyFile(rpcClient, status.Hash)
				if err != nil {
					return utils.WrapErr(err)
				}
				status.IsGood = isGood
				progChan <- status

				switch {
				case !status.IsGood:
					downChan <- status.Hash
				case i < len(hashes):
					downChan <- hashes[i]
					i++
				default:
					return nil
				}
			}
		}
	})

	if err := eGrp.Wait(); err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func pushFile(
	sftpClient *sftp.Client,
	hash hs.FileHash,
	remoteDir string,
	ctx context.Context,
	progChan chan SendStatus,
) (SendStatus, error) {
	status := SendStatus{Hash: hash}

	path := hash.AbsPath(conf.GetConf().MyNode())
	localFile, err := os.Open(path)
	if err != nil {
		return status, utils.WrapErr(err)
	}
	defer localFile.Close()
	fileInfo, err := localFile.Stat()
	if err != nil {
		return status, utils.WrapErr(err)
	}
	status.Total = fileInfo.Size()

	remotePath := filepath.Join(remoteDir, hash.RelPath)
	parentDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(parentDir); err != nil {
		return status, utils.WrapErr(err)
	}

	remoteFile, err := sftpClient.Create(remotePath)
	defer remoteFile.Close()
	if err != nil {
		return status, utils.WrapErrWithInfo(err, remotePath)
	}
	rateKB := conf.GetConf().MaxUpload * 1024
	pw := &progWriter{
		writer: remoteFile,
		total:  status.Total,
		ctx:    ctx,
		progress: func(written int64) {
			status.Curr = written
			progChan <- status
		},
		rate:       rateKB,
		burstSize:  rateKB / 2,
		tokens:     rateKB,
		lastRefill: time.Now(),
	}

	if _, err := io.Copy(pw, localFile); err != nil {
		return status, utils.WrapErr(err)
	}
	return status, nil
}

type progWriter struct {
	writer     io.Writer
	total      int64
	written    int64
	progress   func(written int64)
	ctx        context.Context
	rate       int64
	burstSize  int64
	tokens     int64
	lastRefill time.Time
}

func (pw *progWriter) Write(p []byte) (int, error) {
	if err := pw.ctx.Err(); err != nil {
		return 0, utils.WrapErr(err)
	}
	if pw.rate < 1 {
		n, err := pw.writer.Write(p)
		pw.written += int64(n)
		pw.progress(pw.written)
		return n, err
	}

	var writtenNow int
	for len(p) > 0 {
		pw.refillTokens()
		toWrite := min(int64(len(p)), pw.tokens)
		if toWrite > 0 {
			n, err := pw.writer.Write(p[:toWrite])
			if err != nil {
				return int(pw.written), utils.WrapErr(err)
			}
			p = p[n:]
			writtenNow += n

			bytes := int64(n)
			pw.tokens -= bytes
			pw.written += bytes
			pw.progress(pw.written)
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
