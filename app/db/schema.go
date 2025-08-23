package db

import (
	"github.com/wilymonkey/maeve/app/help"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func createSchema(conn *sqlite.Conn) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS file_meta (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash BLOB NOT NULL UNIQUE,
		size INTEGER NOT NULL,
		mod_time INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS file_path (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS path_meta_link (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path_id INTEGER NOT NULL,
		file_meta_id INTEGER NOT NULL,
		snapshot INTEGER NOT NULL,
		FOREIGN KEY (path_id) REFERENCES path_hash(id),
		FOREIGN KEY (file_meta_id) REFERENCES file_meta(id),
		UNIQUE (path_id, file_meta_id, snapshot)
	);

	CREATE TABLE IF NOT EXISTS hash_sign (
	    id INTEGER PRIMARY KEY CHECK (id = 1),
	    sign BLOB NOT NULL
	);
	`
	if err := sqlitex.ExecScript(conn, schema); err != nil {
		return help.Stacktrace(err, "creating schema", help.DelDB)
	}
	return nil
}
