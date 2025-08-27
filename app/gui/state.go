package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
)

var global State

type State struct {
	Name        binding.String
	MaeveDir    binding.String
	MaxBackups  binding.Int
	MaxUpload   binding.String
	RemoteNodes binding.StringList
	SourceDirs  binding.StringList
	Window      fyne.Window
}

func LoadState(window fyne.Window) {
	cfg := conf.GetConf()

	state := State{
		Name:        binding.BindString(&cfg.DisplayName),
		MaeveDir:    binding.BindString(&cfg.MaeveDir),
		MaxBackups:  binding.BindInt(&cfg.MaxBackups),
		MaxUpload:   binding.NewString(),
		RemoteNodes: binding.BindStringList(&cfg.RemoteNodes),
		SourceDirs:  binding.BindStringList(&cfg.SourceDirs),
		Window:      window,
	}

	state.MaxUpload.Set(utils.BytesToHuman(cfg.MaxUpload))
	state.MaxUpload.AddListener(binding.NewDataListener(func() {
		val := fynext.Unwrap(state.MaxUpload)
		max, err := utils.ParseHumanBytes(val)
		if err != nil {
			state.ShowError(err)
		}
		cfg.MaxUpload = max
		state.MaxUpload.Set(utils.BytesToHuman(cfg.MaxUpload))
	}))

	saveOnChange(state.Name, state.MaeveDir, state.MaxUpload)
	saveOnChange(state.MaxBackups)
	saveOnChange(state.RemoteNodes, state.SourceDirs)

	global = state
}

func saveOnChange[T any](sources ...binding.Item[T]) {
	cfg := conf.GetConf()
	for _, s := range sources {
		s.AddListener(binding.NewDataListener(func() {
			if err := cfg.SaveToFile(); err != nil {
				global.ShowError(err)
			}
		}))
	}
}

func (s *State) ShowError(err error) {
	dialog.ShowError(err, s.Window)
}
