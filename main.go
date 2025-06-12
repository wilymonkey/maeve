package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	snapshotNode := flag.String("snapshot", "", "Create a hardlink based snapshot of a given node")
	nodePath := flag.String("node-path", "", "Get the path of a given node; creates the path if it doesn't exist")
	backupAll := flag.Bool("backup-all", false, "Backup to all nodes in config")
	flag.Parse()

	if err := ReadConfig(*configPath); err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			if err := DefaultConfig(*configPath); err != nil {
				log.Fatalf("Failed to create default config ⇒  %v", err)
			}
			// Try to read the newly created config file.
			if err := ReadConfig(*configPath); err != nil {
				log.Fatalf("Failed to load default config ⇒  %v", err)
			}
		} else {
			log.Fatalf("Failed to load config ⇒  %v", err)
		}
	}

	if *snapshotNode != "" {
		if err := Snapshot(*snapshotNode); err != nil {
			log.Fatalf("Failed to create snapshot ⇒  %v", err)
		}
		return
	}

	if *nodePath != "" {
		nodeDir := NodeDirLatest(*nodePath)
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			log.Fatalf("Failed to create node folder ⇒  %v", err)
		}
		fmt.Println(nodeDir)
		return
	}

	if *backupAll {
		if err := LocalPull(); err != nil {
			log.Fatalf("Failed to clone directories ⇒  %v", err)
		}
		fmt.Println("Local directories cloned with hardlinks")

		if err := Snapshot(Cfg.Name); err != nil {
			log.Fatalf("Failed to snapshot after local pull ⇒  %v", err)
		}

		for _, nodeAddress := range Cfg.RemoteNodes {
			if err := LocalPush(nodeAddress); err != nil {
				log.Printf("Failed to sync to %s ⇒  %v", nodeAddress, err)
				continue
			}
			fmt.Printf("Successfully synced and snapshotted to %s\n", nodeAddress)
		}
		return
	}

	if err := Status(); err != nil {
		log.Fatalf("Unable to start TUI ⇒  %v", err)
	}
}
