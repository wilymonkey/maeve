package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
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

	if err := ReadConfig(); err != nil {
		log.Fatalf("Failed to read config ⇒  %v", err)
	}

	switch {

	case *flagNodePath != "":
		nodePath(*flagNodePath)
		return

	case *flagBackupAll:
		backupAll()
		return

	case *flagHashes != "":
		latestHashes(*flagHashes)
		return

	default:
		if err := Status(); err != nil {
			log.Fatalf("Unable to start TUI ⇒  %v", err)
		}
	}
}

// Print to os.Stdout the latest hashes for a given node.
func latestHashes(node string) {
	dir, err := Config.NodeDirLatest(node)
	if err != nil {
		log.Fatalf("Failure to get hashes ⇒  %v", err)
	}
	hashes, err := ReadFileHashes(dir)
	if err != nil {
		log.Fatalf("Failure to get hashes ⇒  %v", err)
	}

	enc := gob.NewEncoder(os.Stdout)
	if err := enc.Encode(hashes); err != nil {
		log.Fatalf("Failure to encode hashes ⇒  %v", err)
	}
}

func backupAll() {
	if err := LocalPull(); err != nil {
		log.Fatalf("Failed to clone directories ⇒  %v", err)
	}
	fmt.Println("Local directories cloned with hardlinks")

	for _, nodeAddress := range Config.RemoteNodes {
		if err := LocalPush(nodeAddress); err != nil {
			log.Printf("Failed to sync to %s ⇒  %v", nodeAddress, err)
			continue
		}
		fmt.Printf("Successfully synced and snapshotted to %s\n", nodeAddress)
	}

}

func nodePath(dir string) {
	// TODO: Are we keeping this?
}
