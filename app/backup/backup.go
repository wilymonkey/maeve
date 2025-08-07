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

	// OLD
	hashDone    bool
	totalFiles  int
	hashNum     int
	pushProg    pushProgress
	pushingNode int
	pushDone    bool
	err         error
	ctx         context.Context
	ctxCancel   context.CancelFunc
}

func Dialog(state guiState) fyne.CanvasObject {
	return container.NewVBox(
		theme.NewH2("Checking PC State"),
		backupDirMetaTable(state.backupDirMeta),
		theme.NewH2("Preparing Backup Files"),
		theme.NewH2("Send to PCs"),
	)
}

func OnCancel(m guiState) {
	m.ctxCancel()
}

func NewState() guiState {
	linkDirs := make([]BackupDirMeta, len(conf.GetConf().SourceDirs))
	for i, dir := range conf.GetConf().SourceDirs {
		linkDirs[i] = BackupDirMeta{
			path:     dir,
			number:   binding.NewInt(),
			size:     binding.NewInt(),
			hashsums: binding.NewFloat(),
		}
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	return guiState{
		backupDirMeta: linkDirs,
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		pushProg:      newPushProgress(0),
	}
}

func backupDirMetaTable(backupDirMeta []BackupDirMeta) *widget.Table {
	totalRows := len(backupDirMeta) + 1
	totalCols := 3

	return widget.NewTable(
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
				case 2:
					label.SetText("Size")
				case 3:
					label.SetText("Hashsums")
				}
			} else {
				meta := backupDirMeta[row-1]
				switch col {
				case 0:
					label.SetText(meta.path)
				case 1:
					label.SetText(fmt.Sprintf("%d", utils.GetOrPanic(meta.number)))
				case 2:
					label.SetText(fmt.Sprintf("%d", utils.GetOrPanic(meta.size)))
				case 3:
					perc := int(utils.GetOrPanic(meta.hashsums) * 100)
					label.SetText(fmt.Sprintf("%d", perc))
				}
			}
		},
	)
}
