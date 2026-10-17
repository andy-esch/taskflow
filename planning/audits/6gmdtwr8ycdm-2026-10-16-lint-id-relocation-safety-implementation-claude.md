---
schema: 1
id: 6gmdtwr8ycdm
bucket: closed
area: lint-id-relocation-safety-implementation-claude
date: "2026-10-16"
updated_at: "2026-10-16"
---
# Audit: Lint ID relocation safety implementation — claude — 2026-10-16

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

Challenge lint ID spelling relocation across raw-editor races and partial-failure
reporting. This fixes data preservation, not just prettier diagnostics. Work
independently and do not read the sibling report before yours is final.

- **Claude primary:** real filesystem creation/removal ordering, mode preservation,
  fresh authorization and CAS, exact untouched bytes. Cross-check partial progress
  through `Service.RepairPlanning` and one CLI recovery branch.
- **Antigravity primary:** core/CLI/wire receipt semantics, optional-field presence,
  schema compatibility and usable recovery without prose parsing. Cross-check one
  real-store partial failure. A green suite is not sufficient evidence.

## Review target

Source: `/Users/andyeschbacher/git/andy-esch/taskflow`, read-only handoff.
Branch `fix/guard-lint-id-relocation`, base
`e454d00bbad89b2dcd03f6944f1d742fb340988d`. Implementation is **uncommitted**:
capture staged, unstaged, untracked and deleted files, not HEAD alone.

Task: [Lint ID relocation](../tasks/6gmbfxy1d561-preserve-source-evidence-and-partial-outcomes-during-lint-id-relocation.md).
Thread: [Maintenance persistence](../threads/6gmdt09gygfw-harden-local-maintenance-and-relocation-persistence.md).
Audit origin: October 16 concurrency M1/M2. Its duplicate-valid-ID consequence
was narrowed during triage: actual spelling repair retains an invalidly named
original on failure, not the copied-valid-ID fixture used by that audit.

Verify a consumer inventory of these actual implementation paths:

- `internal/store/fix.go`: `FixFrontmatter`, `repairInvalidID`, completed rows and previews.
- `internal/store/fix_relocation.go`: `(*FS).relocateFixedFrontmatter`,
  `verifyFixPath`, two pre-create fault windows and pre-remove hook;
  `atomic.go`: `createFileAtomicExactMode` vs `writeFileAtomic`.
- `internal/core/frontmatter_repair.go`: `FrontmatterRepairFailure` and local action
  evidence; `planning_maintenance.go`: application copying and phase ordering;
  `store.go`: unchanged Fixer signature, extended typed-failure contract.
- `internal/cli/lint.go`, `planning_repair_failure.go`, `exit.go`,
  `render/render.go`: partial output, stdout failure, classification and retry hints.
- `internal/wire/planning_repair.go`, `envelopes.go`, `wire.go`: optional
  `relocation`, `error.planning_repair`, revision **1.89 ADDITIVE** under ADR-0008.
- Regression files: `store/fix_relocation_test.go`, `store/fix_test.go`,
  `core/planning_maintenance_test.go`, `cli/planning_repair_failure_test.go`,
  `wire/planning_repair_test.go`, populated cases in `wire/envelopes_test.go`.
- Owning guides: `docs/architecture/persistence.md` and `planning-data.md`.

Most golden changes are revision-only: owner structurally compared 40 JSON
goldens with `schema_version` removed. The schema golden has real new optional
branches; do not dismiss those as bulk generation noise.

## Intended contract to challenge

1. Relocation retains source mode, refuses destination overwrite, and rechecks
   original bytes plus existing spelling-reference/identity authorization before
   create. After create, both copies are checked before deleting the source.
2. Cooperating writers use the existing repository guard. Raw writers still have
   a final check-to-effect window. No crash-atomic move, inode continuity, reference
   cascade, rollback or automatic deletion of ambiguous copies is promised.
3. A failed move preserves earlier fixed/proposed rows and separate local action
   evidence. It is not a fixed or skipped row. Action flags record this repair's
   history, not whether another actor has since removed a path.
4. Core stops later maintenance phases on relocation failure, copies the outcome,
   retains domain/OS causes, and never retries the whole multi-file batch.
5. Human/JSON output cannot call a partial-only operation “nothing to fix” or count
   it fixed. JSON errors carry the same recovery, even with no earlier repairs or
   a failed stdout writer. A committed destination never advertises blind retry.
6. New metadata is optional, adapter-neutral operation evidence, not path authority
   or domain entity fields. Existing fixed/skipped meanings and exit codes remain.
   Preview writes/removes nothing and keeps the existing proposal shape.

## Mandatory evidence floor

Both reviewers: verify named symbols/tests, run focused race tests, perform one
successful hostile repair and one refused/partial repair through production code.
Check exact untouched bytes independently of decoded semantic values. Keep real
planning records out of fixtures.

- **Claude assigned mutation:** replace `createFileAtomicExactMode` with
  `writeFileAtomic` in the relocation helper, compiler-valid. Require
  `TestFixIDRelocationRefusesRawDestinationCollision` to fail, then restore and
  prove green. Its final `testHookBeforeFixIDExclusiveCreate` runs **after** fresh
  availability/authorization checks, so those checks cannot mask lost exclusivity.
  Owner already killed this mutation; independently reproduce it.
- **Antigravity assigned mutation:** disable outcome mapping in
  `ToPlanningRepairEnvelope` without breaking compilation. Require
  `TestPlanningRepairEnvelopeSeparatesPartialRelocationFromFixedRows` or a named
  JSON case of `TestCLILintFixRetainsPartialRelocationWithoutCountingItFixed` to
  fail for lost evidence, then restore and prove green. Explain why an optional
  field can disappear while schema validation alone stays green.
- Cross-check the other lens: Claude uses the CLI populated recovery/no-prefix
  case; Antigravity executes
  `TestFixIDRelocationRemovalFailureRetainsEarlierRepairsAndLocalOutcome` (real
  store -> real Service) and traces that receipt through the actual CLI mapper.
  Separate seam probes are valid; fabricated receipts are not end-to-end producer proof.
- Exercise non-default optional fields in both semantic assertions and actual
  JSON Schema validation. Also inspect `source_removed=false`, absent metadata on
  ordinary preview, filesystem retry hints and stdout-error joining.
- Execute the inspect-both-paths advice on the partial fixture: establish what
  remains, what changed, and why another automatic attempt must not delete a copy
  or be described as crash recovery. Never “repair” a real planning tree.

## Required hostile angles

Settle each important hypothesis as reproduced / falsified by a named probe /
unresolved, with residual uncertainty rather than a finding quota:

- Late occupied destination, source drift before creation or removal, destination
  edit/deletion after creation, added canonical/cross-kind owner or inbound reference.
- Earlier completed batch progress plus an injected unlink failure; both paths and
  correct action history survive without falsely adding a fixed row.
- Restrictive mode across all ID-led kinds; no preview effects, no generic
  graph-owned dependency normalization or reference-cascade authority.
- Relocation progress lost in the port/application/renderer/error/schema seams;
  a transient-looking OS error or stdout failure inviting duplicate batch replay.

Do a separate systemic pass over helper assumptions and adjacent callers. Ordinary
in-place lint CAS/guard finalization is tracked in `6gmdsxccqqnm`. Task rename's
destination-loss bug is separately reproduced and tracked in `6gmdsxcevgkr`;
`RenameTask` production is not changed here. Challenge the boundary only with an
in-scope consequence; do not report these as newly fixed or untracked discoveries.
Do not confuse source CAS with whole-corpus raw-writer serialization, or claim
action flags certify current path existence.

## Validation and restoration

Owner: `just build`, full `just test` (race), `just lint` (0 issues), focused
core/store/wire/CLI races with `-count=3`, populated schema cases, planning lint
with links. Goldens and schema comments regenerated. `just docs-check` reports
only the intentional uncommitted lint-help change; fresh temporary docgen output
matches `docs/cli`. You may inherit the full-suite baseline, but independently
execute your focused hostile/mutation floor.

Follow the injected independent-clone protocol before inspecting implementation.
Build that clone's own binary; use scratch fixtures/caches outside its tree and
avoid registry pollution. The source checkout stays read-only. Restore all probes
to the captured baseline, leave only your assigned audit changed, verify with the
source helper, then transfer just that audit. Only the helper's sandbox baseline
commit is permitted; no other commits or pushes. Keep the sandbox for receipt
confirmation. A blocked or degraded review is incomplete, never clean by omission.

## Deliverable

Preserve this brief; replace the report slot with findings-first verdict/limits,
verified consumer inventory, consolidated root causes, reproduction/producer trace,
expected vs actual behavior and a bounded fix. Distinguish product defects from
coverage gaps, design preferences, and already tracked work. Create canonical
findings through the sandbox CLI and leave them open for owner triage.

Include the mutation ledger (exact edit, intended test, compiled killed/survived/
invalid result and restored green run), a bounded surviving-hypothesis ledger,
and the helper's final verification attestation. Return the confirmable transfer
receipt in your handoff message; the report cannot attest its own later delivery.

## Reviewer report

Reviewer: Claude (Opus 5.5), 2026-10-16. Independent; the sibling Antigravity report
was not read. Line numbers refer to sandbox baseline `88674e7c`.

### Verdict and limits

**Not ready to close the review AC: three open findings (M1, M2, L1) for owner
triage.** The production relocation itself behaved correctly on every hostile probe.
It refused a late occupied destination, kept source and destination drift, rechecked
authorization, kept source mode, left untouched bytes exact, wrote nothing on preview,
and kept a real (not hook-injected) unlink failure with prefix rows plus the
receipt on both streams. The assigned mutation was killed by the intended test, and
four further guard mutations were killed by their named tests.

What does not hold:

- **M1 (product, recovery guidance).** The "inspect both paths before retrying"
  advice does not lead anywhere. Every retry is refused with "already exists". The
  residual plain-`lint` advice for the retained source says to rename it onto the
  occupied destination. Following that advice with `mv` destroyed a later destination
  edit in the reported two-copy state.
- **M2 (coverage gap; production verified correct).** Three of the change's own
  safety guards can be deleted, compiler-valid, with the full race suite green:
  exact-mode creation, the real `os.Remove` failure receipt, and the CLI retry
  override. Each regression uses a fixture value for which the guard is not needed.
- **L1 (claim truthfulness).** "Stdout failure cannot discard the recovery receipt"
  holds for error-returning writers only. A real broken stdout pipe ends the process
  with SIGPIPE (141) before any stderr recovery. Task scope forbids signal handling,
  so the fix is to narrow the claim.

Limits:

- Lens: Claude assignment (filesystem ordering, mode, fresh authorization/CAS, exact
  bytes), plus the CLI populated recovery and no-prefix cross-check. Antigravity's
  receipt/schema lens was covered only where listed below.
- macOS (APFS) only. The real unlink failure used `chflags uchg`, which is BSD/darwin
  specific. Raw-writer races were driven through the shipped test hooks. The production
  binary has no way to inject a writer between check and effect.
- Not repeated: full `just lint` and `just test`. The full baseline is inherited; I
  ran full `go test -race ./...` twice, and all focused race tests used `-count=3`
  (see Validation run).

### Consolidated root causes

1. **The two-copy state is reported as the safe outcome, but nothing completes from
   it** (M1). The change deliberately keeps both copies and tells the operator to
   retry. The fixer refuses an existing destination, and the pre-existing pure
   `flatNameProblem` advice assumes the canonical path is absent. Both are true at
   once only in the state this change now produces on purpose.
2. **The regression fixtures sit at points where the guard is not needed** (M2).
   `0o600` is unchanged by umask 022. `EIO` is already non-retryable. The removal
   hook fires before the code its test is named after. One pattern masks three
   independent guards.
3. **The seam writer models a stdout failure that a real Go stdout rarely produces**
   (L1). An injected `Write` error is tested. EPIPE on fd 1 is a fatal signal unless
   handled, and a closed fd 1 is reopened on `/dev/null` by the runtime.

## Findings

Created through the sandbox CLI and left open for owner triage.

#### M1. Partial-relocation recovery advice cannot complete and points lint at the occupied destination · **Status:** fixed

**File:** internal/store/flatname.go:75 | **Component:** store/cli recovery guidance
**Effort:** S · **Urgency:** soon

Product defect in recovery guidance. The change deliberately keeps both copies and
reports that two-copy state as the safe outcome (`internal/core/frontmatter_repair.go:27`,
lint help, `docs/architecture/persistence.md`). The only advice it gives is "inspect
both paths before retrying". Running that advice on the state that produces it
(production probes P2/P3):

- **What remains.** Both copies remain. The source bytes are unchanged (SHA), and
  `diff` shows only the `id:` line differs.
- **Retry.** `lint --fix`, before and after the unlink cause is cleared, always gives
  `skipped … id "6fbj87000lt6" not repaired: 6fbj870001t6-bad-id.md already exists`.
  That refusal is correct: no automatic attempt deletes or chooses a copy. But no
  retry can ever complete the move, and neither the error, the help, nor the guide
  says what does.
- **Inspection.** Plain `lint` reports the retained source as `unidentified task
  record … — rename to "6fbj870001t6-bad-id.md" and match the id: field`. That names
  the occupied destination.
- **Following that advice.** This was run on the post-state of
  `TestFixIDRelocationRetainsSourceWhenCreatedDestinationDrifts/edited`, where the
  destination carries a later raw edit. Plain `mv` onto the advised name left 0 of 1
  concurrent-edit lines, and the canonical file then carried `id: 6fbj87000lt6`.

`flatNameProblem` (`internal/store/flatname.go:62-83`) is pre-existing, pure over the
stem, and documented as "manual advice, not authority". Its unstated precondition is
that the canonical path is absent. This change makes that precondition false in
exactly the state it now reports.

Expected: guidance for the reported two-copy state leads to a resolution and never
targets an existing path. Actual: the retry advice cannot complete the move, and the
manual advice is destructive.

**Recommendation:** When the canonical-spelling sibling exists, replace the rename advice at the loader call site with compare-both-copies guidance that never targets that path, and state in the partial error, help and guide that lint --fix refuses while both copies exist and the operator removes the copy not kept; add plain-lint regressions on both partial fixtures.

**Resolution:** Confirmed. Shared directory scans now replace rename advice when
    the canonical sibling is occupied; pure filename advice is conditional on
    absence. Plain-lint and refused-retry regressions cover unchanged and edited
    two-copy states across tasks, audits, research and Threads without changing
    either copy. Partial error, human output, help and persistence guidance
    require deliberate reconciliation rather than blind retry.

#### M2. Relocation mode, real unlink receipt and retry override are unpinned by invariant-neutral fixtures · **Status:** fixed

**File:** internal/store/fix_relocation_test.go:138 | **Component:** store/cli regressions
**Effort:** S · **Urgency:** soon

Coverage gap; production behavior was verified correct in each case. One root cause:
each regression's fixture value is one the guard is not needed for, so deleting the
guard leaves the suite green.

1. **Exact mode.** Every mode test uses `0o600`, and `0o600 &^ 022 == 0o600`, so
   ordinary umask-filtered `createFileAtomic` produces the same mode. Mutation
   `fix_relocation.go:40` `createFileAtomicExactMode` → `createFileAtomic` survived
   full `go test -race ./...`. Consequence (mutant binary, `0664` source, umask 022):
   the destination became `0644` and group write was lost. The real binary keeps
   `0664`. No file under `internal/` mentions umask. AC1's "0600 source" fixture is
   itself umask-invariant. `rename.go:132` shares the unpinned helper (adjacent, not
   in scope).
2. **Real unlink failure.**
   `TestFixIDRelocationRemovalFailureRetainsEarlierRepairsAndLocalOutcome` injects
   through `testHookBeforeFixIDSourceRemove`. That hook fires before both post-create
   checks and before `os.Remove`, so the real failure branch at
   `fix_relocation.go:64-65` never runs. Mutation `return fail(...)` → plain
   `fmt.Errorf` survived the full suite. On a real EPERM (`chflags uchg`), the mutant
   binary emitted 0 stdout bytes and no `error.planning_repair`. The real binary
   emits the full receipt on both streams.
3. **Retry override.** The CLI recovery fixture's cause is `EIO`, which
   `classifyFilesystemError` already marks non-retryable. Neutering
   `exit.go:179` (`Retryable = false`) survived the cli and wire suites. A temporary
   seam probe with a `remove`/`EBUSY` cause reported `retryable: true` for a
   committed destination on the mutant and `false` on real code. The same held for a
   stdout `EAGAIN` joined with an `ErrConflict` cause.

**Recommendation:** Use fixture values the guards are needed for: a 0o664 or 0o666 source under an explicit umask (subprocess or serialized), a seam at the os.Remove call itself rather than before the post-create checks, and an EBUSY or EAGAIN cause in the CLI recovery fixture; re-kill mutations 2, 4 and 9.

**Resolution:** Confirmed coverage gaps; current production guards were correct.
    Added a Unix subprocess with explicit umask 022 and source mode 0664, moved
    removal-failure injection to the actual unlink call after content checks,
    and used EBUSY plus joined stdout EAGAIN fixtures. All three compiler-valid
    guard-removal mutations now fail their intended regressions, were restored,
    and focused race tests pass.

#### L1. Broken stdout pipe discards the recovery receipt despite the stdout-failure claim · **Status:** fixed

**File:** internal/cli/planning_repair_failure.go:9 | **Component:** cli recovery claims
**Effort:** XS · **Urgency:** eventually

Claim truthfulness, not data loss. `planning_repair_failure.go:9` says the receipt
"must survive even if stdout failed while reporting it", and the task checkpoint says
"stdout failure cannot discard the recovery receipt". Real-process probe P4: the real
binary on a real partial (`chflags uchg` unlink failure), with `--json` stdout piped
into a reader that closes immediately. The process exited 141 (SIGPIPE) with 0 bytes
on stderr; both copies stayed on disk. Go terminates on EPIPE for fd 1 unless SIGPIPE
is handled, so `TestCLILintFixKeepsRecoveryWhenStdoutFails`'s error-returning writer
models a failure a real stdout rarely produces. A closed fd 1 (`>&-`) does not fail at
all, because the runtime reopens it on `/dev/null`.

The task scope explicitly excludes adding signal handling, so the bounded fix is to
narrow the documented promise. The filesystem state stays safe, and plain `lint` still
reveals it (subject to M1).

**Recommendation:** Narrow the comment, guide and task claim to writer errors and state that a broken stdout pipe terminates the process before stderr recovery while the two-copy state remains for plain lint; leave CLI-wide SIGPIPE policy to a separate owner decision.

**Resolution:** Confirmed limitation, not a new signal-handling requirement.
    Code comment, persistence guide and task checkpoint now limit receipt
    preservation to returned writer errors and explicitly exclude process
    termination such as SIGPIPE before stderr reporting. The regression name
    reflects that scope; default Go signal policy remains unchanged.

## Review evidence

### Verified inventory

Each name below was checked in the sandbox by `grep -n`, by reading the file, or by
executing it. "Executed" means a named test run or a production-binary probe.

- `internal/store/fix.go`: `FixFrontmatter` (:27) takes the write lock (:38-44). Its
  misspelled-ID branch calls `relocateFixedFrontmatter(path, renamed, content, fixed)`
  (:97). The fixed row is added only after relocation returns nil (:101). On preview,
  `!dryRun` skips relocation and keeps the existing proposal row (:96, executed).
  `repairInvalidID` (:185) adds a non-ENOENT destination stat error (:215-216). Hook
  vars are at :170-171. The completed prefix is returned with the error (:133-151).
- `internal/store/fix_relocation.go`: in `(*FS).relocateFixedFrontmatter` (:16), the
  pre-create hook (:17) comes before the source CAS (:20). The authorization recheck
  via `repairInvalidID` (:26-32) is followed by the mode stat (:33). The final
  `testHookBeforeFixIDExclusiveCreate` (:37) runs before `createFileAtomicExactMode`
  (:40), whose IsExist error maps to `ErrConflict` (:41-42). The `fail` wrapper sets
  only `DestinationWritten: true` (:46-50). Next come the pre-remove hook (:51), the
  destination check (:58), the source check (:61) and `os.Remove` (:64-65).
  `verifyFixPath` (:74).
- `internal/store/atomic.go`: `writeFileAtomic` (:48) renames over the destination.
  `createFileAtomicExactMode` (:92) opens `O_EXCL` in place and calls `f.Chmod(perm)`
  (:106-111). It is unchanged by this diff; it shipped earlier with task rename
  (`internal/store/rename.go:132` is the other caller).
- `internal/core/frontmatter_repair.go`: `FrontmatterRepairRelocation` (:11-16),
  `FrontmatterRepairFailure` (:21-24) with `Error` (:26-28) and `Unwrap` (:30).
  `planning_maintenance.go`: `Relocation` field (:19). `errors.As` copies by value
  (:44-46) and returns before `FixFindingHeaders` (:50). `store.go`: the `Fixer`
  signature is unchanged; the typed-failure comment is at :451-452.
- CLI: `runLintFix` (`lint.go:80`) adds the relocation-aware branch (:86), the stdout
  join (:89) and the wrapper (:95-96). The new `planningRepairCommandFailure`
  (`planning_repair_failure.go:10-18`) is mapped in `exit.go:174-181`, including the
  retry override at :179. `render.go`: `PlanningRepairJSON` (:994),
  `PlanningRepairHuman` (:1001).
- Wire: `FrontmatterRepairRelocationJSON` and `ToPlanningRepairEnvelope`
  (`planning_repair.go:8-28`); `FixEnvelope.Relocation` (`envelopes.go:936`);
  `ErrorItem.PlanningRepair` (:1132); `SchemaVersion = "1.89"` with ADDITIVE note
  (`wire.go:378-383`). In the schema golden, all four relocation fields are
  required and `additionalProperties: false`.
- Goldens: independently re-compared. 41 changed `.golden` files: 40 are equal
  after recursively deleting `schema_version`. `schema_jsonschema.golden` differs
  only by revision plus the `planning_repair`/`relocation`/
  `FrontmatterRepairRelocationJSON` branches. `machine_contract_revision.txt` is `1.89`.
- Regressions executed (all pass, `-race -count=3`): the seven
  `TestFixIDRelocation*` tests (`fix_relocation_test.go:26-232`),
  `TestFixRepairsMisspelledIDInFilenameAndFrontmatter` (`fix_test.go:143`),
  `TestRepairPlanningRetainsPartialRelocationBesideCompletedPrefix`
  (`planning_maintenance_test.go:93`), all four
  `TestCLILintFixRetainsPartialRelocationWithoutCountingItFixed` subcases,
  `TestCLILintFixKeepsRecoveryWhenStdoutFails`,
  `TestPlanningRepairEnvelopeSeparatesPartialRelocationFromFixedRows`, and the
  populated `FixEnvelope`/`ErrorEnvelope` cases (`envelopes_test.go:416,422`).
- Guides: `docs/architecture/persistence.md` "Ordinary frontmatter ID repair" and the
  `planning-data.md` relocation paragraph link to existing files. `just docs-check`
  exits 0 in the sandbox.
- Planned, not shipped: the review AC on task `6gmbfxy1d561` is unchecked. The
  adjacent tasks `6gmdsxccqqnm` and `6gmdsxcevgkr` exist as planning records only.

### Hostile angles settled

| Hypothesis | Result | Evidence |
| --- | --- | --- |
| Late occupied destination clobbered | Falsified | Mutation 1 killed; P1 |
| Source drift before create or remove lost | Falsified | Mutations 5, 7; tests |
| Destination edit/delete after create drops source | Falsified | Mutation 6; test |
| New canonical/cross-kind owner or inbound ref | Falsified pre-create | Mutation 3; P6 |
| Ref added after create, before unlink | By design: final window | Contract 2; code read |
| Earlier prefix + real unlink failure | Falsified (correct) | P2 real EPERM |
| Partial counted fixed or "nothing to fix" | Falsified | P2, P3 human/JSON |
| Mode lost across ID-led kinds | Falsified in prod; unpinned | P1, mutation 2 (M2) |
| Preview writes or adds relocation | Falsified | P5, dry-run test |
| Graph-owned normalization during move | Falsified | P6 skip row |
| Receipt lost at port/app/render/error/schema | Falsified in prod; unpinned | P2, mutation 8 (M2) |
| Transient OS error invites blind retry | Falsified in prod; unpinned | S1, mutation 9 (M2) |
| Stdout failure discards receipt | Reproduced for broken pipe | P4 (L1) |
| Inspect-both-paths advice converges safely | Reproduced: does not | P3 (M1) |

### Mutation ledger

All edits are single-site and compiler-valid (`go vet` clean). Each was restored with
`git checkout -- <file>`, and the package suite was rerun green after each one.

| # | Edit | Intended test | Result |
| --- | --- | --- | --- |
| 1 | `fix_relocation.go:40` `createFileAtomicExactMode` → `writeFileAtomic` (assigned) | `TestFixIDRelocationRefusesRawDestinationCollision` | **Killed**; the only store failure. Winner bytes overwritten and source removed |
| 2 | `:40` → `createFileAtomic` (umask-filtered) | mode tests | **Survived full `go test -race ./...`** |
| 3 | `:30` `if false && target != destination` | `...RechecksSpellingAuthorizationBeforeCreate` | Killed |
| 4 | `:65` `return fail(...)` → plain `fmt.Errorf` | `...RemovalFailureRetainsEarlierRepairsAndLocalOutcome` | **Survived full suite** |
| 5 | `:20` pre-create source CAS disabled (`false && err != nil`) | `...RefusesRawSourceChangeBeforeCreate` | Killed; coordinated with the post-create check, which would otherwise keep the source |
| 6 | `:58` destination check disabled | `...RetainsSourceWhenCreatedDestinationDrifts` | Killed |
| 7 | `:61` post-create source check disabled | `...KeepsRawSourceChangeBeforeRemove` | Killed |
| 8 | (consequence of 4 on the real path) | P2 rerun with mutant binary | Empty stdout, no `planning_repair` |
| 9 | `exit.go:179` override → `_ = ...Retryable` | `TestCLILintFixRetainsPartialRelocation...` | **Survived cli+wire suites**; temp seam probe S1 kills it |

Restored green: `go test -race -count=1 ./internal/store/` ok after 1 and after 3-7;
`./internal/cli/` ok after 9. Final full run under Validation.

### Production probes

Scratch trees were built from `tskflwctl init --no-register` with
`XDG_CONFIG_HOME`/`TSKFLW_CONFIG_HOME` pointed at scratch. No real planning record
was used. Fixture: `tasks/6fbj87000lt6-bad-id.md` (body ends in CRLF, comment and
unknown key) plus an earlier `000000000001-earlier.md` with `tags: a,b`.

- **P1, successful hostile repair** (umask 022, source `0664`). Real binary: the
  destination is `0664`, and its bytes match the source with only the `id:` line
  changed (`cmp` against an independently built expected file). The receipt has no
  `relocation` key (`has("relocation") == false`). Mutant 2 binary: destination is
  `0644`.
- **P2, refused/partial repair with a real unlink failure** (`chflags uchg` on the
  source, no hooks). Exit 1. Stdout `fixed` holds only the earlier row. `relocation`
  has `destination_written: true` and `source_removed: false`.
  `error.planning_repair` is identical to stdout (`diff` of sorted JSON).
  `error.filesystem` is `{permission, remove, retryable: false}`. Both copies exist
  and the source SHA is unchanged. With no prefix, human output shows "partial ID
  relocation" without "nothing to fix" or "file(s) fixed".
- **P3, executing the advice.** On the P2 state, `diff` shows only the `id:` line
  differs. `lint --fix` was retried before and after clearing `uchg`. Both runs give
  `skipped ... 6fbj870001t6-bad-id.md already exists`; nothing is deleted and nothing
  completes. Plain `lint` reports the source as `unidentified task record ... —
  rename to "6fbj870001t6-bad-id.md"`. On the post-state of
  `...WhenCreatedDestinationDrifts/edited`, `mv` as advised left 0 concurrent-edit
  lines and `id: 6fbj87000lt6` under the canonical name (M1). This is not crash
  recovery. The tool never chooses between copies, and that is correct: only a human
  can tell which copy carries intent.
- **P4, stdout failure.** With stdout closed (`>&-`), the run succeeded because the
  runtime reopened fd 1 on `/dev/null`. With stdout piped to a reader that closes
  immediately, exit was 141, stderr was 0 bytes, and both copies remained (L1).
- **P5, preview.** `--dry-run --json` gives `dry_run: true`, no `relocation` key, and
  the existing proposal rows. Both files are SHA-identical before and after.
- **P6, batch and authority.** In the same batch, a second misspelling of one
  canonical ID is refused: "already used by 6fbj870001t6-b.md" (same kind) and
  "already used by task ..." (cross kind). A misspelled task with `depends_on: a,b`
  is skipped with "graph-owned repair refused" and is not moved.
- **S1, temporary CLI seam probe** (deleted). A `FrontmatterRepairFailure` with
  `remove`/`EBUSY` gives `retryable: false` on real code and `true` on mutant 9. An
  `ErrConflict` cause joined with a stdout `EAGAIN` writer error gives
  `filesystem={io, write, /dev/stdout}`, retryable false on real code and true on the
  mutant.
- **S2, temporary wire schema probe** (deleted). The non-default
  `source_removed: true`, `destination_written: false`, `dry_run: true` and skipped
  row survive the mapper. They validate as `FixEnvelope` and nested in `ErrorEnvelope`
  with `filesystem`. Negative controls fail: a missing `source_removed` and a string
  `source_removed` are both rejected.

### Surviving-hypothesis ledger

- **`createFileAtomicWithMode` cleanup removes by path** (`atomic.go:103-121`). After
  a write/chmod/sync/close error it runs `_ = os.Remove(path)`. A raw writer that
  replaced the just-created name in that window would be deleted, and a failed
  cleanup leaves a destination with `Relocation == nil`. *Unresolved*: needs an I/O
  fault plus a concurrent raw writer at a new name; not reproduced. The helper is
  shared with task rename.
- **Stdout-join attribution** (`lint.go:89`, the only `errors.Join(err, outputErr)`
  in `internal/cli`). When the cause has no `PathError`, `error.filesystem` describes
  the stdout write (S1). *Reproduced at seam, not filed*: retry stays false under the
  override, and Go's poller and EINTR handling make a transient stdout class
  practically unreachable for a real stdout.
- **`SourceRemoved` is always false in every producer failure**, and `Error()`
  hardcodes "was written but source removal ... did not complete". Consistent today;
  a design note only.
- **Durability ordering.** File `Sync`, then best-effort directory sync, then unlink,
  with no sync after unlink. A crash can at worst bring back the source, which gives
  two copies and a safe refusal. Where directory fsync is rejected, ordering is
  filesystem-defined. This is covered by "not crash-atomic".
- **Not preserved:** xattrs, ACLs, owner and inode. Covered by "no inode continuity".
- **Antigravity-only angles not repeated:** wire compatibility classification under
  ADR-0008 beyond the golden comparison above. Why an omitted optional field passes
  schema validation is plain from the definition: `relocation` is not required.

### Validation run

- `just build` in the sandbox (`v0.23.0-107-g88674e7c`).
- Focused `go test -race -count=3` for store (`TestFixIDRelocation|TestFixRepairsMisspelledID|TestFixRefuses`),
  core (`TestRepairPlanning`), wire (`PlanningRepair|Envelope`) and cli
  (`TestCLILintFix`): all ok.
- Full `go test -race ./...` ran green with mutant 2 applied (the survival
  evidence). It was rerun on the restored baseline with the report and findings in
  place: 29 packages ok, no failures.
- `just docs-check`: exit 0.
- After finding creation, using the sandbox binary: `audit lint` gave "all audit
  findings pass lint", `lint --links` gave "all planning entities and dependency
  links pass lint", and `audit findings 6gmdtwr8ycdm` lists M1, M2 and L1 as open.
- Temporary probe files `internal/cli/zz_review_probe_test.go` and
  `internal/wire/zz_review_probe_test.go` were deleted. Every mutation was restored
  with `git checkout`. `git status --short` shows only this audit modified.

### Isolation attestation

This is the output of the source helper
`scripts/isolated-review-workspace.sh verify --sandbox <sandbox>`, run after all
probes were restored:

```text
sandbox_path=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/9fc0e07e-b675-43a1-b20f-d105213733cd/scratchpad/sandboxes/isolated-review.yvcQ0s
git_dir=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/9fc0e07e-b675-43a1-b20f-d105213733cd/scratchpad/sandboxes/isolated-review.yvcQ0s/.git
baseline_commit=88674e7c81863597eb8d49930d9ea6fb144963ee
source_blob=9f7efe90456a6d9cd4057c3808a6784e132a19af
source_fingerprint=6e2d10482c7e483e611cc0bb221afdea9c97e58a
deliverable=planning/audits/6gmdtwr8ycdm-2026-10-16-lint-id-relocation-safety-implementation-claude.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

The transfer receipt is returned in the handoff message, not in this report.

## Owner disposition (2026-10-16)

The source helper's `receipt` confirmed `transfer=succeeded`, `delivery=confirmed`
and transferred blob `d234182be721419417552f911e0d9763fee11807` before owner edits.
Reviewer evidence above is retained as the captured baseline assessment; finding
statuses and resolution notes reflect subsequent owner fixes, not a re-review.

M1, M2 and L1 are accepted and fixed. Post-review evidence and exact mutation
outcomes are recorded in the
[implementation task](../tasks/6gmbfxy1d561-preserve-source-evidence-and-partial-outcomes-during-lint-id-relocation.md#claude-review-disposition-2026-10-16).
The shared exclusive-create cleanup/raw-replacement hypothesis remains unproven
and is carried into the
[existing entity-integrity design task](../tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md#newly-observed-questions),
not silently treated as a confirmed incident or required expansion of this fix.
No SIGPIPE behavior was changed. Full race tests, build, lint and planning lint
pass. This audit has served its purpose and is closed; the implementation's review
AC remains unchecked pending the separate Antigravity report.
