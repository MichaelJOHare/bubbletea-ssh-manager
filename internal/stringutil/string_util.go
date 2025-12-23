package stringutil

import "strings"

// NormalizeString returns a normalized string (trimmed, lowercased).
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// SplitStringOnNewline normalizes line endings to Unix-style LF then splits.
//
// Returns a slice of lines.
func SplitStringOnNewline(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
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
	lines := SplitStringOnNewline(s)
	for i := len(lines) - 1; i >= 0; i-- {
		line := NormalizeString(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}
