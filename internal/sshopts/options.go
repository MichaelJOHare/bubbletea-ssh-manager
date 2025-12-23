package sshopts

import "strings"

// Options represents a small subset of SSH algorithm selection settings.
//
// These map to OpenSSH config keys and can be passed to ssh via `-o`.
// Values should be comma-separated algorithm lists (OpenSSH format).
// Examples: "ssh-ed25519,rsa-sha2-512".
type Options struct {
	HostKeyAlgorithms string
	KexAlgorithms     string
	MACs              string
}

// IsZero returns true if all fields are empty or whitespace.
func (o Options) IsZero() bool {
	return strings.TrimSpace(o.HostKeyAlgorithms) == "" &&
		strings.TrimSpace(o.KexAlgorithms) == "" &&
		strings.TrimSpace(o.MACs) == ""
}

// DisplayString returns a human-readable summary for UI display.
//
// Example output:
//   - HostKeyAlgorithms=ssh-ed25519,rsa-sha2-512
//   - KexAlgorithms=curve25519-sha256
//   - MACs=hmac-sha2-256
//
// Note: this is intentionally not used to construct the ssh command line.
// This app connects by alias and relies on OpenSSH to apply ~/.ssh/config.
func (o Options) DisplayString() string {
	parts := make([]string, 0, 3)
	if v := strings.TrimSpace(o.HostKeyAlgorithms); v != "" {
		parts = append(parts, "HostKeyAlgorithms="+v)
	}
	if v := strings.TrimSpace(o.KexAlgorithms); v != "" {
		parts = append(parts, "KexAlgorithms="+v)
	}
	if v := strings.TrimSpace(o.MACs); v != "" {
		parts = append(parts, "MACs="+v)
	}
	return strings.Join(parts, "\n")
}
