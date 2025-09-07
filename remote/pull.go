package remote

import (
	"io"
	"os"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
	"github.com/wilymonkey/maeve/help"
)

func (n *NodeConn) PullDB() error {
	if err := n.addSFTP(); err != nil {
		return err
	}
	if err := n.addBackupDir(); err != nil {
		return err
	}
	sourceFile, err := n.sftpClient.Open(db.DBPath(n.backupDir))
	if err != nil {
		return help.WrapErr(err, "opening remote file")
	}
	defer sourceFile.Close()

	localFile, err := os.Create(db.DBPath(conf.MyNode()))
	if err != nil {
		return help.WrapErr(err, "opening local file")
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, sourceFile)
	if err != nil {
		return help.WrapErr(err, "pulling data from node")
	}

	return nil
}
