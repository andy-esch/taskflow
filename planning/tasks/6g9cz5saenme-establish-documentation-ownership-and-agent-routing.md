---
schema: 1
id: 6g9cz5saenme
status: completed
epic: 21-code-quality-architecture-hardening
description: Define canonical owners, precedence, and task-based navigation so agents load the right project guidance without duplicating mutable facts.
effort: 4-6 hours
tier: 3
priority: medium
autonomy_level: 4
tags: [documentation, agents, architecture, dx]
created: "2026-09-12"
updated_at: "2026-10-12"
depends_on: [6gjthp9frrhq]
started_at: "2026-10-11"
completed_at: "2026-10-11"
---
# Establish documentation ownership and agent routing

## Objective

Give contributors and coding agents one short, tool-neutral entry point that explains which source owns each kind of project truth, how conflicts are handled, and which deeper document to read for a given change.

The starting root has a substantial `CLAUDE.md` but no `AGENTS.md` or `GEMINI.md`. It duplicates
architectural and planning behavior also present in README, `docs/ARCHITECTURE.md`, generated CLI
reference, schema output, and code. Earlier obsolete Thread guidance has already been corrected;
the remaining problem is competing ownership and discovery, not missing Thread commands.

## Acceptance criteria

- [x] A documented source-of-truth matrix assigns canonical ownership for build commands, architectural policy, package contracts, CLI behavior, entity authoring conventions, compatibility promises, decisions, and current planning status.
- [x] A concise root `AGENTS.md` provides tool-neutral non-negotiables, source precedence, validation commands, and a task-to-document routing table without copying large mutable inventories.
- [x] AGENTS.md is the sole shared root guide; CLAUDE.md is removed and no
  GEMINI.md wrapper or competing shared-policy inventory remains.
- [x] Conflict handling distinguishes implemented code/test evidence, generated
  registries/shapes and maintained prose, accepted decisions/promises, and
  planning history; revision tests do not replace compatibility review.
- [x] Any nested agent guidance is limited to genuinely local constraints and links to package or subsystem documentation for explanation.
- [x] README and contributor entry points route readers to the new guide without presenting another competing architecture summary.
- [x] The routed task authoring guidance names shipped guarded dependency
  commands, retains generic-write restrictions, and has regression evidence with
  a classified machine-contract revision.

## Out of scope

- Encoding every product behavior in agent instructions.
- Assuming that every coding agent supports the same tool-specific instruction filename or import syntax.
- Rewriting architecture guides, package documentation, or CLI reference content owned by sibling tasks.

## Research-informed implementation (2026-10-11)

The prerequisite [research task](6gjthp9frrhq-research-documentation-ownership-and-portable-agent-routing.md)
produced [a dated primary-source snapshot](../research/6gjggjw01raw-documentation-ownership-and-portable-agent-routing.md).
The maintainer selected shared `AGENTS.md` ownership and native Claude/Antigravity discovery,
without `CLAUDE.md` or `GEMINI.md` wrappers. Keep vendor loading details out of the shared guide and do not
change user settings. Distinguish implemented evidence, accepted
policy, and planning history rather than declaring that code always overrules a promise.

This slice may remove the duplicated README contributor-command roster and correct its missing
race-detector description. The architecture split and repository-wide reconciliation remain
separate tasks; route findings into those rather than expanding this task into a wholesale rewrite.

Dogfooding the authoring route found a stale `schema task` convention claiming guarded dependency
commands were future work. Correct the owning metadata and adjacent inaccurate registry/layout
comments. Since the convention is published in JSON, classify revision 1.88 as additive and update
the fixtures; no command behavior, payload shape, or persisted format changes. Pin the live
guarded route and generic-write restriction with a CLI regression.

## Delivery and verification (2026-10-11)

- Added the shared [AGENTS.md](../../AGENTS.md) ownership matrix, task-based router, conflict
  handling, and durable safety rules. Removed `CLAUDE.md`; no `GEMINI.md` wrapper remains.
  No nested guide or client setting was added. Loader qualifications live in the dated research.
- Replaced README's duplicated contributor-command roster with links to the shared guide and
  justfile. Preserved product workflows and made the TUI section a peer of Contributing.
- Repointed the three active routines and the store fuzz-test comment to `AGENTS.md`. Routines
  follow its scoped reading routes, require a clean starting checkout, and stop on unexpected
  changes without reverting user-owned work. No routine was executed against this dirty checkout.
- Corrected stale authoring metadata and nearby registry/layout comments at their owning Go
  sources. Revision 1.88 is additive guidance only; JSON shapes, inputs, and mutation behavior
  are unchanged. Updated fixtures and the generated comment map, with a live CLI regression.
- Full race-enabled `just test`, `just lint`, `just build`, and `just docs-check` passed.
  Reverting the dependency convention made the new regression fail; restoring it passed.
  A fixture comparison confirmed changes were limited to the revision and corrected guidance.
- Checked local guide links/anchors, absence of tool-specific wrappers, and planning links.
  Static routing checks do not prove model-client loading or instruction adherence; no authenticated
  client session was run. Larger architecture, package-contract, and drift checks remain sequenced
  in their existing tasks, not silently included here.

## Adapter-neutral closeout baseline (2026-10-05)

The completed port migration added [a planning-data checklist](../../docs/ARCHITECTURE.md#planning-data-change-checklist)
and linked it from README, CLAUDE, the architecture reference, and risk-bearing package comments.
Reuse that route in the tool-neutral agent entry point instead of copying the full contract into
another guide. This does not provide the source-of-truth matrix or complete this task. The current
Thread command guidance was already corrected; treat the obsolete-command example above as
historical motivation, not a current defect claim.

The checklist is integrated into the architecture reference, not a separate top-level document.
Preserve that single owner when adding agent routing; any later extraction follows the planned
documentation ownership and architecture restructuring work.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- [External implementation review brief](../audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md)

## External review closeout (2026-10-12)

The [independent review](../audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md)
confirmed no lost durable safeguards, working reading routes, and only the declared revision/guidance
delta. Both low findings are fixed in this slice:

- L1: the focused regression requires the present-tense guarded route and the complete generic-write
  prohibition, not just names. The reported future-route mutation and a separate permission-only
  mutation each compile and fail the intended test; restoring the guidance passes. This remains a
  deliberately phrase-based policy regression, not a general natural-language semantic validator.
- L2: the ownership matrix and conflict paragraph now distinguish generated registries/shapes
  from maintained prose, and declaration tests from human compatibility judgment.

The transferred audit matched its retained independent sandbox byte-for-byte before owner
dispositions. The review-helper origin/push concern is assigned to
[its own task](6gk06e2g21hv-disable-pushes-from-isolated-review-workspaces.md) and
[review-workspace hardening Thread](../threads/6gk06gktr1q5-harden-isolated-review-workspace-safety-and-delivery-evidence.md),
not made a blocker for documentation delivery. Existing routine-policy contradictions remain in
the final reconciliation task. The ownership task is complete; the docs Thread remains in progress.
