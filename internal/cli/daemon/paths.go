package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

// getPasoDir returns the path to the ~/.paso directory
func getPasoDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
	}
	return filepath.Join(home, ".paso"), nil
}

// getSocketPath returns the full path to the daemon Unix socket
func getSocketPath() (string, error) {
	pasoDir, err := getPasoDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(pasoDir, "paso.sock"), nil
}
