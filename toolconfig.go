package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
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
