---
schema: 1
id: 6gmnhv6b5da7
bucket: closed
area: in-place-lint-repair-safety-implementation-antigravity
date: "2026-10-17"
updated_at: "2026-10-17"
---
# Audit: In-place lint repair safety implementation — antigravity — 2026-10-17

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

### Verdict and scope limits

**Not ready for closure without owner triage of 2 open CLI findings (`M1`, `L1`).**

Across `internal/store`, `internal/core`, `internal/cli`, and `internal/wire`, the core in-place repair safety mechanisms hold under adversarial testing:
- `(*FS).FixFrontmatter` (`internal/store/fix.go:31-186`) acquires `checkedWriteLock()` (`L43`), rechecks `authorizeMutation()` after lock wait (`L54`), verifies regular-file type (`os.Lstat` + `info.Mode().IsRegular()`) and original content hash (`verifyFixInPlacePath`, `L134`, `L198-L210`) immediately before `writeFileAtomic(path, fixed, 0o644)` (`L137`), and joins guard-release errors with any primary error while retaining completed prefix rows (`L47-L51`).
- `(*Service).RepairPlanning` (`internal/core/planning_maintenance.go:31-75`) computes `result.Committed` strictly from non-dry-run non-skipped `Fixes` rows (`!fixed.Skipped`, `L35-L39`) or `result.Relocation != nil && result.Relocation.DestinationWritten` (`L34`), never from preview proposals or `Skipped: true` rows, and wraps failed committed non-relocation runs in `*core.PlanningRepairFailure` (`L44-L46`) without replaying the batch.
- `ToPlanningRepairEnvelope` (`internal/wire/planning_repair.go:18-28`) keeps `PlanningRepairResult.Committed` internal (`internal/wire/planning_repair_test.go:38-80`) and preserves schema revision `1.89`.

However, the required second pass over shared CLI error/output helpers exposed two reachable defects in the `lint --fix --json` failure path when errors are joined or when `render.FixJSON` fails after `RepairPlanning` succeeds:
1. **`M1` (`internal/cli/fserror.go:21-59`):** `filesystemDetails(err)` extracts `Operation` and `Path` from the first `*fs.PathError` / `*os.LinkError` via `errors.As`, but passes the outer joined `err` to `classifyFilesystemError(err)`. When `FixFrontmatter` (`internal/store/fix.go:49`) or `runLintFix` (`internal/cli/lint.go:101`) joins a secondary `PathError` via `errors.Join`, `ErrorEnvelope.Error.Filesystem` splices the first error's `operation` and `path` with a `class` and `retryable` boolean matched across the entire joined error tree. On an uncommitted run (`Skipped: true` rows, `--dry-run`, or first-file failure), a non-retryable planning `PathError` (such as `lstat` `ENOTDIR` or `open` `EIO`) joined with a transient stdout or unlock `PathError` (`EAGAIN`/`EINTR`/`EBUSY`) emits `"retryable": true` for the broken planning path; on a committed run, joining `read` `EBUSY` on a task file with `write` `ENOSPC` or `EACCES` on `/dev/stdout` misattributes `"class": "no-space"` or `"permission"` to the task read.
2. **`L1` (`internal/cli/lint.go:127-138`, `internal/cli/planning_repair_failure.go:18`):** When `RepairPlanning(false)` succeeds (`err == nil`) after committing repairs (`repaired.Committed == true`), `runLintFix` returns a `render.FixJSON` writer error at `internal/cli/lint.go:128` before evaluating `core.BlockingLintResultCount(results2)+len(problems)` at `L134-L138`, dropping `domain.ErrValidation` and degrading the process exit code from `11` (`validation`) to `1` (`error`) even when `Error.PlanningRepair` carries non-empty `unreadable` or `remaining` arrays. In addition, because `RepairPlanning` returned `err == nil`, a standalone `FixJSON` writer failure after committed repairs is not wrapped in `*core.PlanningRepairFailure`, so `Error.Message` contains only the raw writer error without the committed-batch recovery guidance.

**Scope and evidence limits:**
- Per the brief, raw editors still retain the final verify-to-replace race window between `verifyFixInPlacePath` (`internal/store/fix.go:134`) and `os.Rename` inside `writeFileAtomic` (`internal/store/atomic.go:56`); no crash transaction or inode continuity is claimed.
- `FixFindingHeaders` (`internal/core/finding.go:299-352`) delegates per-audit body writes to `TransformAuditBody` -> `writeBody` (`internal/store/body.go:69-73`), which still uses `s.writeLock` (`internal/store/lock.go:103-109`, discarding unlock errors and not re-running `authorizeMutation()` after lock wait). Ordinary audit body write guards are scoped to tasks `6g77rn6hvmh8` / `6g77rn6n6b86` and `6g7wsr8yvt1a`.
- `EBUSY`, `EAGAIN`, and `ENOSPC` in CLI/store recovery tests are synthetic portable `syscall.Errno` values; guard-release errors use `testHookRepositoryUnlockError` (`internal/store/lock.go:132-134`); `syscall.ENOTDIR` in `TestVerifyFixInPlacePathRetainsFilesystemCause` (`internal/store/fix_inplace_test.go:268-277`) is a real kernel error from `os.Lstat`.

## Findings

#### M1. filesystemDetails splices first PathError path with joined error class and retryability · **Status:** fixed

**File:** internal/cli/fserror.go:21 | **Component:** cli
**Effort:** XS · **Urgency:** soon

When err is created with errors.Join (such as FixFrontmatter joining a guard-release error at internal/store/fix.go:49 or runLintFix joining a stdout writer error at internal/cli/lint.go:101), filesystemDetails(err) at internal/cli/fserror.go:27-37 extracts Operation and Path from the first matched *fs.PathError or *os.LinkError via errors.As, but passes the entire joined err to classifyFilesystemError(err). Because classifyFilesystemError evaluates errors.Is across the whole joined error tree in fixed precedence order, a non-retryable primary PathError (e.g. lstat ENOTDIR or open EIO) joined with a transient stdout or unlock PathError (EAGAIN/EINTR/EBUSY) on an uncommitted batch (skipped-only or dry-run) emits retryable=true for the primary planning path, and a committed batch joining PathErrors of different classes (e.g. read EBUSY on a task file plus write ENOSPC/EACCES on /dev/stdout) attributes the second error's class (no-space/permission) to the first error's operation and path.

- **Producer -> consumer trace:** `(*FS).FixFrontmatter` (`internal/store/fix.go:49`) or `runLintFix` (`internal/cli/lint.go:101`) calls `errors.Join(primaryErr, secondaryErr)` -> `WriteError` (`internal/cli/exit.go:154`) calls `filesystemDetails(err)` -> `internal/cli/fserror.go:28-29` binds `pathErr` via `errors.As(err, &pathErr)` (first `*fs.PathError` in DFS order) -> `internal/cli/fserror.go:37` calls `classifyFilesystemError(err)` on the outer joined `err` instead of `pathErr.Err` -> `internal/cli/exit.go:179` only overrides `Retryable = false` when `planningErr.result.Committed` is true.
- **Exact reproduction (executed in sandbox):**
  1. Uncommitted skipped-only batch where `FixFrontmatter` returns `fixes = [{Path: "/planning/tasks/skipped.md", Skipped: true}]` and `fixErr = &os.PathError{Op: "lstat", Path: "/planning/tasks/later.md", Err: syscall.ENOTDIR}` (non-retryable), while `app.Out` fails with `&os.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.EAGAIN}`:
     `WriteError(&stderr, err, true)` emits `{"schema_version":"1.89","error":{"code":"error","message":"lstat /planning/tasks/later.md: not a directory\nwrite /dev/stdout: resource temporarily unavailable","filesystem":{"class":"io","operation":"lstat","path":"/planning/tasks/later.md","retryable":true}}}`.
  2. Committed batch where `FixFrontmatter` returns `fixes = [{Path: "/planning/tasks/earlier.md", Skipped: false}]` and `fixErr = &os.PathError{Op: "read", Path: "/planning/tasks/later.md", Err: syscall.EBUSY}`, while `app.Out` fails with `&os.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.ENOSPC}`:
     `WriteError(&stderr, err, true)` emits `"filesystem":{"class":"no-space","operation":"read","path":"/planning/tasks/later.md","retryable":false}`.
- **Why existing tests missed it:** `TestCLILintFixKeepsCommittedReceiptWhenTransientStdoutErrorIsJoined` (`internal/cli/planning_repair_failure_test.go:121-145`) joined `domain.ErrConflict` (not a `PathError`) with one `PathError`. `TestCLILintFixRetainsOrdinaryRepairRecoveryWithoutBlindReplay/joined stdout failure` (`internal/cli/planning_repair_failure_test.go:148-209`) joined two `PathError`s that both had `Class: "io"` (`EBUSY` and `EAGAIN`) under `committed == true`.

**Recommendation:** Pass the matched pathErr.Err or linkErr.Err (rather than the outer joined err) to classifyFilesystemError in internal/cli/fserror.go:27-38 so Operation, Path, Class, and Retryable always describe the same filesystem error instance, and add regression coverage for joined PathErrors with distinct classes/retryability.

**Resolution:** Reproduced red-first for joined PathError and LinkError causes:
    classification and retryability now use the same matched error that supplies
    operation/path. Kept existing error selection and domain exit policy. Unit
    regressions cover mixed transient/non-retryable causes, and lint command
    tests cover skipped-only, preview and committed prefixes with returned
    stdout failures. Original joined causes and committed retry suppression
    remain intact.

#### L1. Post-repair FixJSON writer error drops residual validation exit code and committed-prefix prose guidance · **Status:** fixed

**File:** internal/cli/lint.go:127 | **Component:** cli
**Effort:** XS · **Urgency:** eventually

When Service.RepairPlanning(false) returns err == nil after committing repairs (repaired.Committed == true), runLintFix at internal/cli/lint.go:127-129 returns a render.FixJSON writer error immediately before evaluating core.BlockingLintResultCount(results2)+len(problems) at lines 134-138, and without wrapping the post-RepairPlanning failure in *core.PlanningRepairFailure. If residual lint issues or unreadable records exist alongside a stdout writer error, domain.ErrValidation is dropped and the command exits 1 (code: error) instead of 11 (code: validation) despite carrying non-empty remaining/unreadable arrays in Error.PlanningRepair. In addition, on any standalone FixJSON writer error after committed repairs, Error.Message contains only the raw writer error without core.PlanningRepairFailure's committed-batch guidance.

- **Producer -> consumer trace:** `app.Svc.RepairPlanning(false)` (`internal/core/planning_maintenance.go:31-75`) succeeds with `err == nil`, `result.Committed == true`, and non-empty `result.Problems` / `result.LintResults` (because `err == nil`, the `defer` at `planning_maintenance.go:44-46` does not wrap in `*core.PlanningRepairFailure`) -> `runLintFix` (`internal/cli/lint.go:127-129`) calls `render.FixJSON(app.Out, ...)` and returns `outputErr` immediately before `L134-L138` -> `runLintFix` `defer` (`internal/cli/lint.go:89-93`) wraps `outputErr` in `&planningRepairCommandFailure{cause: outputErr, ...}`, whose `Error()` (`internal/cli/planning_repair_failure.go:18`) returns `e.cause.Error()` -> `ExitCode(err)` (`internal/cli/exit.go:96-98`) returns `1` (`error`) instead of `11` (`validation`).
- **Exact reproduction (executed in sandbox):**
  With `fixes = [{Path: "/planning/tasks/earlier.md", Changes: ["tags normalized"]}]`, `problems = [{EntityKind: "task", EntityID: "000000000002", EntitySlug: "broken", LocalPath: "/planning/tasks/000000000002-broken.md", Message: "unterminated frontmatter"}]`, and `app.Out` failing with `&os.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.EAGAIN}`:
  `ExitCode(err)` is `1` (want `11`), and `WriteError(&stderr, err, true)` emits `{"schema_version":"1.89","error":{"code":"error","message":"write /dev/stdout: resource temporarily unavailable","planning_repair":{"schema_version":"1.89","dry_run":false,"fixed":[{"path":"/planning/tasks/earlier.md","changes":["tags normalized"]}],"unreadable":[{"entity_kind":"task","entity_id":"000000000002","entity_slug":"broken","path":"/planning/tasks/000000000002-broken.md","message":"unterminated frontmatter"}],"remaining":[],"workspace":{"planning_root":"","source":""}},"filesystem":{"class":"io","operation":"write","path":"/dev/stdout","retryable":false}}}`.

**Recommendation:** Compute any residual domain.ErrValidation before returning from render.FixJSON and join writer errors via errors.Join so exit/error classification stays validation (11), and wrap post-RepairPlanning writer failures after committed ordinary repairs in *core.PlanningRepairFailure.

**Resolution:** Reproduced red-first. Residual validation is now established
    before JSON output and joined with writer errors, preserving exit 11 and
    both causes. Core-owned PlanningRepairResult.RecoveryError supplies ordinary
    committed-prefix guidance for presentation failures without duplicating
    advice or replacing relocation guidance. Tests cover clean, unreadable and
    remaining-issue results, structured receipts, original writer causes and no
    replay.

### Verified producer/consumer inventory

| Layer / File | Verified Symbol / Lines | Verified Contract Role |
| --- | --- | --- |
| `internal/store/fix.go` | `(*FS).FixFrontmatter(dryRun bool) (results []domain.FixResult, err error)` (`L31-L186`), `testHookBeforeFixInPlaceWrite` (`L193`), `verifyFixInPlacePath(path, version string) error` (`L198-L210`) | Acquires `checkedWriteLock()` (`L43`) when `!dryRun`, joins `unlock()` errors via `errors.Join` (`L47-L51`), rechecks `authorizeMutation()` after lock wait (`L54`), and checks `os.Lstat` regular-file mode + `verifyFixPath(path, hashContent(content))` (`L134`, `L199-L209`) before `writeFileAtomic(path, fixed, 0o644)` (`L137`). |
| `internal/store/lock.go` | `(*FS).checkedWriteLock() (func() error, error)` (`L115-L137`), `testHookRepositoryUnlockError` (`L139`) | Acquires process + platform flock and returns a checked release closure joining `unlockPlatform()` and `testHookRepositoryUnlockError()`. |
| `internal/core/frontmatter_repair.go` | `FrontmatterRepairRelocation` (`L11-L16`), `FrontmatterRepairFailure` (`L21-L30`), `PlanningRepairFailure` (`L36-L44`) | `PlanningRepairFailure.Error()` appends `"; earlier planning repairs committed; inspect the reported repairs and run lint before deciding how to continue; do not blindly replay the whole repair batch"` while unwrapping `Cause`. |
| `internal/core/planning_maintenance.go` | `PlanningRepairResult` (`L13-L24`, field `Committed bool` at `L20`), `(*Service).RepairPlanning(dryRun bool)` (`L31-L75`) | Sets `result.Committed` in `defer` (`L32-L41`) iff `!dryRun` and either `result.Relocation != nil && result.Relocation.DestinationWritten` or any `result.Fixes` entry has `!fixed.Skipped`. Wraps non-nil `err` in `*PlanningRepairFailure` when `result.Committed && result.Relocation == nil` (`L44-L46`). |
| `internal/core/store.go` | `Fixer` interface (`L446-L458`) | Documents completed-prefix retention, post-wait authorization recheck, in-place CAS, guard-release error retention, and partial relocation separation. |
| `internal/cli/lint.go` | `newLintCmd` (`L14-L49`), `runLintFix(app *App, dryRun bool) (err error)` (`L85-L140`) | `defer` (`L89-L93`) wraps `err != nil` in `*planningRepairCommandFailure` when `repaired.Committed || repaired.Relocation != nil`; joins `render.PlanningRepairJSON` errors via `errors.Join(err, outputErr)` (`L100-L102`). |
| `internal/cli/planning_repair_failure.go` | `planningRepairCommandFailure` (`L11-L19`) | Carries `cause`, `result`, `dryRun`, and `workspace` to `WriteError`. |
| `internal/cli/exit.go` & `internal/cli/fserror.go` | `WriteError` (`exit.go:137-222`), `filesystemDetails` (`fserror.go:21-39`) | Populates `payload.Error.PlanningRepair` from `planningRepairCommandFailure` and sets `payload.Error.Filesystem.Retryable = false` when `planningErr.result.Committed && payload.Error.Filesystem != nil` (`exit.go:175-182`). |
| `internal/wire/planning_repair.go` & `envelopes.go` | `ToPlanningRepairEnvelope` (`planning_repair.go:18-28`), `FixEnvelope` (`envelopes.go:929-939`), `ErrorItem.PlanningRepair` (`envelopes.go:1132`) | Serializes `fixed`, `unreadable`, `remaining`, `dry_run`, `workspace`, and optional `relocation`; `PlanningRepairResult.Committed` is not serialized. |

### End-to-end real-store cross-check and plain-lint recovery trace

1. **Real `store.FS` -> `core.Service.RepairPlanning(false)` end-to-end trace (real filesystem effects + hook-injected raw drift + synthetic `syscall.EBUSY` unlock error):**
   - Seeded a disposable `t.TempDir()` repository containing `epics/01-e1.md` (`status: active`) and three task files in `tasks/`:
     - `000000000001-earlier.md` (mode `0600`, repairable unquoted colon `description: Phase 1: do it # keep comment`, `tags: one,two`, `custom: preserve`).
     - `000000000002-skipped.md` (graph-owned `depends_on: 000000000001,000000000003`, which `FixFrontmatter` refuses to normalize).
     - `000000000003-later.md` (mode `0600`, repairable frontmatter).
   - Configured `testHookBeforeFixInPlaceWrite` to append `"\nConcurrent raw editor note.\n"` to `000000000003-later.md` immediately before `verifyFixInPlacePath`, and `testHookRepositoryUnlockError` to return `&os.PathError{Op: "flock", Path: root, Err: syscall.EBUSY}`.
   - Calling `core.MustNewService(store.NewFS(root, core.UnrestrictedMutations())).RepairPlanning(false)` produced:
     - `repaired.Committed == true`, `repaired.Relocation == nil`, `len(repaired.Fixes) == 2`:
       - `Fixes[0]`: `Path: ".../tasks/000000000001-earlier.md"`, `Changes: ["description: quoted value containing ':'", "tags: normalized to a YAML list"]`, `Skipped: false`. On disk, `000000000001-earlier.md` was atomically updated with `# keep comment`, `custom: preserve`, body, and mode `0600` intact.
       - `Fixes[1]`: `Path: ".../tasks/000000000002-skipped.md"`, `Changes: ["graph-owned repair refused (depends_on); repair deliberately, run lint, then use guarded dependency operations"]`, `Skipped: true`.
     - `000000000003-later.md` on disk retained the concurrent editor's exact bytes (`winner`) and mode `0600`.
     - `err` satisfied `errors.Is(err, domain.ErrConflict)`, `errors.Is(err, syscall.EBUSY)`, and `errors.As(err, &ordinary)` (`*core.PlanningRepairFailure`), with message:
       `conflict: frontmatter repair file .../tasks/000000000003-later.md changed on disk\nrelease repository frontmatter repair guard: flock ...: resource busy; earlier planning repairs committed; inspect the reported repairs and run lint before deciding how to continue; do not blindly replay the whole repair batch`.

2. **Executed plain-lint recovery on the partially repaired corpus (`svc.Lint()` / `tskflwctl lint`):**
   - Running plain `svc.Lint()` against that exact post-failure directory returned `lintErr == nil`, `len(lintResults) == 0`, and `len(lintProblems) == 2` (which at the CLI boundary exits `11` / `domain.ErrValidation` as expected when residual problems remain):
     - `000000000001-earlier.md` loaded cleanly with zero lint issues (confirming its committed repair persisted cleanly).
     - `lintProblems[0]`: `EntityKind: "task"`, `EntityID: "000000000002"`, `EntitySlug: "skipped"`, `Message: "validation failed: malformed frontmatter: field \"depends_on\" must be a YAML list, but it is a string (\"000000000001,000000000003\")\n       fix: depends_on: [000000000001, 000000000003]"` — isolating the graph-owned field for deliberate operator remediation.
     - `lintProblems[1]`: `EntityKind: "task"`, `EntityID: "000000000003"`, `EntitySlug: "later"`, `Message: "validation failed: malformed frontmatter: field \"description\" (line 2) has a value containing an unquoted ':' — wrap the value in quotes, e.g. description: \"...\""` — isolating the file whose concurrent edit won the CAS check so the operator can inspect the concurrent note before re-running `lint --fix` or editing manually.

### Machine-contract revision 1.89 compatibility assessment (ADR-0008)

Challenging the retention of `SchemaVersion = "1.89"` (`ADDITIVE`) against [ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md):
- **Static schema and wire types:** `PlanningRepairResult.Committed` (`internal/core/planning_maintenance.go:20`) is internal to `core`; no JSON key, type, requiredness, or closed vocabulary changed (`internal/cli/testdata/golden/schema_jsonschema.golden` is unchanged, and `TestPlanningRepairOrdinaryProgressRetainsExistingWireShape` in `internal/wire/planning_repair_test.go:38-80` validates populated `ErrorEnvelope` instances against `#/$defs/ErrorEnvelope`).
- **Behavioral compatibility:** Revision `1.89` (`internal/wire/wire.go:378-383`) already added optional `ErrorItem.planning_repair` (`*FixEnvelope`) and `FilesystemErrorJSON.Retryable` suppression. Extending `ErrorItem.planning_repair` population and `filesystem.retryable = false` from relocation-only commits to ordinary committed prefixes restores the documented meaning of `FilesystemErrorJSON.Retryable` (`"whether repeating the call unchanged could succeed"`, `internal/wire/envelopes.go:1157`) without altering any existing key or vocabulary. Retaining `1.89` is therefore sound under ADR-0008 §3 (note: the comment at `internal/wire/wire.go:382` says `"filesystem retry hints are false after destination commit"` and could be broadened to `"after repair commit"` for prose accuracy).

### Mutation testing table

All mutations were compiler-valid edits executed inside `/private/tmp/isolated-review.QX35Vz`, killed by runtime test assertions, and restored to green before verification:

| ID | Target File & Lines | Exact Compiler-Valid Mutation | Test Command Executed | Runtime Outcome (Killed) | Restored Green |
| --- | --- | --- | --- | --- | --- |
| `M-AG-1` | `internal/cli/exit.go:179` | Changed `if planningErr.result.Committed && payload.Error.Filesystem != nil` to `if false && planningErr.result.Committed && payload.Error.Filesystem != nil` | `go -C /private/tmp/isolated-review.QX35Vz test -v ./internal/cli -run 'TestCLILintFix'` | **FAILED at runtime:** `TestCLILintFixRetainsPartialRelocationWithoutCountingItFixed` (`L91`), `TestCLILintFixKeepsCommittedReceiptWhenTransientStdoutErrorIsJoined` (`L143`), and all 4 committed subtests of `TestCLILintFixRetainsOrdinaryRepairRecoveryWithoutBlindReplay` (`repair_failure`, `post-lint_failure`, `stdout_failure`, `joined_stdout_failure` at `L200`: `commit=true: recovery lost or false retry safety: ... "retryable":true`). | Restored and verified `PASS` (`0.220s`). |
| `M-AG-2` | `internal/core/planning_maintenance.go:36` | Changed `if !fixed.Skipped` to `if !fixed.Skipped || fixed.Skipped` (promoting skipped rows to `Committed = true`) | `go -C /private/tmp/isolated-review.QX35Vz test -v ./internal/core ./internal/cli -run 'TestRepairPlanningClassifiesDurableProgressWithoutReplay\|TestCLILintFixRetainsOrdinaryRepairRecoveryWithoutBlindReplay'` | **FAILED at runtime:** `TestRepairPlanningClassifiesDurableProgressWithoutReplay/skipped` (`planning_maintenance_test.go:137`: `Committed:true`) and `TestCLILintFixRetainsOrdinaryRepairRecoveryWithoutBlindReplay/skipped` (`planning_repair_failure_test.go:200`: `commit=false: recovery lost or false retry safety: ... "retryable":false`). | Restored and verified `PASS` (`0.108s` / `0.180s`). |
| `M-AG-3` | `internal/cli/lint.go:90` | Changed `if err != nil && (repaired.Committed || repaired.Relocation != nil)` to `if err != nil && repaired.Relocation != nil` | `go -C /private/tmp/isolated-review.QX35Vz test -v ./internal/cli -run 'TestCLILintFixRetainsOrdinaryRepairRecoveryWithoutBlindReplay'` | **FAILED at runtime:** all 4 committed subtests (`repair_failure`, `post-lint_failure`, `stdout_failure`, `joined_stdout_failure` at `planning_repair_failure_test.go:200`: `planning_repair` omitted and `"retryable":true`). | Restored and verified `PASS` (`0.338s`). |

### Bounded surviving-hypothesis ledger

| Hypothesis | Verdict | Evidence |
| --- | --- | --- |
| Named return/defer shadowing silently drops the primary or guard-release error in `FixFrontmatter`, `RepairPlanning`, or `runLintFix` | **Falsified** | `FixFrontmatter` (`internal/store/fix.go:31,47-51,151-169,185`) binds named `(results, err)` and joins `unlock()` errors via `errors.Join`. `RepairPlanning` (`internal/core/planning_maintenance.go:31-75`) and `runLintFix` (`internal/cli/lint.go:85-108`) preserve named `err` through their `defer` blocks. Verified by `TestFixInPlaceRetainsCompletedPrefixAndReleaseFailure` and `TestFixFrontmatterReportsGuardReleaseFailureEvenWithoutWrites`. |
| Guard failure after a no-op, denied authorization, or an earlier completed row is lost or reclassified as a successful batch | **Falsified** | `TestFixFrontmatterReportsGuardReleaseFailureEvenWithoutWrites` (`internal/store/fix_inplace_test.go:211-250`) and `TestRepairPlanningClassifiesDurableProgressWithoutReplay` (`internal/core/planning_maintenance_test.go:111-151`) verify that `releaseErr` fails the call in `repaired`, `no-op`, and `skipped` modes while `Committed` is true only for `repaired`. |
| Preview proposals or skipped rows become commits, or a partial relocation is counted fixed | **Falsified** | `RepairPlanning` (`internal/core/planning_maintenance.go:33-40`) gates `Committed` on `!dryRun` and `!fixed.Skipped` or `Relocation.DestinationWritten`. Killed by `M-AG-2`; partial relocation stays out of `Fixes` (`internal/store/fix.go:109-111`, `TestFixFrontmatterReleaseFailurePreservesPartialRelocation`). |
| A helper skips parse errors or a CAS check follows symlinks and overwrites a foreign entry | **Falsified** | `verifyFixInPlacePath` (`internal/store/fix.go:198-210`) uses `os.Lstat` and `!info.Mode().IsRegular()` before byte hashing (`verifyFixPath`), without YAML decoding. Verified across all 5 entity directories and 4 drift modes (`edit`, `delete`, `symlink`, `directory`) in `TestFixInPlaceRefusesRawDriftAcrossEntityKinds` (`internal/store/fix_inplace_test.go:35-102`). |
| Failure after a later phase or output bypasses the ordinary prefix wrapper; a raw transient error causes machine clients to replay already committed work | **Falsified for committed prefixes; Reproduced for post-`RepairPlanning` `FixJSON` exit classification (`L1`) and uncommitted joined `PathError` splicing (`M1`)** | Committed prefixes always populate `Error.PlanningRepair` and set `Error.Filesystem.Retryable = false` (`internal/cli/exit.go:175-182`, killed by `M-AG-1` and `M-AG-3`). However, `filesystemDetails` splices joined `PathError` fields (`M1`), and an early return on `render.FixJSON` writer failure at `internal/cli/lint.go:127-129` drops residual `domain.ErrValidation` classification (`L1`). |
| Populated recovery fields validate but have misleading rows, flags, workspace, or retry semantics; or invent relocation evidence / a `committed` JSON key | **Falsified for `planning_repair` envelope; Reproduced for `filesystem` sibling block under joined `PathError`s (`M1`)** | `ToPlanningRepairEnvelope` (`internal/wire/planning_repair.go:18-28`) omits `Relocation` when `nil`, never emits a `"committed"` key (`internal/wire/planning_repair_test.go:53`), and validates against `#/$defs/ErrorEnvelope`. The sibling `error.filesystem` block can misreport `class`/`retryable` when two `PathError`s are joined (`M1`). |

### Executed versus inherited validation checks

**Independently executed in `/private/tmp/isolated-review.QX35Vz`:**
- `just --justfile /private/tmp/isolated-review.QX35Vz/Justfile --working-directory /private/tmp/isolated-review.QX35Vz build` -> passed (`bin/tskflwctl v0.23.0-115-gf582588c`).
- `go -C /private/tmp/isolated-review.QX35Vz test -race -count=3 ./internal/store -run 'TestFixInPlace|TestFixFrontmatter|TestVerifyFixInPlacePath|TestFixIDRelocation'` -> passed (`5.169s`).
- `go -C /private/tmp/isolated-review.QX35Vz test -race -count=3 ./internal/core ./internal/cli ./internal/wire -run 'TestRepairPlanning|TestCLILintFix|TestPlanningRepair'` -> passed (`1.212s`, `1.381s`, `1.795s`).
- `go -C /private/tmp/isolated-review.QX35Vz test ./internal/cli ./internal/core ./internal/store ./internal/wire` -> full package suites passed (`9.381s`, `0.464s`, `9.641s`, `0.351s`).
- `just --justfile /private/tmp/isolated-review.QX35Vz/Justfile --working-directory /private/tmp/isolated-review.QX35Vz docs-check` -> passed (`go run ./internal/tools/docgen -out docs/cli` + `git diff --exit-code docs/cli`, exit 0 against the sandbox baseline commit).
- `/private/tmp/isolated-review.QX35Vz/bin/tskflwctl -C /private/tmp/isolated-review.QX35Vz lint --links` and `audit lint 6gmnhv6b5da7` -> passed (`✔ all planning entities and dependency links pass lint`, `✔ all audit findings pass lint`).

**Inherited from owner checkpoint (not re-run across the whole repo):**
- Full-repository `just test` (with `-race` across packages outside `store`/`core`/`cli`/`wire`), `just lint` (`golangci-lint`), and schema-comment generation.

### Isolation attestation

```text
sandbox_path=/private/tmp/isolated-review.QX35Vz
git_dir=/private/tmp/isolated-review.QX35Vz/.git
baseline_commit=f582588c0e98259e0896a34540c266a493269658
source_blob=8b5a993e004c7bae824af1d01cb5f706b5ef7de1
source_fingerprint=65af7e2925c8b1665d95db704b8575ba78eac460
deliverable=planning/audits/6gmnhv6b5da7-2026-10-17-in-place-lint-repair-safety-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner disposition (2026-10-17)

Confirmed the source helper's guarded delivery before editing this audit:
baseline `f582588c0e98259e0896a34540c266a493269658`, transferred blob
`53ce4079855f805a69a43e45271fbb74ff650393`, `verify=passed`,
`transfer=succeeded`, `delivery=confirmed`.

Accepted both findings and reproduced them with retained red-first tests. M1 now
classifies the same matched filesystem error that supplied operation/path, keeping
existing selection precedence and outer domain classification. Direct PathError/
LinkError tests and lint command tests cover mixed classes and retryability across
skipped, preview and committed outcomes. L1 now joins residual validation with
JSON-writer failure instead of losing exit 11. A core-owned `RecoveryError` method
also supplies committed-prefix guidance for presentation failures, preserving
joined causes without duplicating advice or replacing relocation reconciliation.
Clean, unreadable and remaining-issue outputs are pinned separately.

This report and Codex's report describe their captured baselines, not these later
owner changes. In particular, Codex's actual-store writer probe expected generic
exit 1 despite residual diagnostics; that was evidence of the old behavior, not
proof that the classification was correct. The complementary review exposed it.
No historical wire changelog was rewritten; these fixes restore coherent existing
field meanings and validation precedence without new keys or exit vocabulary.

After fixes, build, full race suite, lint and repeated focused core/CLI/wire tests
pass. `docs-check` still reports only the intentional uncommitted lint-help diff;
both independent baselines passed regeneration, and no command metadata changed
in this owner triage. Shared persistence limits remain explicitly out of scope.
