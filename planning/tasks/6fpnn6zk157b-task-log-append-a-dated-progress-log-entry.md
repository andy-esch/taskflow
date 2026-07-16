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
