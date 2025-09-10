package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		params      map[string]ToolParameter
		args        map[string]interface{}
		expectError bool
		description string
	}{
		{
			name: "simple echo command",
			cmd:  "echo {message}",
			params: map[string]ToolParameter{
				"message": {Type: "string", Required: true},
			},
			args: map[string]interface{}{
				"message": "hello world",
			},
			expectError: false,
			description: "should substitute message parameter",
		},
		{
			name: "missing required parameter",
			cmd:  "echo {message}",
			params: map[string]ToolParameter{
				"message": {Type: "string", Required: true},
			},
			args:        map[string]interface{}{},
			expectError: true,
			description: "should fail when required parameter missing",
		},
		{
			name: "optional parameter missing",
			cmd:  "echo hello {message}",
			params: map[string]ToolParameter{
				"message": {Type: "string", Required: false},
			},
			args:        map[string]interface{}{},
			expectError: false,
			description: "should succeed when optional parameter missing and remove placeholder",
		},
		{
			name: "multiple parameters",
			cmd:  "echo {greeting} {name}",
			params: map[string]ToolParameter{
				"greeting": {Type: "string", Required: true},
				"name":     {Type: "string", Required: true},
			},
			args: map[string]interface{}{
				"greeting": "hello",
				"name":     "alice",
			},
			expectError: false,
			description: "should substitute multiple parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := generateCommand(tt.cmd, tt.params)

			argsJSON, err := json.Marshal(tt.args)
			assert.NoError(t, err)

			result, err := fn(context.Background(), argsJSON)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, result, "should return command output")
			}
		})
	}
}

func TestParameterSubstitution(t *testing.T) {
	cmd := "echo hello {name} world {age}"
	params := map[string]ToolParameter{
		"name": {Type: "string", Required: true},
		"age":  {Type: "number", Required: false},
	}

	t.Run("both parameters provided", func(t *testing.T) {
		fn := generateCommand(cmd, params)
		argsJSON, _ := json.Marshal(map[string]interface{}{
			"name": "alice",
			"age":  25,
		})
		result, err := fn(context.Background(), argsJSON)
		assert.NoError(t, err)
		assert.Contains(t, result, "hello alice world 25")
	})

	t.Run("optional parameter missing", func(t *testing.T) {
		fn := generateCommand(cmd, params)
		argsJSON, _ := json.Marshal(map[string]interface{}{
			"name": "bob",
		})
		result, err := fn(context.Background(), argsJSON)
		assert.NoError(t, err)
		assert.Contains(t, result, "hello bob world")
		assert.NotContains(t, result, "{age}")
	})
}

func TestLoadToolsIntegration(t *testing.T) {
	config, err := LoadToolConfig("test_tools.toml")
	assert.NoError(t, err)
	assert.Len(t, config.Tools, 1)

	tool := config.Tools[0]
	assert.Equal(t, "test_echo", tool.Name)
	assert.Equal(t, "echo Query: {query} Limit: {limit}", tool.Command)
	assert.Len(t, tool.Parameters, 2)

	fn := generateCommand(tool.Command, tool.Parameters)

	t.Run("with both parameters", func(t *testing.T) {
		argsJSON, _ := json.Marshal(map[string]interface{}{
			"query": "test search",
			"limit": 10,
		})
		result, err := fn(context.Background(), argsJSON)
		assert.NoError(t, err)
		assert.Contains(t, result, "Query: test search Limit: 10")
	})

	t.Run("with only required parameter", func(t *testing.T) {
		argsJSON, _ := json.Marshal(map[string]interface{}{
			"query": "another test",
		})
		result, err := fn(context.Background(), argsJSON)
		assert.NoError(t, err)
		assert.Contains(t, result, "Query: another test")
		assert.Contains(t, result, "Limit:")
		assert.NotContains(t, result, "{limit}")
	})
}
