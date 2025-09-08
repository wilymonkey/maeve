package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/icons"
	"github.com/wilymonkey/maeve/utils"
	"log"
)

func Render() fyne.CanvasObject {
	go func() {
		for {
			log.Println("hello! still runinng")
			utils.Sleep(1000)
		}
	}()
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

	favicon := func() fyne.CanvasObject {
		img := canvas.NewImageFromResource(icons.FaviconSvg)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSquareSize(42))
		return img
	}

	main := container.NewBorder(
		container.NewHBox(
			favicon(),
			fynext.H1("Maeve"),
		),
		fynext.HighBtn("Backup Now", launchBackup),
		nil, nil,
		container.NewGridWithRows(2,
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
		),
	)

	return container.NewPadded(main)
}
