package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handleHostFormSubmit(msg formResultMsg) (model, tea.Cmd) {
	if msg.kind != formResultSubmitted {
		return m, nil
	}
	protocol := strings.ToLower(strings.TrimSpace(msg.protocol))
	alias := strings.TrimSpace(msg.spec.Alias)
	if alias == "" {
		nm, cmd := m.closeHostForm("Alias is required.", statusError)
		return nm, cmd
	}
	if protocol != "ssh" && protocol != "telnet" {
		nm, cmd := m.closeHostForm(fmt.Sprintf("Unknown protocol: %q", protocol), statusError)
		return nm, cmd
	}
	if protocol == "telnet" && strings.TrimSpace(msg.spec.HostName) == "" {
		// close the form and show error, refine this later
		nm, cmd := m.closeHostForm("HostName is required for telnet.", statusError)
		return nm, cmd
	}

	oldAlias := strings.TrimSpace(m.ms.hostFormOldAlias)
	if oldAlias == "" {
		oldAlias = strings.TrimSpace(msg.oldAlias)
	}

	// close the form before doing IO
	m, _ = m.closeHostForm("", statusInfo)

	saveCmd := func() tea.Msg {
		switch msg.mode {
		case modeAdd:
			return formSaveResultMsg{err: AddHostToRootConfig(protocol, msg.spec, msg.opts)}
		case modeEdit:
			if oldAlias == "" {
				return formSaveResultMsg{err: errors.New("missing old alias")}
			}
			return formSaveResultMsg{err: UpdateHostInConfig(protocol, oldAlias, msg.spec, msg.opts)}
		default:
			return formSaveResultMsg{err: errors.New("unknown form mode")}
		}
	}

	return m, tea.Cmd(func() tea.Msg { return saveCmd() })
}

func (m model) handleHostFormSaveResult(msg formSaveResultMsg) (model, tea.Cmd) {
	if msg.err == nil {
		root, err := seedMenu()
		cmd := func() tea.Msg {
			return menuReloadedMsg{root: root, err: err}
		}
		return m, tea.Batch(m.setStatusSuccess("Saved.", statusTTL), cmd)
	}
	if errors.Is(msg.err, os.ErrNotExist) {
		return m, m.setStatusError("Host not found.", statusTTL)
	}
	return m, m.setStatusError("Save failed: "+msg.err.Error(), 0)
}
