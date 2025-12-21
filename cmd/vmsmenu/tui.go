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
	return textinput.Blink
}

// setStatus sets the status message and error flag.
//
// It increments the status token to invalidate any pending clears from
// a previous temporary status.
func (m *model) setStatus(text string, isError bool) {
	m.statusToken++
	m.status = text
	m.statusIsError = isError
}

// setTemporaryStatus sets a temporary status message that clears after duration d.
//
// It returns a command that will clear the status after the duration.
func (m *model) setTemporaryStatus(text string, isError bool, d time.Duration) tea.Cmd {
	m.setStatus(text, isError)
	tok := m.statusToken
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{token: tok}
	})
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

	// seed menu and initial state
	root, seedErr := seedMenu()
	path := []*menuItem{root}
	items := root.children
	litems := toListItems(items)

	// list to display menu items
	d := newMenuDelegatePtr()
	lst := list.New(litems, d, 0, 0)
	lst.Title = "Hosts"
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowHelp(true)

	// build initial bubbletea model
	m := model{
		query:    q,
		delegate: d,
		root:     root,
		path:     path,
		lst:      lst,
	}
	if seedErr != nil {
		m.setStatus("Config: "+lastNonEmptyLine(seedErr.Error()), true)
	}
	m.setCurrentMenu(items)
	m.relayout()
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

	case connectFinishedMsg:
		if msg.err != nil {
			// return error status if any
			if strings.TrimSpace(msg.output) != "" {
				m.setStatus(fmt.Sprintf("%s to %s exited - %s (%v)", msg.protocol, msg.target, msg.output, msg.err), true)
				m.relayout()
				return m, nil
			}
			// else generic error
			m.setStatus(fmt.Sprintf("%s to %s exited - %v", msg.protocol, msg.target, msg.err), true)
			m.relayout()
			return m, nil
		}
		// else success - show returned to TUI message
		m.setStatus(fmt.Sprintf("%s to %s ended.", msg.protocol, msg.target), false)
		m.relayout()
		return m, nil

	// handle key presses
	case tea.KeyMsg:
		switch msg.String() {
		// quit on Ctrl+C or 'q'
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		// show info on selected item with '?'
		case "?":
			if it, ok := m.lst.SelectedItem().(*menuItem); ok {
				if it.kind == itemGroup {
					cmd := m.setTemporaryStatus(fmt.Sprintf("Group: %s (%d items)", it.name, len(it.children)), false, infoStatusTTL)
					m.relayout()
					return m, cmd
				}
				if m.delegate != nil && m.delegate.groupHints != nil {
					if grp := strings.TrimSpace(m.delegate.groupHints[it]); grp != "" {
						cmd := m.setTemporaryStatus(fmt.Sprintf("Host: %s (%s) in %s", it.name, it.protocol, grp), false, infoStatusTTL)
						m.relayout()
						return m, cmd
					}
				}
				cmd := m.setTemporaryStatus(fmt.Sprintf("Host: %s (%s)", it.name, it.protocol), false, infoStatusTTL)
				m.relayout()
				return m, cmd
			}
		// clear search query if non-empty
		case "esc":
			if strings.TrimSpace(m.query.Value()) != "" {
				m.query.SetValue("")
				m.applyFilter("")
				m.relayout()
				return m, nil
			}
		// go back on left arrow if in a group
		case "left":
			if m.inGroup() {
				m.path = m.path[:len(m.path)-1]
				m.query.SetValue("")
				m.setCurrentMenu(m.current().children)
				m.setStatus("", false)
				m.relayout()
			}
			return m, nil
		// enter to navigate into group or connect to host
		case "enter":
			if it, ok := m.lst.SelectedItem().(*menuItem); ok {
				if it.kind == itemGroup {
					m.path = append(m.path, it)
					m.query.SetValue("")
					m.setCurrentMenu(it.children)
					m.setStatus("", false)
					m.relayout()
					return m, nil
				}
				cmd, protocol, target, tail, err := buildConnectCommand(it)
				if err != nil {
					m.setStatus(err.Error(), true)
					m.relayout()
					return m, nil
				}
				m.setStatus(fmt.Sprintf("Starting %s %s…", protocol, target), false)
				m.relayout()
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					out := strings.TrimSpace(tail.String())
					out = lastNonEmptyLine(out)
					return connectFinishedMsg{protocol: protocol, target: target, err: err, output: out}
				})
			}
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
	statusStyle := lipgloss.NewStyle().Foreground(statusColor).PaddingLeft(footerPadLeft)
	searchStyle := lipgloss.NewStyle().Bold(true).PaddingLeft(footerPadLeft)

	lines := []string{m.lst.View()}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusStyle.Render(m.status))
	}
	lines = append(lines, searchStyle.Render(m.query.View()))
	return strings.Join(lines, "\n")
}
