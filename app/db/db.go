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
		return nil, help.WrapError(err, "creating db")
	}

	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenReadWrite, sqlite.OpenCreate)
	if err != nil {
		return nil, help.WrapError(err, "opening connection")
	}
	err = sqlitex.ExecuteTransient(conn,
		"PRAGMA foreign_keys=ON",
		&sqlitex.ExecOptions{},
	)
	if err != nil {
		return nil, help.WrapError(err, "setting foreign keys on")
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
		return nil, help.WrapError(err, "opening connection")
	}
	return conn, nil
}

func DBPath(dirPath string) string {
	return path.Join(dirPath, "maeve.db")
}

func flushWrites(conn *sqlite.Conn) error {
	err := sqlitex.ExecuteTransient(conn, "PRAGMA wal_checkpoint(FULL);", nil)
	if err != nil {
		return help.WrapError(err, "creating checkpoint")
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA synchronous=FULL;", nil)
	if err != nil {
		return help.WrapError(err, "syncing DB")
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA optimize;", nil)
	if err != nil {
		return help.WrapError(err, "optimising DB")
	}
	return nil
}
