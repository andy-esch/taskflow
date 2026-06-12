package tui

// statusView is one task status filter reachable from the `:` command bar and the
// s/S cycle. value is the canonical filter ("" = working set / active; "all" =
// include archived; otherwise an exact status string).
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

// statusViewWords is the `:` Tab-completion vocabulary for status views.
func statusViewWords() []string {
	words := make([]string, 0, len(statusViews)+len(statusViewAliases))
	for _, sv := range statusViews {
		words = append(words, sv.word)
	}
	for _, sv := range statusViewAliases {
		words = append(words, sv.word)
	}
	return words
}

// statusViewFor maps a `:` word to its canonical status value, reporting whether
// the word names a view at all.
func statusViewFor(word string) (string, bool) {
	for _, sv := range statusViews {
		if sv.word == word {
			return sv.value, true
		}
	}
	for _, sv := range statusViewAliases {
		if sv.word == word {
			return sv.value, true
		}
	}
	return "", false
}

// statusViewStep returns the value dir steps from the current view in the cycle
// (wrapping). An unknown current view starts the walk from the default.
func statusViewStep(current string, dir int) string {
	cur := 0
	for i, sv := range statusViews {
		if sv.value == current {
			cur = i
			break
		}
	}
	n := len(statusViews)
	return statusViews[((cur+dir)%n+n)%n].value
}
