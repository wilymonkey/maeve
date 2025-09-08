package db

import (
	"crypto/ed25519"
	"fmt"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/proto"
	"github.com/zeebo/blake3"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func GetVersion(conn *sqlite.Conn) (*proto.DBVersion, error) {
	var result proto.DBVersion
	var err error

	result.Hash, err = hashDB(conn)
	if err != nil {
		return nil, err
	}

	err = sqlitex.ExecuteTransient(conn,
		"SELECT sign FROM hash_sign WHERE id = 1;",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				stmt.ColumnBytes(0, result.HashSign)
				return nil
			},
		})
	if err != nil {
		return nil, fmt.Errorf("finding table version signature: %w", err)
	}

	err = sqlitex.ExecuteTransient(conn,
		`SELECT DISTINCT snapshot 
		FROM path_meta_link
		ORDER BY snapshot DESC 
		LIMIT 1;`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				result.LatestSnapshot = stmt.ColumnInt64(0)
				return nil
			},
		})
	if err != nil {
		return nil, fmt.Errorf("getting latest snapshot: %w", err)
	}

	err = sqlitex.ExecuteTransient(conn,
		"SELECT COUNT(*) FROM path_meta_link;",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				result.Rows = stmt.ColumnInt64(0)
				return nil
			},
		})
	if err != nil {
		return nil, fmt.Errorf("counting path_meta_link rows: %w", err)
	}

	return &result, nil
}

func SetDBVersion(conn *sqlite.Conn, dirPath string) error {
	hash, err := hashDB(conn)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(conf.GetConf().PrivKey, hash)

	err = sqlitex.ExecuteTransient(conn,
		"INSERT OR REPLACE INTO size_signature (id, signature) VALUES (1, ?)",
		&sqlitex.ExecOptions{
			Args: []any{sig},
		})
	if err != nil {
		return fmt.Errorf("inserting db signature: %w", err)
	}
	return nil
}

func hashDB(conn *sqlite.Conn) ([]byte, error) {
	maxRows := 100
	hasher := blake3.New()

	var total int
	err := sqlitex.ExecuteTransient(conn,
		"SELECT COUNT(*) FROM file_meta;",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				total = stmt.ColumnInt(0)
				return nil
			},
		})
	if err != nil {
		return nil, fmt.Errorf("counting file_meta rows: %w", err)
	}

	addHashes := func(stmt *sqlite.Stmt) error {
		var savedHash []byte
		stmt.ColumnBytes(0, savedHash)
		hasher.Write(savedHash)
		return nil
	}
	if total <= maxRows {
		err := sqlitex.ExecuteTransient(conn,
			"SELECT hash FROM file_meta ORDER BY hash;",
			&sqlitex.ExecOptions{ResultFunc: addHashes},
		)
		if err != nil {
			return nil, fmt.Errorf("hashing rows: %w", err)
		}
		return hasher.Sum(nil), nil
	}

	endTx, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return nil, fmt.Errorf("creating immediate transaction: %w", err)
	}
	defer endTx(&err)

	for _, offset := range evenOffsets(total, maxRows) {
		err = sqlitex.Execute(conn, "SELECT hash FROM file_meta ORDER BY hash LIMIT 1 OFFSET ?;",
			&sqlitex.ExecOptions{
				Args:       []any{offset},
				ResultFunc: addHashes,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("selecting evenly spread hashsums: %w", err)
		}
	}

	return hasher.Sum(nil), err
}

func evenOffsets(dataLen, maxRows int) []int {
	offsets := make([]int, maxRows)
	step := dataLen / maxRows
	remainder := dataLen % maxRows

	offset := 0
	for i := range maxRows {
		offsets[i] = offset
		offset += step
		if remainder > 0 {
			offset++
			remainder--
		}
	}
	return offsets
}
