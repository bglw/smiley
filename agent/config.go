package agent

import (
	"fmt"
	"os"
	"path/filepath"
)

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
