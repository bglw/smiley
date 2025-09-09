package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/superfly/contextwindow"
)

var (
	border_4 = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	border_hr = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true).
			BorderForeground(lipgloss.Color("228"))
)

// subwindow height
type WindowSize struct {
	Height int
	Width  int
	Loc    string
}

type Controller interface {
	Update(msg tea.Msg) (Controller, tea.Cmd)
}

type Controllers []Controller

func (cs Controllers) Update(msg tea.Msg) (Controller, tea.Cmd) {
	cmds := []tea.Cmd{}

	for i := range cs {
		c, cmd := cs[i].Update(msg)
		cs[i] = c
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	for _, cmd := range cmds {
		slog.Info("controller cmd", "cmd", cmd)
	}

	return cs, tea.Sequence(cmds...)
}

type rootWindow struct {
	p                   *tea.Program
	status, top, bottom tea.Model
	controllers         Controller
	w, h, th, bh        int

	db *sql.DB
}

func newRootWindow(content string) rootWindow {
	m := rootWindow{}

	m.status = NewStatus()

	box := NewBox("top", "top-inner", false, false, true, false)
	box.Inner = NewViewport("top-inner", content)
	m.top = box

	input := NewTextarea("bottom")
	m.bottom = input

	return m
}

func init() {
	f, err := os.OpenFile("/tmp/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}

func round(tot, pct int) int {
	if tot <= 0 || pct <= 0 {
		return 0
	}

	rows := float64(tot) * float64(pct) / 100.0
	return int(math.Round(rows))
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

func (m rootWindow) Init() tea.Cmd {
	return nil
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
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		slog.Info("keypress", "key", msg)

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		cmds = append(cmds, m.resize(msg.Width, msg.Height))
	}

	m.status, cmd = m.status.Update(msg)
	cmds = append(cmds, cmd)
	m.top, cmd = m.top.Update(msg)
	cmds = append(cmds, cmd)
	m.bottom, cmd = m.bottom.Update(msg)
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

func eprintf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	if len(format) == 0 || format[len(format)-1] != '\n' {
		fmt.Fprintln(os.Stderr)
	}
	os.Exit(1)
}

func mustGetenv(v string) string {
	val := os.Getenv(v)
	if val == "" {
		eprintf("%s not set", v)
	}

	return val
}

func main() {
	eliot, _ := os.ReadFile("hollow.txt")
	_ = eliot
	m := newRootWindow("")

	cfgdir, err := ensureCtxAgentDir()
	if err != nil {
		eprintf("Find ~/.ctxagent: %v", err)
	}

	path := filepath.Join(cfgdir, "contextwindow.db")
	db, err := contextwindow.NewContextDB(path)
	if err != nil {
		eprintf("Open %s: %v", path, err)
	}

	m.db = db

	model, err := contextwindow.NewOpenAIResponsesModel(contextwindow.ResponsesModelGPT5Mini)
	if err != nil {
		eprintf("Connect to LLM: %v", err)
	}

	cw, err := contextwindow.NewContextWindowWithThreading(db, model, "", true)
	if err != nil {
		eprintf("Create context window: %v", err)
	}

	llm := LLMController{
		model:   model,
		context: cw,
		db:      db,
	}

	controllers := Controllers{}
	controllers = append(controllers, &TextAreaInput{})
	controllers = append(controllers, &llm)
	m.controllers = controllers

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.p = p
	llm.context.AddMiddleware(&tuiMiddleware{
		program: p,
	})

	if _, err := p.Run(); err != nil {
		eprintf("main program error: %v", err)
	}

}
