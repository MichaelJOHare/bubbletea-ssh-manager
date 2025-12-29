package connect

import (
	"bubbletea-ssh-manager/internal/host"
	str "bubbletea-ssh-manager/internal/stringutil"
	"strings"
)

// A Target represents a connection target for SSH or Telnet.
// It includes the protocol and host specification.
type Target struct {
	Protocol  string // "ssh" or "telnet"
	host.Spec        // shared host fields (alias/hostname/port/user)
}

// Display returns the human-readable target for status messages.
//
// Examples:
//   - ssh:    mike@krabby <10.0.0.147:22>
//   - telnet: router <10.0.0.1:23>
func (t Target) Display() string {
	alias := strings.TrimSpace(t.Alias)
	user := strings.TrimSpace(t.User)
	hostName := strings.TrimSpace(t.HostName)
	port := strings.TrimSpace(t.Port)

	displayAlias := alias
	if user != "" {
		displayAlias = user + "@" + alias
	}
	if hostName != "" && port != "" {
		return displayAlias + " <" + hostName + ":" + port + ">"
	}
	return displayAlias
}

// WindowTitle returns a stable short title for the terminal/tab.
// Format: "ssh mike@KRABBY" (or group.HOST with host uppercased).
func (t Target) WindowTitle() string {
	protocol := strings.TrimSpace(t.Protocol)
	alias := strings.TrimSpace(t.Alias)
	user := strings.TrimSpace(t.User)

	if alias == "" {
		return protocol
	}

	// uppercase only the host portion of grouped aliases: group.HOST
	aliasTitle := strings.ToUpper(alias)
	if g, h, ok := str.SplitStringOnDelim(alias); ok {
		aliasTitle = g + "." + strings.ToUpper(h)
	}

	if user != "" {
		return protocol + " " + user + "@" + aliasTitle
	}
	return protocol + " " + aliasTitle
}
