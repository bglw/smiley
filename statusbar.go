package main

import (
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

func freshSpinner() spinner.Model {
	return spinner.New(spinner.WithSpinner(spinner.Points))
}

func NewStatus() Status {
	s := Status{
		spinner:  freshSpinner(),
		spinning: false,
	}

	return s
}

func (s Status) Init() tea.Cmd {
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
	case msgInit:
		return s, s.Init()

	case msgWorking:
		if msg == true {
			s.spinning = true
			cmds = append(cmds, s.spinner.Tick)
		} else {
			s.spinning = false
		}

	case spinner.TickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		if s.spinning {
			cmds = append(cmds, cmd)
		} else {
			s.spinner = freshSpinner()
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
