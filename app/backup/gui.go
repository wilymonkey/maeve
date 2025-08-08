package backup

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/theme"
	"github.com/wilymonkey/maeve/utils"
)

type guiState struct {
	version       binding.Float
	backupDirMeta []BackupDirMeta
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

func Dialog(state guiState, exitNow func(), exitOnDone bool) fyne.CanvasObject {
	go func() {
		if err := Backup(state); err != nil {
			state.ctxCancel()
			state.err.Set(err.Error())
		} else if exitOnDone {
			exitNow()
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

	cancelBtn := widget.NewButton(
		"Cancel",
		func() {
			state.ctxCancel()
			exitNow()
		},
	)
	cancelBtn.Importance = widget.DangerImportance
	go func() {
		<-state.ctx.Done()
		cancelBtn.SetText("Okay")
		cancelBtn.Importance = widget.SuccessImportance
		cancelBtn.Refresh()
	}()

	return container.NewBorder(
		nil,
		cancelBtn,
		nil,
		nil,
		body,
	)
}

func NewState() guiState {
	backupDirs := make([]BackupDirMeta, len(conf.GetConf().BackupDirs))
	for i, dir := range conf.GetConf().BackupDirs {
		backupDirs[i] = BackupDirMeta{
			path:     dir,
			number:   binding.NewInt(),
			size:     binding.NewInt(),
			hashsums: binding.NewFloat(),
		}
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	return guiState{
		backupDirMeta: backupDirs,
		err:           binding.NewString(),
		ctx:           ctx,
		ctxCancel:     ctxCancel,

		// OLD
		pushProg: newPushProgress(0),
	}
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
