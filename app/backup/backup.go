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

type nodeDBVersion struct {
	node    string
	version *db.DBVersion
}

func repairDB(conn *sqlite.Conn, state guiState) error {
	localVersion, err := db.GetVersion(conn)
	if err != nil {
		return err
	}
	versionChan := make(chan nodeDBVersion, 10)
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
			versionChan <- nodeDBVersion{node: node, version: dbVersion}
			return nil
		})
	}
	err = eGrp.Wait()
	close(versionChan)
	if err != nil {
		return err
	}

	remoteVersions := collectVersions()
	node := findLatestVersion(localVersion, remoteVersions)
	if node == "" {
		return nil
	}

	return nil
}

func findLatestVersion(local *db.DBVersion, nodeVersions []nodeDBVersion) string {
	agreementMap := make(map[int64]int)
	for _, v := range nodeVersions {
		for _, snapshot := range v.version.Snapshot {
			agreementMap[snapshot.Unix()]++
		}
	}
	for _, snapshot := range local.Snapshot {
		agreementMap[snapshot.Unix()]++
	}
	for snapshot, score := range agreementMap {
		if score < 2 {
			delete(agreementMap, snapshot)
		}
	}

	scores := make([]int, len(nodeVersions))
	var localScore int
	for i, v := range nodeVersions {
		for _, snapshot := range v.version.Snapshot {
			if _, exists := agreementMap[snapshot.Unix()]; exists {
				scores[i]++
			}
		}
	}
	for _, snapshot := range local.Snapshot {
		if _, exists := agreementMap[snapshot.Unix()]; exists {
			localScore++
		}
	}

	var bestNodeIdx int
	var bestNodeScore int
	for i, s := range scores {
		if s > bestNodeScore {
			bestNodeIdx = i
			bestNodeScore = s
		}
	}
	if bestNodeScore > localScore {
		return nodeVersions[bestNodeIdx].node
	} else {
		return ""
	}
}
