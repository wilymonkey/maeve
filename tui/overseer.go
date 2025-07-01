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

func (om overseerModel) Init() tea.Cmd {
	return om.current.Init()
}

func (om overseerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case overseerBackMsg:
		if om.prev != nil {
			om.current = om.prev
			om.prev = nil
			return om, nil
		}
		return om, tea.Quit

	case errorMsg:
		return om, tea.Sequence(
			msg.print(),
			overseerBack,
		)

	default:
		updated, cmd := om.current.Update(msg)
		om.current = updated
		return om, cmd
	}
}

func (om overseerModel) View() string {
	return logo + "\n" + om.current.View()
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
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return utils.PrintErr(err)
	}
	return nil
}
