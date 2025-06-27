package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type homeModel struct {
	spinner  spinner.Model
	status   string
	quitting bool
}

func newHomeModel() homeModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = spinnerStyle
	return homeModel{
		spinner: s,
		status:  "Status not implemented...press q to quit",
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
	if hm.quitting {
		return "Exiting...\n"
	}

	logoStyled := logoStyle.Render(logo)

	return fmt.Sprintf(
		"%s\n\n%s %s\n",
		logoStyled,
		hm.spinner.View(),
		logoStyle.Render(hm.status),
	)
}
