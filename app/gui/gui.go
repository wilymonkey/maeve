package gui

import (
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/backup"
	"github.com/wilymonkey/maeve/fynext"
)

var sshNodeReg = regexp.MustCompile(`^[a-zA-Z0-9_]+@(?:\d{1,3}\.){3}\d{1,3}:\d{1,5}$`)

func Render() fyne.CanvasObject {
	launchBackup := func() {
		lock := dialog.NewCustomWithoutButtons(
			"Backup Lock",
			widget.NewLabel("Currently backing up, stop it unlock this window."),
			global.Window,
		)
		w := backup.Launch(fyne.CurrentApp(), false)
		w.SetOnClosed(func() { lock.Dismiss() })
		w.Show()
		lock.Show()
	}

	main := container.NewBorder(
		container.NewHBox(
			favicon(42),
			fynext.H1("Maeve"),
		),
		fynext.HighBtn("Backup Now", launchBackup),
		nil, nil,
		container.NewGridWithRows(3,
			container.NewBorder(
				fynext.H2("This PC"),
				nil, nil, nil,
				thisPC(),
			),
			container.NewBorder(
				fynext.H2("Folders"),
				addSourceDir(),
				nil, nil,
				sourceDirs(),
			),
			container.NewBorder(
				fynext.H2("Backup PCs"),
				addRemoteNote(),
				nil, nil,
				remoteNotes(),
			),
		),
	)

	return container.NewVScroll(
		container.NewPadded(main),
	)
}

func remoteNotes() fyne.CanvasObject {
	emptyRows := func() fyne.CanvasObject {
		label := widget.NewLabel("")
		deleteBtn := fynext.DeleteBtn(func() {})
		return container.NewBorder(
			nil, nil, nil,
			deleteBtn,
			label,
		)
	}
	updateRows := func(i binding.DataItem, o fyne.CanvasObject) {
		row := o.(*fyne.Container)
		label := row.Objects[0].(*widget.Label)
		deleteBtn := row.Objects[1].(*widget.Button)

		label.Bind(i.(binding.String))
		label.Truncation = fyne.TextTruncateEllipsis

		deleteBtn.OnTapped = func() {
			global.RemoteNodes.Remove(label.Text)
		}
	}
	w := container.NewVScroll(
		fynext.GreyBox(
			container.NewPadded(
				widget.NewListWithData(
					global.RemoteNodes,
					emptyRows,
					updateRows,
				),
			),
		),
	)
	return w
}

func addRemoteNote() *fyne.Container {
	inputEntry := widget.NewEntry()
	addButton := fynext.HighBtn("  +  ", func() {
		global.RemoteNodes.Append(inputEntry.Text)
		inputEntry.SetText("")
	})
	addButton.Disable()
	inputEntry.OnChanged = func(s string) {
		if sshNodeReg.MatchString(s) {
			addButton.Enable()
		} else {
			addButton.Disable()
		}
	}

	return container.NewBorder(
		nil, nil, nil,
		addButton,
		inputEntry,
	)
}
