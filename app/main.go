package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/gui"
	"github.com/wilymonkey/maeve/remote"
	"github.com/wilymonkey/maeve/theme"
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
	flagBackupAll := fs.Bool("backup-all", false, "Backup to all nodes in config")
	flagVersion := fs.Bool("version", false, "Print the current version")

	// Hidden, internal use only
	flagServer := fs.Bool("server", false, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse args: %v", err)
	}

	switch {
	case *flagBackupAll:
		startApp(func(app fyne.App) fyne.Window {
			return backup.Launch(app, true)
		})
		return

	case *flagVersion:
		fmt.Printf("%s\n", conf.Version)
		return

	case *flagServer:
		log.Println("RPC Server: Starting...")
		if err := remote.RunServer(); err != nil {
			log.Fatalf("RPC Server Failed:  %v", err)
		}
		log.Println("RPC Server: Exiting.")
		return

	default:
		startApp(func(app fyne.App) fyne.Window {
			w := app.NewWindow("Maeve")
			gui.LoadState(w)
			w.SetContent(gui.Render())
			return w
		})
	}
}

func startApp(window func(app fyne.App) fyne.Window) {
	a := app.NewWithID("wilymonkey/maeve")
	a.Settings().SetTheme(&theme.Theme{})
	w := window(a)
	w.ShowAndRun()
}
