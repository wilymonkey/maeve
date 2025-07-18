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
		theme.BlueBg(
			container.NewCenter(
				container.NewVBox(
					theme.NewH1("Components"),
					widget.NewLabel("SSH Installed"),
					widget.NewLabel("SSH Running"),
					widget.NewLabel("Maeve"),
				),
			),
		),
		layout.NewSpacer(),
		container.NewVBox(
			theme.NewH1("Maeve Installer"),
			theme.HighButton("Install", func() {}),
		),
	)
}
