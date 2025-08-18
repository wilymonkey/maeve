package hashsums

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/zeebo/blake3"
	"golang.org/x/sync/errgroup"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
)

type FileHash struct {
	Hash     [32]byte
	Snapshot string
	RelPath  string
}

type FileMeta struct {
	Hash    [32]byte
	RelPath string
	Size    int64
	ModTime time.Time
}

func NewFileHash(basepath, targetpath string) (FileHash, error) {
	var filehash FileHash
	hash, err := NewHashsum(targetpath)
	if err != nil {
		return filehash, err
	}

	relPath, err := filepath.Rel(basepath, targetpath)
	snapshot := filepath.Dir(basepath)
	if err != nil {
		return filehash, fmt.Errorf("making relative path: %s -> %s: %w", basepath, targetpath, err)
	}

	return FileHash{
		Hash:     hash,
		Snapshot: snapshot,
		RelPath:  relPath,
	}, nil
}

func (fh *FileHash) AbsPath(node string) string {
	return filepath.Join(conf.GetConf().NodeDir(node), fh.Snapshot, fh.RelPath)
}

func (fh *FileHash) Validate(path string) (bool, error) {
	newHash, err := NewHashsum(path)
	if err != nil {
		return false, utils.WrapErr(err)
	}
	return newHash == fh.Hash, nil
}

// Generates a FileHash for the container for FileHash[] in
// the SelfDir
func GetSelfHashGob() (FileHash, error) {
	var result FileHash
	dir := conf.GetConf().SelfDir()
	path := conf.GetConf().HashFile(dir)
	hash, err := NewHashsum(path)
	if err != nil {
		return result, utils.WrapErr(err)
	}
	result = FileHash{
		RelPath: path,
		Hash:    hash,
	}
	return result, nil
}

func ValidateExisting(hashes []FileHash, node string) ([]FileHash, error) {
	dir := conf.GetConf().NodeDirTemp(node)
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
	for _, _ = range currHashes {
		// TODO: Compare what should be and what shouldn't be there
	}

	return utils.SetToSlice(sourceHashes), nil
}

func NewDirFileHash(basepath string, progChan chan FileHash, ctx context.Context) ([]FileHash, error) {
	hashChan := make(chan FileHash, 100)
	var hashes []FileHash
	wg := utils.GoWait(func() {
		for hash := range hashChan {
			hashes = append(hashes, hash)
		}
	})

	eGrp, ctx := errgroup.WithContext(ctx)
	eGrp.SetLimit(20 * runtime.NumCPU())

	walkErr := filepath.WalkDir(basepath, func(filePath string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := dir.Info()
		if err != nil {
			return fmt.Errorf("getting dir info: %w", err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		eGrp.Go(func() error {
			fh, err := NewFileHash(basepath, filePath)
			if err != nil {
				return err
			}
			hashChan <- fh
			progChan <- fh
			return nil
		})

		return nil
	})

	err := eGrp.Wait()
	close(hashChan)
	wg.Wait()
	if err != nil {
		return hashes, utils.WrapErr(err)
	}
	if walkErr != nil {
		return hashes, utils.WrapErr(walkErr)
	}
	return hashes, nil
}

func NewHashsum(path string) ([32]byte, error) {
	var result [32]byte
	f, err := os.Open(path)
	if err != nil {
		return result, fmt.Errorf("opening file: %s: %w", path, err)
	}
	defer f.Close()

	h := blake3.New()
	if _, err := io.Copy(h, f); err != nil {
		return result, fmt.Errorf("hashing file: %s: %w", path, err)
	}
	copy(result[:], h.Sum(nil))
	return result, nil
}

// Write the given []FileHash to the SelfDir.
func WriteFileHashes(hashes []FileHash) error {
	f, err := os.Create(conf.GetConf().HashFile(conf.GetConf().SelfDir()))
	if err != nil {
		return utils.WrapErr(err)
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(hashes)
}

func readFileHashes(dir string) ([]FileHash, error) {
	f, err := os.Open(conf.GetConf().HashFile(dir))
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
