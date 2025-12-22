package main

import "strings"

// lastNonEmptyLine returns the last non-empty line from the given string.
// Used to extract error messages from command output and parse errors.
func lastNonEmptyLine(s string) string {
	lines := splitStringIntoLines(s)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

// splitStringIntoLines normalizes line endings to Unix-style LF then splits.
func splitStringIntoLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}
