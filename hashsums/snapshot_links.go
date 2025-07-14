package hashsums

import (
	"context"
	"errors"
	"os"

	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
)

func LinkFilesFromSnapshots(node string, remoteHashes []FileHash) ([]FileHash, error) {
	localMaster, err := GetMaster(node)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist) || errors.Is(err, ErrInvalidMH):
			localMaster, err = NewMasterHash(node)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return remoteHashes, nil
				}
				return nil, utils.WrapErr(err)
			}
		default:
			return nil, utils.WrapErr(err)
		}
	}

	sem := utils.NewScalingSema(20)
	defer sem.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eGrp, ctx := errgroup.WithContext(ctx)

	missingChan := make(chan FileHash, 100)
	var missing []FileHash
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
				if err := local.Hardlink(sourcePath, targetPath); err != nil {
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
		return nil, utils.WrapErr(err)
	}
	close(missingChan)

	return missing, nil
}
