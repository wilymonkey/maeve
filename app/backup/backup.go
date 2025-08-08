package backup

import (
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
)

func Backup(state guiState) error {
	conn, err := db.Open(conf.GetConf().MyNode())
	if err != nil {
		return err
	}
	_, err = db.GetVersion(conn)
	if err != nil {
		return err
	}
	return nil
}
