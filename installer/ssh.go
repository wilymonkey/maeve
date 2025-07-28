package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wilymonkey/maeve/installer/utils"
	"golang.org/x/crypto/ssh"
)

func checkSSH() (bool, error) {
	cmd := exec.Command("powershell.exe",
		"-NoProfile",
		"-Command",
		"Get-WindowsCapability -Online | Where-Object Name -like 'OpenSSH*'",
	)
	output, err := cmd.Output()
	if err != nil {
		return false, utils.WrapErr(err)
	}

	outputStr := string(output)
	hasClient := strings.Contains(outputStr, "Name : OpenSSH.Client")
	hasServer := strings.Contains(outputStr, "Name : OpenSSH.Server")
	isInstalled := strings.Contains(outputStr, "State : Installed")

	return hasClient && hasServer && isInstalled, nil
}

func checkSSHRunning() (bool, error) {
	cmd := exec.Command("powershell.exe",
		"-NoProfile",
		"-Command",
		"Get-Service -Name 'sshd'",
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			errOutput := string(exitErr.Stderr)
			if strings.Contains(errOutput, "find any service") {
				return false, nil
			}
		}
		return false, utils.WrapErr(err)
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "Running"), nil
}

// Reads Admin Keys from \ProgramData\ssh\administrators_authorized_keys.
func readSSHKeys() ([]string, error) {
	var keys []string

	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		return keys, utils.WrapErr(fmt.Errorf("Unable to get PROGRAMDATA env"))
	}

	filePath := filepath.Join(programData, "ssh", "administrators_authorized_keys")
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return keys, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			keys = append(keys, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, utils.WrapErr(err)
	}

	return keys, nil
}

func isValidKey(key string) bool {
	if len(key) < 1 {
		return false
	}
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		return false
	}
	return true
}
