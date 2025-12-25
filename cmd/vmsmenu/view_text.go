package main

import (
	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"

	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// hostDetailsText returns the detailed text view for the currently selected host.
//
// If no host is selected, it returns a placeholder message.
func (m model) hostDetailsText() string {
	it, _ := m.lst.SelectedItem().(*menuItem)
	if it == nil || it.kind != itemHost {
		return "Select a host to view details"
	}

	labelStyle := lipgloss.NewStyle().Foreground(fullHelpBorderColor).Bold(true).PaddingLeft(4)
	valueStyle := lipgloss.NewStyle().Foreground(statusColor)

	protocol := str.NormalizeString(it.protocol)
	protoColor := sshHostNameColor
	if protocol == "telnet" {
		protoColor = telnetHostNameColor
	}
	protoValueStyle := lipgloss.NewStyle().Foreground(protoColor).Bold(true)

	rows := make([][2]string, 0, 8)
	rows = append(rows, [2]string{"Protocol", protocol})
	rows = append(rows, [2]string{"Alias", strings.TrimSpace(it.spec.Alias)})
	rows = append(rows, [2]string{"HostName", strings.TrimSpace(it.spec.HostName)})
	rows = append(rows, [2]string{"Port", strings.TrimSpace(it.spec.Port)})
	rows = append(rows, [2]string{"User", strings.TrimSpace(it.spec.User)})

	maxLabelW := 0
	for _, r := range rows {
		maxLabelW = max(maxLabelW, lipgloss.Width(r[0]))
	}

	lines := make([]string, 0, 16)
	header := lipgloss.NewStyle().
		PaddingTop(1).
		PaddingLeft(2).
		Foreground(searchLabelColor).
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
		lines = append(lines, fmt.Sprintf("%s  %s", labelStyle.Render(label+":"), vRendered))
	}

	if protocol == "ssh" {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().
			PaddingTop(1).
			PaddingBottom(1).
			PaddingLeft(2).
			Foreground(searchLabelColor).
			Bold(true).
			Render("SSH OPTIONS"))

		optionsValueStyle := valueStyle.PaddingLeft(4)
		optionsLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ddb034")).Bold(true).PaddingLeft(4)
		if it.options.IsZero() {
			lines = append(lines, optionsValueStyle.Render("(none)"))
		} else {
			for _, line := range config.CreateSSHOptionsEntry(it.options, "") {
				if line == "" {
					continue
				}
				k, v, ok := strings.Cut(line, " ")
				if !ok || strings.TrimSpace(v) == "" {
					continue
				}
				lines = append(lines, fmt.Sprintf("%s: %s", optionsLabelStyle.Render(k), valueStyle.Render(v)))
			}
		}
		lines = append(lines, "") // extra padding at bottom
	}

	return strings.Join(lines, "\n")
}
