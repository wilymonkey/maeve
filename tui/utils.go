package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/tui/style"
)

type errorMsg struct {
	err error
}

func newErrorMsg(e error) errorMsg {
	return errorMsg{err: e}
}

func (e *errorMsg) print() tea.Cmd {
	fail := style.Fail.Render("Failed!")
	return tea.Printf("%s %s\n%v", style.ICross, fail, e.err)
}

func boolView(b bool) string {
	if b {
		return style.ITick
	}
	return style.ICross
}
