package main

import (
	_ "embed"

	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
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

//go:embed systemprompt.default.md
var defaultSystemPrompt string

var optLogTools = flag.Bool("log-tools", false, "Record tool invocations/responses in the transcript")

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

func init() {
	var (
		f   io.Writer = io.Discard
		err error
	)

	if os.Getenv("DEBUG") != "" {
		f, err = os.OpenFile("/tmp/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
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
	var (
		systemMd    = flag.String("system", "", "Path to system.md")
		toolConfig  = flag.String("tools", "", "Path to tools.toml")
		contextDb   = flag.String("db", "", "Path to contextwindow.db")
		contextName = flag.String("name", "", "Optional conversation name")
		maxTokens   = flag.Int("maxtokens", 60_000, "Maximum tokens before compacting")
	)

	flag.Usage = func() {
		fmt.Println("smiley [options]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.Parse()

	prompt := strings.TrimSpace(strings.Join(flag.Args(), " "))

	cfgdir, err := ensureCtxAgentDir()
	if err != nil {
		eprintf("Find ~/.ctxagent: %v", err)
	}

	toolConfigPath := filepath.Join(cfgdir, "tools.toml")

	if *toolConfig != "" {
		toolConfigPath = *toolConfig
	}
	tools, err := LoadToolConfig(toolConfigPath)
	if err != nil {
		if *toolConfig != "" {
			eprintf("Load tool config: %v", err)
		}
		tools = nil
	}

	systemPrompt := defaultSystemPrompt
	systemPromptPath := filepath.Join(cfgdir, "system.md")
	if *systemMd != "" {
		systemPromptPath = *systemMd
	}
	pbuf, err := os.ReadFile(systemPromptPath)
	if err != nil {
		if *systemMd != "" {
			eprintf("Loading %s: %v", systemPromptPath, err)
		}
	} else {
		systemPrompt = string(pbuf)
	}

	path := filepath.Join(cfgdir, "contextwindow.db")
	if *contextDb != "" {
		path = *contextDb
	}
	db, err := contextwindow.NewContextDB(path)
	if err != nil {
		eprintf("Open %s: %v", path, err)
	}

	model, err := contextwindow.NewOpenAIResponsesModel(contextwindow.ResponsesModelGPT5Mini)
	if err != nil {
		eprintf("Connect to LLM: %v", err)
	}

	cw, err := contextwindow.NewContextWindowWithThreading(db, model, *contextName, true)
	if err != nil {
		eprintf("Create context window: %v", err)
	}

	cw.SetSystemPrompt(systemPrompt)
	cw.SetMaxTokens(*maxTokens)

	if tools != nil {
		if err = LoadTools(cw, tools); err != nil {
			eprintf("Loading tool definitions from %s: %v", toolConfigPath, err)
		}
	}

	m := newRootWindow("", cw, prompt, *contextName)
	m.db = db

	llm := LLMController{
		model:   model,
		context: cw,
		db:      db,
	}

	controllers := Controllers{}
	controllers = append(controllers, &TextAreaInput{})
	controllers = append(controllers, &SlashCommandController{})
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
