package remote

import (
	"context"
	"io"
	netRPC "net/rpc"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
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
	localFile, err := os.Open(hash.Path.ResolveSelf())
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

	localFile, err := os.Open(hash.Path.ResolveSelf())
	if err != nil {
		return status, utils.WrapErr(err)
	}
	defer localFile.Close()
	fileInfo, err := localFile.Stat()
	if err != nil {
		return status, utils.WrapErr(err)
	}
	status.Total = fileInfo.Size()

	remotePath := hash.Path.ResolvePrepend(remoteDir)
	parentDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(parentDir); err != nil {
		return status, utils.WrapErr(err)
	}

	remoteFile, err := sftpClient.Create(remotePath)
	defer remoteFile.Close()
	if err != nil {
		return status, utils.WrapErrWithInfo(err, remotePath)
	}
	pw := &ProgressWriter{
		Writer: remoteFile,
		Total:  status.Total,
		ctx:    ctx,
		Progress: func(written int64) {
			status.Curr = written
			progChan <- status
		},
	}

	if _, err := io.Copy(pw, localFile); err != nil {
		return status, utils.WrapErr(err)
	}
	return status, nil
}

type ProgressWriter struct {
	Writer   io.Writer
	Total    int64
	Written  int64
	Progress func(written int64)
	ctx      context.Context
}

// Write implements the io.Writer interface.
// Assumes the Progress field is not nil.
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	select {
	case <-pw.ctx.Done():
		return 0, pw.ctx.Err()
	default:
		// Continue
	}

	n, err = pw.Writer.Write(p)
	pw.Written += int64(n)
	pw.Progress(pw.Written)
	return n, err
}
