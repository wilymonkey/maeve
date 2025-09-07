package remote

import (
	"fmt"
	"io"
	"os"

	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/db"
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
		return fmt.Errorf("opening remote file: %w", err)
	}
	defer sourceFile.Close()

	localFile, err := os.Create(db.DBPath(conf.MyNode()))
	if err != nil {
		return fmt.Errorf("opening local file: %w", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, sourceFile)
	if err != nil {
		return fmt.Errorf("pulling data from node: %w", err)
	}

	return nil
}
