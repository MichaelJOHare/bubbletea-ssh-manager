package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const footerPadLeft = 2

func (m *model) inGroup() bool {
	return len(m.path) > 1
}

func (m *model) relayout() {
	// footer consumes lines at the bottom:
	// - optional hint (only in groups)
	// - optional status line
	// - search input (always)
	footerLines := 2
	if strings.TrimSpace(m.status) != "" {
		footerLines++
	}
	if m.inGroup() {
		footerLines++
	}

	// make sure the list doesn't overwrite the footer
	m.lst.SetSize(m.width, max(0, m.height-footerLines))

	// ensure the text input has enough width to render placeholder/prompt
	// in bubbles/textinput, Width is the content width, not including the prompt
	promptW := lipgloss.Width(m.query.Prompt)
	m.query.Width = max(0, m.width-footerPadLeft-promptW-1)
}
