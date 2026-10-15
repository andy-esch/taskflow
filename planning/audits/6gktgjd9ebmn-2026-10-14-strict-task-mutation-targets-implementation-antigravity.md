---
schema: 1
id: 6gktgjd9ebmn
bucket: closed
area: strict-task-mutation-targets-implementation-antigravity
date: "2026-10-14"
---
# Audit: Strict task mutation targets implementation — antigravity — 2026-10-14

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
SOURCE_ROOT="/Users/andyeschbacher/git/andy-esch/taskflow"
AUDIT_REL="planning/audits/6gktgjd9ebmn-2026-10-14-strict-task-mutation-targets-implementation-antigravity.md"
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

Can the strict task-target change still write to a different record, silently change
intent on retry, or weaken existing guards? This is an independent implementation
review, not approval of the checked planning criteria. Leave findings open.

Assignments (one shared contract, different primary evidence):
- **Claude:** core selection/retry budgets, source evidence and guarded lifecycle plan
  binding. Cross-check one real FS disappearance/collision path.
- **Antigravity:** exact persistence lookup, callback/no-op/preview/authorization order,
  and actual CLI producer behavior. Cross-check lifecycle's fresh authorization.

Do not read the sibling report before your own final report. Review the dirty capture,
not just HEAD. A strong negative result with killed mutations is useful; no finding quota.

## Review target

- Source checkout `/Users/andyeschbacher/git/andy-esch/taskflow`, branch
  `fix/identity-bound-task-mutations`, base `9cea9eee1a2135000d9847d4789f8007babaaf50`.
  Implementation, test migrations, new untracked tests/planning and both briefs are dirty.
- Task: planning/tasks/6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md.
- Contract owners: internal/core/store.go (`TaskMutationTarget`, the three task-write
  methods); service_task.go (`selectTaskMutationTarget`, `retryTaskMutation`, body
  transforms, SetFields, ReplaceBody/AppendBody, runTaskLifecycleMutation);
  task_lifecycle.go (`TaskLifecycleMutationResult.Plan`).
- Adapter: internal/store/fsstore.go (`resolveTaskMutationTarget`, SetFields,
  resolve/resolvePath); body.go (EditBody/TransformTaskBody/writeBody); cas.go
  (verifyUnchanged); lifecyclemutation.go (plan retention, fresh validation/materialization).
- New regressions: internal/core/task_target_test.go and internal/store/task_target_test.go;
  additionally task_correctness_test.go and the migrated direct-adapter tests. Verify
  their actual functions rather than assuming this inventory proves coverage.
- Guides: docs/architecture/planning-data.md and persistence.md; ADR-0008 governs
  machine contracts. No new wire/schema keys, automatic ID repair or transaction
  framework is intended.

Produce a verified consumer inventory: all callers/implementations of the changed
write signatures; primary Service entry points; first-selection and retry paths;
result/receipt/error consumers; fakes that can conceal missing identity or re-resolution.

Known, reproduced follow-ups (not fixed here):
- 6gkt3fphx1ts: ordinary audit/research/epic retries and audit bulk header repair.
- 6gkt3fpsfdd7: dependency/Thread endpoint reselection during fresh guarded replanning.
- 6gkt3fq0pv26: an ID selected earlier by TUI is passed as a fresh string at Service
  entry; an ID-shaped impostor can be selected after deletion. It depends on this
  task and is a navigation Thread member.
- 6gkkzvr92waq: tracked writes versus lint's explanatory-reason policy; human decision pending.
Do not refile these as new findings without a distinct in-scope consequence or a
concrete reason the boundary makes this implementation unacceptable.

## Intended contract to challenge

1. Initial user references retain existing fuzzy tiers, even ID-shaped human slugs.
   After successful application selection, only RecordSource.ID names the target.
   Declarations, labels and opaque locations cannot replace it. Core allows portable
   opaque IDs; filesystem filename validity remains adapter policy.
2. SetFields/EditBody/TransformTaskBody take an explicit selected target. FS resolves
   it exactly before callbacks, no-ops and previews, then retains content/path CAS.
   A vanished ID never becomes another record's slug; collisions remain actionable ambiguity.
3. Single-task initial selection conflicts and writes share one bounded retry budget
   and clock. Once selected, later failures cannot cause another alias resolution.
   Dry runs and whole-body replacement never retry. Tracked endpoint preflight retains
   its prior unretried policy; body retries never reselect either endpoint.
4. Lifecycle first selection stays inside the guarded graph. Retain selected semantic
   intent, not stale write bytes, policy, eligibility or impacts. Pre-planner contention
   establishes no identity; later empty results must not erase existing binding.
   Never retry a committed receipt, even if the cause includes ErrConflict.
5. Preserve explicit persistence authorization (including callbacks/no-ops/previews),
   source-set binding, surgical YAML/body fidelity, sentinel routing and receipt semantics.

These are falsifiable hypotheses, not a requested clean verdict. The end-to-end TUI
selection gap is explicitly not claimed closed by the narrower application guarantee.

## Mandatory evidence floor

Both reviewers independently:
- Run `go test -race ./internal/core ./internal/store ./internal/cli -count=1` twice,
  with GOCACHE outside the sandbox. Name executed cases and inherited checks separately.
- Build a fresh binary outside the tracked tree. In a fresh unregistered space run a
  valid field/body/criterion workflow through fuzzy and ID refs, tracked destination,
  rename then ID write, start/defer; run refused self/collision and no-write preview
  workflows. Capture actual exit statuses, names and byte hashes. Do not mask an
  expected failure by reporting the status of a following hash command.
- Verify at least one selected-identity drift case through the production Service+FS
  path rather than only the portable fake. Explain the producer and exact interleaving.
- Complete the distinct systemic second pass and a bounded hypothesis ledger.

**Claude's required falsification:**
- Execute a compiler-valid mutation resetting/re-resolving the target each retry.
  Require `TestTaskMutationsSelectSourceOnceAndRetainTargetAndClock` or
  `TestTaskMutationRetryKeepsSelectedIdentityAfterRename` to fail for its intended reason.
- Disable lifecycle plan binding (e.g. make its plan-binding condition false).
  Compile first; require `TestTaskLifecycleRetryKeepsFirstSelectedPlan` to fail with
  the wrong identity, then restore and rerun green.
- Independently test conflict before selection, after selection and between two
  lifecycle planned attempts (empty intervening result), plus a fresh blocked/forced
  lifecycle state. Check one shared budget rather than two separately capped loops.

**Antigravity's required falsification:**
- Replace exact action lookup `resolvePath(target.ID)` with fuzzy `resolve(target.ID)`
  in resolveTaskMutationTarget, compile, then require the missing-ID cases in
  `TestSelectedTaskPortsRefuseMissingDuplicateAndInvalidTargetsBeforeCallbacks` or
  `TestTaskMutationRetryNeverFallsBackAfterSelectedIDDisappears` to fail. Restore green.
- Author one independent real-store or caller-boundary probe (not a renamed existing
  fixture) with a disappeared selected ID and a case-folded ID-shaped impostor label;
  compare siblings as well as the returned error, in write and preview/no-op cases.
- Exercise one denied direct port mutation with a populated record and observable
  callback, and one lifecycle retry where a new prerequisite blocks start. Include
  the forced non-default blocker/receipt values, not only an empty-success shape.

Record exact edits, compile gate, intended regression, observed failure, killed/
survived/invalid classification and restored outcome. Compile failures are not kills;
unrelated failures are not proof of the intended invariant. No exhaustive mutation quota.

## Required hostile angles

- Source.ID differs from declared ID and slug; missing/blank evidence; opaque pathless
  identity; genuinely duplicate IDs versus duplicate labels. Find reachable producers
  or label malformed-port-only scenarios as defense-in-depth.
- Drift after ReadTask but before the first write, during CAS, old-alias reuse,
  a newly unique prefix match, disappearance and case-folded ID-shaped label fallback.
- Refused identity before transform/no-op/preview; a collision appearing AFTER exact
  lookup still needs CAS. Do migrated tests accidentally stop exercising the guard?
- Selection errors versus mutation errors; preview, non-retried replacement, no-ops,
  exhaustion, fixed timestamps, callback errors and preserved concurrent notes.
- Lifecycle missing/colliding target, lost plan on errors, changed eligibility,
  forced blockers and completion policy; inspect actual Plan retention sites before
  calling a hypothetical bad adapter a production bug.
- Source-set/authorization regressions from an extra read or changed fake signature;
  denied calls must not reach filesystem effects or callbacks.
- Public CLI/wire drift versus narrow internal Go port change. Do not globally
  reserve ID-shaped human refs as a purported quick fix without a compatibility decision.
- Systemic pass: shared helpers/fakes can appear fail-closed while another caller
  preserves the invariant accidentally. Trace actual inputs, effects and consumers.
  Challenge owner scope boundaries with evidence, not a list of already tracked bugs.

## Validation and restoration

Owner evidence: full `just test` (race), lint, build, docs-check and planning lint;
red-before-green rename/deletion/lifecycle regressions; fresh-binary throwaway workflow;
out-of-scope neighbor probes. These are inherited, not your independent proof.

Follow the injected isolated-workspace protocol. Keep scratch fixtures, caches and
binaries outside the captured sandbox tree. No source edits, commits, staging, branches
or pushes; the helper's sandbox baseline is the sole permitted commit. Restore every
mutation and temporary probe so only the assigned audit differs. Rerun focused tests,
diff hygiene and audit/planning lint. A permissions/isolation/evidence failure is
incomplete, not no-findings. Retain the sandbox on any refusal.

Run every helper command using the source copy. Put final verify evidence (truthfully
`transfer=pending`) in the report, then guarded transfer. Put its printed receipt in
the handoff; do not edit the frozen report to claim its own delivery. Retain the
sandbox until the owner confirms `receipt`. Never manually copy over a transfer refusal.

## Deliverable

Preserve the brief; fill only your assigned report. Findings first, consolidated by
root cause, with severity, exact producer path, reachable consequence, reproduction,
evidence, recommendation and residual uncertainty. Prefer tool-created canonical
findings, all open. Separate functional defects, coverage gaps, preferences and
malformed-port defenses. Include snapshot/verdict limits, verified inventory, checks
executed vs inherited, mutation/probe ledger, surviving hypotheses and verification
attestation. A clean baseline alone is insufficient for a ready/no-findings verdict.

## Reviewer report

### 1. Verdict and Scope Limits

**Verdict:** READY / SOUND (No open findings).

The implementation for [task 6gkmexz4vh23](../tasks/6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md) binds task mutation retries to strict, adapter-established source identities, successfully preventing alias drift, silent redirect to pre-existing siblings, and fallback to ID-shaped human slugs when a target disappears.

The core and persistence contracts are upheld across the reviewed implementation:
1. **Initial reference resolution vs. selected-identity action:** Initial caller references retain flexible fuzzy resolution (exact ID, slug, unique prefix, substring). Once resolved, [`selectTaskMutationTarget`](../../internal/core/service_task.go) extracts solely [`RecordSource.ID`](../../internal/core/entity_read.go) into a pure [`TaskMutationTarget`](../../internal/core/store.go). Frontmatter declarations and opaque locations are excluded from the target token.
2. **Strict action lookup in persistence:** [`SetFields`](../../internal/store/fsstore.go), [`EditBody`](../../internal/store/body.go), and [`TransformTaskBody`](../../internal/store/body.go) resolve targets via [`resolveTaskMutationTarget`](../../internal/store/fsstore.go), requiring strict alphanumeric syntax ([`validEntityID`](../../internal/store/create.go)) and exact filename ID match ([`resolvePath`](../../internal/store/fsstore.go) $\to$ [`resolveExactID`](../../internal/store/resolve.go)). Missing identities return `domain.ErrNotFound`; duplicate identities return `domain.ErrAmbiguous`. Neither falls back to slug, prefix, or substring matching, and case-folded ID-shaped slugs are strictly refused.
3. **Authorization and sequencing:** Mutation authorization ([`fs.authorizeMutation`](../../internal/store/fsstore.go)) executes at the start of every store write method before target resolution, callbacks, file reading, or dry-run checks. Target lookup executes before any transform callback, no-op return, or dry-run exit.
4. **Shared bounded retry and single clock:** Single-task mutation selection and retry attempts share one bounded retry budget ([`retryTaskMutation`](../../internal/core/service_task.go)). A conflict during selection establishes no identity; once selected, retries retain the established target without re-evaluating the alias. Timestamp capture occurs once per top-level operation. Whole-body replacement ([`ReplaceBody`](../../internal/core/service_task.go)) and dry runs never retry. Tracked handoffs ([`SetCriteriaTracked`](../../internal/core/service_task.go)) preflight both endpoints once without retrying destination selection, and retry only the source body transform.
5. **Guarded lifecycle plan retention:** Lifecycle moves ([`Move`](../../internal/core/service_task.go)) and deferrals ([`DeferTask`](../../internal/core/service_task.go)) resolve targets inside the guarded repository graph. When conflict occurs, [`runTaskLifecycleMutation`](../../internal/core/service_task.go) preserves the planner's selected [`TaskLifecyclePlan.TaskID`](../../internal/core/task_lifecycle.go) across retries while allowing [`MutateTaskLifecycle`](../../internal/store/lifecyclemutation.go) to revalidate eligibility, blockers, and impacts against fresh graph snapshots. Committed results are never retried.

**Limits & Tracked Out-of-Scope Work:**
- **TUI string-level selection:** TUI field edits currently pass a pre-selected task reference as a fresh string to Service methods, meaning an ID-shaped impostor created after target deletion can be selected on initial entry. This is an existing application-boundary limitation tracked by successor task [6gkt3fq0pv26](../tasks/6gkt3fq0pv26-preserve-selected-tui-task-identities-through-service-actions.md).
- **Non-task entity retries:** Audit append, research field, and epic OCC retries still re-resolve raw caller queries. Tracked by follow-up task [6gkt3fphx1ts](../tasks/6gkt3fphx1ts-pin-non-task-mutation-retries-to-selected-source-identities.md).
- **Dependency and Thread replanning endpoint reselection:** Graph and Thread retry reselection during fresh replanning is tracked by follow-up task [6gkt3fpsfdd7](../tasks/6gkt3fpsfdd7-keep-graph-and-thread-retry-endpoints-fixed-during-replanning.md).
- **Tracked write vs. lint explanatory reason:** Alignment between criterion write requirements and lint rules is pending human policy decision in [6gkkzvr92waq](../tasks/6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md).

---

### 2. Findings

No open findings. All examined mutation paths, retry loops, authorization gates, and failure modes fail closed as intended.

---

### 3. Verified Consumer Inventory

Every caller, implementation, entry point, retry path, and test fake associated with the strict task target contract was verified in the sandbox:

| Category | Component / Symbol | Source Location | Verified Behavior / Role |
| :--- | :--- | :--- | :--- |
| **Port Definition** | [`TaskMutationTarget`](../../internal/core/store.go) | [`internal/core/store.go`](../../internal/core/store.go) | Minimal identity carrier containing only `ID string`; excludes declarations, locations, and slugs. |
| **Port Definition** | [`TaskStore.SetFields`](../../internal/core/store.go) | [`internal/core/store.go`](../../internal/core/store.go) | Signature takes `TaskMutationTarget`, updates map, and dryRun boolean. |
| **Port Definition** | [`TaskStore.EditBody`](../../internal/core/store.go) | [`internal/core/store.go`](../../internal/core/store.go) | Signature takes `TaskMutationTarget`, body text, appendMode, timestamp, and dryRun. |
| **Port Definition** | [`TaskStore.TransformTaskBody`](../../internal/core/store.go) | [`internal/core/store.go`](../../internal/core/store.go) | Signature takes `TaskMutationTarget`, timestamp, dryRun, and pure body transform callback. |
| **Adapter Implementation** | [`FS.SetFields`](../../internal/store/fsstore.go) | [`internal/store/fsstore.go`](../../internal/store/fsstore.go) | Enforces mutation authorization; resolves target via `resolveTaskMutationTarget`; CAS re-verifies via `resolvePath(target.ID)`. |
| **Adapter Implementation** | [`FS.EditBody`](../../internal/store/body.go) | [`internal/store/body.go`](../../internal/store/body.go) | Enforces mutation authorization; resolves target via `resolveTaskMutationTarget`; CAS re-verifies via `resolvePath(target.ID)`. |
| **Adapter Implementation** | [`FS.TransformTaskBody`](../../internal/store/body.go) | [`internal/store/body.go`](../../internal/store/body.go) | Enforces mutation authorization; resolves target before callback; handles no-op without write lock; CAS re-verifies via `resolvePath(target.ID)`. |
| **Adapter Lookup** | [`FS.resolveTaskMutationTarget`](../../internal/store/fsstore.go) | [`internal/store/fsstore.go`](../../internal/store/fsstore.go) | Checks `validEntityID(target.ID)`; delegates to `resolvePath(target.ID)`; returns `ErrNotFound` on missing, `ErrAmbiguous` on collision. |
| **Service Selection** | [`Service.selectTaskMutationTarget`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Reads task once using caller ref; validates `requireSourceID`; returns `TaskMutationTarget{ID: record.Source.ID}`. |
| **Service Retry** | [`retryTaskMutation`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Shares one bounded retry budget between initial selection and writes; pins `target` across conflict attempts. |
| **Service Entry** | [`Service.SetFields`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Validates fields in lexical order; invokes `retryTaskMutation` calling `s.store.SetFields`. |
| **Service Entry** | [`Service.ReplaceBody`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Selects target once; calls `s.store.EditBody` without automated retry loop. |
| **Service Entry** | [`Service.AppendBody`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Invokes `retryTaskMutation` calling `s.store.EditBody(target, ..., appendMode=true)`. |
| **Service Entry** | [`Service.transformTaskBody`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Primary path for AC mutations: calls `retryTaskMutation` $\to$ `s.store.TransformTaskBody`. |
| **Service Entry** | [`Service.SetCriteriaTracked`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Resolves source and destination IDs; validates self-tracking; calls `transformTaskBodyAt(target)`. |
| **Service Lifecycle** | [`Service.runTaskLifecycleMutation`](../../internal/core/service_task.go) | [`internal/core/service_task.go`](../../internal/core/service_task.go) | Pins selected `Plan.TaskID` into `retryPlanner` across conflict attempts; checks `!result.Committed`. |
| **Lifecycle Adapter** | [`FS.MutateTaskLifecycle`](../../internal/store/lifecyclemutation.go) | [`internal/store/lifecyclemutation.go`](../../internal/store/lifecyclemutation.go) | Executes under repo write lock; reloads graph; re-validates plan; re-derives blockers/impacts fresh on each attempt. |
| **CLI Producer** | `task set` / `task append` / `task ac` | [`internal/cli/task.go`](../../internal/cli/task.go) | CLI entry points parsing flags and invoking corresponding Service methods. |
| **CLI Producer** | `task move` / `task start` / `task defer` | [`internal/cli/task_move.go`](../../internal/cli/task.go) | CLI lifecycle entry points invoking `Service.Move` and `Service.DeferTask`. |
| **Test Double** | `selectedTaskStore` | [`internal/core/task_target_test.go#L13-L62`](../../internal/core/task_target_test.go) | Spies on reads, targets, clocks, and updates; verifies single selection and clock constancy. |
| **Test Double** | `conflictStore` | [`internal/core/occ_retry_test.go#L26-L61`](../../internal/core/occ_retry_test.go) | Exercises OCC conflict handling and retries with `TaskMutationTarget`. |
| **Test Double** | `criterionIdentityStore` | [`internal/core/task_correctness_test.go#L23-L41`](../../internal/core/task_correctness_test.go) | Verifies criterion tracking targets canonical source identity. |
| **Test Double** | `driftAfterTaskSelection` | [`internal/store/task_target_test.go#L272-L286`](../../internal/store/task_target_test.go) | Injects rename, deletion, or duplication between `ReadTask` and the first mutation attempt. |

---

### 4. Checks Executed vs. Inherited

#### Executed Independently by Antigravity in Sandbox:
1. **Full Race Test Suites (x2):**
   `GOCACHE=/tmp/taskflow-review-gocache go test -race ./internal/core ./internal/store ./internal/cli -count=1`
   - Run 1: Exit 0. Passed all tests across `./internal/core` (432 tests, 1.473s), `./internal/store` (397 tests, 10.294s), and `./internal/cli` (560 tests, 10.042s).
   - Run 2: Exit 0. Passed all tests across `./internal/core` (1.470s), `./internal/store` (10.441s), and `./internal/cli` (10.223s).
2. **Fresh Binary Producer Workflows:**
   Built executable to `/tmp/taskflow-review-bin/tskflwctl` outside the tracked tree. In an independent, unregistered space (`/tmp/taskflow-workflow-space`):
   - Initialized space with `--no-register` (repo ID `6gkytmh02s8w`).
   - Created epic `01-demo-epic` and two tasks: `first-alpha` (ID `6gkytstk6ah8`) and `second-beta` (ID `6gkytt5ce33n`).
   - Valid field update by fuzzy ref `alpha`: exit 0; SHA-256 updated to `8eeb6e13...`.
   - Valid body append by exact ID `6gkytstk6ah8`: exit 0; SHA-256 updated to `3b90cf89...`.
   - Valid criterion add by fuzzy ref `alpha`: exit 0; SHA-256 updated to `9f86bd8c...`.
   - Valid criterion check by fuzzy ref `alpha`: exit 0; SHA-256 updated to `c094a27c...`.
   - Valid tracked destination `task ac alpha --tracked 2 --to second-beta --reason "handed to beta"`: exit 0; wrote stable destination ID (`**tracked:** by 6gkytt5ce33n — handed to beta`); SHA-256 updated to `762f2314...`.
   - Rename then ID write: renamed task to `Renamed Omega` (file moved to `tasks/6gkytstk6ah8-renamed-omega.md`, exit 0); updated tags using original stable ID `6gkytstk6ah8`: exit 0; SHA-256 updated to `18d5fd19...`.
   - Lifecycle moves: `task start 6gkytt5ce33n`: exit 0, moved to `in-progress` (SHA-256 `436f48aa...`); `task defer 6gkytt5ce33n --until 2026-11-01`: exit 0, moved to `deferred` with `revisit_at: 2026-11-01` (SHA-256 `e2855e6c...`).
   - Refused self-tracking `task ac 6gkytstk6ah8 --tracked 1 --to 6gkytstk6ah8 --reason "self"`: exit 11 (`validation failed: a criterion cannot be tracked to its own task`); SHA-256 before and after identical (`18d5fd19...`).
   - Collision refusal: duplicated `6gkytstk6ah8` into `6gkytstk6ah8-duplicate.md`; attempted `task set 6gkytstk6ah8 --priority low`: exit 13 (`"6Gkytstk6ah8" matches 2 tasks by id: ambiguous match`); SHA-256 before and after identical (`18d5fd19...`).
   - No-write previews: `task set 6gkytstk6ah8 --priority low --dry-run` (exit 0) and `task ac 6gkytstk6ah8 --uncheck 1 --dry-run` (exit 0): SHA-256 before and after verified identical (`18d5fd19...`).
3. **Required Falsification Mutation:**
   Replaced exact action lookup `s.resolvePath(target.ID)` with fuzzy `s.resolve(target.ID)` in [`internal/store/fsstore.go:349`](../../internal/store/fsstore.go). Verified compilation. Executed tests:
   - `TestSelectedTaskPortsRefuseMissingDuplicateAndInvalidTargetsBeforeCallbacks`: failed in all 6 missing-ID subcases (`bad selected target reached callbacks/effects: err=<nil> want=not found`).
   - `TestTaskMutationRetryNeverFallsBackAfterSelectedIDDisappears`: failed in all 16 subcases across all operations (`missing selected identity was not refused: err=<nil> injected=true retries=1`).
   - Mutation was KILLED. Restored file and re-ran tests green.
4. **Independent Real-Store Disappeared ID & Impostor Probe:**
   Authored standalone test creating original task, sibling, and an impostor task whose slug was the uppercase string of the target ID (`strings.ToUpper(origID)`). Removed target file after selection. Exercised 7 distinct operations: `SetFields` write/preview, `EditBody` write/preview, `TransformTaskBody` write/no-op/preview. Verified that all 7 operations returned `domain.ErrNotFound`, and that neither the sibling nor the impostor was touched.
5. **Independent Denied Mutation & Lifecycle Blocker Probe:**
   - Denied direct port mutation: configured `FS` with `core.ReadOnlyMutations()`; called `TransformTaskBody` with populated record and callback counter. Verified `core.ErrReadOnlyPersistence` returned immediately, 0 callbacks executed, and file content was unchanged.
   - Lifecycle retry with fresh blocker: injected dependency `beta` into `alpha` before lifecycle verify. Verified that unforced transition failed with `*core.TaskEligibilityError` naming `alpha`, `GateBlocked`, and blocker details (`beta`, reason `BlockerNotStarted`). Verified that forced transition (`TaskLifecycleOverrideDependencyGate`) succeeded with `receipt.Forced=true`, `receipt.OverrideApplied=true`, `receipt.Before.Gate=GateBlocked`, non-empty `OutstandingBlockers`, and correct remedy string.
6. **Lint and Code Quality:**
   - `golangci-lint run ./...`: 0 issues.
   - `tskflwctl -C . lint --links`: passed cleanly (exit 0).

#### Inherited from Owner (Labelled, not credited as reviewer execution):
- Full repository-wide test suite (`go test -race ./...`).
- Git historical commits and ADR-0008 context.

---

### 5. Mutation and Probe Ledger

| Mutation / Probe | Target File / Seam | Intended Invariant / Test | Classification | Observed Failure / Evidence |
| :--- | :--- | :--- | :--- | :--- |
| Fuzzy lookup in `resolveTaskMutationTarget` (`s.resolve(target.ID)`) | [`internal/store/fsstore.go:349`](../../internal/store/fsstore.go) | `resolvePath` exact lookup refuses missing IDs without slug fallback | **KILLED** | Pre-compile check passed. `TestSelectedTaskPorts...` failed (missing target returned nil error, callbacks executed). `TestTaskMutationRetryNeverFallsBack...` failed (retries mutated impostor with matching slug). Restored cleanly. |
| Standalone real-store disappeared ID + uppercase impostor probe | Independent probe in `internal/store` | Disappeared ID returns `ErrNotFound` across write, preview, and no-op without mutating uppercase impostor | **CONFIRMED** | All 7 write/preview/no-op operations returned `domain.ErrNotFound`. Sibling and impostor byte hashes remained identical before and after. |
| Denied direct port mutation with observable callback | Independent probe in `internal/store` | `authorizeMutation` fails closed before target resolution or callbacks | **CONFIRMED** | `TransformTaskBody` returned `ErrReadOnlyPersistence`. Callback invocation count was 0. Target file bytes remained untouched. |
| Lifecycle retry with fresh prerequisite & non-default blocker receipt | Independent probe in `internal/store` | Lifecycle retry uses fresh graph to derive non-default blockers and forced receipt | **CONFIRMED** | Unforced retry returned `*TaskEligibilityError` with `GateBlocked` and prerequisite blocker. Forced retry returned `Forced=true`, populated `OutstandingBlockers`, and explanatory remedy. |
| Drift before first mutation attempt | [`internal/store/task_target_test.go:287`](../../internal/store/task_target_test.go) | Target identity survives rename/delete/duplicate between `ReadTask` and first mutation | **CONFIRMED** | Tested rename+reuse, delete+impostor, and duplicate across 9 operations in write and preview. Succeeded on renamed path; refused delete and duplicate without collateral damage. |
| CAS refuses collision appearing after exact lookup | [`internal/store/task_target_test.go:350`](../../internal/store/task_target_test.go) | `verifyUnchanged` re-resolves via `resolvePath` inside write lock | **CONFIRMED** | Tested across all 8 operations. Late collision created before write lock returned `domain.ErrAmbiguous` without retrying or altering files. |
| Fresh CLI binary end-to-end workflow | `/tmp/taskflow-review-bin/tskflwctl` in `/tmp/taskflow-workflow-space` | Valid workflows, rename+ID write, start/defer, self/collision refusal, dry-run previews | **CONFIRMED** | Fuzzy & ID refs succeeded; rename preserved ID targeting; self-tracking exited 11; collision exited 13; previews exited 0 with byte-identical hashes. |

---

### 6. Bounded Hypotheses Ledger

| Hypothesis | Challenged Invariant / Angle | Evidence & Method | Status |
| :--- | :--- | :--- | :--- |
| **H1: Initial user references retain fuzzy tiers while mutations bind strictly to Source.ID** | User may type an ID-shaped slug; selection must resolve it, but retries must not re-run the alias. | Tested via [`TestTaskInitialReferenceStillResolvesAnIDShapedHumanSlug`](../../internal/store/task_target_test.go) and live CLI workflow. `selectTaskMutationTarget` extracts `RecordSource.ID`. Target carries no declarations or locations. | **CONFIRMED** |
| **H2: FS resolves target exactly before callbacks, no-ops, previews, and CAS** | A vanished ID could fall back to an impostor slug or prefix; no-op/preview could bypass target validation. | Tested via required falsification mutation (killed), independent probe with uppercase impostor, and [`TestSelectedTaskPortsRefuseMissingDuplicateAndInvalidTargetsBeforeCallbacks`](../../internal/store/task_target_test.go). All returned `ErrNotFound` or `ErrAmbiguous` before callbacks or effects. | **CONFIRMED** |
| **H3: Selection and writes share one bounded retry budget; replacement & previews never retry** | Conflict during selection could allow separate retry loop; replacement could overwrite concurrent changes. | Tested via [`TestTaskMutationsSelectSourceOnceAndRetainTargetAndClock`](../../internal/core/task_target_test.go) and [`TestTaskMutationSelectionAndWritesShareOneRetryBudget`](../../internal/core/task_target_test.go). Budget and clock are shared; `ReplaceBody` and dry runs make exactly 1 attempt and fail on conflict. | **CONFIRMED** |
| **H4: Lifecycle first selection binds inside graph; retries retain selected Plan across contention** | Contention between attempts or empty intervening result could wipe selected plan or re-resolve raw ref. | Tested via [`TestTaskLifecycleRetryRetainsBindingAcrossUnplannedContention`](../../internal/core/task_target_test.go) and independent blocker probe. Intervening empty result does not erase binding; retry planner preserves plan identity while re-validating fresh graph state. | **CONFIRMED** |
| **H5: Persistence authorization, sentinel routing, and YAML/body fidelity remain preserved** | Denied mutation could leak disk reads or callbacks; sentinels could be collapsed into generic conflicts. | Tested via independent probe with `ReadOnlyMutations()`, live CLI workflows, and [`mutation_authorization_test.go`](../../internal/store/mutation_authorization_test.go). Authorization fails closed before lookup; `ErrAmbiguous` and `ErrNotFound` propagate without being masked as retryable conflicts. | **CONFIRMED** |

---

### 7. Systemic Second Pass Assessment

1. **Exact Lookup vs. Candidate Scanners:**
   The filesystem adapter distinguishes human-reference resolution from action target resolution. [`FS.resolve`](../../internal/store/fsstore.go) runs tiered matching (`resolveID`: exact ID, exact slug, prefix, substring). In contrast, [`FS.resolvePath`](../../internal/store/fsstore.go) and [`FS.resolveTaskMutationTarget`](../../internal/store/fsstore.go) invoke [`resolveExactID`](../../internal/store/resolve.go), which inspects only `cand.id == id`. Because candidate IDs are established from the filename's leading Crockford base32 segment and validated against canonical lowercase Crockford syntax, an impostor file containing an ID as its slug or with case-folded letters cannot match.
2. **Callback and No-Op Sequencing:**
   In [`TransformTaskBody`](../../internal/store/body.go), authorization and exact target lookup occur strictly before reading the file and executing the `transform` callback. When `transform` returns an unchanged body (no-op), the store returns without acquiring the write lock or calling `writeBody`. However, because target validation occurred up-front, an invalid, missing, or colliding target ID is refused even if the requested change would be a no-op.
3. **Dry-Run / Preview Isolation:**
   In all three task write methods, `dryRun=true` runs every validation check, parses the would-be updated frontmatter or body, and returns before acquiring `s.writeLock()` or calling `writeFileAtomic`. Target lookup runs before the dry-run return, ensuring that previews report identical identity errors (`ErrNotFound`, `ErrAmbiguous`, `ErrValidation`) without performing write I/O.
4. **Lifecycle Intervening Contention Invariant:**
   In [`runTaskLifecycleMutation`](../../internal/core/service_task.go), the retry loop initializes `retryPlanner := planner`. When attempt 1 produces a valid `result.Plan.TaskID`, `retryPlanner` is bound to a closure returning that plan. If attempt 2 suffers contention prior to planner invocation (returning an empty `result.Plan` with `domain.ErrConflict`), `retryPlanner` is not overwritten because the condition `if result.Plan.TaskID != ""` evaluates to false. Thus, attempt 3 continues to use the bound closure from attempt 1, preventing accidental re-resolution of the raw caller reference.
5. **Committed Receipt Guard:**
   Both single-task mutations and lifecycle mutations enforce that a mutation which successfully committed its write is never retried, even if an error is encountered during guard cleanup (e.g. `!result.Committed` in `runTaskLifecycleMutation`).

---

### 8. Isolation and Verification Attestation

```
sandbox_path=/private/tmp/isolated-review.0jXPFN
git_dir=/private/tmp/isolated-review.0jXPFN/.git
baseline_commit=67fadfb30df8a1c70e1c0684723a1768e4533cc9
source_blob=30ec53586632dc925935055e0ba9f97a0a4842f8
source_fingerprint=f445e2160591b27cf5563518c9f5f0d2d0396488
deliverable=planning/audits/6gktgjd9ebmn-2026-10-14-strict-task-mutation-targets-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner triage (2026-10-15)

Confirmed the source helper's delivery receipt before triage: `transfer=succeeded`,
`delivery=confirmed`, transferred blob `54660ae165707393974fcfcbf64708afa66bd143`.
No new findings require code changes. The exact-to-fuzzy mutation kill and independent
missing-ID, denied-callback and fresh-blocker probes support the scoped target contract.

This review captured the original dirty snapshot, before Claude M1's parse-free
selection correction. Its clean verdict missed that CLI repair regression and does
not independently approve the later correction. See the
[Claude disposition](6gktgjd3fm74-2026-10-14-strict-task-mutation-targets-implementation-claude.md#owner-triage-2026-10-15)
and [task evidence](../tasks/6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md).

Two report claims need qualification: the explicit committed-result retry guard is
lifecycle-specific; ordinary single-file retries rely on conflicts preceding durable
writes. Authorization-before-lookup describes the direct mutation ports, not core's
preceding read-only selection. Neither implies a universal receipt or authorization
framework. Owner normalization replaced workstation-only links with repository-relative
file links, corrected the CLI/resolveExactID destinations, and removed stale line
anchors; quoted line numbers otherwise describe the captured snapshot.
