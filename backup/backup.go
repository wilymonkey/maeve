package backup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/style"
	"github.com/wilymonkey/maeve/utils"
	"github.com/wilymonkey/maeve/utils/myerr"
)

type Model struct {
	linkDone bool
	hashDone bool
	sendDone bool
	linkDirs map[string]linkPathMeta
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	err      *myerr.ErrMsg
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
	linkDirs := make(map[string]linkPathMeta)
	for _, dir := range cfg.Global.SourceDirs {
		linkDirs[dir] = linkPathMeta{path: dir}
	}
	return Model{
		linkDirs: linkDirs,
		spinner:  s,
		progress: p,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return pullChanges(m.linkDirs) },
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, overseer.Back
		}

	case linkPathMeta:
		m.linkDirs[msg.path] = msg
		return m, nil

	case doneBackup:
		m.linkDone = true
		return m, newHashes

	case doneHashsums:
		m.hashDone = true
		return m, sendRemote

	case doneSend:
		m.sendDone = true
		if cfg.TuiInteractive {
			return m, nil
		} else {
			return m, overseer.Back
		}

	case myerr.ErrMsg:
		m.err = &msg
		if cfg.TuiInteractive {
			return m, nil
		} else {
			return m, overseer.Back
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	spinView := m.spinner.View()

	titleWithProgress(&b, "CREATE BACKUP FILES", spinView, m.linkDone, true)
	m.linkDirsView(&b)

	b.WriteString("\n")
	titleWithProgress(&b, "CALCULATING HASHSUMS", spinView, m.hashDone, m.linkDone)

	b.WriteString("\n")
	titleWithProgress(&b, "SENDING TO REMOTE NODES", spinView, m.sendDone, m.hashDone)

	if m.err != nil {
		m.err.Print(&b)
	}

	b.WriteString("\n")
	b.WriteString(style.Help.Render("Press q to stop"))

	return b.String()
}

func titleWithProgress(b *strings.Builder, title, spinView string, isDone, isPending bool) {
	if isDone {
		b.WriteString(style.TitleSuccess.Render(title))
	} else if isPending {
		fmt.Fprintf(b, "%s %s", spinView, style.Title.MarginTop(0).Render(title))
	} else {
		b.WriteString(style.TitlePending.Render(title))
	}
	b.WriteString("\n")
}

func (m *Model) linkDirsView(b *strings.Builder) {
	for _, dir := range m.linkDirs {
		path := utils.TruncateStr(dir.path, 20)
		size := utils.BytesToHuman(dir.size)
		fmt.Fprintf(b, "%s   Files: %d Size: %s\n", path, dir.number, size)
	}
}

type doneHashsums struct{}

func newHashes() tea.Msg {
	selfDir := cfg.Global.SelfDir()
	if err := hashsums.NewDirFileHash(selfDir); err != nil {
		return myerr.TuiMsg(err)
	}
	return doneHashsums{}
}

type doneSend struct{}

func sendRemote() tea.Msg {
	return doneSend{}
}
