package backup

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	hs "github.com/wilymonkey/maeve/hashsums"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/utils"
)

type hashProg struct {
	hashNum int
}
type doneHashsums struct {
	hashes []hs.FileHash
}

func newHashes(parentCtx context.Context) tea.Msg {
	selfDir := cfg.Global.SelfDir()

	progChan := make(chan hs.FileHash, 100)
	defer close(progChan)
	hashNum := 0
	utils.Throttle(
		progChan,
		func(hash hs.FileHash) {
			hashNum++
		},
		func() {
			overseer.Global.Send(hashProg{hashNum: hashNum})
		},
	)

	hashes, err := hs.NewDirFileHash(selfDir, progChan, parentCtx)
	if err != nil {
		return utils.WrapErr(err)
	}
	hs.WriteFileHashes(hashes)
	return doneHashsums{hashes: hashes}
}
