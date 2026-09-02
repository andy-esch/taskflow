package domain

import "sort"

// Research is a research document: an exploration snapshot, true as of its date.
//
// It is deliberately the THINNEST entity in the registry, and the omissions are
// the design (epic 28, decided 2026-08-14):
//
//   - **No status, no bucket, no lifecycle verbs.** ADR-0001 adopted ADRs precisely
//     because research docs are "open-ended spikes with no lifecycle, numbering, or
//     status discipline" — that characterization is load-bearing, so research keeps
//     no lifecycle here. An ADR is a decision that freezes on acceptance; a research
//     doc is a snapshot that a later doc freely supersedes. Give research a status
//     vocabulary and it becomes a second, worse ADR.
//   - **No cross-references — not even `epic:`.** Other entities link TO research
//     docs with ordinary relative-path body links, which already cascade on rename
//     (store.RenameTask) and are already dangler-checked (`lint --links`). Research
//     carries no back-references, so there is no rollup, no resolved reference, and
//     no many-to-many modelling problem. Provenance stays a body concern.
//
// What's left is the minimum that makes the corpus discoverable: a date to order by,
// a description to list by, and tags to filter by. Created is required — it's the
// chronology anchor, and the id is minted FROM it (ADR-0003 §3), so lexical id order
// is authorship order.
//
// Unknown frontmatter is preserved, not rejected: the legacy corpus carries a
// vestigial `status: reference` on 18 docs, which is deliberately NOT declared in the
// contract and NOT linted — it simply rides along (the surgical-edit guarantee).
type Research struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"`

	// ID is the stable 12-char identifier (ADR-0003 §3): it leads the flat filename
	// (research/<id>-<slug>.md) and is the primary resolution key.
	ID string `yaml:"id"`

	// FilenameID is that same id as parsed from the flat filename's leading field
	// (set by the store via splitFlatName) — the canonical key; the frontmatter `id:`
	// above must equal it, and lint flags drift. Derived, not frontmatter.
	FilenameID string `yaml:"-"`

	// Created is the date the research was done (YYYY-MM-DD), required. The id is
	// minted from it, so this is what makes id order chronological.
	Created string `yaml:"created"`

	// Description is a one-line summary — the whole point of the list view. Optional:
	// the migrated corpus has none, and demanding one would fail lint on 28 files.
	Description string `yaml:"description"`

	Tags []string `yaml:"tags"`

	// Updated is the doc's own last-edited date (stamped by edit/append), distinct
	// from the immutable Created.
	Updated string `yaml:"updated_at"`
}

// CanonicalID is the stable store-resolution identity. Filename identity wins
// for filesystem records so frontmatter drift cannot redirect a read; adapters
// without filename semantics use ID.
func (r Research) CanonicalID() string {
	if r.FilenameID != "" {
		return r.FilenameID
	}
	return r.ID
}

// knownResearchFields is the frontmatter keys the tool recognizes for a research doc,
// DERIVED from the registry's AuthoringFields plus the tool-managed stamps — so a field
// added to the research Descriptor becomes settable without a second list to keep in
// sync (the epic-28 charter: a noun's fields ride one registry).
// knownResearchFields is every frontmatter key the tool RECOGNIZES for a research doc:
// the registry's AuthoringFields plus `updated_at`, the one tool-managed stamp research
// carries. Deriving from the registry means a field added to the Descriptor is recognized
// without a second list to keep in sync.
//
// `id` and `schema` are deliberately absent, matching KnownTaskFieldNames (which also
// omits both while including updated_at) — they are storage machinery, not frontmatter an
// author reasons about, and ProtectedResearchField rejects them before this set is ever
// consulted.
var knownResearchFields = func() map[string]bool {
	m := map[string]bool{"updated_at": true}
	fields, err := AuthoringFields("research")
	if err != nil { // unreachable: "research" is a registered kind
		panic("domain: research kind missing from the entity registry")
	}
	for _, f := range fields {
		m[f.Name] = true
	}
	return m
}()

// KnownResearchField reports whether f is a frontmatter key the tool knows for a
// research doc. `research set` gates on it (unless --force) so a typo'd key can't be
// silently persisted. RECOGNIZED is not the same as SETTABLE — see SettableResearchFields.
func KnownResearchField(f string) bool { return knownResearchFields[f] }

// KnownResearchFieldNames returns every recognized research frontmatter key, sorted — the
// research analog of KnownTaskFieldNames, and what `schema --json`'s research_fields is
// built from.
func KnownResearchFieldNames() []string {
	names := make([]string, 0, len(knownResearchFields))
	for f := range knownResearchFields {
		names = append(names, f)
	}
	sort.Strings(names)
	return names
}

// SettableResearchFields returns the keys `research set` will actually WRITE, sorted:
// recognized minus protected. Distinct from KnownResearchFieldNames because most
// recognized keys can't be set — `created` is encoded in the id and `updated_at` is
// stamped — so a caller offering choices (an error message listing alternatives, a TUI
// edit menu) must use this, or it advertises fields the write path then refuses.
func SettableResearchFields() []string {
	names := make([]string, 0, len(knownResearchFields))
	for f := range knownResearchFields {
		if _, protected := ProtectedResearchField(f); protected {
			continue
		}
		names = append(names, f)
	}
	sort.Strings(names)
	return names
}

// IsResearchListField reports whether a research key is stored as a YAML list (only
// `tags`), so `--set key=value` writes a sequence instead of a corrupting !!str.
func IsResearchListField(f string) bool { return f == "tags" }

// ProtectedResearchField reports whether a field must not be written through
// `research set`, and why. Returning the REASON (not just a bool) keeps the explanation
// next to the rule, so every adapter surfaces the same wording.
//
// `created` is the interesting one: the stable id is minted FROM it (ADR-0003 §3), so
// changing it in place would leave the id encoding one date and the field claiming
// another — silently breaking the id-order-is-date-order invariant for that doc, with
// no way to detect it later. Re-dating a research doc means creating a new one.
func ProtectedResearchField(field string) (string, bool) {
	switch field {
	case "created":
		return "created is encoded in the stable id (ADR-0003 §3) — changing it would desync the two; create a new doc instead", true
	case "id":
		return "id is the immutable key and must match the filename — rename the file instead", true
	case "schema":
		return "schema is the on-disk format version, managed by the tool", true
	case "updated_at":
		return "updated_at is stamped automatically", true
	}
	return "", false
}

// ValidateResearchField checks a constrained research frontmatter field from its string
// form — the research analog of ValidateEpicField.
func ValidateResearchField(field, value string) error {
	switch field {
	case "description":
		return ValidateDescription(value)
	case "created":
		// Unreachable via `research set` (ProtectedResearchField rejects it first), but
		// correct here so any other writer gets the mintable-range rule, not just the shape.
		return ValidateMintableDate(value)
	}
	return nil
}
