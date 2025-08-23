package db

import (
	"time"

	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/app/local"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func MetaFromIds(conn *sqlite.Conn, linkIds []int64) ([]*local.FileMeta, error) {
	result := make([]*local.FileMeta, len(linkIds))

	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return nil, help.DevReport(err, "creating read transaction")
	}
	defer endTx(&err)

	for _, id := range linkIds {
		err := sqlitex.ExecuteTransient(conn,
			`SELECT fm.hash, fp.path, fm.size, fm.mod_time
			FROM path_meta_link pml
			JOIN file_meta fm ON pml.file_meta_id = fm.id
			JOIN file_path fp ON pml.path_id = fp.id
			WHERE pml.id = ?;`,
			&sqlitex.ExecOptions{
				Args: []any{id},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					result = append(result, metaFromStmt(stmt))
					return nil
				},
			})
		if err != nil {
			return nil, help.DevReport(err, "getting FileMeta from link id")
		}
	}

	return result, err
}

func IdsFromMetas(conn *sqlite.Conn, metas []*local.FileMeta) ([]int64, error) {
	result := make([]int64, len(metas))

	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return nil, help.DevReport(err, "creating read transaction")
	}
	defer endTx(&err)

	for _, m := range metas {
		err := sqlitex.ExecuteTransient(conn,
			`SELECT pml.id
			FROM file_meta fm
			JOIN path_meta_link pml ON fm.id = pml.file_meta_id
			WHERE fm.hash = ?;`,
			&sqlitex.ExecOptions{
				Args: []any{m.Hash},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					result = append(result, stmt.ColumnInt64(0))
					return nil
				},
			})
		if err != nil {
			return nil, help.DevReport(err, "getting FileMeta from link id")
		}
	}

	return result, err
}

func GetLatestMeta(conn *sqlite.Conn) ([]*local.FileMeta, error) {
	var result []*local.FileMeta

	err := sqlitex.ExecuteTransient(conn,
		`SELECT fm.hash, fp.path, fm.size, fm.mod_time
			FROM path_meta_link pml
			JOIN file_meta fm ON pml.file_meta_id = fm.id
			JOIN file_path fp ON pml.path_id = fp.id
	        WHERE pml.snapshot = (
	            SELECT MAX(snapshot) FROM path_meta_link
	        );`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				result = append(result, metaFromStmt(stmt))
				return nil
			},
		})
	if err != nil {
		return nil, help.DevReport(err, "getting FileMeta from latest snapshot")
	}

	return result, nil
}

// Requires the stmt columns to be in order fo local.FileMeta fields.
func metaFromStmt(stmt *sqlite.Stmt) *local.FileMeta {
	var meta local.FileMeta
	stmt.ColumnBytes(0, meta.Hash[:])
	meta.RelPath = stmt.ColumnText(1)
	meta.Size = stmt.ColumnInt64(2)
	meta.ModTime = time.Unix(stmt.ColumnInt64(3), 0)
	return &meta
}
