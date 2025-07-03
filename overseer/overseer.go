package overseer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/style"
)

type msgPush struct {
	next tea.Model
}

func Push(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return msgPush{next: m}
	}
}

type msgBack struct{}

func Back() tea.Msg {
	return msgBack{}
}

type Model struct {
	prev    tea.Model
	current tea.Model
}

func New(model tea.Model) Model {
	return Model{
		current: model,
	}
}

func (m Model) Init() tea.Cmd {
	return m.current.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case msgBack:
		if m.prev != nil {
			m.current = m.prev
			m.prev = nil
			return m, nil
		}
		return m, tea.Quit

	case msgPush:
		m.prev = m.current
		m.current = msg.next
		return m, m.current.Init()

	default:
		updated, cmd := m.current.Update(msg)
		m.current = updated
		return m, cmd
	}
}

func (m Model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n", style.ILogo, m.current.View())
	return b.String()
}
