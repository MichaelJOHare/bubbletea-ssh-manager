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
// For host items, it's the name and target concatenated.
// For group items, it's just the name.
func (it *menuItem) FilterValue() string {
	if it.kind == itemHost {
		return it.name + " " + it.protocol + " " + it.target
	}
	return it.name
}

// toListItems converts a slice of menuItem pointers to a slice of list.Item.
func toListItems(items []*menuItem) []list.Item {
	out := make([]list.Item, 0, len(items))
	for _, it := range items {
		out = append(out, it)
	}
	return out
}

// current returns the current menu item (the last in the path).
func (m *model) current() *menuItem {
	return m.path[len(m.path)-1]
}

// setCurrentMenu sets the current menu items and updates the list title.
func (m *model) setCurrentMenu(items []*menuItem) {
	m.allItems = items
	m.lst.SetItems(toListItems(items))

	m.refreshSelection()

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

// refreshSelection keeps the selection valid immediately after swapping items.
//
// Without this, the previous index can be temporarily out of range and
// the highlight may appear one tick late.
func (m *model) refreshSelection() {
	if n := len(m.lst.Items()); n > 0 {
		idx := m.lst.Index()
		idx = min(max(idx, 0), n-1)
		m.lst.Select(idx)
	}
}