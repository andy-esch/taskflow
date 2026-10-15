---
schema: 1
id: 6gkmexz4vh23
status: completed
epic: 21-code-quality-architecture-hardening
description: Prevent retry-time alias drift and missing-ID fallback from silently redirecting task mutations to another record.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, core, store, concurrency]
created: "2026-10-14"
depends_on: [6gkg89ay3b5b]
updated_at: "2026-10-15"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
started_at: "2026-10-14"
completed_at: "2026-10-15"
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

- [x] All retried task body/criterion/field operations retain the initially established source identity through a rename, old-alias reuse and a new unique fuzzy match; the intended task changes and every sibling remains byte-identical.
- [x] A disappeared selected ID never falls back to an ID-shaped or case-folded human slug; it returns an appropriate missing/conflict error with no sibling write, including tracked handoffs and retry exhaustion.
- [x] Missing identity evidence and canonical-ID collisions fail closed with useful existing sentinels; source declarations, display slugs and locations are not action authority.
- [x] Ordinary concurrent body edits still preserve both writers; no-op, preview, callback errors, fixed timestamp, authorization and retry bounds remain covered in portable and real-store tests.
- [x] Lifecycle and adjacent mutation families are inventoried against the same rule; confirmed out-of-scope gaps have explicit scoped tracking rather than unproven safety claims.
- [x] Owning port/retry comments and planning-data guidance describe the actual target-selection guarantee; race tests, lint, build, docs and planning validation pass.
- [x] Task field repair still replaces malformed typed values through Service
  and CLI, including previews; unrelated corruption and invalid body writes
  remain validation errors without effects.

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

## Implementation approach (2026-10-14)

Keep user-facing reference resolution unchanged. Core resolves one task source and
passes an explicit selected-identity value to the three existing task-write methods;
the filesystem adapter uses exact-ID lookup before callbacks, no-op/preview returns
and CAS. This is a narrow internal port contract, not a new capability or generic
mutation framework. Replacement shares the body port but remains non-retried;
interactive editing stays unchanged.

Review correction: initial selection uses parse-free `ResolveTaskSource` on the
existing task port, not `ReadTask`. A malformed typed value must not hide identity
when `task set` can repair it. The write adapter validates the updated document;
no local path capability or public resolution-policy change is needed.

Lifecycle already resolved inside its guarded graph, but its retry repeated the
original caller reference. The correction retains the first selected plan's identity
across retries while recomputing eligibility/impacts on each fresh authoritative
snapshot. Selection stays inside the repository guard.

Branch: `fix/identity-bound-task-mutations`. The separate tracked-reason policy remains
unselected and out of scope; no existing Thread owns this correctness slice.

## Implementation and owner evidence

- The three existing task-write methods now take `core.TaskMutationTarget`, not a
  string user query. Service selects `RecordSource.ID`; FS uses exact ownership
  lookup before callbacks, no-op/preview returns and content CAS. Primary CLI
  references keep their original matching tiers, including ID-shaped human labels.
- Single-task selection conflicts and writes share one bounded retry budget. Successful
  selection is retained even if a later attempt returns an empty error result;
  timestamp capture remains once per operation. Replacement is still non-retried.
  Tracked handoffs retain their unretried endpoint preflight; their body retries
  never re-resolve either endpoint.
- Lifecycle move/defer had the same defect in its guarded planner retry. Four owner
  regressions failed with a successful write to the sibling before the correction.
  First selection stays inside the guard; retries retain the selected semantic plan
  while the adapter recomputes validation, materialization and impacts from fresh data.
- Portable tests cover pathless/opaque source IDs, declaration/slug disagreement,
  empty evidence, sentinel propagation, one read after selection, fixed clocks,
  preview/replace retry refusal, shared-budget exhaustion and lifecycle contention
  returning no plan between two planned attempts.
- Real-store matrices cover checkbox/state/add/remove/replace/tracked/append/field
  edits across cooperating rename, alias reuse, deletion and case-folded ID-shaped
  impostors. Selection-to-first-write drift covers replacement and previews too.
  Strict port tests refuse duplicate/missing/empty/label/case-invalid targets before
  callbacks and effects, including no-op requests. Existing CAS, concurrent append,
  authorization and 50 self-alias regressions remain covered.

The main rename/deletion/lifecycle tests were red before implementation and pass
afterward. Owner survey probes that assert an observed defect were removed from the
repository test suite and retained in local scratch; they are not passing safety tests.

## Neighboring boundary survey and follow-ups

| Boundary | Observed evidence | Owner |
| --- | --- | --- |
| Task body/field operations and lifecycle | Wrong-record retries reproduced; fixed in this task | Current implementation and permanent regressions |
| Audit append and research fields | Rename/CAS retry changed the sibling with nil error | [Non-task retry pinning](6gkt3fphx1ts-pin-non-task-mutation-retries-to-selected-source-identities.md); also inventories epic, finding creation/edit and bulk header repair |
| Dependencies and Thread lifecycle | Fresh guarded retry added an edge to the sibling or started the wrong Thread after rename; both returned nil | [Graph/Thread endpoint binding](6gkt3fpsfdd7-keep-graph-and-thread-retry-endpoints-fixed-during-replanning.md); retain IDs, not stale write sets |
| TUI field action selected before Service entry | Actual queued setFieldCmd edited an ID-shaped impostor after deletion and returned editedMsg | [Selected TUI task actions](6gkt3fq0pv26-preserve-selected-tui-task-identities-through-service-actions.md), sequenced after this port contract and added to the existing navigation Thread |
| Whole-file task edit and non-retried replacement | Interactive edit retains its single-attempt CAS; replacement now shares strict selection but still never auto-retries | No expansion of editor/retry policy here |
| Repair and bulk graph operations | Their explicit-ID/source-occurrence and durability contracts differ from ref-based retries | Inventory included in the graph/Thread follow-up; no unprobed clean verdict |

Audit/research/dependency/Thread probes ran against real FS seams under `-race` and
changed a pre-existing sibling after a stale snapshot was correctly rejected. The TUI
probe exercised the real command closure before watcher refresh. These bound the
claims to deterministic interleavings, not concurrent-frequency measurements.

The two non-task/graph follow-ups are independently actionable; no artificial
dependency on this task or broad integrity research was added. The TUI successor
has a real dependency on this task's selected-target port and lifecycle rule.
The ordinary CLI initial-reference contract must not be tightened globally to
paper over the selected-interface boundary.

## Validation and independent review handoff

Full just test (race), just lint (0 issues), just build, just docs-check, fresh focused
core/store/CLI race runs, architecture navigation, planning lint --links and diff
hygiene pass. The unregistered space at
`/private/tmp/taskflow-task-target-space.U5Oj8W` exercised case-folded prefix fields,
append, checkbox/tracked edits with prose, rename then stable-ID updates, start and
defer. Its lint --links passes; self-tracking by ID/slug refuses with exit 11 in write
and preview, and a field preview plus both refusals retain the source SHA-256.

[Claude](../audits/6gktgjd3fm74-2026-10-14-strict-task-mutation-targets-implementation-claude.md)
leads application/lifecycle binding; [Antigravity](../audits/6gktgjd9ebmn-2026-10-14-strict-task-mutation-targets-implementation-antigravity.md)
leads exact persistence and realistic producer/authorization probes. Both require
compiler-valid mutations killed by the intended regression and a systemic second pass.

Brief provenance: the preceding review found a neighboring raw-alias retry bug, and
the owner then demonstrated that merely passing an ID string still permits fallback.
These briefs challenge both selection and exact action lookup separately, distinguish
earlier TUI selection as tracked scope, and use the newly merged helper's split
verification/transfer-receipt contract instead of requiring self-attested delivery.

All seven criteria are met and this task is completed. Both delivered reviews have
been processed and closed; the bounded neighboring gaps remain tracked separately.

## Claude review disposition (2026-10-15)

Confirmed the helper's delivery receipt before triage. M1 is valid: the initial
typed read prevented field repair before raw-node updates could replace a bad value.
The parse-free source method preserves base behavior while retaining strict action
identity, authorization, CAS and updated-document validation. Tracked handoff endpoint
preflight remains typed and unretried.

The new [CLI regressions](../../internal/cli/task_set_repair_test.go) failed before
the fix for tier/tags/priority writes/previews and existing-corruption attribution.
They now pass, including no-effect refusal of malformed body writes. The portable
selection fake now deliberately fails typed reads. Drift tests exercise the new
source resolver as well as tracked handoff's original typed preflight.

Claude's two non-finding coverage notes are also addressed: real-store tests cover
repository-scoped selection conflicts sharing the CAS retry budget and an empty
intervening lifecycle result retaining identity. Both pass ten race-enabled runs;
no separate task is needed for these bounded tests. Existing neighboring follow-ups
remain independently scoped.

Full race suite, lint, build and docs-check pass. Fresh-binary checks in unregistered
`/private/tmp/taskflow-claude-repair.B6yFGR` confirm repair preview succeeds without
changing SHA-256, an unrelated update exits 11 with the original attribution and
unchanged bytes, and stable-ID repair makes the task readable. M1 is fixed and Claude's
audit closed; this is owner verification, not an independent re-review of the correction.

## Antigravity disposition and completion (2026-10-15)

Confirmed delivery before triage. No new findings: independent exact-lookup mutation,
missing-ID/impostor, denied callback and fresh-blocker probes support this slice.
Its clean verdict covers the original captured snapshot, not Claude M1's later fix;
owner qualification in that audit prevents broader safety claims from overriding the
repair regression or the remaining scoped gaps. Workstation-only evidence links are
now repository-relative, with obsolete destinations corrected.

Completion makes the selected-TUI-action successor eligible in
[refine Thread/TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md).
This task remains an external prerequisite, not an artificial Thread member. Its
one-hop exported neighborhood shows that handoff for the PR. Non-task and graph/Thread
retry work remains independently actionable; the tracked-reason policy still needs
its separate human decision. No broader framework or CLI/wire policy adopted here.
