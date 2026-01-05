package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// getHomeDirectory returns the effective home directory for the current user.
//
// On Windows/MSYS2, prefer $HOME so this matches where MSYS2/OpenSSH tools
// look for config files (eg. ~/.ssh/config).
func getHomeDirectory() (string, error) {
	if h := strings.TrimSpace(os.Getenv("HOME")); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return home, nil
}

// expandPath expands a given path, replacing ~ with the effective home directory.
// Used mainly to expand include paths in config files.
//
// It returns an error if the path is empty or the home directory cannot be determined.
func expandPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(path, "~") {
		home, err := getHomeDirectory()
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

// readLines reads the file at the given path and returns its lines as a slice of strings.
//
// It normalizes line endings to LF.
func readLines(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n"), nil
}

// writeLines writes lines to path with LF endings.
func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return writeFileAtomic(path, []byte(content))
}

// writeFileAtomic writes data to path atomically.
//
// It creates parent directories as needed and preserves existing file permissions.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var mode os.FileMode = 0o644
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode()
	}

	f, err := os.CreateTemp(dir, ".tmp-config-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmp)
	}()

	if err := f.Chmod(mode); err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// on Windows, os.Rename won't overwrite an existing destination
	// remove first, this is not perfectly atomic but it's fine for this
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return os.Rename(tmp, path)
}

// stripComment removes any comment from the given line.
//
// A comment starts with a '#' character.
func stripComment(line string) string {
	if before, _, ok := strings.Cut(line, "#"); ok {
		return before
	}
	return line
}

// isSimpleAlias returns true if the given string is a simple alias (no patterns).
//
// Openssh supports patterns in Host directives; we only treat simple names
// as concrete menu entries.
func isSimpleAlias(s string) bool {
	if s == "" {
		return false
	}
	return !strings.ContainsAny(s, "*?!")
}

// fileContainsAlias returns true if any Host header in lines contains alias.
func fileContainsAlias(lines []string, alias string) bool {
	for _, raw := range lines {
		_, aliases, _, ok := parseHostHeader(raw)
		if !ok {
			continue
		}
		if slices.Contains(aliases, alias) {
			return true
		}
	}
	return false
}
