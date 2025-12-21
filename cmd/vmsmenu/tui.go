package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// newModel creates a new TUI model with initial state
// and seeded menu items.
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
	d := newMenuDelegate()
	lst := list.New(litems, d, 0, 0)
	lst.Title = "Hosts"
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowHelp(true)

	// build initial bubbletea model
	m := model{
		query: q,
		root:  root,
		path:  path,
		lst:   lst,
	}
	if seedErr != nil {
		m.statusIsError = true
		m.status = "Config: " + lastNonEmptyLine(seedErr.Error())
	}
	m.setCurrentMenu(items)
	m.relayout()
	return m
}

// Init returns the initial command for the TUI (blinking cursor).
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles incoming messages and updates the model state accordingly.
// It handles window resize, connection completion, key presses, and updates
// to the text input and list components.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// handle different message types
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.relayout()
		return m, nil

	// update status on connection finish
	case connectFinishedMsg:
		if msg.err != nil {
			m.statusIsError = true
			if strings.TrimSpace(msg.output) != "" {
				m.status = fmt.Sprintf("%s to %s exited: %s (%v)", msg.protocol, msg.target, msg.output, msg.err)
				m.relayout()
				return m, nil
			}
			m.status = fmt.Sprintf("%s to %s exited: %v", msg.protocol, msg.target, msg.err)
			m.relayout()
			return m, nil
		}
		m.statusIsError = false
		m.status = fmt.Sprintf("%s to %s ended.", msg.protocol, msg.target)
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
				m.statusIsError = false
				if it.kind == itemGroup {
					m.status = fmt.Sprintf("Group: %s (%d items)", it.name, len(it.children))
				} else {
					m.status = fmt.Sprintf("Host: %s (%s)", it.name, it.protocol)
				}
				m.relayout()
			}
			return m, nil
		// go back on Esc if in a group
		case "esc":
			if len(m.path) > 1 {
				m.path = m.path[:len(m.path)-1]
				m.query.SetValue("")
				m.setCurrentMenu(m.current().children)
				m.status = ""
				m.statusIsError = false
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
					m.status = ""
					m.statusIsError = false
					m.relayout()
					return m, nil
				}

				// when host selected, build and start connection command
				cmd, protocol, target, tail, err := buildConnectCommand(it)
				if err != nil {
					m.statusIsError = true
					m.status = err.Error()
					m.relayout()
					return m, nil
				}

				// handoff to ssh/telnet and return to TUI when done
				m.statusIsError = false
				m.status = fmt.Sprintf("Starting %s %s…", protocol, target)
				m.relayout()
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					out := strings.TrimSpace(tail.String())
					out = lastNonEmptyLine(out)
					return connectFinishedMsg{protocol: protocol, target: target, err: err, output: out}
				})
			}
			return m, nil
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

	searchStyle := lipgloss.NewStyle().Bold(true).PaddingLeft(footerPadLeft)
	statusColor := lipgloss.Color("243")
	if m.statusIsError {
		statusColor = lipgloss.Color("9")
	}
	statusStyle := lipgloss.NewStyle().Foreground(statusColor).PaddingLeft(footerPadLeft)

	lines := []string{m.lst.View()}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusStyle.Render(m.status))
	}
	lines = append(lines, searchStyle.Render(m.query.View()))
	return strings.Join(lines, "\n")
}
