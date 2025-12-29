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

const statusTTL = 8 * time.Second // duration for non-error info statuses
const (
	grayStatusColor = lipgloss.Color("#bcbcbc") // default status color
	redStatusColor  = lipgloss.Color("#d03f3f") // error status color
	indigoColor     = lipgloss.Color("#736fe4") // indigo
	cyanColor       = lipgloss.Color("#0083b3") // cyan
	yellowColor     = lipgloss.Color("#dec532") // yellow
	brightPinkColor = lipgloss.Color("#ff0087") // pink
	greenColor      = lipgloss.Color("#6fc36f") // green
	pinkColor       = lipgloss.Color("#e15979") // pink
	orangeColor     = lipgloss.Color("#e48315") // orange
)

// Init returns the initial command for the TUI (blinking cursor and window title).
func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("MENU"), textinput.Blink)
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
	s.Style = lipgloss.NewStyle().Foreground(brightPinkColor)

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
// It handles window resize, connection completion, key presses, and updates
// to the text input and list components.
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
//   - list
//   - help (full or brief)
//   - status
//   - preflight spinner
//   - search input or prompt input
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
		return m.viewNormal()
	}
}
