package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Status struct {
	w           int
	spinner     spinner.Model
	spinning    bool
	usage       float64
	currentTool string
	totalTools  int
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

	case msgTokenUsage:
		s.usage = float64(msg)

	case msgToolCall:
		if !msg.complete {
			s.currentTool = msg.name
		} else {
			s.currentTool = ""
			s.totalTools += 1
		}

	case tea.WindowSizeMsg:
		s.w = msg.Width
	}

	return s, tea.Batch(cmds...)
}

func (s Status) View() string {
	if s.w == 0 {
		return ""
	}

	rb := strings.Builder{}
	lb := strings.Builder{}

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#17271a")).
		Background(lipgloss.Color("#dd9f6b"))

	rb.WriteString(barStyle.Render(" "))
	rb.WriteString(barStyle.Render(s.spinner.View()))
	rb.WriteString(barStyle.Bold(true).Render(" | "))
	rb.WriteString(barStyle.Render(fmt.Sprintf("%d%% full", int(math.Round(s.usage*100)))))

	lStyle := barStyle.Align(lipgloss.Right)

	lb.WriteString(lStyle.Render(s.currentTool))
	lb.WriteString(lStyle.Render(" | "))
	lb.WriteString(lStyle.Render(fmt.Sprintf("%d tools", s.totalTools)))

	var (
		rt  = rb.String()
		lt  = lb.String()
		rl  = lipgloss.Width(rt)
		ll  = lipgloss.Width(lt)
		pad = s.w - rl - ll
	)

	rb.WriteString(barStyle.Render(strings.Repeat(" ", pad)))
	rb.WriteString(lt)

	return barStyle.
		Height(1).
		Width(s.w).
		Render(rb.String())
}
