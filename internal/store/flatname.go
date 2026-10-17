package store

import (
	"fmt"
	"strings"

	"github.com/andy-esch/taskflow/internal/id"
)

// splitFlatName parses a flat, id-led entity filename stem (`<id>-<slug>`, the
// basename with `.md` already stripped) into its stable id and human slug. It is
// the Phase-B counterpart to the old "the slug is the whole basename" rule: under
// ADR-0003 §4 tasks and audits live in one flat directory as `<id>-<slug>.md`, so
// identity leads the name and the slug is the renamable remainder.
//
// The split is by POSITION — the id is the fixed-width leading field (id.Length
// chars) followed by a single `-`, and the slug is everything after — never by
// splitting on `-`. A slug routinely contains dashes (`add-retry-backoff`, or an
// audit's `2026-06-16-dispatcher`), so splitting on the first or last dash would
// corrupt it; slicing the fixed id off the front is the only safe parse.
//
// ok is false when stem does not lead with a valid id + separator + non-empty
// slug. That is exactly the carveout gate (ADR-0003 amendment 2026-07-04): a
// non-entity file left in a scanned bucket (`HOWTO-execute`, `README`) is not
// id-led, so it parses to ok=false — neither a resolution candidate nor a file the
// scan mistakes for an entity. A real entity that merely lost its frontmatter is
// still id-led, so it stays ok=true here and fails loud later at parse time.
func splitFlatName(stem string) (entityID, slug string, ok bool) {
	// Need the fixed-width id, its `-` separator, and at least one slug character.
	if len(stem) < id.Length+2 {
		return "", "", false
	}
	if stem[id.Length] != '-' {
		return "", "", false
	}
	cand := stem[:id.Length]
	if !id.Valid(cand) {
		return "", "", false
	}
	return cand, stem[id.Length+1:], true
}

// entityNameProblem is flatNameProblem for a full basename, with the disposition advice
// appended only when the file really is a non-entity — telling someone to move a
// mistyped-id task into meta/ would be actively wrong advice.
func entityNameProblem(base string) (string, error) {
	stem := strings.TrimSuffix(base, ".md")
	if reason := flatNameProblem(stem); reason != "" {
		return reason, errBadEntityID
	}
	return "has no leading id — move it to meta/ or delete it", errNotEntity
}

// flatNameProblem explains a MISSPELLED id, returning "" when the stem simply is not
// id-led at all (a genuine stray). It exists because
// splitFlatName's single ok=false collapsed two very different situations into one message:
// a genuinely non-entity file (notes.md), and a file whose 12-character id is right there
// but contains a character Crockford excludes. The second was reported as "has no leading
// id", which sends the reader looking for a missing id instead of at the one wrong letter.
//
// The returned text is the reason clause; callers supply the surrounding sentence.
func flatNameProblem(stem string) string {
	if len(stem) < id.Length+2 || stem[id.Length] != '-' {
		return ""
	}
	cand := stem[:id.Length]
	reason := invalidEntityIDReason(cand)
	if reason == "" {
		return ""
	}
	lower := strings.ToLower(cand)
	if fixed, ok := id.Canonicalize(lower); ok && len(lower) == len(cand) {
		// Casing and i/l/o spelling can preserve the same decoded identity.
		// Pure filename advice cannot establish destination availability. Scans
		// replace it with two-copy guidance when their entry snapshot has a sibling.
		// Neither diagnostic grants authority to the automatic repair writer.
		return reason + fmt.Sprintf(" — rename to %q only if that path is absent, and match the `id:` field; otherwise compare both copies without overwriting either", fixed+stem[id.Length:]+".md")
	}
	if strings.ContainsAny(cand, "uU") {
		return reason + " — u has no canonical decode, so this id must be replaced deliberately"
	}
	return reason + " — no canonical decode is established; inspect the intended id manually"
}

// occupiedSpellingProblem contextualizes a parse failure using the same directory
// entries as the scan, without another filesystem read or I/O in pure parsers.
// Even a directory or symlink occupies the name: never advise overwriting it.
func occupiedSpellingProblem(base string, entries map[string]bool) string {
	stem := strings.TrimSuffix(base, ".md")
	if len(stem) < id.Length+2 || stem[id.Length] != '-' {
		return ""
	}
	cand := stem[:id.Length]
	reason := invalidEntityIDReason(cand)
	if reason == "" {
		return ""
	}
	lower := strings.ToLower(cand)
	fixed, ok := id.Canonicalize(lower)
	if !ok || len(lower) != len(cand) {
		return ""
	}
	destination := fixed + stem[id.Length:] + ".md"
	if !entries[destination] {
		return ""
	}
	return reason + fmt.Sprintf(" — canonical sibling %q already exists; compare and merge both copies without overwriting either, then deliberately remove or archive the copy not kept outside scanned entity directories; keep the surviving filename and `id:` field consistent; lint --fix refuses while both paths exist", destination)
}

// Use the same strict classification for filenames and write-time validation.
// A case defect is not an excluded character or permission to mint a new ID.
// This formats advice only: id.Valid and automatic repair authority stay strict.
func invalidEntityIDReason(value string) string {
	if id.Valid(value) {
		return ""
	}
	lower := strings.ToLower(value)
	if len(lower) == len(value) && id.Valid(lower) {
		return fmt.Sprintf("id %q must be lowercase (canonical spelling %q)", value, lower)
	}
	if c, pos, bad := id.InvalidChar(value); bad {
		if c >= 'A' && c <= 'Z' {
			_, _, lowerBad := id.InvalidChar(strings.ToLower(string(c)))
			if !lowerBad {
				return fmt.Sprintf("id %q contains uppercase %q at position %d; ids require %d lowercase Crockford base32 characters", value, string(c), pos, id.Length)
			}
		}
		return fmt.Sprintf("id %q contains %q at position %d, outside the canonical lowercase Crockford base32 alphabet (no i, l, o, u)", value, string(c), pos)
	}
	return fmt.Sprintf("id %q is not a valid id (%d lowercase Crockford base32 characters)", value, id.Length)
}
