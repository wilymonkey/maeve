package utils

import (
	"errors"
	"os"
	"path/filepath"
)

const PERMDIR = 0755

// Creates a file and parent directories if missing.
func Create(path string) (*os.File, error) {
	f, err := os.Create(path)
	if err == nil {
		return f, nil
	}
	if os.IsNotExist(err) {
		if mkErr := os.MkdirAll(filepath.Dir(path), os.ModePerm); mkErr != nil {
			return nil, WrapErr(err)
		}

		f, err = os.Create(path) // retry
		if err != nil {
			return nil, WrapErr(err)
		}
		return f, nil
	}
	return nil, WrapErr(err)
}

// Create the file if it doesn't exist, otherwise do nothing.
func TouchFile(path string) error {
	f, err := os.Open(path)
	if err == nil {
		f.Close()
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		baseDir := filepath.Dir(path)
		if err = os.MkdirAll(baseDir, PERMDIR); err != nil {
			return err
		}
		f, err = os.Create(path)
		if err != nil {
			return err
		}
		f.Close()
		return nil
	}
	return err
}
