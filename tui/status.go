package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner  spinner.Model
	status   string
	quitting bool
}

var (
	logoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
)

func newModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	return model{
		spinner: s,
		status:  "Status not implemented...press q to quit",
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Exiting...\n"
	}

	logo := logoStyle.Render(strings.Join([]string{
		"███╗░░░███╗░█████╗░███████╗██╗░░░██╗███████╗",
		"████╗░████║██╔══██╗██╔════╝██║░░░██║██╔════╝",
		"██╔████╔██║███████║█████╗░░╚██╗░██╔╝█████╗░░",
		"██║╚██╔╝██║██╔══██║██╔══╝░░░╚████╔╝░██╔══╝░░",
		"██║░╚═╝░██║██║░░██║███████╗░░╚██╔╝░░███████╗",
		"╚═╝░░░░░╚═╝╚═╝░░╚═╝╚══════╝░░░╚═╝░░░╚══════╝",
	}, "\n"))

	return fmt.Sprintf(
		"%s\n\n%s %s\n",
		logo,
		m.spinner.View(),
		statusStyle.Render(m.status),
	)
}

func Status() error {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
