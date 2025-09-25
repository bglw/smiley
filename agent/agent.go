package agent

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/superfly/contextwindow"
)

type Agent struct {
	lock    sync.Mutex
	model   contextwindow.Model
	context *contextwindow.ContextWindow
	db      *sql.DB

	OnEvent func(Message)
}

func NewAgentForked(db *sql.DB, model contextwindow.Model, contextName, oldName string) (*Agent, error) {
	if contextName == "" {
		contextName = uuid.New().String()
	}

	err := contextwindow.CloneContext(db, oldName, contextName)
	if err != nil {
		return nil, err
	}

	return NewAgent(db, model, contextName)
}

func NewAgent(db *sql.DB, model contextwindow.Model, contextName string) (*Agent, error) {
	if contextName == "" {
		contextName = uuid.New().String()
	}

	cw, err := contextwindow.NewContextWindowWithThreading(db, model, contextName, true)
	if err != nil {
		return nil, fmt.Errorf("create context window: %w", err)
	}

	agent := &Agent{
		model:   model,
		context: cw,
		db:      db,
	}

	cw.AddMiddleware(&agentMiddleware{
		agent: agent,
	})

	return agent, nil
}

func (a *Agent) SendPrompt(prompt string) error {
	a.lock.Lock()
	a.context.AddPrompt(prompt)
	a.lock.Unlock()

	if a.OnEvent != nil {
		a.OnEvent(WorkingMsg{Working: true})
	}

	a.sendTokenUsage()

	go func() {
		response, err := a.callModel()
		if err != nil {
			if a.OnEvent != nil {
				a.OnEvent(ErrorMsg{
					Err: err,
					Msg: err.Error() + "\n",
				})
			}
		} else {
			if a.OnEvent != nil {
				a.OnEvent(ModelResponseMsg{Response: response})
			}
		}

		a.sendTokenUsage()

		if a.OnEvent != nil {
			a.OnEvent(WorkingMsg{Working: false})
		}
	}()

	return nil
}

func (a *Agent) callModel() (string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	response, err := a.context.CallModel(context.TODO())
	if err != nil {
		slog.Info("llm call error", "error", err)
		return "", err
	}

	return response, nil
}

func (a *Agent) SwitchContext(name string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	err := a.context.SwitchContext(name)
	if err != nil {
		slog.Error("switch context", "error", err)
		return fmt.Errorf("switch context: %w", err)
	}

	return nil
}

func (a *Agent) ListContexts() (ret []string, err error) {
	cs, err := a.context.ListContexts()
	if err != nil {
		return nil, err
	}

	for _, c := range cs {
		ret = append(ret, c.Name)
	}

	return ret, nil
}

func (a *Agent) LoadTools(configPath string) error {
	tools, err := LoadToolConfig(configPath)
	if err != nil {
		return fmt.Errorf("load tool config: %w", err)
	}

	if err := LoadTools(a.context, tools); err != nil {
		return fmt.Errorf("load tools: %w", err)
	}

	return nil
}

func (a *Agent) RegisterBuiltinTool(name string, tool BuiltinTool) {
	LoadBuiltin(name, tool)
}

func (a *Agent) GetContextWindow() *contextwindow.ContextWindow {
	return a.context
}

func (a *Agent) SetSystemPrompt(prompt string) {
	a.context.SetSystemPrompt(prompt)
}

func (a *Agent) SetMaxTokens(max int) {
	a.context.SetMaxTokens(max)
}

func (a *Agent) sendTokenUsage() {
	if a.OnEvent == nil {
		return
	}

	usage, err := a.context.TokenUsage()
	if err != nil {
		return
	}

	a.OnEvent(TokenUsageMsg{Usage: usage.Percent})
}

func EnsureCtxAgentDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}

	ctxAgentDir := filepath.Join(homeDir, ".ctxagent")
	if _, err := os.Stat(ctxAgentDir); os.IsNotExist(err) {
		err = os.MkdirAll(ctxAgentDir, 0755)
		if err != nil {
			return "", fmt.Errorf("create .ctxagent directory: %w", err)
		}
	}

	return ctxAgentDir, nil
}
