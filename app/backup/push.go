package backup

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/app/remote"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/fynext/icons"
	"github.com/wilymonkey/maeve/utils"
)

// =======================================
// STATES
// =======================================

const (
	PSConnecting = "Connecting to PC..."
	PSPushDB     = "Sending database to PC..."
	PSLinking    = "Backup PC is linking files..."
)

func PSSendFile(path string) string {
	return fmt.Sprintf("Sending file %s...", path)
}

type pushState struct {
	task      binding.String
	percent   binding.Float
	currSize  binding.Item[int64]
	totalSize binding.Item[int64]
}

func newPushState() pushState {
	return pushState{
		task:      binding.NewString(),
		percent:   binding.NewFloat(),
		currSize:  fynext.BindNewInt64(),
		totalSize: fynext.BindNewInt64(),
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
	objects := make([]fyne.CanvasObject, 0, len(guiState.nodeStates))

	for _, state := range guiState.nodeStates {
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
				return
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
		widget.NewLabelWithData(guiState.pushState.task),
		widget.NewProgressBarWithData(guiState.pushState.percent),
	)
}

// =======================================
// METHODS
// =======================================

func pushChanges() error {
	for _, state := range guiState.nodeStates {
		guiState.pushState.task.Set(PSConnecting)
		nodeConn, err := remote.NewNodeConn(state.name, func(err error) {
			state.err.Set(err)
		})
		if err != nil {
			state.err.Set(err)
			continue
		}
		defer nodeConn.Close()

		state.isConnected.Set(true)
		defer state.isConnected.Set(false)

		guiState.pushState.task.Set(PSPushDB)
		dbPath := db.DBPath(conf.MyNode())
		if err := nodeConn.PushDB(dbPath, guiState.ctx); err != nil {
			state.err.Set(err)
			continue
		}

		guiState.pushState.task.Set(PSLinking)
		missingIds, err := nodeConn.ConformToDB()
		if err != nil {
			state.err.Set(err)
			continue
		}

		metas, err := processIds(missingIds)
		if err != nil {
			state.err.Set(err)
			continue
		}

		var progPath string
		var progSize int64
		var totalProgSize int64
		var sentSize int64
		totalSize := fynext.Unwrap(guiState.pushState.totalSize)
		progChan := make(chan *remote.PushStatus, 10)
		utils.Throttle(
			progChan,
			func(p *remote.PushStatus) {
				if p.Meta.RelPath != progPath {
					progPath = p.Meta.RelPath
					progSize = 0
					sentSize += totalProgSize
					totalProgSize = p.Meta.Size
				} else {
					progSize = p.SentSize
				}
			},
			func() {
				if fynext.Unwrap(guiState.pushState.task) != progPath {
					task := PSSendFile(progPath)
					guiState.pushState.task.Set(task)
				}
				currSize := sentSize + progSize
				guiState.pushState.currSize.Set(currSize)
				percent := float64(currSize) / float64(totalSize)
				guiState.pushState.percent.Set(percent)
			},
		)
		if err := nodeConn.PushLinks(metas, guiState.ctx, progChan); err != nil {
			state.err.Set(err)
			continue
		}

		state.isDone.Set(true)
	}

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
		println(m.RelPath)
		missingSize += m.Size
	}
	totalSize := fynext.Unwrap(guiState.pushState.totalSize)
	guiState.pushState.currSize.Set(totalSize - missingSize)

	return metas, err
}
