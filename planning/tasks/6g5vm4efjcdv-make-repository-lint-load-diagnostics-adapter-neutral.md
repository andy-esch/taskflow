---
schema: 1
id: 6g5vm4efjcdv
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Replace the shared lint unreadable-file bucket with record identity and optional repair locations before remote or served adapters consume it.
effort: 1-2 days
tier: 3
priority: low
autonomy_level: 3
tags: [architecture, diagnostics, lint, ports]
created: "2026-09-01"
updated_at: "2026-09-04"
depends_on: [6g5rxq1ravd3, 6g6scc9jgxae]
---
# Make repository lint load diagnostics adapter-neutral

## Objective

Give core lint one taskflow-owned failed-record contract so local files remain actionable while remote and served adapters retain entity kind and stable identity without invented paths.

## Scope

- Define a neutral lint load problem with entity kind, optional stable ID and slug, optional repair location, and message.
- Adapt task, epic, research, and Thread load failures at secondary-adapter boundaries without changing semantic lint findings.
- Map the neutral contract deliberately in CLI human and machine output and advance the machine schema when implemented.
- Preserve local fix and guarded mutation FileProblem flows where exact filesystem snapshots remain required.

## Acceptance criteria

- [ ] Pathless task, epic, research, and Thread failures retain kind and recoverable identity through Service.Lint and machine output.
- [ ] Local filesystem lint keeps exact actionable locations and messages without extra entity scans.
- [ ] Core lint ports and results no longer require domain.FileProblem or parse locations for identity.
- [ ] Human output, partial-failure exit behavior, schema comments, generated schema, compatibility notes, and fixtures are updated together.
- [ ] Fix and guarded mutation paths retain their existing authoritative local-file behavior.

## Stress tests

Mixed readable and unreadable entity kinds, pathless problems, misleading locations with explicit identity, invalid filenames, duplicate identities, deterministic ordering, and local scan counts.

## Out of scope

Changing lint rules, redesigning ordinary entity list envelopes, implementing a database or HTTP store, or changing guarded mutation policy.

## Sequencing

Begin after the v0.19.0 TUI preview so its shared diagnostic and wire-contract decisions form a
deliberate post-release boundary. This task owns the multi-entity vocabulary first; only then does
`preserve-portable-load-diagnostics-in-board-and-status` carry it through dashboard projections.

## Related

- Thread diagnostic predecessor: make-thread-read-diagnostics-adapter-neutral
- Task graph diagnostic precedent: make-task-graph-load-diagnostics-adapter-neutral
