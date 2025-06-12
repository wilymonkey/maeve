package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunSSH(address string, command string) (string, error) {
	host, port := splitHostPort(address)
	cmd := exec.Command("ssh", "-p", port, host, command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ssh command failed ⇒  %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func RunSCP(address, path string) error {
	host, port := splitHostPort(address)
	// Quirk of the exec command of scp makes it nest "latest" folders.
	target := host + ":" + filepath.Dir(path)
	cmd := exec.Command("scp", "-r", "-P", port, NodeDirLatest(Cfg.Name), target)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("scp command failed ⇒  %v, stderr: %s", err, stderr.String())
	}

	return nil
}

func splitHostPort(nodeAddress string) (host string, port string) {
	parts := strings.Split(nodeAddress, ":")
	if len(parts) == 1 {
		return parts[0], "22"
	}
	return parts[0], parts[1]
}
