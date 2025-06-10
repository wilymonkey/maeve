package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Name        string   `yaml:Name`
	BackupDir   string   `yaml:"BackupDir"`
	RemoteNodes []string `yaml:"RemoteNodes"`
	SourceDirs  []string `yaml:"SourceDirs"`
}

var Global Config

// Reads the config file and makes it available globally.
func Read(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	err = yaml.Unmarshal(data, &Global)
	if err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	if len(Global.SourceDirs) == 0 {
		return fmt.Errorf("SourceDirs specified in config")
	}
	if Global.BackupDir == "" {
		return fmt.Errorf("BackupDir not specified in config")
	}

	return nil
}

func NodeDir(node string) string {
	return filepath.Join(Global.BackupDir, node, "latest")
}
func NodeDirLatest(node string) string {
	return filepath.Join(Global.BackupDir, node, "latest")
}
