package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func effectiveHomeDir() (string, error) {
	if h := strings.TrimSpace(os.Getenv("HOME")); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return home, nil
}

func expandPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(path, "~") {
		home, err := effectiveHomeDir()
		if err != nil {
			return "", err
		}
		rest := strings.TrimPrefix(path, "~")
		rest = strings.TrimPrefix(rest, string(filepath.Separator))
		rest = strings.TrimPrefix(rest, "/")
		return filepath.Join(home, rest), nil
	}
	return path, nil
}

func readLines(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n"), nil
}

func stripComment(line string) string {
	if before, _, ok := strings.Cut(line, "#"); ok {
		return before
	}
	return line
}

func isSimpleAlias(s string) bool {
	if s == "" {
		return false
	}
	// OpenSSH supports patterns in Host directives; we only treat simple names
	// as concrete menu entries.
	return !strings.ContainsAny(s, "*?!")
}
