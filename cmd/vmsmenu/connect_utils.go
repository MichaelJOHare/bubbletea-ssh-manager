package main

import (
	"strings"

	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const connectionAbortedExitStatus = "exit status 512"

type connectTarget struct {
	protocol string // connection protocol (ssh, telnet, etc)
	alias    string // ssh alias / display alias (exactly as in config)
	user     string // username when set
	host     string // hostname when set
	port     string // normalized numeric port when host is set (else may be empty)
}

// Protocol returns the connection protocol (ssh, telnet, etc).
func (t connectTarget) Protocol() string {
	return t.protocol
}

// Display returns the human-readable target for status messages.
// Examples:
//   - ssh:   mike@krabby <10.0.0.147:22>
//   - ssh:   krabby
//   - telnet: router <router:23>
func (t connectTarget) Display() string {
	alias := strings.TrimSpace(t.alias)
	user := strings.TrimSpace(t.user)
	host := strings.TrimSpace(t.host)
	port := strings.TrimSpace(t.port)

	displayAlias := alias
	if user != "" {
		displayAlias = user + "@" + alias
	}
	if host != "" && port != "" {
		return displayAlias + " <" + host + ":" + port + ">"
	}
	return displayAlias
}

// preferredProgramPath returns the preferred full path to the given program name.
//
// On Windows, it prefers MSYS2 binaries if available.
// On other platforms, it looks in the system PATH.
func preferredProgramPath(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("empty program name")
	}

	// prefer MSYS2 binaries when running on Windows
	if runtime.GOOS == "windows" {
		roots := []string{}
		if v := strings.TrimSpace(os.Getenv("BTSM_MSYS_ROOT")); v != "" {
			roots = append(roots, v)
		}
		// default install location
		roots = append(roots, `C:\msys64`)
		for _, root := range roots {
			p := filepath.Join(root, "usr", "bin", name+".exe")
			// check if the file exists and is not a directory
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}

	// fallback to PATH lookup
	p, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return p, nil
}

// isConnectionAborted returns true if the given error indicates
// that the connection was aborted by user (eg. Ctrl+C).
func isConnectionAborted(err error) bool {
	if err == nil {
		return false
	}

	// early exit status match
	s := strings.TrimSpace(err.Error())
	if s == connectionAbortedExitStatus {
		return true
	}
	// catch cases where additional info is included
	if strings.Contains(s, connectionAbortedExitStatus) {
		return true
	}
	// common alternative representations
	ls := strings.ToLower(s)
	if strings.Contains(ls, "signal: interrupt") {
		return true
	}

	return false
}

// WindowTitle returns a stable short title for the terminal/tab.
// Format: "ssh mike@KRABBY" (or group.host with host uppercased).
func (t connectTarget) WindowTitle() string {
	protocol := strings.TrimSpace(t.protocol)
	alias := strings.TrimSpace(t.alias)
	user := strings.TrimSpace(t.user)

	if alias == "" {
		return protocol
	}

	// Uppercase only the host portion of grouped aliases: group.HOST
	aliasTitle := strings.ToUpper(alias)
	if g, h, ok := splitGroupedAlias(alias); ok {
		aliasTitle = g + "." + strings.ToUpper(h)
	}

	if user != "" {
		return protocol + " " + user + "@" + aliasTitle
	}
	return protocol + " " + aliasTitle
}
