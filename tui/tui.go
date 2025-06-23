package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type tuiSwitchToStatus struct{}

type TuiModel struct {
	current tea.Model
	state   string
}

func (m TuiModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tuiSwitchToStatus:
		return m, m.current.Init()

	default:
		updated, cmd := m.current.Update(msg)
		m.current = updated
		return m, cmd
	}
}

func (m TuiModel) View() string {
	return m.current.View()
}
