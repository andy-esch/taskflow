---
schema: 1
id: 6g4tf2yr4nb0
bucket: closed
area: dependency-eligibility-enforcement-codex
date: "2026-08-29"
updated_at: "2026-08-29"
---
# Audit: Dependency eligibility enforcement — Codex adversarial review — 2026-08-29

> Edit findings through `tskflwctl audit finding` so status and resolution metadata remain queryable.

Independent adversarial review of the uncommitted implementation on `feat/dependency-eligibility-enforcement` for task `6g3q4rte8kc1-enforce-dependency-eligibility-across-every-task-start-path`, ADR-0006, and the lifecycle amendment in `docs/ARCHITECTURE.md`. The review covered lifecycle/dependency semantics, every first-party status path, typed force scopes, graph health, resulting task validity, guard ownership, planner re-entry, cooperating/raw writers, CAS windows, create-and-start, CLI/TUI parity, editing bypasses, timestamps/no-ops, impact determinism, receipts/schema/recovery, and batch scaling.

**Verdict:** not ready to close. The centralized policy and cooperating-writer architecture are sound, but H1 permits a mutation to commit and then be reported as failed. M1 leaves an internal creation capability able to manufacture `in-progress` tasks outside the advertised lifecycle boundary. The remaining findings are concrete coverage and adapter-contract gaps.

## Findings

### High

#### H1. A guard-release error can report a committed lifecycle transition as failed and make it retryable  · **Status:** fixed

**File:** `internal/store/lifecyclemutation.go:36-45,97-113`; `internal/core/service_task.go:255-264`; `internal/core/retry.go:54-71`; `internal/cli/moves.go:24-41`; `internal/tui/entity.go:233-255` · **Component:** lifecycle outcome / recovery · **Effort:** M · **Urgency:** acute

**Class:** implementation defect. · **Disposition:** blocking fix in this branch.

`FS.MutateTaskLifecycle` performs the create or replacement and only afterward runs the deferred unlock. If unlock fails, it returns the fully populated `TaskLifecycleMutationResult` together with an error. `TestMutateTaskLifecycleAttributesReleaseFailureAfterCommit` explicitly proves that combination: `result.Changed` is true, the function returns an error, and the task file is already `in-progress`.

The higher layers erase the distinction between “failed before commit” and “cleanup failed after commit”:

- `retryOnConflict` assumes every `ErrConflict` is raised before a write and reruns the whole closure. A release failure wrapping `ErrConflict` therefore retries an already committed move up to four times. For create-and-start, the retry sees the just-created stable ID and converts the successful creation into an apparent collision/conflict.
- `runMoves` ignores `got` whenever `err != nil`, emits only `MoveResult.error`, and omits the lifecycle receipt. The CLI can therefore say a transition failed although it committed.
- The TUI likewise turns any error into `actionErrMsg`, discards the receipt, and deliberately skips the success reload path, leaving both the message and visible state stale.

This is exactly the dangerous unknown-outcome case: a user or agent may retry or compensate for a transition that is already durable.

**Recommendation:** introduce a typed lifecycle mutation failure/outcome analogous to `DependencyMutationFailure`, carrying the receipt and an explicit durable/committed phase. Never auto-retry a lifecycle result once the store reports that its write committed, even when cleanup wraps `ErrConflict`. Teach human output, JSON error output, and the TUI to say that the transition committed but guard cleanup failed, include the task/workspace receipt, and force a TUI refresh. Add service and adapter tests for ordinary and `ErrConflict`-wrapping unlock failures on both existing moves and `new --start`.

**Resolution:** Added an explicit committed phase and typed lifecycle failure
carrying the receipt; service retries stop after commit, CLI human/JSON recovery
retains task and workspace detail, and the TUI warns and reloads. Tests cover
ordinary and conflict-wrapping cleanup failures for existing moves and
create-and-start.

### Medium

#### M1. The ordinary creation port can still persist an unauthorized `in-progress` task  · **Status:** fixed

**File:** `internal/core/store.go:27-32`; `internal/store/create.go:132-161`; `internal/store/create.go:88-103` · **Component:** lifecycle capability boundary · **Effort:** S · **Urgency:** soon

**Class:** implementation defect / latent first-party bypass. · **Disposition:** fix in this branch before declaring the lifecycle capability exclusive.

`TaskStore` says lifecycle changes are absent and `TaskLifecycleMutationStore` is the only status-write capability, but the same interface exposes `CreateTask(domain.Task, ...)`. `FS.CreateTask` validates identity and graph-owned fields but neither validates nor restricts `Task.Status`; `taskFields` serializes the caller-supplied status verbatim. A direct internal caller can therefore create an `in-progress` task without graph authorization, `started_at`, the repository lifecycle guard, impact analysis, or a lifecycle receipt. It can also persist an invalid status.

The current `Service.NewTask` caller is safe—it constructs only `ready-to-start`/`next-up` and routes `Start` through the guarded create-and-start path—so no shipped CLI/TUI bypass was found. The defect is that the stated capability invariant is convention rather than enforcement, making a later first-party adapter or refactor able to bypass it accidentally.

**Recommendation:** make ordinary creation accept only the explicitly non-start creation states used by the product, or replace the arbitrary `domain.Task` status input with a semantic non-start creation request. Reject `in-progress` and invalid statuses inside the store boundary and add a direct-store regression test. Keep guarded create-and-start as the only persistence route that can create `in-progress`.

**Resolution:** Ordinary FS.CreateTask now accepts only ready-to-start or
next-up and rejects in-progress, terminal, empty, and invalid statuses; guarded
create-and-start is the sole creation path into in-progress.

#### M2. The advertised lifecycle-versus-lifecycle/dependency race coverage does not exercise cooperating writers  · **Status:** fixed

**File:** `internal/store/lifecyclemutation_test.go:126-233`; `planning/tasks/6g3q4rte8kc1-enforce-dependency-eligibility-across-every-task-start-path.md:49-58,82` · **Component:** concurrency tests · **Effort:** M · **Urgency:** soon

**Class:** testing gap; the implementation design appears correct. · **Disposition:** add branch coverage or narrow the task's coverage claim.

The task requires racing start against prerequisite lifecycle changes and dependency mutations. The new tests named for those races instead use package hooks plus raw `os.WriteFile` before the whole-graph verification. They prove snapshot-token detection, and the target hook proves the immediate target check, but they do not prove that two real cooperating operations share the same canonical-root lock, serialize across distinct `FS` instances/processes, re-enter with fresh graph state, and preserve default/forced authorization.

The production paths do appear to use the same checked repository guard, and the broader guard suite covers cooperating graph writers. This is therefore not evidence of a current bypass. It is a regression hole at the lifecycle integration seam and makes the task's specific adversarial-coverage claim stronger than its tests.

**Recommendation:** add deterministic two-store or child-process tests using actual lifecycle and dependency services. Hold one operation inside the lifecycle boundary, start the competing mutation, assert serialization, release the first operation, and verify the second either re-authorizes from fresh state or refuses without a stale start. Cover prerequisite reopen/default start and dependency add/remove under both default and dependency-gate override paths.

**Resolution:** Added deterministic two-FS cooperating-writer tests using the
real dependency and lifecycle services. They hold the canonical guard, prove the
competing start waits, then verify fresh default and dependency-override
authorization after dependency add/remove and prerequisite reopen.

#### M3. JSON eligibility failures flatten the typed blocker result into prose  · **Status:** fixed

**File:** `internal/core/task_lifecycle.go:97-123`; `internal/cli/moves.go:24-41`; `internal/wire/envelopes.go:176-194`; `internal/cli/transition_test.go:37-64` · **Component:** CLI JSON contract / recovery · **Effort:** S · **Urgency:** soon

**Class:** machine-contract and test gap. · **Disposition:** fix in branch to satisfy human/machine failure parity, or explicitly track an additive schema follow-up before task closeout.

Core deliberately models a machine-inspectable `TaskEligibilityError` with task ID, derived state, and deterministic blockers. `runMoves` immediately flattens it to `err.Error()`, and `wire.MoveResult` has only `error` plus an optional success-side `lifecycle` payload. A failed `task start --json` therefore requires consumers to parse English to discover the role, gate, blocker IDs/reasons/paths, or the recovery action. In a mixed batch this loss is per-item, where a separate top-level error envelope cannot reconstruct the result safely.

The existing CLI test checks structured JSON only for the forced success path; the default refusal is tested only as human stderr text. That does not establish the task's requested human/machine receipt parity.

**Recommendation:** preserve typed lifecycle failure detail in each failed move row—preferably an additive `lifecycle_failure` or a lifecycle result with an explicit outcome—while retaining the existing `error` string for compatibility. Include state, blockers, override eligibility, and remedy, bump/regenerate the schema as required, and add mixed-batch human/JSON parity goldens.

**Resolution:** MoveResult now retains a typed lifecycle_failure with state,
blockers, requested override, override eligibility, and remedy while preserving
prose error compatibility. Mixed-batch JSON coverage proves success and refusal
remain attributable per item.

#### M4. The TUI routes through the policy but discards successful lifecycle impacts and remedies  · **Status:** fixed

**File:** `internal/tui/entity.go:233-255`; `internal/cli/render/render.go:239-261`; `internal/core/task_lifecycle.go:81-95` · **Component:** TUI lifecycle receipt parity · **Effort:** S · **Urgency:** soon

**Class:** adapter/UX defect. · **Disposition:** fix in branch if TUI receipt parity is part of acceptance; otherwise open a bounded follow-up and make that scope decision explicit.

The TUI correctly calls the shared service, so it cannot bypass eligibility. On success, however, `moveTask` and `deferTaskCmd` discard `TaskLifecycleReceipt` and send only `{slug,to,revisit}`. This hides descendant state changes and the recovery remedy. A TUI user can reopen an upstream completed task, make completed descendants inconsistent, and receive only a generic success flash; the CLI prints the impact count, each state change, and the remedy.

**Recommendation:** carry the receipt in `movedMsg` and surface at least impact count plus remedy in a durable flash/modal/detail affordance, with tests for an upstream reopen. Keep the existing reload, but do not treat reloading rows as a substitute for explaining a cross-task consequence.

**Resolution:** Task TUI moves now carry lifecycle receipts. Cross-task impacts
and remedy are surfaced in the flash, while committed cleanup failures warn and
force a reload instead of appearing uncommitted.

### Low

#### L1. Same-status no-ops bypass destination-specific override validation  · **Status:** fixed

**File:** `internal/core/task_lifecycle.go:227-250`; `internal/core/service_task.go:267-275` · **Component:** force-scope validation / no-op receipts · **Effort:** XS · **Urgency:** eventually

**Class:** implementation defect with no durable mutation. · **Disposition:** small branch fix or tightly scoped follow-up.

`validateExistingLifecyclePlan` returns a no-op at lines 235-238 before checking whether the typed override is legal for the destination. Thus a same-status generic move with `--force` can succeed for a destination where that override is forbidden—for example, `task move <already-ready-task> ready-to-start --force`. The receipt can expose `override: dependency-gate` with `forced: false`, even though dependency-gate override is valid only when entering `in-progress`.

No unauthorized transition commits, but the ordering weakens the typed force contract and makes no-op JSON misleading.

**Recommendation:** validate override/destination scope before the same-status early return, then retain byte-identical no-op behavior. Add table coverage for legal and illegal overrides on same-status moves.

**Resolution:** Destination-specific typed override validation now precedes the
same-status early return; illegal forced no-ops fail while ordinary no-ops
remain byte-identical.

## Confirmed behavior and dispositions

### Implementation behavior found sound

- Every current CLI and TUI lifecycle action routes through `Service.Move` or `Service.DeferTask`; `new --start` routes through the guarded create-and-start plan. No shipped adapter path that starts via generic set/edit was found.
- `task set` and store field mutation reject status, while `task edit` rejects every accepted status delta. Create-and-start validates a prospective ready candidate, active fields, absence of graph-owned fields, and uses exclusive creation under the repository guard.
- The dependency-gate and acceptance-criteria overrides are distinct core values. Apart from L1's no-op validation ordering, force cannot bypass candidate role, global graph health, active-field validity, or completion criteria outside its own scope.
- Global graph-health failure is intentionally repository-wide and fail-closed, even for unrelated lifecycle actions. ADR-0006 and the task choose this behavior; it is a design tradeoff, not a defect.
- Descendant impacts use deterministic downstream traversal and compare complete derived states. Direct dependency impacts are stable-ID sorted. Lifecycle timestamp updates and ordinary same-status no-ops are stable; leaving deferred clears `revisit_at`.
- Wire schema version `1.52`, empty-array normalization, and receipt mapping are covered by the passing full suite and schema goldens.

### Explicit design tradeoffs and non-blocking work

- The root lock is advisory. Cooperating CLI/TUI writers are serialized, but a raw editor can write after the whole-graph scan or after the target hash check and before rename. A raw prerequisite edit in that final interval can invalidate authorization; a raw target edit can be overwritten. `docs/ARCHITECTURE.md:198-209` and ADR-0006's 2026-08-27 amendment accurately say the immediate check narrows but cannot eliminate this window. Do not describe the raw-writer path as airtight or as a true filesystem CAS; the supported guarantee is atomic only among cooperating writers. No implementation change is required unless the product expands its guarantee beyond that documented boundary.
- Batch commands deliberately commit each item independently and report per-item outcomes. Earlier successes remain when a later item fails; this is established CLI behavior, not an atomic-batch promise. Recovery is to inspect the complete receipt, repair/refuse the failing item, and rerun idempotently—subject to H1 being fixed so a reported failure has a known durability state.
- Each dry-run item loads one full graph; each changed item reloads it for verification. On the current 280-task corpus, a locally built branch binary produced warm dry-run timings of approximately 0.12 s for 10 repeated items, 0.59 s for 50, and 1.19 s for 100. The linear behavior is acceptable for the current corpus and supports deferring an invocation-local cache. A benchmark/scan counter should precede any optimization because cache invalidation would interact with authorization correctness.

## Validation

- `go test ./...` — passed.
- `go test -race ./...` — passed.
- `go vet ./...` — passed.
- `git diff --check` — passed.
- `golangci-lint run ./...` and `just lint` — not independently reproducible in this environment; the installed golangci-lint 2.12.2 reports `context loading failed: no go files to analyze`. No `go mod tidy` or other write workaround was run.
- Performance sample used a branch binary built with `GOCACHE=/tmp/taskflow-review-go-cache GOPROXY=off`, then dry-run same-status batches only; it did not mutate planning data.

## Recommended closeout order

1. Fix H1 and add explicit committed-outcome recovery tests before merge.
2. Close the latent creation capability in M1.
3. Add the real cooperating-writer integration coverage in M2.
4. Decide and record whether M3/M4 are branch acceptance or bounded follow-ups; do not mark human/JSON/TUI parity complete while the structured data is discarded.
5. Take L1 opportunistically; retain the documented raw-writer and batch-scan tradeoffs.
