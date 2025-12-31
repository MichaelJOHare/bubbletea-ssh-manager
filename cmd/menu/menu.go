package main

import (
	"time"

	str "bubbletea-ssh-manager/internal/stringutil"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const statusTTL = 10 * time.Second // duration for non-error info statuses

// Init returns the initial command for the TUI (blinking cursor and window title).
func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("SSH Manager"), textinput.Blink)
}

// newModel creates a new TUI model with initial state and seeded menu items.
//
// It returns the initialized model.
func newModel() model {
	theme := DefaultTheme()
	keys := newKeyMap(theme)

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
	s.Style = lipgloss.NewStyle().Foreground(theme.PreflightSpinner)

	// seed menu and initial state
	root, seedErr := seedMenu()
	path := []*menuItem{root}
	items := root.children
	litems := toListItems(items)

	// setup list to display menu items
	d := newMenuDelegatePtr(theme)
	lst := list.New(litems, d, 0, 0)
	lst.InfiniteScrolling = true
	lst.Styles.TitleBar = lst.Styles.TitleBar.Padding(1, 0, 1, 1)
	lst.Styles.Title = lst.Styles.Title.Padding(0, 2)
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowHelp(true)

	// build initial bubbletea model
	m := model{
		theme:    theme,
		keys:     keys,
		query:    q,
		prompt:   u,
		spinner:  s,
		delegate: d,
		root:     root,
		path:     path,
		lst:      lst,
		mode:     modeMenu,
	}

	m.initHelpKeys()
	m.setCurrentMenu(items)
	if seedErr != nil {
		m.setStatusError("Config: "+str.LastNonEmptyLine(seedErr.Error()), 0)
	}
	return m
}

// Update handles incoming messages (ie. result of IO operations)
// and updates the model state accordingly.
//
// It returns the updated model and any command to be executed.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		nm, cmd, _ := m.handleWindowSizeMsg(v)
		return nm, cmd
	case menuReloadedMsg:
		nm, cmd, _ := m.handleMenuReloadedMsg(v)
		return nm, cmd
	case statusClearMsg:
		nm, cmd, _ := m.handleStatusClearMsg(v)
		return nm, cmd
	case spinner.TickMsg:
		nm, cmd, _ := m.handleSpinnerTickMsg(v)
		return nm, cmd
	case preflightTickMsg:
		nm, cmd, _ := m.handlePreflightTickMsg(v)
		return nm, cmd
	case preflightResultMsg:
		nm, cmd, _ := m.handlePreflightResultMsg(v)
		return nm, cmd
	case connectFinishedMsg:
		nm, cmd, _ := m.handleConnectFinishedMsg(v)
		return nm, cmd
	case tea.KeyMsg:
		if nm, cmd, handled := m.handleKeyMsg(v); handled {
			return nm, cmd
		}
	}

	// host form: handle lifecycle + non-key updates without needing special ordering
	if nm, cmd, handled := m.handleHostFormMsg(msg); handled {
		return nm, cmd
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

// View renders the TUI components (in order top to bottom):
//
//   - host form (if open)
//   - host details (if open)
//   - preflight status (if active)
//   - main menu list with status and search/prompt input
//
// It returns the complete string to be displayed.
func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.mode == modeExecuting {
		return ""
	}

	switch m.mode {
	case modeHostForm:
		return m.viewHostForm()
	case modeHostDetails:
		return m.viewHostDetails()
	case modePreflight:
		return m.viewPreflight()
	default:
		return m.viewMenu()
	}
}
