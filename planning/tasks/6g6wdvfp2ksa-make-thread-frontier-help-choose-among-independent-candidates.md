---
schema: 1
id: 6g6wdvfp2ksa
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Expose shared priority, tier, and effort metadata when a Thread has multiple independently eligible candidates.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, cli, ux, dogfood]
created: "2026-09-04"
depends_on: [6g6scc9jgxae]
updated_at: "2026-10-11"
started_at: "2026-10-11"
completed_at: "2026-10-11"
---

# Make Thread frontier help choose among independent candidates

## Objective

Make `thread frontier` useful when graph truth permits several independent tasks at once. Preserve
the frontier as an eligibility projection while carrying enough taskflow-owned planning metadata
for humans and future adapters to distinguish a high-priority foundation task from a low-priority
experiment without issuing a separate `task show` for every candidate.

## Scope

- Carry priority, tier, and effort through the shared core Thread projection used by adapters; the
  CLI must not reload task files or derive these values from rendered text.
- Render compact planning metadata beside eligible candidates in human frontier output and expose
  the same values deliberately in JSON/wire output.
- Keep graph eligibility, in-flight presentation, deterministic ordering, member/external-gate
  roles, and fail-closed health behavior unchanged.
- Treat planning metadata as decision context, not a computed recommendation or scheduler score;
  represent `Unknown` effort and any valid boundary values explicitly.
- Advance and classify the machine-contract revision for wire changes, even when the core
  projection already supplies the data.

### Implementation boundary (2026-10-11)

The current core `ThreadTaskView` already carries the complete task, including priority, tier,
and effort. Reuse that projection rather than adding a second metadata lookup or a new port.
Human output will retain the candidate identity line and add a compact lifecycle/planning line;
in-flight work stays first, and candidates stay ordered by stable ID, not priority.

Expose a typed `planning` object on wire Thread task views, including members, frontier, and
external prerequisites. Keep the compact dependency-query task DTO unchanged. Preserve raw
values (`Unknown` included); empty strings and tier `0` mean unrecorded metadata, shown as
`unset` in human output. This is additive machine-contract context, not a new scheduling policy.

## Acceptance criteria

- [x] With multiple independently eligible members, human frontier output shows each candidate's
      priority, tier, and effort without obscuring its stable ID, slug, or lifecycle status.
- [x] JSON carries the same typed metadata from the shared core projection, and contract/schema
      fixtures cover non-default and unknown values.
- [x] CLI and TUI adapters do not rescan storage, parse locations, or independently reconstruct the
      task metadata.
- [x] Existing frontier membership and ordering are byte-for-byte or semantically unchanged apart
      from the new presentation fields.
- [x] Broken/degraded graphs still fail closed, active work remains distinct from dispatchable work,
      and external gates are not misrepresented as candidates.
- [x] The projection is reusable by a later TUI or web presentation without importing CLI types or
      terminal formatting policy.

## Stress tests

- Several candidates with mixed priorities/tiers, equal metadata, `Unknown` effort, long slugs,
  narrow terminals, active plus eligible work, external gates, unhealthy graphs, and stable JSON
  ordering.

## Out of scope

- Automatically ranking or selecting work, defining organization-wide priority semantics, critical
  path or forecasting calculations, and redesigning the TUI frontier presentation.
- Changing which task statuses or graph states are eligible to start.

## Sequencing

Follow the v0.19.0 checkpoint, where the Thread deliberately fans out into several independent
branches and the missing decision context becomes concrete. This is a contained dogfood usability
slice; it does not block guarded repair, diagnostic portability, or the spatial prototype.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
- Release checkpoint [v0.19.0 TUI preview](6g6scc9jgxae-cut-v0.19.0-as-a-tui-threads-preview.md)
- Earlier presentation correction [Show in-flight work](6g5fthzwbeq1-show-in-flight-work-before-dispatchable-thread-frontier-members.md)

## Implementation outcome (2026-10-11)

Reused `ThreadTaskView.Task`; no new port, storage scan, or scheduling policy.
Human candidates now have a separate lifecycle/priority/tier/effort line, preserving stable
identity lines and in-flight-first presentation. Human-only formatting keeps free-form metadata
on one control-sequence-free line and labels absent values `unset`; raw wire values stay intact.

Revision **1.86 (ADDITIVE)** adds `planning: {priority, tier, effort}` to shared Thread task
views, including members, frontier, and external prerequisites. Compact dependency-query tasks
are unchanged. Generated schema comments, machine goldens, command docs, and README recipes agree.

Validation: `just test` (full race suite), `just lint`, and `just build` pass.
CLI documentation regeneration is byte-stable. Core pathless-reader tests prove one graph read;
semantic schema tests and real-adapter goldens cover nondefault/unknown/unset values, stable order,
active and external roles, degraded/broken fail-closed behavior, and terminal-safe display.
The rebuilt binary was dogfooded against `refine-thread-and-tui-navigation`.

Stress testing found a pre-existing legacy-only degradation diagnosis gap. It is tracked by
[6gjn7axv0gmy](6gjn7axv0gmy-explain-legacy-dependency-degradation-in-thread-projections.md),
added as an independent member of the same Thread; graph policy remains unchanged here.

## Review checkpoint (2026-10-11)

The cleanup pass shortened frontier help, clarified that wire changes require a classified revision
even when core already carries the data, and made the missing-member regression prove that cause
rather than accidentally failing on unsorted membership. Narrow streaming output coverage retains
complete handles and separate planning lines. Full race tests, lint, and build remain green.

One focused independent contract/CLI review completed in
[the Codex audit](../audits/6gjnjssabsy2-2026-10-11-thread-frontier-planning-context-implementation-codex.md).
No new reachable defect was reproduced. Evidence includes populated read/mutation/repair receipts,
exact-schema validation, a semantic golden-delta inventory, hostile text through the real reader,
and killed/restored mutations for tier preservation, sanitization, and coordinated health bypasses.
The owner verified the isolated-workspace transfer logs and retained the review's bounded limits.

Post-review changes only strengthen tests: shared Thread-update/task-impact before/after mappings
now retain nondefault and raw planning values in permanent regressions, and the narrow streaming
test covers OSC terminators, forged rows, carriage returns, whitespace-only input, and Unicode.
The reviewed production behavior is unchanged; full race tests, lint, and build pass again.
The legacy-degradation diagnostic omission remains tracked separately by `6gjn7axv0gmy`.
