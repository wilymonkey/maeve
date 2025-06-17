package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Creates the directory for a given file.
func newDir(filePath string) error {
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("unable to create parent dir %s ⇒  %v", dirPath, err)
	}
	return nil
}
