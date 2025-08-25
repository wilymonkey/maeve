package backup

import (
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/fynext"
)

func runBackup() {
	if err := dispatcher(); err != nil {
		guiState.err.Set(err)
	}

	var nodeHasErr bool
	for _, node := range guiState.nodeStates {
		if fynext.Unwrap(node.err) != nil {
			nodeHasErr = true
			break
		}
	}

	if fynext.Unwrap(guiState.err) == nil && guiState.exitOnDone && !nodeHasErr {
		// TODO: remove this when done.
		// fyne.Do(state.window.Close)
	}

	guiState.ctxCancel()
}

func dispatcher() error {
	if err := repairDB(); err != nil {
		return err
	}

	guiState.nextTask()
	metas, err := newLatest()
	if err != nil {
		return err
	}
	updateTotalSize(metas)

	guiState.nextTask()
	if err = updateDB(metas); err != nil {
		return err
	}

	guiState.nextTask()
	if err = pushChanges(); err != nil {
		return err
	}

	return nil
}

func updateTotalSize(metas []*local.FileMeta) {
	var total int64
	for _, m := range metas {
		total += m.Size
	}
	guiState.pushState.totalSize.Set(total)
}
