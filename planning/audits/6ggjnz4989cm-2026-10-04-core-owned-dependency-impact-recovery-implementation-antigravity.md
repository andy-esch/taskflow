---
schema: 1
id: 6ggjnz4989cm
bucket: closed
area: core-owned-dependency-impact-recovery-implementation-antigravity
date: "2026-10-04"
updated_at: "2026-10-05"
---
# Audit: Core-owned dependency impact and recovery implementation — antigravity — 2026-10-04

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

Challenge the core-owned dependency impact/recovery implementation, not the entire adapter-neutral
refactor. Find demonstrated data-safety, semantic ownership, compatibility, or recovery regressions.
Do not award readiness based on a green suite alone. Do not implement fixes or settle findings.

## Review target

- Task: `planning/tasks/6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md`.
- Thread: `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`.
- Source audit: `planning/audits/6ga93gvffkpc-2026-09-15-adapter-hygiene.md`, M1/L1.
- Baseline HEAD is merged PR #280 (`289b0ba`); the implementation is the captured uncommitted
  handoff overlay, including newly added tests. Do not inspect only the source branch's commits.
- Production: `internal/core/{dependency_operations,service_task,task_lifecycle,thread_mutation}.go`,
  `internal/wire/{dependency,wire}.go`, `internal/cli/render/{dependency,render}.go`, and
  `internal/cli/task_dependency.go`. Review README, architecture, generated CLI reference, and
  machine goldens alongside code. Verify every name here exists before using it as evidence.
- Build a repository-wide consumer inventory: dependency and lifecycle receipts/failures, wire
  conversions, CLI success/error routes, lifecycle moves renderer, TUI receipt/remedy consumption,
  and repair envelopes reusing these DTOs. Identify each warning/guidance decision's owner.

## Intended contract to challenge

1. Core owns warning decisions. `NewlyUnsafe()` preserves the old predicate: a different non-clear
   gate OR newly inconsistent state. Broken → blocked still warns; this is not a severity scale.
   `NewlyInconsistent()` warns only on false → true. Safe, unchanged unsafe, cleared, and role-only
   changes do not manufacture new warnings.
2. Shared core task recovery applies to dependency and lifecycle receipts. Dependency human/JSON
   consumers receive equivalent guidance, including runnable stable-ID blocker commands. Wire
   copies core decisions; renderers do not rederive policy from string gates or boolean pairs.
3. Dependency impacts represent the full proposed plan, even on refusal or partial failure.
   `dry_run`, `applied_task_ids`, `remaining_task_ids`, and lifecycle `committed` are separate
   durability evidence. Failed writes must not claim planned state landed. Partial/all-applied
   conflict failures stay visible and never silently auto-retry after the first durable write.
4. Pre-write CLI failures deliberately omit `error.dependency_mutation`; error text remains
   classified/actionable. Core callers retain the typed receipt. This exclusion is not a missing
   payload bug. Committed-prefix errors carry the structured receipt and the same core advice.
5. Schema 1.81 adds warning flags to shared impact DTOs and optional dependency `remedy`, including
   nested recovery envelopes. No existing field meanings, persisted Markdown schema, eligibility,
   authorization, override, repair transaction, or legacy migration semantics change.

Non-goals: a generic recovery framework, severity ordering, a second backend, TUI unreadable-marker
redesign, constructor mutation-policy design, or fixing the separately tracked init selector issue.
Evidence-backed discoveries outside this task should be proposed as followups, not imposed as
unbounded refactor closure blockers.

## Mandatory evidence floor

- Run baseline focused tests, full tests, race tests, lint, build, and generated-output comparisons
  in your sandbox. Show exact commands and results; report unavailable tooling as unverified.
- Independently reproduce preview, applied warning, no-op, clear, refusal, partial prefix, and
  all-applied failure. For prefix failures use the real filesystem migration tests as well as
  portable service tests; inspect resulting files and retry counts, not just error prose.
- Mutate at least three distinct owner/caller boundaries with compiler-valid probes: core
  predicate, mapper/renderer, and durable-prefix retry. Each named test must fail for the semantic
  reason claimed. Include one coordinated mapper+renderer mutation where their agreement could
  mask policy leakage. Restore and rerun before proceeding. A compile error is not a killed probe.
- Populate schema branches with `newly_unsafe=true`, `newly_inconsistent=true`, and non-empty
  dependency `remedy`, including `error.dependency_mutation`. Validate the emitted envelope against
  its own definition. Test false flags and omitted empty remedy too; empty arrays are not evidence.
- Run the blocker commands actually recommended by a receipt. Confirm the referenced stable IDs
  resolve; on preview distinguish the current graph's blockers from the proposed graph's blockers.
- Produce a claim/evidence matrix and an explicit verdict (ready, ready with tracked followups,
  or blocked). A no-findings verdict must show hostile tests, not paraphrase owner evidence.

## Required hostile angles

### Semantic ownership and presentation

Trace the full core → wire → CLI/TUI path. Can a non-CLI consumer obtain the warning and guidance
without parsing prose or reconstructing graph policy? Does advice survive when flags are false or
when a renderer receives a copied flag that disagrees with raw state? Is there any remaining adapter
policy duplication hiding behind a helper? Separate genuine recovery policy from output formatting.

### Durability and retry

Challenge failures before planning, after planning but before a write, after each atomic write,
after the last write, and during guard cleanup. Change the durable graph between a partial failure
and resuming intent. Do receipts and advice still distinguish current evidence from the abandoned
full plan? Can a committed conflict be swallowed into a successful no-op? Do error wrapping,
`errors.Is`/`errors.As`, exit classification, and structured recovery agree?

### Compatibility and all consumers

Review schema descriptions and the 1.81 additive claim. Shared DTOs appear in repair/lifecycle as
well as ordinary dependency output. Verify all populated shapes, required boolean properties,
optional remedy omission, and older field meanings. Look for legacy manually constructed receipts
or failures whose default values could create inconsistent guidance; identify actual reachable
callers rather than treating arbitrary fabricated invalid objects as production bugs.

### Systemic second pass, particularly for Antigravity

After the first pass, pick two apparently trustworthy abstractions and try to disprove them with
concrete counterexamples. Good candidates: tests compare a mapper to the same mutated predicate;
a fake port reports stronger durability than the real store; a generated schema branch is never
populated; error formatting invents guidance that the structured receipt lacks. Report for each
hypothesis the exact hostile input, path, observed output, and why it survived or was falsified.
Do not merely label patterns as anti-patterns. One demonstrated systemic issue is more useful than
several speculative redesigns. Also explain which owner-authored tests are strongest and weakest.

## Validation and restoration

Follow the mandatory independent-clone protocol injected above. Never execute mutation probes or
generators in the shared source. The sandbox baseline commit is the only permitted review commit;
do not stage, commit again, switch branches, push, or create a PR. Restore probes only in your sandbox.

Suggested commands (verify test names first):

```sh
go test ./internal/core -run 'TestTaskGraphStateImpactNewlyUnsafe|TestThreadProjectionImpactNewlyInconsistent|TestServiceDependency|TestDependencyPreviewRefusal|TestLifecycleThreadRecovery' -count=1
go test ./internal/wire -run 'TestImpactWireCopiesCoreDecisionsAndDependencyRemedy|TestJSONSchema_ValidatesRealOutput' -count=1
go test ./internal/cli/render -run 'TestDependencyHumanUsesCoreGuidance|TestMovesHumanUsesProjectedImpactFlags' -count=1
go test ./internal/cli -run 'TestDependencyMutationPublishesCoreRecoveryIntent|TestWriteErrorCarriesStructuredDependencyMutationRecovery|TestTaskDependAddRejectsCycle' -count=1
go test -race ./internal/store -run 'TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes' -count=20
go test ./...
go test -race ./...
golangci-lint run ./...
just build
./bin/tskflwctl -C . lint
```

Generate CLI docs and schema comments into a private temporary directory and compare them with
`docs/cli` and `internal/wire/schema_comments.json`; do not use `-update` to conceal drift.
For throwaway dogfood use `init --path "$SCRATCH" --taskflow-root planning --no-register`.
Global `-C` does not select init's bootstrap destination today; the separate followup owns that.
Do not mutate real planning states, source audit dispositions, task ACs, or machine fixtures.

## Deliverable

Update only your assigned audit, retaining this brief. Record severity-coded open findings with
reproduction, exact implementation locations, user/machine impact, a bounded recommended fix, and
the regression that would pin it. Keep falsified hypotheses and protocol/test limitations separate
from findings. Include commands/results, mutation outcomes and restoration, consumer inventory,
claim/evidence matrix, isolation attestation, and verdict. Do not duplicate the same root cause
across several bands or invent findings to meet a quota.

## Reviewer report

### Isolation attestation

The reviewer executed the mandatory independent sandbox protocol via `scripts/isolated-review-workspace.sh`:

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Nhk7ui
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Nhk7ui/.git
baseline_commit=8bfc1cfa2def30456e0474338fa066b077e59333
source_blob=887b8e81c3102f17b7c201946f9dd7665cd927ac
source_fingerprint=f50395ebd55515caef67344317659cfb055e85f8
deliverable=planning/audits/6ggjnz4989cm-2026-10-04-core-owned-dependency-impact-recovery-implementation-antigravity.md
deliverable_changed=true
transfer=succeeded
```

All inspection, builds, unit and race tests, dogfood experiments, schema validation, and mutation probes were executed solely inside `$SANDBOX`. The shared source checkout remained strictly read-only.

---

### Executive summary & verdict

**Verdict:** `ready with tracked followups`

The implementation successfully transfers semantic ownership of dependency and lifecycle impact warning decisions (`NewlyUnsafe()`, `NewlyInconsistent()`) and actionable recovery guidance (`remedy`) into `internal/core`. Primary adapters (`internal/cli/render/dependency.go`, `internal/cli/render/render.go`) no longer re-derive graph policies or evaluate string gates. Durability boundaries fail closed: pre-write validation errors omit mutation recovery payloads to avoid misrepresenting planned edges as landed, while committed-prefix failures carry structured recovery receipts and halt auto-retries immediately. Schema 1.81 is strictly additive per ADR-0008.

One medium finding (`M1`) is identified: `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` in `internal/wire/dependency_recovery_test.go` has a coverage blind spot that only evaluates transitions from `GateClear`, allowing a wire mapper that re-derives a naive `After.Gate != GateClear` predicate to pass undetected on unchanged-blocked and inconsistent states.

---

### Findings

#### M1. Wire impact decision regression test masks shallow gate re-derivation on unchanged-unsafe and inconsistent states · **Status:** fixed 2026-10-05 (PR #282)

- **Location:** `internal/wire/dependency_recovery_test.go:11-45`
- **Reproduction / Hostile Evidence:**
  Mutate `toTaskGraphStateImpactsJSON` in `internal/wire/dependency.go:214` to re-derive the naive gate check instead of copying core's decision:
  ```go
  // internal/wire/dependency.go
  NewlyUnsafe: impact.After.Gate != core.GateClear,
  ```
  Run the test suite:
  ```sh
  go test ./internal/wire -run 'TestImpactWireCopiesCoreDecisionsAndDependencyRemedy' -count=1
  # ok  github.com/andy-esch/taskflow/internal/wire  0.204s (EXIT 0)
  go test ./internal/cli -count=1
  # ok  github.com/andy-esch/taskflow/internal/cli   3.482s (EXIT 0)
  ```
  Every test in `internal/wire` and `internal/cli` passes without error.
- **Root Cause & User/Machine Impact:**
  `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` iterates only over `unsafe := []bool{false, true}` where `Before.Gate` is always `GateClear` and `After.Gate` is either `GateClear` (unsafe=false) or `GateBlocked` (unsafe=true). In both cases, `impact.NewlyUnsafe()` happens to match `impact.After.Gate != core.GateClear`.
  The test never exercises:
  1. An already-blocked task (`Before.Gate = GateBlocked, After.Gate = GateBlocked`), where `NewlyUnsafe()` is `false` (no new warning), but `After.Gate != GateClear` evaluates to `true`.
  2. A task with consistency restored or newly inconsistent with clear gate (`After.Gate = GateClear, After.Inconsistent = true`), where `NewlyUnsafe()` is `true`, but `After.Gate != GateClear` evaluates to `false`.
  If an adapter or wire mapper regresses to a shallow `After.Gate != GateClear` check, machine JSON consumers would receive spurious `newly_unsafe: true` flags on already-blocked tasks, triggering false warning alerts and unnecessary inspection workflows.
- **Bounded Recommended Fix:**
  Expand `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` in `internal/wire/dependency_recovery_test.go` to include test table entries covering unchanged-blocked (`GateBlocked -> GateBlocked`), newly inconsistent (`Inconsistent: false -> true` with `GateClear`), and Thread view impacts with `Before.Inconsistent = true, After.Inconsistent = true`.
- **Regression Pin:**
  The expanded test asserts `wire.Impacts[0].NewlyUnsafe == impact.NewlyUnsafe()` across all gate/inconsistency combinations, causing any mapper that re-derives `After.Gate != GateClear` to fail immediately.

---

**Resolution:** Independently reproduced the exact shallow task mapper passing
the original focused and full wire/CLI suites. Expanded the regression to 36
task gate/inconsistency transition pairs and all four Thread consistency
transitions; isolated shallow-gate, changed-gate-only, and
After.Inconsistent-only mutants now fail and restored tests pass. The fix merged
in PR #282; production policy was unchanged. Audit remains closed, and release
inclusion is separate.

### Systemic second pass

#### Hypothesis A: Wire mapper test helper coverage masking (Falsified Test Trustworthiness)
- **Hypothesis:** `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` proves that wire conversion strictly copies core decisions without re-deriving policy.
- **Hostile Probe:** Replaced `NewlyUnsafe: impact.NewlyUnsafe()` with `NewlyUnsafe: impact.After.Gate != core.GateClear` in `internal/wire/dependency.go:214`.
- **Observed Result:** Test passed with exit code 0.
- **Why it was falsified:** The test helper used symmetric default zero-values for `Before`, rendering `NewlyUnsafe()` indistinguishable from `After.Gate != GateClear`. Recorded as finding `M1`.

#### Hypothesis B: Durability boundary and auto-retry masking (Verified Production Safety)
- **Hypothesis:** Durable-prefix recovery might accidentally retry or swallow conflict errors when a partial write lands before failure.
- **Hostile Probe:** Removed `len(result.AppliedTaskIDs) > 0` from the retry guard in `internal/core/dependency_operations.go:161`.
- **Observed Result:**
  - `TestServiceDependencyRefusalAndCommittedFailureRecovery` failed with:
    `typed failure/receipt lost: {Operation:add Changed:false ...}, <nil>`
  - `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries` failed with:
    `calls=2, <nil>`
- **Why it survived:** The check in `internal/core/dependency_operations.go:161` (`len(result.AppliedTaskIDs) > 0`) is essential to prevent a committed prefix from being silently replayed into a no-op that swallows disk/guard failures.

#### Evaluation of Owner-Authored Tests
- **Strongest Tests:**
  - `TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes` (`internal/store/dependency_operations_test.go`): Uses the real filesystem storage adapter (`NewFS`), tests all prefix boundaries (`failAfter = 1, 2, 3`), asserts exact disk changes, verifies error wrapping and remedy distinction, and proves convergence on rerun under `-race -count=20`.
  - `TestMovesHumanUsesProjectedImpactFlagsNotWireStatePolicy` (`internal/cli/render/impact_recovery_test.go`): Adversarially forces `impact.NewlyUnsafe` to disagree with `impact.After.Gate` to guarantee the human renderer respects copied wire booleans rather than parsing wire strings.
  - `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries` (`internal/core/dependency_recovery_test.go`): Verifies `store.calls == 1` even when configured with `WithRetry(4, ...)` to ensure durable prefixes halt retries.
- **Weakest Tests:**
  - `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` (`internal/wire/dependency_recovery_test.go`): Narrow coverage (only tests transitions from `GateClear`) that fails to kill naive gate re-derivation (finding `M1`).

---

### Mandatory evidence floor

#### 1. Baseline command execution
All commands executed in `$SANDBOX`:

| Suite / Command | Exact Command | Result |
| --- | --- | --- |
| Core focused | `go test ./internal/core -run 'TestTaskGraphStateImpactNewlyUnsafe\|TestThreadProjectionImpactNewlyInconsistent\|TestServiceDependency\|TestDependencyPreviewRefusal\|TestLifecycleThreadRecovery' -count=1` | PASS (0.321s) |
| Wire focused | `go test ./internal/wire -run 'TestImpactWireCopiesCoreDecisionsAndDependencyRemedy\|TestJSONSchema_ValidatesRealOutput' -count=1` | PASS (0.205s) |
| Render focused | `go test ./internal/cli/render -run 'TestDependencyHumanUsesCoreGuidance\|TestMovesHumanUsesProjectedImpactFlags' -count=1` | PASS (0.358s) |
| CLI focused | `go test ./internal/cli -run 'TestDependencyMutationPublishesCoreRecoveryIntent\|TestWriteErrorCarriesStructuredDependencyMutationRecovery\|TestTaskDependAddRejectsCycle' -count=1` | PASS (0.405s) |
| Store race (x20) | `go test -race ./internal/store -run 'TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes' -count=20` | PASS (3.701s) |
| Full test suite | `go test ./...` | PASS (34 packages) |
| Full race suite | `go test -race ./...` | PASS (34 packages) |
| Linter | `golangci-lint run ./...` | PASS (0 issues) |
| Build & lint | `just build && ./bin/tskflwctl -C . lint` | PASS (all planning entities pass) |
| CLI docgen drift | `go run ./internal/tools/docgen -out "$TMP_DOCS" && diff -u docs/cli "$TMP_DOCS"` | PASS (0 diff) |
| Schema comment drift | `go run ./internal/tools/schemacomments -out "$TMP_COMMENTS" && diff -u internal/wire/schema_comments.json "$TMP_COMMENTS"` | PASS (0 diff, 267 comments) |

#### 2. Independent reproduction of mutation outcomes
Executed dogfood operations against an isolated throwaway repository initialized via `./bin/tskflwctl init --path "$SCRATCH" --taskflow-root planning --no-register`:

1. **Preview (`--dry-run`):**
   - Command: `tskflwctl task depend add 6gg000000002 --on 6gg000000001 --dry-run --json`
   - Output: `dry_run: true`, `applied_task_ids: []`, `remaining_task_ids: []`, `impacts[0].newly_unsafe: true`, `remedy: "preview only: proposed changes would introduce newly unsafe task state; inspect affected tasks with `tskflwctl task blockers 6gg000000002` and adjust the request or prerequisites before applying"`.
   - **Blocker Command Verification (Preview):** Executed `tskflwctl task blockers 6gg000000002` against current state. Output: `state: candidate/clear eligible=true view: frontier ✔ no blockers`. Accurately distinguishes current graph state from prospective proposed state.
2. **Applied Warning:**
   - Command: `tskflwctl task depend add 6gg000000002 --on 6gg000000001 --json`
   - Output: `dry_run: false`, `applied_task_ids: ["6gg000000002"]`, `remaining_task_ids: []`, `impacts[0].newly_unsafe: true`, `remedy: "inspect each affected task with `tskflwctl task blockers 6gg000000002` and restore sound prerequisites or update its dependencies"`.
   - **Blocker Command Verification (Applied):** Executed `tskflwctl task blockers 6gg000000002` against current state. Output: `state: candidate/blocked eligible=false view: frontier • prereq not-started direct (6gg000000002 -> 6gg000000001)`. The referenced stable ID resolves and reports the actionable blocker.
3. **No-op:**
   - Command: `tskflwctl task depend add 6gg000000002 --on 6gg000000001 --json`
   - Output: `changed: false`, `edges[0].outcome: "skipped"`, `impacts: []`, `applied_task_ids: []`, `remedy` omitted.
4. **Clear:**
   - Command: `tskflwctl task depend remove 6gg000000002 --on 6gg000000001 --json`
   - Output: `changed: true`, `edges[0].outcome: "removed"`, `impacts[0].after.gate: "clear"`, `impacts[0].newly_unsafe: false`, `remedy` omitted.
5. **Refusal (Pre-write Validation):**
   - Command: `tskflwctl task depend add 6gg000000002 --on 6gg000000002 --json`
   - Output: Exit code 11 (`validation`), error payload: `{"schema_version":"1.81","error":{"code":"validation","message":"validation failed: task 6gg000000002 cannot depend on itself; no dependency task files were applied; inspect the failure and current graph before retrying"}}`.
   - `error.dependency_mutation` is strictly omitted (`nil`).
6. **Partial Prefix Failure:**
   - Verified via `TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes` and `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries`.
   - Result: File-backed store leaves exactly `failAfter` files modified on disk, `applied_task_ids` names them, `remaining_task_ids` names the rest, `remedy` requires inspection of durable progress before resuming, and auto-retry is halted (`calls = 1`).
7. **All-Applied Failure:**
   - Verified via `TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes` (`failAfter = 3`).
   - Result: All 3 files durably landed on disk, `remaining_task_ids` is empty, `remedy` explains writes already landed, and retry does not re-write files.

#### 3. Compiler-valid mutation probes & restoration

| Probe ID | Boundary Challenged | Exact Mutation | Target Test | Failure Result | Restoration |
| --- | --- | --- | --- | --- | --- |
| Probe 1 | Core warning decision | `internal/core/task_lifecycle.go:37`: `NewlyUnsafe()` returns `false` | `TestTaskGraphStateImpactNewlyUnsafe` | FAIL: `NewlyUnsafe() = false, want true` across newly_blocked, newly_broken, blocked_to_broken, broken_to_blocked, newly_inconsistent | Restored; test passed (0.201s) |
| Probe 2 | Coordinated Mapper + Renderer | `internal/wire/dependency.go:214`: `NewlyUnsafe: impact.After.Gate != core.GateClear` AND `internal/cli/render/render.go:327`: `if impact.After.Gate != string(core.GateClear)` | `TestMovesHumanUsesProjectedImpactFlagsNotWireStatePolicy` | FAIL: `renderer re-derived policy instead of copying flags=false` | Restored; test passed (0.253s) |
| Probe 3 | Durable-prefix retry halt | `internal/core/dependency_operations.go:161`: Removed `len(result.AppliedTaskIDs) > 0` | `TestServiceDependencyRefusalAndCommittedFailureRecovery` & `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries` | FAIL: Conflict failure swallowed into `err == nil` on retry; calls=2 instead of 1 | Restored; test passed (0.194s) |

#### 4. Schema branch validation
Validated using `github.com/santhosh-tekuri/jsonschema/v6` against `wire.JSONSchema()`:
- `DependencyMutationEnvelope`: Validated with `newly_unsafe = true` and populated `remedy`.
- `DependencyMutationEnvelope`: Validated with `newly_unsafe = false` and omitted empty `remedy` (`omitempty`).
- `MovesEnvelope`: Validated with `newly_unsafe = true` / `newly_inconsistent = true` and `newly_unsafe = false` / `newly_inconsistent = false`.
- `ErrorEnvelope`: Validated with `error.dependency_mutation = nil` (pre-write failure) and populated `error.dependency_mutation` (committed-prefix recovery).
- Verified schema property constraints: `newly_unsafe` is required in `TaskGraphStateImpactJSON`, `newly_inconsistent` is required in `ThreadProjectionImpactJSON`, and `remedy` is optional in `DependencyMutationJSON`.

---

### Repository-wide consumer inventory

| Component / Layer | Type / Path | Role / Consumption | Warning / Guidance Decision Owner |
| --- | --- | --- | --- |
| Core Model | `core.TaskGraphStateImpact` (`internal/core/task_lifecycle.go:26`) | Derived state change projection | Core (`TaskGraphStateImpact.NewlyUnsafe()`) |
| Core Model | `core.ThreadProjectionImpact` (`internal/core/thread_mutation.go:390`) | Thread projection change | Core (`ThreadProjectionImpact.NewlyInconsistent()`) |
| Core Service | `core.DependencyMutationReceipt` (`internal/core/dependency_operations.go:42`) | Mutation receipt for add/remove/migrate | Core (`dependencyMutationRemedy`) |
| Core Service | `core.DependencyMutationFailure` (`internal/core/dependency_operations.go:61`) | Typed error carrying durable prefix receipt | Core (`dependencyMutationRemedy` / `DependencyMutationFailure.Error()`) |
| Core Service | `core.TaskLifecycleReceipt` (`internal/core/task_lifecycle.go:81`) | Lifecycle transition receipt | Core (`taskLifecycleRemedy` / `taskImpactRemedy`) |
| Wire DTO | `wire.TaskGraphStateImpactJSON` (`internal/wire/dependency.go:199`) | Wire projection of task state impact | Copied from Core (`NewlyUnsafe`) |
| Wire DTO | `wire.ThreadProjectionImpactJSON` (`internal/wire/dependency.go:252`) | Wire projection of Thread impact | Copied from Core (`NewlyInconsistent`) |
| Wire DTO | `wire.DependencyMutationJSON` (`internal/wire/dependency.go:328`) | Wire projection of dependency receipt | Copied from Core (`Remedy`) |
| Wire DTO | `wire.TaskGraphRepairJSON` (`internal/wire/dependency_repair.go:78`) | Wire projection of repair receipt | Reuses `toTaskGraphStateImpactsJSON` (Core-owned) |
| CLI Render | `render.DependencyMutationHuman` (`internal/cli/render/dependency.go:207`) | Human output for `task depend add/remove/migrate` | Formats Core decisions (`impact.NewlyUnsafe()`, `receipt.Remedy`) |
| CLI Render | `render.MovesHuman` (`internal/cli/render/render.go:275`) | Human output for `task start/complete/move` | Formats Wire flags (`impact.NewlyUnsafe`, `impact.NewlyInconsistent`) |
| CLI Render | `render.TaskGraphRepairHuman` (`internal/cli/render/dependency.go:104`) | Human output for `task depend repair` | Formats repair receipt impacts; residual warnings owned by repair defect rules |
| CLI Error Route | `cli.dependencyFailure` (`internal/cli/task_dependency.go:27`) | Maps core failures to CLI error envelopes | Core owns prefix classification; CLI omits pre-write payload |
| CLI Error Route | `cli.WriteError` (`internal/cli/exit.go:136`) | Emits classified `ErrorEnvelope` | Domain exit code + Core remedy string in `error.message` |
| TUI Model | `tui.Model.Update` (`internal/tui/model.go:472`) | Status bar flash message on task move | Formats Core `msg.lifecycle.Remedy` directly |
| TUI Entity | `tui.moveTask` / `tui.deferTaskCmd` (`internal/tui/entity.go:363, 381`) | Invokes `core.Service.Move` / `DeferTask` | Handles `TaskLifecycleMutationFailure` committed receipts |

---

### Claim / evidence matrix

| Contract Claim | Design Expectation | Verified Evidence | Verdict |
| --- | --- | --- | --- |
| 1. Core warning ownership | `NewlyUnsafe()` warns on changed non-clear gate or newly inconsistent; safe, unchanged-unsafe, and cleared do not warn. | `TestTaskGraphStateImpactNewlyUnsafe` & `TestServiceDependencySafeAlreadyUnsafeAndClearedNeedNoNewRecovery` pass. Killed by Probe 1. | VERIFIED |
| 2. Shared recovery & equivalence | Dependency and lifecycle receipts share recovery intent; wire copies flags; renderers do not re-derive policy. | `TestMovesHumanUsesProjectedImpactFlagsNotWireStatePolicy` kills wire-string comparison. Dogfood confirms matching human and JSON remedy. | VERIFIED |
| 3. Durability & retry safety | Impacts describe full proposed plan; applied/remaining IDs track persistence; partial/all-applied failures never auto-retry. | `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries` & `TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes` pass under `-race -count=20`. Killed by Probe 3. | VERIFIED |
| 4. Error envelope consistency | Pre-write errors omit `dependency_mutation`; committed prefixes include structured receipt; core remedy preserved. | `TestWriteErrorCarriesStructuredDependencyMutationRecovery` & `TestTaskDependAddRejectsCycleWithValidationExit` pass. Pre-write and post-write JSON verified. | VERIFIED |
| 5. Schema 1.81 compatibility | Additive revision: boolean flags on shared impact DTOs, optional `remedy` on dependency mutation. | ADR-0008 verified; `schema_comments.json` has 267 entries; all golden files pass; schema validation passes for true/false and omitted remedy branches. | VERIFIED |
| 6. Regression test resilience | Regression tests kill shallow adapter re-derivation. | `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` passes when wire mapper re-derives `After.Gate != GateClear`. | TRACKED (Finding M1) |

## Owner triage (2026-10-05)

Accepted M1 as an in-scope test-coverage defect, not a production-policy regression.
The owner independently applied the exact shallow task mapper in a temporary
non-Git copy: the original focused test and full wire/CLI suites passed. The
production mapper already invokes core's decision and requires no semantic change.

Expanded the named wire test to 36 task gate/inconsistency transition pairs across
dependency and lifecycle converters and all four Thread consistency transitions.
It retains core-owned guidance even on false warnings. Core's literal predicate
tests remain separate; this test does not duplicate their policy formula.

The same temporary copy then compiled and rejected three mutants with the expanded
`go test ./internal/wire -run '^TestImpactWireCopiesCoreDecisionsAndDependencyRemedy$' -count=1`:

- `NewlyUnsafe: impact.After.Gate != core.GateClear` fails unchanged blocked/broken
  and newly inconsistent clear-state cases.
- A changed-non-clear-gate-only mapper fails newly inconsistent cases, including
  `clear-false_to_clear-true` and `blocked-false_to_blocked-true`.
- `NewlyInconsistent: impact.After.Inconsistent` fails `thread/true_to_true`.

Every failure is a semantic assertion, not a compile error. Restored focused tests
pass after each probe; the restored production mapper is byte-identical to source.
No probes ran in the shared checkout or the retained reviewer sandbox.

The retained independent sandbox's baseline matches the attestation and its only
worktree difference is this audit. This owner pass verifies M1 independently; the
reviewer's broad inventory is not an assertion that every TUI consumer surfaces
guidance. Codex already demonstrated the Thread-only TUI loss, tracked separately
by [6ggkdbg0816h](../tasks/6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md).

The report above retains its submission-time verdict and claim/evidence labels.
M1's managed status/resolution records the fix merged in PR #282; closure settles
this review, not the tracked TUI task or release inclusion.
