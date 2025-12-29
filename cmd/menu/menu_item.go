package main

import (
	"strings"

	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"

	"github.com/charmbracelet/bubbles/list"
)

type itemKind int // type of menu item: group or host

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

// Title returns the main display name of the menu item.
func (it *menuItem) Title() string {
	return it.name
}

// Description returns a short description of the menu item.
//
// For host items, it's the protocol.
// For group items, it's just "group".
func (it *menuItem) Description() string {
	if it.kind == itemHost {
		return it.protocol
	}
	return "group"
}

// FilterValue returns the string used for filtering this item.
//
// For host items, it's a combination of name, protocol, alias, username, and hostname.
// For group items, it's just the name.
func (it *menuItem) FilterValue() string {
	if it.kind == itemHost {
		parts := []string{it.name, it.protocol}
		if v := strings.TrimSpace(it.spec.Alias); v != "" {
			parts = append(parts, v)
		}
		if v := strings.TrimSpace(it.spec.User); v != "" {
			parts = append(parts, v)
		}
		if v := strings.TrimSpace(it.spec.HostName); v != "" {
			parts = append(parts, v)
		}
		return strings.Join(parts, " ")
	}
	return it.name
}

// current returns the current menu item (the last in the path).
func (m *model) current() *menuItem {
	return m.path[len(m.path)-1]
}

// inGroup returns true if the current path is inside a group (not at root).
func (m *model) inGroup() bool {
	return len(m.path) > 1
}

// toListItems converts a slice of menuItem pointers to a slice of list.Item.
//
// Used to turn custom menu items into list items for the Bubble Tea list component.
func toListItems(items []*menuItem) []list.Item {
	out := make([]list.Item, 0, len(items))
	for _, it := range items {
		out = append(out, it)
	}
	return out
}

// setActiveMenuItem updates the list view to show only the currently selected item.
//
// Used when displaying more info on (or connecting to) a host to focus the view on the selected host.
func (m *model) setActiveMenuItem(listView string) string {
	selected := m.lst.SelectedItem()
	if selected != nil {
		lst := m.lst
		lst.SetItems([]list.Item{selected})
		lst.Select(0)
		listView = lst.View()
	}
	return listView
}

// setCurrentMenu sets the current menu items and updates the list title.
//
// Used when navigating into or out of groups.
func (m *model) setCurrentMenu(items []*menuItem) {
	m.allItems = items
	if m.delegate != nil {
		m.delegate.groupHints = nil
	}
	m.updateItems(toListItems(items))

	parts := make([]string, 0, len(m.path))
	for _, p := range m.path {
		name := strings.TrimSpace(p.name)
		if name == "" {
			continue
		}
		parts = append(parts, name)
	}
	m.lst.Title = strings.Join(parts, " / ")
}

// updateItems sets the list items and resets selection to the first item.
//
// This ensures that the list state remains consistent after filtering or menu changes.
func (m *model) updateItems(items []list.Item) {
	m.lst.SetItems(items)
	m.lst.Select(0) // reset selection to first item
}
