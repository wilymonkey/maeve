package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/wilymonkey/maeve/tui/color"
)

const logo = `
███╗░░░███╗░█████╗░███████╗██╗░░░██╗███████╗
████╗░████║██╔══██╗██╔════╝██║░░░██║██╔════╝
██╔████╔██║███████║█████╗░░╚██╗░██╔╝█████╗░░
██║╚██╔╝██║██╔══██║██╔══╝░░░╚████╔╝░██╔══╝░░
██║░╚═╝░██║██║░░██║███████╗░░╚██╔╝░░███████╗
╚═╝░░░░░╚═╝╚═╝░░╚═╝╚══════╝░░░╚═╝░░░╚══════╝
`

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(color.SpringGreen))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(color.RoyalBlue))
	logoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(color.RoyalBlue))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(color.Gray))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(color.Red))
	// Preset icons
	checkMark = successStyle.Margin(0, 1).Render("✔")
	crossMark = errorStyle.Render("✘")
)
