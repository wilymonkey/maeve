package backup

import (
	"context"

	"github.com/pkg/sftp"
	"github.com/wilymonkey/maeve/cfg"
	hs "github.com/wilymonkey/maeve/hashsums"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/remote"
	"github.com/wilymonkey/maeve/rpc"
	"github.com/wilymonkey/maeve/utils"
)

type doneSend struct{}

type pushingNode struct {
	index int
}

func pushChanges(ctx context.Context, hashes []hs.FileHash) error {
	hashGob, err := hs.GetSelfHashGob()
	if err != nil {
		return utils.WrapErr(err)
	}
	hashes = append(hashes, hashGob)

	for i, node := range cfg.Global.RemoteNodes {
		overseer.Global.Send(pushingNode{index: i})

		if err := pushToNode(ctx, hashes, node); err != nil {
			return utils.WrapErr(err)
		}
	}

	overseer.Global.Send(doneSend{})
	return nil
}

type pushProgress struct {
	operations *utils.UniqueCircSlice[hs.FileHash, remote.SendStatus]
	curr       int
	total      int
}

func newPushProgress(total int) pushProgress {
	return pushProgress{
		operations: utils.NewUniqueCircSlice(
			5,
			func(item remote.SendStatus) hs.FileHash { return item.Hash },
		),
		total: total,
	}
}

func pushToNode(ctx context.Context, hashes []hs.FileHash, node string) error {
	sshClient, err := remote.NewSSHClient(node)
	if err != nil {
		return utils.WrapErr(err)
	}
	defer sshClient.Close()
	rpcSesh, err := sshClient.NewSession()
	if err != nil {
		return utils.WrapErr(err)
	}
	defer rpcSesh.Close()
	rpcClient, err := rpc.New(rpcSesh)
	if err != nil {
		return utils.WrapErr(err)
	}
	defer rpcClient.Close()
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return utils.WrapErr(err)
	}
	defer sftpClient.Close()

	progChan := make(chan remote.SendStatus, 100)
	defer close(progChan)
	pushProg := newPushProgress(len(hashes))

	utils.Throttle(
		progChan,
		func(status remote.SendStatus) {
			if status.IsGood {
				pushProg.curr++
			}
			pushProg.operations.Push(status)
		},
		func() {
			overseer.Global.Send(pushProg)
		})

	currHashes, err := rpc.ValiExisting(rpcClient, hashes)
	if err != nil {
		return utils.WrapErr(err)
	}
	currSet := utils.SliceToSet(currHashes)
	for _, h := range hashes {
		if _, exists := currSet[h]; !exists {
			status, err := remote.NewDoneStatus(h)
			if err != nil {
				return utils.WrapErr(err)
			}
			progChan <- status
		}
	}
	hashes = currHashes

	if err = rpc.LoadMaster(rpcClient); err != nil {
		return utils.WrapErr(err)
	}

	if err := remote.SendFiles(hashes, sftpClient, rpcClient, ctx, progChan); err != nil {
		return utils.WrapErr(err)
	}
	return nil
}
