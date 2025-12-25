package config

import (
	"bubbletea-ssh-manager/internal/sshopts"
	"slices"

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultHostIndent = "    "

// FindHostEntry searches the recursively-parsed config rooted at rootConfigPath
// for a concrete (simple) alias and returns the first match.
//
// If the alias is not found, it returns (nil, nil).
func FindHostEntry(rootConfigPath, alias string) (*HostEntry, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, errors.New("empty alias")
	}
	entries, err := ParseConfigRecursively(rootConfigPath)
	if err != nil {
		// if the root doesn't exist, callers can treat it as "not found"
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	for i := range entries {
		if strings.TrimSpace(entries[i].Spec.Alias) == alias {
			// return a stable pointer
			e := entries[i]
			return &e, nil
		}
	}
	return nil, nil
}

// ResolveHostSourcePath returns the path of the file that defines alias.
//
// This is important when the root config uses Include directives, the caller
// can then edit the correct fragment file.
//
// If the alias is not found, it returns "" and nil.
func ResolveHostSourcePath(rootConfigPath, alias string) (string, error) {
	e, err := FindHostEntry(rootConfigPath, alias)
	if err != nil || e == nil {
		return "", err
	}
	return e.SourcePath, nil
}

// AddHostEntry appends entry as a new Host block to configPath.
//
// It errors if an entry with the same alias already exists in that file.
func AddHostEntry(configPath string, entry HostEntry) error {
	alias := strings.TrimSpace(entry.Spec.Alias)
	if alias == "" {
		return errors.New("entry alias is required")
	}
	if !isSimpleAlias(alias) {
		return fmt.Errorf("unsupported alias pattern: %q", alias)
	}

	lines, err := readLines(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				return err
			}
			lines = nil
		} else {
			return err
		}
	}

	if fileContainsAlias(lines, alias) {
		return fmt.Errorf("host %q already exists in %s", alias, configPath)
	}

	out := appendNonNil(lines, "")
	out = append(out, createHostEntry(entry)...)
	return writeLines(configPath, out)
}

// UpdateHostEntry updates (or renames) a Host entry.
//
// It removes any existing mention of oldAlias in configPath and then appends
// updated as a fresh Host block at the end of the file.
//
// If oldAlias doesn't exist in configPath, it returns os.ErrNotExist.
func UpdateHostEntry(configPath, oldAlias string, updated HostEntry) error {
	oldAlias = strings.TrimSpace(oldAlias)
	if oldAlias == "" {
		return errors.New("old alias is required")
	}
	newAlias := strings.TrimSpace(updated.Spec.Alias)
	if newAlias == "" {
		return errors.New("updated alias is required")
	}
	if !isSimpleAlias(oldAlias) || !isSimpleAlias(newAlias) {
		return errors.New("alias patterns are not supported")
	}

	lines, err := readLines(configPath)
	if err != nil {
		return err
	}
	if !fileContainsAlias(lines, oldAlias) {
		return os.ErrNotExist
	}

	stripped, _ := removeAliasFromLines(lines, oldAlias)
	if oldAlias != newAlias && fileContainsAlias(stripped, newAlias) {
		return fmt.Errorf("host %q already exists in %s", newAlias, configPath)
	}

	out := appendNonNil(stripped, "")
	out = append(out, createHostEntry(updated)...)
	return writeLines(configPath, out)
}

// RemoveHostEntry removes alias from the config file.
//
// If alias appears in a multi-alias Host header, it is removed from that header
// and the rest of the block is preserved.
//
// If alias isn't present, it returns os.ErrNotExist.
func RemoveHostEntry(configPath, alias string) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return errors.New("alias is required")
	}
	if !isSimpleAlias(alias) {
		return fmt.Errorf("unsupported alias pattern: %q", alias)
	}

	lines, err := readLines(configPath)
	if err != nil {
		return err
	}
	if !fileContainsAlias(lines, alias) {
		return os.ErrNotExist
	}
	out, _ := removeAliasFromLines(lines, alias)
	return writeLines(configPath, out)
}

// createHostEntry creates config lines for a HostEntry.
func createHostEntry(entry HostEntry) []string {
	alias := strings.TrimSpace(entry.Spec.Alias)
	indent := DefaultHostIndent

	out := []string{fmt.Sprintf("Host %s", alias)}
	if v := strings.TrimSpace(entry.Spec.HostName); v != "" {
		out = append(out, indent+"HostName "+v)
	}
	if v := strings.TrimSpace(entry.Spec.User); v != "" {
		out = append(out, indent+"User "+v)
	}
	if v := strings.TrimSpace(entry.Spec.Port); v != "" {
		out = append(out, indent+"Port "+v)
	}
	sshOpts := CreateSSHOptionsEntry(entry.SSHOptions, indent)
	if len(sshOpts) > 0 {
		out = append(out, sshOpts...)
	}
	// trailing blank line for readability
	out = append(out, "")
	return out
}

// CreateSSHOptionsEntry creates config lines for non-empty SSH options.
//
// It uses the given indent for each line.
func CreateSSHOptionsEntry(o sshopts.Options, indent string) []string {
	parts := make([]string, 0, 3)
	if v := strings.TrimSpace(o.HostKeyAlgorithms); v != "" {
		parts = append(parts, indent+"HostKeyAlgorithms "+v)
	}
	if v := strings.TrimSpace(o.KexAlgorithms); v != "" {
		parts = append(parts, indent+"KexAlgorithms "+v)
	}
	if v := strings.TrimSpace(o.MACs); v != "" {
		parts = append(parts, indent+"MACs "+v)
	}
	return parts
}

// fileContainsAlias returns true if any Host header in lines contains alias.
func fileContainsAlias(lines []string, alias string) bool {
	for _, raw := range lines {
		_, aliases, _, ok := parseHostHeader(raw)
		if !ok {
			continue
		}
		if slices.Contains(aliases, alias) {
			return true
		}
	}
	return false
}

// removeAliasFromLines removes alias from any Host headers in lines.
//
// Returns updated lines and a bool indicating whether any change was made.
func removeAliasFromLines(lines []string, alias string) ([]string, bool) {
	type block struct {
		header     string
		body       []string
		isHost     bool
		startIndex int
	}

	// split into blocks (host blocks and non-host spans)
	blocks := make([]block, 0, 16)
	i := 0
	for i < len(lines) {
		raw := lines[i]
		if _, _, _, ok := parseHostHeader(raw); ok {
			// host block
			start := i
			header := raw
			i++
			bodyStart := i
			for i < len(lines) {
				if _, _, _, ok2 := parseHostHeader(lines[i]); ok2 {
					break
				}
				i++
			}
			body := append([]string(nil), lines[bodyStart:i]...)
			blocks = append(blocks, block{header: header, body: body, isHost: true, startIndex: start})
			continue
		}

		// non-host single line block (kept as-is)
		blocks = append(blocks, block{header: raw, body: nil, isHost: false, startIndex: i})
		i++
	}

	// reconstruct lines, removing alias where found
	changed := false
	out := make([]string, 0, len(lines))
	for _, b := range blocks {
		if !b.isHost {
			out = append(out, b.header)
			continue
		}

		// host block
		indent, aliases, comment, ok := parseHostHeader(b.header)
		if !ok {
			out = append(out, b.header)
			out = append(out, b.body...)
			continue
		}
		// remove alias from aliases
		kept := make([]string, 0, len(aliases))
		removedHere := false
		for _, a := range aliases {
			if a == alias {
				removedHere = true
				continue
			}
			kept = append(kept, a)
		}

		// if alias not found here, keep block as-is
		if !removedHere {
			out = append(out, b.header)
			out = append(out, b.body...)
			continue
		}
		changed = true

		if len(kept) == 0 {
			// drop entire block
			continue
		}

		// reconstruct header without alias
		header := indent + "Host " + strings.Join(kept, " ")
		if comment != "" {
			header += " " + comment
		}
		out = append(out, header)
		out = append(out, b.body...)
	}

	return out, changed
}

// parseHostHeader returns (indent, aliases, comment, ok).
//
// - indent is leading whitespace on the original line
// - comment includes the leading '#' if present
func parseHostHeader(line string) (string, []string, string, bool) {
	if strings.TrimSpace(line) == "" {
		return "", nil, "", false
	}

	// preserve indent
	trimmedLeft := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmedLeft)]

	// split comment (preserve it)
	before, after, hasComment := strings.Cut(trimmedLeft, "#")
	comment := ""
	if hasComment {
		comment = "#" + after
	}
	before = strings.TrimSpace(before)
	if before == "" {
		return "", nil, "", false
	}
	fields := strings.Fields(before) // split on whitespace
	if len(fields) < 2 {             // need at least "Host" and one alias
		return "", nil, "", false
	}
	if strings.ToLower(fields[0]) != "host" {
		return "", nil, "", false
	}
	aliases := make([]string, 0, len(fields)-1)
	for _, a := range fields[1:] {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		aliases = append(aliases, a)
	}
	if len(aliases) == 0 {
		return "", nil, "", false
	}
	return indent, aliases, strings.TrimRight(comment, "\r\n"), true
}

// appendNonNil appends extra to lines, initializing lines if nil.
func appendNonNil(lines []string, extra ...string) []string {
	if lines == nil {
		return append([]string{}, extra...)
	}
	return append(lines, extra...)
}

// writeLines writes lines to path with LF endings.
func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return writeFileAtomic(path, []byte(content))
}

// writeFileAtomic writes data to path atomically.
//
// It creates parent directories as needed and preserves existing file permissions.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var mode os.FileMode = 0o644
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode()
	}

	f, err := os.CreateTemp(dir, ".tmp-config-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmp)
	}()

	if err := f.Chmod(mode); err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// on Windows, os.Rename won't overwrite an existing destination
	// remove first, this is not perfectly atomic but it's fine for this
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return os.Rename(tmp, path)
}
