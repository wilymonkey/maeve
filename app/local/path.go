package local

import (
	"path/filepath"
	"strings"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/utils"
)

type AbsPath struct {
	Node         string
	SnapshotPath string
	RelPath      string
}

func NewAbsPath(path string) AbsPath {
	rel, err := filepath.Rel(conf.GetConf().BackupDir, path)
	if err != nil {
		panic(utils.WrapErr(err))
	}

	parts := strings.Split(filepath.ToSlash(filepath.Clean(rel)), "/")

	a := AbsPath{}
	if len(parts) > 0 {
		a.Node = parts[0]
	}
	if len(parts) > 1 {
		a.SnapshotPath = parts[1]
	}
	if len(parts) > 2 {
		a.RelPath = filepath.Join(parts[2:]...)
	}

	return a
}

func (a *AbsPath) Path() string {
	return filepath.Join(conf.GetConf().NodeDir(a.Node), a.SnapshotPath, a.RelPath)
}

type RelPath struct {
	Snapshot string
	Path     string
}

func (r *RelPath) Resolve(node string) string {
	return filepath.Join(conf.GetConf().NodeDir(node), r.Snapshot, r.Path)
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
func (s *SnapshotRelPath) Resolve(snapshot string) RelPath {
	return RelPath{Path: filepath.Join(snapshot, s.Path)}
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
