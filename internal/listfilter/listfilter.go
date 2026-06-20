// Package listfilter is the shared substring matcher for bubbles/list. Both the
// CLI picker and the TUI use it for predictable "contains" filtering, so the two
// faces can't drift (the TUI additionally offers fuzzy via list.DefaultFilter).
package listfilter

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
)

// Substring is a case-insensitive "contains" list.FilterFunc — the predictable
// alternative to bubbles/list's default fuzzy matcher, which over-matches long
// structured slugs (e.g. "multiuser" fuzzy-matching unrelated tasks via a
// scattered subsequence). Matches keep input order; MatchedIndexes mark the run
// for highlighting.
func Substring(term string, targets []string) []list.Rank {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		ranks := make([]list.Rank, len(targets))
		for i := range targets {
			ranks[i] = list.Rank{Index: i}
		}
		return ranks
	}
	var ranks []list.Rank
	for i, t := range targets {
		lower := strings.ToLower(t)
		b := strings.Index(lower, term)
		if b < 0 {
			continue
		}
		// Index and slice the SAME (lowercased) string: ToLower can change byte
		// length (e.g. Ⱥ→ⱥ grows 2→3 bytes), so slicing the original t by b would
		// panic. MatchedIndexes are rune positions in the lowercased string; for
		// the rare case-folding-rune-count change the highlight can drift a rune,
		// which is cosmetic — but it can never panic.
		start := utf8.RuneCountInString(lower[:b])
		matched := make([]int, utf8.RuneCountInString(term))
		for k := range matched {
			matched[k] = start + k
		}
		ranks = append(ranks, list.Rank{Index: i, MatchedIndexes: matched})
	}
	return ranks
}
