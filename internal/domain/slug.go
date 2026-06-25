package domain

import (
	"strings"
	"unicode"
)

// Slugify converts a title to a filename slug. The policy is an ALLOWLIST,
// not a denylist (rewritten 2026-06-13 — the old ASCII-punctuation regex let
// unicode punctuation like em-dashes into filenames, and patching classes one
// bug report at a time doesn't converge):
//
//   - letters, digits, and combining marks — any script — are kept (lowercased);
//   - '.' is kept (version numbers), though trimmed at the ends;
//   - apostrophes vanish silently ("don't" → dont, not don-t) — the one
//     language-driven exception;
//   - EVERY other rune is a word break: whitespace, '_', ASCII and unicode
//     punctuation, dashes of all widths, symbols. Runs collapse to one '-'.
//
// The result is trimmed of '-'/'.' and capped at ~80 bytes on a rune boundary,
// then backed up to a '-' so a word is never cut mid-way.
func Slugify(text string) string {
	var b strings.Builder
	pendingBreak := false
	for _, r := range strings.ToLower(strings.TrimSpace(text)) {
		switch {
		case r == '\'' || r == '’' || r == '‘':
			// silent: contractions and possessives read better unbroken
		case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '.':
			if pendingBreak && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			pendingBreak = false
		default:
			pendingBreak = true
		}
	}
	text = strings.Trim(b.String(), "-.")
	if len(text) > 80 {
		// Cut on a RUNE boundary at ≤80 bytes (multibyte scripts must never
		// slice mid-rune into an invalid-UTF-8 filename), then back up to the
		// previous dash so a word isn't cut mid-way.
		cut := 0
		for i := range text {
			if i > 80 {
				break
			}
			cut = i
		}
		text = text[:cut]
		// Back up to the previous word break so a word isn't cut mid-way — but
		// only when the dash is far enough in. A short first word before one very
		// long token (e.g. "ab veryverylongword…") has its only dash near the
		// front; backing up to it would collapse the slug to a useless stub
		// ("ab"), so in that case keep the rune-boundary cut instead.
		if i := strings.LastIndex(text, "-"); i >= len(text)/2 {
			text = text[:i]
		}
		text = strings.Trim(text, "-.")
	}
	return text
}
