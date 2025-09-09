package main

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Status struct {
	w        int
	spinner  spinner.Model
	spinning bool
}

func NewStatus() Status {
	s := Status{
		spinner:  spinner.New(spinner.WithSpinner(spinner.Points)),
		spinning: true,
	}

	return s
}

func (s Status) Init() tea.Cmd {
	slog.Info("here")

	if s.spinning {
		return s.spinner.Tick
	}

	return nil
}

func (s Status) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds = []tea.Cmd{}
	)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		if s.spinning {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		s.w = msg.Width
	}

	return s, tea.Batch(cmds...)
}

func (s Status) View() string {
	bar := strings.Builder{}

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#17271a")).
		Background(lipgloss.Color("#dd9f6b"))

	bar.WriteString(barStyle.Render(" "))
	bar.WriteString(barStyle.Render(s.spinner.View()))
	bar.WriteString(barStyle.Bold(true).Render(" | "))
	bar.WriteString(barStyle.Render("this is a test of the emergency broadcast system"))

	return barStyle.
		Height(1).
		Width(s.w).
		Render(bar.String())
}
