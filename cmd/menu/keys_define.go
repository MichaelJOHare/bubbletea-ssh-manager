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
	infoHelp         = "details"

	// Symbols for key bindings in the host details help view.

	editSymbol   = "E"
	editHelp     = "edit"
	addSymbol    = "A"
	addHelp      = "add"
	removeSymbol = "R"
	removeHelp   = "remove"
)

var (

	// Main help key styles.

	cursorUpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(cursorUpSymbol)   // green
	cursorDownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(cursorDownSymbol) // green
	leftBackStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(leftBackSymbol)  // purple
	quitStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(quitSymbol)        // red
	infoStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(infoSymbol)       // blue
	clearStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("188")).Render(clearSymbol)     // light grey

	// Host details help key styles.

	editStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(editSymbol) // orange
	addStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(addSymbol)   // green
	removeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(removeSymbol) // red

	// Help text color (for key descriptions)

	helpTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246")) // gray

	// Main help text styles

	cursorUpHelpStyle   = helpTextStyle.Render(cursorUpHelp)
	cursorDownHelpStyle = helpTextStyle.Render(cursorDownHelp)
	leftBackHelpStyle   = helpTextStyle.Render(leftBackHelp)
	quitHelpStyle       = helpTextStyle.Render(quitHelp)
	infoHelpStyle       = helpTextStyle.Render(infoHelp)
	clearHelpStyle      = helpTextStyle.Render(clearHelp)

	// Host details help text styles

	leftCloseHelpStyle = helpTextStyle.Render("close")
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
	// left arrow to close the host details modal
	leftCloseHelpKey = key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp(leftBackStyle, leftCloseHelpStyle),
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

	// Host details help layout: one key per column (horizontal).

	moreHelpKeys = [][]key.Binding{
		{leftCloseHelpKey},
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
