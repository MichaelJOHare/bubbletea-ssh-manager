package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const infoStatusTTL = 8 * time.Second

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

	// text input for SSH username prompt (only shown when needed)
	u := textinput.New()
	u.Placeholder = "username"
	u.Prompt = "\nUser: "
	u.Blur()

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
		delegate: d,
		root:     root,
		path:     path,
		lst:      lst,
	}

	m.initHelpKeys()
	m.setCurrentMenu(items)
	if seedErr != nil {
		m.setStatus("Config: "+lastNonEmptyLine(seedErr.Error()), true, 0)
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
		remaining := max(int(time.Until(m.preflightEndsAt).Round(time.Second).Seconds()), 0)
		m.setStatus(fmt.Sprintf("Checking %s %s (%ds)…", m.preflightProtocol, m.preflightHostPort, remaining), false, 0)
		if remaining > 0 {
			return m, preflightTickCmd(msg.token)
		}
		return m, nil

	case preflightResultMsg:
		if !m.preflighting || msg.token != m.preflightToken {
			return m, nil
		}
		protocol := m.preflightProtocol
		hostPort := m.preflightHostPort
		display := m.preflightDisplay
		windowTitle := m.preflightWindowTitle
		cmd := m.preflightCmd
		tail := m.preflightTail

		m.preflighting = false
		m.preflightEndsAt = time.Time{}
		m.preflightProtocol = ""
		m.preflightHostPort = ""
		m.preflightWindowTitle = ""
		m.preflightCmd = nil
		m.preflightTail = nil
		m.preflightDisplay = ""

		if msg.err != nil {
			statusCmd := m.setStatus(fmt.Sprintf("%s %s failed: \n%v", protocol, hostPort, msg.err), true, infoStatusTTL)
			return m, statusCmd
		}

		m.setStatus(fmt.Sprintf("Starting %s %s…", protocol, display), false, 0)

		return m, tea.Sequence(
			tea.SetWindowTitle(windowTitle),
			tea.ExecProcess(cmd, func(err error) tea.Msg {
				out := ""
				if tail != nil {
					out = strings.TrimSpace(tail.String())
					out = lastNonEmptyLine(out)
				}
				return connectFinishedMsg{protocol: protocol, target: display, err: err, output: out}
			}),
		)

	case connectFinishedMsg:
		titleCmd := tea.SetWindowTitle("MENU")
		output := strings.TrimSpace(msg.output)
		if msg.err != nil {
			if output != "" {
				statusCmd := m.setStatus(fmt.Sprintf("%s to %s exited:\n%s (%v)", msg.protocol, msg.target, output, msg.err), true, 0)
				return m, tea.Batch(titleCmd, statusCmd)
			}
			if isConnectionAborted(msg.err) {
				statusCmd := m.setStatus(fmt.Sprintf("%s to %s aborted.", msg.protocol, msg.target), true, infoStatusTTL)
				return m, tea.Batch(titleCmd, statusCmd)
			}
			statusCmd := m.setStatus(fmt.Sprintf("%s to %s exited - %v", msg.protocol, msg.target, msg.err), true, 0)
			return m, tea.Batch(titleCmd, statusCmd)
		}

		statusCmd := m.setStatus(fmt.Sprintf("%s to %s ended.", msg.protocol, msg.target), false, infoStatusTTL)
		return m, tea.Batch(titleCmd, statusCmd)

	// handle key presses
	case tea.KeyMsg:
		if nm, cmd, handled := m.handleKeyMsg(msg); handled {
			return nm, cmd
		}
	}

	// always update query input first
	var cmd1 tea.Cmd
	m.query, cmd1 = m.query.Update(msg)

	// reapply filter whenever the query changes
	m.applyFilter(m.query.Value())

	// then update list navigation
	var cmd2 tea.Cmd
	m.lst, cmd2 = m.lst.Update(msg)

	return m, tea.Batch(cmd1, cmd2)
}

// View renders the TUI components: list, status, hints, and search input.
func (m model) View() string {
	if m.quitting {
		return ""
	}

	statusColor := lipgloss.Color("243")
	if m.statusIsError {
		statusColor = lipgloss.Color("9")
	}
	statusStyle := lipgloss.NewStyle().Foreground(statusColor).PaddingLeft(footerPadLeft).PaddingTop(1)
	searchStyle := lipgloss.NewStyle().Bold(true).PaddingLeft(footerPadLeft)

	lines := []string{m.lst.View()}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusStyle.Render(m.status))
	}
	if m.promptingUser {
		lines = append(lines, searchStyle.Render(m.prompt.View()))
	} else {
		lines = append(lines, searchStyle.Render(m.query.View()))
	}
	return strings.Join(lines, "\n")
}
