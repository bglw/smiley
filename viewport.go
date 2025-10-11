package main

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type Viewport struct {
	lock    sync.Mutex
	Content []string
	live    *strings.Builder
	ID      string
	vm      viewport.Model
	focused bool
}

type msgViewportLog struct {
	Msg   string
	Style lipgloss.Style
}

type msgResetViewport []msgViewportLog

func NewViewport(id, content string) *Viewport {
	v := Viewport{
		ID:      id,
		vm:      viewport.New(0, 0),
		Content: []string{content},
		live:    &strings.Builder{},
	}
	v.vm.SetContent(content)
	v.vm.GotoTop()

	return &v
}

func (v *Viewport) Init() tea.Cmd {
	return nil
}

func (v *Viewport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case msgFocusChanged:
		v.focused = (msg.region == "top")

	case tea.KeyMsg:
		// pgup/pgdown/alt+up/alt+down always work (these are viewport-specific keys)
		if filterKey(msg, "pgup", "pgdown", "alt+up", "alt+down") {
			v.vm, cmd = v.vm.Update(msg)
		} else if v.focused {
			// Arrow keys and other keys only work when focused
			switch msg.Type {
			case tea.KeyUp, tea.KeyDown:
				v.vm, cmd = v.vm.Update(msg)
			case tea.KeyEnd:
				v.vm.GotoBottom()
			}
		}

	case msgResetViewport:
		v.resetViewport(msg)

	case msgViewportLog:
		v.Add(msg.Style.Render(msg.Msg))
		v.vm.GotoBottom()

	case WindowSize:
		if msg.Loc == v.ID {
			v.vm.Height = msg.Height
			v.vm.Width = msg.Width
			v.rewrap()
		}

	default:
		v.vm, cmd = v.vm.Update(msg)
	}

	return v, cmd
}

func (v *Viewport) View() string {
	if v.vm.Height == 0 || v.vm.Width == 0 {
		return ""
	}

	vs := v.vm.View()
	return vs
}

func (v *Viewport) rewrap() {
	if v.vm.Width == 0 {
		return
	}

	slog.Info("call rewrap", "w", v.vm.Width, "h", v.vm.Height, "lines", len(v.Content))

	content := make([]string, len(v.Content))
	for _, c := range v.Content {
		content = append(content, wordwrap.String(c, v.vm.Width-5))
	}
	v.Content = content
}

func (v *Viewport) resetViewport(lines []msgViewportLog) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.Content = []string{}
	v.live.Reset()

	// don't want to call SetContent in a loop
	for _, line := range lines {
		entry := line.Style.Render(wordwrap.String(line.Msg, v.vm.Width-5))
		v.Content = append(v.Content, entry)
		v.live.WriteString(entry + "\n")
	}

	v.vm.SetContent(v.live.String())
}

func (v *Viewport) Add(c string) {
	v.lock.Lock()
	defer v.lock.Unlock()
	entry := wordwrap.String(c, v.vm.Width-5)
	v.Content = append(v.Content, entry)
	v.live.WriteString(entry + "\n")
	v.vm.SetContent(v.live.String())
}
