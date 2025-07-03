package backup

import (
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/style"
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

	const linkTitle = "CREATE BACKUP FILES"
	if m.linkDone {
		b.WriteString(style.TitleSuccess.Render(linkTitle))
	} else {
		b.WriteString(spinView)
		b.WriteString(style.Title.Render(linkTitle))
	}
	b.WriteString("\n")
	b.WriteString(m.linkDirsView(&b))

	const hashTitle = "CALCULATING HASHSUMS"
	if m.hashDone {
		b.WriteString(style.TitleSuccess.Render(hashTitle))
	} else if m.linkDone {
		b.WriteString(spinView)
		b.WriteString(style.Title.Render(hashTitle))
	} else {
		b.WriteString(style.TitlePending.Render(hashTitle))
	}
	b.WriteString("\n")

	const sendTitle = "SENDING TO REMOTE NODES"
	if m.sendDone {
		b.WriteString(style.TitleSuccess.Render(sendTitle))
	} else if m.hashDone {
		b.WriteString(spinView)
		b.WriteString(style.Title.Render(sendTitle))
	} else {
		b.WriteString(style.TitlePending.Render(sendTitle))
	}
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(style.Help.Render("Press q to stop"))

	return b.String()
}

func (m *Model) linkDirsView(b *strings.Builder) string {

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
