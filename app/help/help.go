package help

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
)

type Help int

const (
	AddKey Help = iota
	UpdateMaeve
	DelConfig
	DelKnownHost
	DelBackupDir
	DelPrivateKey
	badNodeConn
	delDB
	checkSource
	checkBackupDir
	checkNodeConn
	devReport
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
	case DelPrivateKey:
		return "delete the private key"
	case delDB:
		return "delete the .db file in the backup folder"
	case badNodeConn:
		return "the connection with the backup PC seems to be bad"
	case checkSource:
		return "check if I have permission to access the backup source files/folders"
	case checkBackupDir:
		return "check if the backup folder is one I have access to"
	case checkNodeConn:
		return "check if the backup pc (name, ip and port) is correct"
	case devReport:
		return "report it to the developer"
	default:
		return "unable to parse help text"
	}
}

func DelDB(err error, task string) error {
	return newHelpError(err, task, delDB)
}

func BadNodeConn(err error, task string) error {
	return newHelpError(err, task, badNodeConn)
}

func CheckSource(err error, task string) error {
	return newHelpError(err, task, checkSource)
}

func CheckBackupDir(err error, task string) error {
	return newHelpError(err, task, checkBackupDir)
}

func CheckNodeConn(err error, task string) error {
	return newHelpError(err, task, checkNodeConn)
}

func DevReport(err error, task string) error {
	return newHelpError(err, task, devReport)
}

func Stacktrace(err error, task string, help Help) error {
	return newHelpError(err, task, help)
}

func getStacktrace(start, number int) string {
	var b strings.Builder
	for i := start + number; i > start; i-- {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			continue
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		// Extract just the function name (without full package path).
		fmt.Fprintf(&b, "%s: ", filepath.Base(fn.Name()))
	}
	utils.Assert("stacktrace shouldn't be empty", b.Len() != 0)
	result := b.String()
	return result[:b.Len()-2]
}

type helpError struct {
	help  Help
	task  string
	err   error
	stack string
}

func (e *helpError) Error() string {
	return fmt.Sprintf(
		"%d||%s||%s||%s",
		e.help,
		e.task,
		e.err.Error(),
		e.stack,
	)
}

func newHelpError(err error, task string, help Help) error {
	return &helpError{
		help:  help,
		task:  task,
		err:   err,
		stack: getStacktrace(2, 4),
	}
}

// Reconstructs a Help Error from the Error() representation.
func FromErr(err error) error {
	parts := strings.Split(err.Error(), "||")
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return DevReport(err, "splitting error into parts")
	}
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return DevReport(err, "parsing help iota")
	}
	return &helpError{
		help:  Help(i),
		task:  parts[1],
		err:   errors.New(parts[2]),
		stack: parts[3],
	}
}

func Render(err error) fyne.CanvasObject {
	var helpErr *helpError
	if errors.As(err, &helpErr) {
		errLbl := widget.NewLabel(helpErr.err.Error())
		errLbl.Wrapping = fyne.TextWrapWord
		sl := widget.NewLabel(helpErr.stack)
		sl.Wrapping = fyne.TextWrapWord

		content := fynext.VBox(
			fynext.SmallTxt("Possible Fix"),
			widget.NewLabel(helpErr.help.string()),
			fynext.SmallTxt("Task Attempted"),
			widget.NewLabel(helpErr.task),
			fynext.SmallTxt("Error"),
			errLbl,
			fynext.SmallTxt("Stacktrace"),
			sl,
		)
		return content
	}

	return widget.NewLabel(err.Error())

}
