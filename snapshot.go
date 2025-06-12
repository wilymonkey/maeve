package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Creates a snapshot of the given node.
func Snapshot(node string) error {
	source := NodeDirLatest(node)
	sourceDir, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceDir.Close()

	_, err = sourceDir.Readdir(1)
	if err != nil {
		return fmt.Errorf("%s latest folder is empty ⇒  %w", node, err)
	}
	target := filepath.Join(NodeDir(node), time.Now().Format(time.DateOnly))
	err = NewSnapshot(source, target)
	if err != nil {
		return fmt.Errorf("unable to create snapshot ⇒  %w", err)
	}

	return nil
}

// Creates all required directories then hardlinks all files from source to target.
func NewSnapshot(source, target string) error {
	sem := NewSemaphore(runtime.NumCPU() * 20)
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

	err := filepath.WalkDir(source, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relPath)

		wg.Add(1)
		go func(path, targetPath string, dir os.DirEntry) {
			defer wg.Done()

			sem.Acquire()
			defer sem.Release()

			dInfo, err := dir.Info()
			if err != nil {
				errChan <- err
				return
			}
			fileMode := dInfo.Mode()

			if dir.IsDir() {
				if err := os.MkdirAll(targetPath, dInfo.Mode().Perm()); err != nil {
					errChan <- err
				}
				return
			}

			if fileMode.IsRegular() {
				if err := os.Link(path, targetPath); err != nil {
					errChan <- err
				}
				return
			}

			if fileMode&os.ModeSymlink != 0 {
				linkTarget, err := os.Readlink(path)
				if err != nil {
					errChan <- err
					return
				}
				if err := os.Symlink(linkTarget, targetPath); err != nil {
					errChan <- err
				}
				return
			}

			// Skip other file types (e.g., devices, sockets)

		}(path, targetPath, dir)

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
