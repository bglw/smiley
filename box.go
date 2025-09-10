package main

import (
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
		b.h -= 1
	}

	if b.style.GetBorderBottom() {
		b.h -= 1
	}

	if b.style.GetBorderRight() {
		b.w -= 1
	}

	if b.style.GetBorderLeft() {
		b.w -= 1
	}

	return func() tea.Msg {
		return WindowSize{
			Loc:    b.ChildID,
			Height: b.h,
			Width:  b.w,
		}
	}
}

func (b Box) Update(msg tea.Msg) (Box, tea.Cmd) {
	var (
		cmds = []tea.Cmd{}
	)

	switch msg := msg.(type) {
	case WindowSize:
		if msg.Loc == b.ID {
			cmds = append(cmds, b.resize(msg))
		}
	}

	return b, tea.Batch(cmds...)
}

func (b Box) View() string {
	content := ""
	if b.Inner != nil {
		content = b.Inner.View()
	}

	vs := b.style.
		Height(b.h).
		Width(b.w).
		AlignVertical(lipgloss.Top).
		Render(content)

	return vs
}
