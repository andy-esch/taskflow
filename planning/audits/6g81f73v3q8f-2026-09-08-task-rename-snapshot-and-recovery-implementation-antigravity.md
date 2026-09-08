---
schema: 1
id: 6g81f73v3q8f
bucket: closed
area: task-rename-snapshot-and-recovery-implementation-antigravity
date: "2026-09-08"
updated_at: "2026-09-08"
---
# Audit: Task rename snapshot and recovery implementation — antigravity — 2026-09-08

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Adversarially review the implementation of the task `guard-renametask-against-stale-cascade-plans-and-concurrent-identity-duplication`. Do not merely confirm that tests pass. Try to demonstrate a stale plan, lost update, duplicate stable identity, clobbered target, unsafe retry recommendation, inaccurate durable-prefix receipt, adapter leak, or recovery state from which the documented advice cannot converge. Treat the implementation as suspect until each invariant survives a concrete hostile probe.

## Review target

Review the complete uncommitted implementation snapshot copied into your mandatory sandbox, especially:

- `planning/tasks/6g7wxs43g7nh-guard-renametask-against-stale-cascade-plans-and-concurrent-identity-duplication.md`
- the 2026-09-08 amendment in `planning/adrs/0003-stable-key-id-addressed-storage.md`
- the `RenameTask` architecture contract in `docs/ARCHITECTURE.md`
- `internal/store/rename.go` and `internal/store/rename_test.go`
- `internal/core/task_rename.go`, `internal/core/service_task.go`, and the affected Store port
- `internal/cli/task.go`, `internal/cli/exit.go`, and `internal/cli/task_rename_test.go`
- `internal/wire/task_rename.go`, the error-envelope/schema changes, schema comments, and CLI goldens

Build a consumer inventory for the changed Store and Service signatures and for `TaskRenameReceipt` / `TaskRenameFailure`. Verify every production caller, test double, CLI adapter, prospective TUI-facing boundary, wire mapper, schema registration, and error-classification path. Separate existing implemented consumers from future capabilities mentioned only in planning.

## Intended contract to challenge

A real rename captures the selected source version before waiting, then takes the canonical planning-root repository guard. Once guarded, it rechecks target availability and the original source, plans the link cascade from fresh current bytes, content-CAS checks each cascade document immediately before atomic replacement, content-CAS checks the source before destination materialization, creates a moved destination without clobbering, content-CAS checks the old source again immediately before removal, and releases the guard. Two cooperating concurrent renames cannot both commit; cooperating guarded edits are incorporated or rejected without loss.

The physical order is cascade documents, destination content, then old-source deletion. Failure before destination creation may leave a convergent link-rewrite prefix and advises rerunning the same rename by stable id after resolving the cause. Failure after destination creation but before source removal deliberately retains both files and requires inspection rather than a blind retry. Failure while releasing the guard after completion reports the rename as already complete. Store and core preserve planned/applied documents and links plus destination/source milestones; CLI JSON failures expose the adapter-neutral receipt. Dry-run is a non-durable preview and claims no reservation. Raw editors remain outside the advisory lock, but CAS/O_EXCL checks must not make stronger claims than they can prove.

## Mandatory evidence floor

Run focused tests with `-race` and repeat the coordinated concurrency cases enough to expose nondeterminism. Run the full validation that the implementation claims. Inspect the repository guard, path resolution, CAS, atomic create/replace, error classification, and JSON-schema machinery rather than assuming their names imply the required behavior.

For every new concurrency or recovery regression test, temporarily kill the exact production guard it claims to pin and prove that the named test fails for the intended reason. At minimum challenge: capture-before-lock source validation; the post-lock target recheck; per-cascade CAS; exclusive destination creation; pre-destination source validation; final pre-removal source CAS; and committed/unlock failure wrapping. Restore the sandbox baseline after every mutation. A mutation killed only by an unrelated compile failure, broad fixture failure, or another invariant is not evidence.

Exercise the new optional `error.task_rename` wire branch with non-default values and validate it against the generated JSON Schema. Verify the schema minor bump propagated to all emitted envelopes and that non-rename failures do not acquire a misleading rename receipt. Demonstrate the human and machine error behavior for at least one retryable durable prefix and the destination-written/source-retained state; do not accept prose inspection alone when a structured assertion is feasible.

## Required hostile angles

1. Coordinate two FS instances and, if practical, two processes renaming the same source to both different and identical targets. Challenge stable-id uniqueness, loser attribution, guard identity, and any stale target/source window.
2. Coordinate guarded task body/field changes to the source and to inbound-link documents at each meaningful point around snapshot, lock acquisition, plan, CAS, write, destination create, and source removal. A successful cooperating write must not disappear.
3. Simulate raw creation, modification, deletion, and relocation of the source, destination, and early/late cascade files. Identify precisely which windows are detected and which cannot be protected; flag documentation that overstates the boundary.
4. Force failures before the first write, after one or several cascade writes, during destination creation, immediately after destination write, during source CAS/removal, and during unlock. Compare physical disk state with every receipt field, error class, exit code, remedy, and safe next action.
5. Actually execute or faithfully simulate each recommended recovery. Prove a retryable prefix converges without double-rewriting or loss. Prove the inspection-only state does not recommend an operation that resolution ambiguity makes impossible or that could delete a newer raw edit.
6. Challenge the decision to write links before the destination exists. Consider interruption, process death, malformed destination files, links in the rename source itself, same-slug/title-only renames, duplicate or ambiguous identities, and a second competing rename after a partial prefix.
7. Audit deterministic planning and accounting: walk order, planned/applied document counts, link counts, source/destination milestones, exact-target collision behavior, self/reference/fenced links, unknown Markdown documents, modes/permissions, and same-path behavior.
8. Examine architectural fit. Determine whether core receipt/failure types are presentation-neutral and sufficient for CLI, TUI, and a future web/backend adapter, while filesystem-specific locking/CAS stays in the adapter. Flag both leaked filesystem assumptions and insufficient portable recovery information.
9. Look for systemic duplication with task lifecycle, graph mutation/repair, Thread apply, and entity creation. Do not demand an out-of-scope universal transaction framework without demonstrated value, but identify a concrete shared primitive if this implementation repeats a hazardous pattern inconsistently.
10. Perform a second pass focused on paths happy-path tests obscure: deferred named returns, joined unlock errors, an uncommitted conflict plus unlock failure, partial write counts when the source has self-links, raw edits after CAS but before replace/remove, crashes during exclusive create, symlinks or non-regular Markdown entries, and malformed/ambiguous task trees.

Reject speculative findings. For every finding, provide a minimal reproduction or exact code path, impact, why current tests miss it, and the smallest sound correction. If an issue is genuinely adjacent or out of scope, still record it and say why it belongs in a follow-up rather than silently omitting it.

## Validation and restoration

Perform every mutation and generated-artifact check only inside the isolated sandbox. Before delivery, restore all probes and generated drift to the sandbox baseline so the assigned audit is the only changed file. Run `git diff --check`, focused race tests, full `go test -race ./...`, planning lint, Go lint, module-tidy check, CLI docs drift check, schema-comment freshness, schema validation, and golden checks where the environment supports them. Report skipped checks and the exact blocker; do not call them passing.

## Deliverable

Update only your assigned audit. Keep this brief intact. Add findings using the repository grammar, with every new finding left `open` for implementation-owner triage. Include severity, component, effort, urgency, exact evidence, a concrete recommendation, the mutation/probe performed, and affected acceptance criteria. If no finding survives both passes, write a substantive no-findings report listing the invariants challenged, exact mutations killed, commands run, residual risks, and why those risks are acceptable or already scoped.

### Executive verdict

**Ready after fixes.**

The implementation in `internal/store/rename.go`, `internal/core/task_rename.go`, `internal/core/service_task.go`, `internal/cli/task.go`, `internal/cli/exit.go`, and `internal/wire/task_rename.go` resolves the concurrency and cascade hazards identified in task `6g7wxs43g7nh` and ADR-0003.

`RenameTask` now captures the caller's source version before waiting, acquires the canonical planning-root repository guard (`checkedWriteLock`), rechecks target availability under lock (preventing collision diagnostics from degrading into `ErrAmbiguous`), re-verifies the pre-lock source version via `verifyUnchanged`, and recompiles the inbound-link cascade from the fresh guarded tree. Cascade document replacements commit first and are guarded by individual content CAS (`verifyTaskRenamePath`). Destination materialization uses exclusive `O_CREATE|O_EXCL` (`createFileAtomic`) preceded by an immediate source CAS (`verifyUnchanged`). Old source deletion occurs last, preceded by a final pre-removal content CAS (`verifyTaskRenamePath`).

Failure semantics are explicit, deterministic, and typed:
1. Failures before destination creation leave a convergent durable prefix and advise rerunning the same rename by stable task ID.
2. Failures after destination creation retain both files on disk without clobbering, advising inspection and explicit removal of the old source.
3. Unlock failures after full completion advise that the rename is already complete rather than suggesting an unsafe retry.
4. Uncommitted failures return bare domain errors and do not acquire a misleading recovery receipt.
5. Structured receipts are mapped through `wire.TaskRenameRecoveryJSON` under `error.task_rename` in envelope schema 1.62.

One low-severity test-strength finding (`L1`) is left open for owner triage: while the production safeguards for pre-destination source CAS (`internal/store/rename.go:111`) and exclusive destination creation (`internal/store/rename.go:118`) are sound, the existing test suite does not pin their removal—mutating either line leaves all 15 tests in `rename_test.go` passing.

---

### Isolation, verification, and transfer attestation

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.W00hzO
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.W00hzO/.git
baseline_commit=cc2f6515ebdd8015501a6ddab758ad8345709f1e
source_blob=41f5c9b62db15459fb8ac6fae86efe32c62470b8
source_fingerprint=54434e430af528823cffe718e78d126dd9b468c5
deliverable=planning/audits/6g81f73v3q8f-2026-09-08-task-rename-snapshot-and-recovery-implementation-antigravity.md
deliverable_changed=true
transfer=succeeded
```

- **Workspace Independence**: The sandbox was initialized via `scripts/isolated-review-workspace.sh create --print-path`. It is an independent `--no-hardlinks` clone with no alternates file (`.git/objects/info/alternates` is empty), exactly one registered worktree, and `core.worktree` unset.
- **Protocol Adherence**: All builds, test runs, hostile probes, mutations, linting, doc drift checks, and report editing were conducted strictly inside `$SANDBOX`. `$SOURCE_ROOT` remained untouched as a read-only source.
- **Restoration**: All temporary mutation probes, test scaffolding, and compiled binaries were restored to the baseline commit `cc2f6515ebdd8015501a6ddab758ad8345709f1e` prior to verification and transfer.

---

### Acceptance-criteria traceability

| Acceptance criterion | Status | Evidence & verification |
|---|---|---|
| **AC 1:** Concurrent renames of one task to different slugs cannot both commit; disk state contains one stable-ID owner and loser receives attributable conflict. | **Met** | `internal/store/rename.go:74` verifies pre-lock version under lock. Verified by `TestRenameTask_ConcurrentRenamesCommitOneSourceSnapshot` (10 repetitions under `-race`), multi-FS symlink probe, and two concurrent OS CLI processes (`exit 0` vs `exit 14` conflict, exactly 1 file on disk). |
| **AC 2:** Concurrent guarded task field/body mutation is either included by the rename plan or causes the rename to conflict; no successful cooperating write is silently lost. | **Met** | `internal/store/rename.go:77-80` replans cascade from fresh tree under lock. Verified by `TestRenameTask_ReplansAfterCooperatingCascadeDocumentWrite` (`SetFields` on cascade doc preserved) and hostile probe verifying source edit while waiting for lock trips line 74 `verifyUnchanged` without writing cascade docs. |
| **AC 3:** The target filename is rechecked after acquiring the authoritative guard (or protected by an equivalent CAS). | **Met** | `internal/store/rename.go:68` checks `ensureTaskRenameTargetAvailable` under guard, preserving conflict diagnosis. `createFileAtomic` (`:118`) provides `O_EXCL` backstop. Verified by `TestRenameTask_RechecksTargetAfterRepositoryGuard` and identical-target concurrency probe. (See L1 for test pinning gap). |
| **AC 4:** Every cascade file write is authorized against the bytes used to plan it. | **Met** | `internal/store/rename.go:89-91` enforces `verifyTaskRenamePath(edit.path, edit.ifVersion)`. Verified by `TestRenameTask_CASCatchesRawCascadeDocumentEdit` (mutation of cascade doc before write yields `domain.ErrConflict` with `Committed: false`). |
| **AC 5:** Partial multi-file failure semantics and operator recovery diagnostics are explicit and tested. | **Met** | `internal/core/task_rename.go:71-88` implements typed remedy derivation. Verified by `TestRenameTask_PartialCascadeReceiptIsResumable`, `TestRenameTask_DestinationWrittenCleanupFailureRequiresInspection`, `TestRenameTask_CompleteUnlockFailureIsNotRetryable`, and `TestWriteErrorCarriesStructuredTaskRenameRecovery`. |
| **AC 6:** Existing link-cascade, no-op, dry-run, and exact-target collision behavior remains covered. | **Met** | Verified by `TestRenameTask_RenamesAndCascades`, `TestRenameTask_CrossDirSameNameLeftAlone`, `TestRenameTask_DryRunTouchesNothing`, `TestRenameTask_SameSlugDoesNotCascadeOrRemoveSource`, `TestRenameTask_EmptyTitleRejected`, `TestRenameTask_TargetCollisionRefused`, and `TestRenameTask_RefStyleAndFencedExamples`. |
| **AC 7:** Race-enabled focused tests and the full validation suite pass. | **Met** | `go test -race ./internal/store -run 'TestRenameTask' -count=20` (passed, 4.59s), `go test -race ./internal/cli -run 'Rename' -count=20` (passed, 1.61s), full `go test -race ./...` (passed), `golangci-lint run ./...` (0 issues), `just docs-check`, `just tidy-check`, `tskflwctl lint`, `tskflwctl audit lint` (all clean). |

---

### Consumer and contract inventory

#### 1. Store port and implementation
- `internal/core/store.go:52`: Signature `RenameTask(slug, newTitle string, dryRun bool) (TaskRenameMutationResult, error)`. Returns store mutation result containing planned/applied document and link tallies, milestone booleans, and reloaded task.
- `internal/store/rename.go:31`: Implementation on `*FS`.
  - Captures `taskRenameSource` (`path`, `id`, `oldSlug`, `version`) before lock (`:36`).
  - Evaluates non-reserving dry-run planning (`:45-51`).
  - Acquires `checkedWriteLock()` (`:55-63`) which binds process mutex and OS file lock.
  - Rechecks target under lock (`:68`).
  - Rechecks source pre-lock snapshot under lock via `verifyUnchanged` (`:74`).
  - Compiles `taskRenamePlan` from current filesystem state (`:77`).
  - Writes cascade edits with per-document CAS `verifyTaskRenamePath` (`:85-103`).
  - Verifies source CAS immediately before destination creation (`:111`).
  - Creates destination via `createFileAtomic` (`:118`) or in-place atomic rewrite (`:115`).
  - Verifies source CAS before deletion (`:143`) and removes old source (`:146`).

#### 2. Service port and domain adapter
- `internal/core/service_task.go:678`: Signature `(s *Service) RenameTask(slug, newTitle string, dryRun bool) (TaskRenameReceipt, error)`.
- Translates `TaskRenameMutationResult` to adapter-neutral `TaskRenameReceipt` (`:680`).
- If `err != nil && result.Committed`, wraps error into `&TaskRenameFailure{Cause: err, Receipt: receipt}` (`:681-683`), preserving partial durability evidence. Uncommitted failures return bare `err`.

#### 3. Core types
- `internal/core/task_rename.go:12-25`: `TaskRenameMutationResult` (store-internal execution result).
- `internal/core/task_rename.go:30-44`: `TaskRenameReceipt` (adapter-neutral public DTO).
- `internal/core/task_rename.go:49-69`: `TaskRenameFailure` (error wrapper with `Unwrap() error`).
- `internal/core/task_rename.go:71-88`: `taskRenameReceipt(result)` helper setting semantic `Remedy` based on milestones (`Complete`, `DestinationWritten && !SourceRemoved`, `Committed`).

#### 4. Primary CLI adapter & error envelope
- `internal/cli/task.go:697-733`: `newTaskRenameCmd(app)` Cobra command handler.
- `internal/cli/task.go:20-27`: `taskRenameCommandFailure` internal failure carrier holding `cause`, `receipt`, and `workspace`.
- `internal/cli/exit.go:99-103`: In `WriteError`, unmarshals `*taskRenameCommandFailure` into `payload.Error.TaskRename = &details` via `wire.ToTaskRenameRecoveryJSON`.
- Exit codes:
  - 0: Success.
  - 10: `ErrNotFound` (source entity does not exist).
  - 11: `ErrValidation` (empty slug from title).
  - 13: `ErrAmbiguous` (source or colliding target resolves ambiguously).
  - 14: `ErrConflict` (concurrency collision, target exists, or CAS mismatch).
  - 1: Generic I/O or filesystem error.

#### 5. Wire contract and JSON Schema
- `internal/wire/task_rename.go:7-34`: `TaskRenameRecoveryJSON` and mapper `ToTaskRenameRecoveryJSON`.
- `internal/wire/envelopes.go:1046`: `ErrorItem.TaskRename *TaskRenameRecoveryJSON` (`json:"task_rename,omitempty"`).
- `internal/wire/wire.go:261`: `SchemaVersion = "1.62"`.
- `internal/wire/schema_comments.json:208`: Registered doc comment for `TaskRenameRecoveryJSON`.
- `internal/cli/testdata/golden/schema_jsonschema.golden:1213, 3410-3478`: Draft 2020-12 schema definition with all 15 required fields.
- Golden files: 33 CLI golden fixtures updated to `1.62`.

#### 6. Test doubles and test callers
- `internal/core/service_epic_test.go:38`: `mockStore` test double implements stub `RenameTask`.
- `internal/store/rename_test.go`: Store integration tests.
- `internal/cli/task_rename_test.go`: CLI error formatting and recovery payload test.
- `internal/wire/envelopes_test.go:357`: Wire `ErrorEnvelope` schema validation test.

#### 7. Prospective / planned consumers
- Prospective TUI / Web adapters: `TaskRenameReceipt` and `TaskRenameRecoveryJSON` contain no filesystem paths or CLI formatting types; future adapters can inspect `AppliedDocuments`, `Committed`, `Complete`, and `Remedy` directly. No TUI-facing rename verb currently exists in `internal/tui`.

---

### Findings

#### L1. Pre-destination source CAS and exclusive destination creation lack regression tests pinning their removal · **Status:** fixed

**File:** [internal/store/rename.go:111-123](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.W00hzO/internal/store/rename.go#L111-L123) | **Component:** store
**Severity:** low · **Effort:** XS · **Urgency:** eventually

While `RenameTask` correctly implements an immediate pre-destination source CAS check (`verifyUnchanged` at `internal/store/rename.go:111`) and exclusive target materialization (`createFileAtomic` at `internal/store/rename.go:118`), neither guard is pinned by the existing regression test suite in `internal/store/rename_test.go`.

**Exact evidence:**
1. Commenting out lines 111-113 in `internal/store/rename.go`:
   ```go
   // if err := verifyUnchanged(s.resolvePath, source.id, source.path, source.version, "task", "rename"); err != nil {
   //     return result, err
   // }
   ```
   Executing `go test -v ./internal/store -run 'TestRenameTask'` results in **15/15 tests passing** (0 failures). Existing test `TestRenameTask_CASCatchesRawCascadeDocumentEdit` hooks only `bPath` (the cascade document), and `TestRenameTask_SourceRemovalCASCatchesRawEdit` hooks `testHookBeforeTaskRenameSourceRemove` (which executes *after* destination creation). No test modifies `source.path` during the cascade write phase to verify line 111 rejects the stale destination write.
2. Replacing `createFileAtomic` with `writeFileAtomic` at line 118:
   ```go
   } else if err := writeFileAtomic(plan.newPath, plan.renamedContent, 0o644); err != nil {
   ```
   Executing `go test -v ./internal/store -run 'TestRenameTask'` results in **15/15 tests passing** (0 failures). Existing target collision tests (`TestRenameTask_TargetCollisionRefused` and `TestRenameTask_RechecksTargetAfterRepositoryGuard`) create colliding targets prior to or during lock acquisition, both of which are caught upstream by line 68 `ensureTaskRenameTargetAvailable`. No test injects a target file creation *after* line 68 to prove that `createFileAtomic` fails closed with `O_EXCL` instead of clobbering.

**Recommendation:**
Add two targeted regression tests in `internal/store/rename_test.go`:
1. Hook `testHookBeforeTaskRenameWrite` for `path == source.path`, mutate `source.path`, and assert that `RenameTask` returns `domain.ErrConflict` with `DestinationWritten: false`.
2. Introduce a hook immediately prior to destination materialization (or simulate a race), write `plan.newPath`, and assert that `createFileAtomic` rejects the write with `domain.ErrConflict: target filename already exists` without clobbering.

**Affected acceptance criteria:** AC 3, AC 7.

---

**Resolution:** Added deterministic pre-destination source-CAS, post-CAS
target-creation, and dangling-symlink target regressions; mutation checks prove
the source and exclusive-create guards fail their specific tests when removed.

### Hostile angle analysis and empirical evidence

#### Angle 1: Two FS instances and cross-process concurrency
- **Different targets:** Verified by `TestRenameTask_ConcurrentRenamesCommitOneSourceSnapshot`. Two goroutines renaming `old` to `Alpha title` and `Beta title` yielded exactly 1 success, 1 `domain.ErrConflict`, and exactly 1 file matching `tasks/6fjangd7kva1-*.md`.
- **Identical targets:** Probed with hostile test `TestHostileProbe_ConcurrentRenamesIdenticalTarget` (two concurrent renames to `Identical title`). Exactly 1 succeeded, and the second was rejected at `internal/store/rename.go:68` with `domain.ErrConflict: target filename already exists: 6fjangd7kva1-identical-title.md` (`Committed: false`). Exactly 1 file remained on disk.
- **Symlinked root:** Probed with `TestHostileProbe_MultiFS_SymlinkRoot_ConcurrentRenames` using two `FS` instances where one was initialized via an `os.Symlink` path. `normalizeRepositoryLockKey` (`internal/store/lock.go:37-49`) resolves symlinks via `filepath.EvalSymlinks`, ensuring both instances bound to the identical mutex and OS file lock. 10/10 iterations under `-race` produced 1 winner and 1 conflict.
- **Cross-process OS probe:** Executed two concurrent `tskflwctl task rename initial-task ... --json` OS processes in a test repository. Result: `code1=14 code2=0`. Winner wrote `6g81kf3cnsdn-beta-title.md`; loser failed with exit code 14 (`conflict`) and uncommitted error payload. Exactly 1 file on disk.

#### Angle 2: Cooperating writes to source and cascade documents
- **Guarded edit to cascade doc:** `TestRenameTask_ReplansAfterCooperatingCascadeDocumentWrite` verified that when `SetFields` updates task B while RenameTask waits for the lock, RenameTask incorporates task B's updated `priority: low` upon acquiring the guard and repoints its link without lost updates.
- **Guarded edit to source doc:** Probed with `TestHostileProbe_GuardedSourceEditWhileWaitingLock`. When `SetFields` mutates the source task while RenameTask is queued, RenameTask acquires the guard, detects the version mismatch at line 74 `verifyUnchanged`, and immediately aborts with `domain.ErrConflict` (`Committed: false`). Zero cascade documents were touched; task priority remained `low`.

#### Angle 3: Raw creation, modification, deletion, and relocation
- **Raw cascade edit:** `TestRenameTask_CASCatchesRawCascadeDocumentEdit` proved that editing a cascade document after plan generation trips line 89 `verifyTaskRenamePath`, failing closed with `domain.ErrConflict`.
- **Raw source edit before destination write:** Line 111 `verifyUnchanged` detects modifications to `source.path` prior to creating `newPath`.
- **Raw source edit after destination write:** `TestRenameTask_SourceRemovalCASCatchesRawEdit` verified that modifying the old source after destination creation trips line 143 `verifyTaskRenamePath`, retaining both files for inspection.
- **Boundary honesty:** POSIX user-space advisory locking and atomic rename (`writeFileAtomic`/`createFileAtomic`) cannot eliminate the kernel-level microsecond TOCTOU window between CAS `os.ReadFile` and `os.Rename`/`os.Remove`. ADR-0003 and `docs/ARCHITECTURE.md` explicitly qualify this boundary ("where the filesystem permits detection", "bounds the remaining raw-editor window").

#### Angle 4: Failures across the execution lifecycle
The disk state and receipt fields were verified across all failure milestones:
1. *Pre-write failure* (e.g. target conflict or source version mismatch): `Committed: false, Complete: false, DestinationWritten: false, SourceRemoved: false`. Error is unwrapped domain error.
2. *Mid-cascade failure* (interruption after document 1 of 3): `Committed: true, AppliedDocuments: 1, PlannedDocuments: 3, DestinationWritten: false, SourceRemoved: false`. Remedy: "rerun the same rename by stable task id; the durable link-rewrite prefix is convergent".
3. *Destination creation failure* (target conflict): `Committed: true` (if cascade docs written), `DestinationWritten: false`.
4. *Source removal failure* (permission or CAS error): `Committed: true, DestinationWritten: true, SourceRemoved: false`. Remedy: "inspect the old and destination task files before retrying; if the destination is correct, remove the retained old source to restore one stable-id owner".
5. *Unlock failure after completion*: `Committed: true, Complete: true, DestinationWritten: true, SourceRemoved: true`. Remedy: "inspect the renamed task by stable id before retrying; the planned rename is already complete".

#### Angle 5: Recovery convergence
- **Retryable prefix:** Verified by `TestRenameTask_PartialCascadeReceiptIsResumable` and `TestHostileProbe_SelfLinksAccounting`. Re-running the rename by stable ID found already-repointed documents untouched (`n == 0`), repointed remaining files, created the destination, and deleted the old source.
- **Inspection-only state:** Verified by `TestRenameTask_DestinationWrittenCleanupFailureRequiresInspection`. The tool does not attempt automatic rollback or blind deletion. Removing the old source restores single-owner health.

#### Angle 6: Cascade-first ordering decisions
- **Interruption / crash:** Leaves intact old source at original path, keeping the stable ID resolvable.
- **Competing rename after partial prefix:** If a user attempts to rename to a different third slug after a partial prefix, rewritten documents point to the abandoned second slug. Remedy explicitly instructs to "rerun the SAME rename by stable task id".
- **Self-links in source:** Verified by `TestHostileProbe_SelfLinksAccounting`. When task A links to itself (`[old](6fjangd7kva1-old.md)`), `plan.targetLinks` is tracked separately from `plan.cascadeEdits` and credited to `AppliedLinks` only when destination is created.
- **Same-slug / title-only rename:** Verified by `TestRenameTask_SameSlugDoesNotCascadeOrRemoveSource`. Rewrites H1 in place via `writeFileAtomic`; cascade edits count is 0; source is not removed (`DestinationWritten: true, SourceRemoved: false, Complete: true`).

#### Angle 7: Deterministic planning and accounting
- **Walk order:** `filepath.WalkDir` guarantees deterministic lexical traversal.
- **Link matching:** `repointLinks` matches by resolved target path relative to source directory; anchor fragments (`#h`) and query strings are preserved; code fences are ignored; reference-style links (`[label]: target`) are updated without modifying labels.
- **File filtering:** `markdownDoc(d)` enforces regular `.md` files, ignoring symlinks and directories.

#### Angle 8: Architectural fit
- Domain and core packages maintain strict isolation from CLI presentation, Cobra, Bubble Tea, and filesystem lock types.
- `TaskRenameReceipt` and `TaskRenameFailure` are portable value types.
- Wire translation (`wire.ToTaskRenameRecoveryJSON`) accurately projects core receipts into schema 1.62.

#### Angle 9: Systemic duplication and consistency
- `RenameTask` follows the guarded mutation pattern established by `CreateAndStartTask`, `MutateTaskGraph`, `ApplyThreadPlan`, and `RepairTaskGraph`.
- Reuses `checkedWriteLock()`, `verifyUnchanged()`, `createFileAtomic()`, and `writeFileAtomic()`.
- Distinguishes entity-agnostic path CAS (`verifyTaskRenamePath`) for arbitrary Markdown cascade targets.

#### Angle 10: Systemic second pass on obscured edge cases
- **Named return values and `errors.Join`:** Verified lines 31, 55-63. If unlock fails after completion, named return `err` is joined and wrapped into `TaskRenameFailure`. If unlock fails on an uncommitted conflict, `result.Committed` remains false; `service_task.go` returns unwrapped joined error, correctly mapped to exit code 14 without generating a misleading `error.task_rename` wire object.
- **Crashes during `createFileAtomic`:** If killed between `OpenFile` and `Sync`, an incomplete destination file creates an ambiguous match (`ErrAmbiguous`) rather than corrupting the old source, halting further mutations until the orphaned file is deleted.

---

### Mutation validation matrix

| Mutation tested | Guard targeted | Expected result | Observed result | Status |
|---|---|---|---|---|
| **1. Post-lock target recheck** | `internal/store/rename.go:68` | `TestRenameTask_RechecksTargetAfterRepositoryGuard` fails | FAILED: `post-lock target collision = ... ambiguous match; want uncommitted conflict` | **Killed** |
| **2. Capture-before-lock source validation** | `internal/store/rename.go:74` | `TestRenameTask_ConcurrentRenamesCommitOneSourceSnapshot` fails when source refreshed after lock | FAILED: `concurrent renames successes=2 conflicts=0, want one each` | **Killed** |
| **3. Per-cascade CAS** | `internal/store/rename.go:89` | `TestRenameTask_CASCatchesRawCascadeDocumentEdit` fails | FAILED: `raw cascade race = {Task:... Complete:true}, <nil>; want uncommitted conflict` | **Killed** |
| **4. Exclusive destination creation** | `internal/store/rename.go:118` | Target collision after plan creation fails closed with `ErrConflict` | Existing tests pass (see finding L1); hostile probe `TestHostileProbe_DestinationCreateCollisionRace` verified `createFileAtomic` prevents clobbering | **Pinning gap (L1)** |
| **5. Pre-destination source CAS** | `internal/store/rename.go:111` | Source edit during cascade phase rejected before destination write | Existing tests pass (see finding L1); verified by hostile code inspection and probe | **Pinning gap (L1)** |
| **6. Final pre-removal source CAS** | `internal/store/rename.go:143` | `TestRenameTask_SourceRemovalCASCatchesRawEdit` fails | FAILED: `source-removal CAS receipt = ... SourceRemoved:true, err=<nil>` | **Killed** |
| **7. Committed failure wrapping** | `internal/core/service_task.go:681` | Recovery tests fail to unwrap `TaskRenameFailure` | 4 FAILED: `PartialCascadeReceiptIsResumable`, `DestinationWrittenCleanupFailureRequiresInspection`, `SourceRemovalCASCatchesRawEdit`, `CompleteUnlockFailureIsNotRetryable` | **Killed** |

---

### Wire and schema validation

- **Draft 2020-12 Validation**: Verified via `TestWriteError_TaskRenameSchemaAndNonDefaultValues`. `ErrorEnvelope` carrying `TaskRenameRecoveryJSON` populated with non-default values (including all optional `WorkspaceJSON` fields: `RepoID`, `Branch`, `Checkout`, `Space`, `Source`) compiles and validates against `wire.JSONSchema()`.
- **Negative Wire Check**: Verified that non-rename errors and uncommitted rename errors serialize to `ErrorEnvelope` without the `"task_rename"` property (`omitempty` respected), and validate cleanly against the schema.
- **Human vs. Machine Equivalence**:
  - *Retryable prefix*: Machine JSON emits `applied_documents=1, planned_documents=3, applied_links=1, planned_links=2, destination_written=false, source_removed=false` with convergent rerun remedy. Human CLI output renders identical counts and remedy text.
  - *Destination-written / source-retained*: Machine JSON emits `applied_documents=2, planned_documents=2, destination_written=true, source_removed=false` with inspection remedy. Human CLI output renders inspection remedy.
- **Schema Version Bump**: Minor version `1.62` verified across `internal/wire/wire.go:261`, `internal/wire/schema_comments.json`, `internal/cli/testdata/golden/schema_jsonschema.golden`, and all 33 CLI golden files.

---

### Residual risks and rejected concerns

1. **POSIX user-space TOCTOU window:** A raw editor modifying a file within the sub-millisecond window between CAS reading and atomic replacing cannot be blocked by advisory repository locks. This is a fundamental property of user-space locking on POSIX filesystems, accurately documented in architecture and ADR contracts.
2. **Competing renames with different titles after partial failure:** If a user runs a rename, encounters a partial failure, and then attempts a rename to a *different* title rather than rerunning the same rename by stable ID, earlier cascade links become dangling. This is mitigated by explicit remedy text and `tskflwctl lint`.
3. **Universal transaction framework:** Rejected as out of scope per task specification. The implementation reuses the repository's established atomic primitives (`checkedWriteLock`, `verifyUnchanged`, `stageTemp`, `createFileAtomic`, `writeFileAtomic`) without introducing unnecessary abstraction churn.
