---
schema: 1
id: 6fjvr03mr9zg
status: ready-to-start
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'Non-entity utility files (HOWTO/README) coexist in tool dirs: no-frontmatter .md is silently ignored (not errored/resolved); .tskflwctlignore adds explicit carveouts. Needed for the flat-layout scan.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, storage, config]
created: "2026-07-04"
---

# Curation carveouts: tolerate non-entity files in tool dirs (frontmatter-gate + .tskflwctlignore)

## Objective

Real planning repos keep **companion/utility files next to entities** — `HOWTO-execute.md`
(the audit-execute cheat sheet every audit links to), routine notes, READMEs, templates.
Today the tool is a purist about its dirs; the flat layout (Phase B) makes that bite. Give
it **curation carveouts** so non-entity files coexist without polluting resolution or
erroring — while a genuinely-broken entity still surfaces loudly.

**Why it's a Phase-B hard requirement (not just nice-to-have):** audits are scanned per
*bucket* dir today (`audits/open|closed|deferred/`), so `audits/HOWTO-execute.md` at the
audits *root* is already invisible to the tool. Once Phase B flattens audits to a single
scanned `audits/` (`markdownCandidates(auditsDir, "")`), that HOWTO becomes a resolution
candidate `id=HOWTO-execute` and a listing `FileProblem`. The flatten can't ship without
this.

## Design — two layers

1. **Automatic: frontmatter-gated entity recognition (zero config).** Split "not mine" from
   "mine but broken":
   - a `.md` with **no frontmatter fence** → **not an entity** → silently skipped in BOTH
     `markdownCandidates` (resolution) AND `scanDir` (listing) — no `FileProblem`, no
     candidate. (Today `scanDir`→`parseAudit`→`missingFrontmatterErr` makes it a
     `FileProblem`, and a flat `markdownCandidates` would make it a candidate.)
   - a `.md` **with** a fence but malformed/missing-required-fields → **a broken entity** →
     `FileProblem` as today (a real error the human must fix).
   This is the clean default: the tool owns frontmatter'd files; everything else is invisible.
2. **Explicit: `.tskflwctlignore` (gitignore-style) — the customization edge.** For files
   that DO carry frontmatter but aren't tool entities (a frontmatter'd template), or a whole
   utility subdir. Git-native glob patterns in the planning root; honored by the scan +
   resolution + `WatchPaths`. Config key alternative (`ignore = [...]`) considered — a
   dotfile is more familiar and diffable.

Best-practice guidance (docs, not enforced): companion docs *can* live in an entity dir
(the carveout allows it), but a `planning/docs/` or co-located `planning/routines/` home
keeps them clearly out of the tracked set.

## Acceptance criteria

- [ ] A no-frontmatter `.md` in a scanned entity dir is invisible to `list`/`show`/`lint`
      (no `FileProblem`, not resolvable) — verified against a flat (Phase B) audit dir.
- [ ] A frontmatter'd-but-malformed entity still surfaces as a `FileProblem`.
- [ ] `.tskflwctlignore` globs exclude matching files/dirs from scan + resolution + watch.

## Out of scope

- Making routines/audits a **first-class tracked entity** with a routine↔audit linkage — a
  larger, exciting direction the user flagged (see Related). This task only makes the tool
  *tolerate* non-entity files; the other would *manage* them.

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- Hard prerequisite for [[flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory]]
  (the flat scan pollution). Surfaced by the desirelines `audits/HOWTO-execute.md` case.
- **Future direction (user, 2026-07-04):** the routine↔audit connection ("a routine produces
  an audit; the audit links back") may become a key tskflwctl feature — routines as a
  first-class entity type, not just tolerated files. Worth its own spike/epic if pursued.
