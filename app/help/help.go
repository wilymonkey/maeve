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
	"github.com/wilymonkey/maeve/utils"
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
	case DelDB:
		return "delete the .db file in the backup folder"
	case DelPrivateKey:
		return "delete the private key"
	case checkBackupDir:
		return "check if the backup folder is one I have access to"
	case checkNodeConn:
		return "check if the backup pc (name, ip and port) is correct"
	case devReport:
		return "report it to the developer"
	default:
		return "...this shouldn't be possible"
	}
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
	before := b.Len()
	for i := start + number; i > start; i-- {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		// Extract just the function name (without full package path).
		fmt.Fprintf(&b, "%s: ", filepath.Base(fn.Name()))
	}
	after := b.Len()
	utils.Assert("stacktrace shouldn't be empty", before == after)
	result := b.String()
	return result[:b.Len()-2]
}

type helpError struct {
	help  Help
	task  string
	err   error
	stack string
}

func newHelpError(err error, task string, help Help) error {
	return &helpError{
		help:  help,
		task:  task,
		err:   err,
		stack: getStacktrace(2, 4),
	}
}

func (e *helpError) Error() string {
	return fmt.Sprintf("I was doing %q but got %q", e.task, e.err.Error())
}

type Widget struct {
	widget.BaseWidget
	bound binding.Item[error]
}

func NewWidget(err binding.Item[error]) *Widget {
	w := &Widget{bound: err}
	w.ExtendBaseWidget(w)
	w.bound.AddListener(binding.NewDataListener(func() {
		w.Refresh()
	}))
	return w
}

func (w *Widget) CreateRenderer() fyne.WidgetRenderer {
	val := fynext.Unwrap(w.bound)

	if val == nil {
		lbl := widget.NewLabel("")
		return widget.NewSimpleRenderer(lbl)
	}

	var hErr *helpError
	if errors.As(val, &hErr) {
		sl := widget.NewLabel(hErr.stack)
		sl.Wrapping = fyne.TextWrapWord

		content := fynext.VBox(
			fynext.SmallTxt("Possible Fix"),
			widget.NewLabel(hErr.help.string()),
			fynext.SmallTxt("Task Attempted"),
			widget.NewLabel(hErr.task),
			fynext.SmallTxt("Error"),
			widget.NewLabel(hErr.err.Error()),
			fynext.SmallTxt("Stacktrace"),
			sl,
		)
		return widget.NewSimpleRenderer(content)
	}

	return widget.NewSimpleRenderer(widget.NewLabel(val.Error()))
}
