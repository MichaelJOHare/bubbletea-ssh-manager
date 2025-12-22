package main

import (
	"bubbletea-ssh-manager/internal/connect"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const preflightTimeout = 10 * time.Second

// preflightTickCmd returns a command that waits 1 second and then sends a preflightTickMsg
// with the given token.
//
// This is used to update the preflight status periodically.
func preflightTickCmd(token int) tea.Cmd {
	return tea.Tick(1*time.Second, func(time.Time) tea.Msg {
		return preflightTickMsg{token: token}
	})
}

// preflightDialCmd returns a command that attempts to dial the given host:port
// and sends a preflightResultMsg with the result.
//
// It uses the given token to identify which preflight this result belongs to.
func preflightDialCmd(token int, hostPort string) tea.Cmd {
	return func() tea.Msg {
		err := connect.PreflightDial(hostPort, preflightTimeout)
		return preflightResultMsg{token: token, err: err}
	}
}

// cancelPreflightCmd cancels the current preflight operation.
//
// It returns the updated model, a command that sends a connectFinishedMsg
// indicating the cancellation, and true if handled.
func (m model) cancelPreflightCmd() (model, tea.Cmd, bool) {
	protocol := strings.TrimSpace(m.preflightProtocol)
	target := strings.TrimSpace(m.preflightDisplay)
	if target == "" {
		target = strings.TrimSpace(m.preflightHostPort)
	}
	m.clearPreflightState()
	return m, func() tea.Msg {
		return connectFinishedMsg{protocol: protocol, target: target, err: connect.ErrAborted}
	}, true
}

// clearPreflightState clears all stored preflight state in the model.
//
// It does not send any messages or commands.
func (m *model) clearPreflightState() {
	m.preflighting = false
	m.preflightEndsAt = time.Time{}
	m.preflightProtocol = ""
	m.preflightHostPort = ""
	m.preflightWindowTitle = ""
	m.preflightCmd = nil
	m.preflightTail = nil
	m.preflightDisplay = ""
}

// startConnect builds and starts the connection command for the given menu item.
//
// It sets the status message and returns a command to execute the connection process.
// If an error occurs while building the command, it sets an error status instead.
func (m model) startConnect(it *menuItem) (model, tea.Cmd, bool) {
	if m.preflighting {
		statusCmd := m.setStatus("Already connecting…", false, statusTTL)
		return m, statusCmd, true
	}

	if it == nil {
		m.setStatus("No host selected.", true, 0)
		return m, nil, true
	}

	cmd, tgt, tail, err := connect.BuildCommand(connect.Request{
		Protocol:    it.protocol,
		DisplayName: it.name,
		Alias:       it.alias,
		User:        it.user,
		Host:        it.hostname,
		Port:        it.port,
	})
	if err != nil {
		m.setStatus(err.Error(), true, 0)
		return m, nil, true
	}

	protocol := tgt.Protocol()
	display := tgt.Display()

	if connect.ShouldPreflight(tgt) {
		hostPort := connect.HostPortForPreflight(tgt)
		if strings.TrimSpace(hostPort) == "" {
			statusCmd := m.setStatus(fmt.Sprintf("%s: missing hostname", protocol), true, statusTTL)
			return m, statusCmd, true
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
			out := ""
			if tail != nil {
				out = strings.TrimSpace(tail.String())
				out = lastNonEmptyLine(out)
			}
			return connectFinishedMsg{protocol: protocol, target: display, err: err, output: out}
		}),
	), true
}
