package main

import (
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

type itemKind int

const (
	itemGroup itemKind = iota
	itemHost
)

type menuItem struct {
	kind itemKind // item type: group or host
	name string   // display name (host alias or group name)

	// host-only fields
	protocol string // "ssh" or "telnet"

	// alias is the ssh-style Host alias from the config
	// for SSH connections we connect by alias
	alias string

	// hostname and port come from HostName/Port directives
	// for Telnet connections we connect by hostname and a numeric port
	hostname string
	port     string

	// ssh-only fields
	// user comes from the SSH-style "User" directive
	user string

	// group-only fields
	children []*menuItem // child menu items
}

type hostWithGroup struct {
	host      *menuItem // host menu item
	groupPath string    // display group path
}

type model struct {
	width  int // window width
	height int // window height

	query         textinput.Model // search input box
	prompt        textinput.Model // generic prompt input (reused for username/addhost/etc)
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

	// preflight state: optional TCP reachability check before handing control to ssh/telnet
	preflighting         bool
	preflightToken       int
	preflightEndsAt      time.Time
	preflightProtocol    string
	preflightHostPort    string
	preflightWindowTitle string
	preflightCmd         *exec.Cmd
	preflightTail        *tailBuffer
	preflightDisplay     string
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
	token int
}

type preflightResultMsg struct {
	token int
	err   error
}

type tailBuffer struct {
	buf []byte // stored bytes
	max int    // max bytes to keep
}
