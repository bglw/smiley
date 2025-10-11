package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	entries []logEntry
	width   int
	height  int
}

type logEntry struct {
	timestamp time.Time
	key       string
	keyType   string
	runes     []rune
	alt       bool
	paste     bool
}

func (e logEntry) String() string {
	parts := []string{
		fmt.Sprintf("[%s]", e.timestamp.Format("15:04:05.000")),
		fmt.Sprintf("Type: %-15s", e.keyType),
		fmt.Sprintf("String: %-20s", e.key),
	}

	if len(e.runes) > 0 {
		parts = append(parts, fmt.Sprintf("Runes: %v", e.runes))
	}

	var flags []string
	if e.alt {
		flags = append(flags, "ALT")
	}
	if e.paste {
		flags = append(flags, "PASTE")
	}
	if len(flags) > 0 {
		parts = append(parts, fmt.Sprintf("Flags: %s", strings.Join(flags, ",")))
	}

	return strings.Join(parts, " | ")
}

func initialModel() model {
	return model{
		entries: []logEntry{},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Log the key press
		entry := logEntry{
			timestamp: time.Now(),
			key:       msg.String(),
			keyType:   msg.Type.String(),
			runes:     msg.Runes,
			alt:       msg.Alt,
			paste:     msg.Paste,
		}
		m.entries = append(m.entries, entry)

		// Quit on ctrl+c
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("Key Logger - Press keys to see them logged (Ctrl+C to quit)\n")
	b.WriteString(strings.Repeat("=", m.width) + "\n\n")

	// Show the last entries that fit on screen
	start := 0
	if len(m.entries) > m.height-5 {
		start = len(m.entries) - (m.height - 5)
	}

	for i := start; i < len(m.entries); i++ {
		b.WriteString(m.entries[i].String())
		b.WriteString("\n")
	}

	// Show summary at bottom
	if len(m.entries) > 0 {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Total keys logged: %d", len(m.entries)))
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
