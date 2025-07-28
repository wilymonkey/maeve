package utils

import "os/exec"

func RunPS(cmd string) error {
	_, err := exec.Command("powershell", "-Command", cmd).CombinedOutput()
	return err
}
