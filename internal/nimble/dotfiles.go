// Copright (c) 2024 Neomantra BV

package nimble

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	NIMBLE_DIR_ENV          = "NIMBLE_DIR"
	NIMBLE_DIR_DEFAULT_NAME = ".nimble" // in home directory
)

func GetNimbleDir() (string, error) {
	nimbleDir := os.Getenv(NIMBLE_DIR_ENV)
	if nimbleDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		nimbleDir = filepath.Join(homeDir, NIMBLE_DIR_DEFAULT_NAME)
	}

	// Try to create the directory (if needed), and return any error
	err := os.MkdirAll(nimbleDir, 0700)
	return nimbleDir, err
}
