package cfg

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/wilymonkey/maeve/utils"
)

var Version = "DEV"

const TIMEFORMAT = "02Jan2006-1504"

var Global MaeveConfig

type MaeveConfig struct {
	Name          string   `yaml:"Name"`
	SSHKey        string   `yaml:"SSHKey"`
	SSHKnownHosts string   `yaml:"SSHKnownHosts"`
	BackupDir     string   `yaml:"BackupDir"`
	MaxBackups    int      `yaml:"MaxBackups"`
	MaxUpload     int64    `yaml:"MaxUpload"`
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

func (c *MaeveConfig) Close() {
}

func (c *MaeveConfig) commit(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return utils.WrapErr(err)
	}

	f, err := utils.Create(path)
	if err != nil {
		return utils.WrapErr(err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return utils.WrapErr(err)
	}
	return nil
}

func (cfg *MaeveConfig) applyDefaults() error {
	def, err := defaultConfig()
	if err != nil {
		return utils.WrapErr(err)
	}

	if cfg.Name == "" {
		cfg.Name = def.Name
	}
	if cfg.SSHKey == "" {
		cfg.SSHKey = def.SSHKey
	}
	if cfg.SSHKnownHosts == "" {
		cfg.SSHKnownHosts = def.SSHKnownHosts
	}
	if cfg.BackupDir == "" {
		cfg.BackupDir = def.BackupDir
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = def.MaxBackups
	}

	// Ignore: MaxUpload, RemoteNodes, SourceDirs

	return nil
}

// Reads/Creates the config file and makes it available globally.
func GetConfig() error {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return utils.WrapErr(err)
	}
	cfgPath := filepath.Join(userDir, "maeve", "config.yaml")

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return utils.WrapErr(err)
		}

		// Create the config file.
		Global, err = defaultConfig()
		if err != nil {
			return utils.WrapErr(err)
		}
		if err := Global.commit(cfgPath); err != nil {
			return utils.WrapErr(err)
		}
		return nil
	}

	if err := yaml.Unmarshal(data, &Global); err != nil {
		return utils.WrapErr(err)
	}
	if err := Global.applyDefaults(); err != nil {
		return utils.WrapErr(err)
	}
	if err := Global.commit(cfgPath); err != nil {
		return utils.WrapErr(err)
	}
	return nil
}

func defaultConfig() (MaeveConfig, error) {
	mc := MaeveConfig{
		MaxBackups:  5,
		MaxUpload:   0,
		RemoteNodes: make([]string, 0),
		SourceDirs:  make([]string, 0),
	}
	hostname, err := os.Hostname()
	if err != nil {
		return mc, utils.WrapErr(err)
	}
	mc.Name = hostname

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return mc, utils.WrapErr(err)
	}
	mc.BackupDir = filepath.Join(homeDir, "Maeve")

	sshDir := filepath.Join(homeDir, ".ssh")
	sshKey, err := findSSHKeys(sshDir)
	if err != nil {
		return mc, utils.WrapErr(err)
	}
	mc.SSHKey = sshKey

	sshKnownHosts := filepath.Join(sshDir, "known_hosts")
	_, err = os.Stat(sshKnownHosts)
	if err != nil {
		return mc, utils.WrapErr(err)
	}
	mc.SSHKnownHosts = sshKnownHosts

	return mc, nil
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
