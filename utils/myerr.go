package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wilymonkey/maeve/style"
)

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

func DummyErr() error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: DummyErr", filename, line)
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
