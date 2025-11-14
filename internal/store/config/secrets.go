package config

import (
	"fmt"
	"os"
	"strings"
)

// ResolveEnv resolves environment variables from profiles.yaml,
// replacing secret://name references with actual values from secrets.yaml
func ResolveEnv(env map[string]string, secrets Secrets) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		val := resolveSecretRef(v, secrets)
		out = append(out, fmt.Sprintf("%s=%s", k, val))
	}
	return out
}

// resolveSecretRef resolves a single secret reference
func resolveSecretRef(value string, secrets Secrets) string {
	const prefix = "secret://"
	if !strings.HasPrefix(value, prefix) {
		return value
	}

	key := strings.TrimPrefix(value, prefix)
	if real, ok := secrets[key]; ok {
		return real
	}

	// Fallback to environment variable if secret not found
	if envVal := os.Getenv(key); envVal != "" {
		return envVal
	}

	// Return empty if not found
	return ""
}

// ValidateSecretsPermissions checks that secrets.yaml has correct permissions (0600)
func ValidateSecretsPermissions(home string) error {
	path := home + "/secrets.yaml"
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, no validation needed
		}
		return err
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		return fmt.Errorf(
			"secrets.yaml has insecure permissions (%04o), should be 0600\n"+
				"Fix: chmod 600 %s/secrets.yaml",
			mode, home)
	}

	return nil
}
