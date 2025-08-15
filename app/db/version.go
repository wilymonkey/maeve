package db

import (
	"crypto/ed25519"
	"encoding/binary"
	"os"
	"time"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/help"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type DBVersion struct {
	LatestSnapshot time.Time
	Size           int64
	SizeSignature  []byte
}

func (v *DBVersion) IsValid(pubkey ed25519.PublicKey) bool {
	sizeMsg := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeMsg, uint64(v.Size))
	return ed25519.Verify(pubkey, sizeMsg, v.SizeSignature)
}

func GetVersion(conn *sqlite.Conn, dirPath string) (*DBVersion, error) {
	var result DBVersion
	var err error

	result.Size, err = getDBSize(&conn, dirPath)
	if err != nil {
		return nil, err
	}

	err = sqlitex.Execute(conn,
		"SELECT signature FROM size_signature WHERE id = 1;",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				stmt.ColumnBytes(0, result.SizeSignature)
				return nil
			},
		})
	if err != nil {
		return nil, help.Stacktrace(err, "querying total file_hash rows", help.DevError)
	}

	stmt := `SELECT DISTINCT snapshot 
		FROM file_hash 
		ORDER BY snapshot DESC 
		LIMIT 1;`
	err = sqlitex.Execute(conn, stmt,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				result.LatestSnapshot = time.Unix(stmt.ColumnInt64(0), 0)
				return nil
			},
		})
	if err != nil {
		return nil, help.Stacktrace(err, "querying db_version rows", help.DevError)
	}

	return &result, nil
}

func SetDBVersion(conn *sqlite.Conn, dirPath string) error {
	size, err := getDBSize(&conn, dirPath)
	if err != nil {
		return err
	}
	sizeMsg := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeMsg, uint64(size))
	signature := ed25519.Sign(conf.GetConf().SSHPrivateKey, sizeMsg)

	const stmt = `INSERT OR REPLACE INTO size_signature (id, signature) VALUES (1, ?)`
	err = sqlitex.ExecuteTransient(conn,
		stmt,
		&sqlitex.ExecOptions{
			Args: []any{signature},
		})
	if err != nil {
		return help.Stacktrace(err, "inserting db size signature", help.DevError)
	}
	return nil
}

// Will close the current connect to DB and reopen it.
func getDBSize(conn **sqlite.Conn, dirPath string) (int64, error) {
	dbPath := dbPath(dirPath)
	if err := flushWrites(*conn); err != nil {
		return 0, err
	}

	if err := (*conn).Close(); err != nil {
		return 0, help.Stacktrace(err, "closing db", help.DevError)
	}

	stat, err := os.Stat(dbPath)
	if err != nil {
		return 0, help.Stacktrace(err, "getting db stats", help.DevError)
	}

	*conn, err = sqlite.OpenConn(dbPath)
	if err != nil {
		return 0, help.Stacktrace(err, "opening db", help.DevError)
	}

	return stat.Size(), nil
}

func flushWrites(conn *sqlite.Conn) error {
	err := sqlitex.ExecuteTransient(conn, "PRAGMA wal_checkpoint(FULL);", nil)
	if err != nil {
		return help.Stacktrace(err, "creating checkpoint", help.DevError)
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA synchronous=FULL;", nil)
	if err != nil {
		return help.Stacktrace(err, "syncing DB", help.DevError)
	}
	err = sqlitex.ExecuteTransient(conn, "PRAGMA optimize;", nil)
	if err != nil {
		return help.Stacktrace(err, "optimising DB", help.DevError)
	}
	return nil
}
