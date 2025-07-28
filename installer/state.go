package main

import (
	"fyne.io/fyne/v2/data/binding"
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
	sshKeys, err := readSSHKeys()
	if err != nil {
		s.LogErr(err)
	}
	s.sshKeys.Set(sshKeys)
	if hasSSH && sshRunning {
		s.percSSH.Set(1)
	}
}

func (s *State) Install() error {
	percSSH, err := s.percSSH.Get()
	if err != nil {
		return utils.WrapErr(err)
	}
	if percSSH != 1 {

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
