package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/home"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/rpc"
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

	if err := cfg.ReadConfig(); err != nil {
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
		startTui(home.New())
	}
}

func printVersion() {
	fmt.Printf("%s\n", cfg.VERSION)
	os.Exit(0)
}

func startTui(start tea.Model) {
	overseer.Global = tea.NewProgram(overseer.New(start), tea.WithAltScreen())

	if _, err := overseer.Global.Run(); err != nil {
		log.Fatalf("Failed to start TUI:  %v", err)
	}
}
