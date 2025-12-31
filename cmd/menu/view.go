package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// viewCenteredModal renders 1-2 boxed panels centered vertically, plus a short help line.
//
// If secondaryContent is empty, only the primary box is rendered.
func (m model) viewCenteredModal(
	primaryBox lipgloss.Style,
	primaryContent string,
	secondaryBox lipgloss.Style,
	secondaryContent string,
	helpKeys []key.Binding,
) string {
	lg := lipgloss.NewStyle()
	renderHelp := func(width int) string {
		h := m.lst.Help
		h.Width = max(0, width)
		return h.ShortHelpView(helpKeys)
	}

	availableW := max(0, m.width)
	panelW := availableW
	if availableW > 0 {
		boxW := lipgloss.Width(primaryBox.Render(primaryContent))
		helpW := lipgloss.Width(renderHelp(availableW))
		panelW = min(max(boxW, helpW), availableW)
		if strings.TrimSpace(secondaryContent) != "" {
			secondaryW := lipgloss.Width(secondaryBox.Render(secondaryContent))
			panelW = min(max(panelW, secondaryW), availableW)
		}
	}

	primary := primaryBox.Width(panelW)
	helpText := renderHelp(max(0, panelW))
	helpText = lg.Width(panelW).Align(lipgloss.Center).PaddingBottom(2).Render(helpText)
	helpH := lipgloss.Height(helpText)
	contentH := max(0, m.height-helpH)

	primaryRendered := primary.Render(primaryContent)
	stacked := primaryRendered
	if strings.TrimSpace(secondaryContent) != "" {
		secondary := secondaryBox.Width(panelW)
		secondaryRendered := secondary.Render(secondaryContent)
		stacked = lipgloss.JoinVertical(lipgloss.Center, primaryRendered, secondaryRendered)
	}

	stackedView := lipgloss.Place(m.width, contentH, lipgloss.Center, lipgloss.Center, stacked)
	helpView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, helpText)
	return strings.Join([]string{stackedView, helpView}, "\n")
}

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
	detailsBox := lg.
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.theme.DetailsBorder).
		PaddingLeft(footerPadLeft).
		PaddingRight(footerPadLeft).
		PaddingTop(1)

	// secondary box is unused here but required by helper signature.
	unusedSecondary := lg
	return m.viewCenteredModal(detailsBox, m.buildHostDetails(), unusedSecondary, "", m.detailsHelpKeys())
}

// viewConfirm renders a confirmation dialog.
//
// Today this is used under host details; future confirmations (e.g. form save/cancel)
// can reuse the same layout by changing the primary content.
func (m model) viewConfirm() string {
	lg := lipgloss.NewStyle()
	primaryBox := lg.
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.theme.DetailsBorder).
		PaddingLeft(footerPadLeft).
		PaddingRight(footerPadLeft).
		PaddingTop(1)

	// render confirm form (if present)
	confirmContent := m.ms.confirmForm.View()
	confirmBox := lg.
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.theme.StatusError).
		Align(lipgloss.Center)

	// For now, confirm is shown beneath host details.
	// When you add confirmations in other contexts, swap this to the appropriate base view.
	primaryContent := m.buildHostDetails()
	return m.viewCenteredModal(primaryBox, primaryContent, confirmBox, confirmContent, m.confirmHelpKeys())
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
