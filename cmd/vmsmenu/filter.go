package main

import (
	"cmp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

func (m *model) applyFilter(q string) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		m.lst.SetItems(toListItems(m.allItems))
		return
	}

	type scored struct {
		item  *menuItem
		score int
	}

	var matches []scored
	for _, it := range m.allItems {
		hay := strings.ToLower(it.FilterValue())
		s, ok := fuzzyScore(q, hay)
		if !ok {
			continue
		}
		matches = append(matches, scored{item: it, score: s})
	}

	slices.SortStableFunc(matches, func(a, b scored) int {
		return cmp.Compare(b.score, a.score)
	})

	filtered := make([]list.Item, 0, len(matches))
	for _, m := range matches {
		filtered = append(filtered, m.item)
	}
	m.lst.SetItems(filtered)
}

// fuzzyScore returns a simple subsequence match score
// higher is better. "ok" is false if q is not a subsequence of s
func fuzzyScore(q, s string) (score int, ok bool) {
	if q == "" {
		return 0, true
	}

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
	if qi != len(q) {
		return 0, false
	}

	// small preference for earlier matches
	if idx := strings.Index(s, string(q[0])); idx >= 0 {
		score -= idx
	}

	return score, true
}
