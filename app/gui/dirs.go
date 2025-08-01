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
	w := container.NewScroll(
		theme.LowPriorBox(
			container.NewPadded(
				widget.NewListWithData(Global.SourceDirs,
					func() fyne.CanvasObject {
						label := widget.NewLabel("")
						deleteBtn := theme.DeleteBtn(func() {})
						return container.NewBorder(
							nil,
							nil,
							nil,
							deleteBtn,
							label,
						)
					},
					func(i binding.DataItem, o fyne.CanvasObject) {
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
							Global.SourceDirs.Remove(val)
						}
					},
				),
			),
		),
	)
	w.SetMinSize(fyne.NewSquareSize(200))
	return w
}

func addSourceDir() fyne.CanvasObject {
	return theme.HighBtn("Add Folder", func() {
		dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				Global.ShowError(err)
				return
			}
			if list == nil {
				return // User canceled
			}
			Global.SourceDirs.Append(list.Path())
		}, Global.Window).Show()
	})
}
