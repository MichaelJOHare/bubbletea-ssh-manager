package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// fileExists checks if the given path exists and is a file.
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	if st, err := os.Stat(path); err == nil {
		return !st.IsDir()
	}
	return false
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
			if fileExists(p) {
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

// newTailBuffer creates a new tailBuffer that keeps up to max bytes.
// Used to capture the tail of command output.
func newTailBuffer(max int) *tailBuffer {
	if max <= 0 {
		max = 4096
	}
	return &tailBuffer{max: max}
}

// Write implements io.Writer for tailBuffer.
func (t *tailBuffer) Write(p []byte) (int, error) {
	// nothing to do
	if len(p) == 0 {
		return 0, nil
	}

	// too much data, just keep the tail
	if len(p) >= t.max {
		t.buf = append(t.buf[:0], p[len(p)-t.max:]...)
		return len(p), nil
	}

	// append and trim from the front if needed
	need := len(t.buf) + len(p) - t.max
	if need > 0 {
		t.buf = t.buf[need:]
	}
	t.buf = append(t.buf, p...)
	return len(p), nil
}

// String returns the contents of the buffer as a Go string.
func (t *tailBuffer) String() string {
	return string(t.buf)
}

// lastNonEmptyLine returns the last non-empty line from the given string.
// Used to extract error messages from command output.
func lastNonEmptyLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

// buildConnectCommand builds the exec.Cmd to connect to the given host menu item.
// It also returns the protocol and target for status messages, and a tailBuffer
// that captures the last part of the command output for error reporting.
func buildConnectCommand(it *menuItem) (*exec.Cmd, string, string, *tailBuffer, error) {
	protocol := strings.TrimSpace(strings.ToLower(it.protocol))
	if protocol != "ssh" && protocol != "telnet" {
		return nil, "", "", nil, fmt.Errorf("unknown protocol for %s: %q", it.name, it.protocol)
	}

	programPath, err := preferredProgramPath(protocol)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("%s not found: %w", protocol, err)
	}

	target := strings.TrimSpace(it.target)
	if target == "" {
		return nil, "", "", nil, fmt.Errorf("empty target")
	}

	args := []string{target}
	if protocol == "telnet" {
		if host, port, ok := splitHostPort(target); ok {
			args = []string{host, port}
		} else {
			fields := strings.Fields(target)
			if len(fields) >= 2 {
				args = []string{fields[0], fields[1]}
			} else {
				args = []string{target}
			}
		}
	}

	cmd := exec.Command(programPath, args...)
	cmd.Stdin = os.Stdin

	// keep streaming to the real terminal (interactive), but also capture the last
	// bit so we can surface errors in the TUI after returning
	tail := newTailBuffer(4096)
	cmd.Stdout = io.MultiWriter(os.Stdout, tail)
	cmd.Stderr = io.MultiWriter(os.Stderr, tail)

	return cmd, protocol, target, tail, nil
}
