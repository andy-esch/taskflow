---
schema: 1
id: 6g5vm4efjcdv
status: completed
epic: 21-code-quality-architecture-hardening
description: Replace the shared lint unreadable-file bucket with record identity and optional repair locations before remote or served adapters consume it.
effort: 1-2 days
tier: 3
priority: low
autonomy_level: 3
tags: [architecture, diagnostics, lint, ports]
created: "2026-09-01"
updated_at: "2026-09-23"
depends_on: [6g5rxq1ravd3, 6g6scc9jgxae]
started_at: "2026-09-23"
completed_at: "2026-09-23"
---
# Make repository lint load diagnostics adapter-neutral

## Objective

Give core lint one taskflow-owned failed-record contract so local files remain actionable while remote and served adapters retain entity kind and stable identity without invented paths.

## Scope

- Define a neutral lint load problem with entity kind, optional stable ID and slug, optional repair location, and message.
- Adapt task, epic, audit, research, and Thread load failures at secondary-adapter boundaries without changing semantic lint findings.
- Map the neutral contract deliberately in CLI human and machine output and advance the machine schema when implemented.
- Preserve local fix and guarded mutation FileProblem flows where exact filesystem snapshots remain required.

## Acceptance criteria

- [x] Pathless task, epic, audit, research, and Thread failures retain kind and recoverable identity through Service.Lint and machine output.
- [x] Local filesystem lint keeps exact actionable locations and messages without extra entity scans.
- [x] Core lint ports and results no longer require domain.FileProblem or parse locations for identity.
- [x] Human output, partial-failure exit behavior, schema comments, generated schema, compatibility notes, and fixtures are updated together.
- [x] Fix and guarded mutation paths retain their existing authoritative local-file behavior.

## Stress tests

Mixed readable and unreadable entity kinds, pathless problems, misleading locations with explicit identity, invalid filenames, duplicate identities, deterministic ordering, and local scan counts.

## Implementation notes

- `core.LintSource` is a consumer-owned port rather than an expansion of the aggregate `Store`; `audit lint` can request its one audit scan without touching unrelated kinds.
- `core.LintLoadProblem` carries entity kind, optional ID/slug, optional location, a local-path marker, and message. Core never reconstructs identity from location.
- The filesystem adapter wraps the existing body-aware task, epic, audit, and research scans, so the portable mapping adds no second read pass. Thread failures enter the same lint result from their existing neutral `ThreadRead` snapshot.
- Machine schema 1.73 is additive: `entity_kind`, `entity_id`, `entity_slug`, and `location` enrich lint/fix unreadable records. The required historical `path` remains exact for local diagnostics and is empty for a pathless source.
- `domain.FileProblem` remains deliberate on ordinary local list, fix, and guarded mutation contracts whose job is exact filesystem repair; this task does not flatten those distinct contracts into the lint port.

## Out of scope

Changing lint rules, redesigning ordinary entity list envelopes, implementing a database or HTTP store, or changing guarded mutation policy.

## Sequencing

Begin after the v0.19.0 TUI preview so its shared diagnostic and wire-contract decisions form a
deliberate post-release boundary. This task owns the multi-entity vocabulary first; only then does
`preserve-portable-load-diagnostics-in-board-and-status` carry it through dashboard projections.

## Related

- Thread diagnostic predecessor: make-thread-read-diagnostics-adapter-neutral
- Task graph diagnostic precedent: make-task-graph-load-diagnostics-adapter-neutral
- Delivery Thread: [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)

Cross-referenced by audit 2026-09-22-correctness-and-errors: M1 (PARTIAL overlap). That finding is a lint *rule* gap — `Service.Lint`'s cross-kind task/Thread id-collision check builds its identity sets from readable records only, so a collision stops being reported once either document is malformed, even though the adapter already recovers the filename id and the duplicate-id check in the same loop consumes it. This task's "Out of scope" still excludes changing lint rules. The repair is now tracked separately in [unreadable cross-kind ID collision lint](6gcqz5aefjjf-lint-cross-kind-task-and-thread-id-collisions-on-unreadable-records.md), with an explicit dependency on this task so both changes converge on one neutral failed-record identity contract.

## Progress log

- 2026-09-23: A boundary audit confirmed this consumer-owned lint port is the intended pattern and
  scoped the remaining portability work into the adapter-neutral data-access Thread. Follow-ups own
  path-independent graph-lint attribution, portable dashboard/list diagnostics, path-free audit
  finding reads, semantic/local path separation, CLI use-case routing, and executable controller
  import boundaries rather than widening this focused unreadable-record change.
- 2026-09-23: Implemented and locally validated the adapter-neutral lint read boundary. Portable
  core tests cover pathless task, epic, audit, research, and Thread failures; filesystem tests cover
  exact local identities, locations, messages, and one existing scan per kind; render/wire tests
  cover identity-first human output plus local and pathless schema 1.73 JSON. `go test -race ./...`,
  `golangci-lint run ./...`, generated-schema freshness, `git diff --check`, and repository
  `lint --json` are clean. The task remains in progress pending the normal adversarial review pass.
