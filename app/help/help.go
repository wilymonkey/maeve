package help

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/fynext"
)

type Help int

const (
	AddKey Help = iota
	UpdateMaeve
	DelConfig
	DelKnownHost
	DelBackupDir
	DelDB
	DelPrivateKey
	CheckNodeConn
	devError
)

func (c Help) string() string {
	switch c {
	case AddKey:
		return "add the Maeve Key to the backup PC"
	case UpdateMaeve:
		return "update Maeve"
	case DelConfig:
		return "delete the user data folder for maeve"
	case DelKnownHost:
		return "the PC we had once connected to has changed, if (and only if) you are certain it's fine, delete the PC entry in \"sshknownkeys\""
	case DelBackupDir:
		return "delete the problematic backup folder"
	case DelDB:
		return "delete the .db file in the backup folder"
	case DelPrivateKey:
		return "delete the private key"
	case CheckNodeConn:
		return "check if the backup pc (name, ip and port) is correct"
	case devError:
		return "report it to the developer"
	default:
		return "...this shouldn't be possible"
	}
}

func DevError(err error, task string) error {
	return Stacktrace(err, task, devError)
}

func Stacktrace(err error, task string, help Help) error {
	var stack strings.Builder
	WriteStacktrace(&stack)
	return &helpError{
		help:  help,
		task:  task,
		err:   err,
		stack: stack.String(),
	}
}

func WriteStacktrace(b *strings.Builder) {
	b.WriteString("...")
	before := b.Len()
	for i := 4; i > 1; i-- {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			continue
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		// Extract just the function name (without full package path).
		fnName := filepath.Base(fn.Name())
		b.WriteString(" →  ")
		b.WriteString(fnName)
	}
	after := b.Len()
	if before == after {
		panic("unable to build stacktrace")
	}
}

type helpError struct {
	help  Help
	task  string
	err   error
	stack string
}

func (e *helpError) Error() string {
	return fmt.Sprintf("I was doing %q but got %q", e.task, e.err.Error())
}

func Widget(errBinding binding.Item[error]) fyne.CanvasObject {
	// Create a container that will update when the error changes
	container := widget.NewCard("", "", nil)

	// Function to update the container's content based on the current error
	updateContent := func(err error) {
		var hErr *helpError
		if errors.As(err, &hErr) {
			e := err.(*helpError)
			sl := widget.NewLabel(e.stack)
			sl.Wrapping = fyne.TextWrapWord
			content := fynext.VBox(
				fynext.SmallTxt("Possible Fix"),
				widget.NewLabel(e.help.string()),
				fynext.SmallTxt("Task Attempted"),
				widget.NewLabel(e.task),
				fynext.SmallTxt("Error"),
				widget.NewLabel(e.err.Error()),
				fynext.SmallTxt("Stacktrace"),
				sl,
			)
			container.SetContent(content)
		} else {
			container.SetContent(widget.NewLabel(err.Error()))
		}
	}

	if err := fynext.GetOrPanic(errBinding); err != nil {
		updateContent(err)
	}

	// Set up listener for changes
	errBinding.AddListener(binding.NewDataListener(func() {
		if err := fynext.GetOrPanic(errBinding); err != nil {
			updateContent(err)
		}
	}))

	return container
}
