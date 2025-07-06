package remote

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	hs "github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/rpc"
	"github.com/wilymonkey/maeve/utils/myerr"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

type SendStatus struct {
	Hash      hs.FileHash
	Curr      int64
	Total     int64
	Verifying bool
	IsGood    bool
}

func SendFiles(
	hashes []hs.FileHash,
	sshClient *ssh.Client,
	ctx context.Context,
	progChan chan SendStatus,
) error {
	eGrp, ctx := errgroup.WithContext(ctx)

	rpcSesh, err := sshClient.NewSession()
	if err != nil {
		return myerr.WrapErr(err)
	}
	defer rpcSesh.Close()
	rpcClient, err := rpc.New(rpcSesh)
	if err != nil {
		return myerr.WrapErr(err)
	}
	defer rpcClient.Close()
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return myerr.WrapErr(err)
	}
	defer sftpClient.Close()

	remoteDir, err := rpc.SelfNodeTempLoc(rpcClient)
	if err != nil {
		return myerr.WrapErr(err)
	}

	downChan := make(chan hs.FileHash, 2)
	i := 0
	verifyChan := make(chan SendStatus, 1)
	doneChan := make(chan struct{}, 1)
	go func() {
		<-doneChan
		close(downChan)
		close(verifyChan)
	}()
	defer close(doneChan)

	eGrp.Go(func() error {
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
					return myerr.WrapErr(err)
				}
				verifyChan <- status
			}
		}
	})

	eGrp.Go(func() error {
		// Init the downloads
		if i < len(hashes) {
			downChan <- hashes[i]
			i++
			if i < len(hashes) {
				return myerr.DummyErr()
				downChan <- hashes[i]
				i++
			}
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
					return myerr.WrapErr(err)
				}
				status.IsGood = isGood
				progChan <- status

				if !status.IsGood {
					downChan <- status.Hash
				} else if i < len(hashes) {
					downChan <- hashes[i]
					i++
				} else {
					doneChan <- struct{}{}
				}
			}
		}
	})

	if err := eGrp.Wait(); err != nil {
		doneChan <- struct{}{}
		return myerr.WrapErr(err)
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
		return status, myerr.WrapErr(err)
	}
	defer localFile.Close()
	fileInfo, err := localFile.Stat()
	if err != nil {
		return status, myerr.WrapErr(err)
	}
	status.Total = fileInfo.Size()

	remotePath := hash.Path.ResolvePrepend(remoteDir)
	parentDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(parentDir); err != nil {
		return status, myerr.WrapErr(err)
	}

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return status, myerr.WrapErr(err)
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
		return status, myerr.WrapErr(err)
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
