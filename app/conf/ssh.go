package conf

import (
	"bytes"
	"net"
	"strings"

	"github.com/wilymonkey/maeve/app/help"
	"golang.org/x/crypto/ssh"
)

type SSHKnownHosts struct {
	Hosts map[string][]byte
}

func NewSSHKnownHosts() SSHKnownHosts {
	return SSHKnownHosts{
		Hosts: make(map[string][]byte),
	}
}

func (s *SSHKnownHosts) Add(hostname string, key ssh.PublicKey) {
	s.Hosts[hostname] = key.Marshal()
}

func (s *SSHKnownHosts) HostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		host := strings.Split(hostname, ":")[0]
		if host == "" {
			return help.Errorf("parsing hostname", "cannot find host for %s", hostname)
		}
		storedKey, exists := s.Hosts[host]
		if !exists {
			s.Add(hostname, key)
			return nil
		}

		if !bytes.Equal(key.Marshal(), storedKey) {
			return help.Errorf("validating keys", "host key mismatch for %s", host)
		}

		return nil
	}
}
