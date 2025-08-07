package backup

import (
	"context"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/data/binding"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/utils"
)

type BackupDirMeta struct {
	path     string
	number   binding.Int
	size     binding.Int
	hashsums binding.Float
}
type doneLinks struct{}

func updateLatest(meta []BackupDirMeta, ctx context.Context) error {
	selfDir := conf.GetConf().SelfDir()

	if err := os.RemoveAll(selfDir); err != nil {
		return utils.WrapErr(err)
	}

	if err := os.MkdirAll(selfDir, os.ModeDir); err != nil {
		return utils.WrapErr(err)
	}

	for _, m := range meta {
		destDir := filepath.Join(selfDir, filepath.Base(m.path))

		metaChan := make(chan int64, 100)
		var totalSize int64
		var totalFiles int
		utils.Throttle(
			metaChan,
			func(size int64) {
				totalFiles++
				totalSize += size
			},
			func() {
				m.number.Set(totalFiles)
				m.size.Set(int(totalSize))
			})

		if err := local.HardlinkDir(m.path, destDir, metaChan, ctx); err != nil {
			return utils.WrapErr(err)
		}
	}

	return nil
}
