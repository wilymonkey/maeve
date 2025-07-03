package hashsums

import (
	"bufio"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local/mypath"
	"github.com/wilymonkey/maeve/utils/semaphore"
	"github.com/zeebo/blake3"
)

type FileHash struct {
	Path mypath.SnapshotPath
	Hash [32]byte
}

func NewDirFileHash(sourcePath string, progChan chan struct{}) error {
	sem := semaphore.NewScaling(20)
	defer sem.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eGrp, ctx := errgroup.WithContext(ctx)

	hashChan := make(chan FileHash, 100)
	var hashsums []FileHash
	go func() {
		for hash := range hashChan {
			hashsums = append(hashsums, hash)
		}
	}()

	err := filepath.WalkDir(sourcePath, func(filePath string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info, err := dir.Info(); err != nil {
			return err
		} else if !info.Mode().IsRegular() {
			return nil
		}

		eGrp.Go(func() error {
			sem.Acquire()
			defer sem.Release()

			hash, err := hashFile(filePath)
			if err != nil {
				return err
			}

			hashChan <- FileHash{Path: mypath.NewSnapshotPath(sourcePath, filePath), Hash: hash}
			progChan <- struct{}{}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk dir %s ⇒  %w", sourcePath, err)
	}
	if err := eGrp.Wait(); err != nil {
		return fmt.Errorf("hardlink files ⇒  %w", err)
	}
	close(hashChan)

	if err := writeFileHashes(hashsums); err != nil {
		return fmt.Errorf("write hashsums to file ⇒  %w", err)
	}

	return nil
}

// Hash on a file.
func hashFile(path string) ([32]byte, error) {
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
