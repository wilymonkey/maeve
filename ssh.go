package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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

func RunRsync(address, path string) error {
	host, port := splitHostPort(address)
	cmd := exec.Command("rsync", "-avzP", "--delete", "-e", "ssh -p "+port, NodeDirLatest(Cfg.Name), host+":"+path)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync command failed ⇒  %v, stderr: %s", err, stderr.String())
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
