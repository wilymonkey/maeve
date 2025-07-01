package style

import (
	"github.com/charmbracelet/lipgloss"
)

const VERSION = "0.1.0"

const red700 = lipgloss.Color("#b91c1c")
const red400 = lipgloss.Color("#f87171")
const green600 = lipgloss.Color("#16a34a")
const pink400 = lipgloss.Color("#e879f9")
const sky600 = lipgloss.Color("#0284c7")
const sky700 = lipgloss.Color("#0369a1")
const zinc200 = lipgloss.Color("#e4e4e7")
const zinc500 = lipgloss.Color("#71717a")

var (
	Success = lipgloss.NewStyle().Foreground(green600)
	Spinner = lipgloss.NewStyle().Foreground(sky600)
	Help    = lipgloss.NewStyle().Foreground(zinc500)
	Fail    = lipgloss.NewStyle().Foreground(red400)
	Title   = lipgloss.NewStyle().Margin(1, 0).Bold(true).Foreground(pink400)
	Bright  = lipgloss.NewStyle().Foreground(zinc200)

	// ICONS
	ITick  = Success.Margin(0, 1).Render("✔")
	ICross = Fail.Margin(0, 1).Render("✘")
	ILogo  = lipgloss.NewStyle().Foreground(sky700).Render(`
███╗   ███╗ █████╗ ███████╗██╗   ██╗███████╗
████╗ ████║██╔══██╗██╔════╝██║   ██║██╔════╝
██╔████╔██║███████║█████╗  ██║   ██║█████╗  
██║╚██╔╝██║██╔══██║██╔══╝  ╚██╗ ██╔╝██╔══╝  
██║ ╚═╝ ██║██║  ██║███████╗ ╚████╔╝ ███████╗
╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝  ╚═══╝  ╚══════╝
`)
)
