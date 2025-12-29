package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"bubbletea-ssh-manager/internal/connect"
	str "bubbletea-ssh-manager/internal/stringutil"

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
	protocol := strings.TrimSpace(m.ms.preflightProtocol)
	target := strings.TrimSpace(m.ms.preflightDisplay)
	if target == "" {
		target = strings.TrimSpace(m.ms.preflightHostPort)
	}
	m.clearPreflightState()
	return m, func() tea.Msg {
		return connectFinishedMsg{protocol: protocol, target: target, err: connect.ErrAborted}
	}, true
}

// clearPreflightState clears all stored preflight state in the model.
func (m *model) clearPreflightState() {
	m.mode = modeMenu
	m.ms.preflightEndsAt = time.Time{}
	m.ms.preflightRemaining = 0
	m.ms.preflightProtocol = ""
	m.ms.preflightHostPort = ""
	m.ms.preflightWindowTitle = ""
	m.ms.preflightCmd = nil
	m.ms.preflightTail = nil
	m.ms.preflightDisplay = ""
}

// startConnect builds and starts the connection command for the given menu item.
//
// It sets the status message and returns a command to execute the connection process.
// If an error occurs while building the command, it sets an error status instead.
func (m model) startConnect(it *menuItem) (model, tea.Cmd, bool) {
	// prevent multiple simultaneous connections
	if m.mode == modePreflight || m.mode == modeExecuting {
		statusCmd := m.setStatus("Already connecting…", false, statusTTL)
		return m, statusCmd, true
	}

	// sanity check
	if it == nil {
		m.setStatus("No host selected.", true, 0)
		return m, nil, true
	}

	// build the connection command
	cmd, tgt, tail, err := connect.BuildCommand(connect.Target{
		Protocol: it.protocol,
		Spec:     it.spec,
	})
	if err != nil {
		m.setStatus(err.Error(), true, 0)
		return m, nil, true
	}

	protocol := str.NormalizeString(tgt.Protocol)
	display := tgt.Display()

	// check if we need to preflight
	if connect.ShouldPreflight(tgt) {
		hostPort := connect.GenerateHostPort(tgt)
		if strings.TrimSpace(hostPort) == "" {
			statusCmd := m.setStatus(fmt.Sprintf("%s: missing hostname", protocol), true, statusTTL)
			return m, statusCmd, true
		}

		m.mode = modePreflight
		m.ms.preflightToken++
		tok := m.ms.preflightToken
		m.ms.preflightRemaining = int(preflightTimeout.Seconds())
		m.ms.preflightEndsAt = time.Now().Add(preflightTimeout)
		m.ms.preflightProtocol = protocol
		m.ms.preflightHostPort = hostPort
		m.ms.preflightWindowTitle = tgt.WindowTitle()
		m.ms.preflightCmd = cmd
		m.ms.preflightTail = tail
		m.ms.preflightDisplay = display

		// clear any existing status to make room for preflight status
		m.status = ""
		m.statusIsError = false

		m.relayout()

		return m, tea.Batch(preflightDialCmd(tok, hostPort), preflightTickCmd(tok), m.spinner.Tick), true
	}

	// no preflight needed; start connection immediately
	m.mode = modeExecuting
	return m, launchExecCmd(tgt.WindowTitle(), cmd, protocol, display, tail), true
}

// launchExecCmd returns a command that exits the TUI and starts
// the given exec.Cmd in the main terminal.
//
// tea.ExitAltScreen is used to to make every connection login
// session starts fresh in the main terminal, avoiding issues
// with leftover TUI artifacts.
//
// It sets the window title before starting the command, and sends a
// connectFinishedMsg when the command exits, capturing any output from
// the provided TailBuffer for error reporting.
func launchExecCmd(windowTitle string, cmd *exec.Cmd, protocol string, target string, tail *connect.TailBuffer) tea.Cmd {
	return tea.Sequence(
		tea.ExitAltScreen,
		tea.SetWindowTitle(windowTitle),
		tea.ExecProcess(cmd, func(err error) tea.Msg {
			out := ""
			if tail != nil {
				out = strings.TrimSpace(tail.String())
				out = str.LastNonEmptyLine(out)
			}
			return connectFinishedMsg{protocol: protocol, target: target, err: err, output: out}
		}),
	)
}
