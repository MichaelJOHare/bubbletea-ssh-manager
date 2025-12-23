package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

const (
	cursorUpSymbol   = "🡩"
	cursorDownSymbol = "🡫"
	leftBackSymbol   = "🡨"
	clearSymbol      = "esc"
	quitSymbol       = "Q"
	infoSymbol       = "?"
)

var (
	cursorUpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(cursorUpSymbol)   // green
	cursorDownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(cursorDownSymbol) // green
	leftBackStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(leftBackSymbol)  // purple
	quitStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(quitSymbol)        // red
	infoStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Render(infoSymbol)       // blue
	clearStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("188")).Render(clearSymbol)     // light grey
)

// New key bindings for the TUI added using AdditionalShortHelpKeys.
var (
	// esc to clear search if non-empty
	escClearKey = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp(clearStyle, "clear"),
	)
	// left arrow to go back if in a group
	leftBackKey = key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp(leftBackStyle, "back"),
	)

	rootHelpKeys  = func() []key.Binding { return []key.Binding{escClearKey} }
	groupHelpKeys = func() []key.Binding { return []key.Binding{leftBackKey, escClearKey} }
)

// initHelpKeys initializes the list's help keys.
//
// This is called once during model initialization.
func (m *model) initHelpKeys() {
	m.lst.KeyMap.CursorUp.SetHelp(cursorUpStyle, "up")
	m.lst.KeyMap.CursorUp.SetKeys("up")
	m.lst.KeyMap.CursorDown.SetHelp(cursorDownStyle, "down")
	m.lst.KeyMap.CursorDown.SetKeys("down")
	m.lst.KeyMap.ShowFullHelp.SetHelp(infoStyle, "info")
	m.lst.KeyMap.ShowFullHelp.SetKeys("?")
	m.lst.KeyMap.Quit.SetHelp(quitStyle, "quit")
	m.lst.KeyMap.Quit.SetKeys("shift+q")
}

// syncHelpKeys updates the list's additional help keys based on navigation state.
//
// This is called from relayout() so help stays in sync as the user navigates.
func (m *model) syncHelpKeys() {
	if m == nil {
		return
	}
	if m.inGroup() || m.promptingUser || m.query.Value() != "" {
		m.lst.AdditionalShortHelpKeys = groupHelpKeys
		return
	}
	m.lst.AdditionalShortHelpKeys = rootHelpKeys
}
