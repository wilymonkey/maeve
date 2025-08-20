package local

import (
	"os"
	"path/filepath"
	"time"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/utils"
)

const DirTimeFormat = "02Jan2006-1504"

func MkDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err // some other error (e.g., permission)
	}
	return true, nil
}

func NewLatestDir() (string, error) {
	currentTime := time.Now().Format(conf.TIMEFORMAT)
	path := filepath.Join(conf.MyNode(), currentTime)
	if err := MkDir(path); err != nil {
		return "", help.CheckBackupDir(err, "creating latest folder")
	}
	return path, nil
}

func RemoveChildDirs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return help.CheckBackupDir(err, "reading files in backup folder")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err := os.RemoveAll(filepath.Join(dir, entry.Name()))
			if err != nil {
				return help.CheckBackupDir(err, "deleting stale folders in backup")
			}
		}
	}
	return nil
}

// Deprecated.
func StampDate(node string) (string, error) {
	oldPath := conf.GetConf().NodeDirTemp(node)
	currentTime := time.Now().Format(conf.TIMEFORMAT)
	newPath := filepath.Join(filepath.Dir(oldPath), currentTime)
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", utils.WrapErr(err)
	}
	return currentTime, nil
}

func CullSnapshots(node string) error {
	snapshots, err := conf.GetConf().NodeSnapshots(node)
	if err != nil {
		return utils.WrapErr(err)
	}
	snapLen := len(snapshots)
	if snapLen > conf.GetConf().MaxBackups {
		for _, p := range snapshots[:snapLen-conf.GetConf().MaxBackups] {
			if err := os.RemoveAll(p); err != nil {
				return utils.WrapErr(err)
			}
		}
	}
	return nil
}
