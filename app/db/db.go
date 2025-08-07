package db

import (
	"fmt"
	"path"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func Open(dirPath string) (*sqlite.Conn, error) {
	dbPath := path.Join(dirPath, "hashtable.db")
	conn, err := sqlite.OpenConn(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening dbPath: %s: %w", dbPath, err)
	}

	if err := initDB(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func initDB(conn *sqlite.Conn) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS file_hash (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash BLOB NOT NULL,
		snapshot TEXT NOT NULL,
		rel_path TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS file_meta (
		hash BLOB PRIMARY KEY,
		rel_path TEXT NOT NULL,
		size INTEGER NOT NULL,
		mod_time TEXT NOT NULL
	);
	`
	if err := sqlitex.ExecScript(conn, schema); err != nil {
		return fmt.Errorf("creating FileHash and FileMeta tables: %w", err)
	}
	return nil
}
