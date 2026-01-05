package tui

import "bubbletea-ssh-manager/internal/config"

// seedMenu creates the initial menu structure.
//
// It builds the menu from existing config files.
// If no config files are found, it returns a stub menu with sample data.
func seedMenu() (*menuItem, error) {
	// try to build menu from config files
	items, err := buildMenuFromConfigs()
	if len(items) > 0 {
		return &menuItem{kind: itemGroup, name: "HOME", children: items}, err
	}

	// fallback stub data so the UI still has something
	// to show when no config files exist yet and hint to user to create them
	l1 := &menuItem{
		kind: itemGroup,
		name: "IF YOU'RE SEEING THIS IT MEANS NO SSH OR TELNET CONFIG FILES WERE FOUND",
		children: []*menuItem{
			{kind: itemHost, name: "stub", protocol: config.ProtocolSSH, spec: config.Spec{Alias: "l2.IA21"}},
			{kind: itemHost, name: "stub", protocol: config.ProtocolTelnet, spec: config.Spec{Alias: "l2.IA21"}},
		},
	}

	l2 := &menuItem{
		kind: itemGroup,
		name: "Create ~/.ssh/config and/or ~/.telnet/config in MSYS2 home directory (see README)",
	}

	return &menuItem{
		kind: itemGroup,
		name: "NO CONFIG FOUND",
		children: []*menuItem{
			l1,
			l2,
			{kind: itemHost, name: "stub", protocol: config.ProtocolSSH, spec: config.Spec{Alias: "devbox"}},
			{kind: itemHost, name: "stub", protocol: config.ProtocolTelnet, spec: config.Spec{Alias: "router", HostName: "router", Port: "23"}},
		},
	}, err
}
