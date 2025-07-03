package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/tui/backup"
	"github.com/wilymonkey/maeve/tui/home"
	"github.com/wilymonkey/maeve/tui/overseer"
	"github.com/wilymonkey/maeve/utils"
)

func AsBackup() error {
	return startTui(backup.New())
}

func AsHome() error {
	return startTui(home.New())
}

func startTui(start tea.Model) error {
	p := tea.NewProgram(overseer.New(start), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return utils.PrintErr(err)
	}
	return nil
}
