package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newModel() model {
	q := textinput.New()
	q.Placeholder = "type to search"
	q.Prompt = "\nSearch: "
	q.Focus()

	root := seedMenu()
	path := []*menuItem{root}
	items := root.children
	litems := toListItems(items)

	d := list.NewDefaultDelegate()
	lst := list.New(litems, d, 0, 0)
	lst.Title = "Hosts"
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowHelp(true)

	m := model{
		query: q,
		root:  root,
		path:  path,
		lst:   lst,
	}
	m.setCurrentMenu(items)
	m.relayout()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.relayout()
		return m, nil

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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
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

				cmd, protocol, target, tail, err := buildConnectCommand(it)
				if err != nil {
					m.statusIsError = true
					m.status = err.Error()
					m.relayout()
					return m, nil
				}

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

func (m model) View() string {
	if m.quitting {
		return ""
	}

	searchStyle := lipgloss.NewStyle().Bold(true).PaddingLeft(footerPadLeft)
	statusColor := lipgloss.Color("241")
	if m.statusIsError {
		statusColor = lipgloss.Color("9")
	}
	statusStyle := lipgloss.NewStyle().Foreground(statusColor).PaddingLeft(footerPadLeft)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(footerPadLeft)

	lines := []string{m.lst.View()}
	if m.inGroup() {
		lines = append(lines, hintStyle.Render("Esc: back"))
	}
	if strings.TrimSpace(m.status) != "" {
		lines = append(lines, statusStyle.Render(m.status))
	}
	lines = append(lines, searchStyle.Render(m.query.View()))
	return strings.Join(lines, "\n")
}
