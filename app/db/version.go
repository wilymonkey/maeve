package db

import (
	"fmt"
	"time"

	"github.com/wilymonkey/maeve/hashsums"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type DBVersion struct {
	hash     [32]byte
	snapshot time.Time
}

func GetVersion(conn *sqlite.Conn) (DBVersion, error) {
	var version DBVersion

	const stmt = `
		SELECT hash, snapshot
		FROM db_version
		ORDER BY snapshot DESC
		LIMIT 1
	`

	err := sqlitex.Execute(conn, stmt, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			stmt.BindBytes(0, version.hash[:])
			version.snapshot = time.Unix(stmt.ColumnInt64(1), 0)
			return nil
		},
	})
	if err != nil {
		return version, fmt.Errorf("getting DBVersion: %w", err)
	}
	return version, nil
}

func newDBVersionTable(conn *sqlite.Conn) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS db_version (
		hash     BLOB NOT NULL,
		snapshot INTEGER NOT NULL
	);`

	if err := sqlitex.ExecuteScript(conn, schema, nil); err != nil {
		return fmt.Errorf("creating DBVersion table: %w", err)
	}
	return nil
}

// Will close the current connect to DB and reopen it.
func SetDBVersion(conn *sqlite.Conn, dbPath string) error {
	if err := conn.Close(); err != nil {
		return fmt.Errorf("closing dbPath: %s: %w", dbPath, err)
	}

	hash, err := hashsums.NewHashsum(dbPath)
	if err != nil {
		return fmt.Errorf("hashing dbPath: %s: %w", dbPath, err)
	}

	conn, err = sqlite.OpenConn(dbPath)
	if err != nil {
		return fmt.Errorf("opening dbPath: %s: %w", dbPath, err)
	}

	const stmt = `INSERT INTO db_version (hash, snapshot) VALUES (?, ?)`
	err = sqlitex.ExecuteTransient(conn,
		`INSERT INTO db_version (hash, snapshot) VALUES (?, ?)`,
		&sqlitex.ExecOptions{
			Args: []any{
				hash,
				time.Now().Unix(),
			},
		})
	if err != nil {
		return fmt.Errorf("adding DBVersion: %w", err)
	}
	return nil
}
