package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/wilymonkey/maeve/cfg"
	hs "github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/utils/myerr"
	"github.com/wilymonkey/maeve/utils/semaphore"
	"golang.org/x/sync/errgroup"
)

// Creates hardlinks for all files from source to target.
//
// CAUTION: Deletes the target directory if it exists.
func HardlinkDir(source, target string, sizeChan chan int64, ctx context.Context) error {
	if err := os.RemoveAll(target); err != nil {
		return myerr.WrapErr(err)
	}

	eGrp, ctx := errgroup.WithContext(ctx)
	eGrp.SetLimit(20 * runtime.NumCPU())

	err := filepath.WalkDir(source, func(sourcePath string, dir os.DirEntry, err error) error {
		// time.Sleep(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue
		}

		if err != nil {
			return err
		}
		info, err := os.Stat(sourcePath)
		if err != nil {
			return myerr.WrapErr(err)
		}
		// Ignore dirs, symlinks, sockets etc.
		if !info.Mode().IsRegular() {
			return nil
		}

		relPath, err := filepath.Rel(source, sourcePath)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relPath)

		eGrp.Go(func() error {
			sizeChan <- info.Size()
			return hardlink(sourcePath, targetPath)
		})

		return nil
	})

	if err := eGrp.Wait(); err != nil {
		return myerr.WrapErr(err)
	}
	if err != nil {
		return myerr.WrapErr(err)
	}

	return nil
}

func TrimSnapshots(node string) error {
	maxBackups := cfg.Global.MaxBackups
	dir := cfg.Global.NodeDir(node)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return myerr.WrapErr(err)
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
			return myerr.WrapErr(err)
		}
	}
	return nil
}

func LinkFilesFromSnapshots(node string, remoteHashes []hs.FileHash) ([]hs.FileHash, error) {
	localMaster, err := hs.GetMasterHash(node)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist) || errors.Is(err, hs.ErrInvalidMH):
			localMaster, err = hs.NewMasterHash(node)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return remoteHashes, nil
				}
				return nil, myerr.WrapErr(err)
			}
		default:
			return nil, myerr.WrapErr(err)
		}
	}

	sem := semaphore.NewScaling(20)
	defer sem.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eGrp, ctx := errgroup.WithContext(ctx)

	missingChan := make(chan hs.FileHash, 100)
	var missing []hs.FileHash
	go func() {
		for hash := range missingChan {
			missing = append(missing, hash)
		}
	}()

	eGrp.Go(func() error {
		sem.Acquire()
		defer sem.Release()
		for _, rHash := range remoteHashes {
			if relPath := localMaster.Exists(rHash); relPath != nil {
				sourcePath := relPath.Resolve(node)
				targetPath := rHash.Path.ResolveTemp(node)
				if err := hardlink(sourcePath, targetPath); err != nil {
					return err
				}
			} else {
				missingChan <- rHash
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
		return nil, myerr.WrapErr(err)
	}
	close(missingChan)

	return missing, nil
}

// Creates a hardlink from sourcePath to targetPath, creating
// directories as needed.
// CAUTION: Assumes the file doesn't exist.
func hardlink(sourcePath, targetPath string) error {
	err := os.Link(sourcePath, targetPath)

	if err == nil {
		return nil
	}

	// Assume the error is the base folder not existing
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return myerr.WrapErr(err)
	}
	// Try to link the file again.
	if err := os.Link(sourcePath, targetPath); err != nil {
		return myerr.WrapErr(err)
	}

	return nil
}
