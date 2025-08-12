package conf

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/crypto/ssh"
)

type SSHKnownHosts struct {
	Hosts map[string]ssh.PublicKey
}

func NewSSHKnownHosts() SSHKnownHosts {
	return SSHKnownHosts{
		Hosts: make(map[string]ssh.PublicKey),
	}
}

func (s *SSHKnownHosts) Add(hostname string, key ssh.PublicKey) {
	s.Hosts[hostname] = key
}

func (s *SSHKnownHosts) HostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		host := strings.Split(hostname, ":")[0]
		if host == "" {
			err := fmt.Errorf("cannot find host for %s", hostname)
			return utils.Stacktrace(err, "parsing hostname")
		}
		storedKey, exists := s.Hosts[host]
		if !exists {
			s.Add(hostname, key)
			return nil
		}

		if !bytes.Equal(key.Marshal(), storedKey.Marshal()) {
			err := fmt.Errorf("host key mismatch for %s: expected %s, got %s",
				host,
				ssh.FingerprintSHA256(storedKey),
				ssh.FingerprintSHA256(key),
			)
			return utils.Stacktrace(err, "validating keys")
		}

		return nil
	}
}
