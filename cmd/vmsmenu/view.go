package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewNormal renders the normal menu view with list, status, and search/prompt.
//
// It focuses the active menu item if prompting for username.
func (m model) viewNormal() string {
	statusColor := statusColor
	if m.statusIsError {
		statusColor = errorStatusColor
	}

	lg := lipgloss.NewStyle()
	statusPadStyle := lg.PaddingLeft(footerPadLeft).PaddingTop(1)
	statusTextStyle := lg.Foreground(statusColor)
	searchStyle := lg.Foreground(searchLabelColor).Bold(true).PaddingLeft(footerPadLeft)
	promptStyle := lg.Foreground(promptLabelColor).Bold(true).PaddingLeft(footerPadLeft)

	listView := m.lst.View()
	if m.promptingUsername {
		listView = m.setActiveMenuItem(listView)
	}

	lines := []string{listView}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusPadStyle.Render(statusTextStyle.Render(m.status)))
	}

	if m.promptingUsername {
		lines = append(lines, promptStyle.Render(m.prompt.View()))
	} else {
		lines = append(lines, searchStyle.Render(m.query.View()))
	}

	return strings.Join(lines, "\n")
}

// viewPreflight renders the preflight view with list and preflight status.
//
// It shows a spinner and countdown timer.
func (m model) viewPreflight() string {

	remaining := max(m.preflightRemaining, 0)
	preflightStatusText := fmt.Sprintf(
		"%s Checking %s %s (%ds)…\nctrl+c to cancel",
		m.spinner.View(),
		m.preflightProtocol,
		m.preflightHostPort,
		remaining,
	)

	lg := lipgloss.NewStyle()
	preflightPadStyle := lg.PaddingLeft(footerPadLeft + 4).PaddingBottom(3).PaddingTop(1)
	preflightTextStyle := lg.Foreground(lipgloss.Color("#8d8d8d"))

	lines := []string{m.lst.View()}
	lines = append(lines, preflightPadStyle.Render(preflightTextStyle.Render(preflightStatusText)))

	return strings.Join(lines, "\n")
}

// viewFullHelp renders the full help view with list, status, and full help text.
//
// It focuses the active menu item if host details aren't open.
func (m model) viewFullHelp() string {
	statusColor := statusColor
	if m.statusIsError {
		statusColor = errorStatusColor
	}

	lg := lipgloss.NewStyle()
	panelW := m.hostDetailsWidth()
	statusPadStyle := lg.PaddingLeft(footerPadLeft).PaddingTop(1)
	statusTextStyle := lg.Foreground(statusColor)
	fullHelpStyle := lg.
		Width(panelW).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(fullHelpBorderColor).
		PaddingLeft(footerPadLeft).
		PaddingRight(footerPadLeft)

	listView := m.lst.View()
	if m.hostDetailsOpen {
		listView = m.viewHostDetails()
	} else {
		listView = m.setActiveMenuItem(listView)
	}

	lines := []string{listView}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusPadStyle.Render(statusTextStyle.Render(m.status)))
	}

	fullHelpView := fullHelpStyle.Render(m.fullHelpText())
	if m.hostDetailsOpen {
		lines = append(lines, lipgloss.PlaceHorizontal(m.width, lipgloss.Center, fullHelpView))
	} else {
		lines = append(lines, fullHelpView)
	}

	return strings.Join(lines, "\n")
}

// viewHostDetails renders the host details box for the currently selected host.
//
// If no host is selected, it shows a placeholder message.
func (m model) viewHostDetails() string {
	lg := lipgloss.NewStyle()
	panelW := m.hostDetailsWidth()
	box := lg.
		Width(panelW).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(fullHelpBorderColor).
		PaddingLeft(1).
		PaddingRight(footerPadLeft).
		PaddingTop(1)

	detailsView := box.Render(m.hostDetailsText())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, detailsView)
}

// hostDetailsWidth returns the target width used for both the full help panel and the host details panel.
//
// The width is the larger of the two rendered panel widths (help vs host details), capped to the
// available terminal width so it doesn't overflow.
func (m model) hostDetailsWidth() int {
	availableW := max(0, m.width)
	if availableW == 0 {
		return 0
	}

	fullHelpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(fullHelpBorderColor).
		PaddingLeft(footerPadLeft).
		PaddingRight(footerPadLeft)

	hostDetailsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(fullHelpBorderColor).
		PaddingLeft(1).
		PaddingRight(footerPadLeft).
		PaddingTop(1)

	fullHelpW := lipgloss.Width(fullHelpStyle.Render(m.fullHelpText()))
	hostDetailsW := lipgloss.Width(hostDetailsStyle.Render(m.hostDetailsText()))
	return min(max(fullHelpW, hostDetailsW), availableW)
}
