package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/zeebo/blake3"
	"golang.org/x/sync/errgroup"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
)

type FileMeta struct {
	Hash    [32]byte
	RelPath string
	Size    int64
	ModTime time.Time
}

func GenFileMeta(path string, hasher *blake3.Hasher) (FileMeta, error) {
	var meta FileMeta
	hash, err := NewHashsum(path, hasher)
	if err != nil {
		return meta, err
	}

	relPath, err := filepath.Rel(conf.MyNode(), path)
	if err != nil {
		return meta, help.DevReport(err, "creating relative path")
	}

	info, err := os.Stat(path)
	if err != nil {
		return meta, help.CheckSource(err, "getting file stats")
	}

	return FileMeta{
		Hash:    hash,
		RelPath: relPath,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

func (h *FileMeta) Validate(path string, hasher *blake3.Hasher) (bool, error) {
	newHash, err := NewHashsum(path, hasher)
	if err != nil {
		return false, err
	}
	return newHash == h.Hash, nil
}

func NewHashsum(path string, hasher *blake3.Hasher) ([32]byte, error) {
	var result [32]byte
	file, err := os.Open(path)
	if err != nil {
		return result, help.CheckSource(err, "opening source file")
	}
	defer file.Close()

	hasher.Reset()
	if _, err := io.Copy(hasher, file); err != nil {
		return result, help.CheckSource(err, "hashing source file")
	}
	copy(result[:], hasher.Sum(nil))
	return result, nil
}

// Writes FileMetas to a channel and closes channel when done.
func WalkDirForMetas(
	sourceDir string,
	ctx context.Context,
	metaChan chan<- *FileMeta,
	preProcess func(path string) (string, error),
) error {
	defer close(metaChan)

	numWorkers := runtime.NumCPU() * 2
	paths := make(chan string, numWorkers*2)
	eGrp, ctx := errgroup.WithContext(ctx)
	for range numWorkers {
		eGrp.Go(func() error {
			hasher := blake3.New()
			for path := range paths {
				prePath, err := preProcess(path)
				if err != nil {
					return err
				}
				meta, err := GenFileMeta(prePath, hasher)
				if err != nil {
					return err
				}
				metaChan <- &meta
			}
			return nil
		})
	}

	walkErr := filepath.WalkDir(sourceDir, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if dir.Type().IsRegular() {
			paths <- path
		}
		return nil
	})

	close(paths)
	if err := eGrp.Wait(); err != nil {
		return err
	}
	if walkErr != nil {
		return walkErr
	}

	return nil
}
