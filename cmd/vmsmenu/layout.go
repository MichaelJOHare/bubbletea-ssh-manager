package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

const footerPadLeft = 2

// Key binding for going back in the menu (when in a group).
var escBackKey = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "back"),
)

// inGroup returns true if the current path is inside a group (not at root).
func (m *model) inGroup() bool {
	return len(m.path) > 1
}

// relayout recalculates the sizes of the list and text input based on the current window size.
func (m *model) relayout() {
	// footer consumes lines at the bottom:
	// - optional hint (only in groups)
	// - optional status line
	// - search input (always)

	// default to 2 lines so search input is padded 1 line above
	// to separate it from help and status
	// add extra line for status when present
	footerLines := 2
	if strings.TrimSpace(m.status) != "" {
		footerLines++
	}

	// make sure the list doesn't overwrite the footer
	m.lst.SetSize(m.width, max(0, m.height-footerLines))

	// ensure the text input has enough width to render placeholder/prompt
	// in bubbles/textinput, Width is the content width, not including the prompt
	promptW := lipgloss.Width(m.query.Prompt)
	m.query.Width = max(0, m.width-footerPadLeft-promptW-1)

	// Keep the list help in sync with our navigation state.
	if m.inGroup() {
		m.lst.AdditionalShortHelpKeys = func() []key.Binding { return []key.Binding{escBackKey} }
	} else {
		m.lst.AdditionalShortHelpKeys = nil
	}
}
