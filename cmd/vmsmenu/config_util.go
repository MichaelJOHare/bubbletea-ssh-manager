package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// userConfigPath returns the path under the user's home directory.
//
// Note: when run under MSYS2/Cygwin on Windows, os.UserHomeDir() points to the
// Windows home directory, not the MSYS2/Cygwin POSIX-style home.
func userConfigPath(parts ...string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(append([]string{home}, parts...)...), nil
}

// readLines reads the given file and returns its lines.
// It normalizes line endings to Unix-style LF.
func readLines(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// normalize Windows CRLF so line splitting behaves predictably
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n"), nil
}

// stripComment removes any comment from the line (text after a '#').
func stripComment(line string) string {
	if before, _, ok := strings.Cut(line, "#"); ok {
		return before
	}
	return line
}

// splitFields splits a line into fields separated by whitespace.
func splitFields(line string) []string {
	return strings.Fields(line)
}

// isSimpleAlias returns true if the alias has no wildcard/negation characters.
func isSimpleAlias(s string) bool {
	if s == "" {
		return false
	}
	// SSH supports patterns in Host directives, we only treat simple names as menu entries
	return !strings.ContainsAny(s, "*?!")
}

// expandPath expands a path that may start with '~' to the user's home directory.
func expandPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
		rest := strings.TrimPrefix(path, "~")
		rest = strings.TrimPrefix(rest, string(filepath.Separator))
		rest = strings.TrimPrefix(rest, "/")
		return filepath.Join(home, rest), nil
	}
	return path, nil
}
