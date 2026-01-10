package tui

import (
	"strconv"
	"strings"

	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"

	"github.com/charmbracelet/lipgloss"
)

type formStatusData struct {
	protocol  config.Protocol // current protocol (ssh/telnet)
	groupName string          // will add to existing group if present, else create new group
	nickname  string          // if no group, nickname is the alias and will appear at root level
	hostname  string          // hostname or IP - for telnet it's required
	port      string          // port number - for telnet it's required if not default
	user      string          // optional user name

	existingGroups []string // existing group names

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
	groupValue   func(...string) string // group name renderer
	head         func(...string) string // header text renderer
}

// newFormStatusRenderers creates form status renderers based on the app theme.
func newFormStatusRenderers(theme Theme) formStatusRenderers {
	return formStatusRenderers{
		label:        lipgloss.NewStyle().Foreground(theme.StatusDefault).Render,
		valueDefault: lipgloss.NewStyle().Foreground(theme.StatusDefault).Render,
		valueSuccess: lipgloss.NewStyle().Foreground(theme.StatusSuccess).Render,
		valueError:   lipgloss.NewStyle().Foreground(theme.StatusError).Render,
		errText:      lipgloss.NewStyle().Foreground(theme.StatusError).Render,
		groupValue:   lipgloss.NewStyle().Foreground(theme.GroupName).Render,
		head:         lipgloss.NewStyle().Foreground(theme.ProtocolSSH).Bold(true).Render,
	}
}

// formatValidated formats a value based on its validation error.
//
// Behavior:
//   - If the value is empty, it uses the default style.
//   - If there is a validation error, it uses the error style and prefixes with an X.
//   - If valid, it uses the success style and suffixes with a checkmark.
func (r formStatusRenderers) formatValidated(s string, err error) string {
	return r.formatValue(s, err, nil, r.valueSuccess, true)
}

// formatUnvalidated formats a value that does not require validation (e.g., user).
func (r formStatusRenderers) formatUnvalidated(s string) string {
	if strings.TrimSpace(s) == "" {
		return r.valueDefault(s)
	}
	return r.valueSuccess(s)
}

// formatValue is a generic formatter for values that may have a validation error.
//
// Behavior:
//   - If there is a validation error, it uses the error style and prefixes with an X.
//   - If empty, it uses the default style.
//   - If valid, it applies transform (if provided), uses render, and optionally suffixes a checkmark.
func (r formStatusRenderers) formatValue(
	s string,
	err error,
	transform func(string) string,
	render func(...string) string,
	withCheck bool,
) string {
	if err != nil {
		return ErrorX + r.errText(strings.TrimSpace(err.Error()))
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return r.valueDefault(s)
	}

	if transform != nil {
		s = transform(s)
	}

	out := render(s)
	if withCheck {
		out += SuccessCheck
	}
	return out
}

// formatGroupName formats the group name with validation error handling.
func (r formStatusRenderers) formatGroupName(s string, err error) string {
	return r.formatValue(s, err, strings.ToUpper, r.groupValue, true)
}

// formatNickname formats the nickname with validation error handling.
func (r formStatusRenderers) formatNickname(s string, err error) string {
	return r.formatValue(s, err, strings.ToLower, r.valueSuccess, true)
}

// hostFormProtocol returns the protocol for the current host form.
//
// It checks the live form value in add mode.
func (m model) hostFormProtocol() config.Protocol {
	protocol := config.ProtocolSSH
	if m.ms.hostFormValues != nil {
		if p := m.ms.hostFormValues.protocol; p != "" {
			protocol = p
		}
	}
	return protocol
}

// existingRootGroupNames returns the list of existing root-level group names.
//
// Used to help users add to existing groups when adding/editing hosts.
func (m model) existingRootGroupNames() []string {
	if m.root == nil {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, it := range m.root.children {
		if it == nil || it.kind != itemGroup {
			continue
		}
		name := strings.ToUpper(strings.TrimSpace(it.name))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// formStatusData collects the current form input values and their validation errors.
//
// It is used to build the form status panel.
func (m model) formStatusData() formStatusData {
	protocol := m.hostFormProtocol()

	groupName := ""
	nickname := ""
	hostname := ""
	port := ""
	user := ""

	// get live form values if available
	if m.ms.hostFormValues != nil {
		groupName = strings.TrimSpace(m.ms.hostFormValues.groupName)
		nickname = strings.TrimSpace(m.ms.hostFormValues.nickname)
		hostname = strings.TrimSpace(m.ms.hostFormValues.hostname)
		port = strings.TrimSpace(m.ms.hostFormValues.port)
		user = strings.TrimSpace(m.ms.hostFormValues.user)
	}

	// for display, show the protocol's default port when empty (ssh=22, telnet=23)
	// this keeps the status panel synced with the *effective* config
	portDisplay := port
	portErr := error(nil)
	if p, err := str.NormalizePort(port, protocol); err != nil {
		portErr = err
	} else {
		portDisplay = p
	}

	return formStatusData{
		protocol:       protocol,
		groupName:      groupName,
		nickname:       nickname,
		hostname:       hostname,
		port:           portDisplay,
		user:           user,
		existingGroups: m.existingRootGroupNames(),

		groupErr:    str.ValidateHostGroup(groupName),
		nicknameErr: str.ValidateHostNickname(nickname),
		hostErr:     str.ValidateHostName(protocol, hostname),
		portErr:     portErr,
	}
}

// buildFormStatusLines builds the lines for the form status panel.
func buildFormStatusLines(r formStatusRenderers, d formStatusData) []string {
	lines := []string{r.head("Current Settings"), ""}

	lines = append(lines, r.label("Group: ")+r.formatGroupName(d.groupName, d.groupErr))

	lines = append(lines, r.label("\nNick: ")+r.formatNickname(d.nickname, d.nicknameErr))

	lines = append(lines, r.label("\nHostName: ")+r.formatValidated(d.hostname, d.hostErr))

	lines = append(lines, r.label("\nPort: ")+r.formatValidated(d.port, d.portErr))

	lines = append(lines, r.label("\nUser: ")+r.formatUnvalidated(d.user))

	// show existing groups if any
	if len(d.existingGroups) > 0 {
		lines = append(lines, r.label("\n\nCurrent Groups:"))
		for i, g := range d.existingGroups {
			// limit to first 6 groups      **** maybe make a help option when on group field to show all? ****
			if i >= 6 {
				remaining := len(d.existingGroups) - i
				lines = append(lines, r.label("  ")+r.valueDefault("+"+strconv.Itoa(remaining)+" more"))
				break
			}
			lines = append(lines, r.label("  ")+r.groupValue(g))
		}
	}

	return lines
}

// hasFormValidationErrors checks if the current form has any validation errors.
//
// This is used to prevent form submission when there are validation errors.
func (m model) hasFormValidationErrors() bool {
	d := m.formStatusData()
	return d.groupErr != nil || d.nicknameErr != nil || d.hostErr != nil || d.portErr != nil
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
