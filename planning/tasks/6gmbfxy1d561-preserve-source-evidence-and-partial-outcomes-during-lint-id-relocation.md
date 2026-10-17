---
schema: 1
id: 6gmbfxy1d561
status: completed
epic: 21-code-quality-architecture-hardening
description: Make lint ID spelling relocation no-clobber and mode-preserving, recheck source bytes, and report destination-written/source-retained failures.
effort: 1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, repair, persistence, concurrency]
created: "2026-10-16"
started_at: "2026-10-16"
updated_at: "2026-10-17"
completed_at: "2026-10-17"
---
# Preserve source evidence and partial outcomes during lint ID relocation

## Objective

Harden the misspelled-ID relocation in `internal/store/fix.go` without changing
which spelling repairs are authorized. The October 16 concurrency audit's M1 and
M2 share this path; the earlier completed lint-fix locking task addressed cooperating
writers, not raw-edit CAS, permission preservation or per-file partial outcomes.

## Scope

- Preserve the source mode using exclusive destination creation; refuse an occupied
  destination, including one appearing after the availability check.
- Recheck original source evidence before destination creation and again before
  source removal. Never remove a concurrent raw edit.
- Report the destination-written/source-retained stage alongside errors and earlier
  batch progress. Choose an adapter-neutral maintenance receipt consistent with the
  existing maintenance port; assess wire compatibility before changing public output.
- Keep both copies for explicit inspection on a post-create conflict. Do not promise
  crash atomicity, add signal handling or automatically delete ambiguous records.
- Correct stale relocation/rename comments in the owning fixer and persistence guide.

## Acceptance criteria

- [x] Source permissions survive repair, including 0600 and 0664 under explicit
  umask 022; destination collision tests prove no clobber.
- [x] Deterministic pre-create and pre-remove source drift tests preserve concurrent bytes.
- [x] Failed unlink or post-create conflict reports both paths and the durable stage without claiming completion or hiding prior repairs.
- [x] Preview makes no changes; permission, authorization, preservation and normal quality gates pass.
- [x] Independent hostile review verifies real-store preservation and partial
  receipt propagation across core, CLI and schema; meaningful findings are fixed
  or explicitly tracked.

## Evidence and sequencing

[Concurrency audit M1/M2](../audits/6gm8g0f66k80-2026-10-16-concurrency-and-atomicity.md).
The source confirms both missing safeguards. A process kill can still leave two
copies; reporting and non-destructive recovery, not eliminating that window, are
the contract. The audit's copied-valid-ID probe is not an actual spelling-repair
interruption: the original misspelled filename has no canonical source ID, so
that probe does not establish a new duplicate-canonical-ID failure. The original
was already unhealthy. The missing CAS and partial receipt remain valid defects.
Do not reopen the historical status-folder move tasks or expand
Finish selected identity safety: this is an independent maintenance-write task.
The broader entity-integrity foundations task may reuse the resulting evidence,
but is not a prerequisite.

This task and the two adjacent follow-ups now belong to the bounded
[maintenance-persistence Thread](../threads/6gmdt09gygfw-harden-local-maintenance-and-relocation-persistence.md).
There are no artificial dependency edges; the confirmed rename destination-loss
case is the recommended next slice, not a prerequisite for this fix.

## Implementation checkpoint (2026-10-16)

Started from merged main `e454d00b` on `fix/guard-lint-id-relocation`. Four retained
red-first tests reproduced mode widening, destination clobber, pre-create source
edit loss and pre-remove source edit loss. They now pass.

The adapter rechecks source bytes plus existing identity/reference authorization,
creates exclusively using the source mode, then verifies both copies before source
removal. Ordinary spelling policy is unchanged; no cascade, ambiguous-copy deletion
or crash-atomicity promise was added. The final check-to-effect window remains.

`Fixer` keeps its existing completed/proposed row signature. A core-owned typed
failure carries progress inside an interrupted relocation, which `RepairPlanning`
copies into its result without entering later phases. Partial moves are not fixed
rows. Human output and JSON stdout/error recovery expose exact local paths and
historical action flags. Returned stdout writer errors retain the receipt for
stderr reporting; process termination (including SIGPIPE on a broken stdout pipe)
can prevent that report. No signal-handling policy was added.

Wire revision **1.89 ADDITIVE** adds optional `relocation` and
`error.planning_repair` evidence, preserving fixed/skipped semantics and exit codes.
The emitted filesystem retry hint is false after destination commit. Source
comments and focused guides document the policy without adding operation evidence
to domain records. Forty existing JSON goldens change only `schema_version`; the
schema golden adds the new optional branches. Generated files were regenerated,
not hand-edited.

## Adjacent work, not implemented here

- [In-place lint CAS and guard finalization](6gmdsxccqqnm-guard-in-place-lint-repairs-and-retain-guard-finalization-failures.md): source-inspection evidence; requires its own deterministic red-first probe.
- [Task rename destination check](6gmdsxcevgkr-verify-the-task-rename-destination-before-removing-its-source.md): confirmed with a compiler-valid temporary real-store probe. It was removed rather than retaining a passing assertion of unsafe behavior.

Review dispositions below record the source findings' resolution. This task is
complete with the PR handoff; the maintenance-persistence Thread stays open for
its two adjacent follow-ups. The completed selected-identity Thread is not reopened.

## Validation and external review checkpoint

- `just build`, full `just test` (race), `just lint` (0 issues), focused
  core/store/wire/CLI repair races (`-count=3`), and planning `lint --links` pass.
- Generated schema comments and CLI goldens are current. `just docs-check` reports
  only the intentional uncommitted lint-help addition; independent fresh docgen
  output matches all of `docs/cli`.
- Owner mutation: replacing exclusive creation with the overwriting primitive
  compiles and fails `TestFixIDRelocationRefusesRawDestinationCollision`, which
  injects the winner after fresh authorization. Exclusive creation was restored;
  the focused race suite and full suite pass again. The earlier authorization seam
  alone would mask this regression and is deliberately tested separately.

Both independent isolated-clone reports have landed. Claude's meaningful findings
are fixed; Antigravity's useful overlap is resolved, but its clean verdict is not
accepted because of material baseline mismatches (see dispositions below):

- [Claude — real-store ordering, mode, source/destination CAS](../audits/6gmdtwr8ycdm-2026-10-16-lint-id-relocation-safety-implementation-claude.md).
- [Antigravity — core/CLI/wire recovery and schema semantics](../audits/6gmdtwrfq2yq-2026-10-16-lint-id-relocation-safety-implementation-antigravity.md).

## Claude review disposition (2026-10-16)

Claude's isolated report was accepted only after the source helper confirmed its
guarded transfer receipt (`delivery=confirmed`). All three findings were meaningful:

- **M1 fixed:** partial-failure messages and help explicitly say retries cannot
  reconcile two copies. Plain lint uses the scan's directory-entry snapshot to
  replace rename advice when the canonical sibling is occupied; pure filename
  advice is conditional on absence. Unchanged and destination-edited partial
  fixtures exercise plain lint and refused retry across all four ID-led kinds,
  proving both copies remain byte-for-byte unchanged. Recovery remains deliberate
  comparison/merge and removal or archival, never automatic copy selection.
- **M2 fixed:** source `0664` under explicit umask `022` runs in an isolated test
  subprocess; unlink failure is injected at the actual remove call, after both
  content checks; CLI fixtures use transient `EBUSY` and joined stdout `EAGAIN`.
  The exact-mode-to-ordinary-create mutation fails the umask test, removal of the
  unlink receipt fails the real-Service prefix test, and disabling the committed
  retry override fails both CLI recovery tests. Each mutation compiled, failed
  for its intended invariant, and was restored before green validation.
- **L1 fixed by narrowing the claim:** returned writer errors preserve the receipt;
  signal/process termination can prevent reporting. Go's local `os/signal` docs
  confirm default SIGPIPE behavior on stdout/stderr. No global signal policy changed.

The report's unresolved exclusive-create cleanup/raw-replacement hypothesis is
recorded in [the existing integrity-design task](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md).
It still needs a deterministic I/O-fault reproduction; it is not claimed as a
confirmed defect or added to this bounded Thread's finish line.

Post-review `just build`, full `just test` (race), `just lint` (0 issues), focused
repair races (`-count=3`), schema-comment generation, planning `lint --links` and
`audit lint` pass. CLI reference is regenerated; `just docs-check` still reports
only the intentional unstaged lint-help addition. At that checkpoint Task/Thread
and AC5 remained open for Antigravity's report; the source findings were tracked.

## Antigravity triage and review closeout (2026-10-17)

The helper confirmed Antigravity's guarded audit transfer before owner edits. Its
one finding, L1, is valid and already fixed by Claude M2's `EBUSY`/`EAGAIN` fixtures
and retry-override mutation check. No duplicate implementation was needed.

Its broader report has material mismatches even against its own captured baseline:
invented core types/helpers, hard-link creation instead of `O_EXCL`, a nonexistent
lock file, six claimed ID-led kinds instead of four, and nonexistent wire fields.
The audit preserves the report with a prominent warning and exact owner errata.
Successful sandbox delivery is not proof that its claimed execution traces are
accurate; this is a processed review with degraded evidence, not a clean verdict.

The owner independently repeated its useful optional-field mutation: deleting the
actual wire relocation mapping compiles and passes shape-only schema validation,
but fails retained wire/CLI semantic receipt tests. Restored mapping passes focused
race runs. Added a retained four-field schema regression covering both values of
each action flag, required-field omission and wrong-type rejection. The wire
changelog now uses the same deliberate-reconciliation guidance as the fixed help.

AC5 is satisfied by Claude's independent hostile real-store/core/CLI/schema probes,
resolved findings and retained owner validation—not by counting Antigravity's
inaccurate inventory as a second independent pass. Original concurrency M1/M2 are
fixed for their validated defects; the correction to M2's claimed duplicate-ID
consequence remains explicit. No automatic ambiguous-copy cleanup was added.

Final build, full race suite, lint (0 issues), focused repair races (`-count=3`),
schema-comment regeneration and planning lint pass. Generated CLI reference is
current; `docs-check` reports only the intentional unstaged lint-help addition.
No design decision or further implementation blocker remains for this task's PR.

## PR handoff and cleanup (2026-10-17)

All five ACs are met and the task is completed through the lifecycle command. The
Thread is healthy with two independently actionable remaining members: task rename
destination verification is the recommended next slice, then in-place lint CAS and
guard finalization. No artificial dependency edges were added for delivery order.

The processed Claude and Antigravity independent clones were removed after receipt
confirmation and owner triage; source audits retain the baseline, delivery evidence
and dispositions. Removed two generated-doc comparison directories as well
(approximately 375 MiB reclaimed). There were no temporary space registrations or
extra Git worktrees to remove. No temporary mutation probes remain in the source.
Shared Go/linter caches and the freshly built binary are retained for normal use.

`just docs-check` now passes with the generated help included in the implementation
commit. The PR includes the exact generated one-hop neighborhood; it contains only
this task because the two follow-ups are independent members, not prerequisites or
dependents.
