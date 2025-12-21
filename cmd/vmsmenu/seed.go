package main

func seedMenu() (*menuItem, error) {
	items, err := buildMenuFromConfigs()
	if len(items) > 0 {
		return &menuItem{kind: itemGroup, name: "Hosts", children: items}, err
	}

	// fallback stub data so the UI still has something
	// to show when no config files exist yet and hint to user to create them
	l2 := &menuItem{
		kind: itemGroup,
		name: "IF YOU'RE SEEING THIS IT MEANS NO SSH OR TELNET CONFIG FILES WERE FOUND",
		children: []*menuItem{
			{kind: itemHost, name: "l2.IA21", protocol: "ssh", target: "l2.IA21"},
			{kind: itemHost, name: "l2.IA22", protocol: "ssh", target: "l2.IA22"},
		},
	}

	return &menuItem{
		kind: itemGroup,
		name: "Hosts",
		children: []*menuItem{
			l2,
			{kind: itemHost, name: "devbox", protocol: "ssh", target: "devbox"},
			{kind: itemHost, name: "router", protocol: "telnet", target: "router:23"},
		},
	}, err
}
