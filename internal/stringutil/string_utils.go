package stringutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"bubbletea-ssh-manager/internal/config"
)

// NormalizeString returns a normalized string (trimmed, lowercased).
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// NormalizePort returns a validated numeric port for a protocol.
//
// Behavior:
//   - If port is empty, returns the default port for the protocol (ssh=22, telnet=23).
//   - If port is numeric, validates it's within 1..65535 and returns it.
func NormalizePort(port string, protocol config.Protocol) (string, error) {
	port = strings.TrimSpace(port)

	if port == "" {
		switch protocol {
		case config.ProtocolSSH:
			return "22", nil
		case config.ProtocolTelnet:
			return "23", nil
		}
	}

	n, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("invalid %s port %q", protocol, port)
	}
	if n < 1 || n > 65535 {
		return "", fmt.Errorf("port out of range: %d", n)
	}
	return strconv.Itoa(n), nil
}

// SplitStringOnDelim splits an alias of the form "group.nickname" into its parts.
//
// Returns ok=false if the alias is not in the expected format.
func SplitStringOnDelim(alias string) (substring1, substring2 string, ok bool) {
	before, after, ok := strings.Cut(alias, ".")
	if !ok {
		return "", "", false
	}
	before = NormalizeString(before)
	after = NormalizeString(after)
	if before == "" || after == "" {
		return "", "", false
	}
	return before, after, true
}

// LastNonEmptyLine returns the last non-empty line from the given string.
//
// Used to extract error messages from command output and parse errors.
func LastNonEmptyLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := NormalizeString(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

// BuildAliasFromGroupNickname constructs a full alias from group and nickname.
//
// It validates the nickname and (optionally) the group, then joins them with a dot.
// If group is empty, it returns a nickname-only alias (for ungrouped hosts).
// Returns an error if validation fails.
func BuildAliasFromGroupNickname(group string, nickname string) (string, error) {
	if err := ValidateHostGroup(group); err != nil {
		return "", err
	}
	if err := ValidateHostNickname(nickname); err != nil {
		return "", err
	}
	g := NormalizeString(FormatAliasForConfig(group))
	n := NormalizeString(FormatAliasForConfig(nickname))
	if n == "" {
		return "", errors.New("nickname is required")
	}
	if g == "" {
		return n, nil
	}
	return g + "." + n, nil
}

// FormatAliasForConfig formats a display name into a config alias.
//
// It trims whitespace, replaces spaces with hyphens, and collapses multiple hyphens.
func FormatAliasForConfig(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "-", " ")
	parts := strings.Fields(s)
	return strings.Join(parts, "-")
}

// SplitAliasForDisplay splits a full alias into group and nickname for display.
//
// It formats each part for display (hyphens to spaces, trimming, case).
func SplitAliasForDisplay(alias string) (groupName string, nickname string) {
	alias = NormalizeString(alias)
	if alias == "" {
		return "", ""
	}
	groupRaw, nickRaw, ok := strings.Cut(alias, ".")
	if ok {
		groupName = FormatDisplayName(groupRaw, true)
		nickname = FormatDisplayName(nickRaw, false)
		return groupName, nickname
	}
	return "", FormatDisplayName(alias, false)
}

// FormatDisplayName formats a raw name for display.
//
// It replaces hyphens with spaces, trims whitespace, collapses
// multiple spaces, and converts to uppercase if isGroup is true.
func FormatDisplayName(raw string, isGroup bool) string {
	s := strings.ReplaceAll(raw, "-", " ")
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	if isGroup {
		return strings.ToUpper(s)
	}
	return strings.ToLower(s)
}
