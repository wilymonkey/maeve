package mypath

import (
	"path/filepath"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/utils/myerr"
)

type RelativePath struct {
	Path string
}

// Resolves to absolute path.
func (r *RelativePath) Resolve(node string) string {
	return filepath.Join(cfg.Global.NodeDir(node), r.Path)
}

type SnapshotPath struct {
	Path string
}

func NewSnapshotPath(sourcePath, filePath string) (SnapshotPath, error) {
	relPath, err := filepath.Rel(sourcePath, filePath)
	if err != nil {
		return SnapshotPath{}, myerr.WrapErr(err)
	}

	return SnapshotPath{Path: relPath}, nil
}

// Resolves to relative path.
func (s *SnapshotPath) Resolve(snapshot string) RelativePath {
	return RelativePath{Path: filepath.Join(snapshot, s.Path)}
}

// Resolves to an absolute path with NodeDirTemp as the base.
func (s *SnapshotPath) ResolveTemp(node string) string {
	return filepath.Join(cfg.Global.NodeDirTemp(node), s.Path)
}

// Resolves to an absolute path with SelfDir as the base.
func (s *SnapshotPath) ResolveSelf() string {
	return filepath.Join(cfg.Global.SelfDir(), s.Path)
}

// Resolves to an absolute path with SelfDir as the base.
func (s *SnapshotPath) ResolvePrepend(prepend string) string {
	return filepath.Join(prepend, s.Path)
}
