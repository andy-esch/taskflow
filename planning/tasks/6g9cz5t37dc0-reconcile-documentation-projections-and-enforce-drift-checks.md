---
schema: 1
id: 6g9cz5t37dc0
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Remove contradictory copies across guides and generated references, then add proportionate checks for links, inventories, examples, and agent routing.
effort: 1 day
tier: 3
priority: medium
autonomy_level: 4
tags: [documentation, ci, agents, maintainability]
created: "2026-09-12"
depends_on: [6g63hjm7cp6w, 6g6x7e2ef37r, 6g9cz5sv6kck]
updated_at: "2026-09-13"
---
# Reconcile documentation projections and enforce drift checks

## Objective

Finish the documentation restructuring with a repository-wide reconciliation pass and small, high-signal checks that keep each projection aligned with its canonical owner.

Known local drift includes the obsolete Thread limitation in `CLAUDE.md`, the “one-screen” label for an 8,000-word architecture document, dated package/status snapshots, and a role table that omits or inconsistently places newer packages. The fix should prevent recurrence without snapshot-testing prose.

## Acceptance criteria

- [ ] README, agent guides, architecture guides, package docs, generated CLI reference, schema guidance, ADR links, and compatibility documents are checked against the source-of-truth matrix and contradictory copies are removed or corrected.
- [ ] Stable links and navigation work after the architecture split, including links from code comments and package docs where appropriate.
- [ ] CI or a documented `just` target verifies generated documentation and package/import inventory freshness, compiles Go examples, and performs a proportionate internal-link check.
- [ ] Agent-specific wrappers or duplicated safety digests have one explicit sync mechanism; mutable command and feature inventories are linked or generated rather than copied.
- [ ] Drift failures explain the canonical source and remediation command instead of reporting only a generic diff.
- [ ] `just test`, `just lint`, documentation generation/checks, `git diff --check`, and planning lint pass.

## Out of scope

- Brittle snapshots of prose, heading order, or line wrapping.
- A new documentation-site framework or hosted portal.
- Treating historical planning audits as normative project documentation.
- Broad code refactors unrelated to documentation ownership or its executable checks.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
