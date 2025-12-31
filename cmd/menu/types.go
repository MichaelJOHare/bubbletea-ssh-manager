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

type modeState struct {
	// prompt state
	pendingHost *menuItem

	// host add/edit state
	hostForm         *huh.Form // host add/edit form
	hostFormMode     formMode  // add vs edit
	hostFormProtocol string    // "ssh" or "telnet" (used for header/config path)
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

type uiMode int

const (
	modeMenu uiMode = iota
	modePromptUsername
	modeHostDetails
	modeHostForm
	modePreflight
	modeExecuting
)

type model struct {
	width  int // window width
	height int // window height

	theme Theme  // active UI theme
	keys  KeyMap // active key mappings

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

type formMode int // add vs edit mode for host entry form

const (
	modeAdd formMode = iota
	modeEdit
)

type form struct {
	protocol   string // "ssh" or "telnet"
	groupName  string // group name portion of alias (display form; spaces allowed)
	nickname   string // host nickname portion of alias (display form; spaces allowed)
	hostname   string // hostname or IP address
	port       string // port number as string
	user       string // user name
	algHostKey string // host key algorithms
	algKex     string // key exchange algorithms
	algMACs    string // MAC algorithms
}

/*
	MESSAGE TYPES
*/

const (
	formResultCanceled formResultKind = iota
	formResultSubmitted
)

type formResultKind int // kind of result from host entry form (submit vs cancel)
type formResultMsg struct {
	kind formResultKind // kind of result (canceled vs submitted)

	// set when kind==formResultSubmitted
	mode     formMode        // add vs edit mode for host entry form
	protocol string          // "ssh" or "telnet"
	oldAlias string          // for edit/rename
	group    string          // group name (display form)
	nickname string          // host nickname (display form)
	spec     host.Spec       // shared host fields (alias/hostname/port/user)
	opts     sshopts.Options // SSH options (only for SSH hosts)
}

type formSaveResultMsg struct {
	err        error     // error during save IO operation
	protocol   string    // "ssh" or "telnet"
	spec       host.Spec // saved host spec
	configPath string    // config file written to (best-effort; set on success)
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
