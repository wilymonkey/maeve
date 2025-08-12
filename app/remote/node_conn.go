package remote

import (
	"bufio"
	"net/rpc"

	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/crypto/ssh"
)

type NodeConn struct {
	sshClient  *ssh.Client
	sshSession *ssh.Session
	rpcClient  *rpc.Client
}

func (n *NodeConn) Close() {
	n.rpcClient.Close()
	n.sshSession.Close()
	n.sshClient.Close()
}

func NewNodeConn(node string, onStderr func(err string)) (*NodeConn, error) {
	sshClient, err := NewSSHClient(node)
	if err != nil {
		return nil, err
	}
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, utils.Stacktrace(err, "creating ssh session")
	}

	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		return nil, utils.Stacktrace(err, "getting stdin pipe")
	}
	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		return nil, utils.Stacktrace(err, "getting stdout pipe")
	}
	stderrPipe, err := sshSession.StderrPipe()
	if err != nil {
		return nil, utils.Stacktrace(err, "getting stderr pipe")
	}

	if err := sshSession.Start("maeve --server"); err != nil {
		return nil, utils.Stacktrace(err, "starting remote maeve as server")
	}

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			onStderr(scanner.Text())
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
