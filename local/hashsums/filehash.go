package hashsums

import (
	"bufio"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sync/errgroup"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local/mypath"
	"github.com/wilymonkey/maeve/utils/myerr"
	"github.com/zeebo/blake3"
)

type FileHash struct {
	Path mypath.SnapshotPath
	Hash [32]byte
}

func NewDirFileHash(sourcePath string, progChan chan FileHash, ctx context.Context) error {
	hashChan := make(chan FileHash, 100)
	defer close(hashChan)
	var hashsums []FileHash
	go func() {
		for hash := range hashChan {
			hashsums = append(hashsums, hash)
		}
	}()

	eGrp, ctx := errgroup.WithContext(ctx)
	eGrp.SetLimit(20 * runtime.NumCPU())

	walkErr := filepath.WalkDir(sourcePath, func(filePath string, dir os.DirEntry, err error) error {
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
		if info, err := dir.Info(); err != nil {
			return err
		} else if !info.Mode().IsRegular() {
			return nil
		}

		eGrp.Go(func() error {
			hash, err := GenHash(filePath)
			if err != nil {
				return myerr.WrapErr(err)
			}

			snapshotPath, err := mypath.NewSnapshotPath(sourcePath, filePath)
			if err != nil {
				return myerr.WrapErr(err)
			}

			fh := FileHash{Path: snapshotPath, Hash: hash}
			hashChan <- fh
			progChan <- fh
			return nil
		})

		return nil
	})

	if err := eGrp.Wait(); err != nil {
		return myerr.WrapErr(err)
	}
	if walkErr != nil {
		return myerr.WrapErr(walkErr)
	}

	if err := writeFileHashes(hashsums); err != nil {
		return myerr.WrapErr(err)
	}

	return nil
}

// Hash on a file.
func GenHash(path string) ([32]byte, error) {
	var result [32]byte

	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := blake3.New()
	reader := bufio.NewReader(file)

	buf := make([]byte, 64*1024) // 64kb reads at one time.
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return result, err
		}
		if n == 0 {
			break
		}
		// Never returns an error.
		_, _ = hash.Write(buf[:n])
	}

	copy(result[:], hash.Sum(nil))
	return result, nil
}

// Write the given []FileHash to the SelfDir.
func writeFileHashes(fileHashes []FileHash) error {
	f, err := os.Create(cfg.Global.HashFile(cfg.Global.SelfDir()))
	if err != nil {
		return fmt.Errorf("unable to open hash file ⇒  %w", err)
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(fileHashes)
}

func ReadFileHashes(dir string) ([]FileHash, error) {
	f, err := os.Open(cfg.Global.HashFile(dir))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hashes []FileHash
	if err := gob.NewDecoder(f).Decode(&hashes); err != nil {
		return nil, fmt.Errorf("unable to decode hash file ⇒  %w", err)
	}
	return hashes, nil
}
