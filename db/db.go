package db

import (
	"fmt"
	"path"

	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func OpenWrite(dirPath string) (*sqlite.Conn, error) {
	dbPath := DBPath(dirPath)
	if err := utils.TouchFile(dbPath); err != nil {
		return nil, fmt.Errorf("creating db: %w", err)
	}

	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenReadWrite, sqlite.OpenCreate)
	if err != nil {
		return nil, fmt.Errorf("opening connection: %w", err)
	}
	err = sqlitex.ExecuteTransient(conn,
		"PRAGMA foreign_keys=ON",
		&sqlitex.ExecOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("setting foreign keys on: %w", err)
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
		return nil, fmt.Errorf("opening connection: %w", err)
	}
	return conn, nil
}

func DBPath(dirPath string) string {
	return path.Join(dirPath, "maeve.db")
}

func flushWrites(conn *sqlite.Conn) error {
	err := sqlitex.ExecuteTransient(conn, "PRAGMA wal_checkpoint(FULL);", nil)
	if err != nil {
		return fmt.Errorf("creating checkpoint: %w", err)
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA synchronous=FULL;", nil)
	if err != nil {
		return fmt.Errorf("syncing DB: %w", err)
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA optimize;", nil)
	if err != nil {
		return fmt.Errorf("optimising DB: %w", err)
	}
	return nil
}
