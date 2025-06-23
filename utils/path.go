package utils

import (
	"fmt"
	"path/filepath"

	"github.com/wilymonkey/maeve/cfg"
)

type RelativePath struct {
	path string
}

// Resolves to absolute path.
func (r *RelativePath) Resolve(node string) string {
	return filepath.Join(cfg.Global.NodeDir(node), r.path)
}

type SnapshotPath struct {
	path string
}

func NewSnapshotPath(sourcePath, filePath string) SnapshotPath {
	relPath, err := filepath.Rel(sourcePath, filePath)
	if err != nil {
		panic(fmt.Errorf("Critical! Unable to get relative path of: %s and %s", sourcePath, filePath))
	}

	return SnapshotPath{path: relPath}
}

// Resolves to relative path.
func (s *SnapshotPath) Resolve(snapshot string) RelativePath {
	return RelativePath{path: filepath.Join(snapshot, s.path)}
}

// Resolves to an absolute path with NodeDirTemp as the base.
func (r *SnapshotPath) ResolveTemp(node string) string {
	return filepath.Join(cfg.Global.NodeDirTemp(node), r.path)
}
