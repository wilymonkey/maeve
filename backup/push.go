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

func pushChanges(ctx context.Context, hashes []hs.FileHash) error {
	for _, node := range cfg.Global.RemoteNodes {
		if err := pushToNode(ctx, hashes, node); err != nil {
			return myerr.WrapErr(err)
		}
	}
	return nil
}

type doneProgress struct {
	operations utils.UniqueCircSlice[remote.SendStatus]
	curr       int
	total      int
}

func pushToNode(ctx context.Context, hashes []hs.FileHash, node string) error {
	sshClient, err := remote.NewSSHClient(node)
	if err != nil {
		return myerr.WrapErr(err)
	}
	defer sshClient.Close()

	progChan := make(chan remote.SendStatus, 100)
	defer close(progChan)
	doneProg := doneProgress{total: len(hashes)}

	utils.Throttle(
		progChan,
		func(status remote.SendStatus) {
			if status.IsGood {
				doneProg.curr++
			}
			doneProg.operations.Push(status)
		},
		func() {
			cfg.TuiProgram.Send(doneProg)
		})

	if err := remote.SendFiles(hashes, sshClient, ctx, progChan); err != nil {
		return myerr.WrapErr(err)
	}
	return nil
}
