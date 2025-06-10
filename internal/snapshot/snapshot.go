package snapshot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	config "github.com/wilymonkey/maeve/internal/config"
)

// Creates a snapshot of the given node.
func Node(node string) error {
	source := config.NodeDirLatest(node)
	sourceDir, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceDir.Close()

	_, err = sourceDir.Readdir(1)
	if err != nil {
		return fmt.Errorf("%s latest folder is empty: %w", node, err)
	}
	target := filepath.Join(config.NodeDir(node), time.Now().Format(time.DateOnly))
	err = Create(source, target)
	if err != nil {
		return fmt.Errorf("unable to create snapshot: %w", err)
	}

	return nil
}

// Creates all required directories then hardlinks all files from source to target.
func Create(source, target string) error {
	if err := createDir(source, target); err != nil {
		return err
	}
	var wg sync.WaitGroup
	errChan := make(chan error, 100)
	var mu sync.Mutex
	var allErrors []error

	go func() {
		for err := range errChan {
			mu.Lock()
			allErrors = append(allErrors, err)
			mu.Unlock()
		}
	}()

	err := filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		targetPath, err := targetPath(source, path, target)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(path, targetPath string, d os.DirEntry) {
			defer wg.Done()

			dInfo, err := d.Info()
			if err != nil {
				errChan <- err
				return
			}
			fileMode := dInfo.Mode()

			if fileMode.IsRegular() {
				if err := os.Link(path, targetPath); err != nil {
					errChan <- err
				}
			} else if fileMode&os.ModeSymlink != 0 {
				linkTarget, err := os.Readlink(path)
				if err != nil {
					errChan <- err
					return
				}
				if err := os.Symlink(linkTarget, targetPath); err != nil {
					errChan <- err
				}
			}
			// Skip other file types (e.g., devices, sockets)
		}(path, targetPath, d)

		return nil
	})

	wg.Wait()
	close(errChan)

	if err != nil {
		mu.Lock()
		allErrors = append(allErrors, err)
		mu.Unlock()
	}

	mu.Lock()
	defer mu.Unlock()
	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}

func createDir(source, target string) error {
	return filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		targetPath, err := targetPath(source, path, target)
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		return os.MkdirAll(targetPath, info.Mode().Perm())
	})
}

func targetPath(source, path, target string) (string, error) {
	relPath, err := filepath.Rel(source, path)
	if err != nil {
		return "", err
	}
	return filepath.Join(target, relPath), nil

}
