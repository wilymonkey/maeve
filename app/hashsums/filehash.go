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
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
)

type OldFileHash struct {
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

func NewFileMeta(path string) (FileMeta, error) {
	var meta FileMeta
	hash, err := NewHashsum(path)
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

func NewFileHash(basepath, targetpath string) (OldFileHash, error) {
	var filehash OldFileHash
	hash, err := NewHashsum(targetpath)
	if err != nil {
		return filehash, err
	}

	relPath, err := filepath.Rel(basepath, targetpath)
	snapshot := filepath.Dir(basepath)
	if err != nil {
		return filehash, fmt.Errorf("making relative path: %s -> %s: %w", basepath, targetpath, err)
	}

	return OldFileHash{
		Hash:     hash,
		Snapshot: snapshot,
		RelPath:  relPath,
	}, nil
}

func (fh *OldFileHash) AbsPath(node string) string {
	return filepath.Join(conf.NodeDir(node), fh.Snapshot, fh.RelPath)
}

func (fh *OldFileHash) Validate(path string) (bool, error) {
	newHash, err := NewHashsum(path)
	if err != nil {
		return false, utils.WrapErr(err)
	}
	return newHash == fh.Hash, nil
}

// Generates a FileHash for the container for FileHash[] in
// the SelfDir
func GetSelfHashGob() (OldFileHash, error) {
	var result OldFileHash
	dir := conf.GetConf().SelfDir()
	path := conf.GetConf().HashFile(dir)
	hash, err := NewHashsum(path)
	if err != nil {
		return result, utils.WrapErr(err)
	}
	result = OldFileHash{
		RelPath: path,
		Hash:    hash,
	}
	return result, nil
}

func ValidateExisting(hashes []OldFileHash, node string) ([]OldFileHash, error) {
	dir := conf.GetConf().NodeDirTemp(node)
	exists, err := local.PathExists(dir)
	if err != nil {
		return nil, utils.WrapErr(err)
	}
	if !exists {
		return hashes, nil
	}

	emptyChan := make(chan OldFileHash)
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
	for range currHashes {
		// TODO: Compare what should be and what shouldn't be there
	}

	return utils.SetToSlice(sourceHashes), nil
}

func NewDirFileHash(basepath string, progChan chan OldFileHash, ctx context.Context) ([]OldFileHash, error) {
	hashChan := make(chan OldFileHash, 100)
	var hashes []OldFileHash
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
		return result, help.CheckSource(err, "opening source file")
	}
	defer f.Close()

	h := blake3.New()
	if _, err := io.Copy(h, f); err != nil {
		return result, help.CheckSource(err, "hashing source file")
	}
	copy(result[:], h.Sum(nil))
	return result, nil
}

// Write the given []FileHash to the SelfDir.
func WriteFileHashes(hashes []OldFileHash) error {
	f, err := os.Create(conf.GetConf().HashFile(conf.GetConf().SelfDir()))
	if err != nil {
		return utils.WrapErr(err)
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(hashes)
}

func readFileHashes(dir string) ([]OldFileHash, error) {
	f, err := os.Open(conf.GetConf().HashFile(dir))
	if err != nil {
		return nil, utils.WrapErr(err)
	}
	defer f.Close()

	var hashes []OldFileHash
	if err := gob.NewDecoder(f).Decode(&hashes); err != nil {
		return nil, utils.WrapErr(err)
	}
	return hashes, nil
}
