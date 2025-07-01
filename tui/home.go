package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
)

type homeModel struct {
	spinner    spinner.Model
	nodeStatus map[string]nodeStatus
}

func newHomeModel() homeModel {
	s := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(spinnerStyle),
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
	return hm.spinner.Tick
}

func (hm homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return hm, overseerBack
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		hm.spinner, cmd = hm.spinner.Update(msg)
		return hm, cmd
	}
	return hm, nil
}

func (hm homeModel) View() string {
	var b strings.Builder
	spinView := hm.spinner.View()
	b.WriteString(titleStyle.Render("REMOTE NODES\n"))
	for _, ns := range hm.nodeStatus {
		ns.View(&b, spinView)
	}
	return b.String()
}
