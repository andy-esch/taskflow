package tui

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
)

// sortKey identifies an interactive sort column. sortDefault keeps the loader's
// order (working-set for tasks, service order otherwise); the rest reorder the
// loaded slice in memory. Adding a column is a new const + a case in lessBy +
// one field in sortFields — the "extensible so new columns are cheap" ask.
type sortKey int

const (
	sortDefault sortKey = iota
	sortPriority
	sortUpdated
	sortTier
	sortSlug
)

// Per-entity `o`-cycle orders: each entity offers only the columns it actually
// has, so cycling never lands on a no-op sort (e.g. tier on epics) that would show
// a chip while nothing reorders. sortDefault is the resting state in each.
var (
	taskSortCols  = []sortKey{sortDefault, sortPriority, sortUpdated, sortTier, sortSlug}
	epicSortCols  = []sortKey{sortDefault, sortPriority, sortSlug}
	auditSortCols = []sortKey{sortDefault, sortUpdated, sortSlug}
)

// sortArrow shows the column's *actual* direction (not just default-vs-reversed):
// updated defaults to newest-first (↓), every other column to ascending (↑), and
// rev flips it — so the glyph always means what you'd expect for that column.
func sortArrow(k sortKey, rev bool) string {
	if (k == sortUpdated) != rev { // updated is descending by default
		return "↓"
	}
	return "↑"
}

func (k sortKey) label() string {
	switch k {
	case sortPriority:
		return "priority"
	case sortUpdated:
		return "updated"
	case sortTier:
		return "tier"
	case sortSlug:
		return "slug"
	default:
		return "default"
	}
}

// sortFields are the comparable values a row exposes for interactive sort. Items
// that lack a field leave it zero/empty (epics/audits have no tier, say), which
// sorts last — sorting still works across entities.
type sortFields struct {
	priorityRank int    // high=0, medium=1, low=2, unset=3
	updated      string // YYYY-MM-DD ("" = unknown, sorts last)
	tier         int    // 0 = unset (sorts last)
	slug         string // the stable tiebreak
}

// priorityRank maps the priority enum to a sort order (high first).
func priorityRank(p string) int {
	switch p {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 3
	}
}

// lessBy reports whether a sorts before b under column k. slug is the universal
// tiebreak so the order is total (stable, deterministic across reloads).
func lessBy(k sortKey, a, b sortFields) bool {
	switch k {
	case sortPriority:
		if a.priorityRank != b.priorityRank {
			return a.priorityRank < b.priorityRank
		}
	case sortUpdated:
		// Most-recent first: the larger date string sorts earlier. "" is lexically
		// smallest, so unknown-updated rows fall to the bottom — what we want.
		if a.updated != b.updated {
			return a.updated > b.updated
		}
	case sortTier:
		if a.tier != b.tier {
			return tierLess(a.tier, b.tier)
		}
	case sortSlug:
		// fall through to the slug tiebreak below
	default:
		return false // sortDefault: no reorder
	}
	return a.slug < b.slug
}

// tierLess orders tiers ascending (1,2,3) with unset (0) last.
func tierLess(a, b int) bool {
	switch {
	case a == 0:
		return false
	case b == 0:
		return true
	default:
		return a < b
	}
}

// sortItems reorders items in place under (key, rev). sortDefault keeps the
// loader's order (only reversing it when rev is set).
func sortItems(items []list.Item, key sortKey, rev bool) {
	if key == sortDefault {
		if rev {
			for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
				items[i], items[j] = items[j], items[i]
			}
		}
		return
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, aok := items[i].(entityItem)
		b, bok := items[j].(entityItem)
		if !aok || !bok {
			return false
		}
		if rev {
			return lessBy(key, b.sortFields(), a.sortFields())
		}
		return lessBy(key, a.sortFields(), b.sortFields())
	})
}
