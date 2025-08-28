package help

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
)

type helpTextKey int

const (
	DevReport helpTextKey = iota
)

var helpText = map[helpTextKey]string{
	DevReport: "report this error to the developer",
}

var helpRegistry = []struct {
	matcher func(error) bool
	helpKey helpTextKey
}{}

func (c helpTextKey) String() string {
	if s, ok := helpText[c]; ok {
		return s
	}
	return helpText[DevReport]
}

type helpError struct {
	Help  helpTextKey
	Task  string
	Err   string
	Stack string
}

func newHelpError(err, task string, help helpTextKey) error {
	return &helpError{
		Help:  help,
		Task:  task,
		Err:   err,
		Stack: newStacktrace(2, 4),
	}
}

func WrapError(err error, task string) error {
	h := DevReport

	for _, entry := range helpRegistry {
		if entry.matcher(err) {
			h = entry.helpKey
		}
	}

	if h == DevReport {
		err = fmt.Errorf("type: %T\nerr: %v", err, err)
	}
	return newHelpError(err.Error(), task, h)
}

func (e *helpError) Error() string {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(e); err != nil {
		return err.Error()
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func DecodeErr(err error) error {
	data, decodeErr := base64.StdEncoding.DecodeString(err.Error())
	if decodeErr != nil {
		return fmt.Errorf("not an encoded helpError: type: %T err: %v", err, err)
	}

	b := bytes.NewBuffer(data)
	var herr helpError
	dec := gob.NewDecoder(b)
	if decodeErr := dec.Decode(&herr); decodeErr != nil {
		return fmt.Errorf("failed to decode error: %v", decodeErr)
	}
	return &herr
}

func Render(err error) fyne.CanvasObject {
	var helpErr *helpError

	if errors.As(err, &helpErr) {
		errLbl := widget.NewLabel(helpErr.Err)
		errLbl.Wrapping = fyne.TextWrapWord
		sl := widget.NewLabel(helpErr.Stack)
		sl.Wrapping = fyne.TextWrapWord

		content := fynext.VBox(
			fynext.SmallTxt("Possible Fix"),
			widget.NewLabel(helpErr.Help.String()),
			fynext.SmallTxt("Task Attempted"),
			widget.NewLabel(helpErr.Task),
			fynext.SmallTxt("Error"),
			errLbl,
			fynext.SmallTxt("Stacktrace"),
			sl,
		)
		return content
	}

	lbl := widget.NewLabel(err.Error())
	lbl.Wrapping = fyne.TextWrapWord
	return fynext.StackOfSize(fyne.NewSize(250, 1), lbl)
}

func newStacktrace(start, number int) string {
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
		fmt.Fprintf(&b, "%s\n", filepath.Base(fn.Name()))
	}

	utils.Assert(b.Len() != 0, "stacktrace shouldn't be empty")

	result := b.String()
	return result[:b.Len()-1] // Trim trailing newline.
}
