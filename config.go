package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Name        string   `yaml:"Name"`
	BackupDir   string   `yaml:"BackupDir"`
	RemoteNodes []string `yaml:"RemoteNodes"`
	SourceDirs  []string `yaml:"SourceDirs"`
}

var Cfg Config

// Reads the config file and makes it available globally.
func ReadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Reading config file ⇒  %w", err)
	}

	err = yaml.Unmarshal(data, &Cfg)
	if err != nil {
		return fmt.Errorf("Parsing config file ⇒  %w", err)
	}

	if len(Cfg.SourceDirs) == 0 {
		return fmt.Errorf("SourceDirs specified in config")
	}
	if Cfg.BackupDir == "" {
		return fmt.Errorf("BackupDir not specified in config")
	}

	return nil
}

func NodeDir(node string) string {
	return filepath.Join(Cfg.BackupDir, node, "latest")
}

func NodeDirLatest(node string) string {
	return filepath.Join(Cfg.BackupDir, node, "latest")
}
