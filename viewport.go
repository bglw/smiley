package main

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Viewport struct {
	Content string
	ID      string
	vm      viewport.Model
	w, h    int
}

func NewViewport(id, content string) Viewport {
	lines := strings.Split(content, "\n")
	slog.Info("viewport init", "totallines", len(lines), "firstline", lines[0])
	
	v := Viewport{
		ID: id,
		vm: viewport.New(0, 0),
	}
	v.vm.SetContent(content)
	v.vm.GotoTop()

	return v
}

func (v Viewport) Init() tea.Cmd {
	return nil
}

func (v Viewport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case WindowSize:
		if msg.Loc == v.ID {
			v.vm.Height = msg.Height
			v.vm.Width = msg.Width
		}
	}

	v.vm, cmd = v.vm.Update(msg)

	return v, cmd
}

func (v Viewport) View() string {
	slog.Info("view", "h", v.vm.Height, "w", v.vm.Width)

	if v.vm.Height == 0 || v.vm.Width == 0 {
		return ""
	}

	vs := v.vm.View()
	slog.Info("view", "newlines", strings.Count(vs, "\n"))
	slog.Info("view", "visiblelinecount", v.vm.VisibleLineCount())
	slog.Info("view", "styleheight", lipgloss.Height(vs))

	return vs
}
