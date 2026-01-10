package tui

import (
	"fmt"
	"strings"

	"bubbletea-ssh-manager/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// detailsStyles holds the styles used in the host details view.
type detailsStyles struct {
	header       lipgloss.Style
	label        lipgloss.Style
	value        lipgloss.Style
	protoSSH     lipgloss.Style
	protoTelnet  lipgloss.Style
	optionsLabel lipgloss.Style
	optionsValue lipgloss.Style
}

func (m model) newDetailsStyles() detailsStyles {
	return detailsStyles{
		header: lipgloss.NewStyle().
			PaddingTop(1).
			PaddingLeft(2).
			Foreground(m.theme.DetailsHeader).
			Bold(true),
		label:        lipgloss.NewStyle().Foreground(m.theme.DetailsLabel).PaddingLeft(4),
		value:        lipgloss.NewStyle().Foreground(m.theme.StatusDefault),
		protoSSH:     lipgloss.NewStyle().Foreground(m.theme.ProtocolSSH).Bold(true),
		protoTelnet:  lipgloss.NewStyle().Foreground(m.theme.ProtocolTelnet).Bold(true),
		optionsLabel: lipgloss.NewStyle().Foreground(m.theme.OptionsLabel).PaddingLeft(4),
		optionsValue: lipgloss.NewStyle().Foreground(m.theme.StatusDefault).PaddingLeft(4),
	}
}

// buildHostDetails returns the detailed text view for the currently selected host.
//
// If no host is selected, it returns a placeholder message.
func (m model) buildHostDetails() string {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		return "Select a host to view details"
	}

	s := m.newDetailsStyles()
	var b strings.Builder

	b.WriteString(s.header.Render("HOST DETAILS"))
	b.WriteString("\n\n")
	b.WriteString(m.buildHostInfo(it, s))

	if it.protocol == config.ProtocolSSH {
		b.WriteString("\n")
		b.WriteString(s.header.PaddingBottom(1).Render("SSH OPTIONS"))
		b.WriteString("\n")
		b.WriteString(m.buildSSHOptions(it, s))
		b.WriteString("\n")
	}

	return b.String()
}

// buildHostInfo renders the core host fields (protocol, alias, hostname, etc.).
func (m model) buildHostInfo(it *menuItem, s detailsStyles) string {
	rows := [][2]string{
		{"Protocol", string(it.protocol)},
		{"Alias", it.spec.Alias},
		{"HostName", it.spec.HostName},
		{"Port", it.spec.Port},
		{"User", it.spec.User},
	}

	maxLabelW := 0
	for _, r := range rows {
		maxLabelW = max(maxLabelW, lipgloss.Width(r[0]))
	}

	var b strings.Builder
	for _, r := range rows {
		label := s.label.Render(fmt.Sprintf("%*s", maxLabelW, r[0]))
		value := m.renderDetailValue(r[0], r[1], it.protocol, s)
		b.WriteString(fmt.Sprintf("%s:  %s\n", label, value))
	}
	return b.String()
}

// renderDetailValue renders a value with appropriate styling based on field name.
func (m model) renderDetailValue(field, value string, proto config.Protocol, s detailsStyles) string {
	if field == "Protocol" {
		if proto == config.ProtocolTelnet {
			return s.protoTelnet.Render(value)
		}
		return s.protoSSH.Render(value)
	}
	return s.value.Render(value)
}

// buildSSHOptions renders the SSH options section.
func (m model) buildSSHOptions(it *menuItem, s detailsStyles) string {
	if it.options.HostKeyAlgorithms == "" &&
		it.options.KexAlgorithms == "" &&
		it.options.MACs == "" {
		return s.optionsValue.Render("(none)")
	}

	var b strings.Builder
	for _, line := range config.BuildSSHOptions(it.options, "") {
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, " ")
		if !ok || v == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", s.optionsLabel.Render(k), s.value.Render(v)))
	}
	return b.String()
}
