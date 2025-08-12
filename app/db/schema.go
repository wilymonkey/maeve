package db

import (
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func createSchema(conn *sqlite.Conn) error {
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
		return utils.Stacktrace(err, "creating schema")
	}
	return nil
}
