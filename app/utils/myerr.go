package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wilymonkey/maeve/style"
)

func ErrContext(context string, err error) error {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	fnName := filepath.Base(fn.Name())
	return fmt.Errorf("[%s]: %s →  %w", fnName, context, err)
}

// Wraps the error with file and codeline location.
func WrapErr(err error) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: %w", filename, line, err)
}

func WrapErrWithInfo(err error, info string) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: (%s) %w", filename, line, info, err)
}

func DummyErr(s string) error {
	return fmt.Errorf("DummyErr: %s", s)
}

func Sleep(t time.Duration) {
	time.Sleep(time.Millisecond * t)
}

func PrintErr(err error) string {
	fail := style.Fail.Render("Failed!")
	s := fmt.Sprintf("%s %s %v", style.ICross, fail, err)
	return style.Wrap(s)
}

func BoolView(b bool) string {
	if b {
		return style.ITick
	}
	return style.ICross
}
