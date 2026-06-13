package domain

import (
	"fmt"
	"strings"
	"unicode"
)

// titleHostile reports whether a rune must not appear in a title destined for a
// filename slug. Two classes: filesystem-reserved ASCII (Windows-invalid, the
// path separator, control chars) and — the case that motivated this guard —
// *non-ASCII* punctuation/symbols (em/en dashes, curly quotes, bullets, arrows)
// that read like part of the name but Slugify silently drops. Benign ASCII
// punctuation (parens, commas, hyphens, …) is allowed: it slugifies predictably
// and isn't filesystem-hostile. Apostrophes (incl. the curly ones Slugify keeps
// silent) are allowed.
func titleHostile(r rune) bool {
	switch r {
	case '\'', '’', '‘': // apostrophes — Slugify drops them, read fine
		return false
	case ':', '/', '\\', '*', '?', '<', '>', '|', '"':
		return true
	}
	if r < 0x20 || r == 0x7f { // control characters
		return true
	}
	return r > unicode.MaxASCII && (unicode.IsPunct(r) || unicode.IsSymbol(r))
}

// ValidateTitle rejects a title containing filename-hostile characters, naming
// the offending runes and suggesting a clean title (so the fix is copy-paste).
// Slugify still normalizes for every internal/legacy path; this is the create
// path's "surface it loudly instead of silently mangling the slug" guard.
func ValidateTitle(title string) error {
	var bad []string
	seen := map[rune]bool{}
	for _, r := range title {
		if titleHostile(r) && !seen[r] {
			seen[r] = true
			bad = append(bad, string(r))
		}
	}
	if len(bad) == 0 {
		return nil
	}
	return fmt.Errorf("%w: title has characters not allowed in a filename (%s); try %q",
		ErrValidation, strings.Join(bad, " "), suggestTitle(title))
}

// suggestTitle drops hostile runes (to spaces) and collapses whitespace into a
// readable, slug-safe title — the suggestion ValidateTitle offers.
func suggestTitle(title string) string {
	var b strings.Builder
	for _, r := range title {
		if titleHostile(r) {
			b.WriteByte(' ')
		} else {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

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
