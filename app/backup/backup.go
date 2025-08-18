package backup

import (
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
)

func Backup() error {
	conn, err := db.Open(conf.GetConf().MyNode())
	if err != nil {
		return err
	}

	if err = repairDB(conn); err != nil {
		return err
	}

	return nil
}
