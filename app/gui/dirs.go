package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/theme"
)

func sourceDirs() fyne.CanvasObject {
	blankRows := func() fyne.CanvasObject {
		label := widget.NewLabel("")
		deleteBtn := theme.DeleteBtn(func() {})
		return container.NewBorder(
			nil, nil, nil,
			deleteBtn,
			label,
		)
	}
	updateRows := func(i binding.DataItem, o fyne.CanvasObject) {
		row := o.(*fyne.Container)
		label := row.Objects[0].(*widget.Label)
		label.Bind(i.(binding.String))
		label.Truncation = fyne.TextTruncateEllipsis
		deleteBtn := row.Objects[1].(*widget.Button)
		deleteBtn.OnTapped = func() {
			val, err := i.(binding.String).Get()
			if err != nil {
				panic(err)
			}
			global.SourceDirs.Remove(val)
		}
	}
	w := container.NewScroll(
		theme.GreyBox(
			container.NewPadded(
				widget.NewListWithData(
					global.SourceDirs,
					blankRows,
					updateRows,
				),
			),
		),
	)
	w.SetMinSize(fyne.NewSquareSize(200))
	return w
}

func addSourceDir() fyne.CanvasObject {
	dlg := func(w fyne.Window) dialog.Dialog {
		return dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				global.ShowError(err)
				return
			}
			if list != nil {
				global.SourceDirs.Append(list.Path())
			}
		}, w)
	}
	return theme.HighBtn("Add Folder", func() {
		theme.ShowWindowDialog(global.Window, "Select Folder", dlg)
	})
}
