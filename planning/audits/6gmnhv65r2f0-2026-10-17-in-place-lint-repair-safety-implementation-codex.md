---
schema: 1
id: 6gmnhv65r2f0
bucket: closed
area: in-place-lint-repair-safety-implementation-codex
date: "2026-10-17"
---
# Audit: In-place lint repair safety implementation — codex — 2026-10-17

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief's bounded evidence floor rather than exhaustively mutating every new test.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Your handoff prompt prints exact `SOURCE_ROOT` and `AUDIT_REL`
values. Substitute them below; never derive them from your current directory, which may be a
different checkout. Then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="<SOURCE_ROOT from your handoff prompt>"
AUDIT_REL="<AUDIT_REL from your handoff prompt>"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)" && cd "${SANDBOX:?create failed: stop and report the blocker}"
```

Every later block begins from `${SANDBOX:?…}`, so it cannot run in the shared checkout after a
failed `create`. Run the helper from `$SOURCE_ROOT`, not from the sandbox: you may mutate the
sandbox's copy while probing, and only the read-only source copy can be trusted to verify your work.

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. The sandbox's push URLs are disabled; never re-enable them or
push to the source path. If creation fails, report the blocker; never fall back to the shared
checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then run
the helper's fail-closed verification:

```sh
cd "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX" &&
  git diff -- "$AUDIT_REL"
```

Copy that `verify` output into the report's isolation attestation: workspace path, resolved Git
directory, baseline commit, captured source blob/fingerprint, deliverable, push guard,
`verify=passed`, and `transfer=pending`. The report is final from this point, so `transfer=pending`
is the truthful value; never edit it to claim delivery. Re-run `verify` (its fields do not depend on
the report text), then transfer:

```sh
"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent or nested `.git`, a
removed or weakened push guard, source-deliverable drift, and empty reports; it copies back only
the assigned audit through a same-directory atomic rename. A successful transfer records a
helper-owned receipt and prints it (`transfer=succeeded` with the transferred blob). A refusal exits
2 and records nothing. Exit 3 means the report was already delivered when a later step failed: run
`"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"` and put its
output, or its refusal, in your handoff. Put the printed receipt in your handoff message. Do not
copy anything else manually. Leave the workspace in place and report its path until the
implementation owner confirms delivery with that same `receipt` command. On refusal, preserve it and
report the conflict rather than resolving it in the shared checkout.

A report without the verification attestation, or a handoff without the transfer receipt, is
incomplete even if its technical findings are otherwise sound.

## Review brief

Challenge the bounded ordinary lint-repair safety implementation before closing the last
member of the maintenance Thread. This is a falsification exercise, not approval paperwork.
Both audits contain the same shared contract; use the lens for your assigned reviewer:

- Codex: persistence, original-byte CAS, post-wait authorization and checked guard finalization.
  Cross-check how your actual store outcomes reach core/CLI recovery.
- Antigravity: core/CLI durable-prefix classification, writer failures and machine-contract
  truthfulness. Cross-check one real store fault rather than trusting synthetic receipts.

Work independently. Do not consult the sibling report until your own report is final.
No finding quota; one demonstrated root cause is more valuable than speculative breadth.

## Review target

Source checkout: /Users/andyeschbacher/git/andy-esch/taskflow.
Branch: fix/guard-in-place-lint-repairs. Base: merged main 43af9bb1 (PR #313).
Implementation is uncommitted: capture all staged, unstaged and untracked files, not HEAD alone.
Read the captured baseline and task checkpoint before interpreting any owner evidence.

Implementing task:
planning/tasks/6gmdsxccqqnm-guard-in-place-lint-repairs-and-retain-guard-finalization-failures.md.
Thread:
planning/threads/6gmdt09gygfw-harden-local-maintenance-and-relocation-persistence.md.

Actual entry points to verify, not declarations to copy from this prompt:
- internal/store/fix.go: FS.FixFrontmatter, verifyFixInPlacePath and its pre-write hook;
  checkedWriteLock in lock.go, existing atomic helper and relocation helper.
- internal/store/fix_inplace_test.go: raw-drift matrix across five entity kinds,
  actual process-guard wait/authorization test, release/prefix and preview/mode tests.
- internal/core/planning_maintenance.go and frontmatter_repair.go:
  RepairPlanning, PlanningRepairResult.Committed, PlanningRepairFailure and distinct
  FrontmatterRepairFailure relocation evidence.
- internal/core/planning_maintenance_test.go and core Fixer port comments in store.go.
- internal/cli/lint.go, exit.go and planning_repair_failure.go; tests in
  internal/cli/planning_repair_failure_test.go.
- internal/wire/planning_repair_test.go: populated ordinary-prefix schema validation;
  existing wire mapper, ErrorEnvelope and filesystem retry meanings.
- docs/architecture/persistence.md, generated docs/cli/tskflwctl_lint.md and ADR-0008.

Broader shared create cleanup, relocation symlink/preflight semantics, arbitrary legacy body
writers and a universal mutation framework are outside this task. Related investigation is
explicitly scoped in task 6g7wsr8yvt1a; ordinary audit body work also has tasks
6g77rn6hvmh8 / 6g77rn6n6b86. Scope is not immunity: demonstrate an in-scope consequence
if an adjacent boundary undermines the claimed result.

## Intended contract to challenge

- The filesystem adapter reauthorizes after acquiring the repair guard and before effects.
  Checked release keeps primary and finalization errors together, even with zero writes.
- An in-place replacement needs current regular-file and original-byte evidence. It must
  not overwrite a newer edit, resurrect a deleted source or replace a byte-identical symlink.
  Malformed YAML remains repairable without typed decoding.
- Completed rows survive later errors. An interrupted relocation remains separate, never
  mislabeled as a completed repair. Unknown content, comments, modes and graph-field refusals
  keep their existing behavior.
- Core's Committed flag describes historical durable progress: completed repairs/cleanup or
  a created relocation destination, never skips or dry-run proposals. No whole-batch replay.
- Ordinary repair failures retain inspection/plain-lint guidance; relocation keeps its more
  specific reconciliation advice. CLI recovery survives returned output errors and transient
  causes must not advertise whole-command retry after commit.
- Existing JSON keys/types/vocabularies/exit codes are unchanged. Committed is internal, not a
  new wire field. Owner kept revision 1.89 as a restoration of existing prefix/retry promises;
  challenge that compatibility judgment against ADR-0008, not merely schema validity.
- No inode-continuity, crash transaction, final verify-to-replace race elimination, universal
  authorization migration or current-path-existence promise is made.

## Mandatory evidence floor

Verify the small actual producer/consumer chain with exact source declarations and fields.
Do not infer an API from this prompt or treat a planning AC as implementation proof.
Keep the verified consumer inventory to the boundaries needed for your conclusions.

Codex:
1. Independently seed a malformed repairable document; preserve unique concurrent bytes
   through a real FixFrontmatter caller. Include one completed earlier repair and a later
   refusal; verify source bytes, counts, type/mode and both primary/release causes.
2. Run compiler-valid separate mutations bypassing only in-place byte CAS and only post-wait
   authorization. Show the intended retained regression fails, then restore and run green.
   If time permits, discard the joined release error while keeping imports valid.

Antigravity:
1. Drive real lint --fix command/core paths for an ordinary committed prefix with a later
   failure, a failed JSON writer after success, and dry-run/skip-only contrast. Check stdout,
   stderr receipt, original causes/exit classification, call count and retry hints.
2. Mutate committed-prefix retry suppression and separately promote a skipped row to
   Committed. Require runtime assertion failures, not compile failures. Restore green.
3. Cross-check one actual filesystem raw-drift or finalization fixture so fabricated portable
   states cannot stand in for an end-to-end producer trace.

Both: run the suggested plain-lint recovery against a disposable partially repaired corpus.
Show what it diagnoses and what remains for operator choice; do not call an error-returning
lint invocation a failed recovery merely because intentional residual defects remain.
For a ready/no-findings assessment, include hostile evidence and the selected mutations.

## Required hostile angles

Settle each primary hypothesis as reproduced / falsified / unresolved:
- Named return/defer shadowing silently drops the primary or guard-release error.
- A guard failure after a no-op, denied authorization, or an earlier completed row is lost
  or incorrectly reclassified as a successful batch.
- Preview proposals or skipped rows become commits, or a partial relocation is counted fixed.
- A helper skips parse errors or a CAS check follows symlinks and overwrites a foreign entry.
- Failure after a later phase or output bypasses the ordinary prefix wrapper; a raw transient
  error causes machine clients to replay already committed work.
- Populated recovery fields validate but have misleading rows, flags, workspace or retry
  semantics. Ordinary receipts should not invent relocation evidence or a committed JSON key.

After the focused pass, perform a distinct systemic pass over shared helpers, adjacent
consumers and test fixtures that may make the invariants accidentally true. Separate actual
reachable defects from coverage gaps, accepted race windows and already tracked policy work.
A different approach or architectural preference is not itself a finding.

## Validation and restoration

Inside the independent captured sandbox, independently run:
- just build (use a sandbox-specific temporary GOCACHE when required).
- go test -race -count=3 ./internal/store -run 'TestFixInPlace|TestFixFrontmatter|TestVerifyFixInPlacePath|TestFixIDRelocation'
- go test -race -count=3 ./internal/core ./internal/cli ./internal/wire -run 'TestRepairPlanning|TestCLILintFix|TestPlanningRepair'
- just docs-check; the helper baseline includes the intentionally regenerated lint page.
- ./bin/tskflwctl -C . lint --links and audit lint for your assigned audit.

Owner ran full just test (race), just build, just lint, repeated focused races and schema-comment
generation. The owner's docs-check against the live index reports the intentional uncommitted
help diff; fresh generation into an independent temporary directory matches docs/cli exactly.
Distinguish inherited checks from your own execution. Only runtime assertions count as killed
mutations; the owner's first discarded-error mutation caused an unused import and was invalid,
then a compiling variant was killed. Never report the invalid attempt as safety evidence.

Use temporary caches, probe fixtures and transcripts outside the sandbox source tree, or remove
only reviewer-owned scratch files before verification. All mutations must be restored to your
captured baseline before transfer. No source-checkout writes, commits or pushes; the helper's
sandbox baseline commit is the sole allowed exception. Follow the injected create/verify/transfer
protocol; keep the sandbox on any failure and report incomplete rather than bypassing it.

## Deliverable

Preserve this brief; fill only your assigned audit's report. Start with findings/verdict and
scope limits. Consolidate by root cause; include actual reachable consequence, exact reproduction,
producer trace, recommendation and residual uncertainty. Create canonical findings with the
sandbox's fresh CLI and leave them open for owner triage.

Include a compact verified inventory, mutation table (exact edit/test/outcome/restored green),
bounded surviving-hypothesis ledger, executed versus inherited checks, and final helper isolation
attestation. Labels must distinguish real filesystem effects, synthetic portable EBUSY/EAGAIN,
hook-injected release errors and actual kernel errors. Return the guarded transfer receipt in
your handoff; report text cannot self-certify delivery. Do not claim universal hardening from a
small passing matrix. No giant API inventory or unsupported no-findings conclusion.

## Reviewer report

**Verdict: no demonstrated in-scope findings; the bounded implementation is ready for owner triage.** Four compiler-valid mutations were killed by the intended retained regressions, and independent real-filesystem probes preserved newer bytes and completed-prefix evidence. This is not task/Thread closure or a claim of universal persistence hardening.

Scope: ordinary in-place frontmatter repair, post-wait authorization, checked finalization, and the necessary core/CLI/wire recovery consumers. The final verify-to-replace window, inode continuity, crash transactions, broader create cleanup, relocation symlink policy and unrelated legacy body writers remain outside this verdict. I did not consult the sibling report. No findings were created because no reachable in-scope defect was demonstrated.

**Captured evidence and verified inventory**

The helper captured the uncommitted handoff into independent baseline
`4802130c5079cdaa84aabdd902c7f963e3b0b119`, whose parent is `43af9bb1`.
I read the implementing task's checkpoint and maintenance Thread in that captured
baseline. Their quality-gate claims are inherited evidence, not my test results.
The inventory below contains inspected implementation, not planned capabilities.

| Boundary | Verified implementation and evidence |
| --- | --- |
| Adapter producer | [fix.go:31](../../internal/store/fix.go#L31) returns named `results, err`; checked guard/deferred join/post-wait authorization are at lines 43–55. The hook and original-byte check precede atomic replacement at lines 131–141; a row is appended only after success. [fix.go:198](../../internal/store/fix.go#L198) uses `Lstat`, requires regular type, then calls [verifyFixPath:82](../../internal/store/fix_relocation.go#L82). [hashContent:25](../../internal/store/cas.go#L25) hashes full bytes with SHA-256, without typed decoding. |
| Guard and persistence helpers | [checkedWriteLock:115](../../internal/store/lock.go#L115) combines the canonical-root process mutex and checked platform release, then joins the test-hook error. On this Unix host [lock_unix.go:21](../../internal/store/lock_unix.go#L21) opens and flocks the planning root directory itself; there is no separate lock-file path. [writeFileAtomic:48](../../internal/store/atomic.go#L48) preserves existing permission bits, stages/fsyncs and renames. |
| Relocation distinction | [relocateFixedFrontmatter:16](../../internal/store/fix_relocation.go#L16) returns `FrontmatterRepairFailure` after destination creation if later checks/removal fail; it does not append a fixed row. The distinct outcome fields are declared in [frontmatter_repair.go:11](../../internal/core/frontmatter_repair.go#L11). |
| Core contract and consumer | [Fixer:444](../../internal/core/store.go#L444) requires completed/proposed prefix retention and prohibits whole-batch retry. [PlanningRepairResult:13](../../internal/core/planning_maintenance.go#L13) includes internal `Committed` and separate `Relocation`; [RepairPlanning:31](../../internal/core/planning_maintenance.go#L31) classifies non-skipped durable rows only outside preview, copies relocation evidence through `errors.As`, and adds [PlanningRepairFailure:36](../../internal/core/frontmatter_repair.go#L36) for ordinary failed progress. |
| CLI recovery | [runLintFix:85](../../internal/cli/lint.go#L85) wraps returned repair, post-lint/reporting and JSON-writer errors after progress; its error-output branch joins output and original causes. [planning_repair_failure.go:11](../../internal/cli/planning_repair_failure.go#L11) retains the result, preview flag and workspace. [exit.go:175](../../internal/cli/exit.go#L175) serializes the receipt and disables filesystem retry after commit; [fserror.go:21](../../internal/cli/fserror.go#L21) extracts OS causes and classifies transient errors. |
| Wire | [ToPlanningRepairEnvelope:18](../../internal/wire/planning_repair.go#L18) maps the ordinary fixed/skipped rows through [ToFixEnvelope:946](../../internal/wire/envelopes.go#L946), optionally copying relocation. [ErrorItem:1126](../../internal/wire/envelopes.go#L1126) already has optional `PlanningRepair *FixEnvelope`. `Committed` has no wire field. [LoadProblem:67](../../internal/core/entity_read.go#L67) uses `LocalPath`, separately from `Location`. |
| Shipped regressions and prose | [fix_inplace_test.go:35](../../internal/store/fix_inplace_test.go#L35) contains the five-kind edit/delete/byte-identical-symlink/directory matrix; lines 104, 152, 180, 211, 252 and 268 cover actual process-mutex waiting, joined prefix failures, preservation/preview, zero-write release failures, partial relocation and real `ENOTDIR`. [planning_maintenance_test.go:111](../../internal/core/planning_maintenance_test.go#L111), [planning_repair_failure_test.go:147](../../internal/cli/planning_repair_failure_test.go#L147) and [planning_repair_test.go:38](../../internal/wire/planning_repair_test.go#L38) cover portable progress, returned writer errors and populated schema validation. [persistence.md:111](../../docs/architecture/persistence.md#L111) and [generated lint help:15](../../docs/cli/tskflwctl_lint.md#L15) describe the bounded recovery contract. |

**Independent producer and recovery probes**

Reviewer-only test sources are preserved outside the sandbox source tree in
`/private/tmp/taskflow-codex-review-probes/`; they were removed from the checkout
before final verification. They are independent probe evidence, not shipped tests.
Transcripts named below are also outside the source tree.

1. `TestReviewerMalformedPrefixDriftAndRecovery` seeded two independent 0600 regular files,
   `tasks/000000000001-first.md` and `tasks/000000000002-later.md`. Each had
   `schema: 1`, a quoted matching ID, `status: pending`,
   `description: Reviewer: unique N`, `tags: alpha,beta`,
   `custom: keep # comment`, and distinct body text. The unquoted description colon
   made both documents malformed but text-repairable. The real pre-write hook appended
   `REVIEWER-UNIQUE-CONCURRENT-BYTES-6gmnhv65r2f0` to the second source, and the real checked-release
   seam returned a distinct injected error.
   Through `core.MustNewService(fs).RepairPlanning(false)`, assertions established:
   exactly two pre-write calls, one release call, one completed first-file row,
   `Committed=true`, no relocation, `errors.Is` for both conflict and release cause,
   exact expected repaired first bytes, exact concurrent second bytes, and regular
   type/0600 mode for both. No retained fixture helper seeded these files.
   Command: `go test -race -count=1 -v ./internal/store -run '^TestReviewer'`.
   Transcript: `/private/tmp/taskflow-codex-independent.log`.

2. The same test ran the freshly built binary's suggested recovery,
   `tskflwctl -C <disposable-root> lint --json`, against that partially repaired corpus.
   Exit 11 was expected: it identified the second source's malformed description
   (including the unquoted-colon explanation) and the first source's deliberately
   unrecognized `pending` status. It preserved the concurrent bytes. The operator
   still needs to inspect/quote the description while preserving the unique body
   and choose the intended valid task state. These residual defects are successful
   diagnostic recovery, not a failed recovery or permission to replay the batch.

3. `TestReviewerDeniedWaitAlsoRetainsRelease` held `processRepositoryLock(root)`,
   waited for the real repair's two pre-wait authorization checks, revoked an
   atomic policy flag, then released the mutex. With an independently injected
   release error, the real caller returned both denial and release causes, zero rows
   and untouched malformed source bytes. This combines denial and finalization,
   beyond the retained wait test's denial-only assertion. It exercises an actual
   process-mutex wait, not cross-process lock contention.

4. `TestReviewerActualStorePrefixThroughCLI` used a counting wrapper that delegates
   unchanged to real `FS.FixFrontmatter`. A repairable task was followed by an
   `epics` entry that was a regular file, producing an actual kernel `ENOTDIR`
   from directory reading after the task replacement. Executing the real
   `lint --fix --json` command through core returned exit classification 1,
   the original errno, one fixer call, repaired regular 0600 task bytes, and matching
   one-row stdout/stderr recovery with correct physical workspace, no relocation,
   `dry_run=false` and `retryable=false`.
   Separate cases omitted the blocking entry and made JSON output return a synthetic
   `*os.PathError{Op:"write", Path:"/dev/stdout", Err:EAGAIN}`, or joined that writer
   error to the real later `ENOTDIR`. Both retained the actual completed repair;
   the joined case retained both causes; all called the fixer once and suppressed retry.
   The successful core workflow's residual required-field diagnostics also survived
   in the writer-failure receipt.
   Command: `go test -race -count=1 -v ./internal/cli -run '^TestReviewerActualStorePrefixThroughCLI$'`.
   Transcript: `/private/tmp/taskflow-codex-real-cli.log`.

5. `TestReviewerPopulatedOrdinaryRecovery` semantically checked and JSON-Schema-validated
   an ordinary `ErrorEnvelope` with a completed row, skipped row, unreadable
   local/opaque locations, remaining dependency issue, populated workspace
   (`PlanningRoot, Space, RepoID, Branch, Checkout, ConfigPath, Source`) and
   nonempty filesystem details. Round-trip assertions checked the exact workspace
   and rows, false retry/dry-run, absent relocation and absent `committed`.
   Command: `go test -race -count=1 -v ./internal/wire -run '^TestReviewerPopulatedOrdinaryRecovery$'`.
   Transcript: `/private/tmp/taskflow-codex-populated-wire.log`.
   One initial scratch-fixture compile typo was corrected to the inspected
   `LoadProblem.LocalPath`; that compile failure is not validation evidence.

Real filesystem effects above are replacement, mode/type observations, byte drift
and `ENOTDIR`. Release failures are hook-injected. Writer `EAGAIN` and the
retained CLI tests' `EBUSY/EAGAIN` are synthetic portable errors, not observed kernel
busy-file behavior.

**Mutation evidence**

Every edit below was made separately in sandbox `internal/store/fix.go`, compiled,
and failed at a runtime assertion. Each original file was restored immediately and
the same test then passed under `-race -count=1`. There were no invalid mutation
attempts counted as kills. Commands used
`go test -race -count=1 -v ./internal/store -run '<selector>'`.

| Exact mutation | Retained test selector | Runtime result; restored result |
| --- | --- | --- |
| In `verifyFixInPlacePath`, `return verifyFixPath(path, version)` → `return nil`; keep `Lstat`/type checks | `^TestFixInPlaceRefusesRawDriftAcrossEntityKinds/tasks/edit$` | FAIL at `fix_inplace_test.go:73`: stale repair accepted, non-skipped row, nil error; restored PASS. |
| Only post-wait `if policyErr := s.authorizeMutation(); policyErr != nil {` → `if policyErr := error(nil); policyErr != nil {`; retain entry and lock-entry checks | `^TestFixInPlaceRechecksAuthorizationAfterWaiting$` | FAIL at line 142: post-wait denial ignored, nil error; restored PASS. |
| Deferred `err = errors.Join(err, fmt.Errorf("release repository frontmatter repair guard: %w", releaseErr))` → `err = errors.Join(err)`; imports remain valid | `^TestFixInPlaceRetainsCompletedPrefixAndReleaseFailure$` | FAIL at line 170: conflict and completed prefix survived, release cause disappeared; restored PASS. |
| In-place helper `info, err := os.Lstat(path)` → `info, err := os.Stat(path)`; preserve byte CAS | `^TestFixInPlaceRefusesRawDriftAcrossEntityKinds/tasks/symlink$` | FAIL at line 73: byte-identical symlink accepted as a repair, nil error; restored PASS. |

Full mutation transcripts:
`/private/tmp/taskflow-codex-mutation-byte-cas.log`,
`/private/tmp/taskflow-codex-mutation-post-wait-auth.log`,
`/private/tmp/taskflow-codex-mutation-joined-release.log`,
`/private/tmp/taskflow-codex-mutation-entry-type.log`.
The restored passes were observed separately, then the complete required focused
matrices were rerun after removing all reviewer-only source files.

**Hostile-hypothesis ledger and distinct systemic pass**

| Challenged hypothesis | Disposition and bounded evidence |
| --- | --- |
| Named return/defer shadowing loses a primary or release cause | **Falsified for exercised returns.** Independent prefix-plus-release and denial-plus-release probes retained both causes; release-only/no-op scenarios also pass. The compiling release mutation proves the retained regression detects this loss. |
| Guard finalization after zero writes or an earlier row becomes success | **Falsified.** Retained `TestFixFrontmatterReportsGuardReleaseFailureEvenWithoutWrites` covers repaired/no-op/skipped/preview; independent denial and durable-prefix probes add the combined causes. Preview acquired no write guard. |
| Proposals/skips become commits, or an interrupted move becomes fixed | **Falsified in the selected producer/consumer matrix.** Core classification test covers repair/cleanup/skip/no-op/preview/partial relocation; retained store partial-relocation/release test preserves the separate outcome. CLI dry-run/skip/no-progress contrasts pass. No claim that arbitrary third-party adapters satisfy the port. |
| Parsing hides malformed sources, or byte CAS follows a foreign symlink | **Falsified at the exercised pre-write boundary.** Independently malformed input was repaired/refused as appropriate; the five-kind matrix preserves edit/deletion/link/directory outcomes. The separate byte-CAS and type mutations each fail despite the other check remaining active. |
| A later phase or output error escapes prefix recovery or advertises replay | **Falsified for returned errors exercised.** Actual later `ENOTDIR`, actual completed task replacement, synthetic post-lint and writer failures retain receipts; fixer count is one. Core error wrapping and CLI named-return defer cover different failure stages. |
| Optional recovery validates but lies about flags, workspace or rows | **Falsified for selected non-default values.** Independent populated semantic/schema checks and actual CLI workspace/row assertions agree; no relocation or new `committed` JSON key appears. |
| Shared helpers/tests make all these checks accidentally true | **Falsified for the challenged mechanisms.** I inspected the raw hash, parse-free verifier, atomic helper and lock layering, then separately disabled byte equality, entry-type checking, late policy and release joining. No nearby caller rescued these mutants. Independent seeds bypassed the retained fixture helper, and actual `ReadDir` failure cross-checked portable receipt stubs. |
| Remaining check-to-effect races or legacy body-writer limitations undermine this bounded claim | **Accepted limits, not reproduced in-scope defects.** `Lstat`/read/rename are not one kernel transaction; an edit after the final check is still possible. [body.go:230](../../internal/store/body.go#L230) still supplies legacy `s.writeLock` to audit body writes, and [lock.go:103](../../internal/store/lock.go#L103) discards checked release there. That unchanged boundary is expressly excluded by this brief; no universal authorization/finalization migration is claimed. |

The systemic pass specifically followed `FixFindingHeaders`'s adjacent producer:
[finding.go:325](../../internal/core/finding.go#L325) accepts a changed successful body
transform and only appends a result when changes exist at line 343. The retained
portable body-prefix test exercises a later body failure, but does not establish
universal durability reporting for legacy writers. Likewise, the existing atomic
helper's best-effort directory sync is not evidence of a crash transaction.
These limits do not turn an interrupted relocation into an ordinary completed row.

**Machine-contract judgment**

I challenged retaining revision 1.89 against [ADR-0008 §2–3](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md#2-the-revision-covers-all-json-output-the-generated-schema-covers-typed-envelopes),
which covers meanings as well as shapes, and inspected
[wire.go:378](../../internal/wire/wire.go#L378). Schema validity alone would not
justify that choice. Here the already-existing optional `planning_repair` accepts
the same `FixEnvelope`; completed/skipped/dry-run meanings and exit vocabulary are
unchanged, and no consumer gains a new requirement or valid retry entitlement after
commit. Broader population of that optional recovery and disabling an unsafe retry
restore the already-existing no-whole-batch-retry port promise (also present in
parent `43af9bb1`'s `internal/core/store.go`). I found no incompatible meaning
change requiring a new revision for this bounded restoration. This does not
authorize future receipt-shape, vocabulary or promised-meaning changes without
ADR-0008 review.

**Executed checks versus inherited claims**

All reviewer Go commands used
`GOCACHE=/private/tmp/taskflow-codex-review-gocache`.

- **Executed PASS:** `just build`, initially and after restoring every mutation.
  Go emitted a nonfatal stat-cache permission warning for its default module cache;
  both builds exited 0 and produced the sandbox binary.
- **Executed PASS**, initially and again after restoration:
  `go test -race -count=3 ./internal/store -run 'TestFixInPlace|TestFixFrontmatter|TestVerifyFixInPlacePath|TestFixIDRelocation'`.
- **Executed PASS**, initially and again after restoration:
  `go test -race -count=3 ./internal/core ./internal/cli ./internal/wire -run 'TestRepairPlanning|TestCLILintFix|TestPlanningRepair'`.
- **Executed PASS:** the independent reviewer tests, four runtime mutation failures
  with restored-green counterparts, and non-default semantic/schema validation above.
- **Executed PASS:** `just docs-check`, twice. The captured baseline includes the
  intentional generated page, so regeneration produced no diff. This independently
  resolves the owner's live-index docs-check limitation for the captured state.
- **Executed PASS:** `./bin/tskflwctl -C . lint --links` and
  `./bin/tskflwctl -C . audit lint 6gmnhv65r2f0`; repeated after report editing.
- **Inherited only, not independently rerun:** owner's full `just test`,
  `just lint` and schema-comment generation. No claim of independent full-suite
  or full-linter execution is made.

Final restored-check transcripts are
`/private/tmp/taskflow-codex-final-build.log`,
`/private/tmp/taskflow-codex-final-store.log`,
`/private/tmp/taskflow-codex-final-consumers.log` and
`/private/tmp/taskflow-codex-final-docs.log`.
Only this audit report differs from the captured baseline; the brief is preserved.

**Isolation attestation**

```text
sandbox_path=/private/tmp/isolated-review.cAXHXI
git_dir=/private/tmp/isolated-review.cAXHXI/.git
baseline_commit=4802130c5079cdaa84aabdd902c7f963e3b0b119
source_blob=4faa68742837ecb880d3cb6bdf832102c6b977bc
source_fingerprint=65af7e2925c8b1665d95db704b8575ba78eac460
deliverable=planning/audits/6gmnhv65r2f0-2026-10-17-in-place-lint-repair-safety-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner disposition (2026-10-17)

Verified the source helper's successful guarded transfer receipt before owner
edits: baseline `4802130c5079cdaa84aabdd902c7f963e3b0b119`, transferred blob
`e8cc676ce2258c29d8bda14ada4b7d44273d9a56`, `verify=passed`,
`transfer=succeeded`, `delivery=confirmed`.

Accepted the bounded no-demonstrated-findings conclusion. Cross-checked the
producer/consumer implementation against the restored sandbox, inspected the
independent filesystem and real-store-to-CLI probe sources/transcripts, and
verified that only the assigned audit differs in that sandbox. The independent
denial-plus-release and actual `ENOTDIR`/writer-failure probes complement the
retained regression matrix; their scratch sources remain reviewer evidence, not
shipped test coverage. The compiling mutation evidence and populated wire checks
support the conclusion without treating schema validity as semantic proof.

No implementation changes are required by this report. Legacy body-writer cleanup
and final check-to-effect races remain the documented limits, not newly completed
work. Independent `just docs-check` passed on the captured baseline, resolving the
owner's live-index limitation for that state. The sibling review is still pending
at this disposition; neither the task nor the Thread is closed by this audit.

Subsequent owner reconciliation: Antigravity's M1/L1 exposed mixed-cause filesystem
classification and post-output residual-validation loss not detected by this
review. Both are fixed with retained red-first tests; see the sibling audit's
owner disposition. This report remains baseline evidence, not an independent
certification of those later changes or of its scratch probe's former exit-1
expectation when residual validation was present.
