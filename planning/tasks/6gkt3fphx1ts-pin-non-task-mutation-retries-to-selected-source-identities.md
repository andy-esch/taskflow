---
schema: 1
id: 6gkt3fphx1ts
status: completed
epic: 21-code-quality-architecture-hardening
description: Stop audit, research and epic retry loops from reselecting another record after alias drift or deletion.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, core, store, concurrency]
created: "2026-10-14"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
updated_at: "2026-10-16"
started_at: "2026-10-15"
completed_at: "2026-10-16"
---
# Pin non-task mutation retries to selected source identities

## Objective and evidence

Extend the strict selected-identity rule to ordinary audit, research and epic
mutations without introducing a universal entity mutation framework.

The owner independently reproduced audit append and research field updates mutating
a pre-existing `alpha-two` after `alpha` was renamed to `omega` in their pre-lock
CAS seams. Both commands returned nil. The source retained its bytes and the sibling
changed. This is deterministic real-store evidence, not measured concurrent frequency.
Probe: `/private/tmp/taskflow-target-neighbors.VNxBO5/neighbor_retry_probe_test.go`
(temporary evidence, not a durable test dependency).

## Scope

- Audit close/defer/reopen, append, finding creation/edit and header canonicalization;
  research field updates/append; epic field/lifecycle updates.
- Resolve initial references once, then use strict adapter-established source identities
  across attempts. A disappeared identity must never fall back to another human label.
- Audit header canonicalization already selected a snapshot occurrence: use its
  `Source.ID`, not the domain slug. Initial duplicate-label selection must remain loud.
- Follow the task mutation target pattern, keeping entity-specific ports and policy.
  Preserve original-reference ergonomics, preview/no-op behavior, authorization,
  timestamp capture, retry bounds and source-preserving CAS.
- Do not silently include interactive editors, creates, graph planners or new wire
  shapes. Record any separately confirmed exposure at those boundaries.

### Earlier selected-caller boundary (2026-10-15)

The [TUI task-action work](6gkt3fq0pv26-preserve-selected-tui-task-identities-through-service-actions.md)
also reproduced an audit lifecycle command restarting selection before core retries:
queue `moveAudit` with the selected source ID, delete that record, then create a
different-ID audit whose human slug equals the missing ID. The real command returned
`movedMsg` and closed the impostor. Probe retained locally at
`/private/tmp/taskflow-selected-action-neighbors.uYLhPU/selected_identity_neighbors_probe_test.go`.

Include narrow selected-identity Service entry points for these non-task callers,
not just pinning after a fresh reference lookup. Inventory epic inline fields and
audit lifecycle delegates; only audit deletion/impostor drift is independently
reproduced here. Reuse ordinary coercion/lifecycle rules and preserve `tea.Cmd`,
session/generation guards and CLI matching. This overlaps the existing entity-specific
port work and needs no separate task or artificial dependency on task-only helpers.

## Acceptance criteria

- [x] Each ordinary retry family retains its first source identity through rename, alias reuse, deletion and ID-shaped label fallback; unrelated records stay byte-identical.
- [x] Audit bulk header repair targets snapshot identities even with duplicate display slugs, and retains partial progress and baseline-aware repair policy.
- [x] Empty identity and duplicate canonical IDs refuse before callbacks, preview/no-op returns or effects, using useful existing sentinels.
- [x] Portable and real-adapter regressions cover transient initial selection, one retry budget, normal concurrent edits, and the observed audit/research wrong-target cases.
- [x] Owning comments/guides and full validation accurately separate fixed families from remaining selection boundaries.
- [x] Already-selected audit lifecycle and other inventoried non-task TUI
  mutation callers do not restart fuzzy selection at Service entry; deletion and
  ID-shaped impostors never change another record.

## Sequencing

Use [strict task targeting](6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md)
as the proven pattern, not an artificial code dependency: these entity-specific fixes
can proceed independently. Feed concrete port/retry decisions into the shared-integrity
design rather than waiting for it or reopening completed legacy tasks.

## Safety-cluster delivery

Member of [Finish selected identity safety](../threads/6gm4h71hm1yt-finish-selected-identity-safety.md),
delivered in one PR alongside graph/Thread retry and selected-path fixes. The tasks
remain independent: delivery order is not a dependency chain or a broader storage redesign.

The initial regression was red in all 18 rename/deletion cases across nine mutation
families. Entity-specific targets and parse-free source selection now keep ordinary
audit, research and epic writes exact from first selection through CAS and retry;
audit repair uses snapshot source IDs. Selected audit/epic TUI callers use explicit
Service APIs. Epics intentionally retain their whole-stem identity semantics.
Duplicate-label audit header repair now has preview/write occurrence tests, and
duplicate canonical IDs refuse before audit callbacks or no-op/preview success.
Portable tests cover opaque identities, transient initial selection, empty conflict
results and one shared selection/write retry budget. Selected audit/epic TUI
commands have missing/impostor regressions through their actual command closures.

Owner validation: `just build`, `just test` (race-enabled), `just lint`,
`just docs-check` and planning `lint --links` pass. Acceptance criteria reflect
implemented owner evidence. Claude's review reproduced an ID-shaped epic-stem CAS
regression, now fixed by passing explicit adapter resolver keys and keeping editor
diagnostic labels separate. Whole-stem ordinary/selected writes and previews have
regressions; store-local ID-shaped audit impostors are also pinned directly.
Antigravity's report was delivered but deferred for evidence-integrity failures,
not counted as a second independent approval. All criteria are met and the task is
completed for PR delivery. Interactive mutation policy and a universal retry/durability
framework remain outside this fix.
