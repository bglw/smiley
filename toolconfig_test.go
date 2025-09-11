package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func processCommandString(cmd string, args map[string]interface{}) string {
	cmdStr := cmd

	bracketRegex := regexp.MustCompile(`\[([^\[\]]+)\]`)
	for {
		matches := bracketRegex.FindStringSubmatch(cmdStr)
		if matches == nil {
			break
		}

		bracketContent := matches[1]
		if hasAllParameters(bracketContent, args) {
			cmdStr = strings.Replace(cmdStr, matches[0], bracketContent, 1)
		} else {
			cmdStr = strings.Replace(cmdStr, matches[0], "", 1)
		}
	}

	for paramName, value := range args {
		if value != nil {
			placeholder := "{" + paramName + "}"
			cmdStr = strings.Replace(cmdStr, placeholder, fmt.Sprintf("%v", value), -1)
		}
	}

	return strings.Join(strings.Fields(cmdStr), " ")
}

func TestBracketSyntax(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		params   map[string]ToolParameter
		args     map[string]interface{}
		expected string
	}{
		{
			name: "echo with flag included",
			cmd:  "echo [-v {verbose}] {message}",
			params: map[string]ToolParameter{
				"verbose": {Type: "string", Required: false},
				"message": {Type: "string", Required: true},
			},
			args: map[string]interface{}{
				"verbose": "true",
				"message": "hello",
			},
			expected: "echo -v true hello",
		},
		{
			name: "echo with flag excluded",
			cmd:  "echo [-v {verbose}] {message}",
			params: map[string]ToolParameter{
				"verbose": {Type: "string", Required: false},
				"message": {Type: "string", Required: true},
			},
			args: map[string]interface{}{
				"message": "hello",
			},
			expected: "echo hello",
		},
		{
			name: "weird syntax with plus",
			cmd:  "echo [+trace {level}] {message}",
			params: map[string]ToolParameter{
				"level":   {Type: "number", Required: false},
				"message": {Type: "string", Required: true},
			},
			args: map[string]interface{}{
				"level":   3,
				"message": "test",
			},
			expected: "echo +trace 3 test",
		},
		{
			name: "multiple brackets",
			cmd:  "echo [-v] [--output {format}] {message}",
			params: map[string]ToolParameter{
				"format":  {Type: "string", Required: false},
				"message": {Type: "string", Required: true},
			},
			args: map[string]interface{}{
				"format":  "json",
				"message": "test",
			},
			expected: "echo --output json test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processCommandString(tt.cmd, tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}
