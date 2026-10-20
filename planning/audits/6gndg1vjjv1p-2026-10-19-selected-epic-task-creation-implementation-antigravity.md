---
schema: 1
id: 6gndg1vjjv1p
bucket: closed
area: selected-epic-task-creation-implementation-antigravity
date: "2026-10-19"
updated_at: "2026-10-20"
---
# Audit: Selected-epic task creation implementation — antigravity — 2026-10-19

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

Challenge the selected-epic task-creation foundation before the TUI form consumes it.
This is not a TUI implementation or a request to build a universal creation API.

Assigned primary lenses:
- Codex: adapter-neutral API/composition, selected identity, shared preparation and compatibility.
- Antigravity: guarded parent validation, authorization after waits, no-clobber and committed/finalization recovery.
Each reviewer independently cross-checks one important invariant from the other lens.
Do not read the sibling report before finalizing your own. No finding quota.

## Review target

Source checkout: /Users/andyeschbacher/git/andy-esch/taskflow.
Base: 5f62e684238209d9a8f111003623cbcc563f9552.
The implementation is UNCOMMITTED on main: capture tracked edits, untracked new
Go files, docs and planning. Reviewing HEAD alone misses this feature.

Implementing task: planning/tasks/6gmy4c0x6e31-bind-task-creation-to-the-selected-epic-identity.md.
Design: planning/research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md.
Owning contracts: docs/architecture/planning-data.md and persistence.md.

Verify these actual paths/symbols in your sandbox:
- internal/core/selected_task_creation.go: SelectedEpicTaskCreationStore,
  WithSelectedEpicTaskCreationStore, NewTaskForSelectedEpic.
- internal/core/service.go: optional aggregate discovery and validateSourceSets.
- internal/core/service_task.go: NewTask and shared prepareTaskCreation.
- internal/core/store.go: EpicMutationTarget and ordinary TaskStore.CreateTask.
- internal/store/create.go: CreateTaskForSelectedEpic, createTask, createEntityFile.
- internal/store/entity_target.go: resolveEpicMutationPath; resolve.go: resolveExactID.
- internal/store/lock.go: checkedWriteLock; epicstore.go: parseEpic.
- internal/core/selected_task_creation_test.go, source_set_test.go;
  internal/store/selected_task_creation_test.go, mutation_authorization_test.go.

Build a repository-wide consumer inventory: callers, composition roots, source-ID/reference semantics, template
consumers and receipts with rg. No production TUI/CLI caller has been switched to
the new API; test-invoked public Service/adapter paths are the reachable feature.
Do not claim a future form or unimplemented selected graph API already exists.

## Intended contract to challenge

1. This optional sibling port does not widen aggregate Store or require local paths.
   Split capabilities need stable, nonzero matching SourceSetID witnesses; explicit
   override wins and missing/typed-nil capability refuses without legacy fallback.
2. Target is an adapter-established exact epic source ID, not a prefix, declaration,
   Location or path. The full epic stem remains identity locally. Epic input is blank
   or exactly the selected ID; implicit substitution is forbidden.
3. Core shares ordinary creation defaults, input validation and scaffolds. Selected
   creation supports ready/next, explicitly rejects Start, and makes one attempt.
   Ordinary CLI matching/create-and-start and wire output remain unchanged.
4. Local parent resolution/parsing occurs inside the guard used for task creation.
   Reauthorize after waiting, before parent reads/create. Read-only/denied previews
   cannot create roots/files or read a malformed parent first.
5. Preserve task/Thread collision checks, no-clobber and exact durable receipts.
   A conflict or cleanup error cannot remint/replay or erase committed evidence.
6. Limits are explicit: dry-run is a non-reserving point-in-time check; raw writers
   retain a final race; later parent removal is not blocked. Same-ID delete/recreate
   is not detectable. No new permanent FK, persisted child IDs or schema is promised.

Criterion authoring, TUI form/session behavior, selected registry removal and other
graph/entity write APIs are separate tickets. Ordinary create's pre-existing
authorization timing is not silently changed by this selected-only entry point;
an adjacent defect still needs a concrete producer and scope classification.

## Mandatory evidence floor

Both: independently run the focused baseline and at least one valid-but-hostile
and one refused fixture through the public selected Service or adapter entry point.
Check durable bytes/modes/record identities, not just error text. Give exact commands
and results; baseline owner gates alone are not a clean review.

Codex must:
- Exercise standalone pathless composition with nil aggregate Store, a mismatched
  witness and an absent/typed-nil capability; confirm no reference-path fallback.
- Compare ordinary vs selected preparation using a non-default case (Next with
  description, explicit metadata or body/template), and inspect the Start refusal.
- Temporarily remove only the new selected-creation source-set validation row.
  Require the intended selected-epic case of
  TestNewServiceRejectsEveryMismatchedSplitCapability to fail; restore green.
- Cross-check one real-store stale-parent or finalization probe.

Antigravity must:
- Run the two deterministic queued-write tests with -race -count=20:
  TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard and
  TestSelectedEpicTaskCreationReauthorizesAfterWaiting. Explain which guard is held
  and why the authorizer signal proves the create is waiting; challenge that oracle.
- Temporarily replace resolveEpicMutationPath with ordinary ResolveEpicPath in
  the new adapter callback. Require
  TestSelectedEpicTaskCreationRefusesFallbackAndUnreadableParents (the prefix
  fixture) to fail, then restore green.
- Temporarily remove ONLY the in-guard reauthorization, preserving outer checks.
  Require TestSelectedEpicTaskCreationReauthorizesAfterWaiting to fail; restore green.
- Exercise a collision and a committed unlock error through the real Service;
  confirm one mint/attempt and the destination/Committed evidence. Cross-check
  capability/source-set binding rather than assuming the registry analogy proves it.

Mutations must compile. A compile error is invalid, not a killed mutant; name the
intended regression that fails and capture its restored pass. Keep a short ledger
(edit, invariant, test, observed result). Do not count unrelated failures as proof.

## Required hostile angles

Separate systemic pass after the local probes:
- Can shared prepareTaskCreation or createTask change legacy behavior, accidentally
  join the wrong parent/namespace, or make tests share the same mistaken oracle?
- Can a delayed permit, epic rename, missing/ambiguous parent, unreadable metadata,
  source declaration, symlink/type drift or direct low-level call bypass strict
  parent ownership? Distinguish actual reachable bugs from the documented raw-writer
  boundary and same-ID replacement limit.
- Does a no-op/preview/error path bypass policy or fabricate a path/identity? Does
  finalization failure survive both core and adapter boundaries without unsafe retry?
- Would another primary/secondary adapter need TUI, filesystem or broad Store types?
  Challenge source-set binding and unsupported capability behavior without demanding
  speculative framework abstractions or a change to existing identity promises.

A no-findings outcome must state which hypotheses were falsified versus unresolved.

## Validation and restoration

Use the injected isolated-review protocol BEFORE inspection or commands. Source is
read-only. No shared worktree/index, independent sibling report access, source Git
writes, pushes or manual report copy-back. The helper's sandbox baseline commit is
the only permitted commit; verify then guarded transfer the assigned audit only.

Use GOCACHE and GOLANGCI_LINT_CACHE in a reviewer-owned temporary directory OUTSIDE
the sandbox corpus. Do not write probe fixtures into the user's real planning space.
Restore temporary production/test mutations and remove only your scratch artifacts
before verification. On capture/permission/evidence failure, keep the sandbox and
report incomplete; never fall back to the live checkout.

Owner baseline: fresh just build; full just test (-race); just lint; planning/link
lint and diff checks passed. Reviewers may inherit those full-suite results but
must independently run focused hostile evidence:
GOCACHE=<review-cache> go test -race ./internal/core ./internal/store -run 'TestSelectedEpic|TestNewService|TestEveryFSMutation' -count=1
For repeat probes restrict -run to the two waiting tests and use -count=20.
Run broader affected tests if your finding/probe changes shared behavior.
A resource/permission failure is not equivalent to passing.

## Deliverable

Keep this brief; replace only the Reviewer report slot. Findings first, with the
injected audit finding grammar, all findings OPEN. For each: root cause, reachable
producer, consequence, reproduction, evidence, recommended direction and remaining
uncertainty. Distinguish product defects, coverage gaps and preferences; consolidate
shared roots rather than inflate severity or count.

Include verified boundary inventory, focused commands/results, mutation ledger and
a bounded surviving-hypothesis list. Preserve documentation limitations truthfully.
Report the final isolation verification attestation (transfer still pending there).
Your handoff must carry the transfer receipt; the owner can confirm it against the
retained sandbox. Do not claim ready if material probes or isolation did not finish.

## Reviewer report

### Summary

No product defects found. One low-severity test-oracle gap (`L1`, coverage gap) was identified and demonstrated in `TestSelectedEpicTaskCreationReauthorizesAfterWaiting`, where the `GuardedMutations` callback closes the synchronization channel before reading the atomic denial flag during the pre-lock check in `checkedWriteLock` and does not assert that the third (post-wait, in-guard) authorization check executed.

### Findings

#### L1. Reauthorization wait test closes readiness channel before evaluating denial in pre-lock check · **Status:** fixed

- **Classification:** Coverage / test-oracle gap (production implementation in `internal/store/create.go:192-197` is sound).
- **Root cause:** In `TestSelectedEpicTaskCreationReauthorizesAfterWaiting` (`internal/store/selected_task_creation_test.go:165-174`), the `core.GuardedMutations` callback increments `checks`, closes `waiting` when `checks == 2` (which occurs inside `checkedWriteLock` at `internal/store/lock.go:116`, *before* `processRepositoryLock(s.root)` at `internal/store/lock.go:122`), and only afterwards evaluates `if denied.Load()` (`internal/store/selected_task_creation_test.go:170`). Neither `TestSelectedEpicTaskCreationReauthorizesAfterWaiting` nor `TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard` asserts `checks == 3` / `authorizations == 3` after the worker goroutine returns.
- **Reachable producer:** `go test -race ./internal/store -run '^TestSelectedEpicTaskCreationReauthorizesAfterWaiting$'` when the scheduler switches to the main test goroutine upon `close(waiting)` (`internal/store/selected_task_creation_test.go:168`) and executes `denied.Store(true)` (`internal/store/selected_task_creation_test.go:186`) before the worker goroutine evaluates `if denied.Load()` (`internal/store/selected_task_creation_test.go:170`) inside call #2 (`checkedWriteLock` entry).
- **Consequence:** Under that scheduling interleaving, call #2 (`checkedWriteLock` pre-lock check at `internal/store/lock.go:116`) returns `core.ErrReadOnlyPersistence` before acquiring `processRepositoryLock(s.root)` or reaching the in-guard reauthorization (`internal/store/create.go:195`, call #3). If the in-guard `s.authorizeMutation()` in `CreateTaskForSelectedEpic` (`internal/store/create.go:195-197`) is removed, `TestSelectedEpicTaskCreationReauthorizesAfterWaiting` still passes whenever preemption occurs between lines 168 and 170.
- **Reproduction & hostile evidence:** Removing `if err := s.authorizeMutation(); err != nil { return err }` from `internal/store/create.go:195-197` and yielding (`time.Sleep(time.Millisecond)`) immediately after `close(waiting)` at `internal/store/selected_task_creation_test.go:168` caused `GOCACHE=/private/tmp/antigravity-review-cache-QSZzN6 go test -C /private/tmp/isolated-review.QSZzN6 -race ./internal/store -run '^TestSelectedEpicTaskCreationReauthorizesAfterWaiting$' -count=1` to pass (`ok github.com/andy-esch/taskflow/internal/store`) with the in-guard reauthorization completely deleted.
- **Recommended direction:** Match `TestFixInPlaceRechecksAuthorizationAfterWaiting` (`internal/store/fix_inplace_test.go:110-118`) by evaluating `if denied.Load()` *before* `if checks == 2 { close(waiting) }` in `TestSelectedEpicTaskCreationReauthorizesAfterWaiting`, and assert `checks == 3` (and `authorizations == 3` in `TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard`) after `<-done`.
- **Remaining uncertainty:** None; the race window in the test oracle is fully explained and reproduced.

**Resolution:** Evaluate the pre-lock permit before signaling readiness, then
    assert all three authorization checks after the worker returns; the
    parent-drift test also asserts three checks. Both queued tests pass with
    -race -count=20. A Go overlay removing only in-guard reauthorization
    compiles and makes the intended test fail (two checks instead of three),
    without modifying production source.

### Queued-write guard and oracle analysis

In both `TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard` (`internal/store/selected_task_creation_test.go:96-149`) and `TestSelectedEpicTaskCreationReauthorizesAfterWaiting` (`internal/store/selected_task_creation_test.go:151-197`):

1. **Held guard:** The test acquires `processRepositoryLock(root)` (`internal/store/lock.go:63-67`), which locks the per-canonical-root in-process mutex `repositoryGuardFor(root).write` (`internal/store/lock.go:16-20, 51-61`).
2. **Why `authorizations == 2` signals arrival at `checkedWriteLock`:** `CreateTaskForSelectedEpic` (`internal/store/create.go:188-211`) delegates to `createTask` (`internal/store/create.go:213-263`) and `createEntityFile` (`internal/store/create.go:70-116`). Authorization check #1 runs at `createEntityFile` entry (`internal/store/create.go:71`). After `os.MkdirAll(s.root, 0o755)` (`internal/store/create.go:94`), `createEntityFile` calls `s.checkedWriteLock()` (`internal/store/create.go:97`). Authorization check #2 runs at `checkedWriteLock` entry (`internal/store/lock.go:116`), immediately before `rejectRepositoryPlannerCall()` (`internal/store/lock.go:119`) and `processRepositoryLock(s.root)` (`internal/store/lock.go:122`). Closing `waiting` on check #2 establishes a happens-before edge (`close(waiting)` $\to$ `<-waiting` $\to$ parent rename / `denied.Store(true)` $\to$ `release()` $\to$ `guard.write.Lock()` return) guaranteeing that `prepare()` (`internal/store/create.go:108`) and its `validateParent` callback (`internal/store/create.go:192-210`) cannot run until after the test mutates state and releases `guard.write`.
3. **Oracle challenge:** Because check #2 runs at `internal/store/lock.go:116` *before* `guard.write.Lock()` at `internal/store/lock.go:122`, closing `waiting` prior to `if denied.Load()` in `TestSelectedEpicTaskCreationReauthorizesAfterWaiting` allows `denied.Store(true)` in the test goroutine to race with `if denied.Load()` in check #2 itself (captured in finding `L1`). In `TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard`, the authorizer always returns `nil`, so check #2 cannot short-circuit and `validateParent()` is guaranteed to run after `os.Rename` and `release()`.

### Verified symbol and boundary inventory

All target paths, symbols, and consumers were verified in `/private/tmp/isolated-review.QSZzN6`:

- `internal/core/selected_task_creation.go`: `SelectedEpicTaskCreationStore` (`L23-25`), `WithSelectedEpicTaskCreationStore` (`L29-35`), `(*Service).NewTaskForSelectedEpic` (`L46-68`).
- `internal/core/service.go`: `Service.selectedTaskCreations` field (`L19`), automatic discovery from aggregate `Store` in `NewService` (`L369-371`), and `{"selected-epic task creation", s.selectedTaskCreations}` in `(*Service).validateSourceSets` (`L424`).
- `internal/core/service_task.go`: `(*Service).NewTask` (`L924-952`), `validateTaskCreationMode` (`L954-959`), and shared `(*Service).prepareTaskCreation` (`L963-1032`).
- `internal/core/store.go`: `EpicMutationTarget` (`L39-42`) and `TaskStore.CreateTask` (`L60-63`).
- `internal/core/creation_receipt.go`: `CreateFinalizationError` (`L9-12`), `LocalCreateOutcome` (`L18-30`), `TaskCreationReceipt` (`L35-40`).
- `internal/store/fsstore.go`: `var _ core.SelectedEpicTaskCreationStore = (*FS)(nil)` (`L99`), `(*FS).SourceSetID` (`L67`), `(*FS).authorizeMutation` (`L77-79`).
- `internal/store/create.go`: `(*FS).createEntityFile` (`L70-116`), `(*FS).writeNewFileUnlocked` (`L125-136`), `(*FS).CreateTask` (`L179-181`), `(*FS).CreateTaskForSelectedEpic` (`L188-211`), `(*FS).createTask` (`L213-263`), `ensureCandidateIDUnique` (`L268-278`), `(*FS).ensureTaskIDNotThread` (`L297-308`).
- `internal/store/entity_target.go`: `(*FS).resolveEpicMutationPath` (`L52-65`).
- `internal/store/resolve.go`: `markdownDoc` (`L24-26`), `epicCandidates` (`L136-158`), `flatCandidates` (`L170-194`), `resolveID` (`L224-285`), `resolveExactID` (`L298-314`).
- `internal/store/lock.go`: `processRepositoryLock` (`L63-67`), `(*FS).checkedWriteLock` (`L115-137`), `testHookRepositoryUnlockError` (`L139`).
- `internal/store/epicstore.go`: `(*FS).resolveEpicPath` (`L261-271`), `parseEpic` (`L273-289`).
- Tests and docs: `internal/core/selected_task_creation_test.go` (`L1-153`), `internal/core/source_set_test.go` (`L25, L44, L75, L107-108`), `internal/store/selected_task_creation_test.go` (`L1-249`), `internal/store/mutation_authorization_test.go` (`L27-30`), `docs/architecture/planning-data.md` (`L69-85`), `docs/architecture/persistence.md` (`L85-92`), `planning/tasks/6gmy4c0x6e31-bind-task-creation-to-the-selected-epic-identity.md` (`L1-84`), `planning/research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md` (`L1-257`).
- Consumer inventory (`rg` across `internal/` and `cmd/`): `appwiring.openPlanning` (`internal/appwiring/wiring.go:70`) composes `core.NewService(fs)`, which automatically binds `SelectedEpicTaskCreationStore` from `*store.FS`. No CLI or TUI controller calls `NewTaskForSelectedEpic` or `CreateTaskForSelectedEpic` yet; the TUI form consumer remains scoped to follow-up task `6g4ym5y2dkp8`.

### Focused commands and results

1. **Focused baseline suite:**
   - Command: `GOCACHE=/private/tmp/antigravity-review-cache-QSZzN6 go test -C /private/tmp/isolated-review.QSZzN6 -race ./internal/core ./internal/store -run 'TestSelectedEpic|TestNewService|TestEveryFSMutation' -count=1`
   - Result: `PASS` (`internal/core` 1.174s, `internal/store` 1.277s).
2. **Queued-write deterministic race repetition (`-race -count=20`):**
   - Command: `GOCACHE=/private/tmp/antigravity-review-cache-QSZzN6 go test -C /private/tmp/isolated-review.QSZzN6 -race ./internal/store -run '^(TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard|TestSelectedEpicTaskCreationReauthorizesAfterWaiting)$' -count=20`
   - Result: `PASS` (`internal/store` 1.235s).
3. **Hostile valid, refused, collision/finalization, and real-store `SourceSetID` probes (temporary `TestAntigravityReviewProbes`, removed before verification):**
   - Command: `GOCACHE=/private/tmp/antigravity-review-cache-QSZzN6 go test -C /private/tmp/isolated-review.QSZzN6 -race -v ./internal/store -run '^TestAntigravityReviewProbes$' -count=1`
   - Result: `PASS` (`internal/store` 1.241s). Verified:
     - Valid-but-hostile parent (`01-202601010000-parent.md` mode `0o600` with misleading frontmatter `id: 99-impostor`, comments, custom keys, and sibling `01-202601010000-parent-two.md` + `99-impostor.md`) with standalone `core.NewService(nil, core.WithSelectedEpicTaskCreationStore(fsB), ...)` (`HasLocalPath` false for epic/task) and `Next: true`, colon-bearing description, and punctuation title `"Hostile / Task: #1"`: minted once (`minted == 1`, `sleeps == 0`), wrote regular file `tasks/6gmy4c0x6e31-hostile-task-1.md` with mode `0o644`, quoted `description: 'has colon: yes'`, `epic: 01-202601010000-parent`, `status: next-up`, reloaded with `record.Source.ID == "6gmy4c0x6e31"`, and preserved parent/sibling bytes and `0o600` mode (`maps.Equal(before, after)`).
     - Mismatched real `*store.FS` instances (`core.NewService(fsA, core.WithSelectedEpicTaskCreationStore(fsB))`): refused at construction with `core.ErrIncompatibleCapabilities` and `domain.ErrValidation`.
     - Refused symlink parent (`01-symlink.md -> 01-real.md`) and cross-kind Thread ID collision (`threads/6gmy4c0x6e31-existing-thread.md` mode `0o600`) across `dryRun={false,true}` with `WithRetry(3, ...)`: refused with `domain.ErrNotFound` and `domain.ErrConflict` respectively, `minted == 1`, `sleeps == 0`, `Committed == false`, `Local == LocalCreateOutcome{}`, and unchanged tree bytes/modes.
     - Same-kind task ID collision (`tasks/6gmy4c0x6e31-existing.md` mode `0o600`) + unlock failure (`testHookRepositoryUnlockError` returning `domain.ErrConflict`) and committed write + unlock failure through `Service` with `WithRetry(3, ...)`: both returned `domain.ErrConflict` and `*core.CreateFinalizationError` with `minted == 1`, `releases == 1`, `sleeps == 0`; collision kept `Committed == false`, `Local == LocalCreateOutcome{}`, and existing `0o600` file untouched; committed failure kept `Committed == true`, `Local.CommittedPath == Local.PlannedPath`, mode `0o644`, and readable task identity `6gmy4c0x6e31`.

### Mutation ledger

| # | Target & compiler-valid edit | Invariant challenged | Intended test & command | Observed mutant result | Restored result |
| --- | --- | --- | --- | --- | --- |
| 1 | `internal/store/create.go:198`: replace `s.resolveEpicMutationPath(target.ID)` with `s.ResolveEpicPath(target.ID)` | Selected epic lookup must use exact whole-stem identity without prefix/fuzzy fallback | `TestSelectedEpicTaskCreationRefusesFallbackAndUnreadableParents` (`go test -race ./internal/store -run '^TestSelectedEpicTaskCreationRefusesFallbackAndUnreadableParents$' -count=1`) | **FAIL** (`prefix_is_not_identity` and `impostor_sibling` accepted in both `dry=false` and `dry=true`) | **PASS** (1.150s) |
| 2 | `internal/store/create.go:195-197`: remove in-guard `if err := s.authorizeMutation(); err != nil { return err }`, keeping outer `createEntityFile` and `checkedWriteLock` checks | Policy revocation while waiting on the repository guard must refuse before reading parent or creating task | `TestSelectedEpicTaskCreationReauthorizesAfterWaiting` (`go test -race ./internal/store -run '^TestSelectedEpicTaskCreationReauthorizesAfterWaiting$' -count=1`) | **FAIL** under unpreempted scheduling (`expired policy permitted create: <nil>`); **PASS (masked)** when preempted after `close(waiting)` (documented as finding `L1`) | **PASS** (1.202s) |
| 3 | `internal/core/service.go:424`: remove `{"selected-epic task creation", s.selectedTaskCreations}` from `validateSourceSets()` | Split `SelectedEpicTaskCreationStore` must publish a non-zero `SourceSetID` matching the service's other capabilities | `TestNewServiceRejectsEveryMismatchedSplitCapability` & `TestNewServiceRejectsMissingOrEmptySourceSetIdentity` (`go test -race ./internal/core -run '^(TestNewServiceRejectsEveryMismatchedSplitCapability\|TestNewServiceRejectsMissingOrEmptySourceSetIdentity)$' -count=1`) | **FAIL** (`selected-epic_creation`, `selected_creation_without_provider`, and `selected_creation_without_witness` accepted) | **PASS** (1.215s) |

### Surviving / falsified systemic hypotheses

- **Shared `prepareTaskCreation` or `createTask` altering legacy behavior or joining the wrong namespace:** *Falsified.* `NewTask` retains `validateTaskCreationMode`, `ReadEpics`, `epicExists`, `canonicalEpic`, `prepareTaskCreation`, and the `p.Start` lifecycle branch (`internal/core/service_task.go:924-952`). `CreateTask` passes `validateParent == nil` to `createTask` (`internal/store/create.go:179-181`), preserving identical task and Thread ID uniqueness checks (`ensureCandidateIDUnique`, `ensureTaskIDNotThread`) and `createEntityFile` behavior.
- **Symlink parent, frontmatter `id:` declaration, prefix/substring impostor, or unreadable parent bypassing strict ownership:** *Falsified.* `epicCandidates` (`internal/store/resolve.go:136-158`) requires `e.Type().IsRegular()` (rejecting symlinks) and keys candidates by filename stem; `resolveExactID` (`internal/store/resolve.go:298-314`) requires exact string equality; `parseEpic` (`internal/store/epicstore.go:273-289`) requires strict YAML frontmatter and overwrites `ep.ID` with the filename stem.
- **Denied/read-only preview or write creating roots/directories or reading malformed parent bytes first:** *Falsified.* `createEntityFile` (`internal/store/create.go:71`) checks `s.authorizeMutation()` before `dryRun` calls `prepare()` and before non-dry-run calls `os.MkdirAll(s.root, 0o755)` (`internal/store/create.go:94`). Failed parent validation returns from `prepare()` before `writeNewFileUnlocked` calls `os.MkdirAll(s.tasksDir, 0o755)` (`internal/store/create.go:126`).
- **Finalization error or ID collision causing `Service.NewTaskForSelectedEpic` to remint or retry:** *Falsified.* `NewTaskForSelectedEpic` (`internal/core/selected_task_creation.go:46-68`) does not wrap `CreateTaskForSelectedEpic` in `retryOnConflict`; it mints once in `prepareTaskCreation` and returns the adapter's exact `TaskCreationReceipt` and error (`*core.CreateFinalizationError` joined with any primary error).
- **Documented boundaries preserved without overclaiming:** Dry-run remains a non-reserving point-in-time check, raw editors outside the advisory lock retain the final check-to-create race window, later parent deletion is not blocked, and same-ID delete/recreate is not distinguished (`docs/architecture/planning-data.md:81-85`, `docs/architecture/persistence.md:85-92`).

### Isolation verification attestation

```text
sandbox_path=/private/tmp/isolated-review.QSZzN6
git_dir=/private/tmp/isolated-review.QSZzN6/.git
baseline_commit=76416eeac9ccc97285cde1a1f5a3cd5e3118a7dc
source_blob=20c74d57e1fe8eac409cbcd1253776d302a009e2
source_fingerprint=475eaedfcbe42c81cce980a2c73b9e850c2c8a5e
deliverable=planning/audits/6gndg1vjjv1p-2026-10-19-selected-epic-task-creation-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```
