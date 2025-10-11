package main

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/superfly/contextwindow"
)

type msgContextTableRows []table.Row

type DatabaseView struct {
	id      string
	cw      *contextwindow.ContextWindow
	w, h    int
	table   table.Model
	focused bool
}

func NewDatabaseView(id string, cw *contextwindow.ContextWindow) *DatabaseView {
	cols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Tokens", Width: 10},
		{Title: "Start", Width: 20},
		{Title: "Duration", Width: 20},
	}

	return &DatabaseView{
		id: id,
		cw: cw,
		table: table.New(
			table.WithFocused(true),
			table.WithColumns(cols),
		),
	}
}

func humanDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	var (
		parts   []string
		days    = int(d.Hours()) / 24
		hours   = int(d.Hours()) % 24
		minutes = int(d.Minutes()) % 60
	)

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d day%s", days, plural(days)))
	}

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hour%s", hours, plural(hours)))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d minute%s", minutes, plural(minutes)))
	}

	if len(parts) == 0 {
		parts = append(parts, "less than a minute")
	}
	return strings.Join(parts, ", ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func (s *DatabaseView) refreshTable() tea.Msg {
	contexts, err := s.cw.ListContexts()
	if err != nil {
		slog.Info("refresh context", "error", err)
		// TODO(tqbf): percolate an error
		return nil
	}

	rows := []table.Row{}

	for _, context := range contexts {
		stats, err := s.cw.GetContextStats(context)
		if err != nil {
			slog.Info("fetch context", "context", context, "error", err)
			continue
		}

		var (
			name       = context.Name
			tokens     = stats.LiveTokens
			start      = context.StartTime
			activeTime = "(no activity)"
		)

		if tokens == 0 {
			continue
		}

		if stats.LastActivity != nil {
			activeTime = humanDuration(start.Sub(*stats.LastActivity))
		}

		rows = append(rows, []string{
			name,
			fmt.Sprintf("%d tokens", tokens),
			fmt.Sprintf("%s", start),
			activeTime,
		})
	}

	return msgContextTableRows(rows)
}

func (s *DatabaseView) Init() tea.Cmd {
	return s.refreshTable
}

func (s *DatabaseView) Update(msg tea.Msg) (ret tea.Model, cmd tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgFocusChanged:
		s.focused = (msg.region == "top")

	case msgInit:
		return s, s.Init()

	case msgContextTableRows:
		s.table.SetRows([]table.Row(msg))

	case tea.KeyMsg:
		if !s.focused {
			break
		}

		switch msg.String() {
		case "down":
			s.table.MoveDown(1)
		case "up":
			s.table.MoveUp(1)
		case "enter":
			row := s.table.SelectedRow()
			if len(row) > 0 {
				name := row[0]
				cmds = append(cmds, func() tea.Msg {
					return msgSelectContext(name)
				})
			}
		}

	case WindowSize:
		if msg.Loc == s.id {
			s.w = msg.Width
			s.h = msg.Height
			s.table.SetHeight(s.h)
			s.table.SetWidth(s.w)
		}
	}

	if len(cmds) > 0 {
		cmd = tea.Sequence(cmds...)
	}

	return s, cmd
}

func (s *DatabaseView) View() string {
	if s.w == 0 || s.h == 0 {
		return ""
	}

	slog.Info("table list")
	for _, row := range s.table.Rows() {
		slog.Info("table", "row", row)
	}

	v := lipgloss.NewStyle().
		Render(s.table.View())
	return v
}
