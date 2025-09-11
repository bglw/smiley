package main

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

type tuiMiddleware struct {
	program *tea.Program
}

func (tm *tuiMiddleware) OnToolCall(ctx context.Context, name, args string) {
	msg := fmt.Sprintf("%s(%s)", name, args)
	tm.program.Send(msgToolCall{
		name: name,
		msg:  msg,
	})
}

func (tm *tuiMiddleware) OnToolResult(ctx context.Context, name, result string, err error) {
	var msg string
	if err != nil {
		msg = fmt.Sprintf("%s: error: %s", name, err.Error())
	} else {
		msg = fmt.Sprintf("%s: (%d bytes)", name, len(result))
	}

	slog.Debug("llm", "name", name, "result", result)

	tm.program.Send(msgToolCall{
		complete: true,
		size:     len(result),
		err:      err,
		msg:      msg,
		name:     name,
	})
}
