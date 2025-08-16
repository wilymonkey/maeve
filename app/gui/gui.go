package gui

import (
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/theme"
)

func Render() fyne.CanvasObject {
	launchBackup := func() {
		backup.Launch(fyne.CurrentApp(), false).Show()
	}

	return container.NewBorder(
		container.NewHBox(
			favicon(42),
			theme.H1("Maeve"),
		),
		theme.HighBtn("Backup Now", launchBackup),
		nil,
		nil,
		container.NewGridWithRows(3,
			container.NewBorder(
				theme.H2("This PC"),
				nil, nil, nil,
				thisPC(),
			),
			container.NewBorder(
				theme.H2("Folders"),
				addSourceDir(),
				nil, nil,
				sourceDirs(),
			),
			container.NewBorder(
				theme.H2("Backup PCs"),
				addRemoteNote(),
				nil, nil,
				remoteNotes(),
			),
		),
	)
}

func remoteNotes() fyne.CanvasObject {
	emptyRows := func() fyne.CanvasObject {
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
			global.RemoteNodes.Remove(val)
		}
	}
	w := container.NewScroll(
		theme.GreyBox(
			container.NewPadded(
				widget.NewListWithData(
					global.RemoteNodes,
					emptyRows,
					updateRows,
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
		global.RemoteNodes.Append(inputEntry.Text)
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
		nil, nil, nil,
		addButton,
		inputEntry,
	)
}
