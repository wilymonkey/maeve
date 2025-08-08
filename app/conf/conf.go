package conf

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/wilymonkey/maeve/utils"
	"github.com/zeebo/blake3"
)

const TIMEFORMAT = "02Jan2006-1504"

var (
	Version   = "DEV"
	maeveConf MaeveConf
	once      sync.Once
)

type MaeveConf struct {
	Name          string
	SSHKey        string
	SSHKnownHosts string
	BackupDir     string
	MaxBackups    int
	MaxUpload     int64
	RemoteNodes   []string
	BackupDirs    []string
}

func GetConf() *MaeveConf {
	once.Do(func() {
		var err error
		maeveConf, err = loadConfig()
		if err != nil {
			panic(err)
		}
	})
	return &maeveConf
}

// Hash file path for a given directory.
func (c *MaeveConf) MasterHashFile(node string) string {
	return filepath.Join(c.NodeDir(node), "maeve_hashmap.gob")
}

// Hash file path for a given directory.
func (c *MaeveConf) HashFile(dir string) string {
	return filepath.Join(dir, "maeve_hashes.gob")
}

func (c *MaeveConf) SelfDir() string {
	return filepath.Join(c.BackupDir, "my_latest")
}

func (c *MaeveConf) MyNode() string {
	sum := blake3.Sum512([]byte(c.SSHKey))
	keyHash := hex.EncodeToString(sum[:16])
	return filepath.Join(c.BackupDir, c.Name+keyHash)
}

func (c *MaeveConf) NodeDir(node string) string {
	return filepath.Join(c.BackupDir, "backups", node)
}

func (c *MaeveConf) NodeDirTemp(node string) string {
	return filepath.Join(c.NodeDir(node), "latest")
}

func (c *MaeveConf) NodeSnapshotDir(node, snapshot string) string {
	return filepath.Join(c.NodeDir(node), snapshot)
}

// Lists snapshots for a given node from oldest to newest.
// Returns absolute paths to those snapshots.
func (c *MaeveConf) NodeSnapshots(node string) ([]string, error) {
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

func (c *MaeveConf) commit(path string) error {
	data, err := json.Marshal(c)
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

func (c *MaeveConf) applyDefaults() error {
	def, err := defaultConfig()
	if err != nil {
		return utils.WrapErr(err)
	}

	if c.Name == "" {
		c.Name = def.Name
	}
	if c.SSHKey == "" {
		c.SSHKey = def.SSHKey
	}
	if c.SSHKnownHosts == "" {
		c.SSHKnownHosts = def.SSHKnownHosts
	}
	if c.BackupDir == "" {
		c.BackupDir = def.BackupDir
	}
	if c.MaxBackups == 0 {
		c.MaxBackups = def.MaxBackups
	}

	// Ignore: MaxUpload, RemoteNodes, SourceDirs

	return nil
}

// Reads/Creates the config file.
func loadConfig() (MaeveConf, error) {
	var conf MaeveConf
	userDir, err := os.UserConfigDir()
	if err != nil {
		return conf, utils.WrapErr(err)
	}
	confPath := filepath.Join(userDir, "maeve", "config.json")

	data, err := os.ReadFile(confPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return conf, utils.WrapErr(err)
		}

		// Create the config file.
		conf, err = defaultConfig()
		if err != nil {
			return conf, utils.WrapErr(err)
		}
		if err := conf.commit(confPath); err != nil {
			return conf, utils.WrapErr(err)
		}
		return conf, nil
	}

	if err := json.Unmarshal(data, &conf); err != nil {
		return conf, utils.WrapErr(err)
	}
	if err := conf.applyDefaults(); err != nil {
		return conf, utils.WrapErr(err)
	}
	if err := conf.commit(confPath); err != nil {
		return conf, utils.WrapErr(err)
	}
	return conf, nil
}

func (c *MaeveConf) SaveToFile() error {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return utils.WrapErr(err)
	}
	confPath := filepath.Join(userDir, "maeve", "config.json")
	if err := c.commit(confPath); err != nil {
		return utils.WrapErr(err)
	}
	return nil
}

func defaultConfig() (MaeveConf, error) {
	mc := MaeveConf{
		MaxBackups:  5,
		MaxUpload:   0,
		RemoteNodes: make([]string, 0),
		BackupDirs:  make([]string, 0),
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
