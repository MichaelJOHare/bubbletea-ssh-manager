package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

const (

	// Symbols for help keys

	cursorUpSymbol   = "🡩 "
	cursorUpHelp     = "up"
	cursorDownSymbol = "🡫 "
	cursorDownHelp   = "down"
	leftBackSymbol   = "🡨 "
	leftBackHelp     = "back"
	clearSymbol      = "esc"
	clearHelp        = "clear"
	addSymbol        = "A"
	addHelp          = "add"
	quitSymbol       = "Q"
	quitHelp         = "quit"
	infoSymbol       = "?"
	infoHelp         = "details"
	editSymbol       = "E"
	editHelp         = "edit"
	removeSymbol     = "R"
	removeHelp       = "remove"
)

func (m *model) keySymbolStyle(color lipgloss.Color, symbol string) string {
	return lipgloss.NewStyle().Foreground(color).Render(symbol)
}

func (m *model) keyHelpTextStyle(text string) string {
	return lipgloss.NewStyle().Foreground(m.theme.HelpText).Render(text)
}

// Key bindings used by the list help.

func (m *model) qQuitKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("shift+q"),
		key.WithHelp(
			m.keySymbolStyle(m.theme.KeyQuit, quitSymbol),
			m.keyHelpTextStyle(quitHelp),
		),
	)
}

func (m *model) escClearKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp(
			m.keySymbolStyle(m.theme.KeyClear, clearSymbol),
			m.keyHelpTextStyle(clearHelp),
		),
	)
}

func (m *model) leftBackKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp(
			m.keySymbolStyle(m.theme.KeyBack, leftBackSymbol),
			m.keyHelpTextStyle(leftBackHelp),
		),
	)
}

func (m *model) leftCloseHelpKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp(
			m.keySymbolStyle(m.theme.KeyBack, leftBackSymbol),
			m.keyHelpTextStyle("close"),
		),
	)
}

func (m *model) editKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp(
			m.keySymbolStyle(m.theme.KeyEdit, editSymbol),
			m.keyHelpTextStyle(editHelp),
		),
	)
}

func (m *model) addKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp(
			m.keySymbolStyle(m.theme.KeyAdd, addSymbol),
			m.keyHelpTextStyle(addHelp),
		),
	)
}

func (m *model) removeKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp(
			m.keySymbolStyle(m.theme.KeyRemove, removeSymbol),
			m.keyHelpTextStyle(removeHelp),
		),
	)
}

func (m *model) mainHelpKeys() []key.Binding  { return []key.Binding{m.addKey()} }
func (m *model) groupHelpKeys() []key.Binding { return []key.Binding{m.leftBackKey()} }
func (m *model) promptHelpKeys() []key.Binding {
	return []key.Binding{m.leftBackKey(), m.escClearKey()}
}
func (m model) moreHelpKeys() [][]key.Binding {
	return [][]key.Binding{
		{m.leftCloseHelpKey()},
		{m.editKey()},
		{m.removeKey()},
	}
}

// initHelpKeys initializes the list's help keys.
//
// This is called once during model initialization.
func (m *model) initHelpKeys() {
	m.lst.KeyMap.CursorUp.SetHelp(
		m.keySymbolStyle(m.theme.KeyCursor, cursorUpSymbol),
		m.keyHelpTextStyle(cursorUpHelp),
	)
	m.lst.KeyMap.CursorUp.SetKeys("up")
	m.lst.KeyMap.CursorDown.SetHelp(
		m.keySymbolStyle(m.theme.KeyCursor, cursorDownSymbol),
		m.keyHelpTextStyle(cursorDownHelp),
	)
	m.lst.KeyMap.CursorDown.SetKeys("down")
	m.lst.KeyMap.ShowFullHelp.SetHelp(
		m.keySymbolStyle(m.theme.KeyInfo, infoSymbol),
		m.keyHelpTextStyle(infoHelp),
	)
	m.lst.KeyMap.ShowFullHelp.SetKeys("?")
	m.lst.KeyMap.Quit.SetHelp(
		m.keySymbolStyle(m.theme.KeyQuit, quitSymbol),
		m.keyHelpTextStyle(quitHelp),
	)
	m.lst.KeyMap.Quit.SetKeys("shift+q")
}
