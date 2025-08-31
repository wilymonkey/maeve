package remote

import (
	"fmt"
	"log"
	"strings"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
	"golang.org/x/crypto/ssh"
)

func NewSSHClient(address string) (*ssh.Client, error) {
	cfg := conf.GetConf()
	user, host, port := parseAddress(address)

	signer, err := ssh.NewSignerFromKey(cfg.SSHPrivateKey)
	if err != nil {
		return nil, help.WrapError(err, "parsing private key")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: cfg.SSHKnownHosts.HostKeyCallback(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), config)
	if err != nil {
		log.Printf("err: %v", err)
		return nil, help.WrapError(err, "dialing ssh server")
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
