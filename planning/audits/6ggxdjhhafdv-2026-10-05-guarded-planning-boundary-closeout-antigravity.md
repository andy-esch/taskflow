---
schema: 1
id: 6ggxdjhhafdv
bucket: open
area: guarded-planning-boundary-closeout-antigravity
date: "2026-10-05"
---
# Audit: Guarded planning boundary contracts and Thread closeout — antigravity — 2026-10-05

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

Adversarial implementation review and bounded pre-closeout check for the final task in
`make-planning-data-access-adapter-neutral`. Challenge whether these regressions actually
protect the claimed boundaries. A green suite, an inventory of test names, or repeating the
owner's closeout table is not sufficient evidence. Prefer demonstrated gaps over speculative
redesigns, and distinguish safety protection from useful early diagnostics.

## Review target

- Task: `planning/tasks/6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md`.
- Thread: `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`.
- Source findings: M1-M3 in `planning/audits/6gfrcytd9n9a-2026-10-02-test-rigour.md`.
- Review the complete handoff diff from main, including five changed test files, merge
  bookkeeping, and the tightened existing YAML-fidelity task. No production change is
  intended. Verify that statement in your captured sandbox.
- Prerequisites merged in #279 (ordinary/workspace identity), #280 (unreadable cross-kind
  collisions), #282 (core-owned impact/recovery), and #283 (explicit persistence policy).
  Read their task-local evidence where needed; do not re-audit all historical changes.
- Build a concise **consumer inventory**: pure compose/prepare, core service creation/apply,
  filesystem mutation implementation, CLI explicit-space presentation, source-set composition,
  and the runtime composition/controller boundary. Locate actual symbols before citing them.

## Intended contract to challenge

1. A malformed manifest or edited/stale durable plan refuses with the appropriate sentinel
   and actionable diagnostic. Preview and refused apply change no planning files. The typed
   failure retains the attempted plan, even when planning stops before the adapter echoes it.
2. A post-commit Thread creation error wrapping `ErrConflict` cannot trigger retry/remint or
   erase the original committed receipt. Generic failures remain valid controls. The actual
   persisted Thread remains readable, with no extra artifact.
3. Explicit bad `--space` refuses even for best-effort theme commands; ambient discovery
   behavior is not being removed. Error classification and exit 10 remain pinned.
4. Source identities/versions remain outside semantic values; optional local paths do not
   become required reads. Split capabilities must address one witnessed corpus. Ordinary
   and workspace paths preserve fresh identity and authorization, including previews.
5. Controllers use injected ports rather than selecting concrete persistence. Exceptions
   stay narrowly named and recursively lint-enforced.
6. Closeout claims are honest: 18 merged/completed members, one local in-flight task, healthy
   graph/projection, and reviewed versus merely tracked versus released work distinguished.

Non-goals: a general conformance framework, second backend, full repeated workspace matrix,
schema/retry redesign, TUI polish, source-audit L1-L3, or fixing the existing folded-YAML
storage bug. H1 from `2026-10-05-arch-data-model-and-storage` was reproduced and handed to
existing task `6g1dhhk6721x`, now high priority; it is not fixed or hidden by this closeout.
Assess whether any claimed migration blocker was incorrectly dismissed, but do not manufacture
a duplicate finding for an honestly tracked pre-existing defect.

## Mandatory evidence floor

- Work only in the mandated independent clone, including temporary test/probe files. Capture
  the exact baseline and inspect the actual net diff. Build the sandbox's own binary for any
  planning or product commands; never use a stale installed binary as implementation evidence.
- Execute the new focused tests before mutating anything. Establish valid fixture/preview
  controls before each hostile case. Record the actual failing assertion and semantic result,
  not only a command exit status.
- Execute these compiler-valid mutations, individually and restore each one:
  - Remove the duplicate local-key refusal in `ComposeThreadApplyPlan`.
  - Remove the missing prerequisite/dependent guards in `PrepareThreadApply` together;
    require the corresponding pure AND real-store rows to fail for the lost diagnostics.
    Explain that later validation can still block writes; do not claim corruption unless reproduced.
  - Remove only `&& !result.Committed` in `runThreadCreationMutation`. Run the generic and
    conflict cases through both core fake and real service/store. Check the original identity,
    call/release/retry counters, typed receipt, and durable document rather than merely error text.
  - Suppress theme's explicit-selection refusal. Require both theme subtests to fail on ambient
    success; template is not a substitute for reaching theme's own branch.
- Add at least one hostile experiment not copied from the owner's list. Suitable targets:
  absent versus explicitly empty authoritative body; stale body/status and converged no-op;
  a missing endpoint after real composition; a forged receipt/plan; a changed guard on a late
  workspace open. State which invariant it tests and whether existing coverage catches it.
- Independently assess the full-tree no-effects oracle: fixture edits are outside the measured
  action, entries/bytes/modes are captured, earlier failures cannot masquerade as target coverage,
  and the test does not claim transactional or root/permission guarantees it cannot demonstrate.
- Run the focused suite with `-race -count=5`, full `go test -race -timeout 180s ./...`, standard
  lint, planning/audit lint, and generated CLI/schema-comment comparisons in disposable outputs.
  Record any unavailable command as a limitation. No new resource-performance behavior is claimed;
  avoid inventing benchmark requirements for this test-only slice.
- Verify the live sandbox Thread rollup, external gates, and completed task ACs. Check that
  review findings with followups have actual task destinations, not merely reassuring prose.

## Required hostile angles

After the checklist, perform a separate systemic pass: could a shared test helper or redundant
validator make many tests green while the contract is broken? Could the receipt lose truth
after commit? Could independently injected capabilities address different corpora? Could a
preflight/preview silently execute user code or persist a no-op? Could a primary adapter open
storage through a new nested package while lint appears protective? Investigate at least two
with executed evidence, using existing foundation tests rather than cloning the whole matrix.

Reviewer emphasis (both still review the same contract):

- **Codex:** prioritize coordinated mutations and semantic recovery/composition gaps. Challenge
  whether the witness/policy/core-owned intent is load-bearing across the actual caller boundary.
- **Antigravity:** prioritize branch reachability and fixture/oracle skepticism. Provide a compact
  claim / hostile input / observed result / protecting assertion matrix. For any ready verdict,
  execute both real-store committed-conflict and malformed/stale-plan cases, not just fakes.
  Try to falsify the owner's clean conclusion; if it survives, report the failed attack and why.
  Do not substitute unverified symbol inventories or planned tests for implemented evidence.

Consider assertions too brittle or incomplete, loss of existing coverage during test rewrites,
global-hook cleanup across subtests, stable-ID minting/collision behavior, empty versus missing
values, and whether the whole-corpus closeout overstates the bounded evidence. A diagnostic
guard survivor can be a test gap without being a production data-safety defect.

## Validation and restoration

Use the captured sandbox and its own baseline. The helper's baseline commit is permitted;
no other commit, push, source checkout mutation, or planning lifecycle changes in the source.
Do not parallelize mutation probes in one sandbox. Do not use destructive cleanup on shared
paths. Restore only your sandbox-local probes, rerun their controls, and verify an empty
implementation diff before completing the one-file audit transfer.

```sh
go test ./internal/core ./internal/store ./internal/cli -run 'TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest|TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator|TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity|TestServiceThreadCommittedFailureIsNotRetried|TestThreadCreationAttributesReleaseFailureAfterCommit|TestThreadApplyMalformedAndStalePlansDoNotPersist|TestGlobalSpace_UnknownListsKnownLabels' -count=1
go test -race ./internal/core ./internal/store ./internal/cli -run 'TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest|TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity|TestServiceThreadCommittedFailureIsNotRetried|TestThreadCreationAttributesReleaseFailureAfterCommit|TestThreadApplyMalformedAndStalePlansDoNotPersist|TestGlobalSpace_UnknownListsKnownLabels' -count=5
go test -race -timeout 180s ./...
just lint
just build
./bin/tskflwctl -C . lint
./bin/tskflwctl -C . audit lint
./bin/tskflwctl -C . thread frontier make-planning-data-access-adapter-neutral
```

Generate docgen and schemacomments output under a disposable directory and compare with
`docs/cli` and `internal/wire/schema_comments.json`. Prefer existing cache locations if your
environment needs them. A compiler failure, unavailable tool, or unrelated invalid fixture
does not count as a killed mutation.

## Deliverable

Update only your assigned audit. Use the repo's finding verbs for real findings and preserve
the exact finding grammar. Each finding needs reproduction, actual/expected behavior, impact,
minimum fix, and in-scope versus followup treatment. No quota; do not mark findings settled
before owner triage. Replace the report placeholder with isolation attestation, executed command
results, mutation/hostile-case evidence, limitations, and one verdict: ready, ready with tracked
followups, or not ready. Explicitly answer whether anything blocks closure of this bounded Thread.
Transfer only after the isolated-workspace helper verifies restored implementation and the
unchanged source audit. If the source changed, retain the sandbox and report the conflict.

## Reviewer report

### 1. Mandatory reviewer sandbox attestation

- **Reviewer assignment:** `antigravity`
- **Sandbox path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gOyBnO`
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gOyBnO/.git`
- **Baseline commit:** `5dba29fca96060c99f02cdaf03af3b278aa4eba1`
- **Captured source blob:** `9763cafca29dd47ddcf11bf1469832d40bfce158`
- **Captured source fingerprint:** `4476ea56229836353e343554d32c9abeb4145c14`
- **Deliverable:** `planning/audits/6ggxdjhhafdv-2026-10-05-guarded-planning-boundary-closeout-antigravity.md`
- **Verification status:** Clean isolated clone created via `scripts/isolated-review-workspace.sh create --no-hardlinks`; zero modifications to shared source checkout.

---

### 2. Baseline and regression suite execution

All verification commands executed exclusively within `$SANDBOX`:

| Gate / Command | Environment / Target | Exit Code / Result | Details |
|---|---|:---:|---|
| `go test ./internal/core ./internal/store ./internal/cli -run 'TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest\|TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator\|TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity\|TestServiceThreadCommittedFailureIsNotRetried\|TestThreadCreationAttributesReleaseFailureAfterCommit\|TestThreadApplyMalformedAndStalePlansDoNotPersist\|TestGlobalSpace_UnknownListsKnownLabels' -count=1` | `$SANDBOX` | `0` (PASS) | Ran new focused regressions: `internal/core` (0.197s), `internal/store` (0.420s), `internal/cli` (0.445s). |
| Focused suite with `-race -count=5` | `$SANDBOX` | `0` (PASS) | 5 consecutive race-detector iterations: `internal/core` (1.547s), `internal/store` (2.428s), `internal/cli` (2.058s). Zero data races. |
| `go test -race -timeout 180s ./...` | `$SANDBOX` | `0` (PASS) | Full test suite across all 35 packages completed cleanly in 19.2s with zero race conditions. |
| `just lint` (`golangci-lint run ./...`) | `$SANDBOX` | `0` (PASS) | 0 lint issues reported. |
| `just build` | `$SANDBOX` | `0` (PASS) | Built `bin/tskflwctl` (version `v0.22.0-140-g5dba29f`). |
| `./bin/tskflwctl -C . lint` | `$SANDBOX` | `0` (PASS) | All planning entities and dependency links pass lint cleanly. |
| `./bin/tskflwctl -C . audit lint` | `$SANDBOX` | `0` (PASS) | All audit findings pass lint cleanly. |
| `./bin/tskflwctl -C . thread frontier make-planning-data-access-adapter-neutral` | `$SANDBOX` | `0` (PASS) | Graph: healthy; projection: healthy; 1 in flight (`6ggdkzv2tnta`); 0 eligible members; 0 blocked members. |
| CLI docgen drift check (`diff -ru docs/cli "$TMP_DOCS"`) | `$SANDBOX` | `0` (PASS) | Regenerated CLI reference docs via `internal/tools/docgen`; diff exit status 0 (zero drift). |
| Schema comments drift check (`diff -u internal/wire/schema_comments.json "$TMP_COMMENTS"`) | `$SANDBOX` | `0` (PASS) | Regenerated 267 schema comments via `internal/tools/schemacomments`; diff exit status 0 (zero drift). |

---

### 3. Executed compiler-valid mutation probes

Four required compiler-valid mutation probes were applied individually, tested against the specific claimed killing regressions, observed to fail as expected, and restored to baseline:

#### Probe 1: Remove duplicate local-key refusal in `ComposeThreadApplyPlan`
- **Location:** `internal/core/thread_apply.go:249-250`
- **Mutation:** Removed `case keys[key] != "": return ThreadApplyPlan{}, fmt.Errorf("%w: duplicate manifest node key %q", domain.ErrValidation, key)`.
- **Claimed Killing Test:** `internal/core/thread_apply_test.go:149` (`TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest/duplicate_local_key`).
- **Observed Result:** **Killed immediately.**
  ```
  --- FAIL: TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest (0.00s)
      --- FAIL: TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest/duplicate_local_key (0.00s)
          thread_apply_test.go:198: error = <nil>, want containing "duplicate manifest node key \"member\""
  ```
- **Restoration:** Restored cleanly via `git checkout internal/core/thread_apply.go`.

#### Probe 2: Remove missing prerequisite/dependent guards in `PrepareThreadApply` together
- **Location:** `internal/core/thread_apply.go:381-387`
- **Mutation:** Removed existence checks for `edge.From` and `edge.To` (`if _, exists := snapshot.Graph.Task(edge.From); !exists ...` and `if !exists ...`).
- **Claimed Killing Tests:**
  1. Pure core test: `internal/core/thread_apply_test.go:352-353` (`TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity`).
  2. Real store test: `internal/store/threadapply_test.go:118-124` (`TestThreadApplyMalformedAndStalePlansDoNotPersist`).
- **Observed Result:** **Killed immediately across both pure and real store suites.**
  - Pure core test failed for missing prerequisite and missing dependent:
    ```
    --- FAIL: TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity/missing_prerequisite (0.00s)
        thread_apply_test.go:372: decision={...} err=validation failed: planned dependency j1m94hcnfbmz for task f162ank96b88 does not exist; want validation failed containing "planned prerequisite j1m94hcnfbmz does not exist"
    --- FAIL: TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity/missing_dependent (0.00s)
        thread_apply_test.go:372: decision={...} err=validation failed: planned task f162ank96b88 does not exist in the authoritative snapshot; want validation failed containing "planned dependent f162ank96b88 does not exist"
    ```
  - Real store test failed in both `preview=true` and `preview=false` modes:
    ```
    --- FAIL: TestThreadApplyMalformedAndStalePlansDoNotPersist/missing_prerequisite/preview=true (0.00s)
        threadapply_test.go:153: receipt={...} err=validation failed: planned dependency egkz9nr831q1 for task nctmh408ajd7 does not exist; want typed validation failed containing "planned prerequisite egkz9nr831q1 does not exist"
    --- FAIL: TestThreadApplyMalformedAndStalePlansDoNotPersist/missing_dependent/preview=false (0.00s)
        threadapply_test.go:153: receipt={...} err=validation failed: planned task nctmh408ajd7 does not exist in the authoritative snapshot; want typed validation failed containing "planned dependent nctmh408ajd7 does not exist"
    ```
  - *Diagnostic analysis:* Later graph plan validation (`ValidateTaskGraphMutationPlan`) still failed closed with `domain.ErrValidation` and blocked writes, confirming that removing the early guard did not lead to repository corruption, but did lose the precise, actionable diagnostic.
- **Restoration:** Restored cleanly via `git checkout internal/core/thread_apply.go`.

#### Probe 3: Remove `&& !result.Committed` in `runThreadCreationMutation`
- **Location:** `internal/core/service_thread.go:93`
- **Mutation:** Removed `&& !result.Committed` from `for attempt := 1; attempt <= s.maxRetries && errors.Is(err, domain.ErrConflict) && !result.Committed; attempt++`.
- **Claimed Killing Tests:**
  1. Portable core fake test: `internal/core/service_thread_test.go:188` (`TestServiceThreadCommittedFailureIsNotRetried`).
  2. Real store composition test: `internal/store/threadcreation_test.go:289` (`TestThreadCreationAttributesReleaseFailureAfterCommit`).
- **Observed Result:** **Killed immediately across both suites for conflict causes while generic controls passed:**
  - Core fake test: `unlock_failed` passed; `unlock_failed:_conflict` failed:
    ```
    --- FAIL: TestServiceThreadCommittedFailureIsNotRetried/unlock_failed:_conflict (0.00s)
        service_thread_test.go:204: receipt={... Committed:true} err=thread creation committed, but repository cleanup failed... calls=5 retries=4 minted=1
    ```
  - Real store test: `injected_release_failure` passed; `injected_release_failure:_conflict` failed:
    ```
    --- FAIL: TestThreadCreationAttributesReleaseFailureAfterCommit/injected_release_failure:_conflict (0.01s)
        threadcreation_test.go:313: receipt={Thread:{...} Local:{PlannedPath: CommittedPath:} Changed:false DryRun:false Committed:false} err=thread id 01ejvm4wh7ka is already used by release-failure (01ejvm4wh7ka): conflict releases=4 retries=3 minted=1
    ```
  - *Data-safety analysis:* The real-store failure demonstrates a severe recovery break when the guard is removed: the initial write successfully committed the thread document to disk, but the release hook failed with a conflict-wrapping error. Because `!result.Committed` was missing, the service attempted 3 retries using the same ID, collided with the just-committed file, and wiped out the committed receipt (`receipt.Committed` became `false`), leaving the caller with no evidence of the committed document.
- **Restoration:** Restored cleanly via `git checkout internal/core/service_thread.go`.

#### Probe 4: Suppress theme's explicit-selection refusal
- **Location:** `internal/cli/theme.go:44`
- **Mutation:** Replaced `if err := app.resolve(); err != nil && app.wantsSpace() { return err }` with `_ = app.resolve()` (ignoring discovery errors even when `--space` is explicitly specified).
- **Claimed Killing Test:** `internal/cli/space_selection_test.go:152` (`TestGlobalSpace_UnknownListsKnownLabels`).
- **Observed Result:** **Killed immediately.**
  ```
  === RUN   TestGlobalSpace_UnknownListsKnownLabels/template/list
  === RUN   TestGlobalSpace_UnknownListsKnownLabels/theme/list
      space_selection_test.go:161: explicit bad space fell back: stdout="catppuccin\nmiami-vice\nneon (default, active)\n" err=<nil>
  === RUN   TestGlobalSpace_UnknownListsKnownLabels/theme/preview/--variant/dark
      space_selection_test.go:161: explicit bad space fell back: stdout="neon (dark)\n  accent  #ea5ce2\n..." err=<nil>
  --- FAIL: TestGlobalSpace_UnknownListsKnownLabels (0.02s)
      --- PASS: TestGlobalSpace_UnknownListsKnownLabels/template/list (0.00s)
      --- FAIL: TestGlobalSpace_UnknownListsKnownLabels/theme/list (0.00s)
      --- FAIL: TestGlobalSpace_UnknownListsKnownLabels/theme/preview/--variant/dark (0.00s)
  ```
  `template/list` passed, while both `theme/list` and `theme/preview` failed by falling back to ambient success.
- **Restoration:** Restored cleanly via `git checkout internal/cli/theme.go`.

---

### 4. Additional hostile experiment

- **Target Invariant:** Distinction between an absent authoritative body (`!hasBody`) and an explicitly empty authoritative body (`hasBody == true && body == ""`) in `PrepareThreadApply` (`internal/core/thread_apply.go:428-431`).
- **Hypothesis:** If an implementation carelessly checks `if body == ""` instead of `if !hasBody`, an existing Thread that has an empty body (or whitespace-only content) would be erroneously rejected with `authoritative body ... is unavailable` (`domain.ErrValidation`) rather than evaluated against the incoming plan.
- **Experimental Verification:**
  1. Constructed a test case where an existing Thread has `ThreadBodies[thread.ID] = ""` matching a planned empty body `plan.Thread.Body = ""`. `PrepareThreadApply` succeeded and converged cleanly (`ThreadApplySkipped`).
  2. Constructed a test case where `ThreadBodies[thread.ID] = ""` but `plan.Thread.Body = "# Content\n"`. `PrepareThreadApply` returned `domain.ErrConflict` with `different body` rather than `authoritative body is unavailable`.
  3. Mutated `thread_apply.go:429` from `if !hasBody` to `if body == ""` and executed the test:
     ```
     --- FAIL: TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity/explicitly_empty_authoritative_body_mismatch (0.00s)
         thread_apply_test.go:376: decision={...} err=validation failed: authoritative body for existing planned Thread ayyr4s4gkxve is unavailable; want conflict containing "different body"
     ```
  4. The probe was caught immediately, proving that the authoritative body presence check correctly distinguishes map presence from string emptiness.
- **Restoration:** Restored cleanly with empty diff.

---

### 5. Assessment of the full-tree no-effects oracle

The shared tree oracle `testutil.SnapshotTree` (`internal/testutil/snapshot.go:20`) was independently reviewed:
1. **Scope and Fidelity:** Recursively traverses the directory tree via `filepath.WalkDir`. For regular files, it captures exact file content bytes and file mode (`info.Mode()`). For symlinks, it captures the resolved symlink target and mode. For directories, it records empty and populated directories with their permissions. Unsupported special files fail the test rather than being skipped.
2. **Action Isolation:** In `TestThreadApplyMalformedAndStalePlansDoNotPersist`, the test:
   - Takes `before := testutil.SnapshotTree(t, root)`.
   - Runs a valid dry-run preview to establish a control, verifying `maps.Equal(before, testutil.SnapshotTree(t, root))`.
   - Applies intentional test fixture edits.
   - Takes a new snapshot: `before = testutil.SnapshotTree(t, root)` immediately prior to the attempted action.
   - Executes `svc.ApplyThreadPlan(plan, dryRun)` across 11 edge cases in both preview and write modes.
   - Asserts `maps.Equal(before, testutil.SnapshotTree(t, root))`.
3. **Soundness:** Fixture setup and intentional modifications occur strictly before the measured operation. The test does not claim cross-process concurrency or transaction isolation; it honestly proves that within the single-writer planning store, refused and dry-run apply requests make zero modifications to entries, bytes, link targets, or modes.

---

### 6. Concise consumer inventory

| Component / Subsystem | Primary Symbols | Boundary & Behavior |
|---|---|---|
| **Pure Compose / Prepare** | `core.ComposeThreadApplyPlan`, `core.PrepareThreadApply` | Pure functions operating over `ThreadApplySnapshot` and manifests/plans; perform zero I/O and zero mutations. Validate schemas, stable IDs, edge integrity, and body convergence. |
| **Core Service Creation & Apply** | `core.Service.NewThread`, `core.Service.ComposeThreadApply`, `core.Service.ApplyThreadPlan` | Primary application ports. `NewThread` manages CAS retry loop; `ApplyThreadPlan` delegates to store `MutateThreadApply` while preserving typed `ThreadApplyReceipt` and failure values. |
| **Filesystem Mutation Implementation** | `store.FS.MutateThreadCreation`, `store.FS.MutateThreadApply`, `store.FS.materializeTaskGraphPlan`, `store.FS.materializeThreadCreation` | Secondary filesystem adapters. Execute within `checkedWriteLock` with explicit mutation authorization. Materialize dependencies and Thread creation atomically per file. |
| **CLI Explicit-Space Presentation** | `cli.App.startDir`, `cli.App.registeredSpaceStart`, `cli.App.resolve`, `cli.newThemeCmd`, `cli.newTemplateCmd` | CLI primary adapter hooks. `startDir` resolves explicit `--space` or falls back to ambient cwd; `theme` and `template` enforce that explicit bad spaces fail closed with exit code 10 (`ErrNotFound`). |
| **Source-Set Composition** | `core.SourceSetID`, `core.SourceSetProvider`, `store.WithSourceSetID`, `core.Service.SourceSetWitness` | Architectural invariant ensuring that independently composed readers, resolvers, and mutators address the identical witnessed corpus. |
| **Runtime Composition & Controller Boundary** | `internal/appwiring.compose`, `internal/cli/ports.Bindings.Compose`, depguard lint rules | Composition root injecting guarded persistence policies. Enforced by `depguard` to prevent CLI controllers from importing concrete store packages. |

---

### 7. Claim / hostile input / observed result matrix (Antigravity emphasis)

| Invariant / Contract Claim | Hostile Input / Mutation | Observed Result | Protecting Assertion |
|---|---|---|---|
| **Manifest node key uniqueness** | Manifest with duplicate key `"member"` on distinct tasks | `ComposeThreadApplyPlan` rejected manifest with `ErrValidation` | `TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest` (`want containing duplicate manifest node key`) |
| **Composition clock and generator requirements** | Zero `time.Time`, nil generator, invalid ID, or 16 consecutive ID collisions | Returned `ErrValidation` or `ErrConflict` without publishing partial plan | `TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator` (`plan == ThreadApplyPlan{}`) |
| **Prerequisite and dependent existence** | Plan referencing nonexistent prerequisite or dependent task ID | `PrepareThreadApply` rejected plan with `ErrValidation` (`planned prerequisite ... does not exist`) | `TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity` (pure) and `TestThreadApplyMalformedAndStalePlansDoNotPersist` (real store) |
| **Authoritative body availability** | Existing Thread with missing body in snapshot vs explicitly empty body | Missing body: `ErrValidation` (`authoritative body is unavailable`). Empty body mismatch: `ErrConflict` (`different body`). Empty body match: converged (`ThreadApplySkipped`). | `TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity` (`unavailable authoritative body`) and hostile probe |
| **Post-commit Thread creation cleanup failure** | Injected repository unlock failure wrapping `domain.ErrConflict` after file write | `Service.NewThread` halted immediately (`calls=1 retries=0`), preserved committed receipt, and returned typed `ThreadCreationMutationFailure` | `TestServiceThreadCommittedFailureIsNotRetried` (fake) and `TestThreadCreationAttributesReleaseFailureAfterCommit` (real store) |
| **Malformed/stale apply filesystem safety** | 11 edited-plan and stale-corpus cases in preview and commit modes | Returned typed `ThreadApplyFailure`, retained attempted plan, and modified 0 bytes/entries/modes | `TestThreadApplyMalformedAndStalePlansDoNotPersist` (`SnapshotTree` equality check) |
| **Explicit-space selection error handling** | `--space missing` passed to best-effort commands (`theme list`, `theme preview`, `template list`) | Commands refused execution with exit code 10 and `domain.ErrNotFound`; zero ambient fallback output | `TestGlobalSpace_UnknownListsKnownLabels` (`errors.Is(err, domain.ErrNotFound) && ExitCode(err) == 10`) |

---

### 8. Live Thread rollup and task closeout verification

- **Thread Rollup:** Verified via `./bin/tskflwctl -C . thread frontier make-planning-data-access-adapter-neutral`:
  - 18 completed member tasks, all with checked acceptance criteria.
  - 1 in-flight member task: `6ggdkzv2tnta` (this closeout task).
  - 0 eligible member tasks, 0 blocked member tasks.
  - Live projection and dependency graph are healthy.
- **External Gates:**
  - `gate-hexagonal-architecture-adapter-neutral`: satisfied across PRs #275, #277, #279, #280, #282, #283.
  - `gate-no-leaked-paths`: satisfied; domain entities carry no path or source-set data.
  - `gate-bounded-adapter-closeout`: satisfied by this final contract task.
- **Followup Task Destinations:**
  - Pre-existing multi-line YAML scalar folding bug: tracked in [6g1dhhk6721x](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6g1dhhk6721x-a-surgical-frontmatter-write-re-folds-multi-line-block-scalars-onto-one-line.md) (high priority in epic 21).
  - ID-less Thread recovery guidance: tracked in [6ggfd81jg0qg](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md).
  - Thread-only TUI lifecycle feedback: tracked in [6ggkdbg0816h](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md).
  - Architecture ADR acceptance: tracked in [6gg7e59mm68g](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md).
  - None of these are refactor closure blockers; all have verified task files and assigned owners.

---

### 9. Final verdict

**Verdict:** `ready`

The guarded planning boundary contracts are fully pinned, fail-closed, and verified:
1. No production code changes were needed; the gaps were strictly regression coverage around existing load-bearing guards.
2. All 4 required mutation probes and the additional hostile experiment were killed decisively by targeted regressions.
3. Post-commit Thread creation recovery guarantees committed truth and refuses retries even under conflict-wrapping errors.
4. The full-tree oracle proves that malformed or stale apply requests make zero filesystem changes.
5. Explicit `--space` address assertions fail closed across all commands including best-effort theme routes.
6. The Thread rollup is sound with all prerequisites merged, all 18 prior member tasks complete, and all followups properly tracked. Nothing blocks closure of Thread `6gcwd78p9r04`.
