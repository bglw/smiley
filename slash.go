package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/superfly/contextwindow"
)

type SlashCommandController struct {
	cw *contextwindow.ContextWindow
}

func (t *SlashCommandController) Update(msg tea.Msg) (Controller, tea.Cmd) {
	switch msg := msg.(type) {
	case msgSlashCommand:
		slashCommands := map[string]func([]string) (string, error){
			"/help":    t.slashHelp,
			"/dump":    t.slashDump,
			"/summary": t.slashSummary,
		}

		if fn, ok := slashCommands[strings.ToLower(msg[0])]; ok {
			res, err := fn([]string(msg))
			if err != nil {
				return t, viewLog("Error: "+err.Error()+"\n", styleErrorText)
			}

			return t, viewLog(res+"\n", styleSlashResult)
		}
	}

	return t, nil
}

func (t *SlashCommandController) slashHelp(args []string) (string, error) {
	return "This is a test of the emergency broadcast system\n", nil
}

func (t *SlashCommandController) slashDump(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("/dump <filename.md>")
	}

	path := args[1]

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("/dump: file %s already exists", path)
	}
	if err != nil {
		return "", fmt.Errorf("/dump: write to %s: %w", path, err)
	}
	defer f.Close()

	records, err := t.cw.LiveRecords()
	if err != nil {
		return "", fmt.Errorf("/dump: failed to read context: %w", err)
	}

	// it really doesn't matter too much to me what this looks like,
	// because i'm always going to be editing these documents anyways.

	fmt.Fprintf(f, "# Conversation Export\n\n")

	for _, record := range records {
		switch record.Source {
		case contextwindow.Prompt:
			fmt.Fprintf(f, "## User\n\n%s\n\n", record.Content)
		case contextwindow.ModelResp:
			fmt.Fprintf(f, "## Assistant\n\n%s\n\n", record.Content)
		case contextwindow.ToolCall:
			fmt.Fprintf(f, "### Tool Call\n\n```\n%s\n```\n\n", record.Content)
		case contextwindow.ToolOutput:
			fmt.Fprintf(f, "### Tool Output\n\n```\n%s\n```\n\n", record.Content)
		}
	}

	return fmt.Sprintf("Exported %d records to %s", len(records), path), nil
}

func (t *SlashCommandController) slashSummary(args []string) (string, error) {
	records, err := t.cw.LiveRecords()
	if err != nil {
		return "", err
	}

	trunc := func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return fmt.Sprintf("%s... + %d bytes", s[:n], len(s)-n)
	}

	buf := &strings.Builder{}

	for i, record := range records {
		switch record.Source {
		case contextwindow.Prompt:
			fmt.Fprintf(buf, "%d (user) %s...\n", i, trunc(record.Content, 40))
		case contextwindow.ModelResp:
			fmt.Fprintf(buf, "%d (model) %s...\n", i, trunc(record.Content, 40))
		case contextwindow.ToolCall:
			fmt.Fprintf(buf, "%d (toolcall) %s...\n", i, trunc(record.Content, 40))
		case contextwindow.ToolOutput:
			fmt.Fprintf(buf, "%d (tool) %s...\n", i, trunc(record.Content, 40))
		}
	}

	return buf.String() + "\n", nil
}
