package local_backup

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wilymonkey/maeve/internal/config"
	"github.com/wilymonkey/maeve/internal/snapshot"
)

// Clones config.SourceDirs to the config.BackupDir under the config.Name node.
func Pull() error {
	err := os.MkdirAll(config.Global.BackupDir, 0755)
	if err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	for _, srcDir := range config.Global.SourceDirs {
		destDir := filepath.Join(config.NodeDirLatest(config.Global.Name), filepath.Base(srcDir))
		err := snapshot.Create(srcDir, destDir)
		if err != nil {
			return fmt.Errorf("cloning %s to %s: %w", srcDir, destDir, err)
		}
	}

	return nil
}

func syncAndSnapshot(node string) error {
	// Ensure remote target directory
	remoteTarget := fmt.Sprintf("%s:%s", node, config.BackupDir)
	sshCmd := exec.Command("ssh", node, "mkdir", "-p", config.BackupDir)
	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating remote target directory on %s: %w, output: %s", node, err, string(output))
	}

	// Sync using rsync
	rsyncCmd := exec.Command("rsync", "-avz", "--delete", "-e", "ssh", config.BackupDir+"/", remoteTarget)
	output, err := rsyncCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync to %s: %w, output: %s", node, err, string(output))
	}

	// Invoke the program remotely in snapshot mode
	remoteCmd := fmt.Sprintf("cd %s && %s --config %s --snapshot", config.BackupDir, os.Args[0], filepath.Base(*flag.String("config", "config.json", "")))
	sshSnapshotCmd := exec.Command("ssh", node, remoteCmd)
	output, err = sshSnapshotCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("creating snapshot on %s: %w, output: %s", node, err, string(output))
	}

	return nil
}
