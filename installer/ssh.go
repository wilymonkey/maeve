package main

import (
	"os/exec"
	"strings"

	"github.com/wilymonkey/maeve/installer/utils"
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
