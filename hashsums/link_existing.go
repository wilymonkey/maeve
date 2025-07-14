package hashsums

import (
	"context"
	"runtime"

	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
)

func LinkExisting(node string, hashes []FileHash) ([]FileHash, error) {
	mh, err := getMaster(node)
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	missingChan := make(chan FileHash, 100)
	defer close(missingChan)
	var missing []FileHash
	go func() {
		for hash := range missingChan {
			missing = append(missing, hash)
		}
	}()

	eGrp, ctx := errgroup.WithContext(context.Background())
	eGrp.SetLimit(20 * runtime.NumCPU())

	for _, h := range hashes {
		eGrp.Go(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if relPath, exists := mh.HashMap[h.Hash]; exists {
				sourcePath := relPath.Resolve(node)
				targetPath := h.Path.ResolveTemp(node)
				if err := local.Hardlink(sourcePath, targetPath); err != nil {
					return utils.WrapErr(err)
				}
			} else {
				missingChan <- h
			}
			return nil
		})
	}

	if err := eGrp.Wait(); err != nil {
		return nil, utils.WrapErr(err)
	}

	return missing, nil
}
