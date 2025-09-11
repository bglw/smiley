package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
		var parsedArgs map[string]interface{}
		if err := json.Unmarshal(args, &parsedArgs); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}

		cmdStr := cmd

		bracketRegex := regexp.MustCompile(`\[([^\[\]]+)\]`)
		for {
			matches := bracketRegex.FindStringSubmatch(cmdStr)
			if matches == nil {
				break
			}

			bracketContent := matches[1]
			if hasAllParameters(bracketContent, parsedArgs) {
				cmdStr = strings.Replace(cmdStr, matches[0], bracketContent, 1)
			} else {
				cmdStr = strings.Replace(cmdStr, matches[0], "", 1)
			}
		}

		paramRegex := regexp.MustCompile(`\{(\w+)\}`)
		for paramName := range params {
			value, exists := parsedArgs[paramName]
			if exists && value != nil {
				cmdStr = paramRegex.ReplaceAllStringFunc(cmdStr, func(match string) string {
					if match == fmt.Sprintf("{%s}", paramName) {
						return fmt.Sprintf("%v", value)
					}
					return match
				})
			} else if params[paramName].Required {
				return "", fmt.Errorf("required parameter %s not provided", paramName)
			}
		}

		cmdStr = strings.Join(strings.Fields(cmdStr), " ")

		cmdParts := strings.Fields(cmdStr)
		if len(cmdParts) == 0 {
			return "", fmt.Errorf("empty command")
		}

		execCmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
		output, err := execCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
		}

		return string(output), nil
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
				generateCommand(toolCfg.Command, toolCfg.Parameters),
			))
	}

	return nil
}

func hasAllParameters(fragment string, args map[string]interface{}) bool {
	paramRegex := regexp.MustCompile(`\{(\w+)\}`)
	matches := paramRegex.FindAllStringSubmatch(fragment, -1)

	if len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		paramName := match[1]
		if _, exists := args[paramName]; !exists {
			return false
		}
	}
	return true
}
