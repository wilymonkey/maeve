package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/wilymonkey/maeve/app/backup"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/gui"
	"github.com/wilymonkey/maeve/app/remote"
	"github.com/wilymonkey/maeve/fynext"
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

	// Hidden, internal use only
	flagServer := fs.Bool("server", false, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse args: %v", err)
	}

	switch {
	case *flagDaemon:
		remote.RunSSHServer()
		return

	case *flagBackup:
		startApp(func(app fyne.App) fyne.Window {
			return backup.Launch(app, true)
		})
		return

	case *flagVersion:
		fmt.Printf("%s\n", conf.Version)
		return

	case *flagServer:
		if err := remote.RunServer(); err != nil {
			log.Fatalf("RPC Server Failed:  %v", err)
		}
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
			app.(desktop.App).SetSystemTrayIcon(gui.FaviconIco)

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
