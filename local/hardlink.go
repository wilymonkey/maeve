package local

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wilymonkey/maeve/utils/myerr"
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

// Creates a hardlink from sourcePath to targetPath, creating
// directories as needed.
//
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

	return err
}
