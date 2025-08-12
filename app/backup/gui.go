package backup

import (
	"context"
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/theme"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
)

type guiState struct {
	backupDirMeta []BackupDirMeta
	nodeStates    map[string]*nodeState
	err           binding.String
	ctx           context.Context
	ctxCancel     context.CancelFunc

	// OLD
	hashDone    bool
	totalFiles  int
	hashNum     int
	pushProg    pushProgress
	pushingNode int
	pushDone    bool
}

func (s *guiState) ErrGroup(scaling int) (*errgroup.Group, context.Context) {
	eGrp, ctx := errgroup.WithContext(s.ctx)
	eGrp.SetLimit(scaling * runtime.NumCPU())
	return eGrp, ctx
}

func (s *guiState) FatalErr(err error) {
	s.ctxCancel()
	if utils.GetOrPanic(s.err) == "" {
		s.err.Set(err.Error())
	}
}

func NewState() guiState {
	cfg := conf.GetConf()

	backupDirs := make([]BackupDirMeta, len(cfg.BackupDirs))
	for i, dir := range cfg.BackupDirs {
		backupDirs[i] = BackupDirMeta{
			path:     dir,
			number:   binding.NewInt(),
			size:     binding.NewInt(),
			hashsums: binding.NewFloat(),
		}
	}

	nodeStates := make(map[string]*nodeState, len(cfg.RemoteNodes))
	for _, node := range cfg.RemoteNodes {
		nodeStates[node] = &nodeState{
			err: binding.NewString(),
		}
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	return guiState{
		backupDirMeta: backupDirs,
		nodeStates:    nodeStates,
		err:           binding.NewString(),
		ctx:           ctx,
		ctxCancel:     ctxCancel,

		// OLD
		pushProg: newPushProgress(0),
	}
}

type nodeState struct {
	err binding.String
}

func Dialog(state guiState, exitNow func(), exitOnDone bool) fyne.CanvasObject {
	go func() {
		if err := Backup(state); err != nil {
			state.FatalErr(err)
		} else if exitOnDone {
			fyne.Do(exitNow)
		}
	}()
	body := container.NewScroll(
		container.NewVBox(
			theme.NewH2("Checking PC State"),
			backupDirMetaTable(state.backupDirMeta),
			theme.NewH2("Preparing Backup Files"),
			theme.NewH2("Send to PCs"),
			ErrorLabel(state.err),
		),
	)
	body.SetMinSize(fyne.NewSize(600, 500))

	var cancelBtn *widget.Button
	cancelBtn = widget.NewButton(
		"Cancel",
		func() {
			if cancelBtn.Importance == widget.DangerImportance {
				state.ctxCancel()
			} else {
				fyne.Do(exitNow)
			}
		},
	)
	cancelBtn.Importance = widget.DangerImportance
	go func() {
		<-state.ctx.Done()
		fyne.Do(func() {
			cancelBtn.SetText("Okay")
			cancelBtn.Importance = widget.SuccessImportance
			cancelBtn.Refresh()
		})
	}()

	return container.NewBorder(
		nil,
		cancelBtn,
		nil,
		nil,
		body,
	)
}

func backupDirMetaTable(backupDirMeta []BackupDirMeta) fyne.CanvasObject {
	totalRows := len(backupDirMeta) + 1
	totalCols := 4

	table := widget.NewTable(
		func() (int, int) {
			return totalRows, totalCols
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(cell widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			row := cell.Row
			col := cell.Col

			if row == 0 {
				switch col {
				case 0:
					label.SetText("Dir")
				case 1:
					label.SetText("Files")
					label.Alignment = fyne.TextAlignCenter
				case 2:
					label.SetText("Size")
					label.Alignment = fyne.TextAlignCenter
				case 3:
					label.SetText("Hashsums")
					label.Alignment = fyne.TextAlignCenter
				}
			} else {
				meta := backupDirMeta[row-1]
				switch col {
				case 0:
					truncPath := utils.TruncateStr(meta.path, 25)
					label.SetText(truncPath)
				case 1:
					label.SetText(fmt.Sprintf("%d", utils.GetOrPanic(meta.number)))
					label.Alignment = fyne.TextAlignCenter
				case 2:
					label.SetText(fmt.Sprintf("%d", utils.GetOrPanic(meta.size)))
					label.Alignment = fyne.TextAlignCenter
				case 3:
					perc := int(utils.GetOrPanic(meta.hashsums) * 100)
					label.SetText(fmt.Sprintf("%d", perc))
					label.Alignment = fyne.TextAlignCenter
				}
			}
		},
	)
	table.SetColumnWidth(0, 250)
	table.SetColumnWidth(1, 90)
	table.SetColumnWidth(2, 90)
	table.SetColumnWidth(3, 90)
	scroll := container.NewScroll(table)
	scroll.SetMinSize(fyne.NewSize(600, 200))
	return scroll
}

func ErrorLabel(err binding.String) fyne.CanvasObject {
	dangerBox := theme.ErrorBox(err)
	dangerBox.Hide()
	err.AddListener(binding.NewDataListener(func() {
		if s, e := err.Get(); s != "" && e == nil {
			dangerBox.Show()
		}
	}))
	return dangerBox
}
