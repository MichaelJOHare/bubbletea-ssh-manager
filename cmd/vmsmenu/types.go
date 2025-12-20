package main

type itemKind int

const (
	itemGroup itemKind = iota
	itemHost
)

type menuItem struct {
	kind itemKind

	name string

	// host-only fields
	protocol string // "ssh" or "telnet"
	target   string

	// group-only fields
	children []*menuItem
}

type model struct {
	width  int
	height int

	query textinputModel

	root     *menuItem
	path     []*menuItem
	allItems []*menuItem
	lst      listModel

	status        string
	statusIsError bool
	quitting      bool
}

type connectFinishedMsg struct {
	protocol string
	target   string
	err      error
	output   string
}
