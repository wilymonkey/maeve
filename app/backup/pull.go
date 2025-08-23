package backup

import (
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
)

// =======================================
// STATES
// =======================================

type pullState struct {
	path   string
	number binding.Int
	size   binding.Item[int64]
}

func newPullState() []*pullState {
	cfg := conf.GetConf()
	states := make([]*pullState, 0, len(cfg.SourceDirs))
	for _, dir := range cfg.SourceDirs {
		states = append(states, &pullState{
			path:   dir,
			number: binding.NewInt(),
			size:   fynext.BindNewInt64(),
		})
	}
	return states
}

// =======================================
// UI
// =======================================

func pullStateTable() fyne.CanvasObject {
	table := widget.NewTable(
		func() (int, int) {
			return len(guiState.pullStates) + 1, 3
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")

			return fynext.LabelDisableUntil(
				label,
				guiState.currTask,
				pulling,
			)
		},
		func(cell widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if cell.Row == 0 {
				switch cell.Col {
				case 0:
					label.Unbind()
					label.SetText("Folder")
					label.Alignment = fyne.TextAlignLeading
				case 1:
					label.SetText("Files")
				case 2:
					label.SetText("Size")
				}
			} else {
				pState := guiState.pullStates[cell.Row-1]
				switch cell.Col {
				case 0:
					label.Unbind()
					txt := utils.TruncateString(pState.path, 50)
					label.SetText(txt)
					label.Alignment = fyne.TextAlignLeading

				case 1:
					label.Bind(binding.IntToString(pState.number))

				case 2:
					sizeStr := binding.NewString()
					label.Bind(sizeStr)
					pState.size.AddListener(binding.NewDataListener(func() {
						size := fynext.Unwrap(pState.size)
						sizeStr.Set(utils.BytesToHuman(size))
					}))

				}
			}
		},
	)
	table.SetColumnWidth(0, 250)
	table.SetColumnWidth(1, 90)
	table.SetColumnWidth(2, 90)
	table.SetColumnWidth(3, 90)
	return table
}

// =======================================
// METHODS
// =======================================

func newLatest() ([]*local.FileMeta, error) {
	if err := removeChildDirs(conf.MyNode()); err != nil {
		return nil, err
	}
	latest, err := newLatestDir()
	if err != nil {
		return nil, err
	}

	var fileMetas []*local.FileMeta

	for _, pState := range guiState.pullStates {
		metaChan := make(chan *local.FileMeta, 100)
		var totalSize int64
		var totalFiles int

		utils.Throttle(
			metaChan,
			func(meta *local.FileMeta) {
				totalFiles++
				totalSize += meta.Size
				fileMetas = append(fileMetas, meta)
			},
			func() {
				fyne.Do(func() {
					pState.number.Set(totalFiles)
					pState.size.Set(totalSize)
				})
			},
		)

		sourceDir := pState.path
		targetDir := filepath.Join(latest, filepath.Base(pState.path))

		if err := os.RemoveAll(targetDir); err != nil {
			return nil, help.CheckBackupDir(err, "deleting folder to link things to")
		}

		err := local.WalkDirForMetas(
			sourceDir,
			guiState.ctx,
			metaChan,
			func(path string) (string, error) {
				relPath, err := filepath.Rel(sourceDir, path)
				if err != nil {
					return "", help.DevReport(err, "creating relative path")
				}
				targetPath := filepath.Join(targetDir, relPath)
				if err := local.Hardlink(path, targetPath); err != nil {
					return "", err
				}
				return targetPath, nil
			},
		)
		if err != nil {
			return nil, err
		}
	}
	return fileMetas, nil
}

func newLatestDir() (string, error) {
	currentTime := time.Now().Format(conf.TimeFormat)
	path := filepath.Join(conf.MyNode(), currentTime)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", help.CheckBackupDir(err, "creating latest folder")
	}
	return path, nil
}

func removeChildDirs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return help.CheckBackupDir(err, "reading files in backup folder")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err := os.RemoveAll(filepath.Join(dir, entry.Name()))
			if err != nil {
				return help.CheckBackupDir(err, "deleting stale folders in backup")
			}
		}
	}
	return nil
}
