package main

import (
	"time"

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

	go func() {
		for {
			time.Sleep(1 * time.Second)
			currentVal, _ := Global.hasSSH.Get()
			Global.hasSSH.Set(!currentVal)
		}
	}()

	w.ShowAndRun()
}
