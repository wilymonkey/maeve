package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Creates hardlinks for all files from source to target.
// CAUTION: Deletes the target directory if it exists.
func HardlinkDir(source, target string) error {
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("delete target dir %s ⇒  %w", target, err)
	}

	sem := NewScalingSemaphore(20)
	defer sem.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eGrp, ctx := errgroup.WithContext(ctx)

	err := filepath.WalkDir(source, func(sourcePath string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Directories are created when files are.
		if dir.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, sourcePath)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relPath)

		eGrp.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			sem.Acquire()
			defer sem.Release()

			return hardlink(sourcePath, targetPath)
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk dir %s ⇒  %w", source, err)
	}
	if err := eGrp.Wait(); err != nil {
		return fmt.Errorf("hardlink files ⇒  %w", err)
	}

	return nil
}

func TrimSnapshots(node string) error {
	maxBackups := Config.MaxBackups
	dir := Config.NodeDir(node)
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

func LinkFilesFromSnapshots(node string, remoteHash []FileHash) ([]FileHash, error) {
	localMaster, err := GetMasterHash(node)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist) || errors.Is(err, ErrInvalidMH):
			localMaster, err = NewMasterHash(node)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return remoteHash, nil
				}
				return nil, fmt.Errorf("create MasterHash ⇒  %w", err)
			}
		default:
			return nil, fmt.Errorf("get MasterHash ⇒  %w", err)
		}
	}

	sem := NewScalingSemaphore(20)
	defer sem.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eGrp, ctx := errgroup.WithContext(ctx)

	missingChan := make(chan FileHash, 100)
	var missing []FileHash
	go func() {
		for hash := range missingChan {
			missing = append(missing, hash)
		}
	}()

	eGrp.Go(func() error {
		sem.Acquire()
		defer sem.Release()
		for _, r := range remoteHash {
			if lPath, exists := localMaster.hashes[r.hash]; exists {
				sourcePath := lPath.toPath(node)
				targetPath := r.path.toTempPath(node)
				if err := hardlink(sourcePath, targetPath); err != nil {
					return err
				}
			} else {
				missingChan <- r
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})

	if err := eGrp.Wait(); err != nil {
		return nil, fmt.Errorf("hardlink snapshot files ⇒  %w", err)
	}
	close(missingChan)

	return missing, nil
}

// Creates a hardlink from sourcePath to targetPath, creating
// directories as needed.
func hardlink(sourcePath, targetPath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat source path ⇒  %w", err)
	}
	// Ignore symlinks, sockets etc.
	if info.Mode().IsRegular() {
		return nil
	}

	err = os.Link(sourcePath, targetPath)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		if err := newDir(targetPath); err != nil {
			return fmt.Errorf("create base dir ⇒  %w", err)
		}
		// Try to link the file again.
		if err := os.Link(sourcePath, targetPath); err != nil {
			return fmt.Errorf("create file link ⇒  %w", err)
		}
	}

	if os.IsExist(err) {
		replaceFile, err := shouldReplace(sourcePath, targetPath)
		if err != nil {
			return fmt.Errorf("compare files ⇒  %w", err)
		}
		if replaceFile {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("remove existing file ⇒  %w", err)
			}
			if err := os.Link(sourcePath, targetPath); err != nil {
				return fmt.Errorf("replace file link ⇒  %w", err)
			}
		}
	}

	return err
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
