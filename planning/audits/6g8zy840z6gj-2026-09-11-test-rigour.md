---
schema: 1
id: 6g8zy840z6gj
bucket: open
area: test-rigour
date: "2026-09-11"
---

# Code Quality Audit: test-rigour — 2026-09-11

> Edit findings through `tskflwctl audit finding` so status and resolution
> metadata stay queryable. Never hand-edit a `**Status:**` line.

Routine: `code-quality-audit` · lens `test-rigour` · ISO week `2026-W37`,
slot `Fri` (index 3).

The lens question: does the suite prove **rejection** as rigorously as success?
Method was guard-coverage tracing — enumerate the load-bearing predicates in the
surfaced packages, name the repository state that trips each, then look for a test
that CONSTRUCTS that state and ASSERTS the refusal.

## Punch list

M1. `epic set --unset description` writes a lint-invalid epic  (effort: XS · urgency: soon)
M2. `TaskLifecycleOverride` closed vocabulary has no exhaustiveness or negative test  (effort: S · urgency: eventually)
L1. `revisit_at`-only-for-deferred guard has no negative test  (effort: XS · urgency: eventually)
L2. `TestService_NewEpic`'s failure table proves less than it claims  (effort: XS · urgency: eventually)
L3. Dependency planner guards the dependent against an unreadable record, not the prerequisites  (effort: XS · urgency: eventually)

## Files audited

- **Signal**: `internal/core/service_epic_test.go` — top score: 6 commits in 30d ×
  `internal/core` referenced by 131 files (log(7)×log(132) ≈ 9.5).
- **Signal**: `internal/core/dependency_graph_test.go` — 5 commits in 30d over the
  same core blast radius (≈ 8.8); the guarded-mutation surface epic 30 is still
  landing into.
- **Adjacency** (from `service_epic_test.go`): `internal/core/service_epic.go`
  (+ `usecases_test.go`, `setepicfields_test.go`) — the test-rigour hop from a test
  file to the production predicates it claims to cover.
- **Adjacency** (from `dependency_graph_test.go`): `internal/store/graphmutation.go`
  + `internal/store/graphmutation_test.go` — the hexagonal-seam hop, from a `core`
  invariant to the `store` guard that enforces it.
- **Random**: `internal/spacestore/fs_test.go` (+ `internal/spacestore/fs.go`).

Two files were pulled in by the high-yield rule ("a guard with no negative test"):
`internal/wire/envelopes_test.go` (2nd-highest churn signal, the `--json` contract
seam) and `internal/core/task_lifecycle.go` + `task_lifecycle_test.go`.

## Commands run

Build (`just` is not installed in this environment; the spec's documented fallback
was used):

```
go build -o bin/tskflwctl ./cmd/tskflwctl     # exit 0
```

Signal and blast radius:

```
git log --since="30 days ago" --name-only --pretty=format: \
  -- 'internal/**/*_test.go' 'internal/cli/testdata/' | ...
grep -rln "internal/core\"" --include=*.go internal cmd | wc -l    # 131
grep -rln "internal/store\"" --include=*.go internal cmd | wc -l   # 26
grep -rln "internal/wire\"" --include=*.go internal cmd | wc -l    # 41
```

Guard sweep — every `fmt.Errorf("%w: ...")` fragment in `internal/core` and
`internal/store` (200 unique), cross-referenced against every `_test.go`. The
message-based hits are noisy by design (good tests assert the sentinel, not the
text), so each candidate was then confirmed by reading the test.

Envelope exhaustiveness, checked mechanically:

```
declared *Envelope types: 53 · cases in TestJSONSchema_ValidatesRealOutput: 53
comm -23 declared tested  →  (empty)
```

M1 was reproduced against a throwaway planning repo under the scratchpad (never
this tree):

```
tskflwctl init && tskflwctl epic new "Probe Epic" --description "a required description"
tskflwctl epic set 01-probe-epic --unset updated_at
  → error: validation failed: updated_at is stamped automatically and cannot be set
tskflwctl epic set 01-probe-epic --unset status
  → error: validation failed: epic status moves via `epic move`, not `set`
tskflwctl epic set 01-probe-epic --unset descriptionn
  → error: validation failed: unknown epic field "descriptionn"
tskflwctl epic set 01-probe-epic --unset description
  → ✔ updated 01-probe-epic            # <-- the finding
tskflwctl lint
  → 01-probe-epic
      description: missing
    error: validation failed: 1 item(s) with issues, 0 unreadable file(s)
```

Validation of this audit is recorded in the PR body.

## Findings

### Critical

(none)

### High

(none)

### Medium

#### M1. `epic set --unset description` writes a lint-invalid epic  · **Status:** open

**File:** `internal/core/service_epic.go:302` | **Component:** core (epic mutation)
**Effort:** XS · **Urgency:** soon

`SetEpicFields` rejects setting or unsetting `status` (:297) and `updated_at`
(:300), and gates unknown field names on the epic registry in both the set (:311)
and unset (:305) branches. Nothing rejects **unsetting a field the epic contract
requires**. `description` is required unconditionally by `NewEpic` (:29–31) and
required unconditionally by `lint`, so `--unset description` drops an epic straight
into the invalid region.

This contradicts a stated non-negotiable in `docs/ARCHITECTURE.md`: *"Reads stay
tolerant so `lint` can REPORT malformed data already on disk; writes refuse to
create it."* Here the write path creates exactly the state the lint path reports.

Note the task path is **not** the same bug and should not be "fixed" to match: a
description-less task is a legal `unstarted` state, and `task start` refuses it at
the lifecycle gate (`a description is required for a next-up/in-progress task`).
Epics have no pre-active state, so for epics the unset has nowhere safe to land.

**Failing scenario:** in any planning repo —
`tskflwctl epic new "Probe Epic" --description "d"` then
`tskflwctl epic set 01-probe-epic --unset description` → `✔ updated 01-probe-epic`,
and the very next `tskflwctl lint` fails with `01-probe-epic / description: missing`.
Reproduced end-to-end; see **Commands run**.

**Why tests didn't catch it:** the epic unset path has **zero** tests. No
`domain.UnsetField` value is ever passed to `SetEpicFields` anywhere in the repo —
not in `service_epic_test.go`, not in `setepicfields_test.go`, not in
`internal/cli/epic_test.go`. The task analogue is covered twice over, including
`TestSetFields_UnsetRejectsUnknownField`
(`internal/core/setfields_coercion_test.go:172`) whose own comment reads *"guards the
gate that the unset path once skipped"* — the same class of bug, already found and
fixed once on the sibling surface, never checked for on this one.

**Recommendation:** in the `UnsetField` branch of `SetEpicFields`, reject unsetting
an epic field the contract requires (today: `description`), mirroring the shape of
the `status`/`updated_at` rejections directly above it. Then add the negative tests
the path has never had: unset `description` → `ErrValidation`; unset `updated_at` →
`ErrValidation`; unset `status` → `ErrValidation`; unset a typo'd field without
`--force` → `ErrValidation`; unset a genuine custom field with `--force` → ok.

**Tightening (adjacent):** `TestService_SetEpicFields_UnknownFieldNeedsForce`
(`service_epic_test.go:715`) claims to "mirror the task contract" but only covers the
set branch — extend it to the unset branch in the same commit.

**Follow-up:** "which fields are required" is currently re-asserted per service
(`NewEpic` and `LintEpic` separately). Declaring it once on the entity descriptor
(`domain/entity.go`) so create, set, unset, and lint all read one table is epic 26's
territory, not this fix.

#### M2. `TaskLifecycleOverride` closed vocabulary has no exhaustiveness or negative test  · **Status:** open

**File:** `internal/core/task_lifecycle.go:229` | **Component:** core (task lifecycle)
**Effort:** S · **Urgency:** eventually

`validTaskLifecycleOverride` closes the override vocabulary with `default: return
false`, and `ValidateTaskLifecyclePlan:211` turns that into `unknown task lifecycle
override %q`. **No test in the repository ever passes an out-of-vocabulary override
value** — every `TaskLifecycleOverride` mention across all `_test.go` files is one of
the three declared constants.

`TestValidateTaskLifecycleNoOpStillValidatesOverrideScope`
(`task_lifecycle_test.go:155`) looks like the covering test but is not: it iterates
the two *valid* overrides against a wrong target status, exercising the per-override
scope checks at :299/:303, never the vocabulary gate itself. That loop would keep
passing with `validTaskLifecycleOverride` broken.

**Context:** no live defect today — this is a regression-detection gap, which is why
it is Medium/eventually rather than a bug. It is reported at Medium because of what
it silently permits, not what it currently does.

**Failing scenario (as a regression, not today's behaviour):** add a fourth override
constant — say `TaskLifecycleOverrideLintInvalid` — and forget its `case` arm in
`validTaskLifecycleOverride`. Every plan carrying the new override is then rejected
with `unknown task lifecycle override "lint-invalid"`, the feature quietly does
nothing, and nothing goes red: no test asserts the vocabulary's membership, and
`.golangci.yml` enables only `standard` + `depguard` — the `exhaustive` linter, which
exists precisely for this, is off.

**Why tests didn't catch it:** `TaskLifecycleOverride` is the one closed vocabulary
in this codebase with no `All*()` accessor. `domain` publishes `AllStatuses`,
`AllEpicStatuses`, `AllAuditBuckets`, and `AllThreadStatuses`, each of which lets a
table test enumerate the real set; the override vocabulary is only ever spelled out
by hand at each use site, so an exhaustive test cannot be written today without
re-declaring the list — the exact duplication that lets vocabularies drift, and the
thing `domain/resolution.go` exists to prevent elsewhere.

**Recommendation:** copy the pattern this repo has already proved. The registry-derived
coverage guard in `internal/wire/envelopes_test.go:383–397` enumerates every registered
envelope via `reflect` so "a newly-added envelope can't be silently left unvalidated" —
it is the best exhaustiveness guard in the tree. Do the same, more cheaply: add
`AllTaskLifecycleOverrides()` beside the constants, then table-test that every member
passes `validTaskLifecycleOverride` and that a synthetic non-member (`"force"`) is
rejected with `ErrValidation` from `ValidateTaskLifecyclePlan`.

**Tightening (adjacent):** `.golangci.yml`'s header already frames linter tightening
as a tracked follow-up. Enabling `exhaustive` scoped to `internal/core` and
`internal/domain` would make every closed-vocabulary `switch` in the planning model
compile-checked rather than convention-checked.

### Low

#### L1. `revisit_at`-only-for-deferred guard has no negative test  · **Status:** open

**File:** `internal/core/task_lifecycle.go:214` | **Component:** core (task lifecycle)
**Effort:** XS · **Urgency:** eventually

`ValidateTaskLifecyclePlan` refuses a plan carrying `RevisitAt` for any target other
than `deferred`. No test constructs that state: the only test that puts a value in a
plan's `RevisitAt` (`internal/store/lifecyclemutation_test.go:38`) hardcodes
`To: domain.StatusDeferred`.

Labelled a smell rather than a defect because it is currently unreachable: the single
production site that sets `RevisitAt` (`internal/core/service_task.go:365`) hardcodes
`To: domain.StatusDeferred`, and `--until` is exposed only on `task defer`
(`internal/cli/task.go:962`).

It still matters as a core-API invariant. `core` is the seam a served adapter
(epic 19) drives directly, without the CLI's flag shape to constrain it, and the
malformed state this prevents is one the read side already knows about —
`domain/validate_test.go:110` pins that an `in-progress` task with a past
`revisit_at` is never due, i.e. the date would be silently inert.

**Recommendation:** one case in `task_lifecycle_test.go` —
`ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{TaskID: id, To: domain.StatusInProgress, RevisitAt: "2026-10-01"}, "")`
→ `errors.Is(err, domain.ErrValidation)`.

#### L2. `TestService_NewEpic`'s failure table proves less than it claims  · **Status:** open

**File:** `internal/core/usecases_test.go:362` | **Component:** core tests
**Effort:** XS · **Urgency:** eventually

The table asserts only `err == nil` — never *which* predicate fired. Its third case
is commented `// empty slug` (`{Title: "!!!", Description: "d", Priority: "medium"}`)
but carries no `Status`, and `""` is not in `epicStatuses` (`domain/epic.go:67`), so
`ValidateEpicStatus` at `service_epic.go:38` rejects it *before* the empty-slug guard
at :45 is ever reached. The case passes on a predicate it does not name.

Not a coverage hole on its own — the slug guard is genuinely covered by
`TestService_Create_EmptySlugStillErrors` (`usecases_test.go:288`), which does pass
`Status: "active"`. But it does mean `NewEpic` has no *intentional* negative test for
an invalid epic status, and a table that passes for the wrong reason is exactly what
lets a guard removal go green.

**Recommendation:** assert `errors.Is(err, domain.ErrValidation)` per case, give every
case a valid `Status: "active"` so each trips the predicate its comment names, and add
an explicit `{Status: "planning"}` case for the status guard.

#### L3. Dependency planner guards the dependent against an unreadable record, not the prerequisites  · **Status:** open

**File:** `internal/core/dependency_operations.go:191` | **Component:** core (dependency planner)
**Effort:** XS · **Urgency:** eventually

`planDependencyEdges` resolves the dependent and then checks it with
`graph.Task(dependentID)`, returning `ErrValidation` if the ID belongs to an
unreadable record (:189–192). Each prerequisite goes through the same
`graph.ResolveTaskID` (:197) and gets no such check, even though that method's own
doc states *"Exact unreadable IDs remain addressable for diagnostic queries"*
(`dependency_graph.go:840`).

Not a live defect, and deliberately not a test gap: neither branch is reachable. An
unreadable record raises `ProblemUnreadable` (`dependency_graph.go:357`), any problem
sets health to `GraphBroken` (:526), and `ValidateTaskGraphMutationSource`
(`dependency_graph_mutation.go:18`) rejects a broken source before the planner is
invoked — in the real adapter (`store/graphmutation.go:57`) and in the core test fake
alike (`dependency_operations_test.go:31`, which mirrors the real precondition
faithfully — worth noting, since a fake that skipped it would have been the more
serious finding).

The asymmetry only becomes load-bearing if that strict gate is ever relaxed to admit
`degraded` sources. Until then it encodes a precondition the code does not actually
make.

**Recommendation:** make the two sides read the same — either drop the unreachable
dependent check with a comment naming `ValidateTaskGraphMutationSource` as the real
gate, or add the symmetric prerequisite check. Do **not** add a test for an
unreachable branch; the honest fix is in the source, not the suite.

## What audited clean

- `internal/core/dependency_operations.go` + `dependency_operations_test.go` —
  every planner guard that is reachable has a test that constructs the state and
  asserts the sentinel: ambiguous ref, duplicate-resolving prerequisites, self
  dependency, and cycle creation are all covered in one table
  (`dependency_operations_test.go:122`), and
  `TestServiceDependencyMutationRetriesOnlyBeforeDurablePrefix` (:184) pins both
  halves of the retry boundary — retried before any write, surfaced with a typed
  durable-prefix receipt after one. This is the strongest negative coverage in the
  packages surveyed.
- `internal/store/graphmutation.go` + `graphmutation_test.go` — 17 tests covering
  guard-release failure after unlocking, planner panic, planner re-entry and nested
  mutation, broken-snapshot rejection before the planner, whole-snapshot CAS against
  a raw edit, per-file CAS after a durable prefix, prefix-safe planner order, and
  `updated_at` stamped only on semantic change. Three internal self-checks in
  `materializeTaskGraphPlan` (path drift, "did not materialize the planned canonical
  set", "did not clear legacy fields") have no test, but they are post-hoc assertions
  on the store's own materializer — a test would only re-assert the same statement.
- `internal/wire/envelopes_test.go` — the exhaustiveness question answered properly.
  All **53** declared `*Envelope` types are in the validation table, all 53 are in the
  `jsonEnvelopes` registry, and `TestJSONSchema_ValidatesRealOutput` closes the loop
  with a `reflect`-derived guard (:383–397) that fails if a registered envelope has no
  case. This is the pattern M2 should copy.
- `internal/spacestore/fs.go` + `fs_test.go` (random pick) — `classifyRegistryError`
  is a three-arm switch whose `default:` passes an error through unclassified, the
  shape this lens treats as high-yield. It is clean: both `userconfig` sentinels that
  exist (`ErrSpaceIDConflict`, `ErrInvalidRegistry`) are mapped, both are asserted
  end-to-end through `domain.Classify` (`fs_test.go:99`, :107), and the `default:` arm
  is asserted directly — `classifyRegistryError(errors.New("disk unavailable"))` must
  return the same error value (:111–114). A `default:` arm with its own test is rare.
- `internal/core/service_task.go:56` — the `task list --unblocked` fail-closed guard
  named in `docs/ARCHITECTURE.md` has negative tests at both layers:
  `TestService_ListTasks_UnblockedUsesStrictGraphAndFailsClosed`
  (`listtasks_test.go:75`) and `TestTaskListUnblockedFailsClosedOnBrokenGraph`
  (`cli/task_dependency_test.go:357`).
- Test-infrastructure race check: the package-level `testHookBeforeGraphVerify` /
  `testHookBeforeGraphWrite` / `testHookAfterGraphWrite` globals in `store` are mutated
  by tests without synchronization, which would be a data race under the race detector
  if any of those tests ran in parallel. No test in `internal/` calls `t.Parallel()`
  (0 occurrences repo-wide), so the hooks are safe as written. Flagged here as the
  constraint that makes them safe, not as a defect.

## External research

Two Medium+ findings (< 3), so the conditional research step applies. Scoped to the
lens and this stack.

- **Exhaustiveness is lintable in Go, and this repo doesn't lint it.** `exhaustive`
  checks that a switch over an enum-like named type lists every member, and it ships
  as a golangci-lint linter — including an opt-in `//exhaustive:enforce` mode so it can
  be adopted file-by-file rather than repo-wide.
  `.golangci.yml` currently enables `default: standard` plus `depguard` only. Given
  that this repo already makes its *architecture* rules executable via `depguard`
  ("kept honest by a `depguard` rule in `.golangci.yml` rather than by memory", per
  `docs/ARCHITECTURE.md`), applying the same philosophy to its closed *vocabularies* is
  a consistent next step rather than a new idea. Directly supports M2.
  https://pkg.go.dev/github.com/nishanths/exhaustive ·
  https://golangci-lint.run/docs/linters/
- **The sentinel-bounded enum table pattern** — declare a trailing sentinel constant
  and size a lookup array by it, so adding a member without its table entry is a
  *compile* error rather than a silent fallthrough. A heavier alternative to M2's
  `All*()` + table test; noted as the stricter option if the override vocabulary ever
  grows past a handful of members. The same write-up makes the point this lens is
  built on — invalid states should not silently propagate.
  https://medium.com/@fanan.dala/writing-safer-go-part-1-enforcing-exhaustiveness-with-sentinel-bounded-enum-table-pattern-eaadb0f677b0
- **Negative/unhappy-path testing as a first-class target** — the general guidance
  matches the lens's own framing: a guard is only proven by a test that constructs the
  refusing state and asserts the refusal, and unreached error paths become dead cleanup
  code. Relevant caveat for L3: where a path is genuinely unreachable, the recommended
  move is an explicit annotation or removal, not a synthetic test — which is why L3's
  recommendation is a source change.
  https://medium.com/better-practices/negative-testing-for-more-resilient-apis-af904a74d32b ·
  https://github.com/launchdarkly-labs/go-coverage-enforcer

No antipattern from this research is currently violated by the code beyond what M1/M2
already name; the `exhaustive` gap is the one proactive recommendation worth carrying
forward.

## Candidate tasks (human to triage)

- `tskflwctl task new "Refuse unsetting an epic's required description" --epic 21-code-quality-architecture-hardening --tags core,validation --tier 2 --priority high --description "epic set --unset description writes a lint-invalid epic; reject required-field unsets in SetEpicFields and add the missing epic unset-path tests (M1)"`
- `tskflwctl task new "Make the task lifecycle override vocabulary exhaustively tested" --epic 21-code-quality-architecture-hardening --tags core,testing --tier 3 --priority medium --description "Add AllTaskLifecycleOverrides() plus a membership/non-membership table test, mirroring the registry-derived coverage guard in wire/envelopes_test.go (M2)"`
- `tskflwctl task new "Enable the exhaustive linter for core and domain" --epic 21-code-quality-architecture-hardening --tags lint,tooling --tier 4 --priority low --description "Scope golangci-lint's exhaustive linter to internal/core and internal/domain so closed-vocabulary switches are compile-checked, the way depguard already makes the layering rules executable (M2 tightening)"`
- `tskflwctl task new "Tighten core negative-test assertions for epic and lifecycle guards" --epic 21-code-quality-architecture-hardening --tags testing --tier 4 --priority low --description "L1/L2/L3: assert the sentinel per case in TestService_NewEpic, add the revisit_at-not-deferred case, and resolve the unreadable-record guard asymmetry in planDependencyEdges"`

## Related-task observations (propose-only)

**No open task overlaps any finding.** The cross-reference ran twice: over
`task list --json -c slug,status,description` (0 of 63 active tasks matched any
fingerprint) and as a direct grep of `planning/tasks/` for `SetEpicFields`,
`epic set`, `--unset`, `TaskLifecycleOverride`, `validTaskLifecycleOverride`,
`planDependencyEdges`, `revisit_at`, and `exhaustive`. That grep returned 17 files;
every one is `completed`. Nothing was marked `tracked` and no task was annotated.

Two of those completed tasks are worth reading before fixing M1 — not as overlap, but
as precedent:

- `planning/tasks/6fbj87001d81-task-set-follow-ups-sentinels-unknown-keys-canonical-field-table.md`
  — the *task*-side sweep of exactly this surface. Its description reads: *"Guard error
  not sentinel-wrapped, unknown --set keys written silently with no unset, epic
  unclearable, updated_at clobbered, list-field drift."* That is M1's checklist, run
  against `SetFields` and never against `SetEpicFields`. The fix for M1 should re-run
  that task's list on the epic path rather than invent a new one.
- `planning/tasks/6g1dnvcawb9c-close-the-research-test-gaps-that-let-two-data-integrity-bugs-ship.md`
  — the same play on the research path, including *"five tests assert less than their
  names claim"*, which is L2's exact shape. Useful as a template for how this repo has
  previously scoped and landed a test-gap task.

**Observation (propose-only, no action taken):** the pattern across those two tasks and
this audit is that each entity's mutation surface gets its guard sweep independently,
and epics have not had one. Whether that warrants a single "sweep `epic set`/`epic edit`
to task-set parity" task rather than the narrower M1 fix is a scoping call for the human
— this audit proposes only the narrow fix.
