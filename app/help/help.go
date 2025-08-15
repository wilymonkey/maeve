package help

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type Help int

const (
	AddKey Help = iota
	UpdateMaeve
	DelConfig
	DelKnownHost
	DelBackupDir
	DelDB
	DelPrivateKey
	CheckNodeConn
	DevError
)

func (c Help) String() string {
	switch c {
	case AddKey:
		return "add the Maeve Key to the backup PC"
	case UpdateMaeve:
		return "update Maeve"
	case DelConfig:
		return "delete the user data folder for maeve"
	case DelKnownHost:
		return "the PC we had once connected to has changed, if (and only if) you are certain it's fine, delete the PC entry in \"sshknownkeys\""
	case DelBackupDir:
		return "delete the problematic backup folder"
	case DelDB:
		return "delete the .db file in the backup folder"
	case DelPrivateKey:
		return "delete the private key"
	case CheckNodeConn:
		return "check if the backup pc (name, ip and port) is correct"
	case DevError:
		return "report this to the developer"
	default:
		return "...this shouldn't be possible"
	}
}

func Stacktrace(err error, task string, help Help) error {
	var b strings.Builder
	fmt.Fprintf(&b, "TRY: %s.\n", help.String())
	fmt.Fprintf(&b, "WAS DOING: %q.\nOUTCOME: %q.\n", task, err.Error())
	b.WriteString("SOURCE: ")
	WriteStacktrace(&b)
	b.WriteRune('.')
	return errors.New(b.String())
}

func WriteStacktrace(b *strings.Builder) {
	b.WriteString("...")
	before := b.Len()
	for i := 4; i > 1; i-- {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			continue
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		// Extract just the function name (without full package path).
		fnName := filepath.Base(fn.Name())
		b.WriteString(" →  ")
		b.WriteString(fnName)
	}
	after := b.Len()
	if before == after {
		panic("unable to build stacktrace")
	}
}
