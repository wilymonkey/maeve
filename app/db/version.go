package db

import (
	"time"

	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type DBVersion struct {
	Snapshot time.Time
	Size     int64
	SizeHash []byte
}

func GetVersion(conn *sqlite.Conn) (*DBVersion, error) {
	var result DBVersion

	err := sqlitex.Execute(conn,
		"SELECT COUNT(*) FROM file_hash;",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				result.Total = stmt.ColumnInt(0)
				return nil
			},
		})
	if err != nil {
		return nil, utils.Stacktrace(err, "querying total file_hash rows")
	}
	if result.Total == 0 {
		return &result, nil
	}

	// Get the latest 50 snapshots
	i := 0
	err = sqlitex.Execute(conn, `
		SELECT DISTINCT snapshot 
		FROM file_hash 
		ORDER BY id DESC 
		LIMIT 50;
        `,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				result.Snapshot[i] = time.Unix(stmt.ColumnInt64(1), 0)
				i++
				return nil
			},
		})
	if err != nil {
		return nil, utils.Stacktrace(err, "querying db_version rows")
	}

	return &result, nil
}
