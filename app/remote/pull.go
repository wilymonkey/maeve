package remote

import (
	"io"
	"os"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/db"
	"github.com/wilymonkey/maeve/app/help"
)

func (n *NodeConn) PullDB(node string) error {
	remoteDir, err := n.backupDir()
	if err != nil {
		return err
	}
	sourceFile, err := n.sftpClient.Open(db.DBPath(remoteDir))
	if err != nil {
		return help.CheckBackupDir(err, "opening remote file")
	}
	defer sourceFile.Close()

	localFile, err := os.Create(db.DBPath(conf.GetConf().MyNode()))
	if err != nil {
		return help.CheckBackupDir(err, "opening local file")
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, sourceFile)
	if err != nil {
		return help.CheckNodeConn(err, "pulling data from node")
	}

	return nil
}
