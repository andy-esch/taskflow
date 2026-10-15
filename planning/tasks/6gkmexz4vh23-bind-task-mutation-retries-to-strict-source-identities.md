---
schema: 1
id: 6gkmexz4vh23
status: next-up
epic: 21-code-quality-architecture-hardening
description: Prevent retry-time alias drift and missing-ID fallback from silently redirecting task mutations to another record.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, core, store, concurrency]
created: "2026-10-14"
depends_on: [6gkg89ay3b5b]
updated_at: "2026-10-14"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
---
# Bind task mutation retries to strict source identities

## Objective

Prevent a task mutation from silently switching records when a caller alias drifts or
the initially resolved record disappears. This is a bounded production correctness
follow-up, not a prerequisite for a universal entity or transaction framework.

## Evidence

Claude [M1](../audits/6gkm2xz5s9qx-2026-10-14-task-correctness-detour-implementation-claude.md)
shows that core retries `TransformTaskBody`, `EditBody` and `SetFields` using the raw
caller query. A rename of `alpha` to `omega` leaves a pre-existing `alpha-two` as the
unique prefix match. The retry mutates that sibling and returns success.

The owner independently reproduced this in real FS code using the pre-lock hooks:
checkbox, criterion state, criterion editing, append and field update all returned nil,
left the renamed source unchanged and changed the sibling. The tracked handoff retained
the original source under this rename interleaving.

Owner deletion probe exposes a second layer: even the resolved `Source.ID` is still
passed to a generic fuzzy resolver. With a pre-existing sibling whose slug equals the
source ID, deleting the source before CAS makes the tracked retry mutate that sibling
and return nil. Resolving once in core is necessary for alias drift, but insufficient
without strict identity-only action lookup. Test fixtures intentionally carry no declared
frontmatter ID; the filename-established identity remains authoritative.

Owner reproduction files are retained locally (not durable CI artifacts) at
`/private/tmp/taskflow-correctness-review.qhNUTt/alias_retry_probe_test.go` and
`/private/tmp/taskflow-correctness-review.qhNUTt/deleted_id_probe_test.go`. Both compiled
and failed the intended stable-target assertions under `go test -race`; the first had
five wrong-target failures and one passing tracked control. This is deterministic seam
evidence, not a measured two-process likelihood estimate.

## Scope

- Resolve the initial caller reference once through the selected task store, require
  adapter-established `RecordSource.ID`, and retain it for the whole mutation/retry.
- Ensure selected-identity actions never fall back to exact/prefix/substring human
  labels when that ID vanishes. Do not use declarations or opaque locations as identity.
- Cover checkbox/state/add/remove/replace criterion writes, tracked handoffs, append
  and field updates. Keep callbacks pure, one captured timestamp, bounded retry, CAS,
  no-op/preview behavior, explicit authorization and sentinel/receipt behavior.
- Inspect lifecycle's separate graph planner/retry path and adjacent entity retries.
  Fix task paths here; document other confirmed occurrences in existing relevant tasks
  or separately scoped follow-ups rather than assuming this is task-only.
- Keep interactive whole-file editing and non-retried replace semantics unchanged
  unless inspection exposes the same wrong-target defect requiring explicit scoping.
- Preserve user-facing fuzzy reference ergonomics at initial selection. If a narrow
  port change is needed, distinguish user references from already-selected identities;
  do not globally reinterpret every ID-shaped slug without a compatibility assessment.

## Acceptance criteria

- [ ] All retried task body/criterion/field operations retain the initially established source identity through a rename, old-alias reuse and a new unique fuzzy match; the intended task changes and every sibling remains byte-identical.
- [ ] A disappeared selected ID never falls back to an ID-shaped or case-folded human slug; it returns an appropriate missing/conflict error with no sibling write, including tracked handoffs and retry exhaustion.
- [ ] Missing identity evidence and canonical-ID collisions fail closed with useful existing sentinels; source declarations, display slugs and locations are not action authority.
- [ ] Ordinary concurrent body edits still preserve both writers; no-op, preview, callback errors, fixed timestamp, authorization and retry bounds remain covered in portable and real-store tests.
- [ ] Lifecycle and adjacent mutation families are inventoried against the same rule; confirmed out-of-scope gaps have explicit scoped tracking rather than unproven safety claims.
- [ ] Owning port/retry comments and planning-data guidance describe the actual target-selection guarantee; race tests, lint, build, docs and planning validation pass.

## Design checkpoint

The invariant is settled: retries must not select a different record. Choose the narrowest
maintainable application/adapter boundary that carries selected identity explicitly.
Stop for human input if implementation changes promised CLI resolution behavior or
requires a broader public contract/mutation-framework expansion. A new framework is
not required merely to fix these observed cases.

## Sequencing and overlap

Follow [the correctness detour](6gkg89ay3b5b-fix-task-identity-guards-and-deterministic-field-updates.md):
reuse its authoritative source-ID guards and tracked rename regression. No dependency
on the broad [shared-integrity design](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md)
is needed; feed concrete retry evidence into that roadmap instead.

The completed [criterion CAS correction](6g6yhyrhp643-close-the-lost-update-window-in-task-acceptance-criteria-writes.md)
fixed content lost updates, not alias/action drift. The completed
[CAS error-classification fix](6g1dpb5gwtt9-verifyunchanged-collapses-every-resolve-error-into-a-futile-retry-conflict.md)
preserved ambiguity in the store's exact-ID guard, but core can still choose another
record before entering that guard. Do not reopen those historical tasks.

Keep this epic-21 correctness work standalone: no existing open Thread owns this scope,
and the documentation and isolated-review Threads are not artificial prerequisites.

## Out of scope

Cross-record transactions, automatic identity repair, ID reuse/incarnation policy,
distributed retry/idempotency frameworks, schema/wire changes and a universal mutation API.
