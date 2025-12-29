package main

import (
	str "bubbletea-ssh-manager/internal/stringutil"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMsg handles app-specific keybindings.
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the query + list components.
func (m model) handleKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	// handle host-details modal first
	if m.hostDetailsOpen {
		if nm, cmd, handled := m.handleHostDetailsKeyMsg(msg); handled {
			return nm, cmd, true
		}
	}

	// preflight is a modal: ignore all keys except quitting/cancel
	if m.preflighting {
		switch msg.String() {
		case "ctrl+c":
			return m.cancelPreflightCmd()
		case "Q", "shift+q":
			m.quitting = true
			return m, tea.Quit, true
		default:
			return m, nil, true
		}
	}

	// keep host details available at any time except during preflight
	if msg.String() == "?" {
		m.hostDetailsOpen = true  // open host details modal
		m.lst.SetShowHelp(false)  // hide base help
		m.setStatus("", false, 0) // hide status
		m.relayout()
		return m, nil, true
	}

	// handle prompt input before search input
	if m.promptingUsername {
		return m.handlePromptKeyMsg(msg)
	}

	return m.handleBaseKeyMsg(msg)
}

// handleHostDetailsKeyMsg handles key messages related to the host details modal.
//
// Host details behaves like a modal:
//   - '?' opens it (no-op if already open)
//   - 'left' closes it
//   - while open, ignore all other keys so search/prompt don't change
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the other handlers.
func (m model) handleHostDetailsKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	if m.hostDetailsOpen {
		switch msg.String() {
		case "left":
			m.hostDetailsOpen = false // close modal
			m.lst.SetShowHelp(true)
			// if we were prompting for username, restore that status message
			if m.promptingUsername {
				m.setStatus(userPromptStatus(m.pendingHost.spec.Alias), false, 0)
			}
			m.relayout()
			return m, nil, true

		case "E":
			return m.openEditHostForm()

		case "A":
			return m.openAddHostForm()

		case "R":
			m.setStatus("Remove not wired yet.", true, statusTTL)
			return m, nil, true
		default:
			return m, nil, true
		}
	}

	return m, nil, false
}

// handlePromptKeyMsg handles key messages when prompting for username.
//
// It returns (newModel, cmd, handled). Always returns handled=true.
func (m model) handlePromptKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		return m.clearPrompt()

	case "left":
		return m.dismissPrompt()

	case "enter":
		u := strings.TrimSpace(m.prompt.Value())
		if u == "" {
			m.setStatus("Username required (left arrow to cancel)", true, 0)
			return m, nil, true
		}
		it := m.pendingHost
		m.dismissPrompt()

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

// handleBaseKeyMsg handles key messages from the base menu context.
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the query + list components.
func (m model) handleBaseKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
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
				return m.beginUserPrompt(it)
			}
			return m.startConnect(it)
		}
		return m, nil, true
	}

	return m, nil, false
}
