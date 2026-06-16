package tui

// statusView is one filter on an entity's view axis — a task status, or an audit
// bucket. word is the `:` vocabulary + s/S cycle label; value is the canonical
// filter ("" = the entity's default working view: active tasks / open audits).
type statusView struct {
	word  string
	value string
}

// statusViews is the single source of truth for task status views: the s/S cycle
// order AND the `:` vocabulary both derive from it, so they can't drift. A test
// (TestStatusViewsCoverAllStatuses) guards that it stays in sync with the domain
// status set. "active" leads as the working-set default.
var statusViews = []statusView{
	{"active", ""},
	{"in-progress", "in-progress"},
	{"next-up", "next-up"},
	{"ready-to-start", "ready-to-start"},
	{"completed", "completed"},
	{"deprecated", "deprecated"},
	{"deferred", "deferred"},
	{"all", "all"},
}

// statusViewAliases are extra `:` words that resolve to a canonical value but are
// not part of the s/S cycle.
var statusViewAliases = []statusView{
	{"working", ""},
	{"working-set", ""},
}

// auditViews is the audits tab's bucket axis — the s/S cycle order AND the `:`
// vocabulary. "open" leads as the default (value "" = the open working set,
// matching the CLI default and loadAuditList); closed/deferred/all reach the
// archived buckets that previously needed `audit list --all`. "deferred"/"all"
// overlap the task axis, so dispatchCommand resolves them against the active tab
// first (see resolveView).
var auditViews = []statusView{
	{"open", ""},
	{"closed", "closed"},
	{"deferred", "deferred"},
	{"all", "all"},
}

// viewWords is the `:` Tab-completion vocabulary for a view axis (cycle + aliases).
func viewWords(views, aliases []statusView) []string {
	words := make([]string, 0, len(views)+len(aliases))
	for _, sv := range views {
		words = append(words, sv.word)
	}
	for _, sv := range aliases {
		words = append(words, sv.word)
	}
	return words
}

// viewFor maps a `:` word to its canonical value on a view axis, reporting whether
// the word names a view at all.
func viewFor(views, aliases []statusView, word string) (string, bool) {
	for _, sv := range views {
		if sv.word == word {
			return sv.value, true
		}
	}
	for _, sv := range aliases {
		if sv.word == word {
			return sv.value, true
		}
	}
	return "", false
}

// viewStep returns the value dir steps from current in the axis cycle (wrapping).
// An unknown current view starts the walk from the default (index 0).
func viewStep(views []statusView, current string, dir int) string {
	cur := 0
	for i, sv := range views {
		if sv.value == current {
			cur = i
			break
		}
	}
	n := len(views)
	return views[((cur+dir)%n+n)%n].value
}
