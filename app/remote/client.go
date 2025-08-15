package remote

import (
	"fmt"
	"strings"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/help"
	"golang.org/x/crypto/ssh"
)

func NewSSHClient(address string) (*ssh.Client, error) {
	cfg := conf.GetConf()
	user, host, port := parseAddress(address)

	signer, err := ssh.NewSignerFromKey(cfg.SSHPrivateKey)
	if err != nil {
		return nil, help.Stacktrace(err, "parsing private key", help.DelPrivateKey)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: cfg.SSHKnownHosts.HostKeyCallback(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), config)
	if err != nil {
		return nil, help.Stacktrace(err, "dialing ssh server", help.CheckNodeConn)
	}
	return client, nil
}

func parseAddress(address string) (user, host, port string) {
	parts := strings.Split(address, "@")
	if len(parts) != 2 {
		parts = append([]string{""}, parts...)
	}
	user = parts[0]

	hostPort := strings.Split(parts[1], ":")
	if len(hostPort) != 2 {
		hostPort = append(hostPort, "22")
	}
	host = hostPort[0]
	port = hostPort[1]

	return user, host, port
}
