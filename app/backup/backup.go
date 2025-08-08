package backup

import (
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/utils"
	"zombiezen.com/go/sqlite"
)

func Backup(state guiState) error {
	_, err := sqlite.OpenConn(conf.GetConf().MyNode())
	return utils.ErrContext("testing", utils.DummyErr("Woot!"))
	if err != nil {
		return utils.ErrContext("opening db connection", err)
	}
	return nil
}
