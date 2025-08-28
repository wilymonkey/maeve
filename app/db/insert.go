package db

import (
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// Returns rows inserted.
func InsertFileMetas(conn *sqlite.Conn, fileMetas []*local.FileMeta) (int64, error) {
	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return 0, help.WrapError(err, "creating immediate transaction")
	}
	defer endTx(&err)

	var startRows int64
	err = sqlitex.ExecuteTransient(conn,
		"SELECT COUNT(*) FROM path_meta_link;",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				startRows = stmt.ColumnInt64(0)
				return nil
			},
		})
	if err != nil {
		return 0, help.WrapError(err, "counting rows before inserts")
	}

	for _, meta := range fileMetas {
		var fileMetaId int
		err = sqlitex.Execute(conn,
			`INSERT INTO file_meta (hash, size, mod_time)
			VALUES (?, ?, ?)
			ON CONFLICT(hash) DO UPDATE SET size = size
			RETURNING id;`,
			&sqlitex.ExecOptions{
				Args: []any{meta.Hash[:], meta.Size, meta.ModTime.Unix()},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					fileMetaId = stmt.ColumnInt(0)
					return nil
				},
			},
		)
		if err != nil {
			return 0, help.WrapError(err, "insert row into file_meta")
		}

		snapshot, rest, err := formatRelpath(meta.RelPath)
		if err != nil {
			return 0, err
		}
		var pathId int
		err = sqlitex.Execute(conn,
			`INSERT INTO file_path (path)
			VALUES (?)
			ON CONFLICT(path) DO UPDATE SET path = path
			RETURNING id;`,
			&sqlitex.ExecOptions{
				Args: []any{rest},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					pathId = stmt.ColumnInt(0)
					return nil
				},
			},
		)
		if err != nil {
			return 0, help.WrapError(err, "insert row into file_path")
		}

		err = sqlitex.Execute(conn,
			`INSERT OR IGNORE INTO path_meta_link (path_id, file_meta_id, snapshot)
			VALUES (?, ?, ?);`,
			&sqlitex.ExecOptions{
				Args: []any{pathId, fileMetaId, snapshot},
			},
		)
		if err != nil {
			return 0, help.WrapError(err, "insert row into path_meta_link")
		}
	}

	var endRows int64
	err = sqlitex.ExecuteTransient(conn,
		"SELECT COUNT(*) FROM path_meta_link;",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				endRows = stmt.ColumnInt64(0)
				return nil
			},
		})
	if err != nil {
		return 0, help.WrapError(err, "counting rows after inserts")
	}

	return endRows - startRows, err
}

// Splits the relpath into a unix timestamp and the remaining path.
func formatRelpath(relpath string) (int64, string, error) {
	root, rest, err := local.SplitAtRootPath(relpath)
	if err != nil {
		return 0, "", err
	}
	snapshot, err := conf.TimeFromString(root)
	if err != nil {
		return 0, "", help.WrapError(err, "parsing root folder as time")
	}

	return conf.TimeToInt64(snapshot), rest, nil
}
