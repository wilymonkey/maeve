package backup

import (
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/hashsums"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
	"zombiezen.com/go/sqlite"
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
	sourceDirs := make([]*pullState, 0, len(cfg.SourceDirs))
	for _, dir := range cfg.SourceDirs {
		sourceDirs = append(sourceDirs, &pullState{
			path:   dir,
			number: binding.NewInt(),
			size: binding.NewItem(func(i1, i2 int64) bool {
				return i1 == i2
			}),
		})
	}
	return sourceDirs
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

func newLatest(conn *sqlite.Conn) ([]*hashsums.FileMeta, error) {
	if err := local.RemoveChildDirs(conf.MyNode()); err != nil {
		return nil, err
	}
	latest, err := local.NewLatestDir()
	if err != nil {
		return nil, err
	}

	var fileMetas []*hashsums.FileMeta

	for _, pState := range guiState.pullStates {
		targetDir := filepath.Join(latest, filepath.Base(pState.path))

		metaChan := make(chan *hashsums.FileMeta, 100)
		var totalSize int64
		var totalFiles int

		utils.Throttle(
			metaChan,
			func(meta *hashsums.FileMeta) {
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

		if err := walkDir(pState.path, targetDir, metaChan); err != nil {
			return nil, err
		}
		close(metaChan)
	}
	return fileMetas, nil
}

func walkDir(sourceDir, targetDir string, metaChan chan *hashsums.FileMeta) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return help.CheckBackupDir(err, "deleting folder to link things to")
	}

	numWorkers := runtime.NumCPU() * 2
	paths := make(chan string, numWorkers*2)
	eGrp, ctx := errgroup.WithContext(guiState.ctx)
	for range numWorkers {
		eGrp.Go(func() error {
			for path := range paths {
				meta, err := processFiles(sourceDir, targetDir, path)
				if err != nil {
					return err
				}
				metaChan <- meta
			}
			return nil
		})
	}

	walkErr := filepath.WalkDir(sourceDir, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if dir.Type().IsRegular() {
			paths <- path
		}
		return nil
	})

	close(paths)
	if err := eGrp.Wait(); err != nil {
		return err
	}
	if walkErr != nil {
		return walkErr
	}

	return nil
}

func processFiles(baseDir, targetDir string, absPath string) (*hashsums.FileMeta, error) {
	relPath, err := filepath.Rel(baseDir, absPath)
	if err != nil {
		return nil, help.DevReport(err, "creating relative path")
	}
	targetPath := filepath.Join(targetDir, relPath)
	if err := local.Hardlink(absPath, targetPath); err != nil {
		return nil, err
	}
	meta, err := hashsums.NewFileMeta(targetPath)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}
