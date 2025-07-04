package myerr

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wilymonkey/maeve/style"
)

// Wraps the error with file and codeline location.
func WrapErr(err error) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: %w", filename, line, err)
}

func Print(err error, b *strings.Builder) {
	fail := style.Fail.Render("Failed!")
	fmt.Fprintf(b, "\n\n%s %s %v\n\n", style.ICross, fail, err)
}

func BoolView(b bool) string {
	if b {
		return style.ITick
	}
	return style.ICross
}
