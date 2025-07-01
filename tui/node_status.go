package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type updateNodeStatus struct {
	status nodeStatus
}

type nodeStatus struct {
	name     string
	isDone   bool
	canSSH   bool
	canMaeve bool
}

func (ns *nodeStatus) View(b *strings.Builder, spinView string) {
	if ns.isDone {
		fmt.Fprintf(b, "\n%s: %s SSH %s Maeve", ns.name, boolView(ns.canSSH), boolView(ns.canMaeve))
	}
	fmt.Fprintf(b, "\n%s: %s", ns.name, spinView)
}

func (ns *nodeStatus) GetState() tea.Msg {
	return nil
}
