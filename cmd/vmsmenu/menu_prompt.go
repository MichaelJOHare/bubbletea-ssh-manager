package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// userPromptStatus returns the status message to show when prompting for a username.
func userPromptStatus(alias string) string {
	return fmt.Sprintf("Enter SSH username for %s", strings.TrimSpace(alias))
}

// beginUserPrompt starts prompting the user for a username for the given host.
//
// It sets the prompt state and status message, and returns the updated model.
func (m model) beginUserPrompt(it *menuItem) (model, tea.Cmd, bool) {
	if it == nil {
		m.setStatus("No host selected.", true, 0)
		return m, nil, true
	}
	m.mode = modePromptUsername
	m.ms.pendingHost = it
	// prefill with existing user (from config or previous override) for convenience
	m.prompt.SetValue(strings.TrimSpace(it.spec.User))
	m.prompt.Focus()
	m.setStatus(userPromptStatus(it.spec.Alias), false, 0)
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

// clearPrompt clears the prompt input if non-empty.
//
// It returns the updated model, a nil command, and true if handled.
func (m model) clearPrompt() (model, tea.Cmd, bool) {
	if strings.TrimSpace(m.prompt.Value()) != "" {
		m.prompt.SetValue("")
	}
	return m, nil, true
}

// dismissPrompt cancels the closes the current prompt.
//
// It returns the updated model, a nil command, and true if handled.
func (m model) dismissPrompt() (model, tea.Cmd, bool) {
	m.mode = modeMenu
	m.ms.pendingHost = nil
	m.prompt.SetValue("")
	m.prompt.Blur()
	m.setStatus("", false, 0)
	return m, nil, true
}
