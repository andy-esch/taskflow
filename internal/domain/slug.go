package domain

import (
	"regexp"
	"strings"
)

var (
	slugPunct = regexp.MustCompile("[:()\\[\\]{}'\"!@#$%^&*+=|\\\\<>,?/~`]")
	slugSpace = regexp.MustCompile(`[\s_]+`)
	slugDash  = regexp.MustCompile(`-{2,}`)
)

// Slugify converts a title to a filename slug, matching the Python pm rules so
// slugs stay stable across the two tools: lowercase, strip punctuation,
// whitespace/underscores → '-', collapse runs, trim, and cap at ~80 chars on a
// '-' boundary.
func Slugify(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	text = slugPunct.ReplaceAllString(text, "")
	text = slugSpace.ReplaceAllString(text, "-")
	text = slugDash.ReplaceAllString(text, "-")
	// Trim leading/trailing '-' and '.': a trailing dot is stripped by Windows
	// and a leading dot makes a hidden file. (The Windows-reserved characters
	// :\*?<>|"/ are already removed by slugPunct above.)
	text = strings.Trim(text, "-.")
	if len(text) > 80 {
		text = text[:80]
		if i := strings.LastIndex(text, "-"); i >= 0 {
			text = text[:i]
		}
	}
	return text
}
