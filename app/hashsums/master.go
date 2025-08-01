package hashsums

import (
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/utils"
)

var ErrNeedsRecreate = errors.New("MasterHash needs to be recreated")

type MasterHash struct {
	Dirs map[string]struct{}
	// hash as key
	HashMap map[[32]byte]local.RelativePath
}

func (mh *MasterHash) Exists(fileHash FileHash) *local.RelativePath {
	if relPath, exists := mh.HashMap[fileHash.Hash]; exists {
		return &relPath
	}
	return nil
}

// Checks folders of a given node to see if the MasterHash contains those
// dirs as well (i.e. they have been added to the hashes map).
// Only returns a value if it's invalid.
func (mh *MasterHash) validate(node string) error {
	entries, err := os.ReadDir(conf.Global.NodeDir(node))
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

// Updates the MasterHash with a given snapshot folder name.
func UpdateMaster(node, snapshot string) error {
	hashes, err := readFileHashes(conf.Global.NodeSnapshotDir(node, snapshot))
	if err != nil {
		return utils.WrapErr(err)
	}
	mh, err := getMaster(node)
	if err != nil {
		return utils.WrapErr(err)
	}
	for _, h := range hashes {
		mh.HashMap[h.Hash] = h.Path.Resolve(snapshot)
	}
	mh.Dirs[snapshot] = struct{}{}
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

	f, err := os.Open(conf.Global.MasterHashFile(node))
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

	baseDir := conf.Global.NodeDir(node)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return mh, nil
		}
		return mh, utils.WrapErr(err)
	}
	tempEntry := filepath.Base(conf.Global.NodeDirTemp(node))

	dirs := make(map[string]struct{})
	hashMap := make(map[[32]byte]local.RelativePath)
	for _, e := range entries {
		if e.IsDir() {
			if e.Name() == tempEntry {
				continue
			}

			hashes, err := readFileHashes(filepath.Join(baseDir, e.Name()))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return mh, utils.WrapErr(err)
			}
			for _, h := range hashes {
				hashMap[h.Hash] = h.Path.Resolve(e.Name())
			}
			dirs[e.Name()] = struct{}{}
		}
	}

	return MasterHash{Dirs: dirs, HashMap: hashMap}, nil
}

func writeMaster(mh *MasterHash, node string) error {
	f, err := utils.Create(conf.Global.MasterHashFile(node))
	if err != nil {
		return utils.WrapErr(err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(mh); err != nil {
		return utils.WrapErr(err)
	}
	return nil
}
