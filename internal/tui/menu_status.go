package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type statusKind int // type of status message (info/success/error)

const (
	statusInfo statusKind = iota
	statusSuccess
	statusError
)

// setStatus sets the status message and kind.
//
// It increments the status token to keep track of which status to clear
// when using a duration. If d > 0, it returns a command to clear the status
// after the specified duration. If d == 0, the status remains until changed.
func (m *model) setStatus(text string, kind statusKind, d time.Duration) tea.Cmd {
	if d < 0 {
		d = 0
	}

	m.statusToken++
	m.status = text
	m.statusKind = kind
	m.relayout()

	if d > 0 {
		tok := m.statusToken
		return tea.Tick(d, func(time.Time) tea.Msg {
			return statusClearMsg{token: tok}
		})
	}
	return nil
}

func (m *model) setStatusInfo(text string, d time.Duration) tea.Cmd {
	return m.setStatus(text, statusInfo, d)
}

func (m *model) setStatusSuccess(text string, d time.Duration) tea.Cmd {
	return m.setStatus(text, statusSuccess, d)
}

func (m *model) setStatusError(text string, d time.Duration) tea.Cmd {
	return m.setStatus(text, statusError, d)
}
