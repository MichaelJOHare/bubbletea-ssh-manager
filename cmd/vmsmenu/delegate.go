package main

import (
	str "bubbletea-ssh-manager/internal/stringutil"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/x/ansi"
)

type menuDelegate struct {
	list.DefaultDelegate                      // embed default delegate to reuse its functionality
	groupHints           map[*menuItem]string // optional group hints per host item
}

// newMenuDelegatePtr creates a new menuDelegate with default settings.
//
// It embeds the default delegate from bubbles/list to leverage existing functionality.
func newMenuDelegatePtr() *menuDelegate {
	d := list.NewDefaultDelegate()
	return &menuDelegate{DefaultDelegate: d}
}

// Render renders a menu item with custom styles based on its kind and state.
//
// It applies different colors for group and host items, and adjusts the description
// to include group hints when available.
func (d *menuDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, desc string
	)

	// copy styles so per-item tweaks don't leak across renders
	styles := d.Styles
	normalTitle := styles.NormalTitle
	selectedTitle := styles.SelectedTitle
	normalDesc := styles.NormalDesc
	selectedDesc := styles.SelectedDesc

	// we only know how to render DefaultItem (which *menuItem is)
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

	// apply per-kind coloring to titles
	mi, _ := item.(*menuItem)
	if mi != nil {
		switch mi.kind {
		case itemGroup:
			normalTitle = normalTitle.Foreground(groupNameColor)
			selectedTitle = selectedTitle.Foreground(groupNameColor)
		case itemHost:
			if d.groupHints != nil {
				if grp := strings.TrimSpace(d.groupHints[mi]); grp != "" {
					desc = str.NormalizeString(mi.protocol) + " • " + grp
				}
			}
			protocol := str.NormalizeString(mi.protocol)
			if protocol == "telnet" {
				normalTitle = normalTitle.Foreground(telnetHostNameColor)
				selectedTitle = selectedTitle.Foreground(telnetHostNameColor)
			} else {
				// default to SSH color (green)
				normalTitle = normalTitle.Foreground(sshHostNameColor)
				selectedTitle = selectedTitle.Foreground(sshHostNameColor)
			}
		}
	}

	// prevent text from exceeding list width
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

	// apply selected vs normal styles
	isSelected := index == m.Index()
	if isSelected {
		title = selectedTitle.Render(title)
		desc = selectedDesc.Render(desc)
	} else {
		title = normalTitle.Render(title)
		desc = normalDesc.Render(desc)
	}

	// render final output
	if d.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc)
		return
	}
	fmt.Fprintf(w, "%s", title)
}
