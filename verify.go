package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/goccy/go-yaml"
	"golang.org/x/crypto/sha3"
)

type FileHash struct {
	relPath string `yaml:"RelativePath:"`
	hash    string `yaml:"SHAKE256:"`
}

func NewSHA3Sums(source string) error {
	sem := NewSemaphore(20)
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
	go func() {
		hashMU.Lock()
		for hash := range hashChan {
			writeHashFile(source, hash)
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
		return errors.Join(allErrors...)
	}

	return nil
}

// Does a rolling hash on a file.
func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha3.NewShake256()
	reader := bufio.NewReader(file)
	// 64kb reads at one time.
	buf := make([]byte, 64*1024)
	for {
		reader_len, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if reader_len == 0 {
			break
		}
		hash.Write(buf[:reader_len])
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeHashFile(path string, hash FileHash) error {
	data, err := yaml.Marshal(hash)
	if err != nil {
		return fmt.Errorf("unable to convert FileHash ⇒  %w", err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("unable to open hash file ⇒  %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("unable to append to hash file ⇒  %w", err)
	}
	return nil
}
