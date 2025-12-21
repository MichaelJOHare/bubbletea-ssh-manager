package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// splitGroupedAlias splits an alias of the form "group.nickname" into its parts.
//
// Returns ok=false if the alias is not in the expected format.
func splitGroupedAlias(alias string) (groupRaw, nicknameRaw string, ok bool) {
	before, after, ok := strings.Cut(alias, ".")
	if !ok {
		return "", "", false
	}
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)
	if before == "" || after == "" {
		return "", "", false
	}
	return before, after, true
}

// splitHostPort attempts to split a target string into host and port.
// It supports "host port" and "host:port" formats (single-colon, numeric port).
//
// Returns ok=false if no valid port could be found.
func splitHostPort(target string) (host, port string, ok bool) {
	fields := strings.Fields(target)
	if len(fields) >= 2 {
		if _, err := strconv.Atoi(fields[1]); err == nil {
			return fields[0], fields[1], true
		}
		return "", "", false
	}

	// host:port (avoid IPv6 nonsense; only accept exactly one colon)
	if strings.Count(target, ":") != 1 {
		return "", "", false
	}
	before, after, okSplit := strings.Cut(target, ":")
	if !okSplit {
		return "", "", false
	}
	if before == "" || after == "" {
		return "", "", false
	}
	if _, err := strconv.Atoi(after); err != nil {
		return "", "", false
	}
	return before, after, true
}

// formatGroupName formats a raw group name for display.
//
// It replaces hyphens with spaces, trims whitespace, collapses
// multiple spaces, and converts to uppercase.
func formatGroupName(raw string) string {
	s := strings.ReplaceAll(raw, "-", " ")
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	return strings.ToUpper(s)
}

// formatNickname formats a raw nickname for display.
//
// It trims whitespace and converts to lowercase.
func formatNickname(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// addMenuHost adds a host menu item to the root or a group, based on the alias format.
func addMenuHost(ungrouped *[]*menuItem, grouped *[]*menuItem, groups map[string]*menuItem, alias, protocol, target string) {
	// grouped alias: add to group (create group if needed)
	groupRaw, nickRaw, ok := splitGroupedAlias(alias)
	if ok {
		groupName := formatGroupName(groupRaw)
		displayName := formatNickname(nickRaw)

		g, exists := groups[groupName]
		if !exists {
			g = &menuItem{kind: itemGroup, name: groupName}
			groups[groupName] = g
			*grouped = append(*grouped, g)
		}
		g.children = append(g.children, &menuItem{kind: itemHost, name: displayName, protocol: protocol, target: target})
		return
	}

	// ungrouped alias: lowercase it for display and add to root
	displayName := formatNickname(alias)
	*ungrouped = append(*ungrouped, &menuItem{kind: itemHost, name: displayName, protocol: protocol, target: target})
}

// buildMenuFromConfigs builds menu items from SSH and Telnet config files.
//
// SSH items connect by alias (ssh reads ~/.ssh/config).
// Telnet items connect by HostName/Port because telnet typically does not use aliases.
func buildMenuFromConfigs() ([]*menuItem, error) {
	sshPath, err := userConfigPath(".ssh", "config")
	if err != nil {
		return nil, err
	}
	telnetPath, err := userConfigPath(".telnet", "config")
	if err != nil {
		return nil, err
	}

	var ungrouped []*menuItem
	var grouped []*menuItem
	groups := map[string]*menuItem{}
	var parseErrs []error

	if sshEntries, err := parseConfigRecursively(sshPath, 0); err == nil {
		for _, e := range sshEntries {
			alias := strings.TrimSpace(e.alias)
			if alias == "" {
				continue
			}
			addMenuHost(&ungrouped, &grouped, groups, alias, "ssh", alias)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		parseErrs = append(parseErrs, fmt.Errorf("read ssh config: %w", err))
	}

	if telnetEntries, err := parseConfigRecursively(telnetPath, 0); err == nil {
		for _, e := range telnetEntries {
			alias := strings.TrimSpace(e.alias)
			host := strings.TrimSpace(e.hostname)
			if alias == "" || host == "" {
				continue
			}
			target := host
			if port := strings.TrimSpace(e.port); port != "" {
				target = host + ":" + port
			}
			addMenuHost(&ungrouped, &grouped, groups, alias, "telnet", target)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		parseErrs = append(parseErrs, fmt.Errorf("read telnet config: %w", err))
	}

	slices.SortStableFunc(ungrouped, func(a, b *menuItem) int {
		return cmp.Compare(a.name, b.name)
	})
	slices.SortStableFunc(grouped, func(a, b *menuItem) int {
		return cmp.Compare(a.name, b.name)
	})
	for _, g := range grouped {
		slices.SortStableFunc(g.children, func(a, b *menuItem) int {
			return cmp.Compare(a.name, b.name)
		})
	}

	items := append(ungrouped, grouped...)
	return items, errors.Join(parseErrs...)
}
