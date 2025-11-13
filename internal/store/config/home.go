package config

import (
	"os"
	"path/filepath"
)

func ResolveHome() (string, error) {
	if custom := os.Getenv("BOBA_HOME"); custom != "" {
		return custom, nil
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".boba"), nil
}
