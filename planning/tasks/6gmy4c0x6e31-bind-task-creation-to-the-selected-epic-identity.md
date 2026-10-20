---
schema: 1
id: 6gmy4c0x6e31
status: completed
epic: 21-code-quality-architecture-hardening
description: Add an adapter-neutral selected-epic creation path so a staged task form cannot silently resolve its displayed epic to a different record.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [tui, authoring, architecture, identity]
created: "2026-10-18"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-20"
started_at: "2026-10-19"
completed_at: "2026-10-20"
---
# Bind task creation to the selected epic identity

## Objective

Make the epic chosen in a task form a strict selected identity, not a human reference
resolved again at submission. This is a reusable core/secondary-adapter capability,
not a TUI form implementation.

## Evidence and scope

NewTask currently reads epics and matches its Epic reference before calling
CreateTask. Existing EpicMutationTarget requires the complete adapter-established
source ID, not a sequence prefix, declaration or display slug. Ordinary CreateTask
serializes collision checks with its write but does not carry a selected parent
expectation. Passing a picker ID into the reference API is therefore not the
already-selected contract used by existing TUI field actions.

Add the smallest explicit selected-epic creation entry point and supporting port
contract needed to refuse missing, duplicate or replaced selection. Retain NewTask's
ordinary CLI reference behavior and shared defaults/validation/scaffolds. Define
where parent evidence is revalidated and what the adapter actually guarantees;
do not advertise database-style foreign keys or atomic parent survival unless
the committing capability implements that guarantee.

## Acceptance criteria

- [x] Core offers a selected-epic task-create contract using adapter-established identity, with normal human-reference creation remaining compatible.
- [x] Source selection cannot fall back to an epic prefix, declaration or impostor label; missing/duplicate/changed selected parents fail closed before a create is reported successful.
- [x] The owning persistence seam revalidates selected-parent expectations within its create/guard contract; documented limits distinguish selection safety from cross-document referential-integrity guarantees.
- [x] Source-set agreement and explicit mutation policy remain enforced, including read-only and dry-run refusal; new capabilities do not impose local paths or unrelated writes on other adapters.
- [x] Preserve core defaults, per-kind validation, no-clobber creation and TaskCreationReceipt committed/finalization evidence; never remint or replay after a committed result.
- [x] Portable/pathless and real guarded-adapter regressions cover selected-parent drift, lookalike labels, contention and finalization failure; owning comments and focused architecture guidance match behavior.

## Out of scope

TUI forms, a universal CreateEntity interface, changing epic identity rules, automatic
epic creation, repairing existing orphan relationships, or a new persisted schema.

## Related

- [Authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Confirmed interaction and adapter-safe gate](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md#adapter-safe-implementation-gate).
- [Task use cases](../../internal/core/service_task.go), [ports](../../internal/core/store.go)
  and [guarded creation](../../internal/store/create.go).
- [TUI consumer](6g4ym5y2dkp8-create-tasks-from-the-tui-and-epic-context.md).

## Implementation checkpoint (2026-10-19)

- Added optional SelectedEpicTaskCreationStore and NewTaskForSelectedEpic; ordinary
  CLI creation keeps reference matching and create-and-start behavior. Shared
  preparation owns defaults/validation/scaffolding, with no parallel form policy.
- The local adapter validates exact readable parent ownership inside ordinary
  guarded creation, reauthorizes after waiting, and preserves create/finalization
  receipts. Selected create-and-start is explicitly unsupported. Preview reserves
  nothing; later parent removal, identical-ID recreation and raw-writer races are
  not claimed to be prevented.
- Portable tests cover pathless-only composition, capability mismatch/absence,
  defaults, input refusal and no remint/replay. Real-adapter tests cover full stems,
  impostors/unreadable parents, queued identity/policy drift, collision and release
  failures. Existing mutation-policy inventory includes the new entry point.
- Validation: fresh build, full race suite, lint, planning/link lint and diff checks
  pass. No CLI metadata or machine-output contract changed.
- Independent reviews are settled below. The TUI consumer is not implemented here.

Completed independent reviews:

- [Codex: portable API and identity/composition](../audits/6gndg1vdbbxd-2026-10-19-selected-epic-task-creation-implementation-codex.md).
- [Antigravity: guarded timing, policy and recovery](../audits/6gndg1vjjv1p-2026-10-19-selected-epic-task-creation-implementation-antigravity.md).

## Review disposition (2026-10-20)

Both isolated-review delivery receipts were confirmed before triage. Neither
review demonstrated a production defect; each exposed a test-oracle weakness:

- Antigravity L1: evaluate the pre-lock permit before signaling the waiting test,
  then require three authorization checks after the worker returns. This prevents
  a pre-lock refusal from masquerading as post-wait reauthorization. Both queued
  tests pass with `-race -count=20`.
- Codex L1: retain parity coverage, but independently assert literal default and
  non-default task metadata, body overrides and exact named-template output through
  both public entry points, in write and preview modes.

Compiler-valid Go overlays demonstrate that removing in-guard reauthorization
fails the intended waiting test and replacing the shared priority input with a
constant fails independent metadata assertions for both entry points. Overlays
never modified production source. Both findings are fixed; no follow-up design
decision or production-policy change was needed.
