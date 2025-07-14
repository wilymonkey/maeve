package cfg

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/wilymonkey/maeve/utils"
)

const VERSION = "0.0.1"
const TIMEFORMAT = "02Jan2006-1504"

var Global MaeveConfig

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

func (c *MaeveConfig) NodeSnapshotDir(node, snapshot string) string {
	return filepath.Join(c.NodeDir(node), snapshot)
}

// Lists snapshots for a given node from oldest to newest.
// Returns absolute paths to those snapshots.
func (c *MaeveConfig) NodeSnapshots(node string) ([]string, error) {
	baseDir := c.NodeDir(node)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, utils.WrapErr(err)
	}
	if len(entries) == 0 {
		return nil, utils.WrapErr(os.ErrNotExist)
	}

	var result []string
	for _, d := range entries {
		if d.IsDir() {
			result = append(result, d.Name())
		}
	}
	sort.Slice(result, func(i, j int) bool {
		ti, err1 := time.Parse(TIMEFORMAT, result[i])
		tj, err2 := time.Parse(TIMEFORMAT, result[j])
		if err1 != nil || err2 != nil {
			return result[i] < result[j]
		}
		return ti.Before(tj)
	})
	for i, snapshot := range result {
		result[i] = c.NodeSnapshotDir(node, snapshot)
	}
	return result, nil
}

// Reads the config file and makes it available globally.
func ReadConfig() error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return utils.WrapErr(err)
	}
	configPath := filepath.Join(userConfigDir, "maeve", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := DefaultConfig(configPath); err != nil {
				return utils.WrapErr(err)
			}
			// Try to read the newly created config file.
			data, err = os.ReadFile(configPath)
			if err != nil {
				return utils.WrapErr(err)
			}
		} else {
			return utils.WrapErr(err)
		}
	}

	err = yaml.Unmarshal(data, &Global)
	if err != nil {
		return utils.WrapErr(err)
	}

	if Global.BackupDir == "" {
		return utils.WrapErr(err)
	}
	if Global.MaxBackups < 1 {
		return utils.WrapErr(err)
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
