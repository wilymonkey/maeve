package backup

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
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
	window        fyne.Window

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

func NewState(window fyne.Window) guiState {
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
		window:        window,

		// OLD
		pushProg: newPushProgress(0),
	}
}

type nodeState struct {
	err binding.String
}

func Launch(app fyne.App, exitOnDone bool) fyne.Window {
	window := app.NewWindow("Maeve - Backing Up")
	state := NewState(window)
	window.SetContent(mainWindow(state, exitOnDone))
	return window
}

func mainWindow(state guiState, exitOnDone bool) fyne.CanvasObject {
	go func() {
		if err := Backup(state); err != nil {
			state.FatalErr(err)
		}
		for _, node := range state.nodeStates {
			if utils.GetOrPanic(node.err) != "" {
				state.FatalErr(errors.New("One or more PCs failed to sync"))
				break
			}
		}
		state.ctxCancel()
		if utils.GetOrPanic(state.err) != "" {
			showError(state)
		} else if exitOnDone {
			fyne.Do(state.window.Close)
		}
	}()

	cancelBtn := func() fyne.CanvasObject {
		var cancelBtn *widget.Button
		cancelBtn = widget.NewButton(
			"Cancel",
			func() {
				if cancelBtn.Importance == widget.MediumImportance {
					state.ctxCancel()
				} else {
					fyne.Do(state.window.Close)
				}
			},
		)
		go func() {
			<-state.ctx.Done()
			fyne.Do(func() {
				cancelBtn.SetText("Okay")
				cancelBtn.Importance = widget.SuccessImportance
				cancelBtn.Refresh()
			})
		}()
		return cancelBtn
	}

	return container.NewBorder(
		nil,
		cancelBtn(),
		nil,
		nil,
		container.NewGridWithRows(3,
			container.NewBorder(
				theme.NewH2("Checking PC State"),
				nil,
				nil,
				nil,
			),
			container.NewBorder(
				theme.NewH2("Preparing Backup Files"),
				nil,
				nil,
				nil,
				backupDirMetaTable(state.backupDirMeta),
			),
			container.NewBorder(
				theme.NewH2("Send to PCs"),
				nil,
				nil,
				nil,
				nodeStatusTable(state.nodeStates),
			),
		),
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

func nodeStatusTable(nodeStates map[string]*nodeState) fyne.CanvasObject {
	cfg := conf.GetConf()

	table := widget.NewTable(
		func() (int, int) {
			totalRows := len(nodeStates) + 1
			totalCols := 2
			return totalRows, totalCols
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(cell widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)

			if cell.Row == 0 {
				switch cell.Col {
				case 0:
					label.SetText("PC")
				case 1:
					label.SetText("State")
					label.Alignment = fyne.TextAlignCenter
				}
			} else {
				name := cfg.RemoteNodes[cell.Row-1]
				pc := nodeStates[name]
				switch cell.Col {
				case 0:
					truncPath := utils.TruncateStr(name, 25)
					label.SetText(truncPath)
				case 1:
					label.Bind(pc.err)
					label.Wrapping = fyne.TextWrapWord
				}
			}
		},
	)
	table.SetColumnWidth(0, 250)
	table.SetColumnWidth(1, 300)
	scroll := container.NewStack(table)
	scroll.Resize(fyne.NewSize(600, 200))
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

func showError(state guiState) {
	var d *dialog.CustomDialog

	cancelBtn := func() fyne.CanvasObject {
		cancelBtn := widget.NewButton("OK", func() { d.Dismiss() })
		cancelBtn.Importance = widget.HighImportance
		return cancelBtn
	}

	d = dialog.NewCustomWithoutButtons(
		"ERROR",
		container.NewBorder(
			nil,
			cancelBtn(),
			nil,
			nil,
			widget.NewLabelWithData(state.err),
		),
		state.window,
	)
	d.Show()
}
