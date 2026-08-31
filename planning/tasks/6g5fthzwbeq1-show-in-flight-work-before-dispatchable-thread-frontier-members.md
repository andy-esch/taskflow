---
schema: 1
id: 6g5fthzwbeq1
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Make human Thread frontier output show active members first and distinguish no additional dispatchable work without changing frontier semantics.
effort: 1-2 hours
tier: 4
priority: low
autonomy_level: 3
tags: [threads, cli, ux]
created: "2026-08-31"
---

# Show in-flight work before dispatchable Thread frontier members

## Objective

Make the human `thread frontier` view useful while a Thread is actively progressing. Preserve
frontier as the set of pending members eligible to start, but show in-flight members before that
dispatchable set so an empty frontier does not read like an idle or completed initiative.

## Scope

- Derive active rows from existing member lifecycle roles; do not add another persisted or derived
  status.
- Render in-flight members before eligible frontier members, retaining deterministic stable-ID
  ordering within each group.
- When active work exists and the eligible set is empty, say `no additional dispatchable member
  tasks`; retain `no dispatchable member tasks` when there is no active work.
- Include each active member's gate state so a forced `in-flight/blocked` or broken state remains
  visible rather than receiving a reassuring active-work gloss.

## Acceptance criteria

- [ ] Human output lists in-flight members before eligible frontier members without inserting them
  into the frontier or changing `eligible`.
- [ ] An active-only Thread reports its active work followed by `no additional dispatchable member
  tasks`.
- [ ] A Thread with neither active nor eligible members retains the existing `no dispatchable member
  tasks` diagnosis.
- [ ] Blocked or broken in-flight members expose their gate state, and existing graph/Thread problem
  diagnostics still render.
- [ ] `thread frontier --json`, `ThreadView.Frontier`, lifecycle authorization, and wire schema remain
  unchanged.

## Stress tests

- Active-only, active plus eligible, multiple active members, forced blocked in-flight work, no
  active/eligible work, and unhealthy projection evidence.

## Out of scope

- Redefining frontier, selecting or ranking the next blocked task, changing `thread list` or `thread
  show`, adding graph traversal, and altering task or Thread lifecycle semantics.

## Sequencing

Independent of deterministic graph rendering and safe to implement separately. It belongs to the
production Thread because it was found by using that Thread as the work-discovery surface.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
