package ui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tnm-yes-anyidea/auD-limit/internal/db"
)

var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

type model struct{
	tracks []db.Track
	cursor int
	quitting bool
}

func NewProgram(tracks []db.Track) (*tea.Program, error) {
	m := model{tracks: tracks}
	p := tea.NewProgram(m)
	return p, nil
}

func (m model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return t })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.tracks)-1 { m.cursor++ }
		case "k", "up":
			if m.cursor > 0 { m.cursor-- }
		case "enter":
			if len(m.tracks) > 0 {
				// TODO: play
				fmt.Fprintf(os.Stderr, "play: %s
", m.tracks[m.cursor].Path)
			}
		}
	case time.Time:
		// tick - ignore for now
	}
	return m, nil
}

func (m model) View() string {
	s := titleStyle.Render("auD-limit (Go) — Library") + "

"
	if len(m.tracks) == 0 { s += "(no tracks)
"; return s }
	for i, t := range m.tracks {
		cursor := " "
		if m.cursor == i { cursor = ">" }
		title := t.Title
		if title == "" { title = "(unknown)" }
		s += fmt.Sprintf("%s %s — %s
", cursor, title, t.Artist)
	}
	s += "
q: quit  j/k: down/up  enter: play (mpv)
"
	return s
}

// Start runs the TUI program (blocking)
func (p *tea.Program) Start() error {
	return p.StartReturningProgram().Run()
}

// helper to adapt to our return pattern
func (p *tea.Program) StartReturningProgram() *tea.Program { return p }