package remote

import (
	"context"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
)

func GetMissing(nodeDir string, metas []*local.FileMeta) ([]*local.FileMeta, error) {
	utils.Assert("At least 1 path in metas", len(metas) > 0)

	root, _, err := local.SplitAtRootPath(metas[0].RelPath)
	if err != nil {
		return nil, err
	}

	var missing []*local.FileMeta
	ctx, cancelCtx := context.WithCancelCause(context.Background())
	metaChan := make(chan *local.FileMeta, 100)
	go func() {
		defer cancelCtx(nil)

		existsMap := make(map[[32]byte]*local.FileMeta, len(metas))
		for _, m := range metas {
			existsMap[m.Hash] = m
		}

		for meta := range metaChan {
			if _, exists := existsMap[meta.Hash]; exists {
				delete(existsMap, meta.Hash)
			} else {
				path := filepath.Join(nodeDir, meta.RelPath)
				if err := os.Remove(path); err != nil {
					err = help.CheckBackupDir(err, "removing partial/incorrect file")
					cancelCtx(err)
					return
				}
			}
		}

		missing = make([]*local.FileMeta, 0, len(existsMap))
		for _, m := range existsMap {
			missing = append(missing, m)
		}
	}()

	sourceDir := filepath.Join(nodeDir, root)
	err = local.WalkDirForMetas(
		sourceDir,
		ctx,
		metaChan, // Writer (closes).
		func(path string) (string, error) {
			return path, nil
		},
	)
	if err != nil {
		return nil, err
	}

	<-ctx.Done()
	return missing, ctx.Err()
}
