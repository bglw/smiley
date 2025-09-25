package main

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/superfly/contextwindow"
)

const (
	screenLog = iota
	screenHistory
)

// subwindow height
type WindowSize struct {
	Height int
	Width  int
	Loc    string
}

type rootWindow struct {
	initialPrompt  string
	contextName    string
	p              *tea.Program
	status, bottom tea.Model
	state          int
	top            Box
	history        *DatabaseView
	log            *Viewport
	controllers    Controller
	w, h, th, bh   int // top height, bottom height

	db *sql.DB
}

type msgInit struct{}
type msgSwitchScreen int

func newRootWindow(
	content string,
	cw *contextwindow.ContextWindow,
	initialPrompt string,
	contextName string,
) rootWindow {
	m := rootWindow{}

	m.contextName = contextName
	m.initialPrompt = initialPrompt + "\n"

	m.status = NewStatus()

	m.history = NewDatabaseView("top-inner", cw)
	m.log = NewViewport("top-inner", content)

	box := NewBox("top", "top-inner", false, false, true, false)
	box.Inner = m.log
	m.top = box
	m.state = screenLog

	input := NewTextarea("bottom")
	m.bottom = input

	return m
}

func (m rootWindow) Init() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return msgInit{}
		},
		func() tea.Msg {
			if m.contextName != "" {
				return msgSelectContext(m.contextName)
			}

			return nil
		},
		func() tea.Msg {
			if m.initialPrompt != "\n" {
				slog.Info("have initial prompt", "prompt", m.initialPrompt)

				// i should be purged i should be flogged
				time.Sleep(500 * time.Millisecond)

				// BUG(tqbf): that sleep actually fixes a problem
				// we were having where the initial prompt, if
				// provided, generated an OpenAI API error about
				// not having the previous-conversation field
				// set on that first prompt. Obviously, need
				// figure that the hell out.

				return msgInputSubmit(m.initialPrompt)
			}

			return nil
		},
	)

}

func (m *rootWindow) resize(w, h int) tea.Cmd {
	if m.h == h && m.w == w {
		return nil
	}

	m.h = h
	m.w = w

	m.th = round(m.h, 90)
	m.bh = round(m.h, 10)

	slog.Info("main dims", "w", m.w, "h", m.h, "th", m.th, "bh", m.bh)

	return tea.Batch(
		func() tea.Msg {
			return WindowSize{
				Width:  w,
				Height: m.th - 1,
				Loc:    "top",
			}
		},
		func() tea.Msg {
			return WindowSize{
				Width:  w,
				Height: m.bh,
				Loc:    "bottom",
			}
		})
}

func (m rootWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds = []tea.Cmd{}
		cmd  tea.Cmd
		keyb bool
	)

	swtch := func(state int) (tea.Model, tea.Cmd) {
		switch state {
		case screenHistory:
			m.top.Inner = m.history
			m.state = screenHistory
			return m, nil
		case screenLog:
			m.top.Inner = m.log
			m.state = screenLog
			return m, nil
		default:
			panic("bad state")
		}
	}

	switch msg := msg.(type) {
	case msgSwitchScreen:
		return swtch(int(msg))

	case tea.KeyMsg:
		slog.Info("keypress", "key", msg)
		keyb = true

		switch {
		case key.Matches(msg, CurrentKeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, CurrentKeyMap.History):
			return swtch(screenHistory)
		case key.Matches(msg, CurrentKeyMap.Log):
			return swtch(screenLog)
		}

	case tea.WindowSizeMsg:
		cmds = append(cmds, m.resize(msg.Width, msg.Height))
	}

	// BUG(tqbf): clean this up

	var rm tea.Model

	if keyb {
		if m.state != screenHistory {
			m.bottom, cmd = m.bottom.Update(msg)
			cmds = append(cmds, cmd)
			rm, cmd = m.log.Update(msg)
			m.log = rm.(*Viewport)
			cmds = append(cmds, cmd)
		} else {
			rm, cmd = m.history.Update(msg)
			m.history = rm.(*DatabaseView)
			cmds = append(cmds, cmd)
		}
	} else {
		m.bottom, cmd = m.bottom.Update(msg)
		cmds = append(cmds, cmd)
		rm, cmd = m.log.Update(msg)
		m.log = rm.(*Viewport)
		cmds = append(cmds, cmd)
		rm, cmd = m.history.Update(msg)
		m.history = rm.(*DatabaseView)
		cmds = append(cmds, cmd)
	}

	m.status, cmd = m.status.Update(msg)
	cmds = append(cmds, cmd)

	m.top, cmd = m.top.Update(msg)
	cmds = append(cmds, cmd)

	m.controllers, cmd = m.controllers.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m rootWindow) View() string {
	if m.h == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.status.View(), m.top.View(), m.bottom.View())
}

func filterKey(msg tea.Msg, keys ...string) bool {
	if km, ok := msg.(tea.KeyMsg); ok {
		ks := km.String()
		for _, k := range keys {
			if ks == k {
				return true
			}
		}
	}

	return false
}
