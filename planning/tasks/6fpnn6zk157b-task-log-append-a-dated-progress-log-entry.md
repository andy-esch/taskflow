---
schema: 1
id: 6fpnn6zk157b
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'task log: append a dated progress entry (structure-aware body write); blocked on choosing the canonical progress-section shape'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, agents, ux, dx]
created: "2026-07-16"
updated_at: "2026-09-20"
audited: "2026-09-20"
audit_sources: [2026-09-20-weekly-task-sweep]
---
> ⚠️ **Spun off 2026-07-16** from
> [structure-aware-body-mutation-and-metadata-reads](6fpfcecdymca-structure-aware-body-mutation-and-metadata-reads.md)
> (item 1 of that batch). The other three items — `task ac`, `task path`/`info`,
> and the section reads — shipped; this one was parked because it needs a design
> decision *first*, so it gets its own home rather than holding the batch open.

## Objective

`task log <slug> --body "…"` (and `--body-file -`): append a **dated** entry to a
task's progress section — the single most frequent body write an agent makes. Today
`task append` is structure-blind: it re-declares `## Progress Log` when one already
exists, producing a duplicate header that then needs a hand-merge. `task log` should
(a) append under the EXISTING progress section, creating it only if absent, and
(b) auto-stamp today's date.

## Decide first — the canonical progress-section shape

`task log` can't be built until we settle what a "progress entry" looks like in THIS
repo, because the two conventions in play disagree, and hard-coding either fights the
other corpus:

- **Single `## Progress Log` + dated bullets** (the requesting agent's repo):

  ```
  ## Progress Log
  - 2026-07-16: shipped X
  - 2026-07-17: shipped Y
  ```

  Compact; one section; `task log` appends a bullet.

- **Per-entry dated headings** (this repo's de-facto corpus, incl. the batch task
  this was spun off from):

  ```
  ## Progress (2026-07-16)
  Shipped X.

  ## Progress (2026-07-17)
  Shipped Y.
  ```

  Matches existing tasks; `task log` appends a new subsection.

**The shape decision blocks implementation** — it is the first acceptance criterion.

## Design notes (once the shape is decided)

- Route the write through the existing `FS.EditBody` / surgical `yaml.Node` path (as
  `task append` / `task set --body` / `task ac` do) so frontmatter, comments, key
  order, and the parse-before-write + compare-and-swap discipline all survive.
- Reuse the fence-aware body-structure model in `internal/domain/body.go`
  (`Section`, `scanAcceptanceCheckboxes`) — the same "structure as first-class"
  muscle: locate the progress section, append under it, create if absent. Note the
  precise-vs-substring lesson from the AC lint guard: a `## Progress …` heading that
  mentions a keyword must not collide with section detection.
- Auto-stamp the date from the injected clock (`s.now()`), never `time.Now()`.

## Acceptance criteria

- [ ] The canonical progress-section shape is decided and recorded here.
- [ ] `task log <slug> --body|--body-file -` appends a dated entry under the existing
      progress section, creating it only if absent — never a duplicate header. Atomic
      (via `EditBody`); frontmatter preserved; date auto-stamped from the clock.
- [ ] `--json` returns the `task_mutation` envelope; `--dry-run` previews without
      writing.
- [ ] Fence-aware (a `## Progress` inside a code block isn't the target); suite + lint
      green; docs (`docgen`) + README updated.

## Related

- Spun off from [structure-aware-body-mutation-and-metadata-reads](6fpfcecdymca-structure-aware-body-mutation-and-metadata-reads.md).
- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md).

## Sweep verification (2026-09-20)

Automated weekly sweep re-read this task against `internal/cli`,
`internal/store`, `internal/domain`, and the `planning/tasks/` corpus at
`934e1cf`. The feature is still unbuilt and the design notes still hold; the one
substantive update is that the corpus evidence behind the blocking decision has
inverted.

**Still unbuilt — verified accurate.** `tskflwctl task --help` lists `ac`,
`append`, `edit`, `set` and no `log`. `task append` is still structure-blind, so
the duplicate-header problem in the Objective is exactly as described.

**Design-note references verified accurate (2026-09-20):**

- `FS.EditBody` → `internal/store/body.go:134` (still the surgical
  parse-before-write + CAS path).
- `domain.Section` → `internal/domain/body.go:216` and
  `scanAcceptanceCheckboxes` → `internal/domain/body.go:251`; both still live in
  the fence-aware body model, alongside `UnterminatedFence` (`body.go:173`),
  which is the fence-awareness primitive AC #4 needs.
- The injected clock is still the idiom — `s.now()` at
  `internal/core/dependency_operations.go:129`, `dependency_repair.go:612`,
  `finding.go:251`.
- The parent batch task `6fpfcecdymca-structure-aware-body-mutation-and-metadata-reads`
  is `completed`, confirming the spin-off note: the other three items shipped and
  only this one is outstanding.

**The corpus evidence for the shape decision has inverted.** This task records
per-entry dated headings (`## Progress (YYYY-MM-DD)`) as "this repo's de-facto
corpus". Counted across `planning/tasks/` at `934e1cf`:

| shape | files |
|---|---|
| `## Progress Log` | 21 |
| `## Implementation progress` | 19 |
| `## Progress (YYYY-MM-DD)` | 7 |

> **Lean (AC #1):** `## Progress Log` is now the majority section header, 21 to
> 7 over per-entry dated headings — the opposite of what this task assumed, so
> the "hard-coding either fights the other corpus" objection now cuts toward
> the single-section shape rather than away from it. Two caveats keep this a
> lean and not a decision: (a) `## Implementation progress` (19 files) is a
> third header spelling that any implementation has to recognise or migrate, and
> (b) within `## Progress Log` the *entry* format is itself split — plain dated
> bullets (`- 2026-06-07: …`) in some files, bold dated sub-entries
> (`**2026-06-07 — title.**` plus bullets) in the older ones. So the decision is
> really two: which header `task log` targets, and which entry form it emits.
> Left for the human — the box stays unticked.

## Progress Log

- 2026-09-20: automated weekly sweep — feature still unbuilt, all design-note references re-verified; corpus recount inverts the shape assumption (## Progress Log 21 files vs ## Progress (date) 7), recorded as a Lean under the blocking decision.
