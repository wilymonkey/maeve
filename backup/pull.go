package backup

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/help"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
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
			return len(global.pullStates) + 1, 3
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")

			return fynext.LabelDisableUntil(
				label,
				global.currTask,
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
				pState := global.pullStates[cell.Row-1]
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
	myNode := conf.MyNode()
	if exists := local.PathOk(myNode); exists {
		if err := removeChildDirs(myNode); err != nil {
			return nil, err
		}
	}

	latest, err := newLatestDir()
	if err != nil {
		return nil, err
	}

	var fileMetas []*local.FileMeta

	for _, pState := range global.pullStates {
		metaChan := make(chan *local.FileMeta, 100)

		sourceDir := pState.path
		targetDir := filepath.Join(latest, filepath.Base(sourceDir))
		linkpath := func(path string) (string, error) {
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return "", help.WrapErr(err, "creating relative path")
			}
			targetPath := filepath.Join(targetDir, relPath)
			if err := local.Hardlink(path, targetPath); err != nil {
				return "", err
			}
			return targetPath, nil
		}

		if err := os.RemoveAll(targetDir); err != nil {
			return nil, help.WrapErr(err, "deleting folder to link things to")
		}

		eGrp, ctx := errgroup.WithContext(global.ctx)
		eGrp.Go(func() error {
			return local.WalkDirForMetas(
				conf.MyNode(),
				sourceDir,
				ctx,
				metaChan,
				linkpath,
			)
		})

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
				pState.number.Set(totalFiles)
				pState.size.Set(totalSize)
			},
		)
		if err := eGrp.Wait(); err != nil {
			return nil, err
		}
	}

	return fileMetas, nil
}

func newLatestDir() (string, error) {
	currentTime := conf.TimeToString(conf.TimeNow())
	path := filepath.Join(conf.MyNode(), currentTime)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", help.WrapErr(err, "creating latest folder")
	}
	return path, nil
}

func removeChildDirs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return help.WrapErr(err, "reading files in backup folder")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err := os.RemoveAll(filepath.Join(dir, entry.Name()))
			if err != nil {
				return help.WrapErr(err, "deleting stale folders in backup")
			}
		}
	}
	return nil
}
