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
