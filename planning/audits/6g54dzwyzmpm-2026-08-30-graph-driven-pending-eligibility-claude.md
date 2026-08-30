---
schema: 1
id: 6g54dzwyzmpm
bucket: closed
area: graph-driven-pending-eligibility-claude
date: "2026-08-30"
updated_at: "2026-08-30"
---
# Audit: Graph-driven pending eligibility — Claude adversarial review — 2026-08-30

> Edit findings through `tskflwctl audit finding` so status and resolution metadata stay queryable.

Independent adversarial review of the uncommitted implementation on
`feat/graph-driven-pending-eligibility` against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) §3 and
[`6g5075cga2nt`](../tasks/6g5075cga2nt-make-dependency-eligibility-graph-driven-for-queued-and-ready-tasks.md).
The whole slice is in the working tree (52 files, +267/−122); `main..HEAD` is empty, so the review
target is `git diff HEAD` — there are no untracked files. Only two production files carry logic
changes: `internal/core/dependency_graph.go` (+10/−1) and `internal/core/task_lifecycle.go`
(+31/−8). The rest is wire descriptions, one schema bump, help text, generated docs, goldens, and
tests.

Coverage: the derived-state algebra (role · gate · eligible · drained · inconsistent) across every
persisted status and every gate class; enumeration of every first-party route into `in-progress`;
`--force` scope against role, gate, and repository health; read/mutate agreement across
`task list --unblocked`, `task blockers`, Thread frontier, and the guarded write; wire/schema
contract accuracy; test rigour measured by mutation; and extension hazards for the Thread and
dependency work still to land.

**Verdict: ready with tracked follow-ups.** The core correction is right and, unusually, is *more*
correct than the task it implements: the status/gate matrix exposed a pre-existing `--force` hole
(any non-clear gate was overrideable, including a locally broken one) and this branch closes it.
Authorization has exactly one owner, every product path funnels through it, and I found no bypass.
Two of the six findings are one-line string fixes inside the delta itself and should land in this
task (**M1**, **L4**); the other four are follow-up work. The larger concern to carry forward is
**M2** — `isPendingWorkRole` centralises the *pending* question but the pending set is still
enumerated in three uncoordinated places, and `roleForStatus`'s `default:` arm will silently
swallow the next status anyone adds without failing a single test.

## Commands run

| Command | Outcome |
| --- | --- |
| `just build` | ok (`bin/tskflwctl`, `v0.17.1-63-g32fd93a-dirty`) |
| `go test ./...` | all packages green |
| `go test -race ./internal/{core,cli,tui,store,wire,domain}/...` | green |
| `golangci-lint run ./...` | 0 issues |
| `go vet ./...` | clean |
| `gofmt -l ./cmd ./internal` | empty |
| `just docs-check` | fails **only** because the branch is uncommitted — the recipe is `git diff --exit-code docs/cli`. Re-running `just docs` reproduced the working tree byte-for-byte (diffstat still 52 files, +267/−122). |
| `./bin/tskflwctl lint` (this repo) | `✔ all planning entities and dependency links pass lint` |

Behavioural repro used a scratch planning repository driven through the real binary: an eight-task
graph crossing `{next-up, ready-to-start}` × `{clear, blocked, broken}`, plus `deferred`,
`completed`, `deprecated`, and `in-progress` sources; forced and unforced starts; `task move` as the
generic escape hatch; a Thread projected under healthy, missing-member, and duplicate-member
membership; and a degraded (legacy-field) repository.

**Mutation probe.** To measure test rigour rather than trust a green run, `isPendingWorkRole` was
temporarily reverted to `role == RoleCandidate`. Nine tests failed across three packages —
`TestTaskGraphPendingRolesShareGraphDrivenEligibility`,
`TestService_ListTasks_UnblockedUsesStrictGraphAndFailsClosed`,
`TestValidateTaskLifecycleStartEligibilityAndTypedOverride`,
`TestValidateTaskLifecycleEveryPersistedStatusEnteringInProgress`,
`TestProjectThreadFrontierUsesSharedEligibility` (core);
`TestTaskGraphQueryCommandsAndUnblockedSelector`, `TestThreadNewListShowPathAndFrontier`,
`TestTaskStartAcceptsQueuedWorkWithClearOrForcedBlockedGate` (cli); and
`TestTUITaskStartUsesDependencyEligibilityPolicy` (tui). The behaviour is pinned end to end, not
merely at the unit layer. The working tree was then restored and re-verified identical to the
branch state (`git diff --stat HEAD` unchanged; suite re-run green). Nothing under review was
modified.

## Findings

### Medium

#### M1. The published `frontier` description names the wrong precondition; a Thread-local membership defect empties the frontier while the same payload reports eligible members and a healthy graph  · **Status:** fixed

**File:** `internal/wire/thread.go:82`; `internal/core/thread_projection.go:177-183`; `internal/cli/thread.go:190` | **Component:** Thread read projection / machine contract
**Effort:** XS · **Urgency:** acute

**Class:** misleading contract published by this delta. · **Disposition:** fix the description in this task; track the regression test.

This branch publishes a new reflected description on the field it changed:

```go
Frontier []ThreadTaskJSON `json:"frontier" jsonschema:"description=next-up and ready-to-start Thread members with clear dependency gates in a healthy graph"`
```

The actual predicate is `view.ProjectionHealth == GraphHealthy` (`thread_projection.go:177`), which
is *strictly stronger* than graph health. A `ValidateThreadDocument` failure or any missing member
sets `ProjectionHealth = GraphBroken` (`thread_projection.go:104,124`) while `graph_health` stays
`healthy`. The envelope itself carries both fields separately, so "in a healthy graph" reads as
`graph_health` — and is then false.

Reproduced with the real binary. Thread members `{ready (ready-to-start, clear), dep-ready,
dr-broken}` plus one hand-added missing id:

```
projection_health broken   graph_health healthy
frontier: []
members : [('ready','candidate','clear',True), ('dep-ready','in-flight','blocked',False),
           ('dr-broken','candidate','broken',False), ('6g54zzzzzzzz','unknown','broken',False)]
problems: ['missing-thread-member']
```

Same result for a duplicate member id (`problems: duplicate member`). In both cases the *same
payload* reports `members[0].state.eligible: true`, `task list --unblocked` still lists `ready`, and
`task start ready` succeeds. So one envelope contradicts itself, and `frontier` contradicts both the
sibling read projection and the mutation.

Why it matters beyond wording: Thread membership mutations have not landed (ADR-0006 §10; `thread
new --task` is the only supported writer, and `lint --fix` deliberately leaves Thread membership
alone per CLAUDE.md), so membership is hand-edited today — and the two likeliest hand-edit mistakes
each zero the frontier for the *entire* Thread at exit code 0. An agent polling `thread frontier`
reads "no work" for a Thread whose members are individually startable.

The fail-closed behaviour itself is correct and the human surface is honest —
`graph: healthy · projection broken · 0 eligible member(s)` followed by
`⚠ missing-thread-member: …`. Only the machine description overstates.

**Why existing tests did not catch it:** `TestProjectThreadFrontierUsesSharedEligibility` exercises
only a healthy projection. `TestProjectThreadCompletedInconsistencyAndMissingMember` asserts
problems and rollup, not the frontier-vs-`members[].eligible` divergence. Nothing asserts that a
published `jsonschema` description matches the predicate it describes.

**Recommendation:** amend the description to name projection health, e.g.
`description=next-up and ready-to-start Thread members with clear dependency gates; empty unless
both graph_health and projection_health are healthy`. Separately (follow-up), add one projection
test pinning "healthy graph + broken projection ⇒ frontier empty while an eligible member is
present", so the divergence is deliberate and covered.

**Resolution:** The schema now names both graph_health and projection_health,
and a regression test pins that broken Thread-local evidence suppresses an
otherwise graph-eligible member from the frontier. No exit-code change was made:
this read command returns the diagnosable projection, whose machine envelope
explicitly carries projection_health and problems, while frontier itself still
fails closed.

#### M2. The pending-work set is enumerated in three uncoordinated places; `isPendingWorkRole` alone does not prevent the drift it was introduced to prevent  · **Status:** fixed

**File:** `internal/core/dependency_graph.go:107-113`, `:685-701`, `:1014-1022` | **Component:** derived-state algebra / extension surface
**Effort:** S · **Urgency:** soon

**Class:** extension hazard. · **Disposition:** tracked follow-up; does not block this task.

The review brief asks whether `isPendingWorkRole` is sufficient to prevent semantic drift. It
centralises the *pending* question well, but the pending set is still spelled out three times, and
only one of them goes through the helper:

1. `isPendingWorkRole(role)` — `dependency_graph.go:111`, role-based, the new helper.
2. `roleForStatus(status)` — `dependency_graph.go:685-701`, status→role, with `default: RoleUnknown`.
3. `blockerReason` — `dependency_graph.go:1014-1016`,
   `case domain.StatusNextUp, domain.StatusReadyToStart: return BlockerNotStarted`, status-based and
   entirely independent of (1) and (2).

Adding a status to `domain.allStatuses` — which review angle 6 names explicitly ("additional pending
statuses") — compiles clean and silently produces:

- `roleForStatus` → `RoleUnknown` via the `default:` arm, with no compile error and no diagnostic;
- `isPendingWorkRole` → false, so the status is never eligible, never in a Thread frontier, and
  never in `task list --unblocked`;
- `Inconsistent` and `Drained` → both false, so the status sits outside every named view in
  ADR-0006 §3;
- `computeSound` → neither sound nor broken, so every downstream task's gate becomes `blocked` with
  reason `invalid-status` (the `default:` at `dependency_graph.go:1025`) — i.e. the forensic
  vocabulary the ADR explicitly reserves for hand-edit damage and old binaries gets applied to a
  legitimate new lifecycle state.

**Why existing tests did not catch it:** `TestValidateTaskLifecycleEveryPersistedStatusEnteringInProgress`
does iterate `domain.AllStatuses()`, but its `default:` arm asserts that the status *refuses* to
start (`task_lifecycle_test.go:110-115`). A silently-`RoleUnknown` new status satisfies that
assertion, so the suite goes green on a wrong classification. There is no test anywhere that asserts
`roleForStatus` is total over `domain.AllStatuses()`.

**Recommendation (narrow, two parts):** (a) make `blockerReason`'s pending case delegate —
`if isPendingWorkRole(roleForStatus(task.Status)) { return BlockerNotStarted }` — so the pending set
has one enumeration; (b) add an exhaustiveness test asserting `roleForStatus(s) != RoleUnknown` for
every `domain.AllStatuses()`, which turns "someone added a status" from a silent misclassification
into a failing test naming the file to edit.

**Resolution:** blockerReason now delegates pending-role classification through
isPendingWorkRole(roleForStatus(...)), and a totality test requires every domain
status to map to a known lifecycle role.

### Low

#### L3. Create-and-start still makes `ready-to-start` authoritative — the one route where readiness remains a precondition  · **Status:** wontfix

**File:** `internal/core/task_lifecycle.go:259-261` and the type comment at `:34-36`; `internal/core/service_task.go:496-499` | **Component:** guarded create-and-start
**Effort:** XS · **Urgency:** eventually

**Class:** residual status-as-authorization. · **Disposition:** tracked follow-up; not user-reachable today.

The contract under review says readiness "is not an authorization prerequisite for starting work",
and every other path now accepts either pending role. `validateCreateAndStartPlan` still hard-requires
one:

```go
if task.Status != domain.StatusReadyToStart {
    return …, fmt.Errorf("%w: create-and-start candidate must begin as ready-to-start, got %s", …)
}
```

`Service.NewTask` sets `next-up` when `p.Next` (`service_task.go:496-499`) and then hands the
document to that validator when `p.Start` is set, so `NewTaskParams{Next: true, Start: true}` fails
deep in the lifecycle validator with a message that contradicts this branch's own contract. The
`TaskLifecycleCreation` doc comment ("Task must be a ready-to-start candidate") says the same.

Not user-reachable: `internal/cli/task.go:181` marks `--next` and `--start` mutually exclusive, and
the TUI has no create-and-start. So this is a core-API and vocabulary defect, not a shipped bug —
but it is exactly the "status metadata authoritative again" residue the review asks about, and it
becomes reachable the moment another adapter (or a bulk-creation path) constructs
`NewTaskParams` directly.

**Why existing tests did not catch it:** no test crosses `Next` with `Start`; the CLI's mutual
exclusion means no golden or integration test can reach the combination.

**Recommendation:** either accept `isPendingWorkRole(roleForStatus(task.Status))` there — a
brand-new task has no dependency edges by construction (`task_lifecycle.go:262-264` forbids them),
so both pending seeds are equally safe — or keep the restriction and reword comment and error to say
it is an internal seed-document invariant, not a lifecycle rule.

**Resolution:** Create-and-start's ready-to-start value is a canonical ephemeral
seed-document shape, never a persisted readiness prerequisite; the CLI
intentionally rejects the contradictory --next/--start combination. Comments and
validation text now make that boundary explicit.

#### L4. `DependencyGateOverrideAllowed`'s doc comment justifies the new rule with a false premise, inviting a revert of the fix it documents  · **Status:** fixed

**File:** `internal/core/task_lifecycle.go:153-159` | **Component:** override policy documentation
**Effort:** XS · **Urgency:** acute

**Class:** misleading rationale on new code. · **Disposition:** fix in this task (one line).

The new helper carries:

> Broken repositories fail before an eligibility error is constructed, so only a blocked
> pending-work role is overrideable.

That premise is about *repository* health, which `ValidateTaskLifecycleSource` already enforces at
`task_lifecycle.go:187-196`. It is not why `GateBroken` is excluded. A gate is `broken` in a
perfectly healthy repository whenever a prerequisite is `deprecated` — `computeSound` returns
`soundResult{broken: true}` for `task.Status == domain.StatusDeprecated`
(`dependency_graph.go:661`), and every `hardBroken` assignment is paired with an `addProblem` call
that would have made the repository unhealthy, so a withdrawn prerequisite is the one remaining way
to reach `GateBroken` under a healthy graph. That is precisely the case this branch fixes, and the
task's own progress note describes it correctly.

Verified: `task start dq-broken --force` in a **healthy** repository →
`role=queued gate=broken`, `override_allowed: false`, blocker reason `withdrawn`,
remedy `repair the broken dependency evidence, then inspect the task before retrying`.

As written, the comment implies the `before.Gate != GateBlocked` term at `:317` is redundant
defence against a case that cannot arise — which is exactly the reasoning a future simplification
pass would use to restore `gate != GateClear` and re-open the hole this task closed.

**Why existing tests did not catch it:** comments are not executable. The behaviour *is* covered
(`TestValidateTaskLifecycleStartEligibilityAndTypedOverride`'s `withdrawn` case now asserts the
forced path still refuses), which is why the wrong rationale survives a green suite.

**Recommendation:** replace the second clause with the real reason — a `broken` gate under a healthy
repository means a withdrawn or otherwise unresolvable prerequisite, which `--force` cannot make
sound; only `blocked` (an unfinished but well-formed prerequisite) is a scheduling decision the
operator is entitled to override.

**Resolution:** The helper comment now explains the reachable healthy-repository
case: a withdrawn prerequisite yields a broken gate that force cannot make
sound.

#### L5. `override_allowed` narrowed its meaning in 1.55 but is the one changed field with no reflected description  · **Status:** fixed

**File:** `internal/wire/dependency.go:259` (field) and `:269` (derivation) | **Component:** machine contract
**Effort:** XS · **Urgency:** eventually

**Class:** contract-surface inconsistency. · **Disposition:** tracked follow-up.

`OverrideAllowed` went from `role == core.RoleCandidate && gate != core.GateClear` to
`failure.DependencyGateOverrideAllowed()` — a genuine semantic narrowing (a broken gate used to
report `true`, and `--force` used to work on it). The delta's own pattern for publishing semantics
is a reflected `jsonschema` description; it added one to `TaskGraphStateJSON.Eligible`
(`dependency.go:28`) and one to `ThreadViewJSON.Frontier` (`thread.go:82`), and both landed in
`schema_jsonschema.golden`. `override_allowed` got neither — its narrowing lives only in the
`wire.go` prose changelog at `:182-185`.

This is the field a machine consumer actually branches on to decide whether to retry with `--force`,
so it is the one most worth self-describing.

**Why existing tests did not catch it:** the schema-comment drift test checks that documented fields
stay in sync, not that a semantically changed field acquires documentation.

**Recommendation:** add
`jsonschema:"description=true only when the refusal is a blocked dependency gate on next-up or ready-to-start work; --force never bypasses a broken gate or another lifecycle role"`
and regenerate the goldens.

**Resolution:** override_allowed now has a reflected JSON Schema description
spelling out the pending-role, blocked-gate-only force contract; goldens were
regenerated.

#### L6. `thread list` still labels the frontier count "ready" — the exact word this task removes from the authorization vocabulary  · **Status:** fixed

**File:** `internal/cli/render/thread.go:41` | **Component:** human render
**Effort:** XS · **Urgency:** eventually

**Class:** incomplete propagation of the rename. · **Disposition:** tracked follow-up.

Under the `FRONTIER` column header (`thread.go:49`), the cell is built as
`fmt.Sprintf("%d ready", len(view.Frontier))`. That count now includes `next-up` members, so
"3 ready" no longer means "3 ready-to-start" — in a tool where `ready-to-start` is a status *and*
`task ready` is a verb, this is the specific overload ADR-0006 §3 is trying to retire. Every other
surface in the delta was reworded ("graph-clear pending member tasks", "clear-gated next-up/ready
members", "only next-up or ready-to-start tasks with a clear dependency gate"); this one was missed.

**Why existing tests did not catch it:** `thread list` human output is asserted on structure, and
the phrase is not in any golden.

**Recommendation:** `fmt.Sprintf("%d eligible", len(view.Frontier))`, matching
`ThreadFrontierHuman`'s own `%d eligible member(s)` at `thread.go:114`.

**Resolution:** Thread list now labels the frontier count as eligible, with a
human-render regression assertion preventing readiness terminology from
returning.

## What held up well

1. **`--force` is genuinely narrow, and provably so.** All four refusal classes verified against the
   real binary: `parked`, `nominally-complete`, `withdrawn`, and non-pending roles refuse with
   `move it to next-up or ready-to-start before starting`; a broken gate refuses with
   `override_allowed: false` and a distinct remedy; a degraded repository refuses in
   `ValidateTaskLifecycleSource` before any eligibility error exists. `task move x next-up --force`
   is rejected outright rather than silently accepting a meaningless override.
2. **The adapter-side reconstruction is gone.** `wire`'s old candidate-only check now calls
   `failure.DependencyGateOverrideAllowed()`. A full grep for `core.Role*` / `core.Gate*` across
   `cli`, `tui`, and `wire` leaves only two `GateClear` comparisons, both used for "newly unsafe"
   impact highlighting in renderers. Authorization has exactly one owner.
3. **No bypass route found.** Every first-party path into `in-progress` funnels through
   `MutateTaskLifecycle`: `task start`, `task move`, `task new --start`, TUI `moveTask` (explicitly
   `TaskLifecycleOverrideNone`, `tui/entity.go:240`), and batch `runMoves`, which re-enters the
   guarded mutation per slug so a batch re-derives the graph after each write. `task set status` is
   rejected in core (`service_task.go:328`) *and* again in the store (`fsstore.go:135`);
   `task edit` refuses a status change by comparing against the pre-edit value
   (`store/edit.go:175-178`); `store.CreateTask` accepts only `ready-to-start`/`next-up`
   (`create.go:146`). `lint --fix` never writes `status`.
4. **Read and mutate agree on the healthy path.** Across all eight probe tasks,
   `task list --unblocked`, `task blockers`'s `state`, the Thread frontier, and `task start` gave
   identical verdicts. The degraded-repository case fails all three closed with attributable
   diagnoses — and `task blockers` carries `health: degraded` in the payload rather than leaving a
   consumer to infer eligibility from an empty blocker list.
5. **The forced-start round trip is honest.** Receipt: `from: next-up`, `forced: true`,
   `before {queued/blocked}` → `after {in-flight/blocked, inconsistent: true}`; `lint` then reports
   `persisted in-flight task has a blocked dependency gate: … (not-started via …)` with a
   deterministic shortest path. ADR-0006's "remains derived as inconsistent until its prerequisites
   become soundly completed" is observably true rather than merely asserted.
6. **Interactive surfaces needed no change and did not half-land.** Completion
   (`completion.go:220-228`) and the bare-verb picker (`fill.go:163-167`) exclude only the
   destination status, so `task start` already offered `next-up` work. This is a place the change
   could plausibly have been left inconsistent.
7. **Version discipline.** `SchemaVersion` bumped to 1.55 with a changelog entry naming the
   behavioural change, both reflected descriptions regenerated into `schema_jsonschema.golden`, all
   thirty goldens re-stamped, and `just docs` regeneration is deterministic and matches the working
   tree byte-for-byte.
8. **The scope discipline is right.** The one production behaviour change beyond the stated task —
   narrowing `--force` from "any non-clear gate" to "blocked only" — is a safety fix the status/gate
   matrix surfaced, is recorded in the task's progress note, and belongs here rather than in a
   follow-up, because leaving it would have widened the hole to a second source role.

## Candidate tasks

- Fold **M1** and **L4** into `6g5075cga2nt` before merge (two one-line strings inside the delta).
- One follow-up task for **M2**: single-source the pending-work set through `isPendingWorkRole` in
  `blockerReason`, and add a `roleForStatus` totality test over `domain.AllStatuses()`. This should
  precede any work that adds a lifecycle status.
- One follow-up task bundling **L3**, **L5**, **L6**: align create-and-start's seed-status rule with
  the pending-work contract, document `override_allowed` as a reflected description, and reword the
  `thread list` frontier column.
- Follow-up (sequence with guarded Thread membership mutations): a projection test pinning
  "healthy graph + broken projection ⇒ empty frontier alongside an eligible member", and a decision
  on whether `thread frontier` should exit non-zero when a Thread-local defect suppresses the
  frontier, so an agent dispatch loop cannot read a repairable defect as "no work".
