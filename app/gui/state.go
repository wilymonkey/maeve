package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/utils"
)

var Global State

type State struct {
	Name        binding.String
	BackupDir   binding.String
	MaxBackups  binding.Int
	MaxUpload   binding.String
	RemoteNodes binding.StringList
	SourceDirs  binding.StringList
	Window      fyne.Window
}

func NewState() State {
	return State{
		Name:        binding.NewString(),
		BackupDir:   binding.NewString(),
		MaxBackups:  binding.NewInt(),
		MaxUpload:   binding.NewString(),
		RemoteNodes: binding.NewStringList(),
		SourceDirs:  binding.NewStringList(),
	}
}

func (s *State) Load(w fyne.Window) error {
	c := conf.GetConf()
	s.Name.Set(c.Name)
	s.Name.AddListener(binding.NewDataListener(func() {
		name := utils.GetOrPanic(s.Name)
		c.Name = name
		if err := c.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.BackupDir.Set(c.BackupDir)
	s.BackupDir.AddListener(binding.NewDataListener(func() {
		backupDir := utils.GetOrPanic(s.BackupDir)
		c.BackupDir = backupDir
		if err := c.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.MaxBackups.Set(c.MaxBackups)
	s.MaxBackups.AddListener(binding.NewDataListener(func() {
		maxBackups := utils.GetOrPanic(s.MaxBackups)
		c.MaxBackups = maxBackups
		if err := c.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.MaxUpload.Set(utils.BytesToHuman(c.MaxUpload))
	s.MaxUpload.AddListener(binding.NewDataListener(func() {
		maxUpload, err := utils.ParseHumanBytes(utils.GetOrPanic(Global.MaxUpload))
		if err != nil {
			Global.ShowError(err)
			return
		}
		c.MaxUpload = maxUpload
		if err := c.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.RemoteNodes.Set(c.RemoteNodes)
	s.RemoteNodes.AddListener(binding.NewDataListener(func() {
		remoteNodes := utils.GetOrPanic(s.RemoteNodes)
		c.RemoteNodes = remoteNodes
		if err := c.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.SourceDirs.Set(c.BackupDirs)
	s.SourceDirs.AddListener(binding.NewDataListener(func() {
		sourceDirs := utils.GetOrPanic(s.SourceDirs)
		c.BackupDirs = sourceDirs
		if err := c.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.Window = w

	return nil
}

func (s State) ShowError(err error) {
	dialog.ShowError(err, s.Window)
}
