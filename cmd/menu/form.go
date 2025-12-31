package main

import (
	"strings"

	str "bubbletea-ssh-manager/internal/stringutil"

	tea "github.com/charmbracelet/bubbletea"
)

// openAddHostForm opens the host add form.
//
// It initializes an empty form for adding a new host.
func (m model) openAddHostForm() (model, tea.Cmd, bool) {
	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormMode = modeAdd
	m.ms.hostFormProtocol = "ssh"
	m.ms.hostFormOldAlias = ""

	v := &form{protocol: "ssh"}
	form := buildHostForm(modeAdd, "", v, m.theme)

	m.ms.hostForm = form
	m.relayout()
	return m, form.Init(), true
}

// openEditHostForm opens the host edit form for the selected host.
//
// It pre-fills the form with the existing host data.
func (m model) openEditHostForm() (model, tea.Cmd, bool) {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		m.setStatusError("Select a host to edit.", statusTTL)
		return m, nil, true
	}

	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormMode = modeEdit
	m.ms.hostFormProtocol = strings.TrimSpace(it.protocol)
	m.ms.hostFormOldAlias = strings.TrimSpace(it.spec.Alias)

	groupName, nickname := str.SplitAliasForDisplay(strings.TrimSpace(it.spec.Alias))

	// prefill form with existing host data
	v := &form{
		protocol:   strings.TrimSpace(it.protocol),
		groupName:  groupName,
		nickname:   nickname,
		hostname:   strings.TrimSpace(it.spec.HostName),
		port:       strings.TrimSpace(it.spec.Port),
		user:       strings.TrimSpace(it.spec.User),
		algHostKey: strings.TrimSpace(it.options.HostKeyAlgorithms),
		algKex:     strings.TrimSpace(it.options.KexAlgorithms),
		algMACs:    strings.TrimSpace(it.options.MACs),
	}
	form := buildHostForm(modeEdit, m.ms.hostFormOldAlias, v, m.theme)

	m.ms.hostForm = form
	m.relayout()
	return m, form.Init(), true
}

// closeHostForm closes the host form and resets related state.
//
// If a non-empty status message is provided, it sets that status.
func (m model) closeHostForm(status string, kind statusKind) (model, tea.Cmd) {
	m.mode = modeMenu
	m.ms.hostForm = nil
	m.ms.hostFormMode = modeAdd
	m.ms.hostFormProtocol = ""
	m.ms.hostFormOldAlias = ""
	m.relayout()
	if strings.TrimSpace(status) == "" {
		return m, nil
	}
	return m, m.setStatus(status, kind, statusTTL)
}

// openRemoveConfirm opens a confirmation dialog for removing the selected host.
//
// It displays a huh.Confirm prompt below the details box.
func (m model) openRemoveConfirm() (model, tea.Cmd, bool) {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		m.setStatusError("Select a host to remove.", statusTTL)
		return m, nil, true
	}

	m.mode = modeConfirm
	m.ms.confirmKind = confirmRemoveHost
	m.ms.confirmHost = it

	protocol := str.NormalizeString(it.protocol)
	alias := strings.TrimSpace(it.spec.Alias)
	form := buildConfirmForm(
		confirmRemoveHost,
		"Remove "+alias+"?",
		"This will remove the host from the config file.",
		protocol,
		alias,
		it,
		m.theme,
	)
	m.ms.confirmForm = form

	m.relayout()
	return m, form.Init(), true
}

// closeConfirm closes the confirmation dialog and returns to the appropriate mode.
func (m model) closeConfirm() (model, tea.Cmd) {
	// always return to menu mode for now
	m.mode = modeMenu
	m.ms.confirmForm = nil
	m.ms.confirmKind = confirmNone
	m.ms.confirmHost = nil
	m.relayout()
	return m, nil
}
