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
	"github.com/wilymonkey/maeve/app/help"
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
	SSHPort       int
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
			return nil, help.WrapErr(err, "parsing config file")
		}
		// File exists since there is no error, so unmarshal it.
	} else if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, help.WrapErr(err, "unmarshalling config file")
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
			return help.WrapErr(err, "getting hostname")
		}
		c.DisplayName = hostname
	}

	if c.SSHPrivateKey == nil || c.SSHPrivateKey.Public() == nil {
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return help.WrapErr(err, "generating ssh key")
		}
		c.SSHPrivateKey = privateKey
	}

	if c.SSHKnownHosts.Hosts == nil {
		c.SSHKnownHosts = NewSSHKnownHosts()
	}

	if c.SSHPort < 1024 {
		c.SSHPort = 2222
	}

	if c.MaeveDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return help.WrapErr(err, "getting home dir")
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
		return help.WrapErr(err, "marshalling config")
	}

	f, err := utils.Create(confPath)
	if err != nil {
		return help.WrapErr(err, "creating config file")
	}
	defer utils.Cleanup(&err, f.Close)

	if _, err := f.Write(data); err != nil {
		return help.WrapErr(err, "writing data to config file")
	}

	return nil
}

// =======================================
// PATHS
// =======================================

func configPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", help.WrapErr(err, "reading user dir")
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
		return snapshot, help.WrapErr(err, "parsing root folder as time")
	}
	return snapshot, nil
}

func TimeFromInt64(t int64) time.Time {
	return time.Unix(t, 0).UTC()
}

func TimeNow() time.Time {
	return time.Now().UTC()
}
