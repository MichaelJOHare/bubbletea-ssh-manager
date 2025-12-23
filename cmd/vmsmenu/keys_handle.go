package main

import (
	str "bubbletea-ssh-manager/internal/stringutil"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMsg handles app-specific keybindings.
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the query + list components.
func (m model) handleKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	// Full help behaves like a modal:
	// - '?' opens it (no-op if already open)
	// - 'left' closes it
	// - while open, ignore all other keys so search/prompt don't change
	if m.fullHelpOpen {
		switch msg.String() {
		case "left":
			m.fullHelpOpen = false
			m.lst.SetShowHelp(true)
			// if we were prompting for username, restore that status message
			if m.promptingUser {
				m.setStatus(fmt.Sprintf("Enter SSH username for %s", strings.TrimSpace(m.pendingHost.spec.Alias)), false, 0)
			}
			m.relayout()
			return m, nil, true
		case "D":
			// toggle detailed host info
			return m, nil, true
		case "E":
			// toggle edit host
			return m, nil, true
		case "A":
			// toggle add host
			return m, nil, true
		case "R":
			// toggle remove host
			return m, nil, true
		// maybe we should capture arrow keys to set a status message?
		// it's easy to miss the fact that the help is open
		default:
			return m, nil, true
		}
	}

	if msg.String() == "?" {
		m.fullHelpOpen = true
		m.lst.SetShowHelp(false)
		m.setStatus("", false, 0)
		m.relayout()
		return m, nil, true
	}

	if m.promptingUser {
		return m.handlePromptKeyMsg(msg)
	}

	return m.handleNormalKeyMsg(msg)
}

// handlePromptKeyMsg handles key messages when prompting for username.
//
// It returns (newModel, cmd, handled). Always returns handled=true.
func (m model) handlePromptKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		return m.clearPromptValue()

	case "left":
		return m.dismissPrompt()

	case "enter":
		u := strings.TrimSpace(m.prompt.Value())
		if u == "" {
			m.setStatus("Username required (left arrow to cancel)", true, 0)
			return m, nil, true
		}
		it := m.pendingHost
		m.promptingUser = false
		m.pendingHost = nil
		m.prompt.SetValue("")
		m.prompt.Blur()
		m.setStatus("", false, 0)

		if it == nil {
			m.setStatus("No host selected.", true, 0)
			return m, nil, true
		}
		it.spec.User = u
		return m.startConnect(it)
	}

	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd, true
}

// handleNormalKeyMsg handles key messages when not prompting for username.
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the query + list components.
func (m model) handleNormalKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch msg.String() {
	// cancel preflight or quit on Ctrl+C
	case "ctrl+c":
		if m.preflighting {
			return m.cancelPreflightCmd()
		}
		m.quitting = true
		return m, tea.Quit, true

	// quit on 'Q'
	case "Q":
		m.quitting = true
		return m, tea.Quit, true

	// esc to clear search if non-empty; otherwise do nothing
	case "esc":
		return m.clearSearch()

	// go back on left arrow if in a group or search is active
	case "left":
		if m.inGroup() {
			m.path = m.path[:len(m.path)-1]
			m.query.SetValue("")
			m.setCurrentMenu(m.current().children)
			m.setStatus("", false, 0)
		} else if m.query.Value() != "" {
			return m.clearSearch()
		}
		return m, nil, true

	// enter to navigate into group or connect to host
	case "enter":
		if it, ok := m.lst.SelectedItem().(*menuItem); ok {
			// navigate into group
			if it.kind == itemGroup {
				m.path = append(m.path, it)
				m.query.SetValue("")
				m.setCurrentMenu(it.children)
				m.setStatus("", false, 0)
				return m, nil, true
			}

			// connect to host
			if str.NormalizeString(it.protocol) == "ssh" {
				return m.beginUserPrompt(it, fmt.Sprintf("Enter SSH username for %s", strings.TrimSpace(it.spec.Alias)))
			}
			return m.startConnect(it)
		}
		return m, nil, true
	}

	return m, nil, false
}
