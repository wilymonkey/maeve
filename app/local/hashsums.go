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

	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/utils"
)

type FileMeta struct {
	Hash    [32]byte
	RelPath string
	Size    int64
	ModTime time.Time
}

func GenFileMeta(path string, baseDir string, hasher *blake3.Hasher) (FileMeta, error) {
	var meta FileMeta
	hash, err := NewHashsum(path, hasher)
	if err != nil {
		return meta, err
	}

	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return meta, help.WrapErr(err, "creating relative path")
	}

	info, err := os.Stat(path)
	if err != nil {
		return meta, help.WrapErr(err, "getting file stats")
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

func (h *FileMeta) AsMapKey() string {
	hashLen := len(h.Hash)
	b := make([]byte, hashLen+len(h.RelPath))
	copy(b, h.Hash[:])
	_, rest := SplitAtRootPath(h.RelPath)
	copy(b[hashLen:], rest)
	return string(b)
}

func NewHashsum(path string, hasher *blake3.Hasher) ([32]byte, error) {
	var result [32]byte
	file, err := os.Open(path)
	if err != nil {
		return result, help.WrapErr(err, "opening source file")
	}
	defer utils.Cleanup(&err, file.Close)

	hasher.Reset()
	if _, err := io.Copy(hasher, file); err != nil {
		return result, help.WrapErr(err, "hashing source file")
	}
	copy(result[:], hasher.Sum(nil))
	return result, err
}

// Writes FileMetas to a channel and closes channel when done.
// Launches in a goroutine.
// Writes FileMetas to a channel and closes channel when done.
func WalkDirForMetas(
	baseDir string, // What the relpath in FileMetas is relative to.
	sourceDir string,
	ctx context.Context,
	metaChan chan<- *FileMeta,
	preProcess func(path string) (string, error),
) error {
	defer close(metaChan)

	pp := preProcess
	if pp == nil {
		pp = func(path string) (string, error) {
			return path, nil
		}
	}

	numWorkers := runtime.NumCPU() * 2
	paths := make(chan string, numWorkers*2)
	eGrp, ctx := errgroup.WithContext(ctx)
	for range numWorkers {
		eGrp.Go(func() error {
			hasher := blake3.New()
			for path := range paths {
				prePath, err := pp(path)
				if err != nil {
					return err
				}
				meta, err := GenFileMeta(prePath, baseDir, hasher)
				if err != nil {
					return err
				}
				metaChan <- &meta
			}
			return nil
		})
	}

	err := filepath.WalkDir(sourceDir, func(path string, dir os.DirEntry, err error) error {
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
	if err != nil {
		return help.WrapErr(err, "walking dir")
	}

	return nil
}
