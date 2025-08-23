package db

import (
	"path"

	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func OpenWrite(dirPath string) (*sqlite.Conn, error) {
	dbPath := DBPath(dirPath)
	if err := utils.TouchFile(dbPath); err != nil {
		return nil, help.Stacktrace(err, "creating db", help.DelBackupDir)
	}

	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenReadWrite, sqlite.OpenCreate)
	if err != nil {
		return nil, help.Stacktrace(err, "opening connection", help.DelBackupDir)
	}

	if err := createSchema(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// Read only open.
func OpenRead(dirPath string) (*sqlite.Conn, error) {
	dbPath := DBPath(dirPath)
	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenReadOnly)
	if err != nil {
		return nil, help.Stacktrace(err, "opening connection", help.DelBackupDir)
	}
	return conn, nil
}

func DBPath(dirPath string) string {
	return path.Join(dirPath, "maeve.db")
}

func flushWrites(conn *sqlite.Conn) error {
	err := sqlitex.ExecuteTransient(conn, "PRAGMA wal_checkpoint(FULL);", nil)
	if err != nil {
		return help.DevReport(err, "creating checkpoint")
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA synchronous=FULL;", nil)
	if err != nil {
		return help.DevReport(err, "syncing DB")
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA optimize;", nil)
	if err != nil {
		return help.DevReport(err, "optimising DB")
	}
	return nil
}
