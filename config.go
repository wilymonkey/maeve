package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type MaeveConfig struct {
	Name        string   `yaml:"Name"`
	BackupDir   string   `yaml:"BackupDir"`
	MaxBackups  int      `yaml:"MaxBackups"`
	RemoteNodes []string `yaml:"RemoteNodes"`
	SourceDirs  []string `yaml:"SourceDirs"`
}

var Config MaeveConfig

// Reads the config file and makes it available globally.
func ReadConfig() error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("unable to get user config dir ⇒  %v", err)
	}
	configPath := filepath.Join(userConfigDir, "maeve", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := DefaultConfig(configPath); err != nil {
				return fmt.Errorf("unable to create default config ⇒  %v", err)
			}
			// Try to read the newly created config file.
			data, err = os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("unable to read default config ⇒  %v", err)
			}
		} else {
			return fmt.Errorf("unable to load config ⇒  %v", err)
		}
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("unable to parse ⇒  %w", err)
	}

	if Config.BackupDir == "" {
		return fmt.Errorf("BackupDir not specified")
	}
	if Config.MaxBackups < 1 {
		return fmt.Errorf("MaxBackups of %d is not valid", Config.MaxBackups)
	}

	return nil
}

// Create default config file and write it to the path given.
func DefaultConfig(configPath string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get hostname ⇒  %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to get current path ⇒  %w", err)
	}
	backupDir := filepath.Join(filepath.Dir(exePath), "backups")

	var config = MaeveConfig{
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

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("unable to create maeve config dir ⇒  %v", err)
	}

	if err := os.WriteFile(configPath, data, 0755); err != nil {
		return fmt.Errorf("unable to write config file ⇒  %w", err)
	}

	return nil
}

func (c *MaeveConfig) HashDir() string {
	return filepath.Join(c.BackupDir, "hashes")
}

func (c *MaeveConfig) SelfDir() string {
	return filepath.Join(c.BackupDir, "hardlinks")
}

func (c *MaeveConfig) NodeDir(nodeName string) string {
	return filepath.Join(c.BackupDir, "backups")
}

func (c *MaeveConfig) NodeDirTemp(nodeName string) string {
	return filepath.Join(c.NodeDir(nodeName), "temp")
}

// Gets the latest snapshot in a given node.
// CAUTION: Deletes files (not directories) found in the given node.
func (c *MaeveConfig) NodeDirLatest(nodeName string) (string, error) {
	baseDir := c.NodeDir(nodeName)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", fmt.Errorf("unable to read %s directory ⇒  %w", nodeName, err)
	}
	if len(entries) == 0 {
		return "", os.ErrNotExist
	}

	var latest os.DirEntry
	for _, dir := range entries {
		if dir.Name() > latest.Name() {
			latest = dir
		}
	}
	latestPath := filepath.Join(baseDir, latest.Name())

	if !latest.IsDir() {
		if err := os.RemoveAll(latestPath); err != nil {
			return "", fmt.Errorf("unable to delete problem file found in %s directory ⇒  %w", nodeName, err)
		}
		// Keep deleting offending files until a dir is returned.
		return c.NodeDirLatest(nodeName)
	}

	return filepath.Join(latestPath), nil
}
