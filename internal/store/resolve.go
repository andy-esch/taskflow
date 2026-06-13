package store

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Fuzzy slug resolution: the keyboard-economy companion to tab-completion.
// An exact id always wins; otherwise a unique case-insensitive prefix, then a
// unique case-insensitive substring, resolves. Ambiguity is explicit — more
// than one match at the winning tier returns ErrAmbiguous listing the
// candidates — and matching is deterministic (candidates are sorted).

// candidate is one resolvable id and where it lives. dir is the status/bucket
// directory name ("" for epics) — shown in ambiguity messages, and convertible
// back to the typed status/bucket by the status==directory invariant.
type candidate struct {
	id   string
	path string
	dir  string
}

// validQueryName rejects queries that could escape the planning tree when
// joined into a path (separators, dot-dot) — a slug is a plain name.
func validQueryName(kind, q string) error {
	if q == "" || strings.ContainsAny(q, `/\`) || strings.Contains(q, "..") {
		return fmt.Errorf("%w: %s name %q must be a plain name (no path separators)", domain.ErrValidation, kind, q)
	}
	return nil
}

// resolveID picks query's match among candidates: exact > unique
// case-insensitive prefix > unique case-insensitive substring. Multiple
// matches at the winning tier are ErrAmbiguous (listing them); none is
// ErrNotFound. kind ("task"/"epic"/"audit") shapes the error messages.
func resolveID(kind, query string, cands []candidate) (candidate, error) {
	if err := validQueryName(kind, query); err != nil {
		return candidate{}, err
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].id != cands[j].id {
			return cands[i].id < cands[j].id
		}
		return cands[i].dir < cands[j].dir
	})
	q := strings.ToLower(query)
	tiers := []func(id string) bool{
		func(id string) bool { return id == query || strings.ToLower(id) == q },
		func(id string) bool { return strings.HasPrefix(strings.ToLower(id), q) },
		func(id string) bool { return strings.Contains(strings.ToLower(id), q) },
	}
	for _, match := range tiers {
		var hits []candidate
		for _, c := range cands {
			if match(c.id) {
				hits = append(hits, c)
			}
		}
		switch len(hits) {
		case 0:
			continue // try the next, looser tier
		case 1:
			return hits[0], nil
		default:
			return candidate{}, fmt.Errorf("%q matches %d %ss: %s: %w",
				query, len(hits), kind, describeCandidates(hits), domain.ErrAmbiguous)
		}
	}
	return candidate{}, fmt.Errorf("%s %q: %w", kind, query, domain.ErrNotFound)
}

// describeCandidates renders an ambiguity list — "a (in-progress), b (open)" —
// so the error itself is enough to retype an unambiguous name.
func describeCandidates(cands []candidate) string {
	parts := make([]string, len(cands))
	for i, c := range cands {
		if c.dir == "" {
			parts[i] = c.id
		} else {
			parts[i] = fmt.Sprintf("%s (%s)", c.id, c.dir)
		}
	}
	return strings.Join(parts, ", ")
}
