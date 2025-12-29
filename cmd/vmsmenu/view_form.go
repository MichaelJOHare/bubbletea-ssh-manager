package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type hostFormValues struct {
	protocol   string
	alias      string
	hostname   string
	port       string
	user       string
	algHostKey string
	algKex     string
	algMACs    string
}

func buildHostFormFromValues(mode hostFormMode, oldAlias string, v *hostFormValues) *huh.Form {
	if v == nil {
		v = &hostFormValues{}
	}
	return buildHostForm(
		mode,
		oldAlias,
		&v.protocol,
		&v.alias,
		&v.hostname,
		&v.port,
		&v.user,
		&v.algHostKey,
		&v.algKex,
		&v.algMACs,
	)
}

func (m model) openAddHostForm() (model, tea.Cmd, bool) {
	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormOldAlias = ""

	v := &hostFormValues{protocol: "ssh"}
	form := buildHostFormFromValues(hostFormAdd, "", v)

	m.ms.hostForm = form
	m.relayout()
	return m, form.Init(), true
}

func (m model) openEditHostForm() (model, tea.Cmd, bool) {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		m.setStatus("Select a host to edit.", true, statusTTL)
		return m, nil, true
	}

	// close other modals
	m.mode = modeHostForm
	m.ms.pendingHost = nil
	m.ms.hostFormOldAlias = strings.TrimSpace(it.spec.Alias)

	v := &hostFormValues{
		protocol:   strings.TrimSpace(it.protocol),
		alias:      strings.TrimSpace(it.spec.Alias),
		hostname:   strings.TrimSpace(it.spec.HostName),
		port:       strings.TrimSpace(it.spec.Port),
		user:       strings.TrimSpace(it.spec.User),
		algHostKey: strings.TrimSpace(it.options.HostKeyAlgorithms),
		algKex:     strings.TrimSpace(it.options.KexAlgorithms),
		algMACs:    strings.TrimSpace(it.options.MACs),
	}
	form := buildHostFormFromValues(hostFormEdit, m.ms.hostFormOldAlias, v)

	m.ms.hostForm = form
	m.relayout()
	return m, form.Init(), true
}

func (m model) closeHostForm(status string, isErr bool) (model, tea.Cmd) {
	m.mode = modeMenu
	m.ms.hostForm = nil
	m.ms.hostFormOldAlias = ""
	m.relayout()
	if strings.TrimSpace(status) == "" {
		return m, nil
	}
	return m, m.setStatus(status, isErr, statusTTL)
}

func buildHostForm(
	mode hostFormMode,
	oldAlias string,
	protocol *string,
	alias *string,
	hostname *string,
	port *string,
	user *string,
	algHostKey *string,
	algKex *string,
	algMACs *string,
) *huh.Form {
	title := "Add Host"
	if mode == hostFormEdit {
		title = "Edit Host"
	}

	protoField := huh.NewSelect[string]().
		Key("protocol").
		Title("Protocol").
		Options(
			huh.NewOption("ssh", "ssh"),
			huh.NewOption("telnet", "telnet"),
		).
		Value(protocol)

	aliasField := huh.NewInput().
		Key("alias").
		Title("Alias").
		Validate(huh.ValidateNotEmpty()).
		Value(alias)

	hostField := huh.NewInput().
		Key("hostname").
		Title("HostName").
		Value(hostname)

	portField := huh.NewInput().
		Key("port").
		Title("Port").
		Value(port)

	userField := huh.NewInput().
		Key("user").
		Title("User").
		Value(user)

	hostKeyField := huh.NewInput().
		Key("hostkeyalgorithms").
		Title("HostKeyAlgorithms").
		Value(algHostKey)

	kexField := huh.NewInput().
		Key("kexalgorithms").
		Title("KexAlgorithms").
		Value(algKex)

	macsField := huh.NewInput().
		Key("macs").
		Title("MACs").
		Value(algMACs)

	fields := []huh.Field{
		huh.NewNote().Title(title).Description("Enter host details and submit."),
	}
	if mode == hostFormAdd {
		fields = append(fields, protoField)
	}
	fields = append(fields,
		aliasField,
		hostField,
		portField,
		userField,
		hostKeyField,
		kexField,
		macsField,
	)

	group := huh.NewGroup(fields...)

	form := huh.NewForm(group).
		WithShowHelp(false).
		WithShowErrors(true)

	form.CancelCmd = func() tea.Msg { return hostFormResultMsg{kind: hostFormResultCanceled} }
	form.SubmitCmd = func() tea.Msg {
		p := strings.ToLower(strings.TrimSpace(*protocol))
		s := host.Spec{
			Alias:    strings.TrimSpace(*alias),
			HostName: strings.TrimSpace(*hostname),
			Port:     strings.TrimSpace(*port),
			User:     strings.TrimSpace(*user),
		}
		opts := sshopts.Options{
			HostKeyAlgorithms: strings.TrimSpace(*algHostKey),
			KexAlgorithms:     strings.TrimSpace(*algKex),
			MACs:              strings.TrimSpace(*algMACs),
		}
		return hostFormResultMsg{kind: hostFormResultSubmitted, mode: mode, protocol: p, oldAlias: strings.TrimSpace(oldAlias), spec: s, opts: opts}
	}

	return form
}

func reloadMenuCmd() tea.Cmd {
	return func() tea.Msg {
		root, err := seedMenu()
		return menuReloadedMsg{root: root, err: err}
	}
}

func (m model) applyReloadedMenu(msg menuReloadedMsg) (model, tea.Cmd) {
	if msg.root == nil {
		return m, m.setStatus("Failed to reload menu.", true, statusTTL)
	}
	m.root = msg.root
	m.path = []*menuItem{msg.root}
	m.query.SetValue("")
	m.setCurrentMenu(msg.root.children)
	m.relayout()
	if msg.err != nil {
		return m, m.setStatus("Config: "+msg.err.Error(), true, statusTTL)
	}
	return m, nil
}

func (m model) handleHostFormSubmit(msg hostFormResultMsg) (model, tea.Cmd) {
	if msg.kind != hostFormResultSubmitted {
		return m, nil
	}
	protocol := strings.ToLower(strings.TrimSpace(msg.protocol))
	alias := strings.TrimSpace(msg.spec.Alias)
	if alias == "" {
		nm, cmd := m.closeHostForm("Alias is required.", true)
		return nm, cmd
	}
	if protocol != "ssh" && protocol != "telnet" {
		nm, cmd := m.closeHostForm(fmt.Sprintf("Unknown protocol: %q", protocol), true)
		return nm, cmd
	}
	if protocol == "telnet" && strings.TrimSpace(msg.spec.HostName) == "" {
		// close the form and show error, refine this later
		nm, cmd := m.closeHostForm("HostName is required for telnet.", true)
		return nm, cmd
	}

	oldAlias := strings.TrimSpace(m.ms.hostFormOldAlias)
	if oldAlias == "" {
		oldAlias = strings.TrimSpace(msg.oldAlias)
	}

	// close the form before doing IO
	m, _ = m.closeHostForm("", false)

	saveCmd := func() tea.Msg {
		switch msg.mode {
		case hostFormAdd:
			return hostFormSaveResultMsg{err: AddHostToRootConfig(protocol, msg.spec, msg.opts)}
		case hostFormEdit:
			if oldAlias == "" {
				return hostFormSaveResultMsg{err: errors.New("missing old alias")}
			}
			return hostFormSaveResultMsg{err: UpdateHostInConfig(protocol, oldAlias, msg.spec, msg.opts)}
		default:
			return hostFormSaveResultMsg{err: errors.New("unknown form mode")}
		}
	}

	return m, tea.Cmd(func() tea.Msg { return saveCmd() })
}

func (m model) handleHostFormSaveResult(msg hostFormSaveResultMsg) (model, tea.Cmd) {
	if msg.err == nil {
		return m, tea.Batch(m.setStatus("Saved.", false, statusTTL), reloadMenuCmd())
	}
	if errors.Is(msg.err, os.ErrNotExist) {
		return m, m.setStatus("Host not found.", true, statusTTL)
	}
	return m, m.setStatus("Save failed: "+msg.err.Error(), true, 0)
}
