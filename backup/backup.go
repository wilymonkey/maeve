package backup

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilymonkey/maeve/cfg"
	hs "github.com/wilymonkey/maeve/local/hashsums"
	"github.com/wilymonkey/maeve/overseer"
	"github.com/wilymonkey/maeve/remote"
	"github.com/wilymonkey/maeve/style"
	"github.com/wilymonkey/maeve/utils"
	"github.com/wilymonkey/maeve/utils/myerr"
)

type Model struct {
	pullDone    bool
	linkDirs    map[string]linkPathMeta
	hashDone    bool
	totalFiles  int64
	hashProg    []hs.FileHash
	pushProg    pushProgress
	pushingNode int
	pushDone    bool
	width       int
	height      int
	spinner     spinner.Model
	progress    progress.Model
	err         error
	ctx         context.Context
	ctxCancel   context.CancelFunc
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
	ctx, ctxCancel := context.WithCancel(context.Background())
	return Model{
		linkDirs:  linkDirs,
		spinner:   s,
		progress:  p,
		ctx:       ctx,
		ctxCancel: ctxCancel,
		pushProg:  newPushProgress(0),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Sequence(
		m.spinner.Tick,
		func() tea.Msg { return pullChanges(m.linkDirs, m.ctx) },
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			if m.ctx.Err() == nil {
				m.ctxCancel()
				return m, nil
			}
			return m, overseer.Back
		}

	case linkPathMeta:
		m.linkDirs[msg.path] = msg
		var totalFiles int64
		for _, value := range m.linkDirs {
			totalFiles += value.number
		}
		m.totalFiles = totalFiles
		return m, nil

	case doneLinks:
		m.pullDone = true
		return m, func() tea.Msg { return newHashes(m.ctx, m.totalFiles) }

	case hashProg:
		m.hashProg = msg.hashes
		return m, nil

	case doneHashsums:
		m.hashDone = true
		return m, func() tea.Msg { return pushChanges(m.ctx, m.hashProg) }

	case pushingNode:
		m.pushingNode = msg.index
		return m, nil

	case pushProgress:
		m.pushProg = msg
		return m, nil

	case doneSend:
		m.pushDone = true
		m.ctxCancel()
		if cfg.TuiInteractive {
			return m, nil
		} else {
			return m, overseer.Back
		}

	case error:
		m.err = msg
		m.ctxCancel()
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

	b.WriteString(titleWithProgress("CREATE BACKUP FILES", spinView, m.pullDone, true))
	m.linkDirsView(&b)

	b.WriteString("\n")
	b.WriteString(titleWithProgress("CALCULATING HASHSUMS", spinView, m.hashDone, m.pullDone))
	perc := float32(len(m.hashProg)) / float32(m.totalFiles) * 100
	fmt.Fprintf(&b, "%d / %d: %0.2f%%\n", len(m.hashProg), m.totalFiles, perc)

	b.WriteString("\n")
	b.WriteString(titleWithProgress("SENDING TO REMOTE NODES", spinView, m.pushDone, m.hashDone))
	m.pushView(&b, spinView)

	if m.err != nil {
		myerr.Print(m.err, &b)
	}

	b.WriteString("\n")
	if m.ctx.Err() != nil {
		b.WriteString(style.Success.Render("Press q to go back"))
	} else {
		b.WriteString(style.Fail.Render("Press q to stop backup"))
	}

	return b.String()
}

func titleWithProgress(title, spinView string, isDone, isPending bool) string {
	var s string
	if isDone {
		s = fmt.Sprintf("%s %s", style.ITick, style.TitleSuccess.Render(title))
	} else if isPending {
		s = fmt.Sprintf(" %s %s", spinView, style.Title.Render(title))
	} else {
		s = style.TitlePending.Render(title)
	}
	return style.My.Render(s) + "\n"
}

func (m *Model) linkDirsView(b *strings.Builder) {
	for _, dirPath := range cfg.Global.SourceDirs {
		dir := m.linkDirs[dirPath]
		path := utils.TruncateStr(dir.path, 30)
		size := utils.BytesToHuman(dir.size)
		fmt.Fprintf(b, "%s   Files: %d Size: %s\n", path, dir.number, size)
	}
}

func (m *Model) pushView(b *strings.Builder, spinView string) {
	for i, node := range cfg.Global.RemoteNodes {
		if m.pushingNode == i {
			fmt.Fprintf(b, "%s %s", node, spinView)
		} else {
			fmt.Fprintf(b, style.Fade.Render(node))
		}
		b.WriteString("\n")
	}

	columns := []table.Column{
		{Title: "File", Width: 20},
		{Title: "Size", Width: 5},
		{Title: "%", Width: 5},
		{Title: "Verified", Width: 5},
	}
	var rows []table.Row
	m.pushProg.operations.ForEach(func(item remote.SendStatus) {
		perc := float32(item.Curr) / float32(item.Total) * 100
		var verified string
		if item.Verifying {
			verified = spinView
		} else if item.IsGood {
			verified = style.ITick
		}
		r := table.Row{
			item.Hash.Path.Path,
			utils.BytesToHuman(item.Total),
			fmt.Sprintf("%0.2f%%", perc),
			verified,
		}
		rows = append(rows, r)
	})

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(7),
	)
	s := table.DefaultStyles()
	s.Header = style.Table
	t.SetStyles(s)
	if m.hashDone && !m.pushDone {
		b.WriteString(t.View())
	}
}

type hashProg struct {
	hashes []hs.FileHash
}
type doneHashsums struct{}

func newHashes(parentCtx context.Context, total int64) tea.Msg {
	selfDir := cfg.Global.SelfDir()

	progChan := make(chan hs.FileHash, 100)
	hashes := make([]hs.FileHash, 0, total)
	utils.Throttle(
		progChan,
		func(hash hs.FileHash) {
			hashes = append(hashes, hash)
		},
		func() {
			cfg.TuiProgram.Send(hashProg{hashes: hashes})
		},
	)

	if err := hs.NewDirFileHash(selfDir, progChan, parentCtx); err != nil {
		return myerr.WrapErr(err)
	}

	close(progChan)
	return doneHashsums{}
}
