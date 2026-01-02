package main

import (
	"fmt"
	"strings"
	"time"

	"bubbletea-ssh-manager/internal/connect"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// handleWindowSizeMsg handles window resize messages.
//
// It updates the model's width and height and relayouts the components.
func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (model, tea.Cmd) {
	m.width, m.height = msg.Width, msg.Height
	m.relayout()
	return m, nil
}

// handleRemoveHostResult processes the async result of a host removal operation.
//
// It updates the model's status and reloads the menu if the removal succeeded.
func (m model) handleRemoveHostResult(msg removeHostResultMsg) (model, tea.Cmd) {
	if msg.err != nil {
		statusCmd := m.setStatusError(fmt.Sprintf("Failed to remove %s: %v", msg.alias, msg.err), 0)
		return m, statusCmd
	}

	// reload menu to reflect the removal
	statusCmd := m.setStatusSuccess(fmt.Sprintf("Removed %s host: %s"+successCheck, msg.protocol, msg.alias), statusTTL)
	reloadCmd := func() tea.Msg {
		root, err := seedMenu()
		return menuReloadedMsg{root: root, err: err}
	}
	return m, tea.Batch(statusCmd, reloadCmd)
}

// handleMenuReloadedMsg handles menu reloaded messages.
//
// It applies the reloaded menu to the model and returns the updated model
// and any command resulting from applying the new menu.
func (m model) handleMenuReloadedMsg(msg menuReloadedMsg) (model, tea.Cmd) {
	if msg.root == nil {
		return m, m.setStatusError("Failed to reload menu.", statusTTL)
	}
	m.root = msg.root
	m.path = []*menuItem{msg.root}
	m.query.SetValue("")
	m.setCurrentMenu(msg.root.children)
	m.relayout()
	if msg.err != nil {
		return m, m.setStatusError("Config: "+msg.err.Error(), statusTTL)
	}
	return m, nil
}

// handleStatusClearMsg handles status clear messages.
//
// It clears the status message if the token matches the current status token.
func (m model) handleStatusClearMsg(msg statusClearMsg) (model, tea.Cmd) {
	if msg.token == m.statusToken {
		m.status = ""
		m.statusKind = statusInfo
		m.relayout()
	}
	return m, nil
}

// handlePreflightTickMsg handles preflight tick messages.
//
// It updates the remaining time for the preflight operation and
// returns a command to schedule the next tick if needed.
func (m model) handlePreflightTickMsg(msg preflightTickMsg) (model, tea.Cmd) {
	if m.mode != modePreflight || msg.token != m.ms.preflightToken {
		return m, nil
	}
	remaining := max(int(time.Until(m.ms.preflightEndsAt).Round(time.Second).Seconds()), 0)
	m.ms.preflightRemaining = remaining
	if remaining > 0 {
		return m, preflightTickCmd(msg.token)
	}
	return m, nil
}

// handleSpinnerTickMsg handles spinner tick messages.
//
// It updates the spinner component if the UI is in preflight mode.
func (m model) handleSpinnerTickMsg(msg spinner.TickMsg) (model, tea.Cmd) {
	if m.mode != modePreflight {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// handlePreflightResultMsg handles preflight result messages.
//
// It processes the result of the preflight check and either starts
// the connection or shows an error status.
func (m model) handlePreflightResultMsg(msg preflightResultMsg) (model, tea.Cmd) {
	if m.mode != modePreflight || msg.token != m.ms.preflightToken {
		return m, nil
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
		return m, statusCmd
	}

	m.mode = modeExecuting
	return m, launchExecCmd(windowTitle, cmd, protocol, display, tail)
}

// handleConnectFinishedMsg handles connection finished messages.
//
// It resets the UI to menu mode and sets an appropriate status message
// based on whether the connection succeeded or failed.
func (m model) handleConnectFinishedMsg(msg connectFinishedMsg) (model, tea.Cmd) {
	m.mode = modeMenu
	titleCmd := tea.SetWindowTitle("MENU")
	output := strings.TrimSpace(msg.output)
	if msg.err != nil {
		if connect.IsConnectionAborted(msg.err) { // test if switching this is correct (may have to change launchExecCmd instead)
			statusCmd := m.setStatusError(fmt.Sprintf("%s to %s aborted.", msg.protocol, msg.target), statusTTL) // eg. if tail != nil && connect.IsConnectionAborted
			return m, tea.Batch(titleCmd, statusCmd)
		}
		if output != "" {
			statusCmd := m.setStatusError(fmt.Sprintf("%s to %s exited:\n%s (%v)", msg.protocol, msg.target, output, msg.err), 0)
			return m, tea.Batch(titleCmd, statusCmd)
		}
		statusCmd := m.setStatusError(fmt.Sprintf("%s to %s exited:\n%v", msg.protocol, msg.target, msg.err), 0)
		return m, tea.Batch(titleCmd, statusCmd)
	}

	statusCmd := m.setStatusSuccess(fmt.Sprintf("%s to %s ended.", msg.protocol, msg.target), statusTTL)
	return m, tea.Batch(titleCmd, statusCmd)
}
