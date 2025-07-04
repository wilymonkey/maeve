package local

import (
	"context"
	"errors"
	"os"

	hs "github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/utils/myerr"
	"github.com/wilymonkey/maeve/utils/semaphore"
	"golang.org/x/sync/errgroup"
)

func LinkFilesFromSnapshots(node string, remoteHashes []hs.FileHash) ([]hs.FileHash, error) {
	localMaster, err := hs.GetMasterHash(node)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist) || errors.Is(err, hs.ErrInvalidMH):
			localMaster, err = hs.NewMasterHash(node)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return remoteHashes, nil
				}
				return nil, myerr.WrapErr(err)
			}
		default:
			return nil, myerr.WrapErr(err)
		}
	}

	sem := semaphore.NewScaling(20)
	defer sem.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eGrp, ctx := errgroup.WithContext(ctx)

	missingChan := make(chan hs.FileHash, 100)
	var missing []hs.FileHash
	go func() {
		for hash := range missingChan {
			missing = append(missing, hash)
		}
	}()

	eGrp.Go(func() error {
		sem.Acquire()
		defer sem.Release()
		for _, rHash := range remoteHashes {
			if relPath := localMaster.Exists(rHash); relPath != nil {
				sourcePath := relPath.Resolve(node)
				targetPath := rHash.Path.ResolveTemp(node)
				if err := hardlink(sourcePath, targetPath); err != nil {
					return err
				}
			} else {
				missingChan <- rHash
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})

	if err := eGrp.Wait(); err != nil {
		return nil, myerr.WrapErr(err)
	}
	close(missingChan)

	return missing, nil
}
