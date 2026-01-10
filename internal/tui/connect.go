package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"bubbletea-ssh-manager/internal/config"
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
func (m model) cancelPreflightCmd() (model, tea.Cmd) {
	protocol := m.ms.preflight.protocol
	target := m.ms.preflight.display
	if target == "" {
		target = m.ms.preflight.hostPort
	}
	m.clearPreflightState()
	return m, func() tea.Msg {
		return connectFinishedMsg{protocol: protocol, target: target, err: connect.ErrAborted}
	}
}

// initPreflightState sets all preflight state fields.
//
// Pass zero values to clear the state. The token is always incremented.
func (m *model) initPreflightState(protocol config.Protocol, hostPort, windowTitle, display string,
	cmd *exec.Cmd, tail *connect.TailBuffer) int {

	m.ms.preflight.token++
	m.ms.preflight.protocol = protocol
	m.ms.preflight.hostPort = hostPort
	m.ms.preflight.windowTitle = windowTitle
	m.ms.preflight.display = display
	m.ms.preflight.cmd = cmd
	m.ms.preflight.tail = tail

	if hostPort != "" {
		m.ms.preflight.remaining = int(preflightTimeout.Seconds())
		m.ms.preflight.endsAt = time.Now().Add(preflightTimeout)
	} else {
		m.ms.preflight.remaining = 0
		m.ms.preflight.endsAt = time.Time{}
	}

	return m.ms.preflight.token
}

// clearPreflightState clears all stored preflight state in the model.
func (m *model) clearPreflightState() {
	m.mode = modeMenu
	m.initPreflightState("", "", "", "", nil, nil)
}

// startConnect builds and starts the connection command for the given menu item.
//
// It sets the status message and returns a command to execute the connection process.
// If an error occurs while building the command, it sets an error status instead.
func (m model) startConnect(it *menuItem) (model, tea.Cmd) {
	if m.mode == modePreflight || m.mode == modeExecuting {
		return m, m.setStatusInfo("Already connecting…", statusTTL)
	}
	if it == nil {
		return m, m.setStatusError("No host selected.", 0)
	}

	cmd, tgt, tail, err := connect.BuildCommand(connect.Target{
		Protocol: it.protocol,
		Spec:     it.spec,
	})
	if err != nil {
		return m, m.setStatusError(err.Error(), 0)
	}

	protocol := tgt.Protocol
	display := tgt.Display()

	// no preflight needed; start connection immediately
	if !connect.ShouldPreflight(tgt) {
		m.mode = modeExecuting
		return m, launchExecCmd(tgt.WindowTitle(), cmd, protocol, display, tail)
	}

	// preflight required
	hostPort := connect.GenerateHostPort(tgt)
	if hostPort == "" {
		return m, m.setStatusError(fmt.Sprintf("%s: missing hostname", string(protocol)), statusTTL)
	}

	m.mode = modePreflight
	tok := m.initPreflightState(protocol, hostPort, tgt.WindowTitle(), display, cmd, tail)
	m.setStatusInfo("", 0)

	return m, tea.Batch(preflightDialCmd(tok, hostPort), preflightTickCmd(tok), m.spinner.Tick)
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
func launchExecCmd(windowTitle string, cmd *exec.Cmd, protocol config.Protocol, target string, tail *connect.TailBuffer) tea.Cmd {
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
