package main

import (
	"bubbletea-ssh-manager/internal/config"

	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"
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

// addMenuItem adds a host/group menu item to the root or a group, based on the alias format.
func addMenuItem(ungrouped *[]*menuItem, groups map[string]*menuItem, host *menuItem) {
	if host == nil {
		return
	}
	alias := strings.TrimSpace(host.alias)
	if alias == "" {
		return
	}

	// grouped alias: add to group (create group if needed)
	groupRaw, nickRaw, ok := splitGroupedAlias(alias)
	if ok {
		groupName := formatGroupName(groupRaw)
		displayName := normalizeString(nickRaw)
		host.name = displayName

		g, exists := groups[groupName]
		if !exists {
			g = &menuItem{kind: itemGroup, name: groupName}
			groups[groupName] = g
		}
		g.children = append(g.children, host)
		return
	}

	// ungrouped alias: lowercase it for display and add to root
	displayName := normalizeString(alias)
	host.name = displayName
	*ungrouped = append(*ungrouped, host)
}

// buildMenuFromConfigs builds menu items from SSH and Telnet config files.
//
// SSH items connect by alias (ssh reads ~/.ssh/config).
// Telnet items connect by HostName/Port because telnet typically does not use aliases.
func buildMenuFromConfigs() ([]*menuItem, error) {
	sshPath, err := config.UserConfigPath(".ssh", "config")
	if err != nil {
		return nil, err
	}
	telnetPath, err := config.UserConfigPath(".telnet", "config")
	if err != nil {
		return nil, err
	}

	var (
		ungrouped []*menuItem              // hosts without groups
		groups    = map[string]*menuItem{} // map of group names
		parseErrs []error                  // parsing errors
	)

	// parse ssh config, add hosts to menu
	if sshEntries, err := config.ParseConfigRecursively(sshPath); err == nil {
		for _, e := range sshEntries {
			alias := strings.TrimSpace(e.Alias)
			if alias == "" {
				continue
			}
			h := &menuItem{kind: itemHost, protocol: "ssh", alias: alias, hostname: e.HostName, port: e.Port, user: e.User}
			addMenuItem(&ungrouped, groups, h)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		parseErrs = append(parseErrs, fmt.Errorf("read ssh config: %w", err))
	}

	// parse telnet config, add hosts to menu
	if telnetEntries, err := config.ParseConfigRecursively(telnetPath); err == nil {
		for _, e := range telnetEntries {
			alias := strings.TrimSpace(e.Alias)
			host := strings.TrimSpace(e.HostName)
			if alias == "" || host == "" {
				continue
			}
			h := &menuItem{kind: itemHost, protocol: "telnet", alias: alias, hostname: host, port: e.Port, user: e.User}
			addMenuItem(&ungrouped, groups, h)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		parseErrs = append(parseErrs, fmt.Errorf("read telnet config: %w", err))
	}

	// convert groups map to slice for sorting
	grouped := make([]*menuItem, 0, len(groups))
	for _, g := range groups {
		grouped = append(grouped, g)
	}

	// sort ungrouped hosts, group names, and children of groups (grouped hosts) alphabetically
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
