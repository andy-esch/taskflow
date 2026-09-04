---
schema: 1
id: 6g6wdvfp2ksa
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Expose shared priority, tier, and effort metadata when a Thread has multiple independently eligible candidates.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, cli, ux, dogfood]
created: "2026-09-04"
depends_on: [6g6scc9jgxae]
updated_at: "2026-09-04"
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
- Advance and document the machine schema if the shared projection changes.

## Acceptance criteria

- [ ] With multiple independently eligible members, human frontier output shows each candidate's
      priority, tier, and effort without obscuring its stable ID, slug, or lifecycle status.
- [ ] JSON carries the same typed metadata from the shared core projection, and contract/schema
      fixtures cover non-default and unknown values.
- [ ] CLI and TUI adapters do not rescan storage, parse locations, or independently reconstruct the
      task metadata.
- [ ] Existing frontier membership and ordering are byte-for-byte or semantically unchanged apart
      from the new presentation fields.
- [ ] Broken/degraded graphs still fail closed, active work remains distinct from dispatchable work,
      and external gates are not misrepresented as candidates.
- [ ] The projection is reusable by a later TUI or web presentation without importing CLI types or
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
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Release checkpoint [v0.19.0 TUI preview](6g6scc9jgxae-cut-v0.19.0-as-a-tui-threads-preview.md)
- Earlier presentation correction [Show in-flight work](6g5fthzwbeq1-show-in-flight-work-before-dispatchable-thread-frontier-members.md)
