package tui

import "bubbletea-ssh-manager/internal/config"

const (
	modeAdd formMode = iota
	modeEdit
)

type formMode int // add vs edit mode for host entry form

type formCanceledMsg struct{}

type formSubmittedMsg struct {
	mode     formMode          // add vs edit mode for host entry form
	protocol config.Protocol   // protocol being edited/added
	oldAlias string            // for edit/rename
	group    string            // group name (display form)
	nickname string            // host nickname (display form)
	spec     config.Spec       // shared host fields (alias/hostname/port/user)
	opts     config.SSHOptions // SSH options (only for SSH hosts)
}

type formSaveResultMsg struct {
	err        error           // error during save IO operation
	protocol   config.Protocol // protocol that was saved
	spec       config.Spec     // saved host spec
	configPath string          // config file written to (best-effort; set on success)
}

type menuReloadedMsg struct {
	root *menuItem // new root menu item
	err  error     // error during reload
}

type statusClearMsg struct {
	token int // token to identify which status to clear
}

type connectFinishedMsg struct {
	protocol config.Protocol // protocol used
	target   string          // display target (eg. host:port)
	err      error           // error from connection attempt
	output   string          // output from ssh/telnet command
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

// confirmResultMsg is sent when a confirmation dialog completes (confirmed or canceled).
type confirmResultMsg struct {
	confirmed bool // true if user confirmed, false if canceled
}

type removeHostResultMsg struct {
	protocol config.Protocol // protocol that was removed
	alias    string          // alias of host that was removed
	err      error           // error during removal
}
