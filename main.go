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
				log.Fatalf("Failed to read default config ⇒  %v", err)
			}
		} else {
			log.Fatalf("Failed to read config ⇒  %v", err)
		}
	}

	if *snapshotNode != "" {
		if err := Snapshot(*snapshotNode); err != nil {
			log.Fatalf("Failed to create snapshot ⇒  %v", err)
		}
		fmt.Println("Snapshot created successfully")
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
		log.Fatalf("Backup not implemented")
		return
	}

	// TODO: Print status

	// err = LocalPull()
	// if err != nil {
	// 	log.Fatalf("Failed to clone directories: %v", err)
	// }
	// fmt.Println("Local directories cloned with hardlinks")

	// for _, node := range Cfg.RemoteNodes {
	// 	err = LocalPush(node)
	// 	if err != nil {
	// 		log.Printf("Failed to sync and snapshot to %s: %v", node, err)
	// 		continue
	// 	}
	// 	fmt.Printf("Successfully synced and snapshotted to %s\n", node)
	// }

}
