package db

import (
	"path"

	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func Open(dirPath string) (*sqlite.Conn, error) {
	dbPath := path.Join(dirPath, "maeve.db")
	if err := utils.TouchFile(dbPath); err != nil {
		return nil, utils.Stacktrace(err, "creating db")
	}

	conn, err := sqlite.OpenConn(dbPath)
	if err != nil {
		return nil, utils.Stacktrace(err, "opening connection")
	}

	if err := createSchema(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func flushWrites(conn *sqlite.Conn) error {
	err := sqlitex.ExecuteTransient(conn, "PRAGMA wal_checkpoint(FULL);", nil)
	if err != nil {
		return utils.Stacktrace(err, "creating checkpoint")
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA synchronous=FULL;", nil)
	if err != nil {
		return utils.Stacktrace(err, "syncing DB")
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA optimize;", nil)
	if err != nil {
		return utils.Stacktrace(err, "optimising DB")
	}
	return nil
}
