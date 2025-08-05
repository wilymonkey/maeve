package local

import (
	"path/filepath"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/utils"
)

type BackupRelPath struct {
	Path string
}

// Resolves to absolute path.
func (r *BackupRelPath) Resolve(node string) string {
	return filepath.Join(conf.GetConf().NodeDir(node), r.Path)
}

type SnapshotRelPath struct {
	Path string
}

func NewSnapshotPath(sourcePath, filePath string) (SnapshotRelPath, error) {
	relPath, err := filepath.Rel(sourcePath, filePath)
	if err != nil {
		return SnapshotRelPath{}, utils.WrapErr(err)
	}

	return SnapshotRelPath{Path: relPath}, nil
}

// Resolves to relative path.
func (s *SnapshotRelPath) Resolve(snapshot string) BackupRelPath {
	return BackupRelPath{Path: filepath.Join(snapshot, s.Path)}
}

// Resolves to an absolute path with NodeDirTemp as the base.
func (s *SnapshotRelPath) ResolveTemp(node string) string {
	return filepath.Join(conf.GetConf().NodeDirTemp(node), s.Path)
}

// Resolves to an absolute path with SelfDir as the base.
func (s *SnapshotRelPath) ResolveSelf() string {
	return filepath.Join(conf.GetConf().SelfDir(), s.Path)
}

// Resolves to an absolute path with SelfDir as the base.
func (s *SnapshotRelPath) ResolvePrepend(prepend string) string {
	return filepath.Join(prepend, s.Path)
}
