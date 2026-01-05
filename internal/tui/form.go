package tui

import (
	"strings"

	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"

	tea "github.com/charmbracelet/bubbletea"
)

type form struct {
	protocol  config.Protocol   // protocol
	groupName string            // group name portion of alias (display form; spaces allowed)
	nickname  string            // host nickname portion of alias (display form; spaces allowed)
	hostname  string            // hostname or IP address
	port      string            // port number as string
	user      string            // user name
	sshOpts   config.SSHOptions // SSH options
}

// openAddHostForm opens the host add form.
//
// It initializes an empty form for adding a new host.
func (m model) openAddHostForm() (model, tea.Cmd) {
	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormMode = modeAdd
	m.ms.hostFormOldAlias = ""

	v := &form{protocol: config.ProtocolSSH}
	form := buildHostForm(modeAdd, "", v, m.theme)

	m.ms.hostForm = form
	m.ms.hostFormValues = v
	m.relayout()
	return m, form.Init()
}

// openEditHostForm opens the host edit form for the selected host.
//
// It pre-fills the form with the existing host data.
func (m model) openEditHostForm() (model, tea.Cmd) {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		m.setStatusError("Select a host to edit.", statusTTL)
		return m, nil
	}

	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormMode = modeEdit
	m.ms.hostFormOldAlias = it.spec.Alias

	groupName, nickname := str.SplitAliasForDisplay(it.spec.Alias)

	// prefill form with existing host data
	v := &form{
		protocol:  it.protocol,
		groupName: groupName,
		nickname:  nickname,
		hostname:  it.spec.HostName,
		port:      it.spec.Port,
		user:      it.spec.User,
		sshOpts: config.SSHOptions{
			HostKeyAlgorithms: it.options.HostKeyAlgorithms,
			KexAlgorithms:     it.options.KexAlgorithms,
			MACs:              it.options.MACs,
		},
	}
	form := buildHostForm(modeEdit, m.ms.hostFormOldAlias, v, m.theme)

	m.ms.hostForm = form
	m.ms.hostFormValues = v
	m.relayout()
	return m, form.Init()
}

// closeHostForm closes the host form and resets related state.
//
// If a non-empty status message is provided, it sets that status.
func (m model) closeHostForm(status string, kind statusKind) (model, tea.Cmd) {
	m.mode = modeMenu
	m.ms.hostForm = nil
	m.ms.hostFormValues = nil
	m.ms.hostFormMode = modeAdd
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
func (m model) openRemoveConfirm() (model, tea.Cmd) {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		m.setStatusError("Select a host to remove.", statusTTL)
		return m, nil
	}

	m.mode = modeConfirm

	protocol := it.protocol
	alias := it.spec.Alias
	title := "Remove " + alias + "?"
	description := "This will remove the host from the config file."
	removeCmd := func() tea.Msg {
		err := config.RemoveHostFromConfig(protocol, alias)
		return removeHostResultMsg{
			protocol: protocol,
			alias:    alias,
			err:      err,
		}
	}
	cancelCmd := m.setStatusError(ErrorX+"Canceled removing "+alias+".", statusTTL)
	form := buildConfirmForm(
		title,
		description,
		m.theme,
	)
	m.ms.confirm = &confirmState{
		form:        form,
		title:       title,
		description: description,
		onConfirm:   tea.Cmd(removeCmd),
		onCancel:    cancelCmd,
	}

	m.relayout()
	return m, form.Init()
}

func (m model) openHostFormConfirm() (model, tea.Cmd) {
	// TODO: implement confirmation prompt when editing/adding host form
	return m, nil
}

// closeConfirm closes the confirmation dialog and returns to the appropriate mode.
func (m model) closeConfirm() (model, tea.Cmd) {
	m.mode = modeMenu
	m.ms.confirm = nil
	m.relayout()
	return m, nil
}
