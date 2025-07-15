package hashsums

import (
	"context"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/zeebo/blake3"
	"golang.org/x/sync/errgroup"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/utils"
)

type FileHash struct {
	Path local.SnapshotPath
	Hash [32]byte
}

func (fh *FileHash) Validate(path string) (bool, error) {
	newHash, err := genHash(path)
	if err != nil {
		return false, utils.WrapErr(err)
	}
	return newHash == fh.Hash, nil
}

// Generates a FileHash for the container for FileHash[] in
// the SelfDir
func GetSelfHashGob() (FileHash, error) {
	var result FileHash
	dir := cfg.Global.SelfDir()
	path := cfg.Global.HashFile(dir)
	hash, err := genHash(path)
	if err != nil {
		return result, utils.WrapErr(err)
	}
	snapshot, err := local.NewSnapshotPath(dir, path)
	if err != nil {
		return result, utils.WrapErr(err)
	}
	result = FileHash{
		Path: snapshot,
		Hash: hash,
	}
	return result, nil
}

func ValidateExisting(hashes []FileHash, node string) ([]FileHash, error) {
	dir := cfg.Global.NodeDirTemp(node)
	exists, err := local.PathExists(dir)
	if err != nil {
		return nil, utils.WrapErr(err)
	}
	if !exists {
		return hashes, nil
	}

	emptyChan := make(chan FileHash)
	defer close(emptyChan)
	go func() {
		for range emptyChan {
		}
	}()
	currHashes, err := NewDirFileHash(dir, emptyChan, context.Background())
	if err != nil {
		return hashes, utils.WrapErr(err)
	}
	sourceHashes := utils.SliceToSet(hashes)
	for _, h := range currHashes {
		if _, exists := sourceHashes[h]; exists {
			delete(sourceHashes, h)
		} else {
			if err := os.Remove(h.Path.ResolveTemp(node)); err != nil {
				return hashes, utils.WrapErr(err)
			}
		}
	}

	return utils.SetToSlice(sourceHashes), nil
}

func NewDirFileHash(sourcePath string, progChan chan FileHash, ctx context.Context) ([]FileHash, error) {
	hashChan := make(chan FileHash, 100)
	defer close(hashChan)
	var hashes []FileHash
	go func() {
		for hash := range hashChan {
			hashes = append(hashes, hash)
		}
	}()

	eGrp, ctx := errgroup.WithContext(ctx)
	eGrp.SetLimit(20 * runtime.NumCPU())

	walkErr := filepath.WalkDir(sourcePath, func(filePath string, dir os.DirEntry, err error) error {
		if err != nil {
			return utils.WrapErr(err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := dir.Info()
		if err != nil {
			return utils.WrapErr(err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		eGrp.Go(func() error {
			hash, err := genHash(filePath)
			if err != nil {
				return utils.WrapErr(err)
			}

			snapshotPath, err := local.NewSnapshotPath(sourcePath, filePath)
			if err != nil {
				return utils.WrapErr(err)
			}

			fh := FileHash{Path: snapshotPath, Hash: hash}
			hashChan <- fh
			progChan <- fh
			return nil
		})

		return nil
	})

	if err := eGrp.Wait(); err != nil {
		return hashes, utils.WrapErr(err)
	}
	if walkErr != nil {
		return hashes, utils.WrapErr(walkErr)
	}
	return hashes, nil
}

// Hash a file.
func genHash(path string) ([32]byte, error) {
	var result [32]byte

	f, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer f.Close()

	h := blake3.New()
	if _, err := io.Copy(h, f); err != nil {
		return result, utils.WrapErr(err)
	}
	copy(result[:], h.Sum(nil))
	return result, nil
}

// Write the given []FileHash to the SelfDir.
func WriteFileHashes(hashes []FileHash) error {
	f, err := os.Create(cfg.Global.HashFile(cfg.Global.SelfDir()))
	if err != nil {
		return utils.WrapErr(err)
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(hashes)
}

func readFileHashes(dir string) ([]FileHash, error) {
	f, err := os.Open(cfg.Global.HashFile(dir))
	if err != nil {
		return nil, utils.WrapErr(err)
	}
	defer f.Close()

	var hashes []FileHash
	if err := gob.NewDecoder(f).Decode(&hashes); err != nil {
		return nil, utils.WrapErr(err)
	}
	return hashes, nil
}
