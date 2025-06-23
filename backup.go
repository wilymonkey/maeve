package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/hashsums"
	"github.com/wilymonkey/maeve/ssh"
)

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
//
// CAUTION: Deletes the Cfg.SelfDir directory.
func LocalPull() error {
	selfDir := cfg.Global.SelfDir()

	if err := os.RemoveAll(selfDir); err != nil {
		return fmt.Errorf("delete self dir ⇒  %w", err)
	}

	if err := os.MkdirAll(selfDir, 0755); err != nil {
		return fmt.Errorf("create self dir ⇒  %w", err)
	}

	for _, srcDir := range cfg.Global.SourceDirs {
		destDir := filepath.Join(selfDir, filepath.Base(srcDir))
		if err := HardlinkDir(srcDir, destDir); err != nil {
			return fmt.Errorf("hardlink dir %s to %s ⇒  %w", srcDir, destDir, err)
		}
	}

	if err := hashsums.NewDirFileHash(selfDir); err != nil {
		return fmt.Errorf("create hashsums for self dir ⇒  %w", err)
	}

	return nil
}

// Pushes changes to a given node address.
func LocalPush(address string) error {
	client, err := ssh.NewSSHClient(address)
	if err != nil {
		return fmt.Errorf("new ssh client ⇒  %w", err)
	}
	defer client.Close()

	// TODO: Don't know what's happening here.

	return nil
}

type BackupModel struct {
	Operation string
}

func initBackupModel() BackupModel {
	return BackupModel{}
}
