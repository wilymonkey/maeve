package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/shared"
	"github.com/wilymonkey/maeve/tui/style"
)

type homeModel struct {
	spinner    spinner.Model
	nodeStatus map[string]nodeStatus
}

func newHomeModel() homeModel {
	s := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(style.Spinner),
	)
	ns := make(map[string]nodeStatus)
	for _, node := range cfg.Global.RemoteNodes {
		ns[node] = nodeStatus{name: node}
	}
	return homeModel{
		spinner:    s,
		nodeStatus: ns,
	}
}

func (hm homeModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, hm.spinner.Tick)
	for _, id := range hm.nodeStatus {
		cmds = append(cmds, id.FetchState)
	}
	return tea.Batch(cmds...)
}

func (hm homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return hm, overseerBack
		case "b":
			return hm, overseerPush(newBackupModel())
		}
	case nodeStatus:
		hm.nodeStatus[msg.name] = msg
		return hm, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		hm.spinner, cmd = hm.spinner.Update(msg)
		return hm, cmd
	}
	return hm, nil
}

func (hm homeModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Version: %s", shared.VERSION)
	b.WriteString("\n")
	spinView := hm.spinner.View()
	b.WriteString(style.Title.Render("REMOTE NODES"))
	b.WriteString("\n")
	for _, ns := range hm.nodeStatus {
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
