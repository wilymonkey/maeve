package remote

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
	hs "github.com/wilymonkey/maeve/hashsums"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/utils"
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
	rpcFuncs := new(RPCFuncs)
	if err := rpc.Register(rpcFuncs); err != nil {
		return utils.Stacktrace(err, "register rpc functions")
	}

	conn := &sshPipeConn{reader: os.Stdin, writer: os.Stdout}
	rpc.ServeConn(conn)
	return nil
}

// Deprecated: Use NewNodeConn.
func New(session *ssh.Session) (*rpc.Client, error) {
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return nil, utils.WrapErr(err)
	}
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, utils.WrapErr(err)
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, utils.WrapErr(err)
	}

	if err := session.Start("maeve --server"); err != nil {
		return nil, utils.WrapErr(err)
	}

	// Report to TUI
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			overseer.Global.Send(func() tea.Msg {
				return utils.WrapErr(fmt.Errorf("remote stderr: %s", scanner.Text()))
			})
		}
	}()

	return rpc.NewClient(&sshPipeConn{reader: stdoutPipe, writer: stdinPipe}), nil
}

type RPCFuncs int

type GetDBVersionArgs struct {
	Node string
}
type GetDBVersionReply struct {
	Version *db.DBVersion
}

func (h *RPCFuncs) GetDBVersion(args *GetDBVersionArgs, reply *GetDBVersionReply) error {
	conn, err := db.Open(args.Node)
	if err != nil {
		return err
	}
	reply.Version, err = db.GetVersion(conn)
	if err != nil {
		return err
	}
	return nil
}

func (nc *NodeConn) GetDBVersion() (*db.DBVersion, error) {
	args := &GetDBVersionArgs{Node: conf.GetConf().Name}
	var reply GetDBVersionReply
	if err := nc.rpcClient.Call("RPCFuncs.GetDBVersion", args, &reply); err != nil {
		return &db.DBVersion{}, utils.Stacktrace(err, "getting rpc db version")
	}
	return reply.Version, nil
}

type TempLocationArgs struct {
	Node string
}
type TempLocationReply struct {
	Path string
}

func (h *RPCFuncs) TempLocation(args *TempLocationArgs, reply *TempLocationReply) error {
	reply.Path = conf.GetConf().NodeDirTemp(args.Node)
	return nil
}

func TempLocation(rpc *rpc.Client) (string, error) {
	args := &TempLocationArgs{Node: conf.GetConf().Name}
	var reply TempLocationReply
	if err := rpc.Call("RPCFuncs.TempLocation", args, &reply); err != nil {
		return "", utils.WrapErr(err)
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

	// path := args.Hash.Path.ResolveTemp(args.Node)
	// hashGood, err := args.Hash.Validate(path)
	// if err != nil {
	// 	return utils.WrapErr(err)
	// }
	// reply.HashGood = hashGood
	return fmt.Errorf("NOT COMPLETE")
}

func VerifyFile(rpc *rpc.Client, hash hs.FileHash) (bool, error) {
	args := &VerifyFuncArgs{
		Hash: hash,
		Node: conf.GetConf().Name,
	}
	var reply VerifyFuncReply
	if err := rpc.Call("RPCFuncs.VerifyTempfile", args, &reply); err != nil {
		return false, utils.WrapErr(err)
	}
	return reply.HashGood, nil
}

type LinkExistingArgs struct {
	Hashes []hs.FileHash
	Node   string
}
type LinkExistingReply struct {
	Hashes []hs.FileHash
}

func (h *RPCFuncs) LinkExisting(args *LinkExistingArgs, reply *LinkExistingReply) error {
	hashes, err := hs.LinkExisting(args.Node, args.Hashes)
	if err != nil {
		return utils.WrapErr(err)
	}
	reply.Hashes = hashes
	return nil
}

func LinkExisting(rpc *rpc.Client, hashes []hs.FileHash) ([]hs.FileHash, error) {
	args := &LinkExistingArgs{Node: conf.GetConf().Name, Hashes: hashes}
	var reply LinkExistingReply
	if err := rpc.Call("RPCFuncs.LinkExisting", args, &reply); err != nil {
		return nil, utils.WrapErr(err)
	}
	return reply.Hashes, nil
}

type ValiExistingArgs struct {
	Hashes []hs.FileHash
	Node   string
}
type ValiExistingReply struct {
	Hashes []hs.FileHash
}

func (h *RPCFuncs) ValiExisting(args *ValiExistingArgs, reply *ValiExistingReply) error {
	hashes, err := hs.ValidateExisting(args.Hashes, args.Node)
	if err != nil {
		return utils.WrapErr(err)
	}
	reply.Hashes = hashes
	return nil
}

func ValiExisting(rpc *rpc.Client, hashes []hs.FileHash) ([]hs.FileHash, error) {
	args := &ValiExistingArgs{
		Hashes: hashes,
		Node:   conf.GetConf().Name,
	}
	var reply ValiExistingReply
	if err := rpc.Call("RPCFuncs.ValiExisting", args, &reply); err != nil {
		return hashes, utils.WrapErr(err)
	}
	return reply.Hashes, nil
}

type FinSnapshotArgs struct {
	Node string
}

func (h *RPCFuncs) FinSnapshot(args *FinSnapshotArgs, reply *struct{}) error {
	if err := local.CullSnapshots(args.Node); err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

func FinSnapshot(rpc *rpc.Client) error {
	args := &FinSnapshotArgs{Node: conf.GetConf().Name}
	var reply struct{}
	if err := rpc.Call("RPCFuncs.FinSnapshot", args, &reply); err != nil {
		return utils.WrapErr(err)
	}
	return nil
}
