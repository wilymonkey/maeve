package backup

import (
	"crypto/ed25519"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/app/hashsums"
	"github.com/wilymonkey/maeve/app/remote"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
)

// =======================================
// STATES
// =======================================

const (
	RDBGetLocal  = "getting version of local database..."
	RDBGetRemote = "fetching remote databases..."
	RDBCompare   = "comparing databases to find the best one..."
)

func RDBFetch(node string) string {
	return fmt.Sprintf("fetching best version from %s...", node)
}

func RDBDone(version *db.DBVersion) string {
	return fmt.Sprintf("Done! Current database has %d entries.", version.Rows)
}

const (
	UDBAddEntries = "adding new entries into database..."
)

func UDBDone(rows int) string {
	return fmt.Sprintf("Done! Added %d entries.", rows)
}

// =======================================
// UI
// =======================================

func repairDBUI() fyne.CanvasObject {
	title := widget.NewLabel("Repairing database:")
	title.TextStyle.Bold = true
	state := widget.NewLabelWithData(guiState.repairDBState)
	state.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		nil, nil,
		fynext.LabelDisableUntil(title, guiState.currTask, repairingDB),
		nil,
		fynext.LabelDisableUntil(state, guiState.currTask, repairingDB),
	)
}

func updateDBUI() fyne.CanvasObject {
	title := widget.NewLabel("Updating database:")
	title.TextStyle.Bold = true
	state := widget.NewLabelWithData(guiState.updateDBState)
	state.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		nil, nil,
		fynext.LabelDisableUntil(title, guiState.currTask, updatingDB),
		nil,
		fynext.LabelDisableUntil(state, guiState.currTask, updatingDB),
	)
}

// =======================================
// METHODS
// =======================================

type nodeVersion struct {
	node    string
	version *db.DBVersion
}

func repairDB(conn **sqlite.Conn) error {
	guiState.repairDBState.Set(RDBGetLocal)
	localVersion, err := db.GetVersion((*conn))
	if err != nil {
		return err
	}

	guiState.repairDBState.Set(RDBGetRemote)
	versionChan := make(chan nodeVersion, 10)
	collectVersions := utils.CollectChan(versionChan)
	eGrp, _ := guiState.ErrGroup(2)
	for node, nodeState := range guiState.nodeStates {
		eGrp.Go(func() error {
			nodeConn, err := remote.NewNodeConn(node, func(err error) {
				nodeState.err.Set(err)
			})
			if err != nil {
				nodeState.err.Set(err)
				return nil
			}
			defer nodeConn.Close()

			dbVersion, err := nodeConn.GetDBVersion()
			if err != nil {
				nodeState.err.Set(err)
				return nil
			}
			versionChan <- nodeVersion{node: node, version: dbVersion}
			return nil
		})
	}
	err = eGrp.Wait()
	close(versionChan)
	if err != nil {
		return err
	}
	remoteVersions := collectVersions()

	guiState.repairDBState.Set(RDBCompare)
	lv := nodeVersion{"", localVersion}
	node := findBestVersion(append([]nodeVersion{lv}, remoteVersions...))
	if node == "" {
		guiState.repairDBState.Set(RDBDone(localVersion))
		return nil
	}

	guiState.repairDBState.Set(RDBFetch(node))
	(*conn).Close()
	nodeConn, err := remote.NewNodeConn(node, func(err error) {
		guiState.err.Set(err)
	})
	if err != nil {
		return err
	}
	defer nodeConn.Close()
	if err := nodeConn.AddSFTP(); err != nil {
		return err
	}
	if err := nodeConn.PullDB(); err != nil {
		return err
	}
	*conn, err = db.Open(conf.MyNode())
	if err != nil {
		return err
	}
	localVersion, err = db.GetVersion((*conn))
	if err != nil {
		return err
	}

	guiState.repairDBState.Set(RDBDone(localVersion))
	return nil
}

func findBestVersion(nodeVersions []nodeVersion) string {
	pubKey := conf.GetConf().PublicKey()

	var nodeIndex int
	var latest int64
	var rows int64
	for i, nv := range nodeVersions {
		if !ed25519.Verify(pubKey, nv.version.Hash, nv.version.HashSign) {
			continue
		}
		if nv.version.LatestSnapshot > latest && nv.version.Rows >= rows {
			latest = nv.version.LatestSnapshot
			rows = nv.version.Rows
			nodeIndex = i
		}
	}
	return nodeVersions[nodeIndex].node
}

func updateDB(conn *sqlite.Conn, fileMetas []*hashsums.FileMeta) error {
	guiState.updateDBState.Set(UDBAddEntries)
	rows, err := db.InsertFileMetas(conn, fileMetas)
	if err != nil {
		return err
	}

	guiState.updateDBState.Set(UDBDone(rows))
	return nil
}
