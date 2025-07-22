package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
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

func DummyErr(s string) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: DummyErr: %s", filename, line, s)
}

func PrintErr(err error) string {
	return fmt.Sprintf("Error: %v", err)
}

func Sleep(t time.Duration) {
	time.Sleep(time.Millisecond * t)
}
