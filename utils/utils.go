package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func PrintErr(err error) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: %w", filename, line, err)
}
