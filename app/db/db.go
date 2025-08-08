package db

import (
	"fmt"
	"path"

	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
)

func Open(dirPath string) (*sqlite.Conn, error) {
	dbPath := path.Join(dirPath, "hashtable.db")
	conn, err := sqlite.OpenConn(dbPath)
	if err != nil {
		ctxMsg := fmt.Sprintf("opening dbPath: %s", dbPath)
		return nil, utils.ErrContext(ctxMsg, err)
	}

	if err := createSchema(conn); err != nil {
		return nil, err
	}

	return conn, nil
}
