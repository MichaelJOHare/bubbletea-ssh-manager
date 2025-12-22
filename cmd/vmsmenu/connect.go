package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const preflightTimeout = 10 * time.Second

// shouldPreflight returns true if the given connectTarget requires preflight checks.
//
// Telnet always requires preflight (to check host/port).
// SSH requires preflight if a hostname is set (to check host/port).
func shouldPreflight(tgt connectTarget) bool {
	switch strings.TrimSpace(tgt.protocol) {
	case "telnet":
		return true
	case "ssh":
		return strings.TrimSpace(tgt.host) != ""
	default:
		return false
	}
}

// preflightTickCmd returns a command that waits 1 second and then sends a preflightTickMsg
// with the given token.
func preflightTickCmd(token int) tea.Cmd {
	return tea.Tick(1*time.Second, func(time.Time) tea.Msg {
		return preflightTickMsg{token: token}
	})
}

// preflightDialCmd returns a command that attempts to dial the given host:port
// and sends a preflightResultMsg with the result.
func preflightDialCmd(token int, hostPort string) tea.Cmd {
	return func() tea.Msg {
		d := net.Dialer{Timeout: preflightTimeout}
		c, err := d.Dial("tcp", hostPort)
		if c != nil {
			_ = c.Close()
		}
		return preflightResultMsg{token: token, err: err}
	}
}

// buildConnectCommand builds the exec.Cmd to connect to the given host menu item.
//
// It returns a connectTarget for display/title, and a tailBuffer
// that captures the last part of the command output for error reporting.
func buildConnectCommand(it *menuItem) (cmd *exec.Cmd, tgt connectTarget, tail *tailBuffer, err error) {
	if it == nil {
		err = fmt.Errorf("nil menu item")
		return
	}

	// determine protocol
	tgt.protocol = normalizeString(it.protocol)
	if tgt.protocol != "ssh" && tgt.protocol != "telnet" {
		err = fmt.Errorf("unknown protocol for %s: %q", it.name, it.protocol)
		return
	}

	// find program path for protocol
	var programPath string
	programPath, err = preferredProgramPath(tgt.protocol)
	if err != nil {
		err = fmt.Errorf("%s not found: %w", tgt.protocol, err)
		return
	}

	// prepare connection parameters
	tgt.alias = strings.TrimSpace(it.alias)
	tgt.user = strings.TrimSpace(it.user)
	tgt.host = strings.TrimSpace(it.hostname)
	port := strings.TrimSpace(it.port)

	if tgt.alias == "" {
		err = fmt.Errorf("empty %s alias", tgt.protocol)
		return
	}

	// build command based on protocol
	var args []string
	switch tgt.protocol {
	case "ssh":
		// ssh connects by alias, hostname/port are only for display
		if tgt.host != "" {
			var p string
			p, err = normalizePort(port, "ssh")
			if err != nil {
				return
			}
			tgt.port = p
		}
		if tgt.user != "" {
			args = []string{"-l", tgt.user, tgt.alias}
		} else {
			args = []string{tgt.alias}
		}

	case "telnet":
		// telnet connects by hostname and port
		if tgt.host == "" {
			err = fmt.Errorf("telnet %q: empty hostname", tgt.alias)
			return
		}
		var p string
		p, err = normalizePort(port, "telnet")
		if err != nil {
			return
		}
		tgt.port = p
		args = []string{tgt.host, p}
	}

	// build exec.Cmd
	cmd = exec.Command(programPath, args...)
	cmd.Stdin = os.Stdin

	// keep streaming to the real terminal (interactive), but also capture the last
	// bit so we can surface errors in the TUI after returning
	tail = newTailBuffer(4096)
	cmd.Stdout = io.MultiWriter(os.Stdout, tail)
	cmd.Stderr = io.MultiWriter(os.Stderr, tail)

	return
}

// startConnect builds and starts the connection command for the given menu item.
//
// It sets the status message and returns a command to execute the connection process.
// If an error occurs while building the command, it sets an error status instead.
func (m model) startConnect(it *menuItem) (model, tea.Cmd, bool) {
	if m.preflighting {
		statusCmd := m.setStatus("Already connecting…", false, infoStatusTTL)
		return m, statusCmd, true
	}

	// build connection command using menu item
	cmd, tgt, tail, err := buildConnectCommand(it)
	if err != nil {
		m.setStatus(err.Error(), true, 0)
		return m, nil, true
	}

	protocol := tgt.Protocol()
	display := tgt.Display()

	if shouldPreflight(tgt) {
		host := strings.TrimSpace(tgt.host)
		port := strings.TrimSpace(tgt.port)
		if host == "" {
			statusCmd := m.setStatus(fmt.Sprintf("%s: missing hostname", protocol), true, infoStatusTTL)
			return m, statusCmd, true
		}
		hostPort := host
		if port != "" {
			hostPort = net.JoinHostPort(host, port)
		}

		m.preflighting = true
		m.preflightToken++
		tok := m.preflightToken
		m.preflightEndsAt = time.Now().Add(preflightTimeout)
		m.preflightProtocol = protocol
		m.preflightHostPort = hostPort
		m.preflightWindowTitle = tgt.WindowTitle()
		m.preflightCmd = cmd
		m.preflightTail = tail
		m.preflightDisplay = display

		m.setStatus(fmt.Sprintf("Checking %s %s (10s)…", protocol, hostPort), false, 0)
		return m, tea.Batch(preflightDialCmd(tok, hostPort), preflightTickCmd(tok)), true
	}

	m.setStatus(fmt.Sprintf("Starting %s %s…", protocol, display), false, 0)

	return m, tea.Sequence(
		tea.SetWindowTitle(tgt.WindowTitle()),
		tea.ExecProcess(cmd, func(err error) tea.Msg {
			out := strings.TrimSpace(tail.String())
			out = lastNonEmptyLine(out)
			return connectFinishedMsg{protocol: protocol, target: display, err: err, output: out}
		}),
	), true
}
