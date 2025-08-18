package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/fynext"
)

func makeGUI() fyne.CanvasObject {
	return container.NewScroll(
		container.NewPadded(
			container.NewBorder(
				container.NewHBox(
					favicon(64),
					fynext.H1("Maeve Installer"),
				),
				fynext.HighBtn("Install", func() {
					if err := Global.Install(); err != nil {
						panic(err)
					}
				}),
				nil,
				nil,
				body(),
			),
		),
	)
}

func body() *fyne.Container {
	keysDisabled := container.NewCenter(fynext.H2("Backup PCs Only"))
	keyTable := widget.NewListWithData(Global.sshKeys,
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			deleteBtn := fynext.DeleteBtn(func() {})
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
				Global.sshKeys.Remove(val)
			}
		})
	keyTable.Hide()
	sshKeys := container.NewScroll(
		container.NewPadded(
			container.NewStack(
				keysDisabled,
				keyTable,
			),
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
		fynext.H2("Is this a Backup PC?"),
		widget.NewCheckWithData("Yes, this is where backups will be kept.", Global.isServer),
		fynext.H2("Maeve Keys"),
		serverExp,
		fynext.GreyBox(sshKeys),
		inputSSH(),
	)
}

func statusAndLabel(t string, b binding.Float) *fyne.Container {
	x := fynext.NewPercStatus(b)
	x.Resize(fyne.NewSize(20, 20))
	return container.New(
		layout.NewCustomPaddedHBoxLayout(0),
		x,
		widget.NewLabel(t),
	)
}

func inputSSH() *fyne.Container {
	inputEntry := widget.NewEntry()
	inputEntry.Disable()
	Global.isServer.AddListener(binding.NewDataListener(func() {
		isServer, err := Global.isServer.Get()
		if err != nil {
			panic(err)
		}
		inputEntry.SetText("")
		if isServer {
			inputEntry.Enable()
			inputEntry.SetPlaceHolder("Paste key here...")
		} else {
			inputEntry.Disable()
			inputEntry.SetPlaceHolder("")
		}
	}))

	addButton := fynext.HighBtn("  +  ", func() {
		Global.AddKey(inputEntry.Text)
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

//go:embed icon.svg
var faviconIcon []byte
var faviconSvg = &fyne.StaticResource{
	StaticName:    "favicon.svg",
	StaticContent: faviconIcon,
}

func favicon(size float32) fyne.CanvasObject {
	img := canvas.NewImageFromResource(faviconSvg)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(size, size))
	return img
}
