package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
)

// Creates hardlinks for all files from source to target.
//
// CAUTION: Deletes the target directory if it exists.
func HardlinkDir(source, target string, sizeChan chan int64, ctx context.Context) error {
	if err := os.RemoveAll(target); err != nil {
		return utils.WrapErr(err)
	}

	eGrp, ctx := errgroup.WithContext(ctx)
	eGrp.SetLimit(20 * runtime.NumCPU())

	err := filepath.WalkDir(source, func(sourcePath string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := os.Stat(sourcePath)
		if err != nil {
			return utils.WrapErr(err)
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
			return Hardlink(sourcePath, targetPath)
		})

		return nil
	})

	if err := eGrp.Wait(); err != nil {
		return utils.WrapErr(err)
	}
	if err != nil {
		return utils.WrapErr(err)
	}

	return nil
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
		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModeDir); err != nil {
			return utils.WrapErr(err)
		}
		// Try to link the file again.
		if err := os.Link(sourcePath, targetPath); err != nil {
			return utils.WrapErr(err)
		}
		return nil
	}

	return utils.WrapErr(err)
}
