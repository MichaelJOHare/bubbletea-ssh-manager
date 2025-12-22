package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

const (
	leftBackSymbol = "🡨"
)

var (
	leftBackStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("140")).Render(leftBackSymbol) // yellow
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
