package main

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Static struct {
	id, content string
	Style       lipgloss.Style
	w, h        int
}

func NewStatic(id, content string, style lipgloss.Style) Static {
	return Static{
		id:      id,
		content: content,
		Style:   style,
	}
}

func (s Static) Init() tea.Cmd {
	return nil
}

func (s Static) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WindowSize:
		slog.Info("smsg", "msg", msg, "id", s.id)
		if msg.Loc == s.id {
			s.w = msg.Width
			s.h = msg.Height
		}
	}

	return s, nil
}

func (s Static) View() string {
	if s.w == 0 || s.h == 0 {
		return ""
	}

	return s.Style.
		Height(s.h).
		Width(s.w).
		MaxHeight(s.h).
		MaxWidth(s.w).
		Render(s.content)
}
