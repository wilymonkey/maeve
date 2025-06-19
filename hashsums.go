package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/sha3"
)

type FileHash struct {
	relPath string
	hash    [32]byte
}

func NewSHA3Sums(source string) ([]FileHash, error) {
	sem := NewScalingSemaphore(20)
	var wg sync.WaitGroup

	errChan := make(chan error, 100)
	var mu sync.Mutex
	var allErrors []error
	go func() {
		for err := range errChan {
			mu.Lock()
			allErrors = append(allErrors, err)
			mu.Unlock()
		}
	}()

	hashChan := make(chan FileHash, 100)
	var hashMU sync.Mutex
	var hashsums []FileHash
	go func() {
		hashMU.Lock()
		for hash := range hashChan {
			hashsums = append(hashsums, hash)
		}
		hashMU.Unlock()
	}()

	err := filepath.WalkDir(source, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir.IsDir() {
			return nil
		}

		wg.Add(1)
		go func(sourcePath, filePath string, dir os.DirEntry) {
			defer wg.Done()

			sem.Acquire()
			defer sem.Release()

			dInfo, err := dir.Info()
			if err != nil {
				errChan <- err
				return
			}

			if dInfo.Mode().IsRegular() {
				hash, err := hashFile(filePath)
				if err != nil {
					errChan <- err
					return
				}
				relPath, err := filepath.Rel(sourcePath, filePath)
				if err != nil {
					errChan <- err
					return
				}
				hashChan <- FileHash{relPath: relPath, hash: hash}
				return
			}

			// Skip other file types (e.g., devices, sockets, symlinks)

		}(source, path, dir)

		return nil
	})

	wg.Wait()
	close(errChan)
	close(hashChan)

	if err != nil {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return nil, errors.Join(allErrors...)
	}

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

	hash := sha3.NewShake256()
	reader := bufio.NewReader(file)
	// 64kb reads at one time.
	buf := make([]byte, 64*1024)
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

	hash.Read(result[:])
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

type MasterHash struct {
	dirs map[string]struct{}
	// hash as key, relPath as value
	hashes map[[32]byte]string
}

// Checks folders of a given node to see if the MasterHash contains those
// dirs as well (i.e. they have been added to the hashes map).
func (mh *MasterHash) validate(node string) (bool, error) {
	entries, err := os.ReadDir(Config.NodeDir(node))
	if err != nil {
		return false, fmt.Errorf("read entries in node %s ⇒  %w", node, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			if _, exists := mh.dirs[e.Name()]; !exists {
				return false, nil
			}
		}
	}

	return true, nil
}

// Retrieves the MasterHash of a given node, creating it if
// it doesn't exist or is invalid.
func GetMasterHash(node string) (*MasterHash, error) {
	var mh *MasterHash

	f, err := os.Open(Config.MasterHashFile(node))
	if err != nil {
		if err == os.ErrNotExist {
			mh, err = newMasterHash(node)
			if err != nil {
				return nil, fmt.Errorf("missing file; create MasterHash ⇒  %w", err)
			}
		}
		return nil, err
	}
	defer f.Close()

	if err := gob.NewDecoder(f).Decode(mh); err != nil {
		return nil, fmt.Errorf("decode MasterHash file ⇒  %w", err)
	}
	isValid, err := mh.validate(node)
	if err != nil {
		return nil, fmt.Errorf("validate MasterHash")
	}
	if !isValid {
		mh, err = newMasterHash(node)
		if err != nil {
			return nil, fmt.Errorf("invalid hashes; create MasterHash ⇒  %w", err)
		}
	}

	return mh, nil
}

// Creates a MasterHash file for a given node.
func newMasterHash(node string) (*MasterHash, error) {
	baseDir := Config.NodeDir(node)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read entries in node %s ⇒  %w", node, err)
	}

	dirs := make(map[string]struct{})
	hashes := make(map[[32]byte]string)
	for _, e := range entries {
		if e.IsDir() {
			hFile, err := ReadFileHashes(filepath.Join(baseDir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("read hash file in snapshot %s ⇒  %w", e.Name(), err)
			}
			for _, h := range hFile {
				hashes[h.hash] = h.relPath
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
