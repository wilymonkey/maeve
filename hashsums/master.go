package hashsums

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/utils"
)

var ErrInvalidMH = errors.New("invalid MasterHash file")

type MasterHash struct {
	dirs map[string]struct{}
	// hash as key, relPath as value
	hashes map[[32]byte]utils.RelativePath
}

func (mh *MasterHash) Exists(fileHash FileHash) *utils.RelativePath {
	if relPath, exists := mh.hashes[fileHash.Hash]; exists {
		return &relPath
	}
	return nil
}

// Checks folders of a given node to see if the MasterHash contains those
// dirs as well (i.e. they have been added to the hashes map).
// Only returns a value if it's invalid.
func (mh *MasterHash) validate(node string) error {
	entries, err := os.ReadDir(cfg.Global.NodeDir(node))
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

	f, err := os.Open(cfg.Global.MasterHashFile(node))
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
	baseDir := cfg.Global.NodeDir(node)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read entries in node %s ⇒  %w", node, err)
	}

	dirs := make(map[string]struct{})
	hashes := make(map[[32]byte]utils.RelativePath)
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
				hashes[h.Hash] = h.Path.Resolve(e.Name())
			}
			dirs[e.Name()] = struct{}{}
		}
	}
	mh := MasterHash{dirs: dirs, hashes: hashes}

	f, err := os.Create(cfg.Global.MasterHashFile(node))
	if err != nil {
		return nil, fmt.Errorf("open MasterHash file ⇒  %w", err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(mh); err != nil {
		return nil, fmt.Errorf("write MasterHash file ⇒  %w", err)
	}

	return &mh, nil
}
