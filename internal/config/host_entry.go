package config

import (
	"fmt"
	"strings"
)

const (
	ProtocolSSH    Protocol = "ssh"
	ProtocolTelnet Protocol = "telnet"
)

type Protocol string // "ssh" or "telnet"

// HostEntry is a minimal representation of a Host block from an SSH-style config.
// It intentionally contains only the fields this project currently supports.
type HostEntry struct {
	Spec       Spec       // host fields that are shared between SSH and Telnet (alias/hostname/port/user)
	SSHOptions SSHOptions // SSH-specific options for this host
	SourcePath string     // path to the config file this entry was read from
}

// Spec is the shared representation of a host endpoint across the project.
//
// It maps directly to the subset of SSH-style config directives we support
// for both SSH and Telnet hosts.
type Spec struct {
	Alias    string // ssh-style Host alias from the config
	HostName string // hostname or IP address
	Port     string // port number as string
	User     string // user name
}

// SSHOptions represents a small subset of SSH algorithm selection settings.
//
// These map to OpenSSH config keys and can be passed to ssh via `-o`.
// Values should be comma-separated algorithm lists (OpenSSH format).
type SSHOptions struct {
	HostKeyAlgorithms string // HostKeyAlgorithms option (e.g. "ssh-rsa,ssh-ed25519")
	KexAlgorithms     string // KexAlgorithms option (e.g. "curve25519-sha256,ecdh-sha2-nistp256")
	MACs              string // MACs option (e.g. "hmac-sha2-256,hmac-sha1")
}

// Normalized returns a copy of the spec with leading/trailing whitespace removed.
//
// This is intended to be applied at boundaries (parsing user input / reading config)
// so most internal code can assume specs are already trimmed.
func (s Spec) Normalized() Spec {
	s.Alias = strings.TrimSpace(s.Alias)
	s.HostName = strings.TrimSpace(s.HostName)
	s.Port = strings.TrimSpace(s.Port)
	s.User = strings.TrimSpace(s.User)
	return s
}

// Normalized returns a copy of the SSH options with leading/trailing whitespace removed.
func (o SSHOptions) Normalized() SSHOptions {
	o.HostKeyAlgorithms = strings.TrimSpace(o.HostKeyAlgorithms)
	o.KexAlgorithms = strings.TrimSpace(o.KexAlgorithms)
	o.MACs = strings.TrimSpace(o.MACs)
	return o
}

// Normalized returns a copy of the host entry with normalized Spec/SSHOptions.
func (e HostEntry) Normalized() HostEntry {
	e.Spec = e.Spec.Normalized()
	e.SSHOptions = e.SSHOptions.Normalized()
	return e
}

// EntryFromSpec creates a HostEntry from the given spec and options.
func EntryFromSpec(spec Spec, opts SSHOptions, sourcePath string) HostEntry {
	return HostEntry{
		Spec:       spec.Normalized(),
		SSHOptions: opts.Normalized(),
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
	entry = entry.Normalized()
	alias := entry.Spec.Alias
	indent := DefaultHostIndent

	out := make([]string, 0, 8)
	if len(preceding) > 0 && preceding[len(preceding)-1] != "" {
		out = append(out, "")
	}
	out = append(out, fmt.Sprintf("Host %s", alias))
	if v := entry.Spec.HostName; v != "" {
		out = append(out, indent+"HostName "+v)
	}
	if v := entry.Spec.User; v != "" {
		out = append(out, indent+"User "+v)
	}
	if v := entry.Spec.Port; v != "" {
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
func BuildSSHOptions(o SSHOptions, indent string) []string {
	o = o.Normalized()
	parts := make([]string, 0, 3)
	if v := o.HostKeyAlgorithms; v != "" {
		parts = append(parts, indent+"HostKeyAlgorithms "+v)
	}
	if v := o.KexAlgorithms; v != "" {
		parts = append(parts, indent+"KexAlgorithms "+v)
	}
	if v := o.MACs; v != "" {
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

	// treat parser input as boundary data, normalize it once here
	value = strings.TrimSpace(value)

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
