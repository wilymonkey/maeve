package local

import (
	"os"
	"path/filepath"
	"time"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/utils"
)

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
