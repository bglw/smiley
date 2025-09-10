package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/superfly/contextwindow"
)

type ToolParameter struct {
	Type        string `toml:"type"`
	Description string `toml:"description"`
	Required    bool   `toml:"required"`
}

type ToolConfig struct {
	Name        string                   `toml:"name"`
	Description string                   `toml:"description"`
	Command     string                   `toml:"command"`
	Parameters  map[string]ToolParameter `toml:"parameters"`
}

type ToolsConfig struct {
	Tools []ToolConfig `toml:"tool"`
}

func LoadToolConfig(configPath string) (*ToolsConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ToolsConfig{}, nil
	}

	var config ToolsConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse tool config %s: %w", configPath, err)
	}

	for i := range config.Tools {
		tool := &config.Tools[i]
		if tool.Parameters == nil {
			tool.Parameters = make(map[string]ToolParameter)
		}

		for paramName, param := range tool.Parameters {
			if param.Type == "" {
				param.Type = "string"
				tool.Parameters[paramName] = param
			}
		}
	}

	return &config, nil
}

type simpleToolFunction func(context.Context, json.RawMessage) (string, error)

func generateCommand(
	cmd string,
	params map[string]ToolParameter) simpleToolFunction {

	return func(ctx context.Context, args json.RawMessage) (string, error) {
		// assume we have parameters:
		//   query (string)
		//   limit (number)
		//
		// then our JSON message will look like:
		//
		// var params struct {
		//   Query string `json:"query"`
		//   Limit int `json:"limit"`
		// }
		//
		// but we're generating this dynamically so we'll instead
		// need to parse this out of a map[string]interface{}
		//
		// var params = map[string]interface{}
		// json.Unmarshall(args, &params)
		//
		// then, whatever "cmd" is, we'll need to os/exec it,
		// substituting in {query} and {limit} into the string
		// we execute.
		//
		// the command will produce a string (or an error);
		// we can return both directly.

		return "", nil /* placeholder */
	}
}

func LoadTools(cw *contextwindow.ContextWindow, cfg *ToolsConfig) error {
	for _, toolCfg := range cfg.Tools {
		tool := contextwindow.NewTool(
			toolCfg.Name,
			toolCfg.Description)
		for pk, pv := range toolCfg.Parameters {
			switch pv.Type {
			case "string":
				tool = tool.AddStringParameter(pk, pv.Description, pv.Required)
			case "number":
				tool = tool.AddNumberParameter(pk, pv.Description, pv.Required)
			default:
				return fmt.Errorf("load tools: unknown parameter type \"%s\" for %s",
					pv.Type, toolCfg.Name)
			}
		}

		cw.AddTool(tool,
			contextwindow.ToolRunnerFunc(
				generateCommand(toolCfg.Name, toolCfg.Parameters),
			))
	}

	return nil
}
