package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// parseConfigRecursively parses an ssh-style config file at the given path,
// handling Include directives recursively up to a certain depth.
//
// Supported directives:
//   - Host
//   - HostName
//   - Port
//   - Include (with basic glob support)
func parseConfigRecursively(path string, depth int) ([]hostEntry, error) {
	// who really needs more than 5 levels of includes anyway
	if depth > 5 {
		return nil, fmt.Errorf("config include depth exceeded")
	}

	// read lines from config file
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	// build host entries
	var out []hostEntry
	currentAliases := []string{}
	localOrder := []string{}
	localSeen := map[string]bool{}
	values := map[string]hostEntry{}

	// get directory of current file for relative includes
	relDir := filepath.Dir(path)

	// process each line
	for _, raw := range lines {
		// strip comments and skip empty lines
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		fields := splitFields(line)
		if len(fields) == 0 {
			continue
		}

		// handle directives
		key := strings.ToLower(fields[0])
		switch key {
		case "include":
			// basic include support, also supports globs
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
					values[a] = hostEntry{alias: a}
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
				e := values[a]
				e.hostname = v
				values[a] = e
			}

		case "port":
			if len(fields) < 2 {
				continue
			}
			p := fields[1]
			for _, a := range currentAliases {
				e := values[a]
				e.port = p
				values[a] = e
			}

		}
	}

	// append local hosts after any included hosts, preserving first appearance
	// order from this file
	for _, a := range localOrder {
		out = append(out, values[a])
	}

	return out, nil
}
