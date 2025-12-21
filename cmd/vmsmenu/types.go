package main

type itemKind int

const (
	itemGroup itemKind = iota
	itemHost
)

type menuItem struct {
	kind itemKind // item type: group or host
	name string   // display name

	// host-only fields
	protocol string // "ssh" or "telnet"
	target   string // hostname or IP address

	// group-only fields
	children []*menuItem // child menu items
}

type hostEntry struct {
	alias    string // nickname for host (ssh Host alias)
	hostname string // actual host name or IP (HostName)
	port     string // optional port number (Port)
}

type hostWithGroup struct {
	host      *menuItem // host menu item
	groupPath string    // display group path
}

type model struct {
	width  int // window width
	height int // window height

	query    textinputModel // search input box
	delegate *menuDelegate

	root     *menuItem   // root menu item
	path     []*menuItem // current navigation path
	allItems []*menuItem // all items in the current menu
	lst      listModel   // list of current menu items

	status        string // status message
	statusIsError bool   // is the status an error message?
	statusToken   int    // increments on status updates; cancels pending clears
	quitting      bool   // is the app quitting?
}

type statusClearMsg struct {
	token int // token to identify which status to clear
}

type connectFinishedMsg struct {
	protocol string // "ssh" or "telnet"
	target   string // hostname or IP address
	err      error  // error from connection attempt
	output   string // output from ssh/telnet command
}

type tailBuffer struct {
	buf []byte // stored bytes
	max int    // max bytes to keep
}
