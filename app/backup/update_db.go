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

const (
	UDB_GetLocal  = "getting version of local database..."
	UDB_GetRemote = "fetching remote databases..."
	UDB_Compare   = "comparing databases to find the best one..."
)

func UDB_Done(version *db.DBVersion) {
	state := fmt.Sprintf("Done! Current database has %d entries.", version.Rows)
	guiState.dbState.Set(state)
}

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

type nodeVersion struct {
	node    string
	version *db.DBVersion
}

func repairDB(conn *sqlite.Conn) error {
	guiState.dbState.Set(UDB_GetLocal)

	localVersion, err := db.GetVersion(conn, conf.GetConf().MyNode())
	if err != nil {
		return err
	}
	versionChan := make(chan nodeVersion, 10)
	collectVersions := utils.CollectChan(versionChan)

	guiState.dbState.Set(UDB_GetRemote)
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

	guiState.dbState.Set(UDB_Compare)
	lv := nodeVersion{"", localVersion}
	node := findBestVersion(append([]nodeVersion{lv}, remoteVersions...))
	if node == "" {
		UDB_Done(localVersion)
		return nil
	}

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
	if err := nodeConn.AddSFTP(); err != nil {
		return err
	}
	localVersion, err = db.GetVersion(conn, conf.GetConf().MyNode())
	if err != nil {
		return err
	}
	UDB_Done(localVersion)

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
