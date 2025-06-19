package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
//
// CAUTION: Deletes the Cfg.SelfDir directory.
func LocalPull() error {
	localDir := Config.SelfDir()

	if err := os.RemoveAll(localDir); err != nil {
		return fmt.Errorf("deleting local latest directory ⇒  %w", err)
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("creating backup directory ⇒  %w", err)
	}

	for _, srcDir := range Config.SourceDirs {
		destDir := filepath.Join(localDir, filepath.Base(srcDir))
		if err := NewSnapshot(srcDir, destDir); err != nil {
			return fmt.Errorf("unable to clone %s to %s ⇒  %w", srcDir, destDir, err)
		}
	}

	hashes, err := NewSHA3Sums(localDir)
	if err != nil {
		return fmt.Errorf("unable to create hashsums for local files ⇒  %w", err)
	}
	if err := WriteFileHashes(hashes, localDir); err != nil {
		return fmt.Errorf("unable to write hashsums to file ⇒  %w", err)
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
