package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/wilymonkey/maeve/installer/theme"
)

func main() {
	Global = NewState()

	a := app.New()
	a.Settings().SetTheme(&theme.Theme{})
	w := a.NewWindow("Maeve Installer")
	w.SetPadded(false)
	w.SetContent(makeGUI())

	w.ShowAndRun()
}
