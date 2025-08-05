package utils

import (
	"os"
	"path/filepath"
)

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
