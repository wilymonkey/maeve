package main

import (
	"fyne.io/fyne/v2/data/binding"
	"github.com/wilymonkey/maeve/installer/utils"
)

var Global State

type State struct {
	percSSH       binding.Float
	percMaeve     binding.Float
	percIntegrity binding.Float
	logs          binding.StringList
	logChan       chan string
}

func NewState() State {
	s := State{
		percSSH:       binding.NewFloat(),
		percMaeve:     binding.NewFloat(),
		percIntegrity: binding.NewFloat(),
		logs:          binding.NewStringList(),
		logChan:       make(chan string),
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
	s.logChan <- utils.PrintErr(err)
}

func (s *State) GetCurrent() {
	hasSSH, err := checkSSH()
	if err != nil {
		s.LogErr(err)
	}
	sshRunning, err := checkSSHRunning()
	if err != nil {
		s.LogErr(err)
	}
	if hasSSH && sshRunning {
		s.percSSH.Set(1)
	}
	s.percIntegrity.Set(0.4)
	s.percMaeve.Set(1)
}
