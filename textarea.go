package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/shlex"
)

type Textarea struct {
	id    string
	ta    textarea.Model
	focus bool
}

type FocusMsg bool

func NewTextarea(id string) Textarea {
	return Textarea{
		id:    id,
		ta:    textarea.New(),
		focus: true,
	}
}

func (t Textarea) Init() tea.Cmd {
	return textarea.Blink // BUG(tqbf): why
}

func (t Textarea) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	isSlash := func() (tea.Cmd, bool) {
		val := t.ta.Value()
		if strings.HasPrefix(val, "/") {
			tokens, _ := shlex.Split(strings.TrimSpace(val))
			t.ta.Reset()
			return func() tea.Msg {
				return msgSlashCommand(tokens)
			}, true
		}
		return nil, false
	}

	switch msg := msg.(type) {
	case msgInit:
		return t, t.Init()

	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyEnter:
			if cmd, ok := isSlash(); ok {
				return t, cmd
			}

		case key.Matches(msg, CurrentKeyMap.Send):
			if cmd, ok := isSlash(); ok {
				return t, cmd
			}

			val := t.ta.Value()
			cmds = append(cmds, func() tea.Msg {
				return msgInputSubmit(val)
			})
			t.ta.Reset()
		}

	case FocusMsg:
		if msg {
			t.ta.Focus()
			t.focus = true
		} else {
			t.focus = false
			t.ta.Blur()
		}

	case WindowSize:
		if msg.Loc == t.id {
			t.ta.SetWidth(msg.Width)
			t.ta.SetHeight(msg.Height)
			if t.focus {
				t.ta.Focus()
			}
		}
	}

	if !filterKey(msg, "pgup", "pgdown", "alt+up", "alt+down") {
		t.ta, cmd = t.ta.Update(msg)
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

func (t Textarea) View() string {
	return t.ta.View()
}
