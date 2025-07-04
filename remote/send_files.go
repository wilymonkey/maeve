package remote

import (
	"context"
	"io"
	"time"

	hs "github.com/wilymonkey/maeve/local/hashsums"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

type SendStatus struct {
	hash       hs.FileHash
	curr       int64
	total      int64
	isVerified bool
}

func SendFiles(
	hashes []hs.FileHash,
	client *ssh.Client,
	ctx context.Context,
	progChan chan SendStatus,
) {
	eGrp, ctx := errgroup.WithContext(ctx)

	downChan := make(chan hs.FileHash, 2)
	defer close(downChan)
	i := 0
	verifyChan := make(chan SendStatus, 1)
	defer close(verifyChan)

	eGrp.Go(func() error {
		for hash := range downChan {
			status := SendStatus{hash: hash}
			// TODO: Send the file.
			verifyChan <- status
		}
		return nil
	})

	eGrp.Go(func() error {
		for status := range verifyChan {
			// TODO: Ask the remote if the file is correct.
			progChan <- status
			if len(hashes) < i {
				downChan <- hashes[i]
				i++
			}
		}
		return nil
	})

	// Init the downloads
	if len(hashes) < i {
		downChan <- hashes[i]
		i++
		if len(hashes) < i {
			downChan <- hashes[i]
			i++
		}
	}
}

type ProgressWriter struct {
	ID           int
	Writer       io.Writer
	Total        int64
	Transferred  int64
	Percent      int
	LastReported time.Time
}

func NewProgressWriter(id int, total int64, writer io.Writer) ProgressWriter {
	return ProgressWriter{ID: id, Writer: writer, Total: total}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.Transferred += int64(n)

	now := time.Now()
	if now.Sub(pw.LastReported) > time.Second || pw.Transferred == pw.Total {
		pw.Percent = int(float64(pw.Transferred) / float64(pw.Total) * 100)
		pw.LastReported = now
	}

	return n, err
}
