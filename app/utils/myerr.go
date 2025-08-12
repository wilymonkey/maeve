package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wilymonkey/maeve/style"
)

func Stacktrace(err error, context string) error {
	var b strings.Builder
	buildStacktrace(&b)
	return fmt.Errorf("%s:\n%s: %w", b.String(), context, err)
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

func Assert(reason string, pred bool) {
	if !pred {
		var b strings.Builder
		buildStacktrace(&b)
		fmt.Fprintf(&b, ": %s", reason)
		panic(b.String())
	}
}

func PanicOnErr(reason string, err error) {
	if err != nil {
		var b strings.Builder
		buildStacktrace(&b)
		fmt.Fprintf(&b, "\n%s: %v", reason, err)
		panic(b.String())
	}
}

func buildStacktrace(b *strings.Builder) {
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
