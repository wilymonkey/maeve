package main

import (
	"os/exec"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

var Global State

type State struct {
	hasSSH     binding.Bool
	sshRunning binding.Bool
	hasMaeve   binding.Bool
	logs       binding.StringList
	logChan    chan string
}

func NewState() State {
	s := State{
		hasSSH:     binding.NewBool(),
		sshRunning: binding.NewBool(),
		hasMaeve:   binding.NewBool(),
		logs:       binding.NewStringList(),
		logChan:    make(chan string),
	}

	go func() {
		for log := range s.logChan {
			s.logs.Append(log)
		}
	}()

	return s
}

func (s *State) Close() {
	close(s.logChan)
}

func (s *State) LogErr(err error) {
	s.logChan <- PrintErr(err)
}

func (s *State) GetCurrent() {
	hasSSH, err := checkSSH()
	if err != nil {
		s.LogErr(err)
	}
	s.hasSSH.Set(hasSSH)
	sshRunning, err := checkSSHRunning()
	if err != nil {
		s.LogErr(err)
	}
	s.sshRunning.Set(sshRunning)
}

func checkSSH() (bool, error) {
	cmd := exec.Command("powershell.exe",
		"-NoProfile",
		"-Command",
		"Get-WindowsCapability -Online | Where-Object Name -like 'OpenSSH*'",
	)
	output, err := cmd.Output()
	if err != nil {
		return false, WrapErr(err)
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
		return false, WrapErr(err)
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "Running"), nil
}
