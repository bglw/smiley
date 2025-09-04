package main

import (
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// 	borderTL = "┌"
	// 	borderTR = "┐"
	// 	borderBL = "⎣"
	// 	borderBR = "⎦"
	// 	borderHR = "─"

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

type rootWindow struct {
	p *tea.Program

	top, bottom tea.Model

	vm viewport.Model

	w, h, th, bh int
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

func (m rootWindow) Init() tea.Cmd {
	return nil
}

func (m *rootWindow) resize(w, h int) tea.Cmd {
	if m.h == h && m.w == w {
		return nil
	}

	m.h = h
	m.w = w

	m.th = round(m.h, 80)
	m.bh = round(m.h, 20)

	slog.Info("main dims", "w", m.w, "h", m.h, "th", m.th, "bh", m.bh)

	return tea.Batch(
		func() tea.Msg {
			return WindowSize{
				Width:  w,
				Height: m.th,
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
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		cmds = append(cmds, m.resize(msg.Width, msg.Height))
	}

	m.top, cmd = m.top.Update(msg)
	cmds = append(cmds, cmd)
	m.bottom, cmd = m.bottom.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m rootWindow) View() string {
	if m.h == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Center, m.top.View(), m.bottom.View())
}

func newRootWindow() rootWindow {
	return rootWindow{}
}

func main() {
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")
	slog.Info("new run")

	m := newRootWindow()

	eliot, _ := os.ReadFile("hollow.txt")
	box := NewBox("top", "top-inner", false, false, true, false)
	box.Inner = NewViewport("top-inner", string(eliot))
	m.top = box

	static := NewStatic("bottom", strings.Repeat("hello\n", 10),
		lipgloss.NewStyle().
			Background(lipgloss.Color("#222")))
	m.bottom = static

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.p = p

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "main program error: %v\n", err)
		os.Exit(1)
	}

}
