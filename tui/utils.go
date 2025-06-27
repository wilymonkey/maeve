package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type errorMsg struct {
	err error
}

func newErrorMsg(e error) errorMsg {
	return errorMsg{err: e}
}

func (e *errorMsg) print() tea.Cmd {
	fail := errorStyle.Render("Failed!")
	return tea.Printf("%s %s\n%v", crossMark, fail, e.err)
}
