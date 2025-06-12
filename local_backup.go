package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Pulls changes from Cfg.SourceDirs.
func LocalPull() error {
	localPath := NodeDirLatest(Cfg.Name)
	err := os.MkdirAll(localPath, 0755)
	if err != nil {
		return fmt.Errorf("creating backup directory ⇒  %w", err)
	}

	for _, srcDir := range Cfg.SourceDirs {
		destDir := filepath.Join(localPath, filepath.Base(srcDir))
		err := NewSnapshot(srcDir, destDir)
		if err != nil {
			return fmt.Errorf("cloning %s to %s ⇒  %w", srcDir, destDir, err)
		}
	}

	return nil
}

// Pushes changes to a given node address.
func LocalPush(nodeAddress string) error {
	remotePath, err := RunSSH(nodeAddress, "maeve -node-path "+Cfg.Name)
	if err != nil {
		return fmt.Errorf("unable to get remotePath ⇒  %w", err)
	}

	if err := RunRsync(nodeAddress, remotePath); err != nil {
		return fmt.Errorf("rsync failed ⇒  %w", err)
	}

	_, err = RunSSH(nodeAddress, "maeve -snapshot "+Cfg.Name)
	if err != nil {
		return fmt.Errorf("unable to get remotePath ⇒  %w", err)
	}

	return nil
}
