package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const footerPadLeft = 2

// syncHelpKeys updates the list's additional help keys based on navigation state.
func (m *model) syncHelpKeys() {
	if m == nil {
		return
	}

	// treat certain states as modals where list navigation/help should not apply
	modal := (m.mode == modePreflight || m.mode == modePromptUsername ||
		m.mode == modeHostDetails || m.mode == modeHostForm || m.mode == modeConfirm)
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
	if m.mode == modePreflight || m.mode == modeHostDetails ||
		m.mode == modeHostForm || m.mode == modeConfirm {
		m.lst.SetShowHelp(false)
	} else {
		m.lst.SetShowHelp(true)
	}

	// set additional help keys based on state
	if m.mode == modePromptUsername {
		m.lst.AdditionalShortHelpKeys = m.promptHelpKeys
		m.lst.KeyMap.Quit.SetKeys()         // Q gets captured by prompt modal
		m.lst.KeyMap.ShowFullHelp.SetKeys() // ? gets captured by prompt modal
		return
	} else {
		m.lst.KeyMap.Quit.SetKeys("Q")
		m.lst.KeyMap.ShowFullHelp.SetKeys("?")
	}
	if m.inGroup() || m.query.Value() != "" {
		m.lst.AdditionalShortHelpKeys = m.groupHelpKeys
		return
	}

	// default: no additional help keys
	m.lst.AdditionalShortHelpKeys = m.mainHelpKeys
}

// relayout resizes the list and form components based on the current window size.
//
// It accounts for footer space used by status, preflight, and search/prompt.
func (m *model) relayout() {
	// footer consumes lines at the bottom:
	// - optional preflight line (spinner + countdown)
	// - optional status line
	// - search/prompt input (hidden when host details modal is open)

	footerLines := 0
	if m.mode == modePreflight {
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

	// ensure the text input has enough width to render placeholder/prompt
	// in bubbles/textinput, Width is the content width, not including the prompt
	if m.mode == modePromptUsername {
		promptW := lipgloss.Width(m.prompt.Prompt)
		m.prompt.Width = max(0, m.width-footerPadLeft-promptW-1)
	} else {
		promptW := lipgloss.Width(m.query.Prompt)
		m.query.Width = max(0, m.width-footerPadLeft-promptW-1)
	}

	// keep the list help in sync with our navigation state
	m.syncHelpKeys()

	// size host add/edit form to the window when open
	if m.ms.hostForm != nil {
		w := max(0, m.width-hostFormStatusOuterWidth-hostFormStatusGap-6)
		h := max(0, m.height-5) // header(1) + footer(2; 1 for help 1 for paginator) + padding(2)
		m.ms.hostForm = m.ms.hostForm.WithWidth(w).WithHeight(h)
	}

	// size confirm prompt to the window when open
	if m.ms.confirm != nil && m.ms.confirm.form != nil {
		maxAvailableW := max(0, m.width-6) // keep some breathing room near borders
		textW := max(lipgloss.Width(m.ms.confirm.title), lipgloss.Width(m.ms.confirm.description))
		w := min(max(20, textW+10), maxAvailableW)
		m.ms.confirm.form = m.ms.confirm.form.WithWidth(w)
	}
}
