package gui

import (
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/theme"
)

func Render() fyne.CanvasObject {
	return container.NewScroll(
		container.NewPadded(
			container.NewBorder(
				container.NewHBox(
					favicon(64),
					theme.NewH1("Maeve"),
				),
				theme.HighBtn("Backup Now", StartBackup),
				nil,
				nil,
				body(),
			),
		),
	)
}

func body() *fyne.Container {
	nodes := container.NewScroll(
		container.NewCenter(theme.NewH2("Backup Locations")),
	)
	nodes.SetMinSize(fyne.NewSquareSize(200))

	dirs := container.NewScroll(
		container.NewCenter(theme.NewH2("Folders")),
	)
	dirs.SetMinSize(fyne.NewSquareSize(200))

	return container.NewVBox(
		theme.NewH2("This PC"),
		thisPC(),
		theme.NewH2("Folders"),
		sourceDirs(),
		addSourceDir(),
		theme.NewH2("Backup PCs"),
		remoteNotes(),
		addRemoteNote(),
	)
}

func remoteNotes() fyne.CanvasObject {
	w := container.NewScroll(
		theme.GreyBox(
			container.NewPadded(
				widget.NewListWithData(Global.RemoteNodes,
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
							Global.RemoteNodes.Remove(val)
						}
					},
				),
			),
		),
	)
	w.SetMinSize(fyne.NewSquareSize(200))
	return w
}

func addRemoteNote() *fyne.Container {
	inputEntry := widget.NewEntry()
	addButton := theme.HighBtn("  +  ", func() {
		Global.RemoteNodes.Append(inputEntry.Text)
		inputEntry.SetText("")
	})
	addButton.Disable()
	inputEntry.OnChanged = func(s string) {
		r := regexp.MustCompile(`^[a-zA-Z0-9_]+@(?:\d{1,3}\.){3}\d{1,3}:\d{1,5}$`)
		if r.MatchString(s) {
			addButton.Enable()
		} else {
			addButton.Disable()
		}
	}

	return container.NewBorder(
		nil,
		nil,
		nil,
		addButton,
		inputEntry,
	)
}

func StartBackup() {
	var d *dialog.CustomDialog
	body, m := backup.Dialog()
	cancelBtn := widget.NewButton(
		"Cancel",
		func() {
			backup.OnCancel(m)
			d.Hide()
		},
	)
	cancelBtn.Importance = widget.DangerImportance
	d = dialog.NewCustomWithoutButtons(
		"Backing Up",
		container.NewBorder(
			nil,
			cancelBtn,
			nil,
			nil,
			body,
		),
		Global.Window,
	)
	d.Show()
}
