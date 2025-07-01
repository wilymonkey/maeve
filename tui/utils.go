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

type boolProgress int

const (
	Processing boolProgress = iota
	True
	False
)

func (b boolProgress) isDone() bool {
	return b != Processing
}

// Ensure you check isDone before accessing this value.
func (b boolProgress) value() bool {
	return b == True
}

func (b boolProgress) View(spinView string) string {
	if b.isDone() {
		if b.value() {
			return checkMark
		}
		return crossMark
	}
	return spinView
}

func boolView(b bool) string {
	if b {
		return checkMark
	}
	return crossMark
}
