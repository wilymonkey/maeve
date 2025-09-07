package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/help"
	"github.com/wilymonkey/maeve/local"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func MetaFromIds(conn *sqlite.Conn, linkIds []int64) ([]*local.FileMeta, error) {
	result := make([]*local.FileMeta, 0, len(linkIds))

	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return nil, help.WrapErr(err, "creating read transaction")
	}
	defer endTx(&err)

	for _, id := range linkIds {
		err := sqlitex.Execute(conn,
			`SELECT fm.hash, pml.snapshot, fp.path, fm.size, fm.mod_time
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
			return nil, help.WrapErr(err, "getting FileMeta from link id")
		}
	}

	if len(result) != len(linkIds) {
		err := fmt.Errorf("only found %d out of %d FileMetas", len(result), len(linkIds))
		return nil, help.WrapErr(err, "check result length")
	}

	return result, err
}

func IdsFromMetas(conn *sqlite.Conn, metas []*local.FileMeta) ([]int64, error) {
	if len(metas) < 1 {
		err := errors.New("no FileMetas given to get Ids for")
		return nil, help.WrapErr(err, "checking input length")
	}

	result := make([]int64, 0, len(metas))

	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return nil, help.WrapErr(err, "creating read transaction")
	}
	defer endTx(&err)

	snapshot, _, err := formatRelpath(metas[0].RelPath)
	if err != nil {
		return nil, err
	}

	for _, m := range metas {
		_, path := local.SplitAtRootPath(m.RelPath)

		err = sqlitex.Execute(conn,
			`SELECT pml.id
			FROM path_meta_link pml
			JOIN file_path fp ON pml.path_id = fp.id
			JOIN file_meta fm ON pml.file_meta_id = fm.id
			WHERE fp.path = ?
			  AND fm.hash = ?
			  AND pml.snapshot = ?;`,
			&sqlitex.ExecOptions{
				Args: []any{path, m.Hash[:], snapshot},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					result = append(result, stmt.ColumnInt64(0))
					return nil
				},
			})
		if err != nil {
			return nil, help.WrapErr(err, "getting link id from FileMeta")
		}
	}

	if len(result) != len(metas) {
		err := fmt.Errorf("found %d out of %d path meta ids", len(result), len(metas))
		return nil, help.WrapErr(err, "checking result length")
	}

	return result, err
}

func GetLatestMeta(conn *sqlite.Conn) ([]*local.FileMeta, error) {
	var result []*local.FileMeta

	err := sqlitex.ExecuteTransient(conn,
		`SELECT fm.hash, pml.snapshot, fp.path, fm.size, fm.mod_time
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
		return nil, help.WrapErr(err, "getting FileMeta from latest snapshot")
	}

	return result, nil
}

// Requires the stmt columns to be in order fo local.FileMeta fields.
func metaFromStmt(stmt *sqlite.Stmt) *local.FileMeta {
	var meta local.FileMeta

	stmt.ColumnBytes(0, meta.Hash[:])
	snapshot := conf.TimeFromInt64(stmt.ColumnInt64(1))
	rest := stmt.ColumnText(2)
	meta.Size = stmt.ColumnInt64(3)
	meta.ModTime = time.Unix(stmt.ColumnInt64(4), 0)

	meta.RelPath = filepath.Join(conf.TimeToString(snapshot), rest)
	return &meta
}
