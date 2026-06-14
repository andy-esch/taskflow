---
status: ready-to-start
epic: 17-pm-go-cli
description: 'Follow-ups deferred from the 2026-06-13 audit: epic numbering, scan dedup, Move CAS, epic JSON shape, TUI/completion layout dup, SetFields bool params'
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, refactor, tech-debt]
created: "2026-06-13"
---

# Address deferred code-audit findings (numbering, dedup, CAS, JSON layout)

## Objective

The 2026-06-13 codebase audit (`planning/audits/open/2026-06-13-codebase-quality-architecture.md`)
fixed the high/quick findings inline; this task collects the ones deliberately
deferred because they're larger refactors or design decisions rather than
one-line fixes. None is urgent — the code is correct today — but each removes a
real drift or robustness gap. Do them as independent commits.

## Acceptance criteria

- [ ] **M3 — epic numbering.** `nextEpicNumber` + `CreateEpic` can mint duplicate
      `NN-` prefixes (two `epic new` with different slugs both get `03-…`; `O_EXCL`
      only guards identical paths), and `%02d` mis-sorts once past 99 epics. Detect
      `NN-` prefix collisions and bump (or retry-on-collision); widen/stop the
      zero-pad. (`store/create.go`)
- [ ] **L4 — `Move` compare-and-swap.** `Move` has no re-resolve before
      write+remove, so a concurrent relocation leaves a duplicate in two status
      dirs. Apply the same CAS guard `SetFields` already has (see
      `harden_test.go`), or document Move as not concurrency-safe. (`store/fsstore.go`)
- [ ] **L2 — scan dedup.** `markdownDoc` already unified the entry filter; finish
      the job by extracting the repeated `ReadDir → skip → ReadFile → parse →
      FileProblem` loop (3 List sites + candidate gatherers) into one generic
      `scanDir` helper. (`store/fsstore.go`, `epicstore.go`, `auditstore.go`)
- [ ] **L6 — epic JSON shape.** `epic list` (`epicJSON`) carries the task rollup;
      `epic show` (`epicMetaJSON`) omits it — two shapes under one schema_version.
      Embed `epicMetaJSON` in `epicJSON` so meta fields are identical and rollup is
      purely additive. (`cli/render/render.go`)
- [ ] **L8 — layout duplication.** The TUI watcher (`tui/watch.go`) and shell
      completion (`cli/completion.go`) each reconstruct `tasks/<status>` /
      `audits/<bucket>` / `epics/` directly. Expose the canonical dir set from the
      store (e.g. `store.WatchDirs(root)`) and have both consume it. Low priority.
- [ ] **M6 (optional) — `SetFields` bool params.** `Service.SetFields` ends in
      `force, dryRun` — two adjacent same-typed positionals with no transpose
      guard. The Store port mutators are single-bool (idiomatic, leave them); only
      revisit if a third flag lands, via a small `MutateOpts` value. Judgment call —
      may be a WONTFIX. (`core/service.go`)
- [ ] `just build && just test && just lint` (or the `go` equivalents) green.

## Out of scope

- The findings already fixed inline (H1, H2, M1, M2, M4, M5, L1, L3, L5, L7, L9) —
  see the audit's progress log.
- The TUI help-scroll bug (M5) — already fixed.

## Related

- Epic [[17-pm-go-cli]]
- Audit: `planning/audits/open/2026-06-13-codebase-quality-architecture.md`
