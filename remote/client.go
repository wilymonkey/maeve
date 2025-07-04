package remote

import (
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local/hashsums"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func NewSSHClient(address string) (*ssh.Client, error) {
	user, host, port, err := parseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("parse node address ⇒  %w", err)
	}

	key, err := os.ReadFile(cfg.Global.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("read ssh key ⇒  %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse ssh key ⇒  %w", err)
	}

	hostKeyCallback, err := knownhosts.New(cfg.Global.SSHKnownHosts)
	if err != nil {
		return nil, fmt.Errorf("create host key callback ⇒  %w", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
	}

	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		return nil, fmt.Errorf("dial ssh ⇒  %w", err)
	}
	return client, nil
}

func parseAddress(address string) (user, host, port string, err error) {
	parts := strings.Split(address, "@")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid format: missing @")
	}
	user = parts[0]

	hostPort := strings.Split(parts[1], ":")
	if len(hostPort) != 2 {
		hostPort = append(hostPort, "22")
	}
	host = hostPort[0]
	port = hostPort[1]

	return user, host, port, nil
}

func remoteHashes(client *ssh.Client) ([]hashsums.FileHash, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new ssh session ⇒  %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("getting stdOut pipe for session ⇒  %w", err)
	}

	err = session.Start(fmt.Sprintf("maeve --latest-hashes %s", cfg.Global.Name))
	if err != nil {
		return nil, fmt.Errorf("start maeve through ssh ⇒  %w", err)
	}

	var hashes []hashsums.FileHash
	err = gob.NewDecoder(stdout).Decode(&hashes)
	if err != nil {
		return nil, fmt.Errorf("decode stdout ⇒  %w", err)
	}

	if err = session.Wait(); err != nil {
		return nil, fmt.Errorf("session wait for cmd ⇒  %w", err)
	}

	return hashes, nil
}
