package main

import (
	"context"
	"database/sql"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/superfly/contextwindow"
)

type msgInputSubmit string
type msgPromptUpdate string
type msgToolCall string
type msgToolResult string

type TextAreaInput struct {
}

var (
	stylePromptText       = lipgloss.NewStyle().Foreground(lipgloss.Color("#a586d9"))
	styleResponseText     = lipgloss.NewStyle().Foreground(lipgloss.Color("#e9f5ea"))
	styleToolLogText      = lipgloss.NewStyle().Foreground(lipgloss.Color("#3c5a42")).Faint(true)
	styleToolResponseText = lipgloss.NewStyle().Foreground(lipgloss.Color("#3c5a42"))
	styleErrorText        = lipgloss.NewStyle().Foreground(lipgloss.Color("#dd9f6b")).Faint(true)
)

func viewLog(msg string, style lipgloss.Style) tea.Cmd {
	return func() tea.Msg {
		return msgViewportLog{
			Msg:   msg,
			Style: style,
		}
	}
}

func (t *TextAreaInput) Update(msg tea.Msg) (Controller, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	_ = cmd

	switch msg := msg.(type) {
	case msgInputSubmit:
		cmds = append(cmds, viewLog(string(msg), stylePromptText))
		cmds = append(cmds, func() tea.Msg {
			return msgPromptUpdate(string(msg))
		})
	}

	return t, tea.Batch(cmds...)
}

type LLMController struct {
	lock    sync.Mutex
	model   contextwindow.Model
	context *contextwindow.ContextWindow
	db      *sql.DB
}

func (t *LLMController) Update(msg tea.Msg) (Controller, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgPromptUpdate:
		t.lock.Lock()
		t.context.AddPrompt(string(msg))
		t.lock.Unlock()

		cmds = append(cmds, t.callModel)
	}

	return t, tea.Batch(cmds...)
}

func (t *LLMController) callModel() tea.Msg {
	t.lock.Lock()
	defer t.lock.Unlock()

	response, err := t.context.CallModel(context.TODO())

	if err != nil {
		return msgViewportLog{
			Msg:   err.Error() + "\n",
			Style: styleErrorText,
		}

	}

	return msgViewportLog{
		Msg:   response + "\n",
		Style: styleResponseText,
	}
}
