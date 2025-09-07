package remote

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/help"
	"github.com/wilymonkey/maeve/local"
	"golang.org/x/sync/errgroup"
)

func GetMissing(nodeDir string, metas []*local.FileMeta) ([]*local.FileMeta, error) {
	if len(metas) == 0 {
		err := fmt.Errorf("no metas received")
		return nil, help.WrapErr(err, "asserting meta array length")
	}

	root, _ := local.SplitAtRootPath(metas[0].RelPath)
	sourceDir := filepath.Join(nodeDir, root)
	exists, err := local.PathExists(sourceDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return metas, nil
	}

	metaChan := make(chan *local.FileMeta, 100)
	eGrp, ctx := errgroup.WithContext(context.Background())
	eGrp.Go(func() error {
		return local.WalkDirForMetas(
			nodeDir,
			sourceDir,
			ctx,
			metaChan,
			nil,
		)
	})

	existsMap := make(map[[32]byte]*local.FileMeta, len(metas))
	for _, m := range metas {
		existsMap[m.Hash] = m
	}

	for meta := range metaChan {
		if _, exists := existsMap[meta.Hash]; exists {
			delete(existsMap, meta.Hash)
			continue
		}

		path := filepath.Join(nodeDir, meta.RelPath)
		if err := os.Remove(path); err != nil {
			return nil, help.WrapErr(err, "removing partial/incorrect file")
		}
	}

	missing := make([]*local.FileMeta, 0, len(existsMap))
	for _, m := range existsMap {
		missing = append(missing, m)
	}

	if err := eGrp.Wait(); err != nil {
		return nil, err
	}

	return missing, nil
}
