package conf

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/wilymonkey/maeve/help"
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
	SSHPrivateKey ed25519.PrivateKey `yaml:"sshprivatekey,flow"`
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
		utils.AssertNoErr("config should be parseable", err)
	})
	return maeveConf
}

func (c *MaeveConf) PublicKey() ed25519.PublicKey {
	return c.SSHPrivateKey.Public().(ed25519.PublicKey)
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
			return nil, help.Stacktrace(err, "parsing config file", help.DelConfig)
		}
		// File exists since there is no error, so unmarshal it.
	} else if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, help.Stacktrace(err, "unmarshalling config file", help.DelConfig)
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
			return help.Stacktrace(err, "getting hostname", help.DevError)
		}
		c.Name = hostname
	}

	if c.SSHPrivateKey == nil || c.SSHPrivateKey.Public() == nil {
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return help.Stacktrace(err, "generating ssh key", help.DevError)
		}
		c.SSHPrivateKey = privateKey
	}

	if c.SSHKnownHosts.Hosts == nil {
		c.SSHKnownHosts = NewSSHKnownHosts()
	}

	if c.MaeveDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return help.Stacktrace(err, "getting home dir", help.DevError)
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

	data, err := yaml.Marshal(c)
	if err != nil {
		return help.Stacktrace(err, "marshalling config", help.DevError)
	}

	f, err := utils.Create(confPath)
	if err != nil {
		return help.Stacktrace(err, "creating config file", help.DelConfig)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return help.Stacktrace(err, "writing data to config file", help.DelConfig)
	}

	return nil
}

func configPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", help.Stacktrace(err, "reading user dir", help.DevError)
	}
	confPath := filepath.Join(userDir, "maeve", "config.yml")
	return confPath, nil
}

func (c *MaeveConf) MyNode() string {
	pubKey, err := ssh.NewPublicKey(c.SSHPrivateKey.Public())
	utils.AssertNoErr("cannot generate pub key from config private key", err)
	sum := blake3.Sum512(pubKey.Marshal())
	keyHash := hex.EncodeToString(sum[:3])
	return filepath.Join(c.MaeveDir, fmt.Sprintf("%s_%s", c.Name, keyHash))
}

func (c *MaeveConf) NodeDir(node string) string {
	return filepath.Join(c.MaeveDir, node)
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
