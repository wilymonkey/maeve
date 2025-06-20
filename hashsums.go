package main

import (
	"bufio"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"path/filepath"

	"github.com/zeebo/blake3"
)

type FileHash struct {
	path RelativeSnapshotPath
	hash [32]byte
}

func NewDirFileHash(source string) ([]FileHash, error) {
	sem := NewScalingSemaphore(20)
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

	err := filepath.WalkDir(source, func(filePath string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir.IsDir() {
			return nil
		}

		eGrp.Go(func() error {
			sem.Acquire()
			defer sem.Release()

			if info, err := dir.Info(); err != nil {
				return err
			} else if !info.Mode().IsRegular() {
				return nil
			}

			hash, err := hashFile(filePath)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(source, filePath)
			if err != nil {
				return err
			}

			hashChan <- FileHash{path: RelativeSnapshotPath{path: relPath}, hash: hash}
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
		return nil, fmt.Errorf("walk dir %s ⇒  %w", source, err)
	}
	if err := eGrp.Wait(); err != nil {
		return nil, fmt.Errorf("hardlink files ⇒  %w", err)
	}
	close(hashChan)

	return hashsums, nil
}

// Does a rolling hash on a file.
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
		hash.Write(buf[:n])
	}

	copy(result[:], hash.Sum(nil))
	return result, nil
}

func WriteFileHashes(fileHashes []FileHash, dir string) error {
	f, err := os.Create(Config.HashFile(dir))
	if err != nil {
		return fmt.Errorf("unable to open hash file ⇒  %w", err)
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(fileHashes)
}

func ReadFileHashes(dir string) ([]FileHash, error) {
	f, err := os.Open(Config.HashFile(dir))
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

var ErrInvalidMH = errors.New("invalid MasterHash file")

type MasterHash struct {
	dirs map[string]struct{}
	// hash as key, relPath as value
	hashes map[[32]byte]RelativeNodePath
}

// Checks folders of a given node to see if the MasterHash contains those
// dirs as well (i.e. they have been added to the hashes map).
// Only returns a value if it's invalid.
func (mh *MasterHash) validate(node string) error {
	entries, err := os.ReadDir(Config.NodeDir(node))
	if err != nil {
		return fmt.Errorf("read entries in node %s ⇒  %w", node, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			if _, exists := mh.dirs[e.Name()]; !exists {
				return ErrInvalidMH
			}
		}
	}

	return nil
}

// Retrieves the MasterHash of a given node.
func GetMasterHash(node string) (*MasterHash, error) {
	var masterHash *MasterHash

	f, err := os.Open(Config.MasterHashFile(node))
	if err != nil {
		return nil, fmt.Errorf("open MasterHash file ⇒  %w", err)
	}
	defer f.Close()

	if err := gob.NewDecoder(f).Decode(masterHash); err != nil {
		return nil, fmt.Errorf("decode MasterHash file ⇒  %w", err)
	}
	if err := masterHash.validate(node); err != nil {
		return nil, fmt.Errorf("validate MasterHash ⇒  %w", err)
	}

	return masterHash, nil
}

var ErrMissingHashFile = errors.New("hash file is missing")

// Creates a MasterHash file for a given node.
func NewMasterHash(node string) (*MasterHash, error) {
	baseDir := Config.NodeDir(node)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read entries in node %s ⇒  %w", node, err)
	}

	dirs := make(map[string]struct{})
	hashes := make(map[[32]byte]RelativeNodePath)
	for _, e := range entries {
		if e.IsDir() {
			hFile, err := ReadFileHashes(filepath.Join(baseDir, e.Name()))
			if err != nil {
				if err == os.ErrNotExist {
					err = ErrMissingHashFile
				}
				return nil, fmt.Errorf("read hash file in snapshot %s ⇒  %w", e.Name(), err)
			}
			for _, h := range hFile {
				hashes[h.hash] = h.path.toNode(e.Name())
			}
			dirs[e.Name()] = struct{}{}
		}
	}
	mh := MasterHash{dirs: dirs, hashes: hashes}

	f, err := os.Create(Config.MasterHashFile(node))
	if err != nil {
		return nil, fmt.Errorf("open MasterHash file ⇒  %w", err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(mh); err != nil {
		return nil, fmt.Errorf("write MasterHash file ⇒  %w", err)
	}

	return &mh, nil
}
