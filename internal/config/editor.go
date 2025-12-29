package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultHostIndent is the default indentation used for options in a Host block.
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
	out = append(out, buildHostEntry(entry)...)
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
	out = append(out, buildHostEntry(updated)...)
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
