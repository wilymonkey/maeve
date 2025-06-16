package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
//
// CAUTION: Deletes the Cfg.SelfDir directory.
func LocalPull() error {
	localPath := Config.SelfDir()

	if err := os.RemoveAll(localPath); err != nil {
		return fmt.Errorf("deleting local latest directory ⇒  %w", err)
	}

	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("creating backup directory ⇒  %w", err)
	}

	for _, srcDir := range Config.SourceDirs {
		destDir := filepath.Join(localPath, filepath.Base(srcDir))
		if err := NewSnapshot(srcDir, destDir); err != nil {
			return fmt.Errorf("unable to clone %s to %s ⇒  %w", srcDir, destDir, err)
		}
	}

	hashes, err := NewSHA3Sums(localPath)
	if err != nil {
		return fmt.Errorf("unable to create hashsums for local files ⇒  %w", err)
	}
	if err := WriteFileHashes(hashes, localPath); err != nil {
		return fmt.Errorf("unable to write hashsums to file ⇒  %w", err)
	}

	return nil
}

// Pushes changes to a given node address.
func LocalPush(nodeAddress string) error {
	remotePath, err := RunSSH(nodeAddress, "maeve --node-path "+Config.Name)
	if err != nil {
		return fmt.Errorf("unable to get remotePath ⇒  %w", err)
	}
	remotePath = strings.TrimSuffix(remotePath, "\n")

	if err := RunSCP(nodeAddress, remotePath); err != nil {
		return fmt.Errorf("rsync failed ⇒  %w", err)
	}

	_, err = RunSSH(nodeAddress, "maeve --snapshot "+Config.Name)
	if err != nil {
		return fmt.Errorf("unable to get remotePath ⇒  %w", err)
	}

	return nil
}
