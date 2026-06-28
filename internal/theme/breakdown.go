package theme

import "strings"

// A breakdown line — "N key · N key · …" (finding counts by urgency, by area) — is
// assembled by several surfaces that disagree on styling: dim separators vs colored
// segments, a "+N more" cap or none. The STRUCTURE (iterate, format each segment,
// join, optional cap) is the shared part; the per-segment styling and the separator
// stay the caller's (audit M10). Breakdown owns the structure and takes those as
// callbacks, so no surface re-rolls the loop.
//
// items are joined with sep; at most max segments show (max <= 0 = all), and when the
// list is capped the more callback supplies the trailing segment (remaining is the
// dropped count). seg formats one item; more may be nil when max <= 0.
func Breakdown[T any](items []T, sep string, max int, seg func(T) string, more func(remaining int) string) string {
	parts := make([]string, 0, len(items))
	for i, it := range items {
		if max > 0 && i >= max {
			if more != nil {
				parts = append(parts, more(len(items)-max))
			}
			break
		}
		parts = append(parts, seg(it))
	}
	return strings.Join(parts, sep)
}
