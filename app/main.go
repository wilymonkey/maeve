package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	tea "github.com/charmbracelet/bubbletea"
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

	if err := conf.LoadConfig(&conf.Global); err != nil {
		log.Fatalf("Failed to read config:  %v", err)
	}

	switch {
	case *flagBackupAll:
		startTui(backup.New())
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
		startGUI()
	}
}

func printVersion() {
	fmt.Printf("%s\n", conf.Version)
	os.Exit(0)
}

func startTui(start tea.Model) {
	overseer.Global = tea.NewProgram(overseer.New(start), tea.WithAltScreen())

	if _, err := overseer.Global.Run(); err != nil {
		log.Fatalf("Failed to start TUI:  %v", err)
	}
}

func startGUI() {
	gui.Global = gui.NewState()

	a := app.NewWithID("wilymonkey/maeve")
	a.Settings().SetTheme(&theme.Theme{})
	w := a.NewWindow("Maeve")
	w.SetPadded(false)
	w.SetContent(gui.Render())

	if err := gui.Global.Load(w); err != nil {
		log.Fatalf("Failed to start UI:  %v", err)
	}

	w.ShowAndRun()
}
