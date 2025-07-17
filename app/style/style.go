package style

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	Reset   = "\x1b[0m"
	Success = lipgloss.NewStyle().Foreground(green600)
	Spinner = lipgloss.NewStyle().Foreground(sky600)
	Fade    = lipgloss.NewStyle().Foreground(zinc500)
	Fail    = lipgloss.NewStyle().Foreground(red400)
	Title   = lipgloss.NewStyle().Bold(true).Foreground(pink400)
	Bright  = lipgloss.NewStyle().Foreground(zinc200)

	// Merged
	TitlePending = Fade.Inherit(Title)
	TitleSuccess = Success.Inherit(Title)
	TitleFail    = Fail.Inherit(Title)

	// Margins
	My = lipgloss.NewStyle().Margin(1, 0)
	Mx = lipgloss.NewStyle().Margin(0, 1)

	// TABLES
	Table = table.Styles{
		Selected: lipgloss.NewStyle(),
		Header:   lipgloss.NewStyle().BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(pink400),
		Cell:     lipgloss.NewStyle(),
	}

	// ICONS
	ITick  = Success.Inherit(Mx).Render("✔")
	ICross = Fail.Inherit(Mx).Render("✘")
	ILogo  = lipgloss.NewStyle().Foreground(sky700).Render(`
███╗   ███╗ █████╗ ███████╗██╗   ██╗███████╗
████╗ ████║██╔══██╗██╔════╝██║   ██║██╔════╝
██╔████╔██║███████║█████╗  ██║   ██║█████╗  
██║╚██╔╝██║██╔══██║██╔══╝  ╚██╗ ██╔╝██╔══╝  
██║ ╚═╝ ██║██║  ██║███████╗ ╚████╔╝ ███████╗
╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝  ╚═══╝  ╚══════╝
`)
)

func Wrap(s string) string {
	return ansi.Wrap(s, 80, "")
}
