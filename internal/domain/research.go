package domain

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
