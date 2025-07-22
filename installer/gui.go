package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/installer/theme"
)

func makeGUI() fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		sideBanner(),
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
				body(),
			),
		),
	)
}

func body() *fyne.Container {
	logs := container.NewScroll(
		widget.NewListWithData(Global.logs,
			func() fyne.CanvasObject {
				return widget.NewLabel("Log")
			},
			func(i binding.DataItem, o fyne.CanvasObject) {
				o.(*widget.Label).Bind(i.(binding.String))
			}),
	)
	logs.SetMinSize(fyne.NewSize(200, 200))

	sshKeys := container.NewScroll(
		container.NewCenter(
			theme.NewH2("Only for Servers"),
		),
	)
	sshKeys.SetMinSize(fyne.NewSize(200, 200))

	return container.NewVBox(
		theme.NewH2("Install as Server?"),
		widget.NewLabel(`Install as a server to allow files to be received by this PC.
It will only allow connections from PCs which have one of the KEYS below.
Remove keys that are no longer used.`),
		widget.NewCheck("Install as Server", func(b bool) {}),
		theme.NewH2("SSH Keys"),
		theme.LowPriorBox(sshKeys),
		theme.NewH2("Logs"),
		theme.LowPriorBox(logs),
	)
}

func statusAndLabel(t string, b binding.Bool) *fyne.Container {
	return container.New(
		layout.NewCustomPaddedHBoxLayout(0),
		theme.BoolImg(b),
		widget.NewLabel(t),
	)
}

func sideBanner() *fyne.Container {
	return theme.PrimaryBox(
		container.NewCenter(
			container.NewVBox(
				container.NewPadded(
					theme.NewH1("Components"),
				),
				container.NewPadded(
					container.NewVBox(
						statusAndLabel("SSH Installed", Global.hasSSH),
						statusAndLabel("SSH Running", Global.sshRunning),
						statusAndLabel("Maeve", Global.hasMaeve),
					),
				),
			),
		),
	)
}
