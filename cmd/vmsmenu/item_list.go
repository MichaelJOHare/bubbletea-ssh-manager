package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

func (it *menuItem) Title() string {
	return it.name
}

func (it *menuItem) Description() string {
	if it.kind == itemHost {
		return it.protocol
	}
	return "group"
}

func (it *menuItem) FilterValue() string {
	if it.kind == itemHost {
		return it.name + " " + it.protocol + " " + it.target
	}
	return it.name
}

func toListItems(items []*menuItem) []list.Item {
	out := make([]list.Item, 0, len(items))
	for _, it := range items {
		out = append(out, it)
	}
	return out
}

func (m *model) current() *menuItem {
	return m.path[len(m.path)-1]
}

func (m *model) setCurrentMenu(items []*menuItem) {
	m.allItems = items
	m.lst.SetItems(toListItems(items))

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
