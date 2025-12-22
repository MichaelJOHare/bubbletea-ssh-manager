package connect

import "strings"

type Target struct {
	protocol string
	alias    string
	user     string
	host     string
	port     string
}

func (t Target) Protocol() string {
	return t.protocol
}

// Display returns the human-readable target for status messages.
// Examples:
//   - ssh:    mike@krabby <10.0.0.147:22>
//   - ssh:    krabby
//   - telnet: router <router:23>
func (t Target) Display() string {
	alias := strings.TrimSpace(t.alias)
	user := strings.TrimSpace(t.user)
	host := strings.TrimSpace(t.host)
	port := strings.TrimSpace(t.port)

	displayAlias := alias
	if user != "" {
		displayAlias = user + "@" + alias
	}
	if host != "" && port != "" {
		return displayAlias + " <" + host + ":" + port + ">"
	}
	return displayAlias
}

// WindowTitle returns a stable short title for the terminal/tab.
// Format: "ssh mike@KRABBY" (or group.HOST with host uppercased).
func (t Target) WindowTitle() string {
	protocol := strings.TrimSpace(t.protocol)
	alias := strings.TrimSpace(t.alias)
	user := strings.TrimSpace(t.user)

	if alias == "" {
		return protocol
	}

	// Uppercase only the host portion of grouped aliases: group.HOST
	aliasTitle := strings.ToUpper(alias)
	if g, h, ok := splitGroupedAlias(alias); ok {
		aliasTitle = g + "." + strings.ToUpper(h)
	}

	if user != "" {
		return protocol + " " + user + "@" + aliasTitle
	}
	return protocol + " " + aliasTitle
}

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
