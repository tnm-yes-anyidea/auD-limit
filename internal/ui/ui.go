package ui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"

	"github.com/tnm-yes-anyidea/auD-limit/internal/db"
	"github.com/tnm-yes-anyidea/auD-limit/internal/player"
)

// Theme describes a small colour palette for the TUI.
type Theme struct {
	Name       string
	Primary    lipgloss.TerminalColor
	Secondary  lipgloss.TerminalColor
	Background lipgloss.TerminalColor
	Accent     lipgloss.TerminalColor
}

var themes = []Theme{
	{"dark", "205", "240", "234", "63"},
	{"light", "21", "236", "252", "208"},
	{"retro", "3", "11", "0", "190"},
}

type model struct {
	tracks   []db.Track
	cursor   int
	quitting bool

	themeIdx int

	// progress bar (visual only for now; update from playback events later)
	p       progress.Model
	percent float64
	playing bool
	mpv     *player.MPV
}

func NewProgram(tracks []db.Track) (*tea.Program, error) {
	pb := progress.New(
		progress.WithDefaultGradient(),
	)
	pb.Width = 40

	m := model{
		tracks:  tracks,
		themeIdx: 0,
		p:       pb,
		percent: 0,
		playing: false,
	}

	// attempt to start mpv controller (non-fatal)
	if mpv, err := player.StartMPV(); err == nil {
		m.mpv = mpv
	}

	prg := tea.NewProgram(m, tea.WithAltScreen())
	return prg, nil
}

func (m model) Init() tea.Cmd {
	return tick()
}

type tickMsg time.Time
type toggleThemeMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*120, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		// simulate/drive visual progress while playing; this should be updated from actual playback position later
		if m.playing {
			if m.percent < 1.0 {
				m.percent += 0.005
			}
		} else {
			if m.percent > 0 {
				m.percent -= 0.001
				if m.percent < 0 {
					m.percent = 0
				}
			}
		}
		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.mpv != nil {
				m.mpv.Close()
			}
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.tracks)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "t":
			return m, func() tea.Msg { return toggleThemeMsg{} }
		case "enter":
			if len(m.tracks) > 0 {
				path := m.tracks[m.cursor].Path
				// envelope to stderr (keeps behaviour similar to original)
				fmt.Fprintf(os.Stderr, "play: %s\n", path)
				if m.mpv != nil {
					_ = m.mpv.Play(path) // best-effort playback via controller
				} else {
					// If controller wasn't available, attempt to start mpv standalone
					go func(p string) { _, _ = player.StartMPV() }(path)
				}
				m.playing = true
				m.percent = 0
			}
		}
	case toggleThemeMsg:
		m.themeIdx = (m.themeIdx + 1) % len(themes)
		accent := lipgloss.Color(themes[m.themeIdx].Accent)
		m.p.Gradient = progress.NewGradient(accent, accent)
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	th := themes[m.themeIdx]

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Primary)).
		Bold(true).
		Padding(0, 1)

	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(th.Background)).
		Foreground(lipgloss.Color(th.Accent)).
		Padding(0, 1)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Secondary)).
		Padding(0, 1)

	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(th.Primary)).
		Foreground(lipgloss.Color(th.Background)).
		Padding(0, 1)

	header := titleStyle.Render("auD-limit (Go) — Library")
	body := ""

	if len(m.tracks) == 0 {
		body = "(no tracks)\n"
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			infoStyle.Render(body),
			"",
			statusStyle.Render(fmt.Sprintf("Theme: %s  Tracks: %d  Playing: %t", th.Name, len(m.tracks), m.playing)),
			infoStyle.Render("q: quit  j/k: down/up  enter: play  t: cycle theme"),
		)
	}

	for i, t := range m.tracks {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		title := t.Title
		if title == "" {
			title = "(unknown)"
		}
		line := fmt.Sprintf("%s %s - %s", cursor, title, t.Artist)
		if m.cursor == i {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Render(line)
		}
		body += line + "\n"
	}

	// progress bar (visual)
	bar := m.p.ViewAs(m.percent)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		body,
		"",
		barStyle.Render(bar),
		infoStyle.Render(fmt.Sprintf("Progress: %0.0f%%", m.percent*100)),
		"",
		statusStyle.Render(fmt.Sprintf("Theme: %s  Tracks: %d  Playing: %t", th.Name, len(m.tracks), m.playing)),
		infoStyle.Render("q: quit  j/k: down/up  enter: play  t: cycle theme"),
	)
}
