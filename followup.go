package main

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type FollowupOption struct {
	Key         string
	Description string
}

type msgShowFollowupModal []FollowupOption
type msgFollowupSelected string

func (mf msgShowFollowupModal) hasFollowups() bool {
	return len([]FollowupOption(mf)) > 0
}

func parseFollowups(text string) []FollowupOption {
	startIdx := strings.Index(text, "<FOLLOWUP>")
	endIdx := strings.Index(text, "</FOLLOWUP>")

	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		return nil
	}

	content := text[startIdx+10 : endIdx]

	re := regexp.MustCompile(`\(F[0-9]+\)`)
	matches := re.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		return nil
	}

	options := []FollowupOption{}

	for i, match := range matches {
		key := content[match[0]+1 : match[1]-1]

		var description string
		if i < len(matches)-1 {
			description = content[match[1]:matches[i+1][0]]
		} else {
			description = content[match[1]:]
		}

		description = strings.TrimSpace(description)

		options = append(options, FollowupOption{
			Key:         key,
			Description: description,
		})

		if len(options) >= 9 {
			break
		}
	}

	return options
}

type FollowupModal struct {
	options       []FollowupOption
	selectedIndex int
	width         int
	height        int
	viewport      viewport.Model
	lineRanges    []struct{ start, end int }
}

func NewFollowupModal(options []FollowupOption) *FollowupModal {
	vp := viewport.New(46, 20)
	return &FollowupModal{
		options:       options,
		selectedIndex: 0,
		width:         50,
		height:        25,
		viewport:      vp,
	}
}

func (m *FollowupModal) Init() tea.Cmd {
	return nil
}

func (m *FollowupModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, func() tea.Msg {
				return msgFollowupSelected("")
			}

		case tea.KeyUp:
			m.selectedIndex = (m.selectedIndex - 1 + len(m.options)) % len(m.options)
			m.autoScroll()
			return m, nil

		case tea.KeyDown:
			m.selectedIndex = (m.selectedIndex + 1) % len(m.options)
			m.autoScroll()
			return m, nil

		case tea.KeyPgUp:
			m.viewport.PageUp()
			return m, nil

		case tea.KeyPgDown:
			m.viewport.PageDown()
			return m, nil

		case tea.KeyEnter:
			selected := m.options[m.selectedIndex].Key
			return m, func() tea.Msg {
				return msgFollowupSelected(selected)
			}
		}
	}

	return m, nil
}

func (m *FollowupModal) autoScroll() {
	if len(m.lineRanges) <= m.selectedIndex {
		return
	}

	startLine := m.lineRanges[m.selectedIndex].start
	endLine := m.lineRanges[m.selectedIndex].end

	targetEnd := endLine
	if m.selectedIndex+1 < len(m.lineRanges) {
		targetEnd = m.lineRanges[m.selectedIndex+1].start
	}

	contentHeight := targetEnd - startLine + 1
	if contentHeight > m.viewport.Height {
		contentHeight = endLine - startLine + 1
	}

	if contentHeight <= m.viewport.Height {
		if startLine < m.viewport.YOffset {
			m.viewport.YOffset = startLine
		} else if targetEnd >= m.viewport.YOffset+m.viewport.Height {
			m.viewport.YOffset = targetEnd - m.viewport.Height + 1
		}
	} else {
		m.viewport.YOffset = startLine
	}
}

func (m *FollowupModal) View() string {
	var content strings.Builder
	currentLine := 0
	m.lineRanges = make([]struct{ start, end int }, len(m.options))

	for i, opt := range m.options {
		startLine := currentLine
		prefix := "   "
		if i == m.selectedIndex {
			prefix = " > "
		}

		wrapped := wordwrap.String(opt.Key+": "+opt.Description, m.width-8)
		lines := strings.Split(wrapped, "\n")

		for j, line := range lines {
			if j == 0 {
				content.WriteString(prefix + line + "\n")
			} else {
				content.WriteString("     " + line + "\n")
			}
			currentLine++
		}

		m.lineRanges[i].start = startLine
		m.lineRanges[i].end = currentLine - 1

		if i < len(m.options)-1 {
			content.WriteString("\n")
			currentLine++
		}
	}

	m.viewport.SetContent(content.String())

	header := lipgloss.NewStyle().
		Width(m.width - 4).
		Foreground(lipgloss.Color("230")).
		Render("Follow-up Options")

	footer := lipgloss.NewStyle().
		Width(m.width - 4).
		Foreground(lipgloss.Color("240")).
		Render("↑/↓: select  PgUp/PgDn: scroll  Enter: submit  Esc: dismiss")

	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		m.viewport.View(),
		"",
		footer,
	)

	modalStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("230"))

	return modalStyle.Render(modalContent)
}
