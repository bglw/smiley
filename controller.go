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
type msgModelResponse string
type msgWorking bool
type msgSelectContext string
type msgTokenUsage float64

type msgToolCall struct {
	name     string
	complete bool
	err      error
	size     int
	msg      string
}

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
	case msgModelResponse:
		cmds = append(cmds, viewLog(string(msg)+"\n", styleResponseText))
		cmds = append(cmds, t.tokenUsage())
		return t, tea.Batch(cmds...)

	case msgToolCall:
		cmds = []tea.Cmd{t.tokenUsage()}

		if *optLogTools {
			if !msg.complete {
				cmds = append(cmds, viewLog(string(msg.msg)+"\n", styleToolLogText))
			} else {
				cmds = append(cmds, viewLog(string(msg.msg)+"\n", styleToolResponseText))
			}
		}

		return t, tea.Batch(cmds...)

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
			t.tokenUsage(),
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
		slog.Info("llm call error", "error", err)
		return msgViewportLog{
			Msg:   err.Error() + "\n",
			Style: styleErrorText,
		}

	}

	return msgModelResponse(response)
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

func (t *LLMController) tokenUsage() tea.Cmd {
	usage, err := t.context.TokenUsage()

	if err != nil {
		return nil
	}

	return func() tea.Msg {
		return msgTokenUsage(usage.Percent)
	}
}
