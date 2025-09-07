package remote

import (
	"bufio"
	"errors"
	"fmt"
	"net/rpc"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type NodeConn struct {
	sshClient  *ssh.Client
	sshSession *ssh.Session
	rpcClient  *rpc.Client
	sftpClient *sftp.Client
	backupDir  string
}

func (n *NodeConn) Close() error {
	if n.sftpClient != nil {
		if err := n.sftpClient.Close(); err != nil {
			return err
		}
	}
	if err := n.rpcClient.Close(); err != nil {
		return err
	}
	if err := n.sshSession.Close(); err != nil {
		return err
	}
	if err := n.sshClient.Close(); err != nil {
		return err
	}
	return nil
}

func NewNodeConn(node string, onStderr func(err error)) (*NodeConn, error) {
	sshClient, err := NewSSHClient(node)
	if err != nil {
		return nil, err
	}
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("creating ssh session: %w", err)
	}

	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("getting stdin pipe: %w", err)
	}
	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("getting stdout pipe: %w", err)
	}
	stderrPipe, err := sshSession.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("getting stderr pipe: %w", err)
	}

	if err := sshSession.Start("maeve --server"); err != nil {
		return nil, fmt.Errorf("starting remote maeve as server: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			err := errors.New(scanner.Text())
			onStderr(fmt.Errorf("scanning error pipe: %w", err))
		}
	}()

	rpcClient := rpc.NewClient(&sshPipeConn{
		reader: stdoutPipe,
		writer: stdinPipe,
	})

	return &NodeConn{
		sshClient:  sshClient,
		sshSession: sshSession,
		rpcClient:  rpcClient,
	}, nil
}

// If connection already exists, this will close and open a new one.
func (n *NodeConn) addSFTP() error {
	if n.sftpClient != nil {
		n.sftpClient.Close()
	}

	var err error
	n.sftpClient, err = sftp.NewClient(n.sshClient)
	if err != nil {
		return fmt.Errorf("getting SFTP client: %w", err)
	}
	return nil
}
