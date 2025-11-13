package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const activeProfileFilename = "active_profile"

// LoadActiveProfile returns the currently persisted profile key.
func LoadActiveProfile(home string) (string, error) {
	path := filepath.Join(home, activeProfileFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveActiveProfile persists the selected profile key with strict permissions.
func SaveActiveProfile(home, profile string) error {
	path := filepath.Join(home, activeProfileFilename)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(profile)), 0o600); err != nil {
		return err
	}
	return nil
}
