package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
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

	leftRightSymbol = "🡨 |🡪 "
	leftRightHelp   = "choose selection"

	enterSymbol = "enter"
	saveHelp    = "save"
	selectHelp  = "select"

	nextSymbol = "🡫 "
	nextHelp   = "next"
	prevSymbol = "🡩 "
	prevHelp   = "prev"
)

type KeyMap struct {
	Quit          key.Binding
	Details       key.Binding
	Add           key.Binding
	Back          key.Binding
	Clear         key.Binding
	CloseDetails  key.Binding
	CloseForm     key.Binding
	FormNext      key.Binding
	FormPrev      key.Binding
	FormSelect    key.Binding
	FormSubmit    key.Binding
	Edit          key.Binding
	Remove        key.Binding
	ConfirmSelect key.Binding
	LeftRight     key.Binding
}

// NewKeyMap creates a new KeyMap with bindings customized to the provided theme.
//
// It styles the key symbols and help text according to the theme colors.
func NewKeyMap(theme Theme) KeyMap {
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
		ConfirmSelect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp(
				keySymbolStyle(theme.KeyEnter, enterSymbol),
				keyHelpTextStyle(selectHelp),
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

func NewFormKeyMap() *huh.KeyMap {
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
