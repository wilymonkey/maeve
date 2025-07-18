package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/installer/theme"
)

func makeGUI() fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		theme.PrimaryBox(
			container.NewCenter(
				container.NewPadded(
					container.NewVBox(
						theme.NewH1("Components"),
						container.New(
							layout.NewCustomPaddedHBoxLayout(0),
							theme.BoolImg(Global.hasSSH),
							widget.NewLabel("SSH Installed"),
						),
						container.New(
							layout.NewCustomPaddedHBoxLayout(0),
							theme.BoolImg(Global.sshRunning),
							widget.NewLabel("SSH Running"),
						),
						container.New(
							layout.NewCustomPaddedHBoxLayout(0),
							theme.BoolImg(Global.hasMaeve),
							widget.NewLabel("Maeve"),
						),
					),
				),
			),
		),
		nil,
		container.NewPadded(
			container.NewBorder(
				container.NewHBox(
					theme.Icon(64),
					theme.NewH1("Maeve Installer"),
				),
				theme.HighButton("Install", func() {}),
				nil,
				nil,
				nil,
			),
		),
	)
}
