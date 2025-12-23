package main

import (
	"bubbletea-ssh-manager/internal/connect"
	str "bubbletea-ssh-manager/internal/stringutil"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const statusTTL = 8 * time.Second // duration for non-error info statuses
const (
	statusColor         = lipgloss.Color("250") // default status color (gray)
	errorStatusColor    = lipgloss.Color("9")   // error status color (red)
	searchLabelColor    = lipgloss.Color("74")  // blue
	promptLabelColor    = lipgloss.Color("221") // yellow
	spinnerColor        = lipgloss.Color("198") // pink
	sshHostNameColor    = lipgloss.Color("10")  // green
	telnetHostNameColor = lipgloss.Color("210") // pink
	groupNameColor      = lipgloss.Color("208") // orange
)

// Init returns the initial command for the TUI (blinking cursor).
func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("MENU"), textinput.Blink)
}

// setStatus sets the status message and error flag.
//
// It increments the status token to keep track of which status to clear
// when using a duration. If d > 0, it returns a command to clear the status
// after the specified duration. If d == 0, the status remains until changed.
func (m *model) setStatus(text string, isError bool, d time.Duration) tea.Cmd {
	if d < 0 {
		d = 0
	}

	m.statusToken++
	m.status = text
	m.statusIsError = isError
	m.relayout()

	if d > 0 {
		tok := m.statusToken
		return tea.Tick(d, func(time.Time) tea.Msg {
			return statusClearMsg{token: tok}
		})
	}
	return nil
}

// newModel creates a new TUI model with initial state and seeded menu items.
//
// It also initializes the text input and list components.
func newModel() model {
	// text input for search query
	q := textinput.New()
	q.Placeholder = "type to search"
	q.Prompt = "\nSearch: "
	q.Focus()

	// text input for generic prompt (only shown when needed)
	u := textinput.New()
	u.Placeholder = "username"
	u.Prompt = "\nUser: "
	u.Blur()

	// spinner for preflight checks
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(spinnerColor)

	// seed menu and initial state
	root, seedErr := seedMenu()
	path := []*menuItem{root}
	items := root.children
	litems := toListItems(items)

	// setup list to display menu items
	d := newMenuDelegatePtr()
	lst := list.New(litems, d, 0, 0)
	lst.InfiniteScrolling = true
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowHelp(true)

	// build initial bubbletea model
	m := model{
		query:    q,
		prompt:   u,
		spinner:  s,
		delegate: d,
		root:     root,
		path:     path,
		lst:      lst,
	}

	m.initHelpKeys()
	m.setCurrentMenu(items)
	if seedErr != nil {
		m.setStatus("Config: "+str.LastNonEmptyLine(seedErr.Error()), true, 0)
	}
	return m
}

// Update handles incoming messages (ie. result of IO operations)
// and updates the model state accordingly.
//
// It handles window resize, connection completion, key presses, and updates
// to the text input and list components.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.relayout()
		return m, nil

	case spinner.TickMsg:
		// only animate the spinner during preflight
		if !m.preflighting {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	// handle status clear messages from temporary statuses
	case statusClearMsg:
		if msg.token == m.statusToken {
			m.status = ""
			m.statusIsError = false
			m.relayout()
		}
		return m, nil

	// handle preflight tick messages to update countdown
	case preflightTickMsg:
		if !m.preflighting || msg.token != m.preflightToken {
			return m, nil
		}
		// Keep this lightweight: update only the remaining seconds for display.
		remaining := max(int(time.Until(m.preflightEndsAt).Round(time.Second).Seconds()), 0)
		m.preflightRemaining = remaining
		if remaining > 0 {
			return m, preflightTickCmd(msg.token)
		}
		return m, nil

	// handle preflight result messages to start connection or show error
	case preflightResultMsg:
		if !m.preflighting || msg.token != m.preflightToken {
			return m, nil
		}

		// capture preflight state
		protocol := m.preflightProtocol
		hostPort := m.preflightHostPort
		display := m.preflightDisplay
		windowTitle := m.preflightWindowTitle
		cmd := m.preflightCmd
		tail := m.preflightTail

		// clear stored preflight state
		m.clearPreflightState()

		// if preflight failed, return error status
		if msg.err != nil {
			statusCmd := m.setStatus(fmt.Sprintf("%s %s failed: \n%v", protocol, hostPort, msg.err), true, statusTTL)
			return m, statusCmd
		}

		// preflight succeeded; start connection
		//m.executing = true
		return m, launchExecCmd(windowTitle, cmd, protocol, display, tail)

	// handle connection finished messages
	case connectFinishedMsg:
		//m.executing = false
		titleCmd := tea.SetWindowTitle("MENU")
		output := strings.TrimSpace(msg.output)
		if msg.err != nil {
			if output != "" {
				statusCmd := m.setStatus(fmt.Sprintf("%s to %s exited:\n%s (%v)", msg.protocol, msg.target, output, msg.err), true, 0)
				return m, tea.Batch(titleCmd, statusCmd)
			}
			if connect.IsConnectionAborted(msg.err) {
				statusCmd := m.setStatus(fmt.Sprintf("%s to %s aborted.", msg.protocol, msg.target), true, statusTTL)
				return m, tea.Batch(titleCmd, statusCmd)
			}
			statusCmd := m.setStatus(fmt.Sprintf("%s to %s exited:\n%v", msg.protocol, msg.target, msg.err), true, 0)
			return m, tea.Batch(titleCmd, statusCmd)
		}

		statusCmd := m.setStatus(fmt.Sprintf("%s to %s ended.", msg.protocol, msg.target), false, statusTTL)
		return m, tea.Batch(titleCmd, statusCmd)

	// handle key presses
	case tea.KeyMsg:
		if nm, cmd, handled := m.handleKeyMsg(msg); handled {
			return nm, cmd
		}
	}

	// always update query input first
	prevQuery := m.query.Value()
	var cmd1 tea.Cmd
	m.query, cmd1 = m.query.Update(msg)
	newQuery := m.query.Value()

	// reapply filter (and refresh help keys) only when the query changes
	if newQuery != prevQuery {
		m.applyFilter(newQuery)
		m.syncHelpKeys()
	}

	// then update list navigation
	var cmd2 tea.Cmd
	m.lst, cmd2 = m.lst.Update(msg)

	return m, tea.Batch(cmd1, cmd2)
}

// View renders the TUI components: list, status, hints, and search input.
//
// It returns the complete string to be displayed.
func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.executing {
		return ""
	}

	// determine status color
	statusColor := statusColor
	if m.statusIsError {
		statusColor = errorStatusColor
	}

	// set styles
	statusPadStyle := lipgloss.NewStyle().PaddingLeft(footerPadLeft).PaddingTop(1)
	statusTextStyle := lipgloss.NewStyle().Foreground(statusColor)
	searchStyle := lipgloss.NewStyle().Foreground(searchLabelColor).Bold(true).PaddingLeft(footerPadLeft)
	promptStyle := lipgloss.NewStyle().Foreground(promptLabelColor).Bold(true).PaddingLeft(footerPadLeft)

	// render status line
	lines := []string{m.lst.View()}
	if m.preflighting && !m.statusIsError {
		remaining := max(m.preflightRemaining, 0)
		line := fmt.Sprintf("%s Checking %s %s (%ds)…", m.spinner.View(), m.preflightProtocol, m.preflightHostPort, remaining)
		lines = append(lines, statusPadStyle.Render(statusTextStyle.Render(line)))
	}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusPadStyle.Render(statusTextStyle.Render(m.status)))
	}

	// render search or prompt input
	if m.promptingUser {
		lines = append(lines, promptStyle.Render(m.prompt.View()))
	} else {
		lines = append(lines, searchStyle.Render(m.query.View()))
	}

	return strings.Join(lines, "\n")
}
