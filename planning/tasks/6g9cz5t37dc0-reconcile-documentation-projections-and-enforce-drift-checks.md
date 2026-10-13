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
depends_on: [6g63hjm7cp6w, 6g6x7e2ef37r, 6g9cz5sv6kck, 6ggxymjbf86r]
updated_at: "2026-10-12"
---
# Reconcile documentation projections and enforce drift checks

## Objective

Finish the documentation restructuring with a repository-wide reconciliation pass and small, high-signal checks that keep each projection aligned with its canonical owner.

Historical local drift included obsolete Thread limitations in `CLAUDE.md`, the “one-screen” label
for a large architecture document, dated package/status snapshots, and incomplete package roles.
The adapter-neutral migration corrected the command guidance and role map; its closeout corrected
the README label and added a short [architecture checklist](../../docs/ARCHITECTURE.md#planning-data-change-checklist).
Re-measure remaining drift rather than repeating those resolved examples. The fix should prevent
recurrence without snapshot-testing prose.

## Acceptance criteria

- [ ] README, agent guides, architecture guides, package docs, generated CLI reference, schema guidance, ADR links, and compatibility documents are checked against the source-of-truth matrix and contradictory copies are removed or corrected.
- [ ] Stable links and navigation work after the architecture split, including links from code comments and package docs where appropriate.
- [ ] CI or a documented `just` target verifies generated documentation and package/import inventory freshness, compiles Go examples, and performs a proportionate internal-link check.
- [ ] Shared agent guidance has one explicit owner in AGENTS.md; routing checks
  detect competing policy copies, and mutable command and feature inventories
  are linked or generated rather than copied.
- [ ] Drift failures explain the canonical source and remediation command instead of reporting only a generic diff.
- [ ] `just test`, `just lint`, documentation generation/checks, `git diff --check`, and planning lint pass.

## Out of scope

- Brittle snapshots of prose, heading order, or line wrapping.
- A new documentation-site framework or hosted portal.
- Treating historical planning audits as normative project documentation.
- Broad code refactors unrelated to documentation ownership or its executable checks.

The [focused pure-layer import guard](6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md)
now follows the dependency-policy ADR and feeds this final reconciliation. Verify guidance against
its actual enforcement, without claiming that import checks prove behavioral safety.

## Agent-routing handoff (2026-10-11)

Use the [ownership research](../research/6gjggjw01raw-documentation-ownership-and-portable-agent-routing.md)
and root `AGENTS.md` ownership matrix. Claude and Antigravity use native discovery; the repository
has neither `CLAUDE.md` nor `GEMINI.md`. Add proportionate checks for the shared root entry point,
working routes, and absence of competing shared-policy inventories. Avoid
snapshotting prose or requiring installed/authenticated model clients in CI. Keep documented
client-loading smoke checks separate from static link validity and behavioral adherence.

The maintainer chose native discovery after reviewing current Claude and Antigravity support.
Keep version/settings/ancestor-file caveats and Gemini CLI's differing default discovery in
the dated research, not another root instruction file.

Reconcile the remaining README product reference against command metadata and tested behavior;
the ownership slice only removes its duplicate contributor workflow. Preserve user-oriented
examples where useful, but route exhaustive mutable inventories to generated reference/schema.

Include the active routines in this reconciliation. The ownership slice repoints their live
entry points and removes unsafe dirty-checkout revert instructions; it is not a full workflow
audit. For example, `weekly-task-sweep.md` forbids tier/priority writes in its scope block but
allows them in step 8. Resolve that policy conflict explicitly rather than silently choosing
either instruction. Check repeated architecture summaries against their owners and verify
safe checkout handling without running routine writes against an unrelated dirty tree.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)

## Architecture extraction handoff (2026-10-12)

The short index and six focused guides now replace the monolithic reference; old section anchors
forward to their new owners. A bounded [navigation test](../../internal/tools/architecturecheck/architecture_test.go)
runs with just test, checking inline local routes/ATX anchors in the index, guides, AGENTS and README,
plus linked package-comment return paths. It deliberately is not a full Markdown/network checker,
semantic prose validator, package inventory verifier or import-map/example gate. Reuse that check
where appropriate; broaden coverage only for demonstrated missing routes. This task still owns
repository-wide reconciliation, root-policy duplication/routine consistency, executable inventory
and example integration. Do not tick its broader criteria just because the extraction test passes.

## External-review coverage handoff (2026-10-12)

The extraction review fixed false-positive index/back-links (fences, inline code and bare mentions), wrapped link labels, and retained index anchors referenced by tasks/Threads. The new incoming-route check derives links from planning/tasks and planning/threads, not a duplicate heading inventory. Historical audits with source-line citations, routines, reference-style Markdown/HTML, full package inventory and semantic claim-to-test agreement remain outside that bounded guard. Check those deliberately in this reconciliation rather than claiming a green navigation test proves repository-wide or semantic fidelity. A nearby userconfig compiler-enforcement claim was corrected to the actual depguard rule; continue auditing package overviews under the sibling task.
