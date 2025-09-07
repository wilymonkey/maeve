package conf

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/wilymonkey/maeve/utils"
	"github.com/zeebo/blake3"
	"golang.org/x/crypto/ssh"
)

var Version = "DEV"

// =======================================
// CONFIG
// =======================================

var (
	maeveConf *MaeveConf
	once      sync.Once
)

type MaeveConf struct {
	DisplayName   string
	SSHPrivateKey ed25519.PrivateKey `yaml:"sshprivatekey,flow"`
	SSHKnownHosts SSHKnownHosts
	SSHAuthKeys   []string
	ServerPort    int
	MaeveDir      string
	MaxBackups    int
	MaxUpload     int64
	RemoteNodes   []string
	SourceDirs    []string
}

func GetConf() *MaeveConf {
	once.Do(func() {
		var err error
		maeveConf, err = loadConfig()
		utils.AssertNoErr(err, "config should be parseable")
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
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
		// File exists since there is no error, so unmarshal it.
	} else if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, fmt.Errorf("unmarshalling config file: %w", err)
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
	if c.DisplayName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("getting hostname: %w", err)
		}
		c.DisplayName = hostname
	}

	if c.SSHPrivateKey == nil || c.SSHPrivateKey.Public() == nil {
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return fmt.Errorf("generating ssh key: %w", err)
		}
		c.SSHPrivateKey = privateKey
	}

	if c.SSHKnownHosts.Hosts == nil {
		c.SSHKnownHosts = NewSSHKnownHosts()
	}

	if c.ServerPort < 1024 {
		c.ServerPort = 2222
	}

	if c.MaeveDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home dir: %w", err)
		}
		c.MaeveDir = filepath.Join(homeDir, "Maeve")
	}

	if c.MaxBackups == 0 {
		c.MaxBackups = 5
	}

	if c.RemoteNodes == nil {
		c.RemoteNodes = make([]string, 0)
	}

	if c.SourceDirs == nil {
		c.SourceDirs = make([]string, 0)
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
		return fmt.Errorf("marshalling config: %w", err)
	}

	f, err := utils.Create(confPath)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer utils.Cleanup(&err, f.Close)

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing data to config file: %w", err)
	}

	return nil
}

// =======================================
// PATHS
// =======================================

func configPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("reading user dir: %w", err)
	}
	confPath := filepath.Join(userDir, "maeve", "config.yml")
	return confPath, nil
}

func MyName() string {
	conf := GetConf()
	pubKey, err := ssh.NewPublicKey(conf.SSHPrivateKey.Public())

	utils.AssertNoErr(err, "cannot generate pub key from config private key")

	sum := blake3.Sum512(pubKey.Marshal())
	keyHash := hex.EncodeToString(sum[:3])
	return fmt.Sprintf("%s_%s", conf.DisplayName, keyHash)
}

func MyNode() string {
	return filepath.Join(GetConf().MaeveDir, MyName())
}

func RemoteTempDir(node string) string {
	conf := GetConf()
	return filepath.Join(conf.MaeveDir, node, "Temp")
}

// =======================================
// TIME
// =======================================

const timeFormat = "02Jan2006-1504"

func TimeToInt64(t time.Time) int64 {
	return t.Unix()
}

func TimeToString(t time.Time) string {
	return t.Format(timeFormat)
}

func TimeFromString(s string) (time.Time, error) {
	snapshot, err := time.Parse(timeFormat, s)
	if err != nil {
		return snapshot, fmt.Errorf("parsing root folder as time: %w", err)
	}
	return snapshot, nil
}

func TimeFromInt64(t int64) time.Time {
	return time.Unix(t, 0).UTC()
}

func TimeNow() time.Time {
	return time.Now().UTC()
}
