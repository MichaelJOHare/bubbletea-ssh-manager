package main

import (
	"strings"

	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type formMode int // add vs edit mode for host entry form

const (
	modeAdd formMode = iota
	modeEdit
)

func buildHostForm(mode formMode, oldAlias string, v *form) *huh.Form {
	if v == nil {
		v = &form{}
	}
	if mode == modeAdd && strings.TrimSpace(v.protocol) == "" {
		v.protocol = "ssh"
	}

	title := "Add Host"
	if mode == modeEdit {
		title = "Edit Host"
	}

	protoField := huh.NewSelect[string]().
		Key("protocol").
		Title("Protocol").
		Options(
			huh.NewOption("ssh", "ssh"),
			huh.NewOption("telnet", "telnet"),
		).
		Value(&v.protocol)

	aliasField := huh.NewInput().
		Key("alias").
		Title("Alias").
		Validate(huh.ValidateNotEmpty()).
		Value(&v.alias)

	hostField := huh.NewInput().
		Key("hostname").
		Title("HostName").
		Value(&v.hostname)

	portField := huh.NewInput().
		Key("port").
		Title("Port").
		Value(&v.port)

	userField := huh.NewInput().
		Key("user").
		Title("User").
		Value(&v.user)

	hostKeyField := huh.NewInput().
		Key("hostkeyalgorithms").
		Title("HostKeyAlgorithms").
		Value(&v.algHostKey)

	kexField := huh.NewInput().
		Key("kexalgorithms").
		Title("KexAlgorithms").
		Value(&v.algKex)

	macsField := huh.NewInput().
		Key("macs").
		Title("MACs").
		Value(&v.algMACs)

	greenEnter := lipgloss.NewStyle().Foreground(greenColor).Render("Enter")
	fields := []huh.Field{
		huh.NewNote().Title(title).Description("Enter host details and press " + greenEnter + " to save."),
	}
	if mode == modeAdd {
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

	form.CancelCmd = func() tea.Msg { return formResultMsg{kind: formResultCanceled} }
	form.SubmitCmd = func() tea.Msg {
		p := strings.ToLower(strings.TrimSpace(v.protocol))
		s := host.Spec{
			Alias:    strings.TrimSpace(v.alias),
			HostName: strings.TrimSpace(v.hostname),
			Port:     strings.TrimSpace(v.port),
			User:     strings.TrimSpace(v.user),
		}
		opts := sshopts.Options{
			HostKeyAlgorithms: strings.TrimSpace(v.algHostKey),
			KexAlgorithms:     strings.TrimSpace(v.algKex),
			MACs:              strings.TrimSpace(v.algMACs),
		}
		return formResultMsg{kind: formResultSubmitted, mode: mode, protocol: p,
			oldAlias: strings.TrimSpace(oldAlias), spec: s, opts: opts}
	}

	return form
}
