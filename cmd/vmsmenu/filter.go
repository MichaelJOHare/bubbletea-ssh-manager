package main

import (
	"cmp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

// applyFilter filters the list items based on the query string q.
//
// It performs a fuzzy match on the items' FilterValue() strings
// and ranks them by match quality.
func (m *model) applyFilter(q string) {
	q = normalizeString(q)
	if q == "" {
		if m.delegate != nil {
			m.delegate.groupHints = nil
		}
		m.setItemsSafely(toListItems(m.allItems))
		return
	}

	type scored struct {
		item  *menuItem
		score int
	}

	// behavior:
	// - inside a group: search only hosts in that group (current page)
	// - at the root: search all hosts globally, plus allow matching group names
	candidates := make([]*menuItem, 0, len(m.allItems))
	if m.inGroup() {
		// in group: current page is a group's host list
		candidates = append(candidates, m.allItems...)
		if m.delegate != nil {
			m.delegate.groupHints = nil
		}
	} else {
		// at root: include all group items and all hosts in tree
		for _, it := range m.allItems {
			if it != nil && it.kind == itemGroup {
				candidates = append(candidates, it)
			}
		}

		// also include all hosts with group hints
		hints := map[*menuItem]string{}
		for _, hwg := range allHostItemsWithGroup(m.root) {
			if hwg.host == nil {
				continue
			}
			candidates = append(candidates, hwg.host)
			if grp := strings.TrimSpace(hwg.groupPath); grp != "" {
				hints[hwg.host] = grp
			}
		}
		if m.delegate != nil {
			m.delegate.groupHints = hints
		}
	}

	// score candidates
	var matches []scored
	// use a set to avoid duplicates
	seen := map[*menuItem]struct{}{}
	// check each candidate for a match
	for _, it := range candidates {
		// skip nil items and duplicates
		if it == nil {
			continue
		}
		if _, ok := seen[it]; ok {
			continue
		}
		// else mark as seen
		seen[it] = struct{}{}

		// fuzzy match
		hay := strings.ToLower(it.FilterValue())
		s, ok := fuzzyScore(q, hay)
		if !ok {
			continue
		}
		matches = append(matches, scored{item: it, score: s})
	}

	// sort matches by score descending
	slices.SortStableFunc(matches, func(a, b scored) int {
		return cmp.Compare(b.score, a.score)
	})

	filtered := make([]list.Item, 0, len(matches))
	for _, sm := range matches {
		filtered = append(filtered, sm.item)
	}
	m.setItemsSafely(filtered)
}

// allHostItemsWithGroup returns all host items in the tree along with a display
// group path that indicates where the host came from.
func allHostItemsWithGroup(root *menuItem) []hostWithGroup {
	if root == nil {
		return nil
	}

	// assumes groups are one level deep (ie. non-nested)
	// root contains (a) ungrouped hosts and (b) group items whose direct children are hosts
	out := make([]hostWithGroup, 0, 64)
	for _, it := range root.children {
		if it == nil {
			continue
		}
		if it.kind == itemHost {
			out = append(out, hostWithGroup{host: it, groupPath: ""})
			continue
		}
		if it.kind != itemGroup {
			continue
		}
		grp := strings.TrimSpace(it.name)
		for _, ch := range it.children {
			if ch == nil || ch.kind != itemHost {
				continue
			}
			out = append(out, hostWithGroup{host: ch, groupPath: grp})
		}
	}
	return out
}

// fuzzyScore returns a simple subsequence match score
// higher is better. "ok" is false if q is not a subsequence of s.
func fuzzyScore(q, s string) (score int, ok bool) {
	// empty query matches everything with score 0
	if q == "" {
		return 0, true
	}

	// check for subsequence and calculate score
	qi := 0
	streak := 0
	for i := 0; i < len(s) && qi < len(q); i++ {
		if s[i] == q[qi] {
			qi++
			streak++
			score += 10 + (streak * 2) // reward contiguous runs
		} else {
			streak = 0
		}
	}

	// if we didn't consume all of q, it's not a match
	if qi != len(q) {
		return 0, false
	}

	// small preference for earlier matches
	if idx := strings.Index(s, string(q[0])); idx >= 0 {
		score -= idx
	}

	return score, true
}
