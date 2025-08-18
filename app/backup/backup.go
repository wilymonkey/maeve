package backup

import (
	"errors"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/fynext"
)

var syncFail = errors.New("One or more PCs failed to sync")

func runBackup() {
	if err := dispatcher(); err != nil {
		guiState.err.Set(err)
	}

	for _, node := range guiState.nodeStates {
		if fynext.Unwrap(node.err) != nil {
			guiState.err.Set(syncFail)
			break
		}
	}

	if fynext.Unwrap(guiState.err) == nil && guiState.exitOnDone {
		// TODO: remove this when done.
		// fyne.Do(state.window.Close)
	}

	guiState.ctxCancel()
}

func dispatcher() error {
	conn, err := db.Open(conf.MyNode())
	if err != nil {
		return err
	}

	if err = updateDB(&conn); err != nil {
		return err
	}

	guiState.nextTask()
	if err = newLatest(); err != nil {
		return err
	}

	return nil
}
