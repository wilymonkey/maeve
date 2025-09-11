package conf

import (
	"bytes"
	"compress/flate"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/wilymonkey/maeve/utils"
)

var Version = "DEV"

// =======================================
// CONFIG
// =======================================

type InMemory struct {
	DisplayName string
	ServerPort  int
	MaxBackups  int
	MaxUpload   int64
	RootDir     string
	RemoteNodes []string
	SourceDirs  []string
	PrivKey     ed25519.PrivateKey
	Cert        *x509.Certificate
}

type OnFile struct {
	DisplayName string   `yaml:"Name"`
	ServerPort  int      `yaml:"Server Port"`
	MaxBackups  int      `yaml:"Max Backups"`
	MaxUpload   int64    `yaml:"Max Upload Speed"`
	RootDir     string   `yaml:"Backup Folder"`
	SourceDirs  []string `yaml:"Source Folders"`
	RemoteNodes []string `yaml:"Backup PCs"`
	CryptoBlob  string   `yaml:"Crypto Blob"`
}

var (
	inMem *InMemory
	once  sync.Once
)

func GetConf() *InMemory {
	once.Do(func() {
		onFile, err := readFile()
		utils.AssertNoErr(err, "reading config from file")
		inMem, err = onFile.toMemory()
		utils.AssertNoErr(err, "reading config into memory")
	})
	return inMem
}

func (c *InMemory) SaveToFile() error {
	cryptoBlob, err := encodeCryptoBlob(c.PrivKey, c.Cert)
	if err != nil {
		return err
	}
	onFile := &OnFile{
		DisplayName: c.DisplayName,
		ServerPort:  c.ServerPort,
		MaxBackups:  c.MaxBackups,
		MaxUpload:   c.MaxUpload,
		RootDir:     c.RootDir,
		RemoteNodes: c.RemoteNodes,
		SourceDirs:  c.SourceDirs,
		CryptoBlob:  cryptoBlob,
	}
	return onFile.writeFile()
}

// =======================================
// OnFile Methods
// =======================================

func (c *OnFile) toMemory() (*InMemory, error) {
	privKey, cert, err := decodeCryptoBlob(c.CryptoBlob)
	if err != nil {
		return nil, err
	}

	return &InMemory{
		DisplayName: c.DisplayName,
		ServerPort:  c.ServerPort,
		MaxBackups:  c.MaxBackups,
		MaxUpload:   c.MaxUpload,
		RootDir:     c.RootDir,
		RemoteNodes: c.RemoteNodes,
		SourceDirs:  c.SourceDirs,
		PrivKey:     privKey,
		Cert:        cert,
	}, nil
}

func readFile() (*OnFile, error) {
	cfg := &OnFile{}
	confPath, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(confPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("unmarshalling config file: %w", err)
		}
	}

	if err := cfg.applyDefaults(); err != nil {
		return nil, err
	}

	if err := cfg.writeFile(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *OnFile) applyDefaults() error {
	if c.DisplayName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("getting hostname: %w", err)
		}
		c.DisplayName = hostname
	}

	if c.ServerPort < 1024 {
		c.ServerPort = 2222
	}

	if c.RootDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home dir: %w", err)
		}
		c.RootDir = filepath.Join(homeDir, "Maeve")
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

	if c.MaxUpload < 0 {
		c.MaxUpload = 0
	}

	if c.CryptoBlob == "" {
		var err error
		c.CryptoBlob, err = genCryptoBlob()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *OnFile) writeFile() error {
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

func genCryptoBlob() (string, error) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("generating private key: %w", err)
	}
	cert, err := genTLSCert(&privKey)
	if err != nil {
		return "", fmt.Errorf("generating tls cert: %w", err)
	}
	return encodeCryptoBlob(privKey, cert)
}

const cryptoBlobKey = "privKey"
const cryptoBlobCert = "cert"

func encodeCryptoBlob(
	privKey ed25519.PrivateKey,
	cert *x509.Certificate,
) (string, error) {
	m := make(map[string][]byte)
	m[cryptoBlobKey] = privKey
	m[cryptoBlobCert] = cert.Raw

	buf := new(bytes.Buffer)
	comp, _ := flate.NewWriter(buf, flate.BestCompression)

	if err := gob.NewEncoder(comp).Encode(m); err != nil {
		return "", fmt.Errorf("encoding crypto blob: %w", err)
	}

	if err := comp.Close(); err != nil {
		return "", fmt.Errorf("closing flate writer: %w", err)
	}

	s := base64.StdEncoding.EncodeToString(buf.Bytes())
	return s, nil
}

func decodeCryptoBlob(s string) (ed25519.PrivateKey, *x509.Certificate, error) {
	compressed, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, nil, fmt.Errorf("translating crypto blob: %w", err)
	}

	zr := flate.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, nil, fmt.Errorf("uncompressing crypto blob: %w", err)
	}
	defer utils.Cleanup(&err, zr.Close)

	var m map[string][]byte
	if err := gob.NewDecoder(zr).Decode(&m); err != nil {
		return nil, nil, fmt.Errorf("decoding crypto blob: %w", err)
	}

	privKey, privKeyOk := m[cryptoBlobKey]
	certDer, cerOk := m[cryptoBlobCert]
	if !privKeyOk || !cerOk {
		return nil, nil, fmt.Errorf("retrieving private key and cert from crypto blob: %w", err)
	}
	cert, err := x509.ParseCertificate(certDer)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing certDer from crypto blob: %w", err)
	}

	return privKey, cert, err
}

func genTLSCert(
	privKey *ed25519.PrivateKey,
	validAddr ...string,
) (*x509.Certificate, error) {
	var ipAddr []net.IP
	for _, ipStr := range validAddr {
		ipAddr = append(ipAddr, net.ParseIP(ipStr))
	}

	template := &x509.Certificate{
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(100, 0, 0), // Valid for 100 years
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IPAddresses: ipAddr,
	}
	certDer, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		privKey.Public().(ed25519.PublicKey),
		privKey,
	)
	if err != nil {
		return nil, fmt.Errorf("creating certDer: %w", err)
	}

	cert, err := x509.ParseCertificate(certDer)
	if err != nil {
		return nil, fmt.Errorf("parssing certDer: %w", err)
	}

	return cert, nil
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
	pubKey := conf.PrivKey.Public().(ed25519.PublicKey)
	keyHash := hex.EncodeToString(pubKey[:3])
	return fmt.Sprintf("%s_%s", conf.DisplayName, keyHash)
}

func MyNode() string {
	return filepath.Join(GetConf().RootDir, MyName())
}

func RemoteTempDir(node string) string {
	conf := GetConf()
	return filepath.Join(conf.RootDir, node, "Temp")
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
