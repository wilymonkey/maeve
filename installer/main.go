package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/wilymonkey/maeve/theme"
)

func main() {
	Global = NewState()
	defer Global.Close()

	a := app.New()
	a.Settings().SetTheme(&theme.Theme{})
	w := a.NewWindow("Maeve Installer")
	Global.CurrentWindow = w
	w.SetPadded(false)
	w.SetContent(makeGUI())

	w.ShowAndRun()
}
