package db

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/wilymonkey/maeve/app/hashsums"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"github.com/zeebo/blake3"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// Returns rows inserted.
func InsertFileMetas(conn *sqlite.Conn, fileMetas []*hashsums.FileMeta) (int64, error) {
	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return 0, help.DevReport(err, "creating immediate transaction")
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
		return 0, help.DevReport(err, "counting rows before inserts")
	}

	for _, meta := range fileMetas {
		var fileMetaId int
		err = sqlitex.Execute(conn,
			`INSERT INTO file_meta (hash, size, mod_time)
			VALUES (?, ?, ?)
			ON CONFLICT(hash) DO UPDATE SET size = size
			RETURNING id;`,
			&sqlitex.ExecOptions{
				Args: []any{meta.Hash, meta.Size, meta.ModTime.Unix()},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					fileMetaId = stmt.ColumnInt(0)
					return nil
				},
			},
		)
		if err != nil {
			return 0, help.DevReport(err, "insert row into file_meta")
		}

		snapshot, pathHash, err := formatRelpath(meta.RelPath)
		if err != nil {
			return 0, err
		}

		var pathHashId int
		err = sqlitex.Execute(conn,
			`INSERT INTO path_hash (hash)
			VALUES (?)
			ON CONFLICT(hash) DO UPDATE SET hash = hash
			RETURNING id;`,
			&sqlitex.ExecOptions{
				Args: []any{pathHash},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					pathHashId = stmt.ColumnInt(0)
					return nil
				},
			},
		)
		if err != nil {
			return 0, help.DevReport(err, "insert row into path_hash")
		}

		err = sqlitex.Execute(conn,
			`INSERT OR IGNORE INTO path_meta_link (path_id, file_meta_id, snapshot)
			VALUES (?, ?, ?)
			ON CONFLICT DO NOTHING;`,
			&sqlitex.ExecOptions{
				Args: []any{pathHashId, fileMetaId, snapshot},
			},
		)
		if err != nil {
			return 0, help.DevReport(err, "insert row into path_meta_link")
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
		return 0, help.DevReport(err, "counting rows after inserts")
	}

	return endRows - startRows, err
}

// Splits the relpath into a unix timestamp and a hash of the remaining path.
func formatRelpath(relpath string) (int64, [32]byte, error) {
	var result [32]byte

	cleaned := filepath.Clean(relpath)
	parts := strings.Split(cleaned, string(filepath.Separator))
	if len(parts) == 0 || parts[0] == "" {
		err := fmt.Errorf("%q has no root folder", relpath)
		return 0, result, help.DevReport(err, "getting root folder")
	}
	snapshot, err := time.Parse(local.DirTimeFormat, parts[0])
	if err != nil {
		return 0, result, help.DevReport(err, "parsing root folder as time")
	}
	if len(parts) < 2 || parts[1] == "" {
		err := fmt.Errorf("%q has no child path", relpath)
		return 0, result, help.DevReport(err, "getting child path")
	}
	childpath := filepath.Join(parts[1:]...)
	result = blake3.Sum256([]byte(childpath))

	return snapshot.Unix(), result, nil
}
