package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

const (

	// Symbols for key bindings in the main help view.

	cursorUpSymbol   = "🡩 "
	cursorUpHelp     = "up"
	cursorDownSymbol = "🡫 "
	cursorDownHelp   = "down"
	leftBackSymbol   = "🡨 "
	leftBackHelp     = "back"
	clearSymbol      = "esc"
	clearHelp        = "clear"
	quitSymbol       = "Q"
	quitHelp         = "quit"
	infoSymbol       = "?"
	infoHelp         = "more"

	// Symbols for key bindings in the "more (?)" help view.

	detailsSymbol = "D"
	detailsHelp   = "details"
	editSymbol    = "E"
	editHelp      = "edit"
	addSymbol     = "A"
	addHelp       = "add"
	removeSymbol  = "R"
	removeHelp    = "remove"
)

var (

	// Main help key styles.

	cursorUpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(cursorUpSymbol)   // green
	cursorDownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(cursorDownSymbol) // green
	leftBackStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(leftBackSymbol)  // purple
	quitStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(quitSymbol)        // red
	infoStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(infoSymbol)       // blue
	clearStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("188")).Render(clearSymbol)     // light grey

	// "More" help key styles.

	detailsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(detailsSymbol) // blue
	editStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(editSymbol)   // orange
	addStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(addSymbol)     // green
	removeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(removeSymbol)   // red

	// Help text color (for key descriptions)

	helpTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246")) // gray

	// Main help text styles

	cursorUpHelpStyle   = helpTextStyle.Render(cursorUpHelp)
	cursorDownHelpStyle = helpTextStyle.Render(cursorDownHelp)
	leftBackHelpStyle   = helpTextStyle.Render(leftBackHelp)
	quitHelpStyle       = helpTextStyle.Render(quitHelp)
	infoHelpStyle       = helpTextStyle.Render(infoHelp)
	clearHelpStyle      = helpTextStyle.Render(clearHelp)

	// "More" help text styles

	leftCloseHelpStyle = helpTextStyle.Render("close")
	detailsHelpStyle   = helpTextStyle.Render(detailsHelp)
	editHelpStyle      = helpTextStyle.Render(editHelp)
	addHelpStyle       = helpTextStyle.Render(addHelp)
	removeHelpStyle    = helpTextStyle.Render(removeHelp)
)

// New key bindings for the TUI added using AdditionalShortHelpKeys.
var (
	// shift+Q to quit the application
	qQuitKey = key.NewBinding(
		key.WithKeys("shift+q"),
		key.WithHelp(quitStyle, quitHelpStyle),
	)
	// esc to clear search if non-empty
	escClearKey = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp(clearStyle, clearHelpStyle),
	)
	// left arrow to go back if in a group
	leftBackKey = key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp(leftBackStyle, leftBackHelpStyle),
	)
	// left arrow to close the full help view
	leftCloseHelpKey = key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp(leftBackStyle, leftCloseHelpStyle),
	)
	// D to show more details about the selected host
	detailsKey = key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp(detailsStyle, detailsHelpStyle),
	)
	// E to edit the selected host
	editKey = key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp(editStyle, editHelpStyle),
	)
	// A to add a new host or group
	addKey = key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp(addStyle, addHelpStyle),
	)
	// R to remove the selected host or group
	removeKey = key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp(removeStyle, removeHelpStyle),
	)

	// Functions returning slices of key bindings for different contexts.

	groupHelpKeys  = func() []key.Binding { return []key.Binding{leftBackKey} }
	promptHelpKeys = func() []key.Binding { return []key.Binding{leftBackKey, escClearKey} }

	// Full help layout: one key per column (horizontal).
	moreHelpColumns = [][]key.Binding{
		{leftCloseHelpKey},
		{detailsKey},
		{editKey},
		{addKey},
		{removeKey},
	}
)

// initHelpKeys initializes the list's help keys.
//
// This is called once during model initialization.
func (m *model) initHelpKeys() {
	m.lst.KeyMap.CursorUp.SetHelp(cursorUpStyle, cursorUpHelpStyle)
	m.lst.KeyMap.CursorUp.SetKeys("up")
	m.lst.KeyMap.CursorDown.SetHelp(cursorDownStyle, cursorDownHelpStyle)
	m.lst.KeyMap.CursorDown.SetKeys("down")
	m.lst.KeyMap.ShowFullHelp.SetHelp(infoStyle, infoHelpStyle)
	m.lst.KeyMap.ShowFullHelp.SetKeys("?")
	m.lst.KeyMap.Quit.SetHelp(quitStyle, quitHelpStyle)
	m.lst.KeyMap.Quit.SetKeys("shift+q")
}

// syncHelpKeys updates the list's additional help keys based on navigation state.
//
// This is called from relayout() so help stays in sync as the user navigates.
func (m *model) syncHelpKeys() {
	if m == nil {
		return
	}

	// treat certain states as modals where list navigation/help should not apply
	modal := m.preflighting || m.promptingUsername || m.fullHelpOpen
	canScroll := !modal && len(m.lst.Items()) > 1
	if canScroll {
		m.lst.KeyMap.CursorUp.SetKeys("up")
		m.lst.KeyMap.CursorDown.SetKeys("down")
	} else {
		m.lst.KeyMap.CursorUp.SetKeys()
		m.lst.KeyMap.CursorDown.SetKeys()
	}

	// during preflight we hide the help entirely (only quitting/cancel is allowed)
	// during full help, the base list help is hidden (custom-rendered modal)
	if m.preflighting || m.fullHelpOpen {
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
