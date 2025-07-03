package backup

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/utils"
	"github.com/wilymonkey/maeve/utils/myerr"
)

type linkPathMeta struct {
	path   string
	number int64
	size   int64
}
type doneBackup struct{}

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
func pullChanges(linkDirs map[string]linkPathMeta) tea.Msg {
	selfDir := cfg.Global.SelfDir()

	if err := os.RemoveAll(selfDir); err != nil {
		return myerr.TuiMsg(err)
	}

	if err := os.MkdirAll(selfDir, 0755); err != nil {
		return myerr.TuiMsg(err)
	}

	sizeChan := make(chan int64, 100)
	defer close(sizeChan)
	var totalSize int64
	var totalFiles int64
	for _, srcDir := range cfg.Global.SourceDirs {
		destDir := filepath.Join(selfDir, filepath.Base(srcDir))
		utils.Throttle(
			sizeChan,
			func(size int64) {
				totalFiles++
				totalSize += size

			},
			func() {
				cfg.TuiProgram.Send(func() tea.Msg {
					return linkPathMeta{
						path:   srcDir,
						number: totalFiles,
						size:   totalSize,
					}
				})
			})

		if err := local.HardlinkDir(srcDir, destDir, sizeChan); err != nil {
			return myerr.TuiMsg(err)
		}
	}

	return doneBackup{}
}
