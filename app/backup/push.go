package backup

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/app/remote"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/fynext/icons"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
)

// =======================================
// STATES
// =======================================

const (
	sayConnecting = "Connecting to PC..."
	sayPushDB     = "Sending database to PC..."
	sayLinking    = "Backup PC is linking files..."
)

func saySendFile(path string) string {
	p := utils.TruncateString(path, 100)
	return fmt.Sprintf("Sending file %s", p)
}

type pushState struct {
	task      binding.String
	percent   binding.Float
	currSize  binding.Item[int64]
	totalSize binding.Item[int64]
	timeLeft  binding.String
	speed     binding.String
}

func newPushState() pushState {
	return pushState{
		task:      binding.NewString(),
		percent:   binding.NewFloat(),
		currSize:  fynext.BindNewInt64(),
		totalSize: fynext.BindNewInt64(),
		timeLeft:  binding.NewString(),
		speed:     binding.NewString(),
	}
}

type nodeState struct {
	name        string
	isConnected binding.Bool
	isDone      binding.Bool
	err         binding.Item[error]
}

func newNodeStates() []*nodeState {
	cfg := conf.GetConf()
	states := make([]*nodeState, 0, len(cfg.RemoteNodes))
	for _, node := range cfg.RemoteNodes {
		states = append(states, &nodeState{
			name:        node,
			isConnected: binding.NewBool(),
			isDone:      binding.NewBool(),
			err:         fynext.BindNewErr(),
		})
	}
	return states
}

// =======================================
// UI
// =======================================

func nodeStateGrid() fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(global.nodeStates))

	for _, state := range global.nodeStates {
		showErr := func() {
			err := fynext.Unwrap(state.err)
			if err != nil {
				showErrorDialog(err)
			}
		}
		name := utils.TruncateString(state.name, 30)
		btn := widget.NewButtonWithIcon(name, nil, showErr)

		state.isConnected.AddListener(binding.NewDataListener(func() {
			if fynext.Unwrap(state.err) != nil {
				return
			}
			if fynext.Unwrap(state.isConnected) {
				btn.Enable()
				btn.Icon = icons.LinkSvg
			} else {
				btn.Disable()
				btn.Icon = nil
			}
			btn.Refresh()
		}))

		state.isDone.AddListener(binding.NewDataListener(func() {
			if fynext.Unwrap(state.err) != nil {
				return
			}
			if fynext.Unwrap(state.isDone) {
				btn.Enable()
				btn.Icon = icons.CheckSvg
				btn.Importance = widget.SuccessImportance
				btn.Refresh()
			}
		}))

		state.err.AddListener(binding.NewDataListener(func() {
			if fynext.Unwrap(state.err) != nil {
				btn.Enable()
				btn.Icon = icons.DangerSvg
				btn.Importance = widget.DangerImportance
				btn.Refresh()
			}
		}))

		objects = append(objects, btn)
	}

	grid := container.NewGridWrap(fyne.NewSize(200, 50), objects...)
	return container.NewVScroll(grid)
}

func pushStateCard() fyne.CanvasObject {
	return fynext.VBox(
		widget.NewLabelWithData(global.pushState.task),
		widget.NewProgressBarWithData(global.pushState.percent),
		container.NewHBox(
			widget.NewLabelWithData(global.pushState.speed),
			layout.NewSpacer(),
			widget.NewLabelWithData(global.pushState.timeLeft),
		),
	)
}

// =======================================
// METHODS
// =======================================

func pushChanges() {
	for _, state := range global.nodeStates {
		if err := pushToNode(state); err != nil {
			state.err.Set(err)
		}
	}
}

func pushToNode(state *nodeState) error {
	global.pushState.task.Set(sayConnecting)
	nodeConn, err := remote.NewNodeConn(state.name, func(err error) {
		state.err.Set(err)
	})
	if err != nil {
		return err
	}
	defer utils.Cleanup(&err, nodeConn.Close)

	state.isConnected.Set(true)
	defer state.isConnected.Set(false)

	global.pushState.task.Set(sayPushDB)
	dbPath := db.DBPath(conf.MyNode())
	if err := nodeConn.PushDB(dbPath, global.ctx); err != nil {
		return err
	}

	global.pushState.task.Set(sayLinking)
	missingIds, err := nodeConn.ConformToDB()
	if err != nil {
		return err
	}

	metas, err := processIds(missingIds)
	if err != nil {
		return err
	}

	progChan := make(chan *remote.PushStatus, 10)
	eGrp, ctx := errgroup.WithContext(global.ctx)
	eGrp.Go(func() error {
		return nodeConn.PushLinks(metas, ctx, progChan)
	})

	var progPath string
	var sentSize int64
	var totalFileSize int64
	var alreadySentSize int64
	totalSize := fynext.Unwrap(global.pushState.totalSize)
	speedSlice := utils.NewCircSlice[int64](10)

	utils.Throttle(
		progChan,
		func(p *remote.PushStatus) {
			if p.Meta.RelPath != progPath {
				progPath = p.Meta.RelPath
				sentSize = 0
				alreadySentSize += totalFileSize
				totalFileSize = p.Meta.Size
			} else {
				sentSize = p.SentSize
			}
		},
		func() {
			if fynext.Unwrap(global.pushState.task) != progPath {
				task := saySendFile(progPath)
				global.pushState.task.Set(task)
			}

			prevCurrSize := fynext.Unwrap(global.pushState.currSize)
			currSize := alreadySentSize + sentSize
			speed := calcSpeed(speedSlice, prevCurrSize, currSize)
			global.pushState.speed.Set(utils.BytesToHuman(speed))

			left := calcTimeLeft(speed, totalSize-currSize)
			global.pushState.timeLeft.Set(left)

			global.pushState.currSize.Set(currSize)
			percent := float64(currSize) / float64(totalSize)
			global.pushState.percent.Set(percent)
		},
	)

	if err := eGrp.Wait(); err != nil {
		return err
	}

	state.isDone.Set(true)
	return nil
}

// Converts linkIds to FileMeta and updates currSize.
func processIds(linkIds []int64) ([]*local.FileMeta, error) {
	conn, err := db.OpenRead(conf.MyNode())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()

	metas, err := db.MetaFromIds(conn, linkIds)
	if err != nil {
		return nil, err
	}

	var missingSize int64
	for _, m := range metas {
		missingSize += m.Size
	}
	totalSize := fynext.Unwrap(global.pushState.totalSize)
	global.pushState.currSize.Set(totalSize - missingSize)

	return metas, err
}

func calcSpeed(speedSlice *utils.CircSlice[int64], prevSize, currSize int64) int64 {
	speed := ((currSize - prevSize) / utils.ThrottleMS) * 1000

	if hasPushed := speedSlice.Push(speed); !hasPushed {
		speedSlice.Pop()
		speedSlice.Push(speed)
	}
	var total int64
	speedSlice.ForEach(func(size int64) {
		total += size
	})
	return total / int64(speedSlice.Len())
}

func calcTimeLeft(speed int64, size int64) string {
	if speed <= 0 {
		return "∞"
	}
	d := time.Duration(size/speed) * time.Second
	return d.String()
}
