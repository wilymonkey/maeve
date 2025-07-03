package home

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/backup"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/style"
)

type Model struct {
	spinner    spinner.Model
	nodeStatus map[string]nodeStatus
}

func New() Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(style.Spinner),
	)
	ns := make(map[string]nodeStatus)
	for _, node := range cfg.Global.RemoteNodes {
		ns[node] = nodeStatus{name: node}
	}
	return Model{
		spinner:    s,
		nodeStatus: ns,
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.spinner.Tick)
	for _, node := range m.nodeStatus {
		cmds = append(cmds, node.FetchState)
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, overseer.Back
		case "b":
			return m, overseer.Push(backup.New())
		}
	case nodeStatus:
		m.nodeStatus[msg.name] = msg
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Version: %s", cfg.VERSION)
	b.WriteString("\n")
	spinView := m.spinner.View()
	b.WriteString(style.Title.Render("REMOTE NODES"))
	b.WriteString("\n")
	for _, ns := range m.nodeStatus {
		ns.View(&b, spinView)
		if ns.err != nil {
			title := style.Fail.Render("Error:")
			fmt.Fprintf(&b, "%s %v\n", title, ns.err)
		}
	}
	b.WriteString(style.Title.Render("COMMANDS"))
	b.WriteString("\n")
	b.WriteString(style.Help.Render("Press b to run backup"))
	b.WriteString("\n")
	b.WriteString(style.Help.Render("Press q to quit"))
	return b.String()
}
