package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/tui/style"
	"github.com/wilymonkey/maeve/utils"
)

type overseerMsgPush struct {
	next tea.Model
}

func overseerPush(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return overseerMsgPush{next: m}
	}
}

type overseerMsgBack struct{}

func overseerBack() tea.Msg {
	return overseerMsgBack{}
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

	case overseerMsgBack:
		if om.prev != nil {
			om.current = om.prev
			om.prev = nil
			return om, nil
		}
		return om, tea.Quit

	case overseerMsgPush:
		om.prev = om.current
		om.current = msg.next
		return om, nil

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
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n", style.ILogo, om.current.View())
	return b.String()
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
