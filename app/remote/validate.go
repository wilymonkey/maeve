package remote

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
)

func GetMissing(nodeDir string, metas []*local.FileMeta) ([]*local.FileMeta, error) {
	if len(metas) == 0 {
		err := fmt.Errorf("no metas received")
		return nil, help.WrapError(err, "asserting meta array length")
	}

	root, _, err := local.SplitAtRootPath(metas[0].RelPath)
	if err != nil {
		return nil, err
	}
	sourceDir := filepath.Join(nodeDir, root)
	exists, err := local.PathExists(sourceDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return metas, nil
	}

	metaChan := make(chan *local.FileMeta, 100)
	var walkErr utils.ThreadSafe[error]
	onError := func(err error) {
		walkErr.Set(err)
	}

	local.WalkDirForMetas(
		sourceDir,
		context.Background(),
		metaChan,
		nil,
		onError,
	)

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
			return nil, help.WrapError(err, "removing partial/incorrect file")
		}
	}

	missing := make([]*local.FileMeta, 0, len(existsMap))
	for _, m := range existsMap {
		missing = append(missing, m)
	}

	if walkErr.Get() != nil {
		return nil, walkErr.Get()
	}

	return missing, nil
}
