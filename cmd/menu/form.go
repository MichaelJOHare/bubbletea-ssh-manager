package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type form struct {
	protocol   string // "ssh" or "telnet"
	alias      string // host alias
	hostname   string // hostname or IP address
	port       string // port number as string
	user       string // user name
	algHostKey string // host key algorithms
	algKex     string // key exchange algorithms
	algMACs    string // MAC algorithms
}

func (m model) openAddHostForm() (model, tea.Cmd, bool) {
	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormOldAlias = ""

	v := &form{protocol: "ssh"}
	form := buildHostForm(m.theme, modeAdd, "", v)

	m.ms.hostForm = form
	m.relayout()
	return m, form.Init(), true
}

func (m model) openEditHostForm() (model, tea.Cmd, bool) {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		m.setStatusError("Select a host to edit.", statusTTL)
		return m, nil, true
	}

	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormOldAlias = strings.TrimSpace(it.spec.Alias)

	// prefill form with existing host data
	v := &form{
		protocol:   strings.TrimSpace(it.protocol),
		alias:      strings.TrimSpace(it.spec.Alias),
		hostname:   strings.TrimSpace(it.spec.HostName),
		port:       strings.TrimSpace(it.spec.Port),
		user:       strings.TrimSpace(it.spec.User),
		algHostKey: strings.TrimSpace(it.options.HostKeyAlgorithms),
		algKex:     strings.TrimSpace(it.options.KexAlgorithms),
		algMACs:    strings.TrimSpace(it.options.MACs),
	}
	form := buildHostForm(m.theme, modeEdit, m.ms.hostFormOldAlias, v)

	m.ms.hostForm = form
	m.relayout()
	return m, form.Init(), true
}

func (m model) closeHostForm(status string, kind statusKind) (model, tea.Cmd) {
	m.mode = modeMenu
	m.ms.hostForm = nil
	m.ms.hostFormOldAlias = ""
	m.relayout()
	if strings.TrimSpace(status) == "" {
		return m, nil
	}
	return m, m.setStatus(status, kind, statusTTL)
}
