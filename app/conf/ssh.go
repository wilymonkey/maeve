package conf

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/wilymonkey/maeve/help"
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
			return help.Stacktrace(err, "parsing hostname", help.DelKnownHost)
		}
		storedKey, exists := s.Hosts[host]
		if !exists {
			s.Add(hostname, key)
			return nil
		}

		if !bytes.Equal(key.Marshal(), storedKey.Marshal()) {
			err := fmt.Errorf("host key mismatch for %s", host)
			return help.Stacktrace(err, "validating keys", help.DelKnownHost)
		}

		return nil
	}
}
