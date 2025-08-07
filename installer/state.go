package main

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"github.com/wilymonkey/maeve/installer/utils"
)

var Global State

type State struct {
	logs          binding.StringList
	logChan       chan string
	percSSH       binding.Float
	percMaeve     binding.Float
	percIntegrity binding.Float
	sshKeys       binding.StringList
	isServer      binding.Bool
	CurrentWindow fyne.Window
}

func NewState() State {
	s := State{
		logs:          binding.NewStringList(),
		logChan:       make(chan string),
		percSSH:       binding.NewFloat(),
		percMaeve:     binding.NewFloat(),
		percIntegrity: binding.NewFloat(),
		sshKeys:       binding.NewStringList(),
		isServer:      binding.NewBool(),
	}

	go func() {
		for log := range s.logChan {
			s.logs.Append(log)
		}
	}()

	return s
}

func (s *State) ShowError(err error) {
	dialog.ShowError(err, s.CurrentWindow)
}

func (s *State) Close() {
	close(s.logChan)
}

// Ensures the key add is unique.
func (s *State) AddKey(key string) {
	sshKeys, err := s.sshKeys.Get()
	if err != nil {
		panic(err)
	}
	if !slices.Contains(sshKeys, key) {
		s.sshKeys.Append(key)
	}
}

func (s *State) Install() error {
	percSSH, err := s.percSSH.Get()
	if err != nil {
		return utils.WrapErr(err)
	}
	if percSSH != 1 {
		err := InstallSSH()
		if err != nil {
			s.percSSH.Set(0)
			return utils.WrapErr(err)
		}
	}

	percMaeve, err := s.percMaeve.Get()
	if err != nil {
		return utils.WrapErr(err)
	}
	if percMaeve != 1 {

	}

	percIntegrity, err := s.percIntegrity.Get()
	if err != nil {
		return utils.WrapErr(err)
	}
	if percIntegrity != 1 {

	}

	return nil
}
