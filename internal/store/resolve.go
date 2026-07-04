package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// markdownDoc reports whether a directory entry is the shape every store scan
// accepts: a regular `.md` file. Requiring a *regular* file (not just non-dir)
// rejects symlinks, so a planted `x.md` link can't be followed out of the
// planning tree — the read-side counterpart to validQueryName's query guard.
func markdownDoc(e os.DirEntry) bool {
	return e.Type().IsRegular() && strings.HasSuffix(e.Name(), ".md")
}

// scanDir reads every regular .md file in dir and parses each through parse. A
// parse failure becomes a FileProblem (the file is skipped, not fatal) so one
// bad file doesn't blind the listing; a missing dir yields nothing; only a real
// read error is fatal. It's the shared body of ListTasks/ListEpics/ListAudits —
// each passes a parse closure binding the file's status/bucket.
func scanDir[T any](dir string, parse func(path string, content []byte) (T, error)) ([]T, []domain.FileProblem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var out []T
	var problems []domain.FileProblem
	for _, e := range entries {
		if !markdownDoc(e) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", path, err)
		}
		v, err := parse(path, content)
		if err != nil {
			problems = append(problems, domain.FileProblem{Path: path, Message: err.Error()})
			continue
		}
		out = append(out, v)
	}
	return out, problems, nil
}

// markdownCandidates lists every regular .md file in dir as a resolution
// candidate, tagging each with dirName (the status/bucket, "" for epics). The
// shared body of the task/epic/audit candidate gatherers.
func markdownCandidates(dir, dirName string) ([]candidate, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var out []candidate
	for _, e := range entries {
		if !markdownDoc(e) {
			continue
		}
		out = append(out, candidate{
			id:   strings.TrimSuffix(e.Name(), ".md"),
			path: filepath.Join(dir, e.Name()),
			dir:  dirName,
		})
	}
	return out, nil
}

// flatCandidates lists every id-led entity file in ONE flat directory as a
// resolution candidate, parsing `<id>-<slug>.md` via splitFlatName so the id and
// slug become separate resolution keys. It is the flat-layout counterpart to
// markdownCandidates' per-status/bucket scan (ADR-0003 §4).
//
// A file whose name is not id-led — a non-entity carveout (`HOWTO-execute.md`,
// `README.md`, or a stray) — is skipped, so it is never a resolution candidate.
// That is the read-side of the carveout gate (ADR-0003 amendment 2026-07-04); the
// loud "not an entity" FileProblem for a stray is the *listing* side's job, not
// resolution's.
func flatCandidates(dir string) ([]candidate, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var out []candidate
	for _, e := range entries {
		if !markdownDoc(e) {
			continue
		}
		entityID, slug, ok := splitFlatName(strings.TrimSuffix(e.Name(), ".md"))
		if !ok {
			continue // not id-led — the carveout gate
		}
		out = append(out, candidate{
			id:   entityID,
			slug: slug,
			path: filepath.Join(dir, e.Name()),
		})
	}
	return out, nil
}

// Fuzzy slug resolution: the keyboard-economy companion to tab-completion.
// An exact id always wins; otherwise a unique case-insensitive prefix, then a
// unique case-insensitive substring, resolves. Ambiguity is explicit — more
// than one match at the winning tier returns ErrAmbiguous listing the
// candidates — and matching is deterministic (candidates are sorted).

// candidate is one resolvable id and where it lives. dir is the status/bucket
// directory name ("" for epics) — shown in ambiguity messages, and convertible
// back to the typed status/bucket by the status==directory invariant.
type candidate struct {
	id   string // resolution key: an epic/legacy stem, or the 12-char stable id under the flat layout
	slug string // human slug (flat tasks/audits) — a second resolution key; "" for epics and legacy candidates
	path string
	dir  string // status/bucket mirror ("" for epics, and for flat entities); retained through the cutover
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
		if cands[i].slug != cands[j].slug {
			return cands[i].slug < cands[j].slug
		}
		return cands[i].dir < cands[j].dir
	})
	q := strings.ToLower(query)
	tiers := []func(key string) bool{
		func(key string) bool { return key == query || strings.ToLower(key) == q },
		func(key string) bool { return strings.HasPrefix(strings.ToLower(key), q) },
		func(key string) bool { return strings.Contains(strings.ToLower(key), q) },
	}
	for _, match := range tiers {
		var hits []candidate
		for _, c := range cands {
			// Under the flat layout a task/audit resolves on either its stable id
			// (a prefix) or its human slug; epic/legacy candidates carry only id
			// (slug ""), so this stays their single-key match unchanged.
			if match(c.id) || (c.slug != "" && match(c.slug)) {
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

// slugCollision reports the directory a `<slug>.md` already occupies among cands
// (an exact-id match), or "" if the slug is free across them. The create path
// uses it to reject a slug that already lives in ANOTHER status dir / audit
// bucket: writeNewFile's O_EXCL only guards the single target path, so without
// this a `task new`/`audit new` could mint a second file with the same slug and
// make every later resolve of it ErrAmbiguous. Best-effort (a scan, not atomic)
// — fine for a single-user CLI, like the epic auto-numbering race.
func slugCollision(slug string, cands []candidate) string {
	for _, c := range cands {
		if c.id == slug {
			return c.dir
		}
	}
	return ""
}

// describeCandidates renders an ambiguity list — "a (in-progress), b (open)" —
// so the error itself is enough to retype an unambiguous name.
func describeCandidates(cands []candidate) string {
	parts := make([]string, len(cands))
	for i, c := range cands {
		switch {
		case c.dir != "":
			parts[i] = fmt.Sprintf("%s (%s)", c.id, c.dir)
		case c.slug != "":
			// Flat entities carry no dir; show slug + id so a dup-slug ambiguity is retypable by id.
			parts[i] = fmt.Sprintf("%s (%s)", c.slug, c.id)
		default:
			parts[i] = c.id
		}
	}
	return strings.Join(parts, ", ")
}
