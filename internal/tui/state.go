package tui

import (
	"os/exec"
	"time"

	"bubbletea-ssh-manager/internal/connect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type confirmState struct {
	form        *huh.Form // confirmation form
	title       string    // title of confirmation
	description string    // description of confirmation
	//returnMode  uiMode  // *** change this so only cancelling on edit goes back to previous mode
	onConfirm tea.Cmd // command to run on confirm
	onCancel  tea.Cmd // command to run on cancel
}

type preflightState struct {
	token       int                 // increments on preflight starts; for tick/result matching
	remaining   int                 // remaining seconds in preflight (for display)
	endsAt      time.Time           // when the preflight should end
	protocol    string              // "ssh" or "telnet"
	hostPort    string              // host:port being checked
	windowTitle string              // original window title before preflight
	cmd         *exec.Cmd           // running preflight command
	tail        *connect.TailBuffer // tail buffer for preflight output
	display     string              // display target (eg. host:port) for status messages
}

type modeState struct {
	// prompt state
	pendingHost *menuItem

	// host add/edit state
	hostForm         *huh.Form // host add/edit form
	hostFormValues   *form     // bound values backing the form fields (live as user types)
	hostFormMode     formMode  // add vs edit
	hostFormOldAlias string    // for edit/rename

	// confirmation dialog state (generic, used for remove/cancel/save confirmations)
	confirm *confirmState

	// preflight check state
	preflight preflightState
}

type uiMode int // current UI mode - menu, prompt, preflight, executing, etc.
const (
	modeMenu uiMode = iota
	modePromptUsername
	modeHostDetails
	modeHostForm
	modePreflight
	modeExecuting
	modeConfirm
)

type formMode int // add vs edit mode for host entry form
const (
	modeAdd formMode = iota
	modeEdit
)
