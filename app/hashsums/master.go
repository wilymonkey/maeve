package hashsums

import (
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/utils"
)

var ErrNeedsRecreate = errors.New("MasterHash needs to be recreated")

type MasterHash struct {
	Dirs map[string]struct{}
	// hash as key
	HashMap map[[32]byte]local.RelPath
}

func (mh *MasterHash) Exists(fileHash OldFileHash) *local.RelPath {
	if relPath, exists := mh.HashMap[fileHash.Hash]; exists {
		return &relPath
	}
	return nil
}

// Checks folders of a given node to see if the MasterHash contains those
// dirs as well (i.e. they have been added to the hashes map).
// Only returns a value if it's invalid.
func (mh *MasterHash) validate(node string) error {
	entries, err := os.ReadDir(conf.NodeDir(node))
	if err != nil {
		return utils.WrapErr(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			if _, exists := mh.Dirs[e.Name()]; !exists {
				return ErrNeedsRecreate
			}
		}
	}

	return nil
}

// Retrieves the MasterHash of a given node, creating it if necessary.
func getMaster(node string) (MasterHash, error) {
	mh, err := readMaster(node)
	if err != nil {
		if errors.Is(err, ErrNeedsRecreate) {
			mh, err = genMasterHash(node)
			if err != nil {
				return mh, utils.WrapErr(err)
			}
			err := writeMaster(&mh, node)
			if err != nil {
				return mh, utils.WrapErr(err)
			}
		} else {
			return mh, utils.WrapErr(err)
		}
	}
	return mh, nil
}

func readMaster(node string) (MasterHash, error) {
	var mh MasterHash

	f, err := os.Open(conf.GetConf().MasterHashFile(node))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return mh, ErrNeedsRecreate
		}
		return mh, utils.WrapErr(err)
	}
	defer f.Close()

	if err := gob.NewDecoder(f).Decode(&mh); err != nil {
		return mh, utils.WrapErr(err)
	}
	if err := mh.validate(node); err != nil {
		return mh, utils.WrapErr(err)
	}

	return mh, nil
}

func genMasterHash(node string) (MasterHash, error) {
	var mh MasterHash

	baseDir := conf.NodeDir(node)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return mh, nil
		}
		return mh, utils.WrapErr(err)
	}
	tempEntry := filepath.Base(conf.GetConf().NodeDirTemp(node))

	dirs := make(map[string]struct{})
	hashMap := make(map[[32]byte]local.RelPath)
	for _, e := range entries {
		if e.IsDir() {
			if e.Name() == tempEntry {
				continue
			}

			_, err := readFileHashes(filepath.Join(baseDir, e.Name()))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return mh, utils.WrapErr(err)
			}
			dirs[e.Name()] = struct{}{}
		}
	}

	return MasterHash{Dirs: dirs, HashMap: hashMap}, nil
}

func writeMaster(mh *MasterHash, node string) error {
	f, err := utils.Create(conf.GetConf().MasterHashFile(node))
	if err != nil {
		return utils.WrapErr(err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(mh); err != nil {
		return utils.WrapErr(err)
	}
	return nil
}
