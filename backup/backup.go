package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local"
	"github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/remote"
	"github.com/wilymonkey/maeve/style"
	"github.com/wilymonkey/maeve/utils/myerr"
)

const descBackup = "Create backup files"
const descHashsums = "Calculating hashsums"

type Model struct {
	currOp   string
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
}

func New() Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(style.Spinner),
	)
	return Model{
		currOp:   descBackup,
		spinner:  s,
		progress: p,
	}
}

func (bm Model) Init() tea.Cmd {
	return tea.Batch(pullChanges, bm.spinner.Tick)
}

func (bm Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		bm.width, bm.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return bm, overseer.Back
		}

	case doneBackup:
		bm.currOp = descHashsums
		return bm, tea.Sequence(
			tea.Printf("%s %s", style.ITick, descBackup),
			newHashes,
		)

	case doneHashsums:
		bm.currOp = ""
		return bm, tea.Sequence(
			tea.Printf("%s %s", style.ITick, descHashsums),
			overseer.Back,
		)

	case spinner.TickMsg:
		var cmd tea.Cmd
		bm.spinner, cmd = bm.spinner.Update(msg)
		return bm, cmd

	case progress.FrameMsg:
		newModel, cmd := bm.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			bm.progress = newModel
		}
		return bm, cmd
	}
	return bm, nil
}

func (m Model) View() string {
	if m.currOp == "" {
		return style.Success.Margin(1, 4).Render("Done!")
	}

	spin := m.spinner.View() + " "
	prog := m.progress.View()

	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog))

	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(m.currOp)
	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog))
	gap := strings.Repeat(" ", cellsRemaining)

	help := style.Help.Render("Press q to stop the backup")

	return spin + info + gap + prog + "\n\n" + help
}

type doneBackup struct{}

// Pulls changes from Cfg.SourceDirs and writes hashes to file.
func pullChanges() tea.Msg {
	selfDir := cfg.Global.SelfDir()

	if err := os.RemoveAll(selfDir); err != nil {
		return myerr.NewErrMsg(err)
	}

	if err := os.MkdirAll(selfDir, 0755); err != nil {
		return myerr.NewErrMsg(err)
	}

	for _, srcDir := range cfg.Global.SourceDirs {
		destDir := filepath.Join(selfDir, filepath.Base(srcDir))
		if err := local.HardlinkDir(srcDir, destDir); err != nil {
			return myerr.NewErrMsg(err)
		}
	}

	return doneBackup{}
}

type doneHashsums struct{}

func newHashes() tea.Msg {
	selfDir := cfg.Global.SelfDir()
	if err := hashsums.NewDirFileHash(selfDir); err != nil {
		return myerr.NewErrMsg(err)
	}
	return doneHashsums{}
}

// Pushes changes to a given node address.
func sendChanges(address string) error {
	client, err := remote.NewSSHClient(address)
	if err != nil {
		return fmt.Errorf("new ssh client ⇒  %w", err)
	}
	defer client.Close()

	// TODO: Don't know what's happening here.

	return nil
}
