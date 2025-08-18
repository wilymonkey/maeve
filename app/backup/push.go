package backup

import (
	"context"

	"github.com/pkg/sftp"
	"github.com/wilymonkey/maeve/app/conf"
	hs "github.com/wilymonkey/maeve/app/hashsums"
	"github.com/wilymonkey/maeve/app/remote"
	"github.com/wilymonkey/maeve/utils"
)

type doneSend struct{}

type pushingNode struct {
	index int
}

func pushChanges(ctx context.Context, hashes []hs.OldFileHash) error {
	hashGob, err := hs.GetSelfHashGob()
	if err != nil {
		return utils.WrapErr(err)
	}
	hashes = append(hashes, hashGob)

	for _, node := range conf.GetConf().RemoteNodes {
		// overseer.Global.Send(pushingNode{index: i})
		if err := pushToNode(ctx, hashes, node); err != nil {
			return utils.WrapErr(err)
		}
	}

	// overseer.Global.Send(doneSend{})
	return nil
}

type pushProgress struct {
	operations *utils.UniqueCircSlice[hs.OldFileHash, remote.SendStatus]
	curr       int
	total      int
}

func newPushProgress(total int) pushProgress {
	return pushProgress{
		operations: utils.NewUniqueCircSlice(
			5,
			func(item remote.SendStatus) hs.OldFileHash { return item.Hash },
		),
		total: total,
	}
}

func pushToNode(ctx context.Context, hashes []hs.OldFileHash, node string) error {
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
	rpcClient, err := remote.New(rpcSesh)
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
			// overseer.Global.Send(pushProg)
		})

	currHashes, err := remote.ValiExisting(rpcClient, hashes)
	if err != nil {
		return utils.WrapErr(err)
	}
	hashes, err = updateProg(hashes, currHashes, progChan)

	currHashes, err = remote.LinkExisting(rpcClient, hashes)
	if err != nil {
		return utils.WrapErr(err)
	}
	hashes, err = updateProg(hashes, currHashes, progChan)

	if len(hashes) > 0 {
		if err := remote.SendFiles(hashes, sftpClient, rpcClient, ctx, progChan); err != nil {
			return utils.WrapErr(err)
		}
	}

	if err := remote.FinSnapshot(rpcClient); err != nil {
		return utils.WrapErr(err)
	}

	return nil
}

// Updates progChan using what hashes are missing.
func updateProg(prev, curr []hs.OldFileHash, progChan chan remote.SendStatus) ([]hs.OldFileHash, error) {
	currSet := utils.SliceToSet(curr)
	for _, h := range prev {
		if _, exists := currSet[h]; !exists {
			status, err := remote.NewDoneStatus(h)
			if err != nil {
				return nil, utils.WrapErr(err)
			}
			progChan <- status
		}
	}
	return curr, nil
}
