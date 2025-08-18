package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/hashsums"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
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

func pullStateUI() fyne.CanvasObject {
	fmtState := func(pstate *pullState) string {
		num := fynext.Unwrap(pstate.number)
		size := utils.BytesToHuman(fynext.Unwrap(pstate.size))
		path := utils.TruncateString(pstate.path, 30)
		return fmt.Sprintf("%s: Found %d, totalling %s", path, num, size)
	}

	objects := make([]fyne.CanvasObject, 0, len(guiState.pullStates))

	for _, pstate := range guiState.pullStates {
		label := widget.NewLabel("")
		objects = append(objects, label)

		var showLabel binding.DataListener
		showLabel = binding.NewDataListener(func() {
			if fynext.Unwrap(pstate.number) > 0 {
				label.Show()
				pstate.number.RemoveListener(showLabel)
			}
		})

		pstate.number.AddListener(binding.NewDataListener(func() {
			label.SetText(fmtState(pstate))
		}))

	}

	vbox := container.NewVBox()
	vbox.Objects = objects
	return container.NewVScroll(vbox)
}

// =======================================
// METHODS
// =======================================

func newLatest() error {
	if err := local.RemoveChildDirs(conf.MyNode()); err != nil {
		return err
	}
	latest, err := local.NewLatestDir()
	if err != nil {
		return err
	}

	for _, pState := range guiState.pullStates {
		targetDir := filepath.Join(latest, filepath.Base(pState.path))

		metaChan := make(chan *hashsums.FileMeta, 100)
		var totalSize int64
		var totalFiles int
		var fileMetas []*hashsums.FileMeta

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
			return err
		}
		close(metaChan)
	}

	return nil
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
