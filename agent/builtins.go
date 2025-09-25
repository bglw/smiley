package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/superfly/contextwindow"
)

type BuiltinTool interface {
	ToolDescription() string
	Init(*contextwindow.ContextWindow) error
	Run(context.Context, json.RawMessage) (string, error)
}

type Todo struct {
	lock  sync.Mutex
	todos []string
}

func (t *Todo) ToolDescription() string {
	return `
name = "todo"
description = """
Manage a list of todo steps; good for multistep processes, internal planning, scratch memory.

You MUST call this tool with an "action".

tool({"action":"add", "entry":"a cool new todo list entry")

"add" adds elements to the list.
"list" gives a numbered list of the current todo elements.
"delete" deletes a numbered entry from the list.
"""

[parameters]
action = { type = "string", description = "add or delete or list", required = true }
entry = { type = "string", description = "A blob of text to add to the todo list", required = false }
number = { type = "number", description = "The number of an entry to delete", required = false }
`
}

func (t *Todo) Init(cw *contextwindow.ContextWindow) error {
	return nil
}

func (t *Todo) Run(ctx context.Context, rawargs json.RawMessage) (string, error) {
	slog.Info("todo run", "args", string(rawargs))

	t.lock.Lock()
	defer t.lock.Unlock()

	out := &bytes.Buffer{}

	args := map[string]any{}
	if err := json.Unmarshal(rawargs, &args); err != nil {
		return "", err
	}

	action, ok := mapGet[string](args, "action")
	if !ok {
		return "", fmt.Errorf("no valid action")
	}

	switch action {
	case "add":
		entry, ok := mapGet[string](args, "entry")
		if !ok {
			return "", fmt.Errorf("no entry")
		}
		t.todos = append(t.todos, entry)

	case "list":
		fmt.Fprintf(out, "<todo_entries>\n")

		for i, e := range t.todos {
			fmt.Fprintf(out, "%d. %s\n", i+1, e)
		}

		fmt.Fprintf(out, "</todo_entries>\n")

	case "delete":
		num, ok := mapGet[int](args, "number")
		if !ok {
			return "", fmt.Errorf("no number")
		}

		todo := t.todos
		newdos := []string{}
		for i, e := range todo {
			if i+1 != num {
				newdos = append(newdos, e)
			} else {
				fmt.Fprintf(out, "deleted %d\n", num)
			}
		}
		t.todos = newdos

	default:
		return "", fmt.Errorf("invalid action %s", action)
	}

	return out.String(), nil
}

type Review struct {
	cw *contextwindow.ContextWindow
}

type Lobotomize struct {
	cw *contextwindow.ContextWindow
}

func (r *Review) ToolDescription() string {
	return `
name = "review"
description = """
Dump a zero-indexed numbered list of elements in the current
conversation, truncated. If truncated, the element is followed
by a count of the remaining bytes. 
"""
`
}

func (r *Review) Init(cw *contextwindow.ContextWindow) error {
	r.cw = cw
	return nil
}

func (r *Review) Run(ctx context.Context, rawargs json.RawMessage) (string, error) {
	records, err := r.cw.LiveRecords()
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

func (l *Lobotomize) ToolDescription() string {
	return `
name = "lobotomize"
description = """
Remove elements from the "live" context window; they will not in the
future be sent back to the LLM model.

Elements are numbered starting at 0. Use the "review" tool to identify
elements you might remove.
"""

[parameters]
start = { type = "number", description = "first element to kill", required = true }
end = { type = "number", description = "last element to kill", required = true }
`
}

func (l *Lobotomize) Init(cw *contextwindow.ContextWindow) error {
	l.cw = cw
	return nil
}

func (l *Lobotomize) Run(ctx context.Context, rawargs json.RawMessage) (string, error) {
	args := map[string]any{}
	if err := json.Unmarshal(rawargs, &args); err != nil {
		return "", err
	}

	_, ok := mapGet[float64](args, "start")
	if !ok {
		return "", fmt.Errorf("must provide 'start' parameter")
	}

	_, ok = mapGet[float64](args, "end")
	if !ok {
		return "", fmt.Errorf("must provide 'end' parameter")
	}

	// TODO: API has changed, need to find new way to set record live state
	return "Lobotomize functionality temporarily disabled due to API changes\n", nil
}

func mapGet[T any](m map[string]any, key string) (T, bool) {
	var zero T
	v, ok := m[key]
	if !ok {
		return zero, false
	}
	x, ok := v.(T)
	return x, ok
}
