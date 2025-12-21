package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	hostNameColor  = lipgloss.Color("10")  // green
	groupNameColor = lipgloss.Color("208") // orange
)

type menuDelegate struct {
	list.DefaultDelegate
}

func newMenuDelegate() menuDelegate {
	d := list.NewDefaultDelegate()
	return menuDelegate{DefaultDelegate: d}
}

func (d menuDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, desc  string
		matchedRunes []int
	)

	// Copy styles so per-item tweaks don't leak across renders.
	styles := d.Styles
	normalTitle := styles.NormalTitle
	selectedTitle := styles.SelectedTitle
	dimmedTitle := styles.DimmedTitle

	// We only know how to render DefaultItem (which *menuItem is).
	i, ok := item.(list.DefaultItem)
	if !ok {
		return
	}
	title = i.Title()
	desc = i.Description()

	width := m.Width()
	if width <= 0 {
		return
	}

	// Apply per-kind coloring to titles.
	if mi, ok := item.(*menuItem); ok {
		switch mi.kind {
		case itemGroup:
			normalTitle = normalTitle.Foreground(groupNameColor)
			selectedTitle = selectedTitle.Foreground(groupNameColor)
			dimmedTitle = dimmedTitle.Foreground(groupNameColor)
		case itemHost:
			normalTitle = normalTitle.Foreground(hostNameColor)
			selectedTitle = selectedTitle.Foreground(hostNameColor)
			dimmedTitle = dimmedTitle.Foreground(hostNameColor)
		}
	}

	// Prevent text from exceeding list width.
	textWidth := max(width-normalTitle.GetPaddingLeft()-normalTitle.GetPaddingRight(), 0)
	title = ansi.Truncate(title, textWidth, "…")
	if d.ShowDescription {
		var lines []string
		for i, line := range strings.Split(desc, "\n") {
			if i >= d.Height()-1 {
				break
			}
			lines = append(lines, ansi.Truncate(line, textWidth, "…"))
		}
		desc = strings.Join(lines, "\n")
	}

	isSelected := index == m.Index()
	emptyFilter := m.FilterState() == list.Filtering && m.FilterValue() == ""
	isFiltered := m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	if isFiltered {
		matchedRunes = m.MatchesForItem(index)
	}

	if emptyFilter {
		title = dimmedTitle.Render(title)
		desc = styles.DimmedDesc.Render(desc)
	} else if isSelected && m.FilterState() != list.Filtering {
		if isFiltered {
			unmatched := selectedTitle.Inline(true)
			matched := unmatched.Inherit(styles.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		title = selectedTitle.Render(title)
		desc = styles.SelectedDesc.Render(desc)
	} else {
		if isFiltered {
			unmatched := normalTitle.Inline(true)
			matched := unmatched.Inherit(styles.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		title = normalTitle.Render(title)
		desc = styles.NormalDesc.Render(desc)
	}

	if d.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc) //nolint: errcheck
		return
	}
	fmt.Fprintf(w, "%s", title) //nolint: errcheck
}
