package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const footerPadLeft = 2

// relayout recalculates the sizes of the list and text input based on the current window size.
func (m *model) relayout() {
	// footer consumes lines at the bottom:
	// - optional hint (only in groups)
	// - optional status line
	// - search input (always)

	// default to 2 lines so search input is padded by 1 line above
	footerLines := 2
	// if status is set add...
	if strings.TrimSpace(m.status) != "" {
		// one line for the status itself
		footerLines++
		// extra line for error status since they can be multi-line
		if m.statusIsError && strings.Contains(m.status, "\n") {
			footerLines++
		}
		// one line for padding above status
		footerLines++
	}

	// make sure the list doesn't overwrite the footer
	m.lst.SetSize(m.width, max(0, m.height-footerLines))

	// ensure the text input has enough width to render placeholder/prompt
	// in bubbles/textinput, Width is the content width, not including the prompt
	if m.promptingUser {
		promptW := lipgloss.Width(m.prompt.Prompt)
		m.prompt.Width = max(0, m.width-footerPadLeft-promptW-1)
	} else {
		promptW := lipgloss.Width(m.query.Prompt)
		m.query.Width = max(0, m.width-footerPadLeft-promptW-1)
	}

	// keep the list help in sync with our navigation state
	m.syncHelpKeys()
}
