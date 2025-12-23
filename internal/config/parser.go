package config

import (
	"bubbletea-ssh-manager/internal/host"
	"fmt"
	"path/filepath"
	"strings"
)

// HostEntry is a minimal representation of a Host block from an SSH-style config.
// It intentionally contains only the fields this project currently supports.
type HostEntry = host.Spec

const maxIncludeDepth = 5

// UserConfigPath returns a path under the effective home directory.
//
// On Windows/MSYS2, prefer $HOME so this matches where MSYS2/OpenSSH tools
// look for config files (eg. ~/.ssh/config).
func UserConfigPath(parts ...string) (string, error) {
	home, err := effectiveHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, parts...)...), nil
}

// ParseConfigRecursively parses an SSH-style config file at the given path,
// following Include directives recursively (up to a small depth limit).
//
// Supported directives:
//   - Include (with basic glob support)
//   - Host
//   - HostName
//   - User
//   - Port
func ParseConfigRecursively(path string) ([]HostEntry, error) {
	return parseConfigRecursively(path, 0)
}

func parseConfigRecursively(path string, depth int) ([]HostEntry, error) {
	if depth > maxIncludeDepth {
		return nil, fmt.Errorf("config include depth exceeded")
	}

	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	var out []HostEntry
	currentAliases := []string{}
	localOrder := []string{}
	localSeen := map[string]bool{}
	values := map[string]*HostEntry{}

	relDir := filepath.Dir(path)

	for _, raw := range lines {
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		key := strings.ToLower(fields[0])
		switch key {
		case "include":
			for _, incRaw := range fields[1:] {
				inc, err := expandPath(incRaw)
				if err != nil {
					continue
				}
				if !filepath.IsAbs(inc) {
					inc = filepath.Join(relDir, inc)
				}
				matches, err := filepath.Glob(inc)
				if err != nil {
					continue
				}
				for _, m := range matches {
					more, err := parseConfigRecursively(m, depth+1)
					if err != nil {
						continue
					}
					out = append(out, more...)
				}
			}

		case "host":
			currentAliases = currentAliases[:0]
			for _, a := range fields[1:] {
				if !isSimpleAlias(a) {
					continue
				}
				currentAliases = append(currentAliases, a)
				if _, ok := values[a]; !ok {
					values[a] = &HostEntry{Alias: a}
				}
				if !localSeen[a] {
					localSeen[a] = true
					localOrder = append(localOrder, a)
				}
			}

		case "hostname":
			if len(fields) < 2 {
				continue
			}
			v := fields[1]
			for _, a := range currentAliases {
				if it := values[a]; it != nil {
					it.HostName = v
				}
			}

		case "port":
			if len(fields) < 2 {
				continue
			}
			p := fields[1]
			for _, a := range currentAliases {
				if it := values[a]; it != nil {
					it.Port = p
				}
			}

		case "user":
			if len(fields) < 2 {
				continue
			}
			u := fields[1]
			for _, a := range currentAliases {
				if it := values[a]; it != nil {
					it.User = u
				}
			}
		}
	}

	for _, a := range localOrder {
		if it := values[a]; it != nil {
			out = append(out, *it)
		}
	}

	return out, nil
}
