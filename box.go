package main

import (
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Box struct {
	Inner       tea.Model
	ID, ChildID string
	style       lipgloss.Style
	w, h        int
}

func NewBox(id, childid string, t, r, b, l bool) Box {
	return Box{
		ChildID: childid,
		ID:      id,
		style: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), t, r, b, l),
	}
}

func (b Box) Init() tea.Cmd {
	return nil
}

func (b *Box) resize(msg WindowSize) tea.Cmd {
	b.w = msg.Width
	b.h = msg.Height

	if b.style.GetBorderTop() {
		msg.Height -= 1
	}

	if b.style.GetBorderBottom() {
		msg.Height -= 1
	}

	if b.style.GetBorderRight() {
		msg.Width -= 1
	}

	if b.style.GetBorderLeft() {
		msg.Width -= 1
	}

	return func() tea.Msg {
		return WindowSize{
			Loc:    b.ChildID,
			Height: msg.Height,
			Width:  msg.Width,
		}
	}
}

func (b Box) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds = []tea.Cmd{}
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case WindowSize:
		slog.Info("heyo", "msg", msg, "id", b.ID)

		if msg.Loc == b.ID {
			cmds = append(cmds, b.resize(msg))
		}
	}

	if b.Inner != nil {
		b.Inner, cmd = b.Inner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return b, tea.Batch(cmds...)
}

func (b Box) View() string {
	slog.Info("box", "h", b.h, "w", b.w)

	content := ""
	if b.Inner != nil {
		content = b.Inner.View()
	}

	slog.Info("viewport content first line", "line", strings.Split(content, "\n")[0])

	vs := b.style.
		Height(b.h).
		Width(b.w).
		AlignVertical(lipgloss.Top).
		Render(content)

	slog.Info("lipgloss rendered first line", "line", strings.Split(vs, "\n")[0])

	slog.Info("box", "newlines", strings.Count(vs, "\n"))
	slog.Info("box", "styleheight", lipgloss.Height(vs))

	return vs
}
