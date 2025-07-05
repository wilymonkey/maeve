package rpc

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	hs "github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/utils/myerr"
	"golang.org/x/crypto/ssh"
)

type sshPipeConn struct {
	reader io.Reader
	writer io.WriteCloser
}

func (c *sshPipeConn) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (c *sshPipeConn) Write(b []byte) (n int, err error) {
	return c.writer.Write(b)
}

func (c *sshPipeConn) Close() error {
	return c.writer.Close()
}

func (c *sshPipeConn) LocalAddr() net.Addr                { return nil }
func (c *sshPipeConn) RemoteAddr() net.Addr               { return nil }
func (c *sshPipeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *sshPipeConn) SetWriteDeadline(t time.Time) error { return nil }
func (c *sshPipeConn) SetDeadline(t time.Time) error      { return nil }

func RunServer() error {
	hashFuncs := new(RPCFuncs)
	if err := rpc.Register(hashFuncs); err != nil {
		return myerr.WrapErr(err)
	}

	conn := &sshPipeConn{reader: os.Stdin, writer: os.Stdout}
	rpc.ServeConn(conn)
	return nil
}

func New(session *ssh.Session) (*rpc.Client, error) {
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return nil, myerr.WrapErr(err)
	}
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, myerr.WrapErr(err)
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, myerr.WrapErr(err)
	}

	if err := session.Start("maeve --server"); err != nil {
		return nil, myerr.WrapErr(err)
	}

	// Report to TUI
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			cfg.TuiProgram.Send(func() tea.Msg {
				return myerr.WrapErr(fmt.Errorf("remote stderr: %s", scanner.Text()))
			})
		}
	}()

	return rpc.NewClient(&sshPipeConn{reader: stdoutPipe, writer: stdinPipe}), nil
}

type RPCFuncs int

type NodeFuncArgs struct {
	Node string
}
type NodeFuncReply struct {
	Path string
}

func (h *RPCFuncs) NodeLocation(args *NodeFuncArgs, reply *NodeFuncReply) error {
	reply.Path = cfg.Global.NodeDirTemp(args.Node)
	return nil
}

func SelfNodeTempLoc(rpc *rpc.Client) (string, error) {
	args := &NodeFuncArgs{Node: cfg.Global.Name}
	var reply NodeFuncReply
	if err := rpc.Call("RPCFuncs.NodeLocation", args, &reply); err != nil {
		return "", myerr.WrapErr(err)
	}
	return reply.Path, nil
}

type VerifyFuncArgs struct {
	Hash hs.FileHash
	Node string
}
type VerifyFuncReply struct {
	HashGood bool
}

func (h *RPCFuncs) VerifyTempfile(args *VerifyFuncArgs, reply *VerifyFuncReply) error {
	path := args.Hash.Path.ResolveTemp(args.Node)
	newHash, err := hs.GenHash(path)
	if err != nil {
		return myerr.WrapErr(err)
	}
	reply.HashGood = newHash == args.Hash.Hash
	return nil
}

func VerifyFile(rpc *rpc.Client, hash hs.FileHash) (bool, error) {
	args := &VerifyFuncArgs{
		Hash: hash,
		Node: cfg.Global.Name,
	}
	var reply VerifyFuncReply
	if err := rpc.Call("RPCFuncs.VerifyTempFile", args, &reply); err != nil {
		return false, myerr.WrapErr(err)
	}
	return reply.HashGood, nil
}
