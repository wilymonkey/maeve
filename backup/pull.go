package backup

import (
	"context"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/utils"
)

type linkPathMeta struct {
	path   string
	number int64
	size   int64
}
type doneLinks struct{}

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
func pullChanges(linkDirs map[string]linkPathMeta, ctx context.Context) tea.Msg {
	selfDir := cfg.Global.SelfDir()

	if err := os.RemoveAll(selfDir); err != nil {
		return utils.WrapErr(err)
	}

	if err := os.MkdirAll(selfDir, 0755); err != nil {
		return utils.WrapErr(err)
	}

	for srcDir := range linkDirs {
		destDir := filepath.Join(selfDir, filepath.Base(srcDir))

		sizeChan := make(chan int64, 100)
		var totalSize int64
		var totalFiles int64
		utils.Throttle(
			sizeChan,
			func(size int64) {
				totalFiles++
				totalSize += size
			},
			func() {
				overseer.Global.Send(
					linkPathMeta{
						path:   srcDir,
						number: totalFiles,
						size:   totalSize,
					},
				)
			})

		if err := local.HardlinkDir(srcDir, destDir, sizeChan, ctx); err != nil {
			return utils.WrapErr(err)
		}
		close(sizeChan)
	}

	return doneLinks{}
}
