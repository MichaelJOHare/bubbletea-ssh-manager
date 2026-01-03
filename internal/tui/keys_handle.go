package tui

import (
	"strings"

	str "bubbletea-ssh-manager/internal/stringutil"

	"github.com/charmbracelet/bubbles/key"
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
		nm, cmd := m.handleHostFormKeyMsg(msg)
		return nm, cmd, true

	case modeHostDetails:
		nm, cmd := m.handleHostDetailsKeyMsg(msg)
		return nm, cmd, true

	case modeConfirm:
		nm, cmd := m.handleConfirmKeyMsg(msg)
		return nm, cmd, true

	case modePreflight:
		// preflight is a modal: ignore all keys except quitting/cancel
		switch {
		case msg.String() == "ctrl+c":
			nm, cmd := m.cancelPreflightCmd()
			return nm, cmd, true
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit, true
		default:
			return m, nil, true
		}

	case modePromptUsername:
		nm, cmd := m.handlePromptKeyMsg(msg)
		return nm, cmd, true
	}

	return m.handleBaseKeyMsg(msg)
}

// handleHostFormKeyMsg handles key messages related to the host add/edit form.
//
// Host add/edit behaves like a modal:
//   - while open, it routes all keys to the form
//   - 'enter' on input fields attempts to submit the form
//   - on select fields, it selects the option (default behavior)
//
// It returns an error checked (newModel, cmd).
func (m model) handleHostFormKeyMsg(msg tea.KeyMsg) (model, tea.Cmd) {
	// host add/edit is a modal: route keys to the form
	if m.ms.hostForm == nil {
		return m, nil
	}

	if key.Matches(msg, m.keys.FormSubmit) {
		// if focused field is not a select (i.e. protocol selector)
		if _, ok := m.ms.hostForm.GetFocusedField().(*huh.Select[string]); !ok {
			// and is an input, attempt to submit the form
			if _, ok := m.ms.hostForm.GetFocusedField().(*huh.Input); ok {
				mdl, cmd := m.ms.hostForm.Update(msg)
				// update model's form reference
				if f, ok := mdl.(*huh.Form); ok {
					m.ms.hostForm = f
				}
				// if the form already completed/aborted, don't double-submit
				if m.ms.hostForm.State != huh.StateNormal {
					m.relayout()
					return m, cmd
				}
				// if there are validation errors, don't submit
				if len(m.ms.hostForm.Errors()) > 0 { // this will be changed after adding confirmation prompt
					m.relayout()
					return m, cmd
				}
				m.relayout()
				return m, tea.Batch(cmd, m.ms.hostForm.SubmitCmd)
			}
		}
	}

	mdl, cmd := m.ms.hostForm.Update(msg)
	if f, ok := mdl.(*huh.Form); ok {
		m.ms.hostForm = f
	}
	m.relayout()
	return m, cmd
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
func (m model) handleHostDetailsKeyMsg(msg tea.KeyMsg) (model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.CloseDetails):
		m.mode = modeMenu // close modal
		m.lst.SetShowHelp(true)
		// if we were prompting for username, restore that status message
		if m.ms.pendingHost != nil {
			m.mode = modePromptUsername
			m.setStatusInfo(userPromptStatus(m.ms.pendingHost.spec.Alias), 0)
		}
		m.relayout()
		return m, nil

	case key.Matches(msg, m.keys.Edit):
		nm, cmd := m.openEditHostForm()
		return nm, cmd

	case key.Matches(msg, m.keys.Remove):
		nm, cmd := m.openRemoveConfirm()
		return nm, cmd

	default:
		return m, nil
	}
}

// handleConfirmKeyMsg handles key messages when a confirmation prompt is displayed.
//
// It returns (newModel, cmd, handled). Always returns handled=true.
func (m model) handleConfirmKeyMsg(msg tea.KeyMsg) (model, tea.Cmd) {
	if m.ms.confirm == nil || m.ms.confirm.form == nil {
		// no form, best-effort return
		m.mode = modeMenu
		m.relayout()
		return m, nil
	}

	// update the confirm form with the key message
	mdl, cmd := m.ms.confirm.form.Update(msg)
	if f, ok := mdl.(*huh.Form); ok {
		m.ms.confirm.form = f
	}
	m.relayout()
	return m, cmd
}

// handlePromptKeyMsg handles key messages when prompting for username.
//
// It returns (newModel, cmd, handled). Always returns handled=true.
func (m model) handlePromptKeyMsg(msg tea.KeyMsg) (model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Clear):
		return m.clearPrompt()

	case key.Matches(msg, m.keys.Back):
		return m.dismissPrompt()

	case msg.String() == "enter":
		u := strings.TrimSpace(m.prompt.Value())
		if u == "" {
			m.setStatusError("Username required (left arrow to cancel)", 0)
			return m, nil
		}
		it := m.ms.pendingHost
		m.dismissPrompt()

		if it == nil {
			m.setStatusError("No host selected.", 0)
			return m, nil
		}
		it.spec.User = u
		return m.startConnect(it)
	}

	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

// handleBaseKeyMsg handles key messages from the base menu context.
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the query + list components.
func (m model) handleBaseKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch {
	// quit on Ctrl+C
	case msg.String() == "ctrl+c":
		m.quitting = true
		return m, tea.Quit, true

	// quit on 'Q'
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit, true

	// open add host form on 'A'
	case key.Matches(msg, m.keys.Add):
		nm, cmd := m.openAddHostForm()
		return nm, cmd, true

	// esc to clear search if non-empty; otherwise do nothing
	case key.Matches(msg, m.keys.Clear):
		nm, cmd := m.clearSearch()
		return nm, cmd, true

	// go back on left arrow if in a group or search is active
	case key.Matches(msg, m.keys.Back):
		if m.inGroup() {
			m.path = m.path[:len(m.path)-1]
			m.query.SetValue("")
			m.setCurrentMenu(m.current().children)
			m.setStatusInfo("", 0)
		} else if m.query.Value() != "" {
			nm, cmd := m.clearSearch()
			return nm, cmd, true
		}
		return m, nil, true

	// show host details on '?'
	case key.Matches(msg, m.keys.Details):
		m.mode = modeHostDetails
		m.lst.SetShowHelp(false) // hide base help
		m.setStatusInfo("", 0)   // hide status
		m.relayout()
		return m, nil, true

	// enter to navigate into group or connect to host
	case msg.String() == "enter":
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
				nm, cmd := m.beginUserPrompt(it)
				return nm, cmd, true
			}
			nm, cmd := m.startConnect(it)
			return nm, cmd, true
		}
		return m, nil, true
	}

	return m, nil, false
}
