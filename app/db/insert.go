package db

import (
	"github.com/wilymonkey/maeve/app/hashsums"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/zeebo/blake3"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// Returns rows inserted.
func InsertFileMetas(conn *sqlite.Conn, fileMetas []*hashsums.FileMeta) (int, error) {
	var rows int

	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return 0, help.DevReport(err, "creating immediate transaction")
	}
	defer endTx(&err)

	for _, meta := range fileMetas {
		var fileMetaId int64
		err = sqlitex.Execute(conn,
			`INSERT OR REPLACE INTO file_meta (hash, size, mod_time)
			VALUES (?, ?, ?)
			RETURNING id;`,
			&sqlitex.ExecOptions{
				Args: []any{meta.Hash, meta.Size, meta.ModTime.Unix()},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					fileMetaId = stmt.ColumnInt64(0)
					return nil
				},
			},
		)
		if err != nil {
			return 0, help.DevReport(err, "insert row into file_meta")
		}

		snapshot, err := local.UnixFromPath(meta.RelPath)
		if err != nil {
			return 0, err
		}
		relPathHash := blake3.Sum256([]byte(meta.RelPath))
		err = sqlitex.Execute(conn,
			`INSERT OR REPLACE INTO path_to_hash (rel_path_hash, file_meta_id, snapshot)
			VALUES (?, ?, ?);`,
			&sqlitex.ExecOptions{
				Args: []any{relPathHash, fileMetaId, snapshot},
			},
		)
		if err != nil {
			return 0, help.DevReport(err, "insert row into path_to_hash")
		}

		rows++
	}

	return rows, err
}
