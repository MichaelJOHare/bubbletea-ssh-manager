package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// tailBuffer is an io.Writer that keeps only the last N bytes written to it.
type tailBuffer struct {
	buf []byte // stored bytes
	max int    // max bytes to keep
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

	if _, err := exec.LookPath(protocol); err != nil {
		return nil, "", "", nil, fmt.Errorf("%s not found on PATH: %w", protocol, err)
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

	cmd := exec.Command(protocol, args...)
	cmd.Stdin = os.Stdin

	// keep streaming to the real terminal (interactive), but also capture the last
	// bit so we can surface errors in the TUI after returning
	tail := newTailBuffer(4096)
	cmd.Stdout = io.MultiWriter(os.Stdout, tail)
	cmd.Stderr = io.MultiWriter(os.Stderr, tail)

	return cmd, protocol, target, tail, nil
}
