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
			container.NewScroll(
				container.NewBorder(
					container.NewHBox(
						theme.Favicon(64),
						theme.NewH1("Maeve Installer"),
					),
					theme.HighButton("Install", func() {}),
					nil,
					nil,
					body(),
				),
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
	logs.SetMinSize(fyne.NewSquareSize(180))

	keysDisabled := container.NewCenter(theme.NewH2("Backup PCs Only"))
	keyTable := widget.NewListWithData(Global.sshKeys,
		func() fyne.CanvasObject {
			return widget.NewLabel("Key")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})
	keyTable.Hide()
	sshKeys := container.NewScroll(
		container.NewStack(
			keysDisabled,
			keyTable,
		),
	)
	sshKeys.SetMinSize(fyne.NewSquareSize(200))
	Global.isServer.AddListener(binding.NewDataListener(func() {
		isServer, err := Global.isServer.Get()
		if err != nil {
			panic(err)
		}
		if isServer {
			keysDisabled.Hide()
			keyTable.Show()
		} else {
			keysDisabled.Show()
			keyTable.Hide()
		}
	}))
	serverExp := widget.NewLabel(`Keys are like a name for a particular PC's.
A backup PC will only accept files from the list of names (keys) that you have approved below.
Copy your Maeve key from the PC with the original files and paste it below.
If a key is no longer being used please remove it.`)
	serverExp.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		theme.NewH2("Is this a Backup PC?"),
		widget.NewCheckWithData("Install as Backup PC", Global.isServer),
		theme.NewH2("Maeve Keys"),
		serverExp,
		theme.LowPriorBox(sshKeys),
		inputSSH(),
		theme.NewH2("Logs"),
		theme.LowPriorBox(logs),
	)
}

func statusAndLabel(t string, b binding.Float) *fyne.Container {
	x := theme.NewPercStatus(b)
	x.Resize(fyne.NewSize(20, 20))
	return container.New(
		layout.NewCustomPaddedHBoxLayout(0),
		x,
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
						statusAndLabel("SSH", Global.percSSH),
						statusAndLabel("Maeve", Global.percMaeve),
						statusAndLabel("Integrity", Global.percIntegrity),
					),
				),
			),
		),
	)
}

func inputSSH() *fyne.Container {
	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder("Paste key here...")

	addButton := theme.HighButton("  +  ", func() {
		Global.sshKeys.Append(inputEntry.Text)
		inputEntry.SetText("")
	})
	addButton.Disable()

	inputEntry.OnChanged = func(s string) {
		if isValidKey(s) {
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
