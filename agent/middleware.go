package agent

import (
	"context"
	"fmt"
	"log/slog"
)

type agentMiddleware struct {
	agent *Agent
}

func (am *agentMiddleware) OnToolCall(ctx context.Context, name, args string) {
	if am.agent.OnEvent != nil {
		msg := fmt.Sprintf("%s(%s)", name, args)
		am.agent.OnEvent(ToolCallMsg{
			Name: name,
			Args: args,
			Msg:  msg,
		})
	}
}

func (am *agentMiddleware) OnToolResult(ctx context.Context, name, result string, err error) {
	if am.agent.OnEvent == nil {
		return
	}

	var msg string
	if err != nil {
		msg = fmt.Sprintf("%s: error: %s", name, err.Error())
	} else {
		msg = fmt.Sprintf("%s: (%d bytes)", name, len(result))
	}

	slog.Debug("llm", "name", name, "result", result)

	am.agent.OnEvent(ToolCallMsg{
		Name:     name,
		Complete: true,
		Size:     len(result),
		Err:      err,
		Msg:      msg,
	})
}
