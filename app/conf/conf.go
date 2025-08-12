package conf

import (
	"crypto/ed25519"
	"crypto/rand"
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
	"golang.org/x/crypto/ssh"
)

const (
	TIMEFORMAT = "02Jan2006-1504"
)

var (
	Version   = "DEV"
	maeveConf *MaeveConf
	once      sync.Once
)

type MaeveConf struct {
	Name          string
	SSHPrivateKey ed25519.PrivateKey
	SSHKnownHosts SSHKnownHosts
	MaeveDir      string
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
	return maeveConf
}

// Reads/Creates the config file.
func loadConfig() (*MaeveConf, error) {
	conf := &MaeveConf{}
	confPath, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(confPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, utils.Stacktrace(err, "parsing config file")
		}
		// File exists since there is no error, so unmarshal it.
	} else if err := json.Unmarshal(data, &conf); err != nil {
		return nil, utils.Stacktrace(err, "unmarshalling config file")
	}

	if err := conf.applyDefaults(); err != nil {
		return nil, err
	}
	if err := conf.SaveToFile(); err != nil {
		return nil, err
	}

	return conf, nil
}

func (c *MaeveConf) applyDefaults() error {
	if c.Name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return utils.Stacktrace(err, "getting hostname")
		}
		c.Name = hostname
	}

	if c.SSHPrivateKey == nil || c.SSHPrivateKey.Public() == nil {
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return utils.Stacktrace(err, "generating ssh key")
		}
		c.SSHPrivateKey = privateKey
	}

	if c.SSHKnownHosts.Hosts == nil {
		c.SSHKnownHosts = NewSSHKnownHosts()
	}

	if c.MaeveDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return utils.Stacktrace(err, "getting home dir")
		}
		c.MaeveDir = filepath.Join(homeDir, "Maeve")
	}

	if c.MaxBackups == 0 {
		c.MaxBackups = 5
	}

	if c.RemoteNodes == nil {
		c.RemoteNodes = make([]string, 0)
	}

	if c.BackupDirs == nil {
		c.BackupDirs = make([]string, 0)
	}

	// Ignore: MaxUpload

	return nil
}

func (c *MaeveConf) SaveToFile() error {
	confPath, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return utils.Stacktrace(err, "marshalling config")
	}

	f, err := utils.Create(confPath)
	if err != nil {
		return utils.Stacktrace(err, "creating config file")
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return utils.Stacktrace(err, "writing data to config file")
	}

	return nil
}

func configPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", utils.Stacktrace(err, "reading user dir")
	}
	confPath := filepath.Join(userDir, "maeve", "config.yml")
	return confPath, nil
}

func (c *MaeveConf) MyNode() string {
	pubKey, err := ssh.NewPublicKey(c.SSHPrivateKey.Public())
	utils.PanicOnErr("cannot generate pub key from config private key", err)
	sum := blake3.Sum512(pubKey.Marshal())
	keyHash := hex.EncodeToString(sum[:3])
	return filepath.Join(c.MaeveDir, c.Name+"_"+keyHash)
}

// TODO: REMOVE ALL OF THE FOLLOWING

// Hash file path for a given directory.
func (c *MaeveConf) MasterHashFile(node string) string {
	return filepath.Join(c.NodeDir(node), "maeve_hashmap.gob")
}

// Hash file path for a given directory.
func (c *MaeveConf) HashFile(dir string) string {
	return filepath.Join(dir, "maeve_hashes.gob")
}

func (c *MaeveConf) SelfDir() string {
	return filepath.Join(c.MaeveDir, "my_latest")
}

func (c *MaeveConf) NodeDir(node string) string {
	return filepath.Join(c.MaeveDir, "backups", node)
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
