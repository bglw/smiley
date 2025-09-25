package main

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/superfly/contextwindow"

	"smiley/agent"
)

// controllers are thingies that do the Update protocol but not the View protocol.

type msgInputSubmit string
type msgPromptUpdate string
type msgModelResponse string
type msgWorking bool
type msgSelectContext string
type msgTokenUsage float64
type msgSlashCommand []string

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
	// magenta
	stylePromptText = lipgloss.NewStyle().Foreground(lipgloss.Color("#a586d9"))

	// bright white
	styleResponseText = lipgloss.NewStyle().Foreground(lipgloss.Color("#e9f5ea"))

	// bright black
	styleToolLogText      = lipgloss.NewStyle().Foreground(lipgloss.Color("#3c5a42")).Faint(true)
	styleToolResponseText = lipgloss.NewStyle().Foreground(lipgloss.Color("#3c5a42"))

	// bright yellow
	styleErrorText = lipgloss.NewStyle().Foreground(lipgloss.Color("#dd9f6b")).Faint(true)

	// bright pink?
	styleSlashResult = lipgloss.NewStyle().Foreground(lipgloss.Color("#e592ab"))
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

type TUIAgentController struct {
	agent *agent.Agent
}

func (t *TUIAgentController) Update(msg tea.Msg) (Controller, tea.Cmd) {
	switch msg := msg.(type) {
	case msgModelResponse:
		return t, viewLog(string(msg)+"\n", styleResponseText)

	case msgToolCall:
		if *optLogTools {
			if !msg.complete {
				return t, viewLog(string(msg.msg)+"\n", styleToolLogText)
			} else {
				return t, viewLog(string(msg.msg)+"\n", styleToolResponseText)
			}
		}
		return t, nil

	case msgSelectContext:
		return t.selectContext(string(msg))

	case msgPromptUpdate:
		return t, func() tea.Msg {
			err := t.agent.SendPrompt(string(msg))
			if err != nil {
				return msgViewportLog{
					Msg:   err.Error() + "\n",
					Style: styleErrorText,
				}
			}
			return nil
		}
	}

	return t, nil
}

func (t *TUIAgentController) selectContext(name string) (Controller, tea.Cmd) {
	err := t.agent.SwitchContext(name)
	if err != nil {
		slog.Error("switch context", "error", err)
		return t, nil
	}

	records, err := t.agent.GetContextWindow().LiveRecords()
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
