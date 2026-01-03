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

// newBinding is a helper to create a key.Binding with styled help text.
// It applies the provided colors to the key symbol and help description.
func newBinding(keys []string, symbol string, help string, color lipgloss.Color, helpColor lipgloss.Color) key.Binding {
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(
			lipgloss.NewStyle().Foreground(color).Render(symbol),
			lipgloss.NewStyle().Foreground(helpColor).Render(help),
		),
	)
}

// NewKeyMap creates a new KeyMap with bindings customized to the provided theme.
//
// It styles the key symbols and help text according to the theme colors.
func NewKeyMap(theme Theme) KeyMap {

	return KeyMap{
		Quit: newBinding(
			[]string{"Q"},
			quitSymbol,
			quitHelp,
			theme.KeyQuit,
			theme.HelpText,
		),
		Details: newBinding(
			[]string{"?"},
			detailsSymbol,
			detailsHelp,
			theme.KeyInfo,
			theme.HelpText,
		),
		Add: newBinding(
			[]string{"A"},
			addSymbol,
			addHelp,
			theme.KeyAdd,
			theme.HelpText,
		),
		Back: newBinding(
			[]string{"left"},
			leftSymbol,
			backHelp,
			theme.KeyBack,
			theme.HelpText,
		),
		Clear: newBinding(
			[]string{"esc"},
			clearSymbol,
			clearHelp,
			theme.KeyClear,
			theme.HelpText,
		),
		CloseDetails: newBinding(
			[]string{"left"},
			leftSymbol,
			closeHelp,
			theme.KeyBack,
			theme.HelpText,
		),
		CloseForm: newBinding(
			[]string{"esc"},
			clearSymbol,
			closeFormHelp,
			theme.KeyClose,
			theme.HelpText,
		),
		FormNext: newBinding(
			[]string{"tab", "down"},
			nextSymbol,
			nextHelp,
			theme.KeyCursor,
			theme.HelpText,
		),
		FormPrev: newBinding(
			[]string{"shift+tab", "up"},
			prevSymbol,
			prevHelp,
			theme.KeyCursor,
			theme.HelpText,
		),
		FormSelect: newBinding(
			[]string{"enter"},
			enterSymbol,
			selectHelp,
			theme.KeyEnter,
			theme.HelpText,
		),
		FormSubmit: newBinding(
			[]string{"enter"},
			enterSymbol,
			saveHelp,
			theme.KeyEnter,
			theme.HelpText,
		),
		Edit: newBinding(
			[]string{"E"},
			editSymbol,
			editHelp,
			theme.KeyEdit,
			theme.HelpText,
		),
		Remove: newBinding(
			[]string{"R"},
			removeSymbol,
			removeHelp,
			theme.KeyRemove,
			theme.HelpText,
		),
		ConfirmSelect: newBinding(
			[]string{"enter"},
			enterSymbol,
			selectHelp,
			theme.KeyEnter,
			theme.HelpText,
		),
		LeftRight: newBinding(
			[]string{"left", "right"},
			leftRightSymbol,
			leftRightHelp,
			theme.KeyCursor,
			theme.HelpText,
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
