package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	"github.com/xldplx/plx/internal/cache"
	"github.com/xldplx/plx/internal/config"
	"github.com/xldplx/plx/internal/git"
	"github.com/xldplx/plx/internal/scanner"
)

type scanDoneMsg []*git.Repo
type errMsg error

// Model represents the Bubbletea state for plx.
type Model struct {
	cfg          *config.Config
	repos        []*git.Repo
	filtered     []*git.Repo
	cursor       int
	textInput    textinput.Model
	filtering    bool
	width        int
	height       int
	loading      bool
	statusMsg    string
	SelectedPath string
	Quitting     bool
}

// New creates a new TUI Model initialized with cached repos if available.
func New(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Filter repositories (fuzzy)..."
	ti.CharLimit = 50
	ti.Width = 40

	cached, _ := cache.Load()
	repos := []*git.Repo{}
	if cached != nil && len(cached.Repos) > 0 {
		repos = cached.Repos
	}

	m := Model{
		cfg:       cfg,
		repos:     repos,
		filtered:  repos,
		cursor:    0,
		textInput: ti,
		filtering: false,
		loading:   len(repos) == 0,
	}
	return m
}

// Init starts initial commands (e.g. background scan to freshen repo states).
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		triggerScan(m.cfg),
	)
}

func triggerScan(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		s := scanner.New(cfg)
		repos, err := s.Scan()
		if err != nil {
			return errMsg(err)
		}
		_ = cache.Save(repos)
		return scanDoneMsg(repos)
	}
}

// Update handles incoming messages and events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case scanDoneMsg:
		selectedPath := ""
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			selectedPath = m.filtered[m.cursor].Path
		}

		m.loading = false
		m.repos = msg
		m.applyFilter()

		// Preserve selection across rescans
		if selectedPath != "" {
			for i, r := range m.filtered {
				if r.Path == selectedPath {
					m.cursor = i
					break
				}
			}
		}
		m.statusMsg = fmt.Sprintf("Scanned %d repositories", len(msg))
		return m, nil

	case errMsg:
		m.loading = false
		m.statusMsg = fmt.Sprintf("Scan error: %v", msg)
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			switch msg.Type {
			case tea.KeyEsc:
				m.filtering = false
				m.textInput.Blur()
				return m, nil
			case tea.KeyEnter:
				m.filtering = false
				m.textInput.Blur()
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				m.applyFilter()
				m.cursor = 0
				return m, cmd
			}
		}

		// Normal navigation mode
		switch {
		case key.Matches(msg, keys.Quit):
			m.Quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Filter):
			m.filtering = true
			m.textInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, keys.Clear):
			if m.textInput.Value() != "" {
				m.textInput.SetValue("")
				m.applyFilter()
				m.cursor = 0
			}

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Select):
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.SelectedPath = m.filtered[m.cursor].Path
				m.Quitting = true
				return m, tea.Quit
			}

		case key.Matches(msg, keys.Editor):
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				target := m.filtered[m.cursor].Path
				editor := m.cfg.DefaultEditor
				if editor == "" {
					editor = "code"
				}
				c := exec.Command(editor, target)
				if runtime.GOOS == "windows" && editor == "code" {
					c = exec.Command("cmd", "/c", "code", target)
				}
				_ = c.Start()
				m.statusMsg = fmt.Sprintf("Opened %s in %s", m.filtered[m.cursor].Name, editor)
			}

		case key.Matches(msg, keys.Rescan):
			m.loading = true
			m.statusMsg = "Scanning workspaces..."
			return m, triggerScan(m.cfg)
		}
	}

	return m, nil
}

func (m *Model) applyFilter() {
	query := strings.TrimSpace(m.textInput.Value())
	if query == "" {
		m.filtered = m.repos
		if m.cursor >= len(m.filtered) {
			m.cursor = max(0, len(m.filtered)-1)
		}
		return
	}

	names := make([]string, len(m.repos))
	for i, r := range m.repos {
		names[i] = r.Name + " " + r.Path
	}

	matches := fuzzy.Find(query, names)
	matchedRepos := make([]*git.Repo, len(matches))
	for i, match := range matches {
		matchedRepos[i] = m.repos[match.Index]
	}

	m.filtered = matchedRepos
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}