package backup

import (
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
	"github.com/wilymonkey/maeve/remote"
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
)

func Backup(state state) error {
	conn, err := db.Open(conf.GetConf().MyNode())
	if err != nil {
		return err
	}

	if err = repairDB(conn, state); err != nil {
		return err
	}

	return nil
}

type nodeVersion struct {
	node    string
	version *db.DBVersion
}

func repairDB(conn *sqlite.Conn, state state) error {
	localVersion, err := db.GetVersion(conn, conf.GetConf().MyNode())
	if err != nil {
		return err
	}
	versionChan := make(chan nodeVersion, 10)
	collectVersions := utils.CollectChan(versionChan)

	eGrp, _ := state.ErrGroup(2)
	for node, nodeState := range state.nodeStates {
		eGrp.Go(func() error {
			nodeConn, err := remote.NewNodeConn(node, func(err string) {
				state.FatalErr(errors.New(err))
			})
			if err != nil {
				nodeState.err.Set(err.Error())
				return nil
			}
			defer nodeConn.Close()

			dbVersion, err := nodeConn.GetDBVersion()
			if err != nil {
				nodeState.err.Set(err.Error())
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
	lv := nodeVersion{"", localVersion}
	node := findBestVersion(append([]nodeVersion{lv}, remoteVersions...))
	if node == "" {
		return nil
	}

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
