---
schema: 1
id: 6gbn4g1eypzs
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Give Thread list the shared output and column-projection contract without flattening its richer typed envelope.
effort: S
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, threads, json, agents]
created: "2026-09-19"
---

# Make Thread list projectable like other planning entity lists

## Objective

Give `thread list` the same bounded-output vocabulary used by other first-class planning entities:
human table/name views, CSV, and caller-selected projected JSON. Retain the full typed Thread
envelope as the authoritative rich projection rather than flattening topology and diagnostics into
columns.

## Scope

- Add `-o/--output` and `-c/--columns` through the shared list-mode conventions, including help,
  validation, completion, and structured JSON errors.
- Define one Thread column registry with stable `id`, human `slug`, lifecycle status, concise
  progress fields, frontier count, and description; keep slug as the quiet display handle.
- Preserve current default human output and full `--json` envelope behavior when columns are not
  requested, including repository graph and projection diagnostics.
- Project canonical keys in caller order and make any compact health/progress values explicit
  rather than leaking formatted table strings.

## Acceptance criteria

- [ ] `thread list` supports the established table, CSV, name, and JSON output modes plus validated
      column selection without changing its default human presentation.
- [ ] Projected JSON exposes stable Thread IDs and caller-ordered canonical keys; aliases, if any,
      are explicit and collision-checked.
- [ ] Full `thread list --json` retains its typed members, gates, graph health, projection health,
      and diagnostics rather than being routed through the compact registry.
- [ ] Help, completion, table/CSV/projected-JSON goldens, registry invariants, and hostile selector
      tests pin the contract.
- [ ] The additive machine-contract revision and generated docs are updated; focused tests, full
      tests, lint, planning lint, and diff checks pass.

## Out of scope

- Pagination or query filtering; the bounded-list task owns the cross-entity design.
- Projecting every nested Thread field into a scalar column.
- Changing Thread graph, plan, frontier, lifecycle, or preview semantics.

## Related

- Audit [Machine contract, M3](../audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md)
- ADR [Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Task [Bound and page agent-facing list queries](6gbn4g1j40pj-bound-and-page-agent-facing-list-queries.md)
