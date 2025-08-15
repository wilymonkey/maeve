package db

import (
	"path"

	"github.com/wilymonkey/maeve/help"
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
)

func Open(dirPath string) (*sqlite.Conn, error) {
	dbPath := dbPath(dirPath)
	if err := utils.TouchFile(dbPath); err != nil {
		return nil, help.Stacktrace(err, "creating db", help.DelBackupDir)
	}

	conn, err := sqlite.OpenConn(dbPath)
	if err != nil {
		return nil, help.Stacktrace(err, "opening connection", help.DelBackupDir)
	}

	if err := createSchema(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func dbPath(dirPath string) string {
	return path.Join(dirPath, "maeve.db")
}
