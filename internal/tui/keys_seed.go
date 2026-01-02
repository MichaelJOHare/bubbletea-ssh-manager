package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

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
func (m model) confirmHelpKeys() []key.Binding {
	return []key.Binding{m.keys.LeftRight, m.keys.ConfirmSelect}
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
