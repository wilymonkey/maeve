package backup

import (
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/fynext"
)

func runBackup() {
	if err := dispatcher(); err != nil {
		global.err.Set(err)
	}

	var nodeHasErr bool
	for _, node := range global.nodeStates {
		if fynext.Unwrap(node.err) != nil {
			nodeHasErr = true
			break
		}
	}

	if fynext.Unwrap(global.err) == nil && global.exitOnDone && !nodeHasErr {
		// TODO: remove this when done.
		// fyne.Do(state.window.Close)
	}

	global.ctxCancel()
}

func dispatcher() error {
	if err := repairDB(); err != nil {
		return err
	}

	global.nextTask()
	metas, err := newLatest()
	if err != nil {
		return err
	}
	updateTotalSize(metas)

	global.nextTask()
	if err = updateDB(metas); err != nil {
		return err
	}

	global.nextTask()
	pushChanges()

	return nil
}

func updateTotalSize(metas []*local.FileMeta) {
	var total int64
	for _, m := range metas {
		total += m.Size
	}
	global.pushState.totalSize.Set(total)
}
