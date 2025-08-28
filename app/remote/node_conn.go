package remote

import (
	"bufio"
	"errors"
	"net/rpc"

	"github.com/pkg/sftp"
	"github.com/wilymonkey/maeve/app/help"
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
		return nil, help.WrapError(err, "creating ssh session")
	}

	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		return nil, help.WrapError(err, "getting stdin pipe")
	}
	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		return nil, help.WrapError(err, "getting stdout pipe")
	}
	stderrPipe, err := sshSession.StderrPipe()
	if err != nil {
		return nil, help.WrapError(err, "getting stderr pipe")
	}

	if err := sshSession.Start("maeve --server"); err != nil {
		return nil, help.WrapError(err, "starting remote maeve as server")
	}

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			err := errors.New(scanner.Text())
			onStderr(help.WrapError(err, "scanning error pipe"))
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

func (n *NodeConn) addSFTP() error {
	var err error
	n.sftpClient, err = sftp.NewClient(n.sshClient)
	if err != nil {
		return help.WrapError(err, "getting SFTP client")
	}
	return nil
}
