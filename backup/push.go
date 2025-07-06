package backup

import (
	"context"

	"github.com/wilymonkey/maeve/cfg"
	hs "github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/remote"
	"github.com/wilymonkey/maeve/utils"
	"github.com/wilymonkey/maeve/utils/myerr"
)

type doneSend struct{}

type pushingNode struct {
	index int
}

func pushChanges(ctx context.Context, hashes []hs.FileHash) error {
	for i, node := range cfg.Global.RemoteNodes {
		cfg.TuiProgram.Send(pushingNode{index: i})
		if err := pushToNode(ctx, hashes, node); err != nil {
			return myerr.WrapErr(err)
		}
	}
	return nil
}

type pushProgress struct {
	operations *utils.UniqueCircSlice[remote.SendStatus]
	curr       int
	total      int
}

func newPushProgress(total int) pushProgress {
	return pushProgress{
		operations: utils.NewUniqueCircSlice[remote.SendStatus](5),
		total:      total,
	}
}

func pushToNode(ctx context.Context, hashes []hs.FileHash, node string) error {
	sshClient, err := remote.NewSSHClient(node)
	if err != nil {
		return myerr.WrapErr(err)
	}
	defer sshClient.Close()

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
			cfg.TuiProgram.Send(pushProg)
		})

	if err := remote.SendFiles(hashes, sshClient, ctx, progChan); err != nil {
		return myerr.WrapErr(err)
	}
	return nil
}
