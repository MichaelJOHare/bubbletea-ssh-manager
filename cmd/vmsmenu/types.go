package main

import (
	"bubbletea-ssh-manager/internal/connect"
	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

type itemKind int // type of menu item : group or host

const (
	itemGroup itemKind = iota
	itemHost
)

type menuItem struct {
	// common fields
	kind itemKind // item type: group or host
	name string   // display name (host alias or group name)

	// host-only fields
	protocol string          // "ssh" or "telnet"
	spec     host.Spec       // shared host fields (alias/hostname/port/user)
	options  sshopts.Options // SSH options (only for SSH hosts)

	// group-only fields
	children []*menuItem // child menu items
}

type model struct {
	width  int // window width
	height int // window height

	query         textinput.Model // search input box
	prompt        textinput.Model // generic prompt input (reused for username/addhost/etc)
	spinner       spinner.Model   // spinner for preflight checks
	promptingUser bool            // whether we're currently prompting for a username
	pendingHost   *menuItem       // host waiting for username input
	delegate      *menuDelegate   // list delegate for rendering items

	root     *menuItem   // root menu item
	path     []*menuItem // current navigation path
	allItems []*menuItem // all items in the current menu
	lst      list.Model  // list of current menu items

	status        string // status message
	statusIsError bool   // is the status an error message?
	statusToken   int    // increments on status updates; tracked to clear status
	quitting      bool   // is the app quitting?
	executing     bool   // running external ssh/telnet session (blank the TUI)

	// preflight state: optional TCP reachability check before handing control to ssh/telnet
	preflighting         bool      // are we in a preflight check?
	preflightToken       int       // increments on preflight starts; for tick/result matching
	preflightRemaining   int       // remaining seconds in preflight (for display)
	preflightEndsAt      time.Time // when the preflight should end
	preflightProtocol    string    // "ssh" or "telnet"
	preflightHostPort    string    // host:port being checked
	preflightWindowTitle string    // original window title before preflight

	// preflight command and output capture
	preflightCmd     *exec.Cmd           // command being run for preflight
	preflightTail    *connect.TailBuffer // buffer for capturing command output
	preflightDisplay string              // display target (eg. host:port) for status messages
}

type statusClearMsg struct {
	token int // token to identify which status to clear
}

type connectFinishedMsg struct {
	protocol string // "ssh" or "telnet"
	target   string // display target (eg. host:port)
	err      error  // error from connection attempt
	output   string // output from ssh/telnet command
}

type preflightTickMsg struct {
	// should match model's preflightToken
	token int // token to identify which preflight to update
}

type preflightResultMsg struct {
	// should match model's preflightToken
	token int // token to identify which preflight to complete

	err error // error from preflight check
}
