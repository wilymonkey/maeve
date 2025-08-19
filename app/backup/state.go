package backup

import (
	"context"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/fynext"
	"golang.org/x/sync/errgroup"
)

var guiState *state

type state struct {
	currTask      binding.Int
	repairDBState binding.String
	pullStates    []*pullState
	updateDBState binding.String
	ctx           context.Context
	ctxCancel     context.CancelFunc
	window        fyne.Window
	err           binding.Item[error]
	exitOnDone    bool

	// OLD
	nodeStates  map[string]*nodeStatus
	hashDone    bool
	totalFiles  int
	hashNum     int
	pushProg    pushProgress
	pushingNode int
	pushDone    bool
}

func (s *state) ErrGroup(scaling int) (*errgroup.Group, context.Context) {
	eGrp, ctx := errgroup.WithContext(s.ctx)
	eGrp.SetLimit(scaling * runtime.NumCPU())
	return eGrp, ctx
}

type currTask int

const (
	repairingDB currTask = iota
	pulling
	updatingDB
	pushing
	done
)

type nodeStatus struct {
	err binding.Item[error]
}

func loadState(window fyne.Window, exitOnDone bool) {
	cfg := conf.GetConf()

	nodeStates := make(map[string]*nodeStatus, len(cfg.RemoteNodes))
	for _, node := range cfg.RemoteNodes {
		nodeStates[node] = &nodeStatus{
			err: fynext.NewErrBinding(),
		}
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	guiState = &state{
		currTask:      binding.NewInt(),
		repairDBState: binding.NewString(),
		pullStates:    newPullState(),
		updateDBState: binding.NewString(),
		nodeStates:    nodeStates,
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		window:        window,
		err:           fynext.NewErrBinding(),
		exitOnDone:    exitOnDone,

		// OLD
		pushProg: newPushProgress(0),
	}

	guiState.err.AddListener(binding.NewDataListener(func() {
		err := fynext.Unwrap(guiState.err)
		if err != nil {
			guiState.ctxCancel()
			showErrorDialog()
		}
	}))
}

func (s *state) nextTask() {
	curr := fynext.Unwrap(s.currTask)
	if curr < int(done) {
		s.currTask.Set(curr + 1)
	}
}
