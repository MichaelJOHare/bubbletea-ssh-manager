package main

import (
	"os/exec"
	"time"

	"bubbletea-ssh-manager/internal/connect"
	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
)

type uiMode int

const (
	modeMenu uiMode = iota
	modePromptUsername
	modeHostDetails
	modeHostForm
	modePreflight
	modeExecuting
)

type modeState struct {
	// prompt state
	pendingHost *menuItem

	// host add/edit state
	hostForm         *huh.Form // host add/edit form
	hostFormOldAlias string    // for edit/rename

	// preflight state
	preflightToken       int                 // increments on preflight starts; for tick/result matching
	preflightRemaining   int                 // remaining seconds in preflight (for display)
	preflightEndsAt      time.Time           // when the preflight should end
	preflightProtocol    string              // "ssh" or "telnet"
	preflightHostPort    string              // host:port being checked
	preflightWindowTitle string              // original window title before preflight
	preflightCmd         *exec.Cmd           // running preflight command
	preflightTail        *connect.TailBuffer // tail buffer for preflight output
	preflightDisplay     string              // display target (eg. host:port) for status messages
}

type model struct {
	width  int // window width
	height int // window height

	root     *menuItem       // root menu item
	path     []*menuItem     // current navigation path
	allItems []*menuItem     // all items in the current menu
	lst      list.Model      // list of current menu items
	delegate *menuDelegate   // list delegate for rendering items
	query    textinput.Model // search input box
	prompt   textinput.Model // generic prompt input (reused for username/addhost/etc)
	spinner  spinner.Model   // spinner for preflight checks

	mode uiMode    // current UI mode
	ms   modeState // current mode state

	status      string     // status message
	statusKind  statusKind // status style (info/success/error)
	statusToken int        // increments on status updates; tracked to clear status
	quitting    bool       // is the app quitting?
}

type formResultKind int // kind of result from host entry form (submit vs cancel)

const (
	formResultCancelled formResultKind = iota
	formResultSubmitted
)

/*
	MESSAGE TYPES
*/

type formResultMsg struct {
	kind formResultKind

	// set when kind==formResultSubmitted
	mode     formMode        // add vs edit mode for host entry form
	protocol string          // "ssh" or "telnet"
	oldAlias string          // for edit/rename
	spec     host.Spec       // shared host fields (alias/hostname/port/user)
	opts     sshopts.Options // SSH options (only for SSH hosts)
}

type formSaveResultMsg struct {
	err error // error during save IO operation
}

type menuReloadedMsg struct {
	root *menuItem // new root menu item
	err  error     // error during reload
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
	token int   // token to identify which preflight to complete
	err   error // error from preflight check
}
