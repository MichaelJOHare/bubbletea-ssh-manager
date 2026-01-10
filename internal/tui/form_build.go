package tui

import (
	"strings"

	"bubbletea-ssh-manager/internal/config"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	hostFormStatusInnerWidth = 30                           // calculated based on content
	hostFormStatusOuterWidth = hostFormStatusInnerWidth + 2 // border left+right
	hostFormStatusGap        = 1                            // gap between form and status panel
)

// buildConfirmForm returns a generic confirmation form.
//
// The form sends a confirmResultMsg when completed (confirmed or canceled).
func buildConfirmForm(title string, description string, appTheme Theme) *huh.Form {
	var confirmed bool

	confirmField := huh.NewConfirm().
		Key("confirm").
		Title(title).
		Description(description).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed)

	form := huh.NewForm(huh.NewGroup(confirmField)).
		WithShowHelp(false).
		WithShowErrors(false).
		WithTheme(confirmFormTheme(appTheme))

	form.SubmitCmd = func() tea.Msg {
		return confirmResultMsg{
			confirmed: confirmed,
		}
	}
	form.CancelCmd = func() tea.Msg {
		return confirmResultMsg{
			confirmed: false,
		}
	}

	return form
}

// buildHostForm returns a Form which represents the data model for the host add/edit form.
//
// It holds the input values for the various fields.
func buildHostForm(mode formMode, oldAlias string, v *form, appTheme Theme) *huh.Form {
	if v == nil {
		v = &form{}
	}
	if mode == modeAdd && v.protocol == "" {
		v.protocol = config.ProtocolSSH
	}

	mainGroup := buildMainFieldGroup(mode, v)
	sshOptsGroup := buildSSHOptionsGroup(v)

	form := huh.NewForm(mainGroup, sshOptsGroup).
		WithShowHelp(false).
		WithShowErrors(false).
		WithKeyMap(NewFormKeyMap()).
		WithTheme(hostFormTheme(appTheme))

	form.CancelCmd = func() tea.Msg { return formCanceledMsg{} }
	form.SubmitCmd = buildSubmitCmd(mode, oldAlias, v)

	return form
}

// buildMainFieldGroup creates the primary host fields group.
func buildMainFieldGroup(mode formMode, v *form) *huh.Group {
	note := huh.NewNote().
		Description("Enter host details and press " + GreenEnter() + " to save.")

	fields := []huh.Field{note}
	if mode == modeAdd {
		fields = append(fields, buildProtocolField(v))
	}
	fields = append(fields,
		buildInputField("group", "Group", &v.groupName),
		buildInputField("nickname", "Nickname", &v.nickname),
		buildInputField("hostname", "Hostname", &v.hostname),
		buildInputField("port", "Port", &v.port),
		buildInputField("user", "User", &v.user),
	)
	return huh.NewGroup(fields...)
}

// buildProtocolField creates the protocol selector field.
func buildProtocolField(v *form) *huh.Select[config.Protocol] {
	return huh.NewSelect[config.Protocol]().
		Key("protocol").
		Title("Protocol").
		Options(
			huh.NewOption("ssh", config.ProtocolSSH),
			huh.NewOption("telnet", config.ProtocolTelnet),
		).
		Value(&v.protocol)
}

// buildInputField creates a simple Huh text input field.
func buildInputField(key, title string, value *string) *huh.Input {
	return huh.NewInput().
		Key(key).
		Title(title).
		Value(value)
}

// buildSSHOptionsGroup creates the SSH-specific options Huh group.
func buildSSHOptionsGroup(v *form) *huh.Group {
	note := huh.NewNote().Description(sshOptionsHelpText())

	return huh.NewGroup(
		note,
		buildInputField("hostkeyalgorithms", "HostKeyAlgorithms", &v.sshOpts.HostKeyAlgorithms),
		buildInputField("kexalgorithms", "KexAlgorithms", &v.sshOpts.KexAlgorithms),
		buildInputField("macs", "MACs", &v.sshOpts.MACs),
	).WithHideFunc(func() bool {
		return v.protocol != config.ProtocolSSH
	})
}

// sshOptionsHelpText returns the help text for the SSH options group.
func sshOptionsHelpText() string {
	lines := []string{
		"Optional SSH settings. Leave blank to use defaults. Press " + GreenEnter() + " to save.",
		"Prefix options with a " + GreenPlus() + " to append, " + RedMinus() + " to remove, " + PurpleCaret() + " to prepend.",
		"",
		"_It's generally recommended to append to defaults rather than override them.",
		"Multiple algorithms can be comma separated. See ssh config man page for details.",
	}
	return strings.Join(lines, "\n")
}

// buildSubmitCmd creates the form submit command handler.
func buildSubmitCmd(mode formMode, oldAlias string, v *form) func() tea.Msg {
	return func() tea.Msg {
		p := v.protocol
		if p == "" {
			p = config.ProtocolSSH
		}

		spec := config.Spec{
			HostName: v.hostname,
			Port:     v.port,
			User:     v.user,
		}

		opts := config.SSHOptions{}
		if p == config.ProtocolSSH {
			opts = v.sshOpts
		}

		return formSubmittedMsg{
			mode:     mode,
			protocol: p,
			oldAlias: oldAlias,
			group:    v.groupName,
			nickname: v.nickname,
			spec:     spec,
			opts:     opts,
		}
	}
}

// buildHostFormHeader builds the host form header boundary and returns the
// computed values used elsewhere (for the status panel, etc.).
func (m model) buildHostFormHeader() (header string) {
	action := "Adding"
	if strings.TrimSpace(m.ms.hostFormOldAlias) != "" || m.ms.hostFormMode == modeEdit {
		action = "Editing"
	}

	protocol := m.hostFormProtocol()

	configPath := "(unknown)"
	if action == "Adding" {
		if p, err := config.GetConfigPathForProtocol(protocol); err == nil && strings.TrimSpace(p) != "" {
			configPath = p
		}
	} else {
		oldAlias := strings.TrimSpace(m.ms.hostFormOldAlias)
		if p, err := config.GetConfigPathForAlias(protocol, oldAlias); err == nil && strings.TrimSpace(p) != "" {
			configPath = p
		}
	}

	text := action + " Host in " + configPath
	headerText := lipgloss.NewStyle().
		Foreground(m.theme.ProtocolTelnet).
		Padding(0, 2, 0, 2)
	header = lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		headerText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(m.theme.SearchLabel),
	)
	return header
}

// buildHostFormFooter builds the host form footer boundary containing the
// current short help line.
func (m model) buildHostFormFooter(panelW int) string {
	h := m.lst.Help
	h.Width = max(0, panelW)

	enterBinding := m.keys.FormSubmit
	if m.ms.hostForm != nil {
		if f := m.ms.hostForm.GetFocusedField(); f != nil {
			if _, ok := f.(*huh.Select[config.Protocol]); ok {
				enterBinding = m.keys.FormSelect
			}
		}
	}

	bindings := append(m.formHelpKeys(), enterBinding)
	helpText := h.ShortHelpView(bindings)

	pad := lipgloss.NewStyle().Padding(0, 2, 0, 2)
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		pad.Render(helpText),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(m.theme.SearchLabel),
	)
}

// buildHostFormPaginator builds the paginator view for the host form.
//
// It only shows when the protocol is "ssh" and there are multiple pages
// (SSH options).
func (m model) buildHostFormPaginator() string {
	if m.ms.hostForm == nil {
		return ""
	}
	if m.hostFormProtocol() != config.ProtocolSSH {
		return ""
	}

	page := 0
	if f := m.ms.hostForm.GetFocusedField(); f != nil {
		switch f.GetKey() {
		case "hostkeyalgorithms", "kexalgorithms", "macs":
			page = 1
		}
	}

	p := paginator.New(paginator.WithPerPage(1), paginator.WithTotalPages(2))
	p.Type = paginator.Dots
	p.Page = page

	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("•")

	return lipgloss.NewStyle().Foreground(m.theme.StatusDefault).Render(p.View())
}
