package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// beginUserPrompt starts prompting the user for a username for the given host.
//
// It sets the prompt state and status message, and returns the updated model.
func (m model) beginUserPrompt(it *menuItem, title string) (model, tea.Cmd, bool) {
	if it == nil {
		m.setStatus("No host selected.", true, 0)
		return m, nil, true
	}
	m.promptingUser = true
	m.pendingHost = it
	// prefill with existing user (from config or previous override) for convenience
	m.prompt.SetValue(strings.TrimSpace(it.spec.User))
	m.prompt.Focus()
	m.setStatus(title, false, 0)
	m.toggleCursorKeys(false)
	return m, nil, true
}

// clearSearch clears the search query if non-empty.
//
// It returns the updated model, a nil command, and true if handled.
func (m model) clearSearch() (model, tea.Cmd, bool) {
	if strings.TrimSpace(m.query.Value()) != "" {
		m.query.SetValue("")
		m.applyFilter("")
		m.relayout()
	}
	return m, nil, true
}

// clearPromptValue clears the prompt input value if non-empty.
//
// It returns the updated model, a nil command, and true if handled.
func (m model) clearPromptValue() (model, tea.Cmd, bool) {
	if strings.TrimSpace(m.prompt.Value()) != "" {
		m.prompt.SetValue("")
	}
	return m, nil, true
}

// dismissPrompt cancels the closes the current prompt.
//
// It returns the updated model, a nil command, and true if handled.
func (m model) dismissPrompt() (model, tea.Cmd, bool) {
	m.promptingUser = false
	m.pendingHost = nil
	m.prompt.SetValue("")
	m.prompt.Blur()
	m.toggleCursorKeys(true)
	m.setStatus("", false, 0)
	return m, nil, true
}
