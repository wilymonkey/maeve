package backup

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"github.com/wilymonkey/maeve/fynext"
)

var guiState *state

type state struct {
	// STATES
	currTask      binding.Int
	repairDBState binding.String
	pullStates    []*pullState
	updateDBState binding.String
	nodeStates    []*nodeState
	pushState     pushState

	// WINDOW MANAGEMENT
	ctx        context.Context
	ctxCancel  context.CancelFunc
	window     fyne.Window
	err        binding.Item[error]
	exitOnDone bool
}

type currTask int

const (
	repairingDB currTask = iota
	pulling
	updatingDB
	pushing
	done
)

func loadState(window fyne.Window, exitOnDone bool) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	guiState = &state{
		// STATES
		currTask:      binding.NewInt(),
		repairDBState: binding.NewString(),
		pullStates:    newPullState(),
		updateDBState: binding.NewString(),
		nodeStates:    newNodeStates(),
		pushState:     newPushState(),

		// WINDOW MANAGEMENT
		ctx:        ctx,
		ctxCancel:  ctxCancel,
		window:     window,
		err:        fynext.BindNewErr(),
		exitOnDone: exitOnDone,
	}

	guiState.err.AddListener(binding.NewDataListener(func() {
		err := fynext.Unwrap(guiState.err)
		if err != nil {
			guiState.ctxCancel()
			showErrorDialog(err)
		}
	}))
}

func (s *state) nextTask() {
	curr := fynext.Unwrap(s.currTask)
	if curr < int(done) {
		s.currTask.Set(curr + 1)
	}
}
