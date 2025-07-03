package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/goccy/go-yaml"
)

const VERSION = "0.0.1"

var Global MaeveConfig
var TuiProgram *tea.Program
var TuiInteractive bool

type MaeveConfig struct {
	Name          string   `yaml:"Name"`
	SSHKey        string   `yaml:"SSHKey"`
	SSHKnownHosts string   `yaml:"SSHKnownHosts"`
	BackupDir     string   `yaml:"BackupDir"`
	MaxBackups    int      `yaml:"MaxBackups"`
	RemoteNodes   []string `yaml:"RemoteNodes"`
	SourceDirs    []string `yaml:"SourceDirs"`
}

// Hash file path for a given directory.
func (c *MaeveConfig) MasterHashFile(node string) string {
	return filepath.Join(c.NodeDir(node), "maeve_hashmap.gob")
}

// Hash file path for a given directory.
func (c *MaeveConfig) HashFile(dir string) string {
	return filepath.Join(dir, "maeve_hashes.gob")
}

func (c *MaeveConfig) SelfDir() string {
	return filepath.Join(c.BackupDir, "my_files")
}

func (c *MaeveConfig) NodeDir(node string) string {
	return filepath.Join(c.BackupDir, "backups", node)
}

func (c *MaeveConfig) NodeDirTemp(node string) string {
	return filepath.Join(c.NodeDir(node), "temp")
}

// Gets the latest snapshot in a given node.
// CAUTION: Deletes files (not directories) found in the given node.
func (c *MaeveConfig) NodeDirLatest(node string) (string, error) {
	baseDir := c.NodeDir(node)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", fmt.Errorf("unable to read %s directory ⇒  %w", node, err)
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
			return "", fmt.Errorf("unable to delete problem file found in %s directory ⇒  %w", node, err)
		}
		// Keep deleting offending files until a dir is returned.
		return c.NodeDirLatest(node)
	}

	return filepath.Join(latestPath), nil
}

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

	err = yaml.Unmarshal(data, &Global)
	if err != nil {
		return fmt.Errorf("unable to parse ⇒  %w", err)
	}

	if Global.BackupDir == "" {
		return fmt.Errorf("BackupDir not specified")
	}
	if Global.MaxBackups < 1 {
		return fmt.Errorf("MaxBackups of %d is not valid", Global.MaxBackups)
	}

	return nil
}

// Create default config file and write it to the path given.
func DefaultConfig(configPath string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("get hostname ⇒  %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir ⇒  %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	sshKey, err := findSSHKeys(sshDir)
	if err != nil {
		return fmt.Errorf("find a private ssh key ⇒  %w", err)
	}
	sshKnownHosts := filepath.Join(sshDir, "known_hosts")
	_, err = os.Stat(sshKnownHosts)
	if err != nil {
		return fmt.Errorf("find a ssh known hosts ⇒  %w", err)
	}

	var config = MaeveConfig{
		Name:          hostname,
		BackupDir:     filepath.Join(homeDir, "Maeve"),
		SSHKey:        sshKey,
		SSHKnownHosts: sshKnownHosts,
		MaxBackups:    5,
		RemoteNodes:   make([]string, 0),
		SourceDirs:    make([]string, 0),
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("convert config to yaml ⇒  %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("create config dir ⇒  %v", err)
	}

	if err := os.WriteFile(configPath, data, 0755); err != nil {
		return fmt.Errorf("unable to write config file ⇒  %w", err)
	}

	return nil
}

func findSSHKeys(sshDir string) (string, error) {
	var key string

	err := filepath.WalkDir(sshDir, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := dir.Name()
		if !dir.IsDir() {
			if isID, _ := filepath.Match("id_*", name); isID {
				if isPub, _ := filepath.Match("*.pub", name); !isPub {
					key = path
					return filepath.SkipDir
				}
			}
		}

		return nil
	})

	if key == "" {
		return "", os.ErrNotExist
	}
	return key, err
}
