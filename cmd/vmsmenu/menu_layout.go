package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const footerPadLeft = 2

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

// syncHelpKeys updates the list's additional help keys based on navigation state.
func (m *model) syncHelpKeys() {
	if m == nil {
		return
	}

	// treat certain states as modals where list navigation/help should not apply
	modal := m.preflighting || m.promptingUsername || m.hostDetailsOpen || m.hostFormOpen()
	canScroll := !modal && len(m.lst.Items()) > 1
	if canScroll {
		m.lst.KeyMap.CursorUp.SetKeys("up")
		m.lst.KeyMap.CursorDown.SetKeys("down")
	} else {
		m.lst.KeyMap.CursorUp.SetKeys()
		m.lst.KeyMap.CursorDown.SetKeys()
	}

	// during preflight we hide the help entirely (only quitting/cancel is allowed)
	// during host details, the base list help is hidden (custom-rendered modal)
	if m.preflighting || m.hostDetailsOpen || m.hostFormOpen() {
		m.lst.SetShowHelp(false)
	} else {
		m.lst.SetShowHelp(true)
	}

	// set additional help keys based on state
	if m.promptingUsername {
		m.lst.AdditionalShortHelpKeys = promptHelpKeys
		m.lst.KeyMap.Quit.SetKeys() // shift+Q gets captured by prompt modal
		return                      // since a username can have a capital Q in it
	} else {
		m.lst.KeyMap.Quit.SetKeys("shift+q")
	}
	if m.inGroup() || m.query.Value() != "" {
		m.lst.AdditionalShortHelpKeys = groupHelpKeys
		return
	}

	// default: no additional help keys
	m.lst.AdditionalShortHelpKeys = nil
}

// relayout recalculates the sizes of the list and text input based on the current window size.
func (m *model) relayout() {
	// footer consumes lines at the bottom:
	// - optional preflight line (spinner + countdown)
	// - optional status line
	// - search/prompt input (hidden when host details modal is open)

	footerLines := 0
	if m.preflighting {
		// if preflight line is rendered, reserve space for it
		// in viewPreflight() it's rendered with PaddingBottom(3) + PaddingTop(1),
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

	// size host add/edit form to the window when open
	if m.hostForm != nil {
		w := max(0, m.width-6)
		h := max(0, m.height-6)
		m.hostForm = m.hostForm.WithWidth(w).WithHeight(h)
	}
}
