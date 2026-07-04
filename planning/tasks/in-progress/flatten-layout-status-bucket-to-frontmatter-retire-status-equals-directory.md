---
schema: 1
status: in-progress
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Move status/bucket into frontmatter as source of truth and flatten tasks/audits to one dir each (id-led filenames); update store, layout, WatchPaths, resolution, completion. Per ADR-0003.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [core, storage]
created: "2026-07-01"
updated_at: "2026-07-04"
id: 6fhnydm03edq
started_at: "2026-07-03"
---
# Flatten the layout — Phase B: id-led flat files, retire status == directory

## Status / history (2026-07-04)

**Plan only — not started.** This task was originally the *Phase A* umbrella; that scope
(frontmatter-authoritative status/bucket with the dirs kept as a lock-step mirror) shipped
piecemeal under the sibling tasks (`replace-misfiled`, audits-authoritative, machine-contract)
plus [[version-aware-occ-content-hash-token-and-plain-retry]], all merged. Re-scoped here to
**Phase B**: the irreversible cutover that deletes the directory mirror entirely. Its slug and
description already fit Phase B.

## Objective

Flatten `tasks/<status>/<slug>.md` and `audits/<bucket>/<slug>.md` to ONE dir each with
**id-led filenames** (`tasks/<id>-<slug>.md`, `audits/<id>-<slug>.md`); status/bucket live
only in frontmatter (already authoritative). Resolve by **id-prefix (+ slug)**. A status
change becomes a pure **in-place frontmatter edit** (no relocation). Delete the status/bucket
subdirs, the mirror, WatchPaths' per-status globs, completion's dir-globbing, and the whole
**Misfiled** concept (no folder left to disagree). Rides the one-time migration script;
coordinates with scheme-2 references + rename.

## Target state

- `tasks/<id>-<slug>.md`, `audits/<id>-<slug>.md`, `epics/<id>-<slug>.md` — **epics are
  already flat, the template** (`markdownCandidates(dir, "")`).
- Resolution: **id is the primary key** (filename leads with it) — id-prefix, then slug (the
  trailing part), then fuzzy, over ONE dir.
- `task move`/`start`/… = an in-place `status:` frontmatter edit (exactly what `MoveEpic`
  does today). No `os.Rename` between dirs.
- Status/bucket read purely from frontmatter; a missing/unknown status is a hard read problem
  (no folder to fall back to — `replace-misfiled` already flags a missing status).

## Simplifications Phase B DELIVERS (not just costs)

- **`moveTask`/`MoveAudit` collapse to in-place edits** (model on `MoveEpic`) → the
  write-then-remove **dual-file window vanishes** (fsstore.go:222-263, auditstore.go:160-195),
  and with it the crash-duplicate recovery net. The OCC verify→write + **flock carry over
  unchanged** to the in-place write.
- **Resolution collapses** to the epic pattern (one scan; `taskCandidates`/`auditCandidates`'
  `AllStatuses()`/`AllAuditBuckets()` loops go away).
- **Dup-slug ambiguity vanishes** — id-led filenames are unique by construction; the
  `duplicateSlugIssues` lint and the ErrAmbiguous-across-dirs class retire.
- **`WatchPaths` collapses** to the 3 entity dirs (fsstore.go:54-63).
- **Misfiled / FolderStatus / *FellBack-fallback / the misfiled fixer / wire
  `misfiled`+`declared_status` / Summary `misfiled` / TUI misfiled markers all DELETE**
  (~84 refs across domain/store/wire/cli/tui).

## TRAPS / footguns / rakes (read before writing code)

1. **BIG-BANG cutover, not in-tool reversible.** Code supports flat-only; the migration script
   does the one-time data move; **git is the only undo**. Unlike Phase A (one-line revert),
   reverting Phase B = `git revert` + re-migrate. Run the migration on a COPY first; land code +
   migration together.
2. **The filename now encodes identity — id-in-filename vs id-in-frontmatter can drift.** Make
   the **filename id the resolution key** (primary), frontmatter `id:` the mirror; add a lint
   for filename-id ≠ frontmatter-id. Parse `<id>-<slug>.md` by **slicing the fixed 12-char id**
   (slugs contain dashes — do NOT split on `-`).
3. **Missing/unknown status has NO folder to fall back to.** Phase A's `StatusFellBack` fallback
   becomes a read problem, not a silent recovery. Lean: surface a `FileProblem` (resilient read)
   + the existing lint flag; do NOT invent an "unknown/" dir.
4. **A status change stops firing a file-*move* fsnotify event.** TUI live-reload + shell
   completion leaned on the relocation. After flatten a status change is an in-place WRITE (the
   `tasks/` watcher still fires), but **completion that globbed `tasks/<status>/` must read
   frontmatter per file** to filter by status. Re-point both.
5. **`.Dir()` is used for two things — untangle them.** As a *path component* (delete: fsstore,
   auditstore, create, fix, completion, domain/layout) AND, sometimes, as the canonical status
   string. Keep the status enum + its string; delete only the **path** uses.
6. **The wire contract loses `misfiled`/`declared_status` + the Summary `misfiled` count** →
   `schema_version` bump + golden regen. An agent on the old contract must see the version move.
7. **Audit filename convention.** Audit slug is `YYYY-MM-DD-area` (date immutable, the human
   key). Flat: `<id>-YYYY-MM-DD-area.md` — id leads for resolution/uniqueness, date stays in the
   slug.
8. **Discovery/anchoring.** `config` walks up for `tasks/`; flat still has `tasks/` (files
   directly inside). Keep the anchor; `.gitkeep` becomes unnecessary once files live there.
9. **Intertwined with scheme-2 references.** Body wikilinks `[[slug]]` + any path refs must
   resolve post-flatten. The migration rewrites on-disk refs; the flatten code resolves by
   id-prefix; the rename verb cascades. Three coordinated tasks, not one.
10. **OCC/flock carries over — simpler.** In-place writes keep the version-CAS + flock; no more
    relocation → no dual-file window. Don't reintroduce a relocation path.

## Sequenced implementation (each step green within the branch)

1. **Flat filename + one-dir scan.** `<id>-<slug>.md` parsing (fixed 12-char id); a flat
   candidate scan mirroring `epicstore`'s `markdownCandidates(dir, "")` for tasks + audits.
2. **Resolution: id-prefix then slug** over the flat dir; retire the across-dirs dup-slug class.
3. **Move = in-place edit.** `moveTask`/`MoveAudit` → `updateFrontmatter({status/bucket})` in
   place (model on `MoveEpic`), keeping OCC verify→write + flock; delete relocation + mkdir +
   the dual-file recovery net.
4. **Delete the mirror.** Remove Misfiled/FolderStatus/*FellBack-fallback/the misfiled fixer;
   parse status/bucket from frontmatter only (missing → FileProblem + lint).
5. **Watch + completion off frontmatter.** `WatchPaths` → 3 entity dirs; completion reads
   frontmatter for status/bucket filters.
6. **Wire + TUI contract.** Drop `misfiled`/`declared_status`/Summary `misfiled`; bump
   `schema_version`; remove TUI misfiled markers/legend; regen goldens + `docs/cli`.
7. **Migration + refs (companion tasks).** Land code + run the migration on this repo +
   desirelines; coordinate scheme-2 ref rewrites.
8. **Docs.** ARCHITECTURE (flat, id-keyed, no mirror), CLAUDE.md (retire `tasks/<status>/`
   mirror language), ADR-0003 §4 status.

## Test pins

- resolve by id-prefix; by slug; slug-collision disambiguated by id.
- `<id>-<slug>.md` parse (12-char id + dashed slug); filename-id ≠ frontmatter-id → lint.
- move = in-place frontmatter edit (file path unchanged, no relocation); OCC + flock still fire
  (extend `TestConcurrentAppends_NoLostUpdates`-style coverage to the in-place move).
- missing/unknown status → FileProblem + lint (no folder fallback).
- `WatchPaths` = 3 dirs; completion status-filter reads frontmatter.
- wire: no `misfiled`/`declared_status`; `schema_version` bumped; goldens regen.
- dup-slug class gone (two same-slug different-id files both resolve by id).

## Reversibility

Phase A was a one-line revert (dirs kept). Phase B is the irreversible cutover: revert =
`git revert` the code + restore the pre-migration tree from git. ALWAYS run the migration on a
copy first, and commit the pre-migration tree so git is a clean undo.

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- ADR [[0003-stable-key-id-addressed-storage]] §2, §4, §6 (migration)
- Companion: [[one-time-migration-script-this-repo-desirelines]] (the data cutover) ·
  [[scheme-2-references-and-rename-verb-with-link-cascade]] (id-prefix refs, rename cascade)
- Built on: [[version-aware-occ-content-hash-token-and-plain-retry]] — the in-place write keeps
  its OCC + flock; the dual-file window it guarded is gone once moves stop relocating.
