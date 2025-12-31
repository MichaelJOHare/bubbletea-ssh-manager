package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Symbols + help labels used throughout the app.

	cursorUpSymbol   = "🡩 "
	cursorUpHelp     = "up"
	cursorDownSymbol = "🡫 "
	cursorDownHelp   = "down"

	leftSymbol    = "🡨 "
	backHelp      = "back"
	closeHelp     = "close"
	clearSymbol   = "esc"
	clearHelp     = "clear"
	closeFormHelp = "close"
	addSymbol     = "A"
	addHelp       = "add"
	quitSymbol    = "Q"
	quitHelp      = "quit"
	detailsSymbol = "?"
	detailsHelp   = "details"
	editSymbol    = "E"
	editHelp      = "edit"
	removeSymbol  = "R"
	removeHelp    = "remove"

	leftRightSymbol = "🡨/🡪 "
	leftRightHelp   = "make selection"

	enterSymbol = "enter"
	saveHelp    = "save"
	selectHelp  = "select"

	nextSymbol = "🡫 "
	nextHelp   = "next"
	prevSymbol = "🡩 "
	prevHelp   = "prev"
)

type KeyMap struct {
	Quit         key.Binding
	Details      key.Binding
	Add          key.Binding
	Back         key.Binding
	Clear        key.Binding
	CloseDetails key.Binding
	CloseForm    key.Binding
	FormNext     key.Binding
	FormPrev     key.Binding
	FormSelect   key.Binding
	FormSubmit   key.Binding
	Edit         key.Binding
	Remove       key.Binding
	LeftRight    key.Binding
}

// newKeyMap creates a new KeyMap with bindings customized to the provided theme.
//
// It styles the key symbols and help text according to the theme colors.
func newKeyMap(theme Theme) KeyMap {
	keySymbolStyle := func(color lipgloss.Color, symbol string) string {
		return lipgloss.NewStyle().Foreground(color).Render(symbol)
	}
	keyHelpTextStyle := func(text string) string {
		return lipgloss.NewStyle().Foreground(theme.HelpText).Render(text)
	}

	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("Q"),
			key.WithHelp(
				keySymbolStyle(theme.KeyQuit, quitSymbol),
				keyHelpTextStyle(quitHelp),
			),
		),
		Details: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp(
				keySymbolStyle(theme.KeyInfo, detailsSymbol),
				keyHelpTextStyle(detailsHelp),
			),
		),
		Add: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp(
				keySymbolStyle(theme.KeyAdd, addSymbol),
				keyHelpTextStyle(addHelp),
			),
		),
		Back: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp(
				keySymbolStyle(theme.KeyBack, leftSymbol),
				keyHelpTextStyle(backHelp),
			),
		),
		Clear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp(
				keySymbolStyle(theme.KeyClear, clearSymbol),
				keyHelpTextStyle(clearHelp),
			),
		),
		CloseDetails: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp(
				keySymbolStyle(theme.KeyBack, leftSymbol),
				keyHelpTextStyle(closeHelp),
			),
		),
		CloseForm: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp(
				keySymbolStyle(theme.KeyClose, clearSymbol),
				keyHelpTextStyle(closeFormHelp),
			),
		),
		FormNext: key.NewBinding(
			key.WithKeys("tab", "down"),
			key.WithHelp(
				keySymbolStyle(theme.KeyCursor, nextSymbol),
				keyHelpTextStyle(nextHelp),
			),
		),
		FormPrev: key.NewBinding(
			key.WithKeys("shift+tab", "up"),
			key.WithHelp(
				keySymbolStyle(theme.KeyCursor, prevSymbol),
				keyHelpTextStyle(prevHelp),
			),
		),
		FormSelect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp(
				keySymbolStyle(theme.KeyEnter, enterSymbol),
				keyHelpTextStyle(selectHelp),
			),
		),
		FormSubmit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp(
				keySymbolStyle(theme.KeyEnter, enterSymbol),
				keyHelpTextStyle(saveHelp),
			),
		),
		Edit: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp(
				keySymbolStyle(theme.KeyEdit, editSymbol),
				keyHelpTextStyle(editHelp),
			),
		),
		Remove: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp(
				keySymbolStyle(theme.KeyRemove, removeSymbol),
				keyHelpTextStyle(removeHelp),
			),
		),
		LeftRight: key.NewBinding(
			key.WithKeys("left", "right"),
			key.WithHelp(
				keySymbolStyle(theme.KeyCursor, leftRightSymbol),
				keyHelpTextStyle(leftRightHelp),
			),
		),
	}
}

func newFormKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()

	// add tab/shift+tab navigation for input, note, and text fields
	km.Input.Next = key.NewBinding(key.WithKeys("tab", "down"))
	km.Input.Prev = key.NewBinding(key.WithKeys("shift+tab", "up"))
	km.Note.Next = key.NewBinding(key.WithKeys("tab", "down"))
	km.Note.Prev = key.NewBinding(key.WithKeys("shift+tab", "up"))
	km.Text.Next = key.NewBinding(key.WithKeys("tab", "down"))
	km.Text.Prev = key.NewBinding(key.WithKeys("shift+tab", "up"))

	// disable select filtering: /
	km.Select.Filter = key.NewBinding(key.WithKeys())
	km.Select.SetFilter = key.NewBinding(key.WithKeys())
	km.Select.ClearFilter = key.NewBinding(key.WithKeys())

	km.Quit = key.NewBinding(key.WithKeys("esc", "ctrl+c"))
	return km
}

func (m *model) mainHelpKeys() []key.Binding  { return []key.Binding{m.keys.Add} }
func (m *model) groupHelpKeys() []key.Binding { return []key.Binding{m.keys.Back} }
func (m *model) promptHelpKeys() []key.Binding {
	return []key.Binding{m.keys.Back, m.keys.Clear}
}
func (m model) detailsHelpKeys() []key.Binding {
	return []key.Binding{m.keys.CloseDetails, m.keys.Edit, m.keys.Remove}
}
func (m model) formHelpKeys() []key.Binding {
	return []key.Binding{m.keys.CloseForm, m.keys.FormPrev, m.keys.FormNext}
}

// initHelpKeys initializes the list's help keys.
//
// This is called once during model initialization.
func (m *model) initHelpKeys() {
	// base list help (up/down keys + details + quit)
	m.lst.KeyMap.CursorUp.SetHelp(
		lipgloss.NewStyle().Foreground(m.theme.KeyCursor).Render(cursorUpSymbol),
		lipgloss.NewStyle().Foreground(m.theme.HelpText).Render(cursorUpHelp),
	)
	m.lst.KeyMap.CursorUp.SetKeys("up")
	m.lst.KeyMap.CursorDown.SetHelp(
		lipgloss.NewStyle().Foreground(m.theme.KeyCursor).Render(cursorDownSymbol),
		lipgloss.NewStyle().Foreground(m.theme.HelpText).Render(cursorDownHelp),
	)
	m.lst.KeyMap.CursorDown.SetKeys("down")
	infoH := m.keys.Details.Help()
	m.lst.KeyMap.ShowFullHelp.SetHelp(infoH.Key, infoH.Desc)
	m.lst.KeyMap.ShowFullHelp.SetKeys("?")
	qH := m.keys.Quit.Help()
	m.lst.KeyMap.Quit.SetHelp(qH.Key, qH.Desc)
	m.lst.KeyMap.Quit.SetKeys("Q")
}
