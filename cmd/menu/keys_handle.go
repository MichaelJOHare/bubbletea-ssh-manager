package main

import (
	"strings"

	str "bubbletea-ssh-manager/internal/stringutil"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// handleKeyMsg handles app-specific keybindings.
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the query + list components.
func (m model) handleKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch m.mode {
	case modeHostForm:
		// host add/edit is a modal: route keys to the form (with a couple of escapes)
		switch msg.String() {
		case "esc":
			nm, cmd := m.closeHostForm("Canceled add/edit host.", statusError)
			return nm, cmd, true
		}
		var cmd tea.Cmd
		if m.ms.hostForm != nil {
			mdl, c := m.ms.hostForm.Update(msg)
			if f, ok := mdl.(*huh.Form); ok {
				m.ms.hostForm = f
			}
			cmd = c
		}
		m.relayout()
		return m, cmd, true

	case modeHostDetails:
		return m.handleHostDetailsKeyMsg(msg)

	case modePreflight:
		// preflight is a modal: ignore all keys except quitting/cancel
		switch msg.String() {
		case "ctrl+c":
			return m.cancelPreflightCmd()
		case "Q", "shift+q":
			m.quitting = true
			return m, tea.Quit, true
		default:
			return m, nil, true
		}

	case modePromptUsername:
		return m.handlePromptKeyMsg(msg)
	}

	// default/menu behavior
	if msg.String() == "?" {
		m.mode = modeHostDetails
		m.lst.SetShowHelp(false) // hide base help
		m.setStatusInfo("", 0)   // hide status
		m.relayout()
		return m, nil, true
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
	switch msg.String() {
	case "left":
		m.mode = modeMenu // close modal
		m.lst.SetShowHelp(true)
		// if we were prompting for username, restore that status message
		if m.ms.pendingHost != nil {
			m.mode = modePromptUsername
			m.setStatusInfo(userPromptStatus(m.ms.pendingHost.spec.Alias), 0)
		}
		m.relayout()
		return m, nil, true

	case "E":
		return m.openEditHostForm()

	case "R":
		m.setStatusError("Remove not wired yet.", statusTTL)
		return m, nil, true
	default:
		return m, nil, true
	}
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
			m.setStatusError("Username required (left arrow to cancel)", 0)
			return m, nil, true
		}
		it := m.ms.pendingHost
		m.dismissPrompt()

		if it == nil {
			m.setStatusError("No host selected.", 0)
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
		if m.mode == modePreflight {
			return m.cancelPreflightCmd()
		}
		m.quitting = true
		return m, tea.Quit, true

	// open add host form on 'A'
	case "A":
		return m.openAddHostForm()

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
			m.setStatusInfo("", 0)
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
				m.setStatusInfo("", 0)
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
