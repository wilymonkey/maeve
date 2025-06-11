package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Pulls changes from Cfg.SourceDirs to the Cfg.BackupDir.
func LocalPull() error {
	err := os.MkdirAll(Cfg.BackupDir, 0755)
	if err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	for _, srcDir := range Cfg.SourceDirs {
		destDir := filepath.Join(NodeDirLatest(Cfg.Name), filepath.Base(srcDir))
		err := SnapshotCreate(srcDir, destDir)
		if err != nil {
			return fmt.Errorf("cloning %s to %s: %w", srcDir, destDir, err)
		}
	}

	return nil
}

// Pushes the local backup to a given node.
func LocalPush(node string) error {
	// Ensure remote target directory
	remoteTarget := fmt.Sprintf("%s:%s", node, Cfg.BackupDir)
	sshCmd := exec.Command("ssh", node, "mkdir", "-p", Cfg.BackupDir)
	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating remote target directory on %s: %w, output: %s", node, err, string(output))
	}

	// Sync using rsync
	rsyncCmd := exec.Command("rsync", "-avz", "--delete", "-e", "ssh", Cfg.BackupDir+"/", remoteTarget)
	output, err := rsyncCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync to %s: %w, output: %s", node, err, string(output))
	}

	// Invoke the program remotely in snapshot mode
	remoteCmd := fmt.Sprintf("cd %s && %s --config %s --snapshot", Cfg.BackupDir, os.Args[0], filepath.Base(*flag.String("config", "config.json", "")))
	sshSnapshotCmd := exec.Command("ssh", node, remoteCmd)
	output, err = sshSnapshotCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("creating snapshot on %s: %w, output: %s", node, err, string(output))
	}

	return nil
}
