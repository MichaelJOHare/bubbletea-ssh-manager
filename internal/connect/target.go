package connect

import (
	"strings"

	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"
)

// A Target represents a connection target for SSH or Telnet.
// It includes the protocol and host specification.
type Target struct {
	Protocol    config.Protocol // "ssh" or "telnet"
	config.Spec                 // shared host fields (alias/hostname/port/user)
}

// Display returns the human-readable target for status messages.
//
// Examples:
//   - ssh:    mike@krabby <10.0.0.147:22>
//   - telnet: router <10.0.0.1:23>
func (t Target) Display() string {
	alias := t.Alias
	user := t.User
	hostName := t.HostName
	port := t.Port

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
	protocol := string(t.Protocol)
	alias := t.Alias
	user := t.User

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
