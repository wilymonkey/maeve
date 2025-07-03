package mypath

import (
	"fmt"
	"path/filepath"

	"github.com/wilymonkey/maeve/cfg"
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

func NewSnapshotPath(sourcePath, filePath string) SnapshotPath {
	relPath, err := filepath.Rel(sourcePath, filePath)
	if err != nil {
		panic(fmt.Errorf("Critical! Unable to get relative path of: %s and %s", sourcePath, filePath))
	}

	return SnapshotPath{Path: relPath}
}

// Resolves to relative path.
func (s *SnapshotPath) Resolve(snapshot string) RelativePath {
	return RelativePath{Path: filepath.Join(snapshot, s.Path)}
}

// Resolves to an absolute path with NodeDirTemp as the base.
func (r *SnapshotPath) ResolveTemp(node string) string {
	return filepath.Join(cfg.Global.NodeDirTemp(node), r.Path)
}
