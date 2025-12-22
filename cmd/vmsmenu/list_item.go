package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

// Title returns the name of the menu item.
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
// For host items, it's a combination of name, protocol, alias, hostname, and port.
// For group items, it's just the name.
func (it *menuItem) FilterValue() string {
	if it.kind == itemHost {
		parts := []string{it.name, it.protocol}
		if v := strings.TrimSpace(it.alias); v != "" {
			parts = append(parts, v)
		}
		if v := strings.TrimSpace(it.user); v != "" {
			parts = append(parts, v)
		}
		if v := strings.TrimSpace(it.hostname); v != "" {
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
func toListItems(items []*menuItem) []list.Item {
	out := make([]list.Item, 0, len(items))
	for _, it := range items {
		out = append(out, it)
	}
	return out
}

// setCurrentMenu sets the current menu items and updates the list title.
func (m *model) setCurrentMenu(items []*menuItem) {
	m.allItems = items
	if m.delegate != nil {
		m.delegate.groupHints = nil
	}
	m.setItemsSafely(toListItems(items))

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

// setItemsSafely refreshes the list items and updates the selection index immediately.
//
// Without this, the previous index can be temporarily out of range and
// the highlight may appear one tick late when filtering rapidly.
func (m *model) setItemsSafely(items []list.Item) {
	m.lst.SetItems(items)
	if n := len(m.lst.Items()); n > 0 {
		idx := m.lst.Index()
		idx = min(max(idx, 0), n-1)
		m.lst.Select(idx)
	}
}
