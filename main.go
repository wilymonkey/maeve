package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/hashsums"
	"github.com/wilymonkey/maeve/tui"
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

	// Hidden, internal use only
	flagHashes := fs.String("latest-hashes", "", "")
	flagNodePath := fs.String("node-path", "", "")

	fs.Parse(os.Args[1:])

	if err := cfg.ReadConfig(); err != nil {
		log.Fatalf("Failed to read config ⇒  %v", err)
	}

	switch {
	case *flagBackupAll:
		tui.AsBackup()
		return

	case *flagNodePath != "":
		nodePath(*flagNodePath)
		return

	case *flagHashes != "":
		latestHashes(*flagHashes)
		return

	default:
		tui.AsHome()
	}
}

// Print to os.Stdout the latest hashes for a given node.
func latestHashes(node string) {
	dir, err := cfg.Global.NodeDirLatest(node)
	if err != nil {
		log.Fatalf("Failure to get hashes ⇒  %v", err)
	}
	hashes, err := hashsums.ReadFileHashes(dir)
	if err != nil {
		log.Fatalf("Failure to get hashes ⇒  %v", err)
	}

	enc := gob.NewEncoder(os.Stdout)
	if err := enc.Encode(hashes); err != nil {
		log.Fatalf("Failure to encode hashes ⇒  %v", err)
	}
}

func nodePath(dir string) {
	// TODO: Are we keeping this?
}
