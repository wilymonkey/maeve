package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
//
// CAUTION: Deletes the Cfg.SelfDir directory.
func LocalPull() error {
	localDir := Config.SelfDir()

	if err := os.RemoveAll(localDir); err != nil {
		return fmt.Errorf("deleting local latest directory ⇒  %w", err)
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("creating backup directory ⇒  %w", err)
	}

	for _, srcDir := range Config.SourceDirs {
		destDir := filepath.Join(localDir, filepath.Base(srcDir))
		if err := NewSnapshot(srcDir, destDir); err != nil {
			return fmt.Errorf("unable to clone %s to %s ⇒  %w", srcDir, destDir, err)
		}
	}

	hashes, err := NewSHA3Sums(localDir)
	if err != nil {
		return fmt.Errorf("unable to create hashsums for local files ⇒  %w", err)
	}
	if err := WriteFileHashes(hashes, localDir); err != nil {
		return fmt.Errorf("unable to write hashsums to file ⇒  %w", err)
	}

	return nil
}

// If parent folders or files simply get renamed, we don't want to copy the source file
// in those cases. Just change the path.
type LinkAction struct {
	fromPath string
	toPath   string
	hash     string
}

// Pushes changes to a given node address.
func LocalPush(address string) error {
	client, err := NewSSHClient(address)
	if err != nil {
		return fmt.Errorf("new ssh client ⇒  %w", err)
	}
	defer client.Close()

	remoteHash, err := remoteHashes(client)
	if err != nil {
		log.Printf("Unable to read remote hashes ⇒  %w", err)
		// Assume the remote is a new node.
		remoteHash = make([]FileHash, 0)
	}
	sourceHash, err := ReadFileHashes(Config.SelfDir())
	if err != nil {
		return fmt.Errorf("read self source hash ⇒  %w", err)
	}
	matches := matchingHashes(sourceHash, remoteHash)

	return nil
}

func LinkSnapshotFiles(node string, source []FileHash) ([]FileHash, error) {
	latest, err := Config.NodeDirLatest(node)
	if err != nil {
		if err == os.ErrNotExist {

		}

		return nil, fmt.Errorf("get latest snapshot ⇒  %w", err)
	}
	local, err := ReadFileHashes(latest)
	if err != nil {
		return nil, fmt.Errorf("read self source hash ⇒  %w", err)
	}

	localSet := make(map[string]FileHash)
	for _, f := range local {
		localSet[f.hash] = f
	}

	sem := ScalingSemaphore(20)
	var wg sync.WaitGroup

	errChan := make(chan error, 100)
	var errorMU sync.Mutex
	var allErrors []error
	go func() {
		errorMU.Lock()
		for err := range errChan {
			allErrors = append(allErrors, err)
		}
		errorMU.Unlock()
	}()

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

		for _, s := range source {
			hash := s.hash
			if l, exists := localSet[hash]; exists {

				baseDir := Config.NodeDirTemp(node)
				sourcePath := filepath.Join(baseDir, l.relPath)
				targetPath := filepath.Join(baseDir, s.relPath)

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

					errChan <- err
				}
			} else {
				missingChan <- s
			}
		}
	}()

	wg.Wait()
	close(errChan)
	close(missingChan)

	if err != nil {
		errorMU.Lock()
		allErrors = append(allErrors, err)
		errorMU.Unlock()
	}

	if len(allErrors) > 0 {
		return nil, errors.Join(allErrors...)
	}
	return missing, nil
}

func matchingHashes(source, remote []FileHash) []FileHash {
	set := make(map[string]FileHash)
	for _, f := range remote {
		set[f.hash] = f
	}

	var matches []LinkAction
	for _, remoteHash := range source {
		hash := remoteHash.hash
		if sourceHash, exists := set[hash]; exists {
			action := LinkAction{fromPath: remoteHash.relPath, toPath: sourceHash.relPath, hash: hash}
			matches = append(matches, action)
		}
	}

	return matches
}
