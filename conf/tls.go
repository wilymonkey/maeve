package conf

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// CertGenerator handles certificate generation and management
type CertGenerator struct {
	keyType      string // "ed25519" or "ecdsa"
	certLifetime time.Duration
	keyDir       string
}

// NewCertGenerator creates a new certificate generator
func NewCertGenerator(keyType string, certLifetime time.Duration, keyDir string) *CertGenerator {
	return &CertGenerator{
		keyType:      keyType,
		certLifetime: certLifetime,
		keyDir:       keyDir,
	}
}

// KeyPair represents a generated key pair and certificate
type KeyPair struct {
	PrivateKey  []byte
	Certificate *x509.Certificate
	SPKIHash    string // SHA-256 hash of the public key in hex format
}

// GenerateKeyPair generates a new key pair and self-signed certificate
func (cg *CertGenerator) GenerateKeyPair(hostnames []string, ips []string) (*KeyPair, error) {
	var priv crypto.Signer
	var err error

	// Generate private key based on selected algorithm
	switch cg.keyType {
	case "ed25519":
		_, priv, err = ed25519.GenerateKey(rand.Reader)
	case "ecdsa":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", cg.keyType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"gRPC App"},
			CommonName:   "gRPC Mutual TLS",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(cg.certLifetime),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  true, // Self-signed, so it's a CA
	}

	// Add hostnames and IPs to certificate
	for _, h := range hostnames {
		template.DNSNames = append(template.DNSNames, h)
	}
	for _, ip := range ips {
		if parsedIP := net.ParseIP(ip); parsedIP != nil {
			template.IPAddresses = append(template.IPAddresses, parsedIP)
		}
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated certificate: %w", err)
	}

	// Calculate SPKI hash (SHA-256 of public key)
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	spkiHash := sha256.Sum256(pubKeyBytes)

	// Marshal private key to PKCS#8 format
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	return &KeyPair{
		PrivateKey:  privBytes,
		Certificate: cert,
		SPKIHash:    hex.EncodeToString(spkiHash[:]),
	}, nil
}

func (cg *CertGenerator) checkFilePermissions(path string, expected fs.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm() != expected {
		return fmt.Errorf("invalid file permissions for %s: expected %o, got %o",
			path, expected, info.Mode().Perm())
	}
	return nil
}

// TLSCredentials creates gRPC TLS credentials from the key pair
func (keyPair *KeyPair) TLSCredentials() (credentials.TransportCredentials, error) {
	// Parse private key
	privKey, err := x509.ParsePKCS8PrivateKey(keyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(keyPair.Certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Create TLS certificate
	tlsCert := tls.Certificate{
		Certificate: [][]byte{keyPair.Certificate},
		PrivateKey:  privKey,
		Leaf:        cert,
	}

	// Create TLS config with mutual authentication
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    x509.NewCertPool(),
		RootCAs:      x509.NewCertPool(),
		MinVersion:   tls.VersionTLS13,
	}

	// Add our own CA to both client and server pools (since we're self-signed)
	tlsConfig.ClientCAs.AddCert(keyPair.Certificate)
	tlsConfig.RootCAs.AddCert(keyPair.Certificate)

	return credentials.NewTLS(tlsConfig), nil
}

// SPKIPinVerifier creates a certificate verifier that pins SPKI hashes
func SPKIPinVerifier(allowedSPKIs map[string]string) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return errors.New("no certificates presented")
		}

		// Parse the leaf certificate
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}

		// Marshal public key to SPKI format
		pubKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to marshal public key: %w", err)
		}

		// Calculate SPKI hash
		spkiHash := sha256.Sum256(pubKeyBytes)
		spkiHashStr := hex.EncodeToString(spkiHash[:])

		// Check if SPKI is in allowed list
		if _, allowed := allowedSPKIs[spkiHashStr]; !allowed {
			return fmt.Errorf("certificate SPKI hash %s not in allowed list", spkiHashStr)
		}

		return nil
	}
}

func example_use() {
	// Configuration
	config := map[string]any{
		"key_type":      "ed25519", // or "ecdsa"
		"cert_lifetime": 365 * 24 * time.Hour,
		"key_dir":       "./tls-keys",
		"hostnames":     []string{"localhost", "grpc-app"},
		"ips":           []string{"127.0.0.1", "::1"},
	}

	// Create certificate generator
	generator := NewCertGenerator(
		config["key_type"].(string),
		config["cert_lifetime"].(time.Duration),
		config["key_dir"].(string),
	)

	// Generate or load key pair
	var keyPair *KeyPair
	var err error

	if _, err := os.Stat(filepath.Join(config["key_dir"].(string), "key.der")); os.IsNotExist(err) {
		// Generate new key pair
		fmt.Println("Generating new TLS key pair...")
		keyPair, err = generator.GenerateKeyPair(
			config["hostnames"].([]string),
			config["ips"].([]string),
		)
		if err != nil {
			panic(fmt.Errorf("failed to generate key pair: %w", err))
		}

		fmt.Printf("Generated new key pair with SPKI: %s\n", keyPair.SPKIHash)
	} else {
		// Load existing key pair
		fmt.Println("Loading existing TLS key pair...")
		keyPair, err = generator.LoadKeyPair()
		if err != nil {
			panic(fmt.Errorf("failed to load key pair: %w", err))
		}
		fmt.Printf("Loaded key pair with SPKI: %s\n", keyPair.SPKIHash)
	}

	// Create TLS credentials for gRPC
	tlsCreds, err := keyPair.TLSCredentials()
	if err != nil {
		panic(fmt.Errorf("failed to create TLS credentials: %w", err))
	}

	// Example: Create gRPC server with TLS
	server := grpc.NewServer(grpc.Creds(tlsCreds))
	fmt.Printf("gRPC server configured with mutual TLS\n")

	// Example: Create gRPC client with TLS
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		panic(fmt.Errorf("failed to create client connection: %w", err))
	}
	defer conn.Close()

	fmt.Printf("gRPC client configured with mutual TLS\n")

	// Store the server and connection for later use
	_ = server
	_ = conn
}

// KeyManager for managing allowed/denied SPKI hashes
type KeyManager struct {
	allowedKeys map[string]string // SPKI hash -> role
	deniedKeys  map[string]bool   // SPKI hash -> true if denied
}

func NewKeyManager() *KeyManager {
	return &KeyManager{
		allowedKeys: make(map[string]string),
		deniedKeys:  make(map[string]bool),
	}
}

func (km *KeyManager) AllowKey(spkiHash, role string) {
	km.allowedKeys[spkiHash] = role
	delete(km.deniedKeys, spkiHash)
}

func (km *KeyManager) DenyKey(spkiHash string) {
	km.deniedKeys[spkiHash] = true
	delete(km.allowedKeys, spkiHash)
}

func (km *KeyManager) IsAllowed(spkiHash string) bool {
	_, allowed := km.allowedKeys[spkiHash]
	_, denied := km.deniedKeys[spkiHash]
	return allowed && !denied
}

func (km *KeyManager) GetRole(spkiHash string) (string, bool) {
	role, exists := km.allowedKeys[spkiHash]
	return role, exists
}
