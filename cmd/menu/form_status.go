package main

import (
	"strings"

	str "bubbletea-ssh-manager/internal/stringutil"

	"github.com/charmbracelet/lipgloss"
)

type formStatusData struct {
	protocol  string // "ssh" or "telnet"
	groupName string // will add to existing group if present, else create new group
	nickname  string // if no group, nickname is the alias and will appear at root level
	hostname  string // hostname or IP - for telnet it's required
	port      string // port number - for telnet it's required if not default
	user      string // optional user name

	groupErr    error // validation error
	nicknameErr error // validation error
	hostErr     error // validation error
	portErr     error // validation error
}

type formStatusRenderers struct {
	label        func(...string) string // label renderer
	valueDefault func(...string) string // default value renderer
	valueSuccess func(...string) string // success value renderer
	valueError   func(...string) string // error value renderer
	errText      func(...string) string // error text renderer
	head         string                 // header text renderer

	successCheck string // success suffix
	errorX       string // error prefix
}

// newFormStatusRenderers creates form status renderers based on the app theme.
func newFormStatusRenderers(theme Theme) formStatusRenderers {
	return formStatusRenderers{
		label:        lipgloss.NewStyle().Foreground(theme.StatusDefault).Render,
		valueDefault: lipgloss.NewStyle().Foreground(theme.StatusDefault).Render,
		valueSuccess: lipgloss.NewStyle().Foreground(theme.StatusSuccess).Render,
		valueError:   lipgloss.NewStyle().Foreground(theme.StatusError).Render,
		errText:      lipgloss.NewStyle().Foreground(theme.StatusError).Render,
		head:         lipgloss.NewStyle().Foreground(theme.ProtocolSSH).Bold(true).Render("Current Settings"),

		successCheck: " ✔️",
		errorX:       "❌ ",
	}
}

// formatValidated formats a value based on its validation error.
//
// Behavior:
//   - If the value is empty, it uses the default style.
//   - If there is a validation error, it uses the error style and prefixes with an X.
//   - If valid, it uses the success style and suffixes with a checkmark.
func (r formStatusRenderers) formatValidated(s string, err error) string {
	if strings.TrimSpace(s) == "" {
		return r.valueDefault(s)
	}
	if err != nil {
		return r.errorX + r.valueError(s)
	}
	return r.valueSuccess(s) + r.successCheck
}

// formatUnvalidated formats a value that does not require validation (e.g., user).
func (r formStatusRenderers) formatUnvalidated(s string) string {
	if strings.TrimSpace(s) == "" {
		return r.valueDefault(s)
	}
	return r.valueSuccess(s)
}

// hostFormProtocol returns the protocol for the current host form.
//
// It checks the live form value in add mode.
func (m model) hostFormProtocol() string {
	protocol := strings.ToLower(strings.TrimSpace(m.ms.hostFormProtocol))
	if protocol == "" {
		protocol = "ssh"
	}
	if m.ms.hostFormMode == modeAdd && m.ms.hostForm != nil {
		if p := strings.ToLower(strings.TrimSpace(m.ms.hostForm.GetString("protocol"))); p != "" {
			protocol = p
		}
	}
	return protocol
}

// formStatusData collects the current form input values and their validation errors.
//
// It is used to build the form status panel.
func (m model) formStatusData() formStatusData {
	protocol := m.hostFormProtocol()
	groupName := strings.TrimSpace(m.ms.hostForm.GetString("group"))
	nickname := strings.TrimSpace(m.ms.hostForm.GetString("nickname"))
	hostname := strings.TrimSpace(m.ms.hostForm.GetString("hostname"))
	port := strings.TrimSpace(m.ms.hostForm.GetString("port"))
	user := strings.TrimSpace(m.ms.hostForm.GetString("user"))

	return formStatusData{
		protocol:  protocol,
		groupName: groupName,
		nickname:  nickname,
		hostname:  hostname,
		port:      port,
		user:      user,

		groupErr:    str.ValidateHostGroup(groupName),
		nicknameErr: str.ValidateHostNickname(nickname),
		hostErr:     str.ValidateHostName(protocol, hostname),
		portErr:     str.ValidateHostPort(protocol, port),
	}
}

// buildFormStatusLines builds the lines for the form status panel.
func buildFormStatusLines(r formStatusRenderers, d formStatusData) []string {
	lines := []string{r.head, ""}

	lines = append(lines, r.label("Group: ")+r.formatValidated(d.groupName, d.groupErr))
	if d.groupErr != nil {
		lines = append(lines, r.errText(strings.TrimSpace(d.groupErr.Error())))
	}

	lines = append(lines, r.label("\nNick: ")+r.formatValidated(d.nickname, d.nicknameErr))
	if d.nicknameErr != nil {
		lines = append(lines, r.errText(strings.TrimSpace(d.nicknameErr.Error())))
	}

	lines = append(lines, r.label("\nHostName: ")+r.formatValidated(d.hostname, d.hostErr))
	if d.hostErr != nil {
		lines = append(lines, r.errText(strings.TrimSpace(d.hostErr.Error())))
	}

	lines = append(lines, r.label("\nPort: ")+r.formatValidated(d.port, d.portErr))
	if d.portErr != nil {
		lines = append(lines, r.errText(strings.TrimSpace(d.portErr.Error())))
	}

	lines = append(lines, r.label("\nUser: ")+r.formatUnvalidated(d.user))
	return lines
}

// buildFormStatusPanel builds the form status panel view.
//
// It uses the current form data and theme to render the panel.
func (m model) buildFormStatusPanel(bodyH int) string {
	r := newFormStatusRenderers(m.theme)
	d := m.formStatusData()
	lines := buildFormStatusLines(r, d)
	statusBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.DetailsBorder).
		Padding(0, 1).
		Width(hostFormStatusInnerWidth).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(hostFormStatusOuterWidth, bodyH, lipgloss.Center, lipgloss.Center, statusBox)
}
