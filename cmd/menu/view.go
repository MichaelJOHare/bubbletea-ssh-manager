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
	statusColor := m.theme.StatusDefault
	switch m.statusKind {
	case statusError:
		statusColor = m.theme.StatusError
	case statusSuccess:
		statusColor = m.theme.StatusSuccess
	}

	lg := lipgloss.NewStyle()
	statusPadStyle := lg.PaddingLeft(footerPadLeft).PaddingTop(1)
	statusTextStyle := lg.Foreground(statusColor)
	searchStyle := lg.Foreground(m.theme.SearchLabel).Bold(true).PaddingLeft(footerPadLeft)
	promptStyle := lg.Foreground(m.theme.UsernamePrompt).Bold(true).PaddingLeft(footerPadLeft)

	listView := m.lst.View()
	if m.mode == modePromptUsername {
		listView = m.setActiveMenuItem(listView)
	}

	lines := []string{listView}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusPadStyle.Render(statusTextStyle.Render(m.status)))
	}

	if m.mode == modePromptUsername {
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

	remaining := max(m.ms.preflightRemaining, 0)
	preflightStatusText := fmt.Sprintf(
		"%s Checking %s %s (%ds)…\nctrl+c to cancel",
		m.spinner.View(),
		m.ms.preflightProtocol,
		m.ms.preflightHostPort,
		remaining,
	)

	lg := lipgloss.NewStyle()
	preflightPadStyle := lg.PaddingLeft(footerPadLeft + 4).PaddingBottom(3).PaddingTop(1)
	preflightTextStyle := lg.Foreground(m.theme.PreflightText)

	lines := []string{m.lst.View()}
	lines = append(lines, preflightPadStyle.Render(preflightTextStyle.Render(preflightStatusText)))

	return strings.Join(lines, "\n")
}

// viewHostDetails renders the host details modal with details + CRUD help keys.
func (m model) viewHostDetails() string {

	lg := lipgloss.NewStyle()
	panelW := m.hostDetailsWidth()
	hostBox := lg.
		Width(panelW).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.theme.DetailsBorder).
		PaddingLeft(footerPadLeft).
		PaddingRight(footerPadLeft).
		PaddingTop(1)
	helpStyle := lg.
		Width(panelW).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.theme.DetailsBorder).
		PaddingLeft(footerPadLeft).
		PaddingRight(footerPadLeft).
		Align(lipgloss.Center)

	detailsView := hostBox.Render(m.hostDetailsText())
	detailsView = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, detailsView)
	// render help at the inner content width so it can be centered inside the padded box
	innerHelpW := max(0, panelW-2-(footerPadLeft*2))
	helpView := helpStyle.Render(m.detailsHelpText(innerHelpW))
	lines := []string{detailsView}
	lines = append(lines, lipgloss.PlaceHorizontal(m.width, lipgloss.Center, helpView))

	return strings.Join(lines, "\n")
}

// viewHostForm renders the host add/edit form centered in the terminal window.
//
// If no form is open, it returns an empty string.
func (m model) viewHostForm() string {
	if m.ms.hostForm == nil {
		return ""
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.theme.DetailsBorder).
		PaddingLeft(2).
		PaddingRight(2).
		PaddingTop(1).
		PaddingBottom(1)

	content := box.Render(m.ms.hostForm.View())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
