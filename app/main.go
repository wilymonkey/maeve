package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/gui"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/rpc"
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
		startGUI(backupWindow)
		return

	case *flagVersion:
		printVersion()
		return

	case *flagServer:
		log.Println("RPC Server: Starting...")
		if err := rpc.RunServer(); err != nil {
			log.Fatalf("RPC Server Failed:  %v", err)
		}
		log.Println("RPC Server: Exiting.")
		return

	default:
		overseer.GlobalInteractive = true
		startGUI(gui.Render)
	}
}

func printVersion() {
	fmt.Printf("%s\n", conf.Version)
	os.Exit(0)
}

func startGUI(content func() fyne.CanvasObject) {
	gui.Global = gui.NewState()

	a := app.NewWithID("wilymonkey/maeve")
	a.Settings().SetTheme(&theme.Theme{})
	w := a.NewWindow("Maeve")
	w.SetPadded(false)
	w.SetContent(content())

	if err := gui.Global.Load(w); err != nil {
		log.Fatalf("Failed to start UI:  %v", err)
	}

	w.ShowAndRun()
}

func backupWindow() fyne.CanvasObject {
	state := backup.NewState()
	body := backup.Dialog(state)
	cancelBtn := widget.NewButton(
		"Cancel",
		func() {
			backup.OnCancel(state)
			fyne.CurrentApp().Quit()
		},
	)
	cancelBtn.Importance = widget.DangerImportance
	return container.NewPadded(
		container.NewBorder(
			nil,
			cancelBtn,
			nil,
			nil,
			body,
		),
	)
}
