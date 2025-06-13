package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	snapshotNode := flag.String("snapshot", "", "Create a hardlink based snapshot of a given node")
	nodePath := flag.String("node-path", "", "Get the path of a given node; creates the path if it doesn't exist")
	backupAll := flag.Bool("backup-all", false, "Backup to all nodes in config")
	flag.Parse()

	if err := ReadConfig(); err != nil {
		log.Fatalf("Failed to read config ⇒  %v", err)
	}

	switch {

	case *snapshotNode != "":
		if err := Snapshot(*snapshotNode); err != nil {
			log.Fatalf("Failed to create snapshot ⇒  %v", err)
		}
		return

	case *nodePath != "":
		nodeDir := NodeDirLatest(*nodePath)
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			log.Fatalf("Failed to create node folder ⇒  %v", err)
		}
		fmt.Print(nodeDir)
		return

	case *backupAll:
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

	default:
		if err := Status(); err != nil {
			log.Fatalf("Unable to start TUI ⇒  %v", err)
		}
	}
}
