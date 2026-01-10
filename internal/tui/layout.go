package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// layout constants
const (
	footerPadLeft         = 2  // match list padding
	footerDefaultLines    = 2  // search input padding
	footerPreflightLines  = 6  // spinner area: PaddingBottom(3) + PaddingTop(1) + text height(2)
	hostFormPadding       = 6  // left(3) + right(3)
	hostFormHeaderFooter  = 5  // header(1) + footer(2) + padding(2)
	confirmDialogPadding  = 6  // left(3) + right(3)
	confirmDialogMinWidth = 20 // minimum dialog width
	confirmDialogExtraW   = 10 // padding + border space
)

// isModalMode returns true if the current mode is a modal overlay that
// should suppress list navigation and help.
func (m *model) isModalMode() bool {
	switch m.mode {
	case modePreflight, modePromptUsername, modeHostDetails, modeHostForm, modeConfirm:
		return true
	}
	return false
}

// hidesListHelp returns true if the current mode should hide the list's help bar.
func (m *model) hidesListHelp() bool {
	switch m.mode {
	case modePreflight, modeHostDetails, modeHostForm, modeConfirm:
		return true
	}
	return false
}

// syncHelpKeys updates the list's additional help keys based on navigation state.
func (m *model) syncHelpKeys() {
	if m == nil {
		return
	}

	m.syncScrollKeys()
	m.lst.SetShowHelp(!m.hidesListHelp())
	m.syncAdditionalHelpKeys()
}

// syncScrollKeys enables or disables list cursor navigation based on mode.
func (m *model) syncScrollKeys() {
	canScroll := !m.isModalMode() && len(m.lst.Items()) > 1
	if canScroll {
		m.lst.KeyMap.CursorUp.SetKeys("up")
		m.lst.KeyMap.CursorDown.SetKeys("down")
	} else {
		m.lst.KeyMap.CursorUp.SetKeys()
		m.lst.KeyMap.CursorDown.SetKeys()
	}
}

// syncAdditionalHelpKeys sets the list's additional help keys based on state.
func (m *model) syncAdditionalHelpKeys() {
	if m.mode == modePromptUsername {
		m.lst.AdditionalShortHelpKeys = m.promptHelpKeys
		m.lst.KeyMap.Quit.SetKeys()         // Q gets captured by prompt modal
		m.lst.KeyMap.ShowFullHelp.SetKeys() // ? gets captured by prompt modal
		return
	}

	m.lst.KeyMap.Quit.SetKeys("Q")
	m.lst.KeyMap.ShowFullHelp.SetKeys("?")

	switch {
	case m.inGroup() || m.query.Value() != "":
		m.lst.AdditionalShortHelpKeys = m.groupHelpKeys
	default:
		m.lst.AdditionalShortHelpKeys = m.mainHelpKeys
	}
}

// relayout resizes the list and form components based on the current window size.
//
// It accounts for footer space used by status, preflight, paginator, and search/prompt.
func (m *model) relayout() {
	m.resizeList()
	m.resizePrompt()
	m.syncHelpKeys()
	m.resizeHostForm()
	m.resizeConfirmDialog()
}

// footerHeight calculates how many lines the footer area consumes.
func (m *model) footerHeight() int {
	lines := footerDefaultLines
	if m.mode == modePreflight {
		lines = footerPreflightLines
	}
	if m.status != "" {
		lines += 1 + lipgloss.Height(m.status) // padding + status text
	}
	return lines
}

// resizeList adjusts the list dimensions to fit above the footer.
func (m *model) resizeList() {
	m.lst.SetSize(m.width, max(0, m.height-m.footerHeight()))
}

// resizePrompt sizes the active text input to fill available width.
func (m *model) resizePrompt() {
	if m.mode == modePromptUsername {
		promptW := lipgloss.Width(m.prompt.Prompt)
		m.prompt.Width = max(0, m.width-footerPadLeft-promptW-1)
	} else {
		promptW := lipgloss.Width(m.query.Prompt)
		m.query.Width = max(0, m.width-footerPadLeft-promptW-1)
	}
}

// resizeHostForm sizes the host add/edit form to the window.
func (m *model) resizeHostForm() {
	if m.ms.hostForm == nil {
		return
	}
	w := max(0, m.width-hostFormStatusOuterWidth-hostFormStatusGap-hostFormPadding)
	h := max(0, m.height-hostFormHeaderFooter)
	m.ms.hostForm = m.ms.hostForm.WithWidth(w).WithHeight(h)
}

// resizeConfirmDialog sizes the confirmation dialog to fit its content.
func (m *model) resizeConfirmDialog() {
	if m.ms.confirm == nil || m.ms.confirm.form == nil {
		return
	}
	maxW := max(0, m.width-confirmDialogPadding)
	textW := max(lipgloss.Width(m.ms.confirm.title), lipgloss.Width(m.ms.confirm.description))
	w := min(max(confirmDialogMinWidth, textW+confirmDialogExtraW), maxW)
	m.ms.confirm.form = m.ms.confirm.form.WithWidth(w)
}
