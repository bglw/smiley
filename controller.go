package main

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/superfly/contextwindow"
)

// controllers are thingies that do the Update protocol but not the View protocol.

type msgInputSubmit string
type msgPromptUpdate string
type msgToolCall string
type msgToolResult string
type msgWorking bool
type msgSelectContext string

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

type Controller interface {
	Update(msg tea.Msg) (Controller, tea.Cmd)
}

type Controllers []Controller

func (cs Controllers) Update(msg tea.Msg) (Controller, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	for i := range cs {
		c, cmd := cs[i].Update(msg)
		cs[i] = c
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		cmd = tea.Sequence(cmds...)
	}

	return cs, cmd
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
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case msgToolCall:
		return t, viewLog(string(msg)+"\n", styleToolLogText)

	case msgToolResult:
		return t, viewLog(string(msg)+"\n", styleToolResponseText)

	case msgSelectContext:
		return t.selectContext(string(msg))

	case msgPromptUpdate:
		t.lock.Lock()
		t.context.AddPrompt(string(msg))
		t.lock.Unlock()

		cmds = append(cmds, []tea.Cmd{
			func() tea.Msg {
				return msgWorking(true)
			},
			t.callModel,
			func() tea.Msg {
				return msgWorking(false)
			},
		}...)
	}

	if len(cmds) > 0 {
		cmd = tea.Sequence(cmds...)
	}

	return t, cmd
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

func (t *LLMController) selectContext(name string) (Controller, tea.Cmd) {
	t.lock.Lock()
	defer t.lock.Unlock()

	err := t.context.SwitchContext(name)
	if err != nil {
		slog.Error("switch context", "error", err)
		return t, nil
	}

	records, err := t.context.LiveRecords()
	if err != nil {
		slog.Error("read records", "error", err)
		return t, nil
	}

	resetMsg := []msgViewportLog{}
	for _, r := range records {
		switch r.Source {
		case contextwindow.Prompt:
			resetMsg = append(resetMsg, msgViewportLog{
				Style: stylePromptText,
				Msg:   r.Content,
			})
		case contextwindow.ModelResp:
			resetMsg = append(resetMsg, msgViewportLog{
				Style: styleResponseText,
				Msg:   r.Content,
			})
		case contextwindow.ToolCall:
			resetMsg = append(resetMsg, msgViewportLog{
				Style: styleToolLogText,
				Msg:   r.Content,
			})
		case contextwindow.ToolOutput:
			resetMsg = append(resetMsg, msgViewportLog{
				Style: styleToolResponseText,
				Msg:   r.Content,
			})
		}
	}

	return t, tea.Sequence(
		func() tea.Msg {
			return msgResetViewport(resetMsg)
		},
		func() tea.Msg {
			return msgSwitchScreen(screenLog)
		},
	)
}
