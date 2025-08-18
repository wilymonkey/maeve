package utils

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

func DummyErr(s string) error {
	return fmt.Errorf("DummyErr: %s", s)
}

func Sleep(t time.Duration) {
	time.Sleep(time.Millisecond * t)
}

// If the world state has been violated, panic the program.
func Assert(reason string, pred bool) {
	if !pred {
		panic(reason)
	}
}

// If there is an error, panic and crash the program.
func AssertNoErr(reason string, err error) {
	if err != nil {
		panic(reason)
	}
}
