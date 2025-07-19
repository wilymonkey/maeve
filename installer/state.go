package main

import "fyne.io/fyne/v2/data/binding"

var Global State

type State struct {
	hasSSH     binding.Bool
	sshRunning binding.Bool
	hasMaeve   binding.Bool
	logs       binding.StringList
}

func NewState() State {
	return State{
		hasSSH:     binding.NewBool(),
		sshRunning: binding.NewBool(),
		hasMaeve:   binding.NewBool(),
		logs:       binding.NewStringList(),
	}
}
