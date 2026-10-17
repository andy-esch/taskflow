---
schema: 1
id: 6gmke6f7nnw0
bucket: closed
area: task-rename-destination-safety-implementation-antigravity
date: "2026-10-17"
updated_at: "2026-10-17"
---
# Audit: Task rename destination safety implementation — antigravity — 2026-10-17

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

Challenge a bounded source-deletion safeguard: can task rename still remove its
last valid source or describe committed work/retry safety incorrectly after a raw
writer changes the newly created destination? Treat the stated contract as a
hypothesis, not a verdict. No findings or clean assessment have been prewritten.

Claude: primary persistence ordering, source/destination evidence and reachable
data loss; cross-check one actual core/CLI recovery test. Antigravity: primary
core/CLI error receipt, retry and compatibility; cross-check the real-store drift
test. Both do a separate systemic pass, independently of the sibling report.

## Review target

Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
Base: merged main `36348b5b` (PR #312), **plus all staged, unstaged and untracked
implementation captured by the sandbox helper**. Reviewing HEAD alone misses this work.
Task: `planning/tasks/6gmdsxcevgkr-verify-the-task-rename-destination-before-removing-its-source.md`.
Thread: `planning/threads/6gmdt09gygfw-harden-local-maintenance-and-relocation-persistence.md`.

Verify these entry points in your captured source before claiming coverage:

- `internal/store/rename.go`: `FS.RenameTask`, `verifyTaskRenameDestination`,
  `verifyTaskRenamePath`, the existing pre-source-removal hook, and unchanged
  exclusive create in `internal/store/atomic.go`.
- `internal/store/rename_destination_test.go` and existing `rename_test.go`.
- `internal/core/task_rename.go`, `task_rename_test.go`, and
  `Service.RenameTask` in `internal/core/service_task.go`.
- `internal/cli/task.go` rename command; `internal/cli/exit.go` recovery mapping;
  `internal/cli/task_rename_test.go` (port stubs, not real filesystem interleavings).
- Unchanged `internal/wire/task_rename.go` DTOs/mappers; current revision 1.89;
  `docs/architecture/persistence.md#ordinary-create-and-task-rename` and ADR-0008.

Require a short verified consumer inventory, not a reconstructed architecture:
at most eight rows, each naming the actual symbol, exact source declaration or
field access, producer/consumer and test command. Copy declarations from the
sandbox (`go doc` / `nl -ba`); never paraphrase signatures into invented APIs.
Before submitting, reconcile each claimed type/field/helper against that source.
A mismatch makes the affected evidence incomplete, not verified. The previous
review round exposed reports inventing hard-link writes, lock files and receipt
fields; a passing sandbox transfer only proves delivery, not technical accuracy.

## Intended contract to challenge

1. Links commit first, exclusive destination creation second, source removal last.
   Immediately before removing a distinct old source, the destination must remain
   a regular file with the exact materialized bytes, and source CAS must still pass.
2. Destination deletion, edit, directory or byte-identical symlink must not justify
   removing source. Refusal leaves raw changes intact; no rollback chooses a copy.
3. Errors retain original causes and historical progress: committed inbound links
   and destination write remain reported; source removal/completion remain false.
   Flags do not certify current existence after an external writer acts.
4. Core decides recovery. Pre-destination prefixes remain convergent by stable ID;
   post-destination failures require inspection/reconciliation, not blind retry.
   A transient-looking filesystem cause must not contradict that advice in JSON.
5. Preview, same-path retitle, no-op, mode preservation, authorization, complete
   success and completed-but-guard-release-failed behavior remain compatible.

No crash atomicity, raw-editor locking, inode continuity or elimination of final
check-to-effect races is promised. An identical regular-file replacement is not
an inode-identity violation of a byte-CAS contract. No new fields or closed vocabulary
were added. Owner classifies this as restoration of existing safety/recovery
semantics with revision 1.89 retained; challenge that against ADR-0008 if justified.

Already tracked, not silently solved here: ordinary in-place lint CAS/guard release
(`6gmdsxccqqnm`), shared create cleanup/raw-path replacement and general durability
semantics (`6g7wsr8yvt1a`). Do not reopen completed lint spelling relocation merely
because it is adjacent; explain a demonstrated in-scope consequence or name the
appropriate follow-up. No universal persistence abstraction is requested.

## Mandatory evidence floor

Build your own captured binary. Independently run the focused tests below with
`-race -count=3` in your sandbox; the full owner suite may be inherited:

```sh
go test -race -count=3 ./internal/store -run 'TestRenameTask|TestVerifyTaskRenameDestination'
go test -race -count=3 ./internal/core ./internal/cli -run 'TestRenameTaskPreservesCause|TestTaskRename|TestWriteErrorCarriesStructuredTaskRenameRecovery'
```

Required primary-lens falsification, then restore and run green:

- Claude: compiler-valid removal of the production call to
  `verifyTaskRenameDestination` must fail
  `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt`'s deletion cases.
  Separately bypass only its regular-file check: a byte-identical symlink must fail
  the symlink cases even though content hashes agree. Retain source-CAS to prove
  it cannot mask the missing destination guard.
- Antigravity: compiler-valid removal of the task-rename filesystem retry override
  in `WriteError` must fail `TestTaskRenameCommandRetainsDestinationFailureAndRequiresInspection`
  (`transient/json`) and `TestTaskRenameFilesystemRetryHintRespectsDestinationCommit`.
  Demonstrate why the raw `EBUSY` fixture is retryable while the committed command
  is not; pre-destination retry behavior is a required contrast, not collateral damage.

Each reviewer executes at least one hostile real-store drift case through
`FS -> Service.RenameTask`, confirming exact original source bytes/mode, changed
destination state, cascade prefix counts and incomplete/error result. The retained
production-hook fixture is acceptable; label the hook precisely and do not claim
an actual kernel unlink failure. `ENOTDIR` is a real path-inspection failure from a
direct verifier probe; CLI `EBUSY` uses an injected portable result, not a real busy file.

Run the suggested recovery against the state that emits it: missing destination,
two surviving divergent copies, or a replaced destination must not lead to deletion
of unique content. Inspect before retrying and distinguish terminal completion from
resumable pre-destination prefix. Use scratch records, never real planning members.

No new optional fields were added. Verify the existing populated error receipt and
unchanged `TaskRenameJSON` mapping instead of inventing new keys/schema branches.
If a test cannot exercise a claimed boundary, mark that as a coverage limit.

## Required hostile angles

- Destination check ordering relative to post-write hooks, original source CAS,
  unlink and receipt updates. Can any error claim completion or erase a prefix?
- Non-regular destinations versus byte equality; unreadable/missing paths versus
  arbitrary identity/inode expectations. Preserve sentinel and OS causes correctly.
- Same-path retitle/no-op versus distinct-path move; dry-run must never reach effects.
- Does the new check accidentally re-resolve a duplicate-ID pair after creating the
  destination? Verification should use the already planned local path and bytes.
- Command output and recovery must preserve actual paths, stable ID, counts,
  booleans and original error class without claiming filesystem work for a port fake.
- Could retry hints contradict core after destination creation, source removal,
  or guard-release failure? Do not turn safe pre-destination prefix recovery into
  a blanket no-retry policy by accident.
- Separate systemic pass: helpers or invariant-neutral fixtures that can mask a
  missing guard; adjacent callers with demonstrated consequences; guide/type-comment
  truthfulness and compatibility classifications. Consolidate related root causes.

## Validation and restoration

Owner: build, full race suite, lint (0 issues), docs-check, schema-comment generation,
focused rename/repair races (`-count=3`), plus compiler-valid guard-removal checks
for destination verification, regular-file requirement and CLI retry suppression.
All mutations were restored. Independently repeat your primary floor, not just
the owner's commands. Planning `lint --links` and `audit lint` are also green.

Follow the mandatory independent-clone protocol injected above. The source stays
read-only; do not create a Git worktree or alter a sibling audit. Scratch fixtures
and caches stay outside the captured tree (or remove only your own probes).
The helper baseline is your only allowed commit; all other commits/pushes are forbidden.
Restore probes to that baseline, leave only your assigned audit changed, verify
using the source helper, and transfer only that audit. Keep the sandbox until
the owner confirms its receipt. A blocked/partial review is incomplete, never clean.

## Deliverable

Preserve this brief. Supply findings first, then a compact verified inventory,
consolidated producer trace, mutation ledger (edit, intended test, compiled
killed/survived/invalid outcome, restored green) and bounded surviving-hypothesis
ledger. Distinguish defects, coverage gaps, preferences and already tracked work.
Create canonical findings with the sandbox CLI and leave them open for owner triage.

Keep the report concise: no giant pseudocode inventory or unsupported claim that
every hostile angle was falsified. Quote only the few actual declarations/error
fields needed to reproduce conclusions. Explicitly identify synthetic fixtures,
inherited checks and checks not run. Include final helper verification attestation;
return its guarded transfer receipt in the handoff, not a self-attested delivery claim.

## Reviewer report

Reviewer: Antigravity, 2026-10-17. Independent review from isolated clone `/private/tmp/isolated-review.rKZHgW` (baseline commit `aff2bda6`); the sibling Claude report was not read.

### Verdict and scope limits

**Not ready to close the independent-review AC without triaging one open finding (`M1`).**

- **Primary contract (destination safety before source removal, core/CLI error receipt, retry hint suppression, and wire compatibility):** Verified sound across real-store, core, and CLI boundaries. `(*FS).RenameTask` (`internal/store/rename.go:161-170`) checks `verifyTaskRenameDestination(plan.newPath, hashContent(plan.renamedContent))` (`os.Lstat`, `info.Mode().IsRegular()`, and `verifyTaskRenamePath` SHA-256 byte check) and `verifyTaskRenamePath(source.path, source.version)` before `os.Remove(source.path)`. Destination deletion, edit, directory replacement, and byte-identical symlink replacement all return `domain.ErrConflict` while retaining intact `0600` source bytes, untouched raw destination changes, and durable cascade prefixes (`Committed: true`, `DestinationWritten: true`, `SourceRemoved: false`, `Complete: false`). `(*Service).RenameTask` (`internal/core/service_task.go:1023-1031`) and `taskRenameFailureRemedy` (`internal/core/task_rename.go:106-117`) select stage-appropriate recovery without replaying the rename or fabricating local paths for pathless stores. `WriteError` (`internal/cli/exit.go:195-197`) suppresses `payload.Error.Filesystem.Retryable = false` when `renameErr.receipt.DestinationWritten` is true while preserving `Retryable: true` for transient `syscall.EBUSY` failures on resumable pre-destination cascade prefixes (`DestinationWritten == false`). Retaining `wire.SchemaVersion = "1.89"` (`internal/wire/wire.go:383`) is truthful under ADR-0008.
- **Systemic second-pass finding (`M1`):** While `verifyTaskRenameDestination` (`internal/store/rename.go:336`) and `createFileAtomicExactMode` (`internal/store/atomic.go:97`, `O_EXCL`) treat symlinks as existing directory entries, the pre-flight target check `(*FS).ensureTaskRenameTargetAvailable` (`internal/store/rename.go:308`) uses `os.Stat(newPath)` instead of `os.Lstat(newPath)`. When `newPath` is a pre-existing dangling symlink, `os.Stat` follows the broken link and returns `ENOENT` (`os.IsNotExist(err) == true`), so `ensureTaskRenameTargetAvailable` returns `nil`. On `--dry-run`, `RenameTask` reports a clean preview (`err == nil`, `Changed: true`) for a target filename that `O_EXCL` will reject. On a real write with any inbound link (`len(plan.cascadeEdits) > 0`), `RenameTask` rewrites and commits all cascade documents to point to the dangling symlink (`Committed: true`, `AppliedDocuments: 1`, `AppliedLinks: 1`, `DestinationWritten: false`) before `createFileAtomicExactMode` fails with `domain.ErrConflict`. `TestRenameTask_ExclusiveDestinationCreatePreservesDanglingSymlink` (`internal/store/rename_test.go:474-498`) masked this defect by calling `os.Remove(bPath)` at line 476 and planting the symlink before `RenameTask` rather than via `testHookBeforeTaskRenameDestinationCreate`.

## Findings

#### M1. Pre-flight target availability follows symlinks while dangling-symlink regression masks cascade prefix · **Status:** fixed

**File:** internal/store/rename.go:308 | **Component:** store/rename target check and test fixture
**Effort:** XS · **Urgency:** soon

`(*FS).ensureTaskRenameTargetAvailable` (`internal/store/rename.go:302-314`) runs before the inbound-link cascade loop (`internal/store/rename.go:74,242`) and on `--dry-run` (`internal/store/rename.go:52`), but calls `os.Stat(newPath)` (`internal/store/rename.go:308`) instead of `os.Lstat(newPath)`. When `newPath` is a pre-existing dangling symlink, `os.Stat` follows the symlink to its missing target and returns `ENOENT` (`os.IsNotExist(err) == true`), so `ensureTaskRenameTargetAvailable` returns `nil`:

1. **Multi-step real write commits cascade links before `O_EXCL` fails:** When the renamed task has any inbound links (`len(plan.cascadeEdits) > 0`), `RenameTask("old", "New title", false)` passes `ensureTaskRenameTargetAvailable`, durably rewrites every inbound-link document in `plan.cascadeEdits` (`internal/store/rename.go:95-113`) to point to the colliding dangling symlink `<id>-<newSlug>.md`, sets `result.Committed = true` (`AppliedDocuments: 1, AppliedLinks: 1, DestinationWritten: false`), and only then fails at `createFileAtomicExactMode` (`internal/store/atomic.go:97`, `os.O_CREATE|os.O_EXCL`) with `domain.ErrConflict: target filename already exists: 6fjangd7kva1-new-title.md` (`internal/store/rename.go:135-137`). `Service.RenameTask` (`internal/core/service_task.go:1026-1028`) then wraps that pre-existing target collision as a `*core.TaskRenameFailure` whose `Remedy` advises rerunning the rename by stable ID, while inbound links on disk now point at the dangling symlink instead of the retained `source.path`.
2. **Dry-run preview false positive:** `RenameTask("old", "New title", true)` returns `err == nil` (`DryRun: true, Changed: true, PlannedDocuments: 2, PlannedLinks: 1`) even though `newPath` is already occupied by a directory entry that `createFileAtomicExactMode` will reject with `domain.ErrConflict`.
3. **Invariant-neutral test fixture masked the pre-cascade pass-through:** `TestRenameTask_ExclusiveDestinationCreatePreservesDanglingSymlink` (`internal/store/rename_test.go:474-498`) seeds `renameRepo(t)`, calls `os.Remove(bPath)` at line 476 to delete the inbound-link document, plants the dangling symlink before invoking `RenameTask`, and asserts `!result.Committed`. Deleting `bPath` emptied `plan.cascadeEdits`, masking that `ensureTaskRenameTargetAvailable` returned `nil`; and planting the symlink before `RenameTask` (rather than in `testHookBeforeTaskRenameDestinationCreate` as `TestRenameTask_ExclusiveDestinationCreateRejectsPostCASTarget` does at `internal/store/rename_test.go:452`) only reached `createFileAtomicExactMode` because `ensureTaskRenameTargetAvailable` failed open. Retaining `bPath` in `TestRenameTask_ExclusiveDestinationCreatePreservesDanglingSymlink` immediately fails (`Committed: true, AppliedDocuments: 1, AppliedLinks: 1`).
4. **Adjacent `internal/store` callers sharing the same root cause (for follow-up triage in `6g7wsr8yvt1a` / `6gmdsxccqqnm`):** `(*FS).repairInvalidID` (`internal/store/fix.go:213`), `(*FS).materializeThreadCreation` (`internal/store/threadcreation.go:135`, called by `(*FS).MutateThreadApply` at `internal/store/threadapply.go:87` before task dependency writes commit at `:127-145`), `(*FS).createEntityFile` (`internal/store/create.go:85`), and `(*FS).materializeTaskLifecycleCreation` (`internal/store/lifecyclemutation.go:255`) also use `os.Stat` before `O_EXCL` creation, and `verifyFixPath` (`internal/store/fix_relocation.go:58,83`) reads the destination with `os.ReadFile` without an `os.Lstat` regular-file check.

**Recommendation:** Use os.Lstat in ensureTaskRenameTargetAvailable so pre-existing dangling symlinks fail before cascade writes and on dry-run, keep bPath in pre-existing collision tests, and inject the post-CAS dangling-symlink race through testHookBeforeTaskRenameDestinationCreate.

**Resolution:** Confirmed red-first: a pre-existing dangling symlink let live
    rename commit one cascade document and let preview succeed. Target preflight
    now uses Lstat before any cascade; real-write and preview regressions
    preserve source, inbound bytes, and the symlink. The exclusive-create
    symlink test now injects only after preflight so it independently exercises
    O_EXCL. Adjacent creation and lint-relocation hypotheses are explicitly
    recorded in task 6g7wsr8yvt1a, not claimed fixed here.

### Verified consumer inventory

All declarations and line numbers below were verified directly against `/private/tmp/isolated-review.rKZHgW` at baseline commit `aff2bda6ca13fdf8d94bd4a0e6c4356f155893a3`.

| Symbol | Exact source declaration / field access | Producer -> Consumer | Verification command |
| --- | --- | --- | --- |
| `(*FS).RenameTask`, `verifyTaskRenameDestination`, `verifyTaskRenamePath` | `func (s *FS) RenameTask(slug, newTitle string, dryRun bool) (result core.TaskRenameMutationResult, err error)` (`internal/store/rename.go:34`); `func verifyTaskRenameDestination(path, ifVersion string) error` (`:335`); `func verifyTaskRenamePath(path, ifVersion string) error` (`:318`) | `store.FS` (`internal/store/rename.go:161,164`) -> `core.TaskStore` (`internal/core/store.go:80`) | `go test -race -count=3 ./internal/store -run 'TestRenameTask\|TestVerifyTaskRenameDestination'` |
| `createFileAtomicExactMode` | `func createFileAtomicExactMode(path string, data []byte, perm os.FileMode) error` (`internal/store/atomic.go:92`, unchanged) | `store.createFileAtomicWithMode` (`internal/store/atomic.go:96`) -> `(*FS).RenameTask` (`internal/store/rename.go:135`) | `go test -race -count=3 ./internal/store -run 'TestRenameTask_ExclusiveDestinationCreate\|TestFixIDRelocationPreservesModeDespiteUmask'` |
| `TaskRenameMutationResult`, `TaskRenameReceipt`, `TaskRenameFailure` | `type TaskRenameMutationResult struct` (`DestinationWritten bool` at `internal/core/task_rename.go:27`, `SourceRemoved bool` at `:28`); `type TaskRenameReceipt struct` (`DestinationWritten bool` at `:50`, `SourceRemoved bool` at `:51`, `Remedy string` at `:52`); `type TaskRenameFailure struct { Cause error; Receipt TaskRenameReceipt }` (`:58-61`) | `store.FS.RenameTask` -> `core.taskRenameReceipt` (`internal/core/task_rename.go:95`) | `go test -race -count=3 ./internal/core -run 'TestRenameTaskPreservesCauseAndSelectsRecoveryFromDurableActions'` |
| `(*Service).RenameTask`, `taskRenameFailureRemedy` | `func (s *Service) RenameTask(slug, newTitle string, dryRun bool) (TaskRenameReceipt, error)` (`internal/core/service_task.go:1023`); `func taskRenameFailureRemedy(result TaskRenameMutationResult) string` (`internal/core/task_rename.go:106`) | `core.Service` (`internal/core/service_task.go:1026-1028`) -> `cli.newTaskRenameCmd` (`internal/cli/task.go:732-737`) | `go test -race -count=3 ./internal/core -run 'TestRenameTaskPreservesCause'` |
| `taskRenameCommandFailure`, `newTaskRenameCmd` | `type taskRenameCommandFailure struct { cause error; receipt core.TaskRenameReceipt; workspace wire.WorkspaceJSON }` (`internal/cli/task.go:20-24`); `func newTaskRenameCmd(app *App) *cobra.Command` (`:719`) | `cli.newTaskRenameCmd` (`internal/cli/task.go:734-737`) -> `cli.WriteError` (`internal/cli/exit.go:188-198`) | `go test -race -count=3 ./internal/cli -run 'TestTaskRenameCommandRetainsDestinationFailureAndRequiresInspection'` |
| `WriteError`, `filesystemDetails` | `func WriteError(w io.Writer, err error, asJSON bool)` (`internal/cli/exit.go:137`); `if renameErr.receipt.DestinationWritten && payload.Error.Filesystem != nil { payload.Error.Filesystem.Retryable = false }` (`:195-197`); `func filesystemDetails(err error) *wire.FilesystemErrorJSON` (`internal/cli/fserror.go:21`) | `cli.WriteError` -> `wire.ErrorEnvelope` (`Error.TaskRename *TaskRenameRecoveryJSON` at `internal/wire/envelopes.go:1134`, `Error.Filesystem *FilesystemErrorJSON` at `:1139`) | `go test -race -count=3 ./internal/cli -run 'TestTaskRenameFilesystemRetryHintRespectsDestinationCommit\|TestTaskRenameCommandRetainsDestinationFailureAndRequiresInspection'` |
| `TaskRenameJSON`, `TaskRenameRecoveryJSON`, `ToTaskRenameRecoveryJSON`, `SchemaVersion` | `type TaskRenameJSON struct` (`internal/wire/task_rename.go:8`, unchanged); `type TaskRenameRecoveryJSON struct` (`:39`, unchanged); `func ToTaskRenameRecoveryJSON(receipt core.TaskRenameReceipt, workspace WorkspaceJSON) TaskRenameRecoveryJSON` (`:68`, unchanged); `const SchemaVersion = "1.89"` (`internal/wire/wire.go:383`) | `wire.ToTaskRenameRecoveryJSON` -> `cli.WriteError` (`internal/cli/exit.go:190-191`) | `go test -race -count=3 ./internal/cli -run 'TestTaskRenameJSONReportsPreviewAndDurableOutcome\|TestWriteErrorCarriesStructuredTaskRenameRecovery'` |

### Consolidated producer trace and recovery execution

1. **Real-store destination-drift cross-check (`FS.RenameTask` -> `Service.RenameTask`):**
   - In `(*FS).RenameTask` (`internal/store/rename.go:95-173`), inbound-link cascade writes run first (`:95-113`, advancing `AppliedDocuments`, `AppliedLinks`, and `Committed = true`). Next, source CAS (`verifyUnchanged`, `:121`) and `os.Stat(source.path)` (`:124`) precede `createFileAtomicExactMode(plan.newPath, plan.renamedContent, sourceInfo.Mode().Perm())` (`:135`), which sets `result.AppliedDocuments++`, `result.AppliedLinks += plan.targetLinks`, `result.Committed = true`, and `result.DestinationWritten = true` (`:141-144`).
   - When `plan.newPath != source.path` (`:151`), `testHookBeforeTaskRenameSourceRemove` runs (`:152-156`), followed by `verifyTaskRenameDestination(plan.newPath, hashContent(plan.renamedContent))` (`:161`) and `verifyTaskRenamePath(source.path, source.version)` (`:164`) immediately before `os.Remove(source.path)` (`:167`). Neither post-destination verifier re-runs `s.resolve` or `s.resolvePath`, avoiding false duplicate-ID ambiguity while both `<id>-<old>.md` and `<id>-<new>.md` exist.
   - `verifyTaskRenameDestination` (`internal/store/rename.go:335-347`) calls `os.Lstat(path)` (`:336`), maps `os.IsNotExist(err)` to `domain.ErrConflict` (`:338-340`), preserves non-`ENOENT` `*os.PathError` causes (`:341`, verified by `TestVerifyTaskRenameDestinationPreservesFilesystemCause` using a real `syscall.ENOTDIR` parent file at `internal/store/rename_destination_test.go:129-140`), rejects non-regular files (`!info.Mode().IsRegular()`, `:343-345`) with `domain.ErrConflict`, and delegates byte comparison to `verifyTaskRenamePath` (`:346`).
   - Across all 8 subtests of `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt` (`{no-cascade,cascade} x {deleted,edited,directory,symlink}`, `internal/store/rename_destination_test.go:18-127`, driven via the synthetic production hook `testHookBeforeTaskRenameSourceRemove`), `Service.RenameTask` returns `*core.TaskRenameFailure` wrapping `domain.ErrConflict` with `Committed: true`, `DestinationWritten: true`, `SourceRemoved: false`, `Complete: false`, `SourcePath: oldPath`, `DestinationPath: newPath`, exact `AppliedDocuments`/`AppliedLinks` (`1/0` without cascade, `2/1` with cascade), intact `0600` source bytes, and untouched raw destination state.

2. **Core/CLI error receipt and retry hint contrast (`Service.RenameTask` -> `newTaskRenameCmd` -> `WriteError`):**
   - `(*Service).RenameTask` (`internal/core/service_task.go:1023-1031`) invokes `s.store.RenameTask` once (no automatic retry loop) and populates `receipt.Remedy = taskRenameFailureRemedy(result)` whenever `err != nil && result.Committed`.
   - `taskRenameFailureRemedy` (`internal/core/task_rename.go:106-117`) branches on durable state:
     - `result.Complete`: `"inspect the renamed task by stable id before retrying; the planned rename is already complete"`
     - `result.DestinationWritten && !result.SourceRemoved`: `"inspect the old and destination task paths before retrying; compare and reconcile any surviving copies, and remove the old source only after verifying the destination preserves the intended content"`
     - `result.Committed` (pre-destination cascade prefix): `"resolve the reported error, then rerun the same rename by stable task id; the durable link-rewrite prefix is convergent"`
   - Why raw `EBUSY` is retryable while post-destination `EBUSY` is not: `classifyFilesystemError` (`internal/cli/fserror.go:41-59`) classifies `syscall.EAGAIN`, `syscall.EINTR`, and `syscall.EBUSY` as `class: "io", retryable: true` because repeating an uncommitted or idempotent syscall can succeed once contention clears. When a task rename fails during the pre-destination cascade prefix (`Committed: true, DestinationWritten: false`), the old source remains the sole task record and the cascade prefix is convergent, so `WriteError` (`internal/cli/exit.go:188-198`) preserves `payload.Error.Filesystem.Retryable == true`. Once `DestinationWritten == true`, however, both `<id>-<old>.md` and `<id>-<new>.md` may exist on disk (or the rename is already complete), so repeating the command cannot reconcile them and will fail with `ambiguous` (exit 13) or `conflict` (exit 14). `WriteError` (`internal/cli/exit.go:195-197`) therefore overrides `payload.Error.Filesystem.Retryable = false` when `renameErr.receipt.DestinationWritten && payload.Error.Filesystem != nil`. (Note: the CLI `EBUSY` tests use an injected portable `partialDestinationRenameStore` / `taskRenameCommandFailure` fixture, not a kernel-busy file.)

3. **Execution of recovery advice on scratch records:**
   - Executed in an isolated scratch tree outside the repository with `/private/tmp/isolated-review.rKZHgW/bin/tskflwctl` on the two-copy divergent state (`tasks/6fjangd7kva1-old.md` and `tasks/6fjangd7kva1-new-title.md` containing `"Unique raw destination edit."`):
     - Blind retry by stable ID (`tskflwctl -C <scratch> task rename 6fjangd7kva1 "New title" --json`) fails closed with exit `13` (`{"schema_version":"1.89","error":{"code":"ambiguous","message":"\"6fjangd7kva1\" matches 2 tasks by id: new-title (6fjangd7kva1), old (6fjangd7kva1): ambiguous match"}}`) without modifying either file.
     - Blind retry by old slug (`tskflwctl -C <scratch> task rename old "New title" --json`) fails closed with exit `14` (`{"schema_version":"1.89","error":{"code":"conflict","message":"conflict: target filename already exists: 6fjangd7kva1-new-title.md"}}`) without modifying either file.
     - Following the emitted remedy—comparing both files, verifying `6fjangd7kva1-new-title.md` preserves the intended content including `"Unique raw destination edit."`, removing `6fjangd7kva1-old.md`, and running `tskflwctl -C <scratch> task rename 6fjangd7kva1 "New title" --json` and `tskflwctl -C <scratch> lint`—converges cleanly (`complete: true, changed: false, committed: false`, exit `0`, lint clean) with all unique content preserved.

4. **ADR-0008 compatibility check (`SchemaVersion = "1.89"`):**
   - `internal/wire/task_rename.go` is unchanged. No JSON property name, type, requiredness, closed vocabulary, selector, or exit code changed. Suppressing a contradictory `filesystem.retryable: true` hint after `destination_written: true` and refining human `remedy` prose restores the documented meaning of `FilesystemErrorJSON.Retryable` (`internal/wire/envelopes.go:1154-1157`) and `TaskRenameJSON` (`internal/wire/task_rename.go:20-23`). Retaining revision `1.89` without a revision bump is consistent with ADR-0008 §2–3.

### Mutation ledger

All mutations were compiler-valid, executed inside `/private/tmp/isolated-review.rKZHgW`, and restored with `git checkout -- <file>` followed by green `-race -count=3` verification.

| # | Exact edit | Intended test(s) | Outcome | Restored verification |
| --- | --- | --- | --- | --- |
| 1 | `internal/cli/exit.go:195-197`: remove `if renameErr.receipt.DestinationWritten && payload.Error.Filesystem != nil { payload.Error.Filesystem.Retryable = false }` (primary assigned mutation) | `TestTaskRenameCommandRetainsDestinationFailureAndRequiresInspection` (`transient/json`), `TestTaskRenameFilesystemRetryHintRespectsDestinationCommit` | **Compiled & Killed**: `transient/json` failed at `task_rename_test.go:218` and `destination_written=true` failed at `task_rename_test.go:249` (`"retryable":true`) | Restored; `go test -race -count=3 ./internal/cli -run 'TestTaskRename\|TestWriteErrorCarriesStructuredTaskRenameRecovery'` passed |
| 2 | `internal/cli/exit.go:195`: change `renameErr.receipt.DestinationWritten` to `renameErr.receipt.Committed` (blanket no-retry contrast probe) | `TestTaskRenameFilesystemRetryHintRespectsDestinationCommit` (`written=false`) | **Compiled & Killed**: failed at `task_rename_test.go:249` (`destination_written=false gave incorrect retry hint: ... "retryable":false`) | Restored; `go test -race -count=3 ./internal/cli -run 'TestTaskRename\|TestWriteErrorCarriesStructuredTaskRenameRecovery'` passed |
| 3 | `internal/core/task_rename.go:110-111`: remove `case result.DestinationWritten && !result.SourceRemoved:` in `taskRenameFailureRemedy` | `TestRenameTaskPreservesCauseAndSelectsRecoveryFromDurableActions`, `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt`, `TestTaskRenameCommandRetainsDestinationFailureAndRequiresInspection` | **Compiled & Killed**: all three tests failed because destination-written failures fell through to the convergent cascade-prefix remedy | Restored; focused `-race -count=3` in `./internal/core`, `./internal/store`, `./internal/cli` passed |
| 4 | `internal/store/rename.go:161-163`: remove call to `verifyTaskRenameDestination(plan.newPath, hashContent(plan.renamedContent))` while retaining `verifyTaskRenamePath(source.path, source.version)` | `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt` | **Compiled & Killed**: all 8 subtests (`{no-cascade,cascade} x {deleted,edited,directory,symlink}`) failed at `rename_destination_test.go:79` (`Complete:true SourceRemoved:true err=<nil>`) | Restored; `go test -race -count=3 ./internal/store -run 'TestRenameTask\|TestVerifyTaskRenameDestination'` passed |
| 5 | `internal/store/rename.go:343-345`: remove `if !info.Mode().IsRegular()` in `verifyTaskRenameDestination` | `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt` (`symlink` and `directory` subtests) | **Compiled & Killed**: `no-cascade/symlink` and `cascade/symlink` failed at `rename_destination_test.go:79` (`Complete:true SourceRemoved:true err=<nil>`); `directory` subtests failed at `:82` (`is a directory` instead of `ErrConflict`) | Restored; `go test -race -count=3 ./internal/store -run 'TestRenameTask\|TestVerifyTaskRenameDestination'` passed |
| 6 | `internal/store/rename_test.go:476-478`: remove `os.Remove(bPath)` in `TestRenameTask_ExclusiveDestinationCreatePreservesDanglingSymlink` (hostile fixture probe for `M1`) | `TestRenameTask_ExclusiveDestinationCreatePreservesDanglingSymlink` | **Compiled & Killed (reproduced `M1`)**: failed at `rename_test.go:483` with `Committed:true AppliedDocuments:1 AppliedLinks:1 DestinationWritten:false` because `ensureTaskRenameTargetAvailable` used `os.Stat` | Restored; `git checkout -- internal/store/rename_test.go` |
| 7 | `internal/store/rename.go:135`: replace `createFileAtomicExactMode` with `createFileAtomic` | `TestRenameTask_PreservesSourceFileMode`, `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt` | **Compiled & Survived `TestRenameTask*`**: rename mode tests use `0o600` (`0o600 &^ 0o022 == 0o600`); helper itself is pinned with `0o664` under `umask 0o022` in `TestFixIDRelocationPreservesModeDespiteUmask` (`internal/store/fix_relocation_unix_test.go:18`) | Restored; `git checkout -- internal/store/rename.go` |

### Surviving-hypothesis and coverage ledger

- **Pre-existing dangling symlink at rename target (`M1`, reproduced defect):** `(*FS).ensureTaskRenameTargetAvailable` (`internal/store/rename.go:308`) uses `os.Stat` rather than `os.Lstat`, allowing a dangling symlink at `newPath` to pass pre-flight checks on `--dry-run` and to commit inbound-link cascade writes before `createFileAtomicExactMode` (`O_EXCL`) rejects the target; `TestRenameTask_ExclusiveDestinationCreatePreservesDanglingSymlink` (`internal/store/rename_test.go:476`) masked this by deleting `bPath`. Filed as open finding `M1`.
- **Synthetic vs real fault boundaries (coverage limits, not filed as defects):**
  - `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt` (`internal/store/rename_destination_test.go:41`) and `TestRenameTask_DestinationWrittenCleanupFailureRequiresInspection` (`internal/store/rename_test.go:614`) inject via `testHookBeforeTaskRenameSourceRemove` (`internal/store/rename.go:152`), which fires before `verifyTaskRenameDestination`, `verifyTaskRenamePath`, and `os.Remove(source.path)`. Neither test injects a kernel failure at `os.Remove(source.path)` (`internal/store/rename.go:167`) itself.
  - `TestVerifyTaskRenameDestinationPreservesFilesystemCause` (`internal/store/rename_destination_test.go:129-140`) exercises a real OS `syscall.ENOTDIR` `*os.PathError` on `os.Lstat` by placing a regular file as the parent path component.
  - `TestTaskRenameCommandRetainsDestinationFailureAndRequiresInspection` (`internal/cli/task_rename_test.go:164-226`) and `TestTaskRenameFilesystemRetryHintRespectsDestinationCommit` (`:228-252`) use a stub `partialDestinationRenameStore` and synthetic `&os.PathError{Op: "read", Path: ..., Err: syscall.EBUSY}` rather than a real kernel `EBUSY`.
  - `RenameTask` mode tests (`internal/store/rename_test.go:75`, `internal/store/rename_destination_test.go:33`) use `0o600`, which is invariant under `umask 0o022` (Mutation #7); `createFileAtomicExactMode` itself is covered under explicit `umask 0o022` with `0o664` in `TestFixIDRelocationPreservesModeDespiteUmask` (`internal/store/fix_relocation_unix_test.go:18`).
- **Final check-to-unlink race window (by design):** A raw writer modifying `plan.newPath` or `source.path` after `verifyTaskRenameDestination` / `verifyTaskRenamePath` (`internal/store/rename.go:161,164`) and before `os.Remove(source.path)` (`:167`) cannot be prevented without OS transactions; explicitly documented in `internal/store/rename.go:32-33` and `docs/architecture/persistence.md:103`.
- **Already tracked adjacent work (not reopened as separate findings):**
  - In-place `FixFrontmatter` CAS and checked guard release: tracked in `planning/tasks/6gmdsxccqqnm-guard-in-place-lint-repairs-and-retain-guard-finalization-failures.md`.
  - `createFileAtomicWithMode` error-path `os.Remove(path)` cleanup race, shared durability/retry policy across remaining CLI recovery envelopes in `internal/cli/exit.go:165-218`, and adjacent `os.Stat`-before-`O_EXCL` / `verifyFixPath` checks (`internal/store/fix.go:213`, `internal/store/fix_relocation.go:58,83`, `internal/store/threadcreation.go:135`, `internal/store/create.go:85`, `internal/store/lifecyclemutation.go:255`): tracked in `planning/tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md` (and cross-referenced in `M1`).

### Validation and checks not run

- **Executed in sandbox `/private/tmp/isolated-review.rKZHgW`:**
  - `go build -o bin/tskflwctl ./cmd/tskflwctl`: exit 0.
  - `go test -race -count=3 ./internal/store -run 'TestRenameTask|TestVerifyTaskRenameDestination'`: pass (`2.133s`).
  - `go test -race -count=3 ./internal/core ./internal/cli -run 'TestRenameTaskPreservesCause|TestTaskRename|TestWriteErrorCarriesStructuredTaskRenameRecovery'`: pass (`1.320s` / `1.491s`).
  - `./bin/tskflwctl -C /private/tmp/isolated-review.rKZHgW lint --links`: pass (`✔ all planning entities and dependency links pass lint`).
  - `./bin/tskflwctl -C /private/tmp/isolated-review.rKZHgW audit lint`: pass (`✔ all audit findings pass lint`).
  - Scratch recovery probe in `/Users/andyeschbacher/.gemini/antigravity-cli/brain/04b71713-90b2-40c9-8b68-5528e77deb7c/scratch/reconcile-probe` (cleaned up after execution): verified exit 13 on ID retry, exit 14 on old-slug retry, and exit 0 after reconciling copies.
- **Inherited from owner checkpoint (not re-run across the whole repo):** full-repository `just test`, `just lint`, `just docs-check`, and schema-comment generation.

### Isolation attestation

Output of `/Users/andyeschbacher/git/andy-esch/taskflow/scripts/isolated-review-workspace.sh verify --sandbox /private/tmp/isolated-review.rKZHgW` after restoring all mutation probes and removing scratch files:

```text
sandbox_path=/private/tmp/isolated-review.rKZHgW
git_dir=/private/tmp/isolated-review.rKZHgW/.git
baseline_commit=aff2bda6ca13fdf8d94bd4a0e6c4356f155893a3
source_blob=f461d52cb074382f32fff23e1c725b59cd799b27
source_fingerprint=feb712cfa5ad2688ee1a0499c2dd01aa9de64bb3
deliverable=planning/audits/6gmke6f7nnw0-2026-10-17-task-rename-destination-safety-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner disposition (2026-10-17)

The guarded handoff receipt subsequently confirmed `verify=passed`,
`transfer=succeeded`, and `delivery=confirmed` for baseline `aff2bda6ca13fdf8d94bd4a0e6c4356f155893a3`,
with transferred blob `14bbeb69d85fa12e149dfa7ccfc467a324cf0f31`.
The report above remains a review of that baseline, not of these later owner edits.

Accepted M1 and reproduced both live-prefix and preview failures with the retained
inbound-link fixture. Target preflight now uses `Lstat`; both operations refuse
before effects. Moved the exclusive-create symlink injection to the final-create
hook so preflight cannot mask a broken `O_EXCL` guarantee. Also closed mutation
#7's caller-coverage gap with `TestRenameTask_PreservesModeDespiteUmask`: substituting
ordinary creation now fails at 0644 versus the required 0664 under explicit 022.
The mutation was restored before validation.

The adjacent-path tracking claim above was broader than the planning text actually
supported. Added explicit source-inspection investigation items to task
`6g7wsr8yvt1a` for remaining `Stat` preflights and lint relocation's symlink-following
verifier; these are not represented as reproduced or fixed by this rename change.
No crash-transaction or final raw-writer race guarantee was added.

Owner validation after fixes: `just build`, full race `just test`, `just lint`
(0 issues), and `just docs-check` pass. Claude's separate audit remains pending
and untouched; task completion is not inferred from this audit's disposition.
