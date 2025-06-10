package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/wilymonkey/maeve/internal/config"
	"github.com/wilymonkey/maeve/internal/snapshot"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	snapshotNode := flag.String("snapshot", "", "Create a hardlink based snapshot of a given node")
	flag.Parse()

	err := config.Read(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	if *snapshotNode != "" {
		err = snapshot.Node(*snapshotNode)
		if err != nil {
			log.Fatalf("Failed to create snapshot: %v", err)
		}
		fmt.Println("Snapshot created successfully")
		return
	}

	err = cloneLocal()
	if err != nil {
		log.Fatalf("Failed to clone directories: %v", err)
	}
	fmt.Println("Local directories cloned with hardlinks")

	for _, node := range config.RemoteNodes {
		err = syncAndSnapshot(node, config)
		if err != nil {
			log.Printf("Failed to sync and snapshot to %s: %v", node, err)
			continue
		}
		fmt.Printf("Successfully synced and snapshotted to %s\n", node)
	}
}
