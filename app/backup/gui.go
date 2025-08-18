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
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/fynext"
	"golang.org/x/sync/errgroup"
)

var guiState *state

type state struct {
	currTask      binding.Int
	dbState       binding.String
	backupDirMeta []BackupDirMeta
	nodeStates    map[string]*nodeStatus
	ctx           context.Context
	ctxCancel     context.CancelFunc
	window        fyne.Window
	err           binding.Item[error]

	// OLD
	hashDone    bool
	totalFiles  int
	hashNum     int
	pushProg    pushProgress
	pushingNode int
	pushDone    bool
}

func (s *state) ErrGroup(scaling int) (*errgroup.Group, context.Context) {
	eGrp, ctx := errgroup.WithContext(s.ctx)
	eGrp.SetLimit(scaling * runtime.NumCPU())
	return eGrp, ctx
}

type currTask int

const (
	updatingDB currTask = iota
	pulling
	pushing
	done
)

type nodeStatus struct {
	err binding.Item[error]
}

func loadState(window fyne.Window) {
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

	nodeStates := make(map[string]*nodeStatus, len(cfg.RemoteNodes))
	for _, node := range cfg.RemoteNodes {
		nodeStates[node] = &nodeStatus{
			err: fynext.NewErrBinding(),
		}
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	guiState = &state{
		currTask:      binding.NewInt(),
		dbState:       binding.NewString(),
		backupDirMeta: backupDirs,
		nodeStates:    nodeStates,
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		window:        window,
		err:           fynext.NewErrBinding(),

		// OLD
		pushProg: newPushProgress(0),
	}
}

func Launch(app fyne.App, exitOnDone bool) fyne.Window {
	window := app.NewWindow("Maeve - Backing Up")
	loadState(window)
	window.SetContent(mainWindow(exitOnDone))
	return window
}

func mainWindow(exitOnDone bool) fyne.CanvasObject {
	go func() {
		if err := Backup(guiState); err != nil {
			guiState.err.Set(err)
		}
		for _, node := range guiState.nodeStates {
			if fynext.GetOrPanic(node.err) != nil {
				guiState.err.Set(errors.New("One or more PCs failed to sync"))
				break
			}
		}
		guiState.ctxCancel()
		if fynext.GetOrPanic(guiState.err) != nil {
			guiState.showError()
		} else if exitOnDone {
			// TODO: remove this when done.
			// fyne.Do(state.window.Close)
		}
	}()

	cancelBtn := func() fyne.CanvasObject {
		var cancelBtn *widget.Button
		cancelBtn = widget.NewButton(
			"Cancel",
			func() {
				if cancelBtn.Importance == widget.MediumImportance {
					guiState.ctxCancel()
				} else {
					fyne.Do(guiState.window.Close)
				}
			},
		)
		go func() {
			<-guiState.ctx.Done()
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
		nil, nil,
		container.NewGridWithRows(2,
			container.NewBorder(
				fynext.ColorWithin(
					fynext.H2("Updating My State"),
					guiState.currTask,
					updatingDB,
					pulling,
				),
				nil, nil, nil,
				container.NewVBox(
					updateDBUI(),
					backupDirMetaTable(guiState.backupDirMeta),
				),
			),
			container.NewBorder(
				fynext.ColorWithin(
					fynext.H2("Sending Files to PCs"),
					guiState.currTask,
					pushing,
					done,
				),
				nil, nil, nil,
				nodeStatusTable(guiState.nodeStates),
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
			return fynext.LabelDisableUntil(
				widget.NewLabel(""),
				guiState.currTask,
				pulling,
			)
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
					label.SetText(meta.path)
					fynext.TruncLabel(label, 200)
				case 1:
					label.SetText(fmt.Sprintf("%d", fynext.GetOrPanic(meta.number)))
					label.Alignment = fyne.TextAlignCenter
				case 2:
					label.SetText(fmt.Sprintf("%d", fynext.GetOrPanic(meta.size)))
					label.Alignment = fyne.TextAlignCenter
				case 3:
					perc := int(fynext.GetOrPanic(meta.hashsums) * 100)
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
	return table
}

func nodeStatusTable(nodeStatus map[string]*nodeStatus) fyne.CanvasObject {
	cfg := conf.GetConf()
	objects := make([]fyne.CanvasObject, 0, len(cfg.RemoteNodes)*2)
	for _, node := range cfg.RemoteNodes {
		status := nodeStatus[node]
		sLabel := help.Widget(status.err)
		border := container.NewBorder(
			nil, nil,
			fynext.LabelDisableUntil(
				widget.NewLabel(node),
				guiState.currTask,
				pushing,
			),
			nil,
			sLabel,
		)
		sep := widget.NewSeparator()
		objects = append(objects, border, sep)
	}
	vbox := container.NewVBox()
	vbox.Objects = objects[:len(objects)-1] // Remove trailing separator.
	return container.NewVScroll(vbox)
}

func ErrorLabel(err binding.String) fyne.CanvasObject {
	dangerBox := fynext.ErrorBox(err)
	dangerBox.Hide()
	err.AddListener(binding.NewDataListener(func() {
		if s, e := err.Get(); s != "" && e == nil {
			dangerBox.Show()
		}
	}))
	return dangerBox
}

func (state *state) showError() {
	dlg := func(w fyne.Window) dialog.Dialog {
		var d *dialog.CustomDialog

		cancelBtn := func() fyne.CanvasObject {
			cancelBtn := widget.NewButton("OK", func() { d.Dismiss() })
			cancelBtn.Importance = widget.HighImportance
			return cancelBtn
		}

		d = dialog.NewCustomWithoutButtons(
			"ERROR",
			container.NewVBox(
				help.Widget(state.err),
				cancelBtn(),
			),
			w,
		)
		return d
	}

	fyne.Do(func() {
		fynext.ShowWindowDialog(state.window, "ERROR", dlg)
	})
}
