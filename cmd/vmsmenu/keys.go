package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

const (
	leftBackSymbol = "🡨"
	quitSymbol     = "Q"
)

var (
	leftBackStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(leftBackSymbol) // yellow
	quitStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(quitSymbol)       // red
)

// New key bindings for the TUI added using AdditionalShortHelpKeys.
var (
	// esc to clear search if non-empty
	escClearKey = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear"),
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
	m.lst.KeyMap.CursorUp.SetHelp("🡩", "up")
	m.lst.KeyMap.CursorUp.SetKeys("up")
	m.lst.KeyMap.CursorDown.SetHelp("🡫", "down")
	m.lst.KeyMap.CursorDown.SetKeys("down")
	m.lst.KeyMap.ShowFullHelp.SetHelp("?", "info")
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
	if m.inGroup() || m.promptingUser {
		m.lst.AdditionalShortHelpKeys = groupHelpKeys
		return
	}
	m.lst.AdditionalShortHelpKeys = rootHelpKeys
}
