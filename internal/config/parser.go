package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// maxIncludeDepth is the maximum depth for recursive Include parsing.
const maxIncludeDepth = 5

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
	// directory of the current config file (for relative includes)
	relDir := filepath.Dir(path)

	// parse lines and split into directives, stripping comments and blank lines
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
				inc, err := expandPath(incRaw) // expand ~ in include path
				if err != nil {
					continue
				}
				if !filepath.IsAbs(inc) {
					inc = filepath.Join(relDir, inc) // make relative to current config file
				}
				matches, err := filepath.Glob(inc)
				if err != nil {
					continue
				}
				for _, m := range matches {
					more, err := parseConfigRecursively(m, depth+1) // recurse into included files
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
				// if new alias, create new HostEntry, else reuse existing
				currentAliases = append(currentAliases, a)
				if it, ok := values[a]; !ok || it == nil {
					values[a] = &HostEntry{Spec: Spec{Alias: a}, SourcePath: path}
				}
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
			for _, a := range currentAliases {
				if it := values[a]; it != nil {
					setHostDirective(key, value, it)
				}
			}
		}
	}

	// collect output entries in order
	for _, a := range localOrder {
		if it := values[a]; it != nil {
			out = append(out, it.Normalized())
		}
	}

	return out, nil
}

// parseHostHeader returns (indent, aliases, comment, ok).
//
// - indent is leading whitespace on the original line
// - comment includes the leading '#' if present
func parseHostHeader(line string) (string, []string, string, bool) {
	if strings.TrimSpace(line) == "" {
		return "", nil, "", false
	}

	// preserve indent
	trimmedLeft := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmedLeft)]

	// split comment (preserve it)
	before, after, hasComment := strings.Cut(trimmedLeft, "#")
	comment := ""
	if hasComment {
		comment = "#" + after
	}
	before = strings.TrimSpace(before)
	if before == "" {
		return "", nil, "", false
	}
	fields := strings.Fields(before) // split on whitespace
	if len(fields) < 2 {             // need at least "Host" and one alias
		return "", nil, "", false
	}
	if strings.ToLower(fields[0]) != "host" {
		return "", nil, "", false
	}
	aliases := make([]string, 0, len(fields)-1)
	for _, a := range fields[1:] {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		aliases = append(aliases, a)
	}
	if len(aliases) == 0 {
		return "", nil, "", false
	}
	return indent, aliases, strings.TrimRight(comment, "\r\n"), true
}
