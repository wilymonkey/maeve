package myerr

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/style"
	"path/filepath"
	"runtime"
)

// Wraps the error with file and codeline location.
func WrapErr(err error) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: %w", filename, line, err)
}

type ErrMsg struct {
	err error
}

func TuiMsg(e error) ErrMsg {
	return ErrMsg{err: e}
}

func (e *ErrMsg) Print() tea.Cmd {
	fail := style.Fail.Render("Failed!")
	return tea.Printf("%s %s\n%v", style.ICross, fail, e.err)
}

func BoolView(b bool) string {
	if b {
		return style.ITick
	}
	return style.ICross
}
