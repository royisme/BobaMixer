// Package secrets manages scoped secret material for tool execution.
package secrets

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/royisme/bobamixer/internal/bobaerrors"
)

// Secrets holds resolved key/value pairs accessible by secret:// references.
type Secrets struct {
	Values map[string]string `json:"values" yaml:"values"`
}

// ValidatePermissions ensures the file is only readable/writable by the user (0600).
func ValidatePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat secrets file: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("secrets file %s must be chmod 600: %w", path, bobaerrors.ErrSecretsPerm)
	}
	return nil
}

// Load reads the secrets document and returns a Secrets struct.
func Load(path string) (*Secrets, error) {
	if err := ValidatePermissions(path); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path) //nolint:gosec // secrets path supplied by user configuration
	if err != nil {
		return nil, fmt.Errorf("read secrets file: %w", err)
	}
	if len(data) == 0 {
		return &Secrets{Values: map[string]string{}}, nil
	}
	values, err := parseValues(data)
	if err != nil {
		return nil, err
	}
	return &Secrets{Values: values}, nil
}

func parseValues(data []byte) (map[string]string, error) {
	values := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sep := strings.IndexAny(line, ":=")
		if sep == -1 {
			return nil, fmt.Errorf("invalid secrets line: %s", line)
		}
		key := strings.TrimSpace(line[:sep])
		value := strings.TrimSpace(line[sep+1:])
		if key == "" {
			return nil, fmt.Errorf("missing key in secrets line: %s", line)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan secrets: %w", err)
	}
	return values, nil
}

// ResolveEnv merges profile env directives with resolved secret values.
func ResolveEnv(profileEnv map[string]string, s *Secrets) (env []string, missing []string) {
	keys := make([]string, 0, len(profileEnv))
	for k := range profileEnv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := map[string]string{}
	if s != nil && s.Values != nil {
		for k, v := range s.Values {
			values[strings.TrimSpace(k)] = v
		}
	}

	for _, key := range keys {
		val := profileEnv[key]
		if strings.HasPrefix(val, "secret://") {
			lookup := strings.TrimPrefix(val, "secret://")
			secretVal, ok := values[lookup]
			if !ok || secretVal == "" {
				missing = append(missing, lookup)
				continue
			}
			env = append(env, fmt.Sprintf("%s=%s", key, secretVal))
			continue
		}
		env = append(env, fmt.Sprintf("%s=%s", key, val))
	}
	return env, missing
}
