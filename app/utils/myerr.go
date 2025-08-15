package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wilymonkey/maeve/help"
	"github.com/wilymonkey/maeve/style"
)

// Wraps the error with file and codeline location.
func WrapErr(err error) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: %w", filename, line, err)
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

func Assert(reason string, pred bool) {
	if !pred {
		var b strings.Builder
		help.WriteStacktrace(&b)
		fmt.Fprintf(&b, ": %s", reason)
		panic(b.String())
	}
}

func AssertNoErr(reason string, err error) {
	if err != nil {
		var b strings.Builder
		help.WriteStacktrace(&b)
		fmt.Fprintf(&b, "\n%s: %v", reason, err)
		panic(b.String())
	}
}
