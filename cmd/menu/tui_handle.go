package main

import (
	"fmt"
	"strings"
	"time"

	"bubbletea-ssh-manager/internal/connect"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// handleWindowSizeMsg handles window resize messages.
//
// It updates the model's width and height and relayouts the components.
func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (model, tea.Cmd, bool) {
	m.width, m.height = msg.Width, msg.Height
	m.relayout()
	return m, nil, true
}

// handleHostFormMsg handles all host-form related messages, including:
//   - form result messages (submitted/canceled)
//   - async save results
//   - generic non-key messages while the form is open (cursor blink, etc)
//
// Key messages are handled by handleKeyMsg so it can intercept esc/left.
func (m model) handleHostFormMsg(msg tea.Msg) (model, tea.Cmd, bool) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, nil, false
	}

	switch v := msg.(type) {
	case formResultMsg:
		switch v.kind {
		case formResultCancelled:
			nm, cmd := m.closeHostForm("Canceled.", statusError)
			return nm, cmd, true
		case formResultSubmitted:
			nm, cmd := m.handleHostFormSubmit(v)
			return nm, cmd, true
		default:
			return m, nil, true
		}

	case formSaveResultMsg:
		nm, cmd := m.handleHostFormSaveResult(v)
		return nm, cmd, true
	}

	if m.mode != modeHostForm {
		return m, nil, false
	}
	if m.ms.hostForm == nil {
		return m, nil, true
	}
	mdl, cmd := m.ms.hostForm.Update(msg)
	if f, ok := mdl.(*huh.Form); ok {
		m.ms.hostForm = f
	}
	m.relayout()
	return m, cmd, true
}

// handleMenuReloadedMsg handles menu reloaded messages.
//
// It applies the reloaded menu to the model and returns the updated model
// and any command resulting from applying the new menu.
func (m model) handleMenuReloadedMsg(msg menuReloadedMsg) (model, tea.Cmd, bool) {
	if msg.root == nil {
		return m, m.setStatusError("Failed to reload menu.", statusTTL), true
	}
	m.root = msg.root
	m.path = []*menuItem{msg.root}
	m.query.SetValue("")
	m.setCurrentMenu(msg.root.children)
	m.relayout()
	if msg.err != nil {
		return m, m.setStatusError("Config: "+msg.err.Error(), statusTTL), true
	}
	return m, nil, true
}

// handleSpinnerTickMsg handles spinner tick messages.
//
// It updates the spinner component if the UI is in preflight mode.
func (m model) handleSpinnerTickMsg(msg spinner.TickMsg) (model, tea.Cmd, bool) {
	if m.mode != modePreflight {
		return m, nil, true
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd, true
}

// handleStatusClearMsg handles status clear messages.
//
// It clears the status message if the token matches the current status token.
func (m model) handleStatusClearMsg(msg statusClearMsg) (model, tea.Cmd, bool) {
	if msg.token == m.statusToken {
		m.status = ""
		m.statusKind = statusInfo
		m.relayout()
	}
	return m, nil, true
}

// handlePreflightTickMsg handles preflight tick messages.
//
// It updates the remaining time for the preflight operation and
// returns a command to schedule the next tick if needed.
func (m model) handlePreflightTickMsg(msg preflightTickMsg) (model, tea.Cmd, bool) {
	if m.mode != modePreflight || msg.token != m.ms.preflightToken {
		return m, nil, true
	}
	remaining := max(int(time.Until(m.ms.preflightEndsAt).Round(time.Second).Seconds()), 0)
	m.ms.preflightRemaining = remaining
	if remaining > 0 {
		return m, preflightTickCmd(msg.token), true
	}
	return m, nil, true
}

// handlePreflightResultMsg handles preflight result messages.
//
// It processes the result of the preflight check and either starts
// the connection or shows an error status.
func (m model) handlePreflightResultMsg(msg preflightResultMsg) (model, tea.Cmd, bool) {
	if m.mode != modePreflight || msg.token != m.ms.preflightToken {
		return m, nil, true
	}

	protocol := m.ms.preflightProtocol
	hostPort := m.ms.preflightHostPort
	display := m.ms.preflightDisplay
	windowTitle := m.ms.preflightWindowTitle
	cmd := m.ms.preflightCmd
	tail := m.ms.preflightTail

	m.clearPreflightState()

	if msg.err != nil {
		statusCmd := m.setStatusError(fmt.Sprintf("%s %s failed: \n%v", protocol, hostPort, msg.err), statusTTL)
		return m, statusCmd, true
	}

	m.mode = modeExecuting
	return m, launchExecCmd(windowTitle, cmd, protocol, display, tail), true
}

// handleConnectFinishedMsg handles connection finished messages.
//
// It resets the UI to menu mode and sets an appropriate status message
// based on whether the connection succeeded or failed.
func (m model) handleConnectFinishedMsg(msg connectFinishedMsg) (model, tea.Cmd, bool) {
	m.mode = modeMenu
	titleCmd := tea.SetWindowTitle("MENU")
	output := strings.TrimSpace(msg.output)
	if msg.err != nil {
		if output != "" {
			statusCmd := m.setStatusError(fmt.Sprintf("%s to %s exited:\n%s (%v)", msg.protocol, msg.target, output, msg.err), 0)
			return m, tea.Batch(titleCmd, statusCmd), true
		}
		if connect.IsConnectionAborted(msg.err) {
			statusCmd := m.setStatusError(fmt.Sprintf("%s to %s aborted.", msg.protocol, msg.target), statusTTL)
			return m, tea.Batch(titleCmd, statusCmd), true
		}
		statusCmd := m.setStatusError(fmt.Sprintf("%s to %s exited:\n%v", msg.protocol, msg.target, msg.err), 0)
		return m, tea.Batch(titleCmd, statusCmd), true
	}

	statusCmd := m.setStatusSuccess(fmt.Sprintf("%s to %s ended.", msg.protocol, msg.target), statusTTL)
	return m, tea.Batch(titleCmd, statusCmd), true
}
