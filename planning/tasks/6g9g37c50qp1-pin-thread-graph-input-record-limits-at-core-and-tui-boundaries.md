---
schema: 1
id: 6g9g37c50qp1
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Add regression coverage for duplicate-node and malformed dependency-rank record limits already enforced by Thread focus and spatial preflight.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [threads, graph, hardening, tests]
created: "2026-09-12"
depends_on: [6g86c7y6hn41]
updated_at: "2026-09-12"
---
# Pin Thread graph input record limits at core and TUI boundaries

## Objective

Turn the already-shipped malformed-record and duplicate-node guards used by Thread neighborhood
selection and spatial rendering into executable regression contracts.

## Acceptance criteria

- [ ] Core neighborhood-selector tests reject duplicate node identities without relying on later
  map collapse or a neighboring validator.
- [ ] TUI preflight tests cover too many raw dependency-rank records, too many task records across
  ranks, and repeated rank identities where applicable.
- [ ] Each assertion proves the guard fails before route layout or terminal-canvas materialization
  and retains the existing bounded diagnostic.
- [ ] Focused core/TUI suites, the race-enabled affected package suite, and planning lint pass.

## Related

- Tracked from finding L5 in the
  [Claude review of the one-hop Thread focus implementation](../audits/6g9fg4a8xvyc-2026-09-12-tui-one-hop-thread-focus-implementation-claude.md).
- Hardens existing limits; it does not change their values or add a new graph representation.

## Out of scope

- Increasing graph limits, changing repair behavior, or introducing a third-party graph library.
