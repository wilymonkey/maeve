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
	MaxBackups  int      `yaml:"MaxBackups"`
	RemoteNodes []string `yaml:"RemoteNodes"`
	SourceDirs  []string `yaml:"SourceDirs"`
}

var Cfg Config

// Reads the config file and makes it available globally.
func ReadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read ⇒  %w", err)
	}

	err = yaml.Unmarshal(data, &Cfg)
	if err != nil {
		return fmt.Errorf("unable to parse ⇒  %w", err)
	}

	if Cfg.BackupDir == "" {
		return fmt.Errorf("BackupDir not specified")
	}
	if Cfg.MaxBackups < 1 {
		return fmt.Errorf("MaxBackups of %d is not valid", Cfg.MaxBackups)
	}

	return nil
}

// Create default config file and write it to the path given.
func DefaultConfig(path string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get hostname ⇒  %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to get current path ⇒  %w", err)
	}
	backupDir := filepath.Join(filepath.Dir(exePath), "backups")

	var config = Config{
		Name:        hostname,
		BackupDir:   backupDir,
		MaxBackups:  5,
		RemoteNodes: make([]string, 0),
		SourceDirs:  make([]string, 0),
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("unable to convert struct to yaml ⇒  %w", err)
	}
	if err := os.WriteFile(path, data, 0755); err != nil {
		return fmt.Errorf("unable to write config file ⇒  %w", err)
	}

	return nil
}

func NodeDir(nodeName string) string {
	return filepath.Join(Cfg.BackupDir, nodeName)
}

func NodeDirLatest(nodeName string) string {
	return filepath.Join(NodeDir(nodeName), "latest")
}
