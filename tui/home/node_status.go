package home

import (
	"bytes"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/ssh"
	"github.com/wilymonkey/maeve/tui/shared"
	"github.com/wilymonkey/maeve/tui/style"
)

type nodeStatus struct {
	name     string
	isDone   bool
	canSSH   bool
	canMaeve bool
	err      error
}

func (ns *nodeStatus) View(b *strings.Builder, spinView string) {
	name := style.Bright.Render(ns.name)
	if ns.isDone {
		fmt.Fprintf(b, "%s: %s SSH %s Maeve\n", name, shared.BoolView(ns.canSSH), shared.BoolView(ns.canMaeve))
	} else {
		fmt.Fprintf(b, "%s: %s\n", name, spinView)
	}
}

func (ns nodeStatus) withErr(err error) nodeStatus {
	ns.isDone = true
	ns.err = err
	return ns
}

func (ns nodeStatus) FetchState() tea.Msg {
	sshClient, err := ssh.NewSSHClient(ns.name)
	if err != nil {
		return ns.withErr(err)
	}
	defer sshClient.Close()
	ns.canSSH = true

	sesh, err := sshClient.NewSession()
	if err != nil {
		return ns.withErr(err)
	}
	defer sesh.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	sesh.Stdout = &stdoutBuf
	sesh.Stderr = &stderrBuf

	err = sesh.Run("maeve --version")
	if err != nil {
		return ns.withErr(err)
	}
	if stderrBuf.String() != "" {
		return ns.withErr(fmt.Errorf("stderrBuf: %s", stderrBuf.String()))
	}
	ns.canMaeve = true

	ns.isDone = true
	return ns
}
