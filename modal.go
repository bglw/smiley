package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Modal struct {
	width  int
	height int
}

func NewModal() *Modal {
	return &Modal{
		width:  50,
		height: 10,
	}
}

func (m *Modal) Init() tea.Cmd {
	return nil
}

func (m *Modal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Modal) View() string {
	content := `Hello World!

This is a test modal window.
It has multiple lines of static content.

Press ctrl+k to dismiss this modal.

This is just a proof of concept
to demonstrate that modals work
in this codebase.`

	modalStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("230")).
		Align(lipgloss.Left, lipgloss.Top)

	return modalStyle.Render(content)
}
