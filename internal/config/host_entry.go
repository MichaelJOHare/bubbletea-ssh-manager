package config

import (
	"fmt"
	"strings"

	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"
)

// HostEntry is a minimal representation of a Host block from an SSH-style config.
// It intentionally contains only the fields this project currently supports.
type HostEntry struct {
	Spec       host.Spec       // host fields that are shared between SSH and Telnet (alias/hostname/port/user)
	SSHOptions sshopts.Options // SSH-specific options for this host
	SourcePath string          // path to the config file this entry was read from
}

// EntryFromSpec creates a HostEntry from the given spec and options.
func EntryFromSpec(spec host.Spec, opts sshopts.Options, sourcePath string) HostEntry {
	return HostEntry{
		Spec:       spec,
		SSHOptions: opts,
		SourcePath: sourcePath,
	}
}

// buildHostEntry creates config lines for a HostEntry.
//
// It centralizes spacing rules:
//   - If preceding is non-empty and does not already end with a blank line, it
//     adds exactly one blank line before the Host block.
//   - It always adds a trailing blank line.
func buildHostEntry(entry HostEntry, preceding []string) []string {
	alias := strings.TrimSpace(entry.Spec.Alias)
	indent := DefaultHostIndent

	out := make([]string, 0, 8)
	if len(preceding) > 0 && preceding[len(preceding)-1] != "" {
		out = append(out, "")
	}
	out = append(out, fmt.Sprintf("Host %s", alias))
	if v := strings.TrimSpace(entry.Spec.HostName); v != "" {
		out = append(out, indent+"HostName "+v)
	}
	if v := strings.TrimSpace(entry.Spec.User); v != "" {
		out = append(out, indent+"User "+v)
	}
	if v := strings.TrimSpace(entry.Spec.Port); v != "" {
		out = append(out, indent+"Port "+v)
	}
	sshOpts := BuildSSHOptions(entry.SSHOptions, indent)
	if len(sshOpts) > 0 {
		out = append(out, sshOpts...)
	}
	// trailing blank line for readability
	out = append(out, "")
	return out
}

// BuildSSHOptions creates a formatted output for non-empty SSH options.
//
// It uses the given indent for each line.
func BuildSSHOptions(o sshopts.Options, indent string) []string {
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

// setHostDirective applies a single Host block directive to the given HostEntry.
//
// It returns true if the directive was recognized and applied.
func setHostDirective(key string, value string, entry *HostEntry) bool {
	if entry == nil {
		return false
	}

	// standard host directives shared between SSH and Telnet
	switch key {
	case "hostname":
		entry.Spec.HostName = value
		return true
	case "port":
		entry.Spec.Port = value
		return true
	case "user":
		entry.Spec.User = value
		return true

	// SSH options (a small subset), values are usually comma separated
	case "hostkeyalgorithms":
		entry.SSHOptions.HostKeyAlgorithms = value
		return true
	case "kexalgorithms":
		entry.SSHOptions.KexAlgorithms = value
		return true
	case "macs":
		entry.SSHOptions.MACs = value
		return true
	}

	return false
}
