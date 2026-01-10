package tui

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"

	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"
)

// addMenuItem adds a host/group menu item to the root or a group, based on the alias format.
func addMenuItem(ungrouped *[]*menuItem, groups map[string]*menuItem, host *menuItem) {
	if host == nil {
		return
	}
	alias := host.spec.Alias
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
		ungrouped []*menuItem
		groups    = map[string]*menuItem{}
		parseErrs []error
	)

	parseErrs = append(parseErrs, parseConfigToMenu(sshPath, config.ProtocolSSH, &ungrouped, groups))
	parseErrs = append(parseErrs, parseConfigToMenu(telnetPath, config.ProtocolTelnet, &ungrouped, groups))

	items := buildSortedMenuItems(ungrouped, groups)
	return items, errors.Join(parseErrs...)
}

// parseConfigToMenu parses a config file and adds entries to the menu structure.
func parseConfigToMenu(path string, protocol config.Protocol, ungrouped *[]*menuItem, groups map[string]*menuItem) error {
	entries, err := config.ParseConfigRecursively(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s config: %w", protocol, err)
	}

	for _, e := range entries {
		if !isValidEntry(e, protocol) {
			continue
		}
		h := &menuItem{kind: itemHost, protocol: protocol, spec: e.Spec, options: e.SSHOptions}
		addMenuItem(ungrouped, groups, h)
	}
	return nil
}

// isValidEntry checks if a config entry has required fields for the given protocol.
func isValidEntry(e config.HostEntry, protocol config.Protocol) bool {
	if e.Spec.Alias == "" {
		return false
	}
	if protocol == config.ProtocolTelnet && e.Spec.HostName == "" {
		return false
	}
	return true
}

// buildSortedMenuItems converts groups map to slice and sorts all items alphabetically.
func buildSortedMenuItems(ungrouped []*menuItem, groups map[string]*menuItem) []*menuItem {
	sortByName := func(a, b *menuItem) int { return cmp.Compare(a.name, b.name) }

	grouped := make([]*menuItem, 0, len(groups))
	for _, g := range groups {
		slices.SortStableFunc(g.children, sortByName)
		grouped = append(grouped, g)
	}

	slices.SortStableFunc(ungrouped, sortByName)
	slices.SortStableFunc(grouped, sortByName)

	return append(ungrouped, grouped...)
}
