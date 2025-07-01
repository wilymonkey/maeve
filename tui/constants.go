package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/wilymonkey/maeve/tui/color"
)

const VERSION = "0.1.0"

var (
	successStyle = lipgloss.NewStyle().Foreground(color.SpringGreen)
	spinnerStyle = lipgloss.NewStyle().Foreground(color.Sky600)
	helpStyle    = lipgloss.NewStyle().Foreground(color.Gray)
	errorStyle   = lipgloss.NewStyle().Foreground(color.Red)
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(color.Pink400)
	// Preset icons
	checkMark = successStyle.Margin(0, 1).Render("✔")
	crossMark = errorStyle.Render("✘")
	logo      = lipgloss.NewStyle().Foreground(color.Sky700).Render(`
███╗   ███╗ █████╗ ███████╗██╗   ██╗███████╗
████╗ ████║██╔══██╗██╔════╝██║   ██║██╔════╝
██╔████╔██║███████║█████╗  ██║   ██║█████╗  
██║╚██╔╝██║██╔══██║██╔══╝  ╚██╗ ██╔╝██╔══╝  
██║ ╚═╝ ██║██║  ██║███████╗ ╚████╔╝ ███████╗
╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝  ╚═══╝  ╚══════╝
`)
)
