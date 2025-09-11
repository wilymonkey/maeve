package utils

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

// Logs error if closer fails.
func Cleanup(err *error, closer func() error) {
	if cErr := closer(); cErr != nil {
		if *err == nil {
			*err = cErr
		}
	}
}

// Wraps the error with file and codeline location.
func WrapErr(err error) error {
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)

	return fmt.Errorf("%s@%d: %w", filename, line, err)
}

func DummyErr(s string) error {
	return fmt.Errorf("DummyErr: %s", s)
}

func Sleep(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))
}

func Assert(truth bool, msg string) {
	if !truth {
		log.Fatal(msg)
	}
}

func Assertf(truth bool, msg string, a ...any) {
	if !truth {
		log.Fatalf(msg, a...)
	}
}

func AssertNoErr(err error, msg string) {
	if err != nil {
		log.Fatalf("While: %s, Err: %v", msg, err)
	}
}
