package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type tailBuffer struct {
	buf []byte
	max int
}

func newTailBuffer(max int) *tailBuffer {
	if max <= 0 {
		max = 4096
	}
	return &tailBuffer{max: max}
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(p) >= t.max {
		t.buf = append(t.buf[:0], p[len(p)-t.max:]...)
		return len(p), nil
	}

	need := len(t.buf) + len(p) - t.max
	if need > 0 {
		t.buf = t.buf[need:]
	}
	t.buf = append(t.buf, p...)
	return len(p), nil
}

func (t *tailBuffer) String() string {
	return string(t.buf)
}

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
		return nil, "", "", nil, fmt.Errorf("empty host alias")
	}

	cmd := exec.Command(protocol, target)
	cmd.Stdin = os.Stdin

	// keep streaming to the real terminal (interactive), but also capture the last
	// bit so we can surface errors in the TUI after returning
	tail := newTailBuffer(4096)
	cmd.Stdout = io.MultiWriter(os.Stdout, tail)
	cmd.Stderr = io.MultiWriter(os.Stderr, tail)

	return cmd, protocol, target, tail, nil
}

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
