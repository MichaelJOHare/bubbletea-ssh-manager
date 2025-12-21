package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMsg handles app-specific keybindings.
//
// It returns (newModel, cmd, handled). If handled is false, the caller should
// pass the message through to the query + list components.
func (m model) handleKeyMsg(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch msg.String() {
	// quit on Ctrl+C or 'q'
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit, true

	// show info on selected item with '?'
	case "?":
		if it, ok := m.lst.SelectedItem().(*menuItem); ok {
			if it.kind == itemGroup {
				cmd := m.setTemporaryStatus(fmt.Sprintf("Group: %s (%d items)", it.name, len(it.children)), false, infoStatusTTL)
				m.relayout()
				return m, cmd, true
			}
			if m.delegate != nil && m.delegate.groupHints != nil {
				if grp := strings.TrimSpace(m.delegate.groupHints[it]); grp != "" {
					cmd := m.setTemporaryStatus(fmt.Sprintf("Host: %s (%s) in %s", it.name, it.protocol, grp), false, infoStatusTTL)
					m.relayout()
					return m, cmd, true
				}
			}
			cmd := m.setTemporaryStatus(fmt.Sprintf("Host: %s (%s)", it.name, it.protocol), false, infoStatusTTL)
			m.relayout()
			return m, cmd, true
		}
		return m, nil, true

	// esc to clear search if non-empty; otherwise do nothing
	case "esc":
		if strings.TrimSpace(m.query.Value()) != "" {
			m.query.SetValue("")
			m.applyFilter("")
			m.relayout()
		}
		return m, nil, true

	// go back on left arrow if in a group
	case "left":
		if m.inGroup() {
			m.path = m.path[:len(m.path)-1]
			m.query.SetValue("")
			m.setCurrentMenu(m.current().children)
			m.setStatus("", false)
			m.relayout()
		}
		return m, nil, true

	// enter to navigate into group or connect to host
	case "enter":
		if it, ok := m.lst.SelectedItem().(*menuItem); ok {
			// navigate into group
			if it.kind == itemGroup {
				m.path = append(m.path, it)
				m.query.SetValue("")
				m.setCurrentMenu(it.children)
				m.setStatus("", false)
				m.relayout()
				return m, nil, true
			}

			// else connect to host
			cmd, protocol, target, tail, err := buildConnectCommand(it)
			if err != nil {
				m.setStatus(err.Error(), true)
				m.relayout()
				return m, nil, true
			}
			m.setStatus(fmt.Sprintf("Starting %s %s…", protocol, target), false)
			m.relayout()
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				out := strings.TrimSpace(tail.String())
				out = lastNonEmptyLine(out)
				return connectFinishedMsg{protocol: protocol, target: target, err: err, output: out}
			}), true
		}
		return m, nil, true
	}

	return m, nil, false
}
