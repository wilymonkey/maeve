package backup

import (
	"context"
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/theme"
	"github.com/wilymonkey/maeve/utils"
)

type guiState struct {
	linkDirs    []dirMeta
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

func Dialog() (fyne.CanvasObject, guiState) {
	m := New()
	return container.NewVBox(
		theme.NewH2("Checking PC State"),
		theme.NewH2("Preparing Backup Files"),
		theme.NewH2("Send to PCs"),
	), m
}

func OnCancel(m guiState) {
	m.ctxCancel()
}

func New() guiState {
	linkDirs := make([]dirMeta, len(conf.GetConf().SourceDirs))
	for i, dir := range conf.GetConf().SourceDirs {
		linkDirs[i] = dirMeta{
			path:     dir,
			number:   binding.NewInt(),
			size:     binding.NewInt(),
			hashsums: binding.NewFloat(),
		}
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	return guiState{
		linkDirs:  linkDirs,
		ctx:       ctx,
		ctxCancel: ctxCancel,
		pushProg:  newPushProgress(0),
	}
}

func makeMetaTable(metaMap map[string]dirMeta) *widget.Table {
	keys := make([]string, 0, len(metaMap))
	for k := range metaMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	totalRows := len(keys) + 1
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
				meta := metaMap[keys[row-1]]
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
