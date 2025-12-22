package main

import (
	"strings"
)

// normalizeString returns a normalized string (trimmed, lowercased).
func normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// lastNonEmptyLine returns the last non-empty line from the given string.
//
// Used to extract error messages from command output and parse errors.
func lastNonEmptyLine(s string) string {
	lines := splitStringOnNewline(s)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

// splitStringOnNewline normalizes line endings to Unix-style LF then splits.
//
// Returns a slice of lines.
func splitStringOnNewline(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

// splitStringOnDelim splits a string on the first dot (.) delimiter.
//
// Used to split grouped aliases like "group.nickname".
// Returns ok=false if the string does not contain a valid delimiter.
func splitStringOnDelim(alias string) (groupRaw, nicknameRaw string, ok bool) {
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
