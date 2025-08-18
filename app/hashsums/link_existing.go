package hashsums

import (
	"context"
	"errors"
	"os"
	"runtime"

	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
)

func LinkExisting(node string, hashes []OldFileHash) ([]OldFileHash, error) {
	mh, err := getMaster(node)
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	missingChan := make(chan OldFileHash, 100)
	var missing []OldFileHash
	wg := utils.GoWait(func() {
		for hash := range missingChan {
			missing = append(missing, hash)
		}
	})

	eGrp, ctx := errgroup.WithContext(context.Background())
	eGrp.SetLimit(20 * runtime.NumCPU())

	for _, h := range hashes {
		eGrp.Go(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if relPath, exists := mh.HashMap[h.Hash]; exists {
				sourcePath := relPath.Resolve(node)

				// TODO: Give proper temp directory.
				h.Snapshot = "latest"

				targetPath := h.AbsPath(node)
				if err := local.Hardlink(sourcePath, targetPath); err != nil {
					if errors.Is(err, os.ErrNotExist) {
						missingChan <- h
						return nil
					}
					return utils.WrapErr(err)
				}
			} else {
				missingChan <- h
			}
			return nil
		})
	}
	err = eGrp.Wait()
	close(missingChan)
	wg.Wait()
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	return missing, nil
}
