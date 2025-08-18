package backup

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/app/remote"
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
)

// =======================================
// STATES
// =======================================

const (
	UDBGetLocal  = "getting version of local database..."
	UDBGetRemote = "fetching remote databases..."
	UDBCompare   = "comparing databases to find the best one..."
)

func UDBFetch(node string) string {
	return fmt.Sprintf("fetching best version from %s...", node)
}

func UDBDone(version *db.DBVersion) string {
	return fmt.Sprintf("Done! Current database has %d entries.", version.Rows)
}

// =======================================
// UI
// =======================================

func updateDBUI() fyne.CanvasObject {
	title := widget.NewLabel("Update database:")
	title.TextStyle.Bold = true
	state := widget.NewLabelWithData(guiState.dbState)
	state.Wrapping = fyne.TextWrapWord
	return container.NewBorder(
		nil, nil,
		title,
		nil,
		state,
	)
}

// =======================================
// METHODS
// =======================================

type nodeVersion struct {
	node    string
	version *db.DBVersion
}

func updateDB(conn **sqlite.Conn) error {
	guiState.dbState.Set(UDBGetLocal)
	localVersion, err := db.GetVersion((*conn))
	if err != nil {
		return err
	}

	guiState.dbState.Set(UDBGetRemote)
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

	guiState.dbState.Set(UDBCompare)
	lv := nodeVersion{"", localVersion}
	node := findBestVersion(append([]nodeVersion{lv}, remoteVersions...))
	if node == "" {
		guiState.dbState.Set(UDBDone(localVersion))
		return nil
	}

	guiState.dbState.Set(UDBFetch(node))
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
	*conn, err = db.Open(conf.GetConf().MyNode())
	if err != nil {
		return err
	}
	localVersion, err = db.GetVersion((*conn))
	if err != nil {
		return err
	}

	guiState.dbState.Set(UDBDone(localVersion))
	return nil
}

func findBestVersion(nodeVersions []nodeVersion) string {
	pubKey := conf.GetConf().PublicKey()

	var nodeIndex int
	var latest time.Time
	var rows int64
	for i, nv := range nodeVersions {
		if !ed25519.Verify(pubKey, nv.version.Hash, nv.version.HashSign) {
			continue
		}
		if nv.version.LatestSnapshot.After(latest) && nv.version.Rows >= rows {
			latest = nv.version.LatestSnapshot
			rows = nv.version.Rows
			nodeIndex = i
		}
	}
	return nodeVersions[nodeIndex].node
}
