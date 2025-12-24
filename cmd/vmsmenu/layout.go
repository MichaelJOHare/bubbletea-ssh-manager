package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const footerPadLeft = 2

// fullHelpText renders a custom full help view containing only our app-level keys.
// We use a local copy of the help model to ensure width is consistent.
func (m model) fullHelpText() string {
	h := m.lst.Help
	h.Width = max(0, m.width-footerPadLeft)

	// Make full help use the same dotted separators as short help.
	h.FullSeparator = h.ShortSeparator
	h.Styles.FullSeparator = h.Styles.ShortSeparator
	return h.FullHelpView(moreHelpColumns)
}

// syncTitleStyles updates the list title and title bar styles based on the current width.
func (m *model) syncTitleStyles() {
	if m == nil {
		return
	}

	// Default bubbles/list styles:
	// - TitleBar padding: (top=0, right=0, bottom=1, left=1)
	// - Title padding: (vertical=0, horizontal=1)
	m.lst.Styles.TitleBar = m.lst.Styles.TitleBar.
		Padding(1, 0, 1, 1)

	m.lst.Styles.Title = m.lst.Styles.Title.
		Padding(0, 2)
}

// relayout recalculates the sizes of the list and text input based on the current window size.
func (m *model) relayout() {
	// footer consumes lines at the bottom:
	// - optional preflight line (spinner + countdown)
	// - optional status line
	// - search/prompt input (hidden when full help is open)

	footerLines := 0
	if m.fullHelpOpen {
		footerLines += lipgloss.Height(m.fullHelpText()) + 2 // +2 for padding above and border
	} else if m.preflighting {
		// if preflight line is rendered, reserve space for it
		// in viewPreflight() it's rendered with PaddingBottom(3) and PaddingTop(1),
		//  so add 4 for the padding plus the actual rendered height of the
		// preflight status text (2).
		footerLines = 4 + 2
	} else {
		footerLines = 2 // default to 2 lines so search input is padded by 1 line above
	}

	if strings.TrimSpace(m.status) != "" {
		// in viewNormal() status is also rendered with PaddingTop(1), so add 1 for the
		// padding plus the actual rendered height of the status text
		footerLines += 1 + lipgloss.Height(m.status)
	}

	// make sure the list doesn't overwrite the footer
	m.lst.SetSize(m.width, max(0, m.height-footerLines))
	m.syncTitleStyles()

	// ensure the text input has enough width to render placeholder/prompt
	// in bubbles/textinput, Width is the content width, not including the prompt
	if m.promptingUsername {
		promptW := lipgloss.Width(m.prompt.Prompt)
		m.prompt.Width = max(0, m.width-footerPadLeft-promptW-1)
	} else {
		promptW := lipgloss.Width(m.query.Prompt)
		m.query.Width = max(0, m.width-footerPadLeft-promptW-1)
	}

	// keep the list help in sync with our navigation state
	m.syncHelpKeys()
}
