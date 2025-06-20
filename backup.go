package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
//
// CAUTION: Deletes the Cfg.SelfDir directory.
func LocalPull() error {
	selfDir := Config.SelfDir()

	if err := os.RemoveAll(selfDir); err != nil {
		return fmt.Errorf("delete self dir ⇒  %w", err)
	}

	if err := os.MkdirAll(selfDir, 0755); err != nil {
		return fmt.Errorf("create self dir ⇒  %w", err)
	}

	for _, srcDir := range Config.SourceDirs {
		destDir := filepath.Join(selfDir, filepath.Base(srcDir))
		if err := HardlinkDir(srcDir, destDir); err != nil {
			return fmt.Errorf("hardlink dir %s to %s ⇒  %w", srcDir, destDir, err)
		}
	}

	hashes, err := NewDirFileHash(selfDir)
	if err != nil {
		return fmt.Errorf("create hashsums for self dir⇒  %w", err)
	}
	if err := WriteFileHashes(hashes, selfDir); err != nil {
		return fmt.Errorf("write hashsums to file ⇒  %w", err)
	}

	return nil
}

// Pushes changes to a given node address.
func LocalPush(address string) error {
	client, err := NewSSHClient(address)
	if err != nil {
		return fmt.Errorf("new ssh client ⇒  %w", err)
	}
	defer client.Close()

	// TODO: Don't know what's happening here.

	return nil
}
