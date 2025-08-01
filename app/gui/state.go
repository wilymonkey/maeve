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
	var c conf.MaeveConf
	if err := conf.LoadConfig(&c); err != nil {
		return utils.WrapErr(err)
	}

	s.Name.Set(c.Name)
	s.Name.AddListener(binding.NewDataListener(func() {
		name := utils.GetOrPanic(s.Name)
		conf.Global.Name = name
		if err := conf.Global.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.BackupDir.Set(c.BackupDir)
	s.BackupDir.AddListener(binding.NewDataListener(func() {
		backupDir := utils.GetOrPanic(s.BackupDir)
		conf.Global.BackupDir = backupDir
		if err := conf.Global.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.MaxBackups.Set(c.MaxBackups)
	s.MaxBackups.AddListener(binding.NewDataListener(func() {
		maxBackups := utils.GetOrPanic(s.MaxBackups)
		conf.Global.MaxBackups = maxBackups
		if err := conf.Global.SaveToFile(); err != nil {
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
		conf.Global.MaxUpload = maxUpload
		if err := conf.Global.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.RemoteNodes.Set(c.RemoteNodes)
	s.RemoteNodes.AddListener(binding.NewDataListener(func() {
		remoteNodes := utils.GetOrPanic(s.RemoteNodes)
		conf.Global.RemoteNodes = remoteNodes
		if err := conf.Global.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.SourceDirs.Set(c.SourceDirs)
	s.SourceDirs.AddListener(binding.NewDataListener(func() {
		sourceDirs := utils.GetOrPanic(s.SourceDirs)
		conf.Global.SourceDirs = sourceDirs
		if err := conf.Global.SaveToFile(); err != nil {
			s.ShowError(err)
		}
	}))

	s.Window = w

	return nil
}

func (s State) ShowError(err error) {
	dialog.ShowError(err, s.Window)
}
