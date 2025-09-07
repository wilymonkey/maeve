package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/gui"
	"github.com/wilymonkey/maeve/icons"
	"github.com/wilymonkey/maeve/remote"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fs.VisitAll(func(f *flag.Flag) {
			if f.Usage != "" {
				fmt.Fprintf(os.Stderr, "  -%s\t%s\n", f.Name, f.Usage)
			}
		})
	}

	// Visible
	flagBackup := fs.Bool("backup", false, "Backup to all nodes in config")
	flagVersion := fs.Bool("version", false, "Print the current version")
	flagDaemon := fs.Bool("daemon", false, "Run in daemon mode")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse args: %v", err)
	}

	switch {
	case *flagDaemon:
		remote.RunServer()
		return

	case *flagBackup:
		startApp(func(app fyne.App) fyne.Window {
			return backup.Launch(app, true)
		})
		return

	case *flagVersion:
		fmt.Printf("%s\n", conf.Version)
		return

	default:
		startApp(func(app fyne.App) fyne.Window {
			w := app.NewWindow("Maeve")
			gui.LoadState(w)
			w.SetContent(gui.Render())
			w.SetCloseIntercept(func() {
				w.Hide()
			})

			app.(desktop.App).SetSystemTrayMenu(maevetray(w))
			app.(desktop.App).SetSystemTrayIcon(icons.FaviconIco)

			return w
		})
	}
}

func startApp(window func(app fyne.App) fyne.Window) {
	a := app.NewWithID("Maeve")
	a.Settings().SetTheme(&fynext.Theme{})
	w := window(a)
	w.SetPadded(false)
	w.ShowAndRun()
}

func maevetray(w fyne.Window) *fyne.Menu {
	return fyne.NewMenu("Maeve",
		fyne.NewMenuItem("Show", func() {
			w.Show()
		}),
	)
}
