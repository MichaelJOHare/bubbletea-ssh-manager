package config

import (
	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"
	"fmt"
	"path/filepath"
	"strings"
)

// HostEntry is a minimal representation of a Host block from an SSH-style config.
// It intentionally contains only the fields this project currently supports.
type HostEntry struct {
	Spec       host.Spec       // host fields that are shared between SSH and Telnet (alias/hostname/port/user)
	SSHOptions sshopts.Options // SSH-specific options for this host
	SourcePath string          // path to the config file this entry was read from
}

// maxIncludeDepth is the maximum depth for recursive Include parsing.
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
//   - SSH options: HostKeyAlgorithms, KexAlgorithms, MACs
//
// It returns a slice of HostEntry structs representing the parsed hosts.
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

	// store output entries, current host aliases, and parsed values
	var out []HostEntry
	currentAliases := []string{}
	localOrder := []string{}
	localSeen := map[string]bool{}
	values := map[string]*HostEntry{}

	// helper to get or create a HostEntry for the given alias
	getOrCreate := func(alias string) *HostEntry {
		if it, ok := values[alias]; ok && it != nil {
			return it
		}
		it := &HostEntry{Spec: host.Spec{Alias: alias}, SourcePath: path}
		values[alias] = it
		return it
	}
	// helper to apply a function to all current aliases' HostEntrys
	applyToCurrent := func(apply func(*HostEntry)) {
		for _, a := range currentAliases {
			if it := values[a]; it != nil {
				apply(it)
			}
		}
	}

	// directory of the current config file (for relative includes)
	relDir := filepath.Dir(path)

	// parse lines and split into directives
	for _, raw := range lines {
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		// handle include and host directives specially
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
				getOrCreate(a)
				if !localSeen[a] {
					localSeen[a] = true
					localOrder = append(localOrder, a)
				}
			}

		// handle other directives generically
		default:
			if len(fields) < 2 {
				continue
			}
			value := strings.Join(fields[1:], " ")
			applyToCurrent(func(it *HostEntry) {
				parseHostBlockDirective(key, value, it)
			})
		}
	}

	// collect output entries in order
	for _, a := range localOrder {
		if it := values[a]; it != nil {
			out = append(out, *it)
		}
	}

	return out, nil
}

// parseHostBlockDirective parses and applies a single Host block directive
// to the given HostEntry.
//
// It returns true if the directive was recognized and applied.
func parseHostBlockDirective(key string, value string, entry *HostEntry) bool {
	if entry == nil {
		return false
	}

	// standard host directives shared between SSH and Telnet
	switch key {
	case "hostname":
		entry.Spec.HostName = value
		return true
	case "port":
		entry.Spec.Port = value
		return true
	case "user":
		entry.Spec.User = value
		return true

	// SSH options (a small subset), values are usually comma separated
	case "hostkeyalgorithms":
		entry.SSHOptions.HostKeyAlgorithms = value
		return true
	case "kexalgorithms":
		entry.SSHOptions.KexAlgorithms = value
		return true
	case "macs":
		entry.SSHOptions.MACs = value
		return true
	}

	return false
}
