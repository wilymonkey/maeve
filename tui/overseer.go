package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/utils"
)

type overseerBackMsg struct{}

func overseerBack() tea.Msg {
	return overseerBackMsg{}
}

type overseerModel struct {
	prev    tea.Model
	current tea.Model
}

func (m overseerModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m overseerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case overseerBackMsg:
		if m.prev != nil {
			m.current = m.prev
			m.prev = nil
			return m, nil
		}
		return m, tea.Quit

	default:
		updated, cmd := m.current.Update(msg)
		m.current = updated
		return m, cmd
	}
}

func (m overseerModel) View() string {
	return m.current.View()
}

func AsBackup() error {
	return startTui(newBackupModel())
}

func AsHome() error {
	return startTui(newHomeModel())
}

func startTui(current tea.Model) error {
	p := tea.NewProgram(overseerModel{
		current: current,
	})
	if _, err := p.Run(); err != nil {
		return utils.PrintErr(err)
	}
	return nil
}
