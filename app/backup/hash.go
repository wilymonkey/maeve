package backup

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/app/conf"
	hs "github.com/wilymonkey/maeve/app/hashsums"
	"github.com/wilymonkey/maeve/utils"
)

type hashProg struct {
	hashNum int
}
type doneHashsums struct {
	hashes []hs.FileHash
}

func newHashes(parentCtx context.Context) tea.Msg {
	selfDir := conf.GetConf().SelfDir()

	progChan := make(chan hs.FileHash, 100)
	defer close(progChan)
	hashNum := 0
	utils.Throttle(
		progChan,
		func(hash hs.FileHash) {
			hashNum++
		},
		func() {
			// TODO: Send report.
		},
	)

	hashes, err := hs.NewDirFileHash(selfDir, progChan, parentCtx)
	if err != nil {
		return utils.WrapErr(err)
	}
	hs.WriteFileHashes(hashes)
	return doneHashsums{hashes: hashes}
}
