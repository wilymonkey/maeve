package local

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilymonkey/maeve/utils"
)

// Gets the first folder in the given path.
func SplitAtRootPath(relpath string) (string, string) {
	cleaned := filepath.Clean(relpath)
	parts := strings.SplitN(cleaned, string(filepath.Separator), 2)

	utils.Assertf(
		len(parts) > 1 && parts[0] != "" && parts[1] != "",
		"%q must have root and child path", relpath,
	)

	return parts[0], parts[1]
}

// Checks if the path exists, returning false on errors.
func PathOk(path string) bool {
	exists, err := PathExists(path)
	if err != nil {
		return false
	}
	return exists
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("checking if path exists: %w", err)
	}
	return true, nil
}

// Creates a Hardlink from sourcePath to targetPath, creating
// directories as needed.
func Hardlink(sourcePath, targetPath string) error {
	err := os.Link(sourcePath, targetPath)
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("creating folders in latest: %w", err)
		}
		// Try to link the file again.
		if err := os.Link(sourcePath, targetPath); err != nil {
			return fmt.Errorf("linking file into latest AGAIN: %w", err)
		}
		return nil
	}

	return fmt.Errorf("linking file into latest: %w", err)
}

func CullSnapshots(node string) error {
	// snapshots, err := conf.GetConf().NodeSnapshots(node)
	// if err != nil {
	// 	return utils.WrapErr(err)
	// }
	// snapLen := len(snapshots)
	// if snapLen > conf.GetConf().MaxBackups {
	// 	for _, p := range snapshots[:snapLen-conf.GetConf().MaxBackups] {
	// 		if err := os.RemoveAll(p); err != nil {
	// 			return utils.WrapErr(err)
	// 		}
	// 	}
	// }
	return nil
}
