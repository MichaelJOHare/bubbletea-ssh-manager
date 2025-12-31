package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewMenu renders the normal menu view with list, status, and search/prompt.
//
// It focuses the active menu item if prompting for username.
func (m model) viewMenu() string {
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

// viewHostDetails renders the host details modal with details + edit/remove help keys.
func (m model) viewHostDetails() string {
	lg := lipgloss.NewStyle()
	renderHelp := func(width int) string {
		h := m.lst.Help
		h.Width = max(0, width)
		return h.ShortHelpView(m.detailsHelpKeys())
	}

	detailsBox := lg.
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.theme.DetailsBorder).
		PaddingLeft(footerPadLeft).
		PaddingRight(footerPadLeft).
		PaddingTop(1)

	availableW := max(0, m.width)
	panelW := availableW
	if availableW > 0 {
		boxW := lipgloss.Width(detailsBox.Render(m.buildHostDetails()))
		helpW := lipgloss.Width(renderHelp(availableW))
		panelW = min(max(boxW, helpW), availableW)
	}
	details := detailsBox.Width(panelW)

	helpText := renderHelp(max(0, panelW))
	helpText = lg.Width(panelW).Align(lipgloss.Center).PaddingBottom(2).Render(helpText)
	helpH := lipgloss.Height(helpText)
	contentH := max(0, m.height-helpH)

	detailsView := details.Render(m.buildHostDetails())
	detailsView = lipgloss.Place(m.width, contentH, lipgloss.Center, lipgloss.Center, detailsView)
	helpView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, helpText)
	return strings.Join([]string{detailsView, helpView}, "\n")
}

// viewHostForm renders the host add/edit form centered in the terminal window.
//
// If no form is open, it returns an empty string.
func (m model) viewHostForm() string {
	if m.ms.hostForm == nil {
		return ""
	}

	formBox := lipgloss.NewStyle().
		PaddingRight(2).
		PaddingBottom(1)

	formContent := formBox.Render(m.ms.hostForm.View())
	panelW := lipgloss.Width(formContent)

	header := m.buildHostFormHeader()
	headerH := lipgloss.Height(header)

	footer := m.buildHostFormFooter(panelW)
	footerH := lipgloss.Height(footer)

	bodyH := max(0, m.height-headerH-footerH)

	right := m.buildFormStatusPanel(bodyH)

	leftW := max(0, m.width-hostFormStatusOuterWidth-hostFormStatusGap)
	left := lipgloss.NewStyle().Width(leftW).Height(bodyH).Render(formContent)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", hostFormStatusGap), right)

	return strings.Join([]string{header, body, footer}, "\n")
}
