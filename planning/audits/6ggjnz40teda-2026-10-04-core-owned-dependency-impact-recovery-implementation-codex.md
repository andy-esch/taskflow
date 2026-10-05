---
schema: 1
id: 6ggjnz40teda
bucket: closed
area: core-owned-dependency-impact-recovery-implementation-codex
date: "2026-10-04"
updated_at: "2026-10-04"
---
# Audit: Core-owned dependency impact and recovery implementation — codex — 2026-10-04

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
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

**Verdict: ready with tracked followups.** Both review passes are complete against the captured implementation overlay. No new in-scope data-safety, warning-policy, retry, or machine-contract regression was demonstrated. M1 remains **open** as a bounded, pre-existing TUI consumer followup; it is not imposed as closure work for the dependency-impact task. No implementation fix or finding disposition was applied.

### Scope and consumer inventory

The comparison was `git diff 289b0ba` in the independent clone, including uncommitted/new tests captured by its baseline. The task remains `in-progress`; planning acceptance criteria and owner claims were treated as requirements, not executed evidence. No source task, Thread, sibling audit, or source-audit disposition was changed.

| Verified producer/consumer | Ownership and observed boundary |
| --- | --- |
| `internal/core/task_lifecycle.go:38`; `internal/core/thread_mutation.go:401` | `TaskGraphStateImpact.NewlyUnsafe()` owns changed-non-clear-gate OR false→true inconsistency; `ThreadProjectionImpact.NewlyInconsistent()` owns false→true only. Broken→blocked is still a warning. |
| `internal/core/dependency_operations.go:138`, `:171`, `:198` | `runDependencyMutation`, `dependencyReceipt`, and `dependencyMutationRemedy` retain proposed impacts, calculate planned/applied/remaining IDs, and select no-write/partial/all-applied advice. `:161` stops retry once any write landed. |
| `internal/core/service_task.go:548`, `:567`, `:579`, `:597` | Lifecycle retry is bounded by `Committed`; receipts share `taskImpactRemedy` with dependency mutations. Thread and typed-override advice remain separate core decisions. Preview advice explicitly describes proposed changes. |
| `internal/core/dependency_operations.go:66`, `:90`; `internal/core/task_lifecycle.go:185`, `:196` | Dependency failure text includes the receipt remedy and wraps the cause. Lifecycle cleanup text reports commitment and inspection; the richer task/Thread remedy stays in its receipt. `errors.Is`/`errors.As` retain sentinel and typed evidence. |
| `internal/store/graphmutation.go:26`, `:105`; `internal/store/lifecyclemutation.go:90` | The filesystem port owns authoritative reads, validation, CAS, per-file atomic replacement and durable result recording. Lifecycle copies analysis and Thread impacts before the write. Test hooks interrupt real production operations; they are not a substitute backend. |
| `internal/store/lock.go:114`; `internal/store/lock_unix.go:20` | Process guard plus advisory `flock` on the planning root directory, not a guessed lock-file path. Cleanup errors are joined after releasing both guards. |
| `internal/wire/dependency.go:208`, `:262`, `:275`, `:342` | Shared impact mappers call core methods and copy remedy. Dependency/lifecycle DTOs normalize required arrays; optional empty remedy is omitted. Warning booleans are always serialized and required by the schema. |
| `internal/cli/task_dependency.go:27`, `:90`; `internal/cli/exit.go:164` | Success selects human/JSON presentation. Pre-write failure intentionally stays the ordinary classified error; only an applied prefix receives `error.dependency_mutation`. `WriteError` converts the retained receipt without inventing policy. |
| `internal/cli/moves.go:46`; `internal/cli/task.go:169`; `internal/cli/exit.go:174`; `internal/wire/envelopes.go:1163` | Committed lifecycle failure survives batch/error wrapping and creation routes; nested recovery converts the same receipt with workspace context. |
| `internal/cli/render/dependency.go:229`, `:235`; `internal/cli/render/render.go:327`, `:338`, `:344` | Dependency human output calls the core warning method; Moves human output consumes copied wire flags. Both print nonempty owner advice independently of task warning classification. Committed Moves errors also print receipt advice at `:301`. |
| `internal/tui/entity.go:356`, `:378`; `internal/tui/model.go:472` | Move/defer commands preserve receipts and committed cleanup warnings. The Model's task-impact-count condition suppresses Thread-only recovery advice: demonstrated M1. |
| `internal/core/dependency_repair.go:647`; `internal/wire/dependency_repair.go:117`; `internal/cli/render/dependency.go:20`; `internal/cli/exit.go:169` | Repair success/error payloads reuse both shared impact mappers. Repair keeps its existing final/prefix graph semantics, distinct from ordinary dependency proposed-plan receipts. Human repair diagnosis formats source edits; it does not implement the newly-unsafe predicate. |

Repository-wide search was `rg -n 'TaskGraphStateImpact|ThreadProjectionImpact|\.Remedy|DependencyMutationFailure|TaskLifecycleMutationFailure' internal --glob '*.go' --glob '!**/*_test.go'`; the complete result is `consumers.txt` in the private evidence directory. Unrelated space-health, rename, and Thread-operation remedy types were excluded after inspecting their identity. The production dependency failure constructor is in `runDependencyMutation`, which always fills core guidance on failure. Its legacy `Error()` fallback can assist manually constructed empty-remedy failures, but no reachable production constructor losing that receipt field was found; arbitrary invalid external structs are not reported as a shipped bug.

### First pass: independent behavioral evidence

The private evidence directory is `$SANDBOX/.git/isolated-review-workspace/evidence`. Probe sources and logs are retained there, outside the transferable tree. Reviewer-added tests were temporary and are **not committed regression coverage**. To reproduce them, copy `probe-store.go`, `probe-wire.go`, `probe-tui.go`, or `probe-cli-render.go` to the matching package's `audit_probe_test.go` **in the retained sandbox only**, then run the named commands below. Remove them again before helper verification.

| Claim challenged | Independent hostile input and observed result |
| --- | --- |
| Preview is proposed, not durable | Real CLI scaffold and two tasks; dry JSON and human add both report newly unsafe `clear→blocked` and the same preview-qualified advice. All task bytes remain equal. Planned IDs contain the dependent; applied IDs are empty. |
| Advice commands resolve stable identity | Receipt command `tskflwctl task blockers 6ggjtrct1dxz` executes, and `task show` resolves that ID. In preview the command reports current `gate=clear`, `blockers=[]`, while the receipt proposes `after.gate=blocked`. After apply it reports `gate=blocked` with prerequisite `6ggjtrcpzph1`, reason `not-started`, direct=true. |
| Applied warning, no-op, cleared impact | Real add names the one applied task, no remaining IDs, newly-unsafe=true. Repeating it returns changed=false, impacts=[], and no remedy property. Remove returns `blocked→clear`, newly-unsafe=false, and no remedy. An independent human corpus with the same stable IDs/initial bytes prints exactly the JSON advice. |
| Refusal and classification | Self-dependency exits **11**, JSON code `validation`, no `error.dependency_mutation`, actionable no-write text. Missing dependency fails before planning, has no plan/applied IDs, and suggests `task depend repair`. That exact diagnostic command executes successfully against the broken graph and returns the source-level missing-dependency defect. Omission of the pre-write payload is intentional. |
| Refused planned impacts remain proposed | `TestAuditRealPreplanCASAndCleanupReceipts/cas` raw-edits an unrelated task after planning, before whole-graph CAS. Retry is disabled to observe the refusal. Error remains `ErrConflict`, the target bytes do not change, applied=[], remaining=[target], but the full proposed unsafe impact remains in the typed receipt. |
| Every durable migration prefix survives | `TestAuditRealConflictEveryPrefixAndLiveResume` interrupts the real store after writes 1/2/3 with `domain.ErrConflict` and configures four possible retries. All stop after **one** mutation call: applied counts 1/2/3, remaining 2/1/0, planned count 3. Byte changes match exactly the applied ID set; bodies remain intact, applied legacy keys disappear, and every prefix remains non-broken. All-applied advice explicitly says writes already landed. |
| Resume uses the live graph | After each incomplete prefix, the reviewer replaces a remaining `blocked_by` declaration with `[]` and appends an operator note. Explicit resume is the second mutation call, writes only remaining task files, preserves the note, and produces **no** canonical dependency on that edited task. It does not replay the abandoned full plan. Final graph is healthy. |
| Last-write/cleanup conflict stays visible | A real single-file dependency add followed by guard-cleanup `ErrConflict` returns applied=[target], remaining=[], full unsafe impact, all-applied text, and **one** mutation call despite retry allowance 4. No successful no-op replaces the failure. |
| Lifecycle durability is separate | `TestAuditRealLifecycleCleanupNeverRetries` reopens a completed prerequisite, then injects cleanup conflict. The receipt is committed with one newly unsafe downstream impact and stable-ID advice, the prerequisite status really changed, downstream file bytes did not, and lifecycle calls=1 despite retry allowance 4. Its emitted blocker command executes on the live graph and exposes the now-inconsistent completed dependent. |
| False flags do not erase useful advice | `TestAuditAdviceSurvivesFalseFlagsAndEmptyTaskImpacts` passes for dependency guidance with unchanged blocked state, and Moves success/committed-error guidance with no task impacts and false Thread flags. Nonempty remedy is still printed. |
| All shared wire shapes accept populated fields | `TestAuditPopulatedSharedSchemaBranches`: 24 subtests across dependency success, Moves, repair, `error.dependency_mutation`, `error.task_lifecycle`, and `error.graph_repair`; true/false task flags in every impact and true/false Thread flags in the four Thread-bearing shapes. Dependency/lifecycle advice is nonempty or omitted when empty; repair has no remedy field and those remedy loop variants repeat its shape. Every emitted object validates against its own `$defs` entry. Removing a required task/Thread boolean is rejected. These are converter/schema tests, not claims that a constructed fixture authorized a domain write. |
| Completed Thread consumer can fail despite sound core ownership | The real sole-member completed-Thread reopen yields committed=true, task impacts=0, Thread impacts=1, newly inconsistent=true, nonempty core remedy. `Model.Update` drops that advice. This is the M1 followup, established independently of source planning documents. |

The CLI scenario is reproduced by `python3 .git/isolated-review-workspace/evidence/dogfood.py`; exact subprocess commands, statuses, stdout/stderr are in `dogfood-results.json` and `dogfood-*.out/.err`. It used `init --path "$SCRATCH" --taskflow-root planning --no-register` and a private config home. No global `-C` bootstrap assumption or source planning mutation was used. The dependency and lifecycle recommended blocker commands and the broken-graph diagnostic repair command were executed against the states recommending them. No destructive repair was guessed from a JSON target.

### Mutation effectiveness

All **19 probes compiled**, then failed for the claimed semantic reason (exit 1, never a build error). `mutations.py`, `mutations.json`, each named `.log`, `real_mutants.py`, and `real-mutants.json` retain the exact edits and test commands. Production files were restored from the sandbox baseline after every probe. The restored focused suite passed after the first batch; each real-store mutant was immediately followed by its restored passing test, and the final restored full suite passed after removing all reviewer test files.

| Mutant | Named regression that failed / observation |
| --- | --- |
| Core task always false | `TestTaskGraphStateImpactNewlyUnsafe`: newly blocked/broken, changed non-clear gate and new inconsistency literal expectations fail. |
| Treat non-clear changes as a severity/clear-only rule | Same test: blocked→broken and **broken→blocked** fail. |
| Core Thread always false | `TestThreadProjectionImpactNewlyInconsistent`: false→true fails. |
| Wire task flag false | `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` and `TestDependencyMutationPublishesCoreRecoveryIntent` fail. |
| Wire Thread flag false | `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` fails. |
| Wire dependency remedy empty | Same wire test fails retained-owner-advice comparison. |
| Moves rederives task policy from state strings | `TestMovesHumanUsesProjectedImpactFlagsNotWireStatePolicy` fails its contradictory false flag. |
| Moves rederives Thread policy from booleans | Same hostile test fails. |
| **Coordinated mapper + renderer duplicate the old raw task/Thread rules** | Wire test **passes** because current values agree; hostile Moves test **fails** because copied flags deliberately disagree with raw states. Agreement between nearby layers alone is insufficient ownership evidence. |
| Coordinated core task method false + mapper uses old raw rule | Wire test fails method/mapper disagreement, independently of the core literal table. |
| Dependency human invents old blocker/remove-edge advice | `TestDependencyHumanUsesCoreGuidanceWithoutInventingRecovery` fails empty/nonempty owner-advice cases. |
| Core dependency receipt loses remedy | `TestServiceDependencyAndLifecycleShareRecoveryIntent` and `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries` fail. The error-text fallback does not rescue the structured receipt. |
| Remove dependency durable-prefix retry stop | `TestServiceDependencyRefusalAndCommittedFailureRecovery` and `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries` fail: call count becomes 2 and a successful/no-op receipt hides the original error. |
| Task preview advice uses applied wording | `TestServiceDependencyAndLifecycleShareRecoveryIntent` fails preview qualification. |
| Dependency failure stops unwrapping cause | Core typed failure/sentinel test and `TestWriteErrorCarriesStructuredDependencyMutationRecovery` fail; JSON classification changes from conflict to generic error. |
| Thread preview advice uses applied wording | `TestLifecycleThreadRecoveryQualifiesPreviewAndRetainsOverride` fails its preview context. |
| Remove dependency retry stop with **real FS** | Reviewer `TestAuditRealConflictEveryPrefixAndLiveResume` fails all 1/2/3 cases and `TestAuditRealPreplanCASAndCleanupReceipts/cleanup` fails. Calls become 2; final prefix/cleanup becomes successful no-op, losing durability evidence. |
| Remove lifecycle committed retry stop with **real FS** | `TestAuditRealLifecycleCleanupNeverRetries` fails: calls=2, returned changed=false/committed=false and error=nil although the first write landed. |
| Gate renderer advice on an unsafe/count condition | `TestAuditAdviceSurvivesFalseFlagsAndEmptyTaskImpacts` fails at the dependency false-warning case, proving nonempty advice is not disposable formatting. This combined probe was killed at that first assertion; its Moves edit is not claimed as independently killed. |

### Second pass: challenged abstractions and settled hypotheses

1. **A mapper compared to its own policy source proves ownership.** Counterexample: duplicate the raw gate/inconsistency rules in wire and Moves together. The wire oracle remains green. The dissonant copied-flag Moves fixture fails, and the separate core literal predicate table kills policy changes. Then change the core method while leaving a raw-rule mapper: the wire oracle fails. Result: the weakness of the mapper test in isolation is demonstrated; current production ownership survives the coordinated challenge. No policy-duplication finding is inferred merely from the weak oracle.
2. **A fake port's applied prefix is enough evidence for real recovery.** Counterexample attempt: force actual filesystem writes to return conflict after every atomic prefix, after the last write, and during unlock; inspect bytes, IDs, call counts, graph health, and live-edit resumption. Current code preserves and exposes every event. Removing the stop condition makes both portable and real tests swallow the event into a successful remainder/no-op, including cleanup. Result: the stronger-durability concern is falsified for these real paths; claims are limited to the tested port/platform rather than generalized to an unshipped backend.
3. **Preserving the lifecycle receipt means every UI surfaces its guidance.** A real completed Thread with one completed member and no descendants is reopened without force. Core computes the right Thread warning and remedy; TUI drops it because the task impact list is empty. Result: demonstrated presentation loss, M1, pre-existing relative to `289b0ba`. No speculative warning redesign or wider refactor is required.

Strongest owner tests: the independent expected-value table in `TestTaskGraphStateImpactNewlyUnsafe`, especially broken→blocked; the contradictory-flag `TestMovesHumanUsesProjectedImpactFlagsNotWireStatePolicy`; and portable retry tests asserting call counts plus typed durable receipts. The existing real migration prefix test is useful for bytes/health/resumption, but injects a generic error, so it alone cannot detect retrying committed **conflict**; the reviewer real conflict probes settle that gap.

Weakest owner tests: the wire test derives expected flags from the same core method, so duplicated equivalent policy can pass; the registry-wide schema test enumerates envelopes but its repair arrays and nested dependency-error impact array are empty, so registry coverage alone does not exercise new required flags there; and the existing TUI reopen fixture includes a descendant, masking M1. The reviewer supplied hostile evidence for each instead of treating broad test names as proof.

### Validation commands and results

Environment: `go version go1.27.1 darwin/arm64`; `uname -sm` → `Darwin arm64`. All commands below ran inside the independent clone. Go and lint used private writable caches; no golden `-update` or generator overwrite was used.

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/core ./internal/wire ./internal/cli/render ./internal/cli -run 'TestTaskGraphStateImpactNewlyUnsafe|TestThreadProjectionImpactNewlyInconsistent|TestServiceDependency|TestDependencyPreviewRefusal|TestLifecycleThreadRecovery|TestImpactWireCopiesCoreDecisionsAndDependencyRemedy|TestJSONSchema_ValidatesRealOutput|TestDependencyHumanUsesCoreGuidance|TestMovesHumanUsesProjectedImpactFlags|TestDependencyMutationPublishesCoreRecoveryIntent|TestWriteErrorCarriesStructuredDependencyMutationRecovery|TestTaskDependAddRejectsCycle' -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test ./... -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./... -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/store -run TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes -count=20
GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache golangci-lint run ./...
GOCACHE=/tmp/taskflow-review-go-cache just build
GOCACHE=/tmp/taskflow-review-go-cache go run ./internal/tools/docgen -out .git/isolated-review-workspace/evidence/generated-cli
GOCACHE=/tmp/taskflow-review-go-cache go run ./internal/tools/schemacomments -out .git/isolated-review-workspace/evidence/schema_comments.json
diff -qr docs/cli .git/isolated-review-workspace/evidence/generated-cli
cmp internal/wire/schema_comments.json .git/isolated-review-workspace/evidence/schema_comments.json
TSKFLW_CONFIG_HOME="$PWD/.git/isolated-review-workspace/evidence/home" ./bin/tskflwctl -C . lint --json --no-input
```

Every command returned **0**. Focused filters intentionally match prefixes: `TestTaskDependAddRejectsCycle` selects the verified `TestTaskDependAddRejectsCycleWithValidationExit`, and `TestDependencyPreviewRefusal` selects `TestDependencyPreviewRefusalDoesNotClaimAppliedChanges`. Logs: `baseline-full.log`, `baseline-race.log`, `baseline-prefix-race.log`, `baseline-lint.log`, `baseline-build.log`, `generated.log`, and `planning-lint.json`. Lint JSON is `unreadable=[]`, `issues=[]`. Build produced a working binary; Go also emitted a nonfatal denied module stat-cache write outside the sandbox, recorded in its build log. The CLI dogfood used that binary successfully. Fresh CLI docs and 267 schema comments match checked-in output exactly.

Additional independent commands:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/store ./internal/wire -run 'TestAuditReal|TestAuditPopulatedSharedSchemaBranches' -count=1 -v
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/cli/render -run TestAuditAdviceSurvivesFalseFlagsAndEmptyTaskImpacts -count=1 -v
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/store ./internal/wire ./internal/cli/render -run 'TestAuditReal|TestAuditPopulated|TestAuditAdvice' -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/tui -run '^TestAuditTUIThreadOnlyRemedy$' -count=1 -v
```

First three pass (exit 0); TUI exits 1 with the demonstrated M1 assertion, not a compilation error. `independent-pass.log`, `false-flags-guidance.log`, `independent-race.log`, and `independent-probes.log` contain the outputs. After all probes were archived/removed, `GOCACHE=/tmp/taskflow-review-go-cache go test ./... -count=1` passed again (`restored-full.log`). Full baseline race and lint passed; no claim is made that the failing temporary TUI assertion is a shipped test.

Compatibility check: `internal/wire/wire.go:339` documents revision 1.81; required flags and optional dependency remedy agree with schema and output. README, `docs/ARCHITECTURE.md`, and generated dependency reference distinguish warning from durability. Existing persisted Markdown schema, authorization/eligibility predicates, and repair/migration mechanics were unchanged in the implementation delta. Machine goldens and `machine_contract_revision.txt` advance to 1.81; full tests execute `TestGolden_MachineContract`, `TestGolden_ProjectionContract`, and `TestMachineGoldenRevisionMatchesSchemaVersion`. Existing repair golden has empty impacts, so its pass was supplemented with populated schema cases rather than cited as evidence for those branches.

Limitations: hook-injected failures exercise real atomic writes and cleanup attribution, not physical disk-full or kernel unlock faults. The executed platform is macOS arm64; Windows runtime locking and a future second backend remain unverified. The fixed private fixtures test representative dependency/lifecycle paths, not every possible graph topology. No frontend screenshot or interactive TUI session was used; M1 follows the actual production command/message/update path over real files. No protocol deviation occurred: only the brief and initial helper copy were accessed in the shared checkout until final transfer.

### Isolation and guarded transfer attestation

The helper created an independent `--no-hardlinks` clone with an overlaid baseline. No additional commit, staging, branch switch, source restore/cleanup, or manual copy-back occurred. All mutations, tests, generators, report/finding writes, and scratch scenarios ran in this clone. Probe files were restored or archived in its private Git metadata; `git status --short` permits only this audit. M1 was allocated through the sandbox `audit finding new` command and remains open (its JSON receipt reports findings=1, open_findings=1).

Final helper commands, with stdout retained as `final-verify.log` and `final-transfer.log`:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- planning/audits/6ggjnz40teda-2026-10-04-core-owned-dependency-impact-recovery-implementation-codex.md
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YPM4xL
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YPM4xL/.git
baseline_commit=94d9205f0d54a21ea8a46fda252a6b6f87132ca4
source_blob=4816b4770db30181dad5cc78854b42884c78dc86
source_fingerprint=f50395ebd55515caef67344317659cfb055e85f8
deliverable=planning/audits/6ggjnz40teda-2026-10-04-core-owned-dependency-impact-recovery-implementation-codex.md
deliverable_changed=true
transfer=succeeded
```

Only this audit is transferred. The sandbox and private evidence are retained pending owner receipt confirmation.

## Owner triage (2026-10-04)

Accepted the review's evidence-backed verdict: no new core, warning-policy, retry,
or machine-contract defect was demonstrated. M1 reproduced independently through
the real sole-member completed-Thread move/message/update path; it is tracked by
[6ggkdbg0816h](../tasks/6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md)
in the TUI refinement Thread, after the core contract, not added to adapter-neutral
refactor closeout. The reviewer report above describes its submission-time state;
the finding status and resolution below record the owner's disposition.

Promoted the strongest supplementary evidence into permanent task-local tests:
the real migration prefix test now injects conflict with four retries available,
asserts a single mutation attempt and matching durable bytes, and resumes against
operator-edited remaining declarations. A schema matrix covers populated true/false
task and Thread warnings, optional remedy presence/omission, and structurally
deleted required flags across ordinary/repair/lifecycle success and error shapes.
An isolated guard-removal probe fails all three real conflict-prefix cases.
These are test-only improvements over the captured review baseline; production
code is unchanged and Antigravity's assigned audit is untouched.

The retained reviewer evidence was inspected; its source restoration/transfer
attestation is complete. Audit closed after the sole finding was handed to its task;
closure does not claim the TUI bug was fixed or that this implementation was merged.

## Findings

#### M1. TUI drops recovery guidance for a Thread-only lifecycle impact (followup) · **Status:** tracked by 6ggkdbg0816h

**File:** internal/tui/model.go:472 | **Component:** tui
**Effort:** S · **Urgency:** eventually

This is a demonstrated, pre-existing presentation defect and a bounded followup; it does not block the core-owned dependency-impact implementation.

Reproduction: the private `probe-tui.go` contains `TestAuditTUIThreadOnlyRemedy`. It writes a real completed task with no descendants and a valid completed Thread containing that task, then invokes the production `moveTask` command to reopen the task to `ready-to-start` and feeds the resulting `movedMsg` through `Model.Update`. Running `GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/tui -run '^TestAuditTUIThreadOnlyRemedy$' -count=1 -v` fails after the successful durable transition. The receipt has `Committed=true`, zero task impacts, one Thread impact with `NewlyInconsistent()=true`, and remedy `inspect each newly inconsistent Thread and restore sound member or external-gate evidence`. The flash is only `moved audit-single-member → ready-to-start`.

Locations: `internal/tui/entity.go:356` preserves the real lifecycle receipt; `internal/tui/model.go:472` nests remedy output under `len(msg.lifecycle.Impacts) > 0`. `internal/core/thread_projection.go:308` marks the completed Thread inconsistent after its member stops being drained, and `internal/core/service_task.go:587` supplies Thread recovery advice independently of downstream task impacts. The success path is reachable without a forced transition or fabricated receipt. `git diff 289b0ba -- internal/tui/model.go` is empty, and the non-preview Thread remedy already exists in that baseline, establishing that this handoff did not introduce the defect.

Impact: a TUI user can successfully reopen work underneath a completed Thread without seeing the core-owned inconsistency warning or its recovery guidance. Core callers and CLI/wire receipts retain the evidence; the loss occurs at TUI presentation. The existing `TestTUITaskReopenSurfacesDescendantImpactsAndRemedy` fixture has a downstream task, so it exercises the branch that hides this defect.

Pin the followup with the same real-filesystem, sole-member completed-Thread transition and require the TUI flash to retain the exact receipt remedy when `Impacts` is empty. Include a nonempty-Thread-impact presentation assertion driven by the core flag. No source fix, task-state change, or disposition change was made in this audit.

**Recommendation:** Print nonempty lifecycle Remedy independently of downstream task counts, and surface Thread impacts using the core warning decision.

**Resolution:** Confirmed independently by the implementation owner with a real
completed sole-member Thread and production moveTask -> Model.Update in an
isolated temporary copy: committed=true, zero task impacts, one newly
inconsistent Thread impact, nonempty core remedy, but the flash omits it.
Followup 6ggkdbg0816h owns unconditional remedy presentation, core-driven Thread
impact feedback, and command/message/update regressions. Added to the TUI
refinement Thread after 6ggdkztshmzz; this pre-existing consumer defect does not
expand the adapter-neutral closure path.
