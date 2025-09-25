package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

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
	InfoCommand string                   `toml:"info_command"`
	Parameters  map[string]ToolParameter `toml:"parameters"`
	Builtin     bool                     `toml:"builtin"`
}

type ToolsConfig struct {
	Tools []ToolConfig `toml:"tool"`
}

func toolFromInfoCommand(cmdline string) (ret ToolConfig, err error) {
	parts := strings.Fields(strings.Join(strings.Fields(cmdline), " "))
	if len(parts) == 0 {
		return ret, fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return ret, fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	if err = toml.Unmarshal(output, &ret); err != nil {
		return ret, fmt.Errorf("parse failed: %w\nOutput: %s", err, string(output))
	}

	return ret, nil
}

func LoadToolConfig(configPath string) (*ToolsConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ToolsConfig{}, nil
	}

	var config ToolsConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("parse tool config %s: %w", configPath, err)
	}

	for i := range config.Tools {
		tool := &config.Tools[i]

		if tool.InfoCommand != "" {
			tool, err := toolFromInfoCommand(tool.InfoCommand)
			if err != nil {
				return nil, fmt.Errorf("parse tool info: %w", err)
			}
			config.Tools[i] = tool
		} else if tool.Parameters == nil {
			tool.Parameters = make(map[string]ToolParameter)
		} else {
			for paramName, param := range tool.Parameters {
				if param.Type == "" {
					param.Type = "string"
					tool.Parameters[paramName] = param
				}
			}
		}

		if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(tool.Name) {
			return nil, fmt.Errorf("tool name '%s' contains invalid characters, must match [a-zA-Z0-9_-]", tool.Name)
		}

		if tool.Name == "" {
			return nil, fmt.Errorf("tool name cannot be empty")
		}

		if tool.Description == "" && !tool.Builtin {
			return nil, fmt.Errorf("tool '%s' must have a description", tool.Name)
		}

		if tool.Command == "" && !tool.Builtin {
			return nil, fmt.Errorf("tool '%s' must have a command", tool.Name)
		}
	}

	return &config, nil
}

var (
	bracketRegex = regexp.MustCompile(`\[([^\[\]]+)\]`)
	paramRegex   = regexp.MustCompile(`\{(\w+)\}`)
)

type simpleToolFunction func(context.Context, json.RawMessage) (string, error)

func generateCommand(
	cmd string,
	params map[string]ToolParameter) simpleToolFunction {

	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var parsedArgs map[string]interface{}
		if err := json.Unmarshal(args, &parsedArgs); err != nil {
			return "", fmt.Errorf("execute tool \"%s\": failed to parse arguments: %w", cmd, err)
		}

		cmdStr := cmd

		// for each [optional] [--flag], extract the flag, see
		// if we have the needed {params} to expand it, otherwise
		// zap the whole [--optional flag].
		//
		// ie, for "tcpdump tcp [and port {port}]", when {port}
		// is optional.

		for {
			matches := bracketRegex.FindStringSubmatch(cmdStr)
			if matches == nil {
				break
			}

			optionalFlag := matches[1]
			if hasAllParameters(optionalFlag, parsedArgs) {
				cmdStr = strings.Replace(cmdStr, matches[0], optionalFlag, 1)
			} else {
				cmdStr = strings.Replace(cmdStr, matches[0], "", 1)
			}
		}

		// now substitute in the parameters themselves

		for paramName := range params {
			value, exists := parsedArgs[paramName]
			if exists && value != nil {
				cmdStr = paramRegex.ReplaceAllStringFunc(
					cmdStr,
					func(match string) string {
						if match == fmt.Sprintf("{%s}", paramName) {
							return fmt.Sprintf("%v", value)
						}
						return match
					})
			} else if params[paramName].Required {
				return "", fmt.Errorf("execute tool \"%s\": required parameter %s not provided", cmd, paramName)
			}
		}

		// now run the command

		cmdStr = strings.Join(strings.Fields(cmdStr), " ")

		cmdParts := strings.Fields(cmdStr)
		if len(cmdParts) == 0 {
			return "", fmt.Errorf("execute tool \"%s\": empty command", cmd)
		}

		execCmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
		output, err := execCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
		}

		return string(output), nil
	}
}

// TODO(tqbf): this is repulsive but whatever
var (
	lockBuiltins sync.Mutex
	builtins     = map[string]BuiltinTool{}
)

func LoadBuiltin(name string, t BuiltinTool) {
	lockBuiltins.Lock()
	defer lockBuiltins.Unlock()

	builtins[name] = t
}

func loadBuiltin(cw *contextwindow.ContextWindow, cfg ToolConfig) error {
	lockBuiltins.Lock()
	defer lockBuiltins.Unlock()

	bt, ok := builtins[cfg.Name]
	if !ok {
		return fmt.Errorf("builtin %s: not found", cfg.Name)
	}

	desc := bt.ToolDescription()

	var builtinCfg ToolConfig
	if err := toml.Unmarshal([]byte(desc), &builtinCfg); err != nil {
		return fmt.Errorf("parse builtin tool description: %w", err)
	}

	tool := contextwindow.NewTool(builtinCfg.Name, builtinCfg.Description)
	for pk, pv := range builtinCfg.Parameters {
		switch pv.Type {
		case "string":
			tool = tool.AddStringParameter(pk, pv.Description, pv.Required)
		case "number":
			tool = tool.AddNumberParameter(pk, pv.Description, pv.Required)
		default:
			return fmt.Errorf("load builtin %s: unknown parameter type \"%s\"",
				builtinCfg.Name, pv.Type)
		}
	}

	if err := bt.Init(cw); err != nil {
		return fmt.Errorf("load builtin: init: %w", err)
	}

	cw.AddTool(tool, bt)
	return nil
}

func LoadTools(cw *contextwindow.ContextWindow, cfg *ToolsConfig) error {
	for _, toolCfg := range cfg.Tools {
		if toolCfg.Builtin {
			if err := loadBuiltin(cw, toolCfg); err != nil {
				return fmt.Errorf("load tools: builtin: %w", err)
			}
			continue
		}

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

// for a given optional [--flag {foo}{bar}], check to see if we have
// both {foo} and {bar}, so we can either substitute it in or zap the
// whole flag.
func hasAllParameters(fragment string, args map[string]interface{}) bool {
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
