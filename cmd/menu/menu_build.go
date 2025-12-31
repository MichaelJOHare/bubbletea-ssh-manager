package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"
)

// addMenuItem adds a host/group menu item to the root or a group, based on the alias format.
func addMenuItem(ungrouped *[]*menuItem, groups map[string]*menuItem, host *menuItem) {
	if host == nil {
		return
	}
	alias := strings.TrimSpace(host.spec.Alias)
	if alias == "" {
		return
	}

	// grouped alias: add to group (create group if needed)
	groupRaw, nickRaw, ok := str.SplitStringOnDelim(alias)
	if ok {
		groupName := str.FormatDisplayName(groupRaw, true)
		displayName := str.FormatDisplayName(nickRaw, false)
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
	displayName := str.FormatDisplayName(alias, false)
	host.name = displayName
	*ungrouped = append(*ungrouped, host)
}

// buildMenuFromConfigs builds menu items from SSH and Telnet config files.
//
// SSH items connect by alias (ssh reads ~/.ssh/config).
// Telnet items connect by HostName/Port because telnet typically does not use aliases.
func buildMenuFromConfigs() ([]*menuItem, error) {
	sshPath, err := config.GetConfigPath(".ssh", "config")
	if err != nil {
		return nil, err
	}
	telnetPath, err := config.GetConfigPath(".telnet", "config")
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
			alias := strings.TrimSpace(e.Spec.Alias)
			if alias == "" {
				continue
			}
			h := &menuItem{kind: itemHost, protocol: "ssh", spec: e.Spec, options: e.SSHOptions}
			addMenuItem(&ungrouped, groups, h)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		parseErrs = append(parseErrs, fmt.Errorf("read ssh config: %w", err))
	}

	// parse telnet config, add hosts to menu
	if telnetEntries, err := config.ParseConfigRecursively(telnetPath); err == nil {
		for _, e := range telnetEntries {
			alias := strings.TrimSpace(e.Spec.Alias)
			host := strings.TrimSpace(e.Spec.HostName)
			if alias == "" || host == "" {
				continue
			}
			h := &menuItem{kind: itemHost, protocol: "telnet", spec: e.Spec}
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
