package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type slashFunc func(args []string) (string, error)

var slashCommands = map[string]slashFunc{
	"/help": slashHelp,
}

type SlashCommandController struct {
}

func (t *SlashCommandController) Update(msg tea.Msg) (Controller, tea.Cmd) {
	switch msg := msg.(type) {
	case msgSlashCommand:
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

func slashHelp(args []string) (string, error) {
	return "This is a test of the emergency broadcast system\n", nil
}
