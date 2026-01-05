package tui

import (
	"fmt"
	"strings"

	"bubbletea-ssh-manager/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// buildHostDetails returns the detailed text view for the currently selected host.
//
// If no host is selected, it returns a placeholder message.
func (m model) buildHostDetails() string {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		return "Select a host to view details"
	}

	labelStyle := lipgloss.NewStyle().Foreground(m.theme.DetailsLabel).PaddingLeft(4)
	valueStyle := lipgloss.NewStyle().Foreground(m.theme.StatusDefault)

	protocol := it.protocol
	protoColor := m.theme.ProtocolSSH
	if protocol == config.ProtocolTelnet {
		protoColor = m.theme.ProtocolTelnet
	}
	protoValueStyle := lipgloss.NewStyle().Foreground(protoColor).Bold(true)

	rows := make([][2]string, 0, 8)
	rows = append(rows, [2]string{"Protocol", string(protocol)})
	rows = append(rows, [2]string{"Alias", it.spec.Alias})
	rows = append(rows, [2]string{"HostName", it.spec.HostName})
	rows = append(rows, [2]string{"Port", it.spec.Port})
	rows = append(rows, [2]string{"User", it.spec.User})

	maxLabelW := 0
	for _, r := range rows {
		maxLabelW = max(maxLabelW, lipgloss.Width(r[0]))
	}

	lines := make([]string, 0, 16)
	header := lipgloss.NewStyle().
		PaddingTop(1).
		PaddingLeft(2).
		Foreground(m.theme.DetailsHeader).
		Bold(true).
		Render("HOST DETAILS")
	lines = append(lines, header)
	lines = append(lines, "")

	for _, r := range rows {
		label := fmt.Sprintf("%*s", maxLabelW, r[0])
		v := r[1]
		vRendered := valueStyle.Render(v)
		if r[0] == "Protocol" {
			vRendered = protoValueStyle.Render(v)
		}
		lines = append(lines, fmt.Sprintf("%s:  %s", labelStyle.Render(label), vRendered))
	}

	if protocol == config.ProtocolSSH {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().
			PaddingTop(1).
			PaddingBottom(1).
			PaddingLeft(2).
			Foreground(m.theme.DetailsHeader).
			Bold(true).
			Render("SSH OPTIONS"))

		optionsValueStyle := valueStyle.PaddingLeft(4)
		optionsLabelStyle := lipgloss.NewStyle().Foreground(m.theme.OptionsLabel).PaddingLeft(4)
		noOptionsPresent := it.options.HostKeyAlgorithms == "" &&
			it.options.KexAlgorithms == "" &&
			it.options.MACs == ""
		if noOptionsPresent {
			lines = append(lines, optionsValueStyle.Render("(none)"))
		} else {
			for _, line := range config.BuildSSHOptions(it.options, "") {
				if line == "" {
					continue
				}
				k, v, ok := strings.Cut(line, " ")
				if !ok || v == "" {
					continue
				}
				lines = append(lines, fmt.Sprintf("%s: %s", optionsLabelStyle.Render(k), valueStyle.Render(v)))
			}
		}
		lines = append(lines, "") // extra padding at bottom
	}

	return strings.Join(lines, "\n")
}
