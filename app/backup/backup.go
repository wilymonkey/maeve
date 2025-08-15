package backup

import (
	"errors"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
	"github.com/wilymonkey/maeve/remote"
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
)

func Backup(state guiState) error {
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

func (nv *nodeVersion) snapshot() float64 {
	return float64(nv.version.LatestSnapshot.Unix())
}

func (nv *nodeVersion) size() float64 {
	return float64(nv.version.Size)
}

func repairDB(conn *sqlite.Conn, state guiState) error {
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
	var verified []nodeVersion
	var latestSnapshot float64
	latestSnapshotWeight := 0.25
	var largestSize float64
	largestSizeWeight := 1 - latestSnapshot
	for _, nv := range nodeVersions {
		if nv.version.IsValid(pubKey) {
			verified = append(verified, nv)
			largestSize = max(nv.size(), largestSize)
			latestSnapshot = max(nv.snapshot(), latestSnapshot)
		}
	}

	type nodeWeight struct {
		node   string
		weight float64
	}
	var bestNode nodeWeight
	for _, nv := range verified {
		w1 := (nv.snapshot() / latestSnapshot) * latestSnapshotWeight
		w2 := (nv.size() / largestSize) * largestSizeWeight
		total := w1 + w2
		if bestNode.weight < total {
			bestNode = nodeWeight{nv.node, total}
		}
	}
	return bestNode.node
}
