package store

import "github.com/andy-esch/taskflow/internal/id"

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
