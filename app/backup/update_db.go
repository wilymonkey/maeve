package backup

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/wilymonkey/maeve/app/remote"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/utils"
	"golang.org/x/sync/errgroup"
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
	UDBDupliCheck = "checking to see if this is a duplicate snapshot..."
	UDBAddEntries = "adding new entries into database..."
	UDBNameChange = "renaming duplicate snapshot..."
)

func UDBDone(rows int64) string {
	return fmt.Sprintf("Done! Added %d entries.", rows)
}

// =======================================
// UI
// =======================================

func repairDBUI() fyne.CanvasObject {
	title := widget.NewLabel("Repairing database:")
	title.TextStyle.Bold = true
	state := widget.NewLabelWithData(global.repairDBState)
	state.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		nil, nil,
		fynext.LabelDisableUntil(title, global.currTask, repairingDB),
		nil,
		fynext.LabelDisableUntil(state, global.currTask, repairingDB),
	)
}

func updateDBUI() fyne.CanvasObject {
	title := widget.NewLabel("Updating database:")
	title.TextStyle.Bold = true
	state := widget.NewLabelWithData(global.updateDBState)
	state.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		nil, nil,
		fynext.LabelDisableUntil(title, global.currTask, updatingDB),
		nil,
		fynext.LabelDisableUntil(state, global.currTask, updatingDB),
	)
}

// =======================================
// METHODS
// =======================================

type nodeVersion struct {
	node    string
	version *db.DBVersion
}

func repairDB() error {
	global.repairDBState.Set(RDBGetLocal)
	localVersion, err := getLocalVer()
	if err != nil {
		return err
	}

	global.repairDBState.Set(RDBGetRemote)

	eGrp, ctx := errgroup.WithContext(global.ctx)
	eGrp.SetLimit(10)

	versionChan := make(chan nodeVersion, 10)
	collectVersions := utils.CollectChan(versionChan)

	for _, nodeState := range global.nodeStates {
		eGrp.Go(func() error {
			node := nodeState.name
			if err := ctx.Err(); err != nil {
				return err
			}

			nodeConn, err := remote.NewNodeConn(node, func(err error) {
				nodeState.err.Set(err)
			})
			if err != nil {
				nodeState.err.Set(err)
				return nil
			}
			defer utils.Cleanup(&err, nodeConn.Close)

			nodeState.isConnected.Set(true)
			defer nodeState.isConnected.Set(false)

			dbVersion, err := nodeConn.GetDBVersion()
			if err != nil {
				nodeState.err.Set(err)
				return nil
			}
			versionChan <- nodeVersion{node: node, version: dbVersion}
			return err
		})
	}
	err = eGrp.Wait()
	close(versionChan)
	if err != nil {
		return err
	}
	remoteVersions := collectVersions()

	global.repairDBState.Set(RDBCompare)
	lv := nodeVersion{"", localVersion}
	node := findBestVersion(append([]nodeVersion{lv}, remoteVersions...))
	if node == "" {
		global.repairDBState.Set(RDBDone(localVersion))
		return nil
	}

	global.repairDBState.Set(RDBFetch(node))
	nodeConn, err := remote.NewNodeConn(node, func(err error) {
		global.err.Set(err)
	})
	if err != nil {
		return err
	}
	defer utils.Cleanup(&err, nodeConn.Close)

	if err := nodeConn.PullDB(); err != nil {
		return err
	}

	localVersion, err = getLocalVer()
	if err != nil {
		return err
	}

	global.repairDBState.Set(RDBDone(localVersion))
	return err
}

func getLocalVer() (*db.DBVersion, error) {
	myNode := conf.MyNode()
	exists, err := local.PathExists(db.DBPath(myNode))
	if err != nil {
		return nil, err
	}
	if !exists {
		return &db.DBVersion{}, nil
	}

	conn, err := db.OpenRead(myNode)
	if err != nil {
		return nil, err
	}
	defer utils.Cleanup(&err, conn.Close)

	ver, err := db.GetVersion(conn)
	if err != nil {
		return nil, err
	}

	return ver, err
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

func updateDB(fileMetas []*local.FileMeta) error {
	global.updateDBState.Set(UDBDupliCheck)

	conn, err := db.OpenWrite(conf.MyNode())
	if err != nil {
		return err
	}
	defer utils.Cleanup(&err, conn.Close)

	prev, err := db.GetLatestMeta(conn)
	if err != nil {
		return err
	}
	metasAreDifferent := func() bool {
		if len(prev) != len(fileMetas) {
			return true
		}

		prevSet := make(map[string]struct{}, len(prev))
		for _, v := range prev {
			prevSet[v.AsMapKey()] = struct{}{}
		}

		for i, entry := range fileMetas {
			if _, exists := prevSet[entry.AsMapKey()]; !exists {
				log.Printf("found bad entry after %d iters", i)
				return true
			}
		}
		return false
	}

	var rows int64
	if metasAreDifferent() {
		global.updateDBState.Set(UDBAddEntries)

		rows, err = db.InsertFileMetas(conn, fileMetas)
		if err != nil {
			return err
		}
	} else {
		global.updateDBState.Set(UDBNameChange)

		myNode := conf.MyNode()
		prevRoot, _ := local.SplitAtRootPath(prev[0].RelPath)
		currRoot, _ := local.SplitAtRootPath(fileMetas[0].RelPath)
		currPath := filepath.Join(myNode, currRoot)
		prevPath := filepath.Join(myNode, prevRoot)

		if currPath != prevPath {
			if err := os.Rename(currPath, prevPath); err != nil {
				return help.WrapErr(err, "renaming current snapshot to previous one")
			}
		}
	}

	global.updateDBState.Set(UDBDone(rows))
	return err
}
