package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Creates a snapshot and hashsum file of a given node. Trims snapshots at the end.
func Snapshot(nodeName string) error {
	source, err := Config.NodeDirLatest(nodeName)
	sourceDir, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceDir.Close()

	_, err = sourceDir.Readdir(1)
	if err != nil {
		return fmt.Errorf("%s latest folder is empty ⇒  %w", nodeName, err)
	}

	snapshotName := time.Now().Format("20060102-1504")
	target := filepath.Join(NodeDir(nodeName), snapshotName)
	if err = NewSnapshot(source, target); err != nil {
		return fmt.Errorf("unable to create snapshot ⇒  %w", err)
	}

	if err = NewSHA3Sums(nodeName, snapshotName); err != nil {
		return fmt.Errorf("unable to create hashsums ⇒  %w", err)
	}

	if err = trimSnapshots(nodeName); err != nil {
		return fmt.Errorf("unable to trim snapshots ⇒  %w", err)
	}

	return nil
}

// Creates hardlinks for all files from source to target.
// CAUTION: Deletes the target directory if it exists.
func NewSnapshot(source, target string) error {
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("deleting %s ⇒  %w", target, err)
	}

	sem := NewScalingSemaphore(20)
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
		// Directories are created when files are created.
		if dir.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relPath)

		wg.Add(1)
		go func(sourcePath, targetPath string, dir os.DirEntry) {
			defer wg.Done()

			sem.Acquire()
			defer sem.Release()

			dInfo, err := dir.Info()
			if err != nil {
				errChan <- err
				return
			}
			fileMode := dInfo.Mode()

			if fileMode.IsRegular() {
				if err := os.Link(sourcePath, targetPath); err != nil {

					if os.IsNotExist(err) {
						if err := newDir(targetPath); err != nil {
							errChan <- err
						}
						// Try to link the file again.
						if err := os.Link(sourcePath, targetPath); err != nil {
							errChan <- err
						}
						return
					}

					if os.IsExist(err) {
						replaceFile, err := shouldReplace(sourcePath, targetPath)
						if err != nil {
							errChan <- fmt.Errorf("unable to stat file ⇒  %w", err)
						}
						if replaceFile {
							if err := os.Remove(targetPath); err != nil {
								errChan <- fmt.Errorf("unable to remove target file ⇒  %w", err)
							}
							// Try to link the file again.
							if err := os.Link(sourcePath, targetPath); err != nil {
								errChan <- err
							}
						}
						return
					}

					errChan <- err
				}
				return
			}

			// Skip other file types (e.g., devices, sockets, symlinks)

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

func shouldReplace(source, target string) (bool, error) {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return false, err
	}
	targetInfo, err := os.Stat(target)
	if err != nil {
		return false, err
	}

	return sourceInfo.ModTime().After(targetInfo.ModTime()), nil
}

func trimSnapshots(node string) error {
	// Increased by 1 to ignore the "latest" folder.
	maxBackups := Config.MaxBackups
	dir := NodeDir(node)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("unable to read entries in node %s ⇒  %w", node, err)
	}

	if len(entries) <= maxBackups {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, folder := range entries[:len(entries)-maxBackups] {
		path := filepath.Join(dir, folder.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("unable to delete snapshot %s ⇒  %w", path, err)
		}
	}
	return nil
}

func LinkFilesFromSnapshots(node string, sourceHash []FileHash) ([]FileHash, error) {
	localMaster, err := GetMasterHash(node)
	if err != nil {
		return nil, fmt.Errorf("get MasterHash ⇒  %w", err)
	}

	sem := NewScalingSemaphore(20)
	var wg sync.WaitGroup

	// TODO: Stop all linking when 1 error is encountered.

	missingChan := make(chan FileHash, 100)
	var missingMU sync.Mutex
	var missing []FileHash
	go func() {
		missingMU.Lock()
		for hash := range missingChan {
			missing = append(missing, hash)
		}
		missingMU.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		sem.Acquire()
		defer sem.Release()

		for _, s := range sourceHash {
			if lPath, exists := localMaster.hashes[s.hash]; exists {
				// TODO: Link local path with source path.
			} else {
				missingChan <- s
			}
		}
	}()

	wg.Wait()
	close(missingChan)

	return missing, nil
}
