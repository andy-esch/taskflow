---
schema: 1
id: 6g5axb85endz
bucket: open
area: bulk-thread-apply-implementation-claude
date: "2026-08-30"
updated_at: "2026-08-31"
---
# Audit: Bulk-link existing tasks into Threads with resumable apply — Claude — 2026-08-30

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
> should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for task
[`6g3q4rtv8d0a`](../tasks/6g3q4rtv8d0a-bulk-link-existing-tasks-into-threads-with-resumable-apply.md)
on branch `feat/bulk-link-existing-tasks`, against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially its bulk-composition,
compile/apply, mutation safety, implementation-consequence, and 2026-08-30 V1-contract sections.

Assume the implementation may be subtly wrong even though its local suite is green. Look for
semantic contradictions, incomplete authorization, stale-snapshot claims, misleading receipts,
non-resumable prefixes, identity confusion, compatibility breaks, and seams that will make later
Thread graph views or TUI work unsafe. Do not reward complexity or test volume by itself. Equally,
do not manufacture findings: settle a concern when code plus an adversarial reproduction disproves
it.

## Review target

The branch is based at `67a68bd`. The implementation is split across tracked and untracked files,
so inspect `git status --short`, `git diff HEAD`, and all relevant untracked files. Primary new
production/test files are:

- `internal/core/thread_apply.go` and `internal/core/thread_apply_test.go`;
- `internal/core/service_thread_apply.go` and `internal/core/service_thread_apply_test.go`;
- `internal/store/threadapply.go`, `internal/store/threadapply_test.go`, and
  `internal/store/threadapply_benchmark_test.go`;
- `internal/cli/thread_apply.go`, `internal/cli/thread_apply_test.go`, and the apply renderers;
- `internal/wire/thread.go` plus the schema/envelope changes; and
- generated `docs/cli/tskflwctl_thread_compose.md` and `tskflwctl_thread_apply.md`.

Also inspect the supporting changes in `internal/core/store.go`, `internal/core/service.go`,
`internal/store/fsstore.go`, `internal/cli/root.go`, `internal/cli/exit.go`, README, architecture,
ADR-0006, the spike record, the implementation task, and the schema/golden artifacts. Ignore
unrelated concurrent work under `planning/meta/`, `routines/`, and
`planning/tasks/6g54ay3njm8y-adopt-branch-protection-ruleset-as-code-.github-ruleset.json.md`.
The two new review-audit documents are review scaffolding, not implementation evidence.

The intended contract is:

- `thread compose` consumes one strict literal YAML/JSON manifest of exact existing task IDs and
  local keys, validates the proposed global edge union, and emits one preallocated unstarted Thread
  plus additive dependency intent without mutating planning entities.
- `member` defaults true. A `member: false` node is permitted only when it is actually a transitive
  prerequisite of a member in the proposed final graph. Inline task creation is outside V1.
- Compose refuses an ID-less planning repository and creates a mode-`0600`, no-clobber durable plan.
  Compose may read its authoring manifest from stdin; apply requires a durable plan path.
- Apply treats dependencies and Thread creation as additive intent. Existing edges and an exactly
  matching Thread are skips; unrelated task edits and dependencies survive; same-ID/different-
  content is a conflict.
- One compound mutation capability reloads live planning identity, the strict task graph, every
  Thread and body under the canonical-root guard; dependency-owner files land first in a validated
  prefix-safe order and the Thread lands last.
- Every interruption is resumable from the same materialized plan. Receipts distinguish pending,
  applied, and skipped operations, durable commit from semantic completion, and pre-commit from
  post-commit failure. Automatic retry is limited to incomplete pre-commit conflicts.
- Cooperating writers serialize on the shared repository guard. Raw editors remain outside it and
  are detected by whole-source, immediate-target, and final-convergence re-reads only where the
  filesystem design can observe them; the implementation does not claim a multi-file transaction.
- At 1,000 tasks and 300 dependency-owner writes, the reference budget is at most one second for
  guarded read/validate/materialize and at most five seconds total on a development-class machine.

## Required hostile angles

1. Attack both file formats. Probe unknown and duplicate fields, duplicate YAML keys, aliases,
   multi-document input, null/empty collections, schema zero versus one, malformed dates/statuses,
   repeated nodes/tasks/edges, whitespace variants, self-edges, missing tasks, and local-key
   confusion. Confirm human manifests cannot smuggle shell interpolation or partial references and
   hand-edited plans cannot escape the Thread directory through ID/slug/body fields.
2. Re-derive `member: false` semantics against both existing and newly proposed transitive paths.
   Try disconnected non-members, non-member chains, multiple members, existing external edges,
   cycles, and a gate that becomes upstream only through several proposed edges. Look for a role
   claim compose accepts but runtime Thread projection would describe differently.
3. Trace compose end to end: config identity, strict repository reads, cross-kind ID uniqueness,
   ordinary Thread-template rendering, deterministic normalization, no entity mutation, mode 0600,
   no-clobber output, file/directory sync behavior, stdin, dry-run, and JSON. Challenge whether a
   failed or interrupted output write can leave a recovery token that appears valid.
4. Trace apply authorization from Cobra through Service into the control-inverted store boundary.
   Confirm live configuration rediscovery occurs inside the guarded operation and correctly handles
   ID-less repos, wrong repo IDs, moved checkouts, changed pointers, symlinks/canonical roots, and
   programmatic stores without an identity reader. Look for validation performed only before lock
   acquisition or Store re-entry from the planner callback.
5. Audit additive intent carefully. Existing edge-only plans must not rewrite/canonicalize a task;
   needed additions must preserve unrelated frontmatter, body, permissions, dependencies, and raw
   edits between retries. Challenge sorting/deduplication, several edges owned by one task, timestamp
   stamping, exact Thread/body equivalence, cross-kind collisions, and same-ID differences in every
   lifecycle or metadata field.
6. Audit the compound transaction in exact order: guard, identity, strict graph/Thread snapshot,
   pure plan, independent store revalidation, materialization, dry-run/no-op, whole-source CAS,
   per-target CAS, atomic task replacements, final convergence re-read, exclusive Thread create,
   and unlock. Prove every physical prefix remains graph-valid and the Thread never advertises an
   edge that was not durable first.
7. Attack interruption and recovery after every physical write, including multiple logical edges in
   one task-file replacement, raw removal of an edge that just landed, a pre-existing exact Thread
   whose dependencies need repair, concurrent exact/different Thread creation, final-create error,
   and guard-release error. Verify operation states, `changed`, `committed`, and `complete` never lie,
   and that retry guidance is safe even when final convergence cannot be classified precisely.
8. Stress cooperating-writer claims with independent stores/processes where practical: bulk apply
   versus direct dependency mutation, task lifecycle, Thread membership/lifecycle, and another bulk
   apply. Distinguish waiting plus fresh re-authorization from mere mutual exclusion. Separately
   probe raw-editor windows without attributing transactional guarantees the design disclaims.
9. Inspect human and machine contracts on success, no-op, dry-run, validation failure, conflict,
   interrupted prefix, and post-commit failure. Check stable IDs, operation action/state vocabulary,
   non-nil arrays, plan path/workspace, error exit classification, recovery detail when failure
   occurs before the store accepts the plan, and schema 1.57/golden accuracy.
10. Challenge complexity and performance. Re-run the representative benchmark, inspect the
    O(W × (V+E)) prefix validator and lock-held fsync cost, and assess whether the stated one-/five-
    second budgets are reproducible and appropriately scoped. Determine whether any graph-library
    need is now concrete rather than hypothetical, and whether contention makes the V1 design
    operationally unsafe before the nominal benchmark limit.
11. Assess test quality rather than only coverage. Use targeted mutation probes or temporary local
    reversions where useful and restore them afterward. Look for global-hook tests that are unsafe
    under parallel/race execution, overfit assertions, tests that would pass if a guard/CAS/final
    re-read were removed, and generated contracts refreshed without behavioral proof.
12. Compare README, architecture, generated CLI docs, ADR, spike lessons, implementation task, wire
    comments, and actual behavior. Flag any obsolete inline-`new_task` promise, any stronger raw-
    editor guarantee than the code provides, or any mismatch in dry-run/completion semantics.
    Confirm the task is correctly dogfooded in `complete-production-threads` without this review
    scaffolding or bulk implementation mutating unrelated Thread members.

Run proportionate validation, including full tests, focused race tests, vet/static analysis,
planning lint, schema/golden checks, `git diff --check`, and the representative benchmark. Record
exact commands and outcomes. If regenerating artifacts as a check, leave the worktree exactly as
you found it except for this assigned audit file.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict (`ready`, `ready with tracked follow-ups`, or `not ready`);
- the reviewed commit/worktree state and exact commands run;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete recommendation;
- a short acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

Do not modify implementation files, the ADR, task, Thread, generated artifacts, or the other
reviewer's audit. Do not create follow-up tasks or pre-resolve findings; the implementation owner
will triage them after both independent reviews arrive.

## Executive verdict

**Ready with tracked follow-ups.**

The compound apply is sound where it matters most. Every hostile probe I aimed at the durable
contract — path escape through a hand-edited plan, format smuggling, additive-intent preservation,
prefix graph-validity, identity fail-closure, cooperating-writer serialization, byte-identical
idempotent retry — came back clean, and the two performance budgets reproduce on this machine. The
write ordering is right: dependency owners land first in a prefix-valid order, the Thread lands
last, and no receipt I could produce claimed a durability that had not happened.

Three findings deserve triage before this is called done. One is a genuine semantic contradiction
(`member: false` is validated against a transitive relation that the runtime Thread projection can
never display). One is a coverage hole that a mutation probe proved (the all-or-nothing whole-source
CAS can be deleted with the entire suite still green, while its pre-existing sibling in
`MutateTaskGraph` is properly pinned). One is a quantified complexity finding (a provably redundant
per-prefix revalidation consumes 59–64% of the phase that owns the one-second budget, and the
recorded performance note measures it against the wrong denominator). None of the three corrupts a
repository; all three weaken a contract the ADR states explicitly.

## Reviewed state and commands

Worktree: branch `feat/bulk-link-existing-tasks`, based at `67a68bd`, uncommitted
(`49 files changed, 749 insertions(+), 82 deletions(-)` plus the untracked implementation files).
Machine: Apple M5, darwin 25.6.0, Go toolchain from the repo. The worktree was restored to exactly
this state after every probe (verified by diffstat and per-file SHA-1).

| Command | Outcome |
| --- | --- |
| `just build` | ok |
| `go build ./...` · `go vet ./...` | clean |
| `go test ./...` | all packages pass |
| `go test -race -count=1 ./internal/store/... ./internal/core/... ./internal/cli/...` | pass, no race reports |
| `just lint` (golangci-lint) | 0 issues |
| `./bin/tskflwctl lint` (planning) | all entities and dependency links pass |
| `git diff --check` | clean |
| `just docs` then `diff -rq` against a pre-run copy | no drift; generated CLI reference is current |
| `go test ./internal/store/ -bench BenchmarkThreadApplyRepresentative -benchtime 3x` | dry-run 522 ms/op, apply 3.499 s/op |

Live end-to-end probing used six throwaway planning repositories built with the freshly built
binary (`init` → `epic new` → `task new` → `thread compose` → `thread apply`), covering manifest
attacks, hand-edited plan attacks, preservation, repair, ID-less repositories, wrong-repository
plans, and a two-process concurrent apply of one 40-node/39-edge plan.

Mutation probes (each applied, measured, then reverted and checksum-verified):

| Probe | Target | Result |
| --- | --- | --- |
| 1 | disable whole-source CAS, `store/threadapply.go:123` | **suite still green** → M2 |
| 2 | disable final-convergence pending-dependency check, `:160` | `TestThreadApplyReverifiesRepairedEdgesWhenThreadAlreadyExists` fails (good) |
| 3 | skip final re-prepare when the Thread already exists, `:151` | same test fails (good) |
| 4 | remove per-target CAS, `:131` | two tests fail (good) |
| 5 | disable `MutateTaskGraph` whole-source CAS, `store/graphmutation.go:87` | `TestMutateTaskGraphCASRejectsRawEditBeforeApply` fails (good — the contrast that makes M2 a gap) |
| 6 | disable per-prefix revalidation, `core/dependency_graph_mutation.go:85` | suite green; timing → M3 |

## Findings

### Medium

#### M1. `member: false` is validated transitively, but a Thread can only ever display a *direct* external gate  · **Status:** open

**File:** `internal/core/thread_apply.go:310` (and `:489-497`) vs `internal/core/thread_projection.go:157-160` | **Component:** core/thread-apply · core/thread-projection
**Effort:** S · **Urgency:** soon

Compose accepts a non-member node when it transitively reaches any member: `thread_apply.go:310`
calls `taskReachesAny`, which at `:489-497` walks `graph.Downstream(start)` — the **transitive**
dependent closure. The runtime projection derives external gates at `thread_projection.go:157-160`
from `graph.Prerequisites(taskID)`, which at `dependency_graph.go:866-868` returns
`g.dependencies[taskID]` — **direct** prerequisites only. The two definitions disagree, and the ADR
now disagrees with itself: §114 says "A **direct** prerequisite outside $M$ remains an **external
gate**" and §372 says "`member: false` permits an explicitly declared existing external gate", while
the 2026-08-30 amendment says `member: false` is "a claim that the node is a real **transitive**
upstream gate of a member".

**Reproduction** (live `thread compose` → `thread apply` → `thread show`): three tasks, manifest
nodes `a{member:false} → b{member:false} → c{member:true}` with edges `a→b`, `b→c`. Compose accepted
the manifest and apply landed both edges. `thread show` then reported members `task-c` and external
gates `task-b` — and `task-a`, the node compose validated as "an upstream gate of a Thread member",
appeared in `members`, `external_gates` and `frontier` as nothing at all. It is invisible in both the
human and `--json` projections.

The only user-visible purpose of a `member: false` node is to declare an external gate. Compose
confirms the role, mutates the repository to create the edge, and the Thread then never mentions it.
An author reasonably reads compose's acceptance as "this gate is now tracked by the Thread"; it is
not. This is also precisely the seam later Thread graph views and the TUI will build on, so the
ambiguity compounds rather than staying local.

**Recommendation:** Choose one definition and make compose, `ProjectThread` and the ADR use one
word. Tightening compose to require a *direct* prerequisite of a member in the proposed final graph
is the smaller change and matches ADR §114 plus today's projection; widening `ProjectThread` to
transitive gates is defensible but changes the meaning of every existing Thread view. Either way,
retire the contradictory sentence in the 2026-08-30 amendment.

#### M2. The all-or-nothing whole-source CAS is untested — the full suite passes with it deleted  · **Status:** fixed

**File:** `internal/store/threadapply.go:123-125` (tests: `internal/store/threadapply_test.go:277`, `:365`) | **Component:** store/thread-apply
**Effort:** XS · **Urgency:** soon

`threadapply.go:123-125` is the pre-write preflight that refuses when the repository changed under
the plan. Rewriting its condition as `if false && (...)` leaves
`go test ./internal/store/ ./internal/core/ ./internal/cli/` **entirely green**. The identical probe
against the pre-existing sibling at `internal/store/graphmutation.go:87` fails
`TestMutateTaskGraphCASRejectsRawEditBeforeApply` — so the older path pins this guard and the new
compound path does not. Probes against the final convergence re-read (`:151`, `:160`) and the
per-target CAS (`:131`) each break the suite, which makes this the single uncovered guard in the new
operation.

The cause is visible in the test file: both uses of `testHookBeforeThreadApplyVerify`
(`threadapply_test.go:277` and `:365`) inject an identity change and a guard-hold. Neither performs a
raw content edit at that seam — the exact scenario the guard exists for. `MutateTaskGraph`'s hook
comment describes that job explicitly ("interleaves a non-cooperating raw edit before the all-file
CAS"); the analogous apply test was not written.

The per-target CAS only covers files the plan writes. A raw edit to a task the plan does *not*
target — one that introduces a cycle, or removes a prerequisite — is invisible to it. With the
whole-source check gone, apply proceeds on a stale snapshot, lands a durable dependency prefix, and
only then fails at the final convergence re-read. That converts a clean pre-commit refusal
(`committed: false`) into a committed partial write. `docs/ARCHITECTURE.md:246-248` states this
guarantee as documentation ("A raw edit before the first write fails the whole-source CAS"), which
makes the missing regression more load-bearing, not less.

**Recommendation:** Add the `MutateTaskGraph` analogue — a `testHookBeforeThreadApplyVerify` that
raw-edits a task the plan does **not** write, asserting `errors.Is(err, domain.ErrConflict)` and
`result.Committed == false`. A second case covering the Threads half of the condition
(`sameThreadSourceSnapshot`) would close it fully; that half is currently unexercised in either
direction.

**Resolution:** Added raw-edit regressions for both halves of the whole-source
CAS: an unrelated task edit and a concurrent Thread addition now prove apply
refuses before its first durable write.

#### M3. The redundant per-prefix revalidation dominates the one-second budget, and the recorded note measures it against the wrong denominator  · **Status:** fixed

**File:** `internal/core/dependency_graph_mutation.go:85-90` (caller: `internal/core/thread_apply.go:382-393`) | **Component:** core/dependency-graph-mutation
**Effort:** S · **Urgency:** eventually

`dependency_graph_mutation.go:85-90` rebuilds and re-analyses the entire task graph once per planned
write — O(W·(V+E)). For **this** caller the check is provably unreachable: every write is additive.
`thread_apply.go:382-393` seeds `desired[edge.To]` from the task's current `DependsOn` and only ever
appends, so each prefix's edge set is a subset of the validated final edge set, and both acyclicity
and reference-validity are monotone under edge-subset. A prefix cannot fail unless the final graph
fails.

Measured on an Apple M5 with `-benchtime 3x` against
`BenchmarkThreadApplyRepresentativeGuardedDryRun`:

| Scale | With prefix loop | Prefix loop disabled | Cost of the loop |
| --- | --- | --- | --- |
| 1,000 tasks / 300 writes (the reference gate) | 522 ms | 212 ms | **310 ms — 59%** |
| 2,000 tasks / 600 writes | 1,992 ms | 726 ms | **1,266 ms — 64%** |

At 2× the representative scale the redundant validator **alone** exceeds the ADR's one-second budget
for guarded read/validate/materialize, while everything else in that phase still fits inside it.

The implementation task records: "Per-file atomic durability, not the owned prefix validator,
dominates the real lock hold, so V1 keeps the simple validator". That is true of the 3.5 s *total*
(the validator is ~9% of it) and the fsync attribution checks out independently — but it is the wrong
comparison for the phase the ADR gives its own separate budget. Against that budget the validator is
the dominant term and the one with superlinear growth, so the recorded trigger condition ("crossing
either budget… triggers incremental prefix-validation work") is much closer than the note implies.

**Recommendation:** Do not delete the loop — `ValidateTaskGraphMutationPlan` is shared with
`MutateTaskGraph`, which serves non-additive plans (`ClearLegacy`, dependency removal) where prefix
validation is genuinely required. Gate it instead: skip per-prefix revalidation when every write's
`DependsOn` is a superset of the task's current set and no `ClearLegacy` is requested, recording the
monotonicity argument in the comment. Then restate the ADR/task performance note in terms of the
phase each budget actually governs.

**Resolution:** The shared validator now recognizes canonical edge-only
supersets and proves their prefixes by monotonicity while retaining graph
reconstruction for removals and legacy clearing. The 1000-task/300-write
planning benchmark fell to about 0.22 seconds; ADR, architecture, and task notes
were corrected.

### Low

#### L1. Apply-time validation does not re-derive compose's "at least one member" invariant  · **Status:** fixed

**File:** `internal/core/thread_apply.go:266`; missing in `PrepareThreadApply` and `internal/core/thread_creation.go:102-131` | **Component:** core/thread-apply
**Effort:** XS · **Urgency:** eventually

`thread_apply.go:266` rejects a memberless manifest, but neither `PrepareThreadApply` nor
`ValidateThreadCreationPlan` has an equivalent check. Editing a materialized plan's `thread.tasks` to
`[]` and running `thread apply` succeeds and creates a memberless Thread.

This contradicts ADR contract #2 ("Apply-time validation is authoritative. Compose-time success
cannot justify… accepting a hand-edited apply plan with invalid creation fields"). Severity stays low
because memberless Threads are legal through `thread new` (README: "empty Threads are valid"), so
nothing invalid is produced — this is a contract asymmetry, not corruption.

Related scope note worth stating in the ADR: the `member: false` gate rule (M1) is *structurally*
compose-only. The plan records members and edges but not non-member nodes, so apply can never
re-derive it. "Apply rechecks every referenced task and Thread invariant" currently reads as if it
could.

**Recommendation:** Add the member-count check to `PrepareThreadApply`, and name in the ADR which
compose-only invariants apply deliberately cannot re-derive.

**Resolution:** PrepareThreadApply now independently rejects a hand-edited
memberless bulk plan, preserving compose's narrower at-least-one-member
contract; focused coverage and ADR/task wording were added.

#### L2. `changed: true` with every operation `skipped` and `committed: false` is reachable  · **Status:** fixed

**File:** `internal/store/threadapply.go:76` and `:165-168` | **Component:** store/thread-apply
**Effort:** XS · **Urgency:** eventually

`result.Changed` is frozen at plan time (`:76`). The final convergence branch can then downgrade the
Thread operation to `skipped` and return `Complete = true` (`:165-168`) without revising `Changed`.
When the plan's only pending operation is the Thread create and another writer creates a
byte-identical Thread between the plan and the final re-prepare, the receipt reads
`changed: true, committed: false, complete: true` with every operation `skipped` — "changed" with
nothing changed and nothing written.

This is a code-path finding, not an end-to-end reproduction: I could not force the window through the
CLI. Two concurrent processes applying the same plan serialize cleanly on the guard, and the loser
correctly reports `changed:false, complete:true, committed:false` with all operations skipped. A
deterministic reproduction is available in-package via `testHookBeforeThreadApplyWrite("thread", …)`.

**Recommendation:** Either recompute `Changed` from the final operation states before returning, or
document `changed` explicitly as "work was outstanding when the plan was authorized" and let the
per-operation `applied`/`skipped` vocabulary carry the durable truth. Today the field's name invites
the stronger reading.

**Resolution:** When final revalidation turns the only pending Thread create
into an exact skip, Changed is recomputed from actually applied operations. A
deterministic raw-create race now pins changed=false, committed=false,
complete=true.

#### L3. `thread compose --dry-run` skips the no-clobber precondition  · **Status:** fixed

**File:** `internal/cli/thread_apply.go:63-66` | **Component:** cli/thread-apply
**Effort:** XS · **Urgency:** eventually

The `--out` existence check lives inside `if !app.DryRun`, so dry-run never evaluates it.
`thread compose --from m1.yaml --out plan1.yaml --dry-run` exits 0 with `"written":false` against an
existing plan file, while the identical real run exits 14 with
`thread apply plan plan1.yaml already exists: conflict`.

The command's own help says "--dry-run prints the same plan without creating the output file", and
the global flag promises "validation still runs". A preview that cannot possibly succeed reports
success — exactly the case a scripted preflight would rely on.

**Recommendation:** Stat `--out` during dry-run and return the same `ErrConflict`.

**Resolution:** Compose dry-run now evaluates the output path's no-clobber
precondition with Lstat and returns the same conflict as real compose; CLI
coverage pins the behavior.

#### L4. `core.ThreadApplyPlan` is published verbatim as the `thread compose --json` contract  · **Status:** fixed

**File:** `internal/wire/thread.go:280` | **Component:** wire/thread
**Effort:** S · **Urgency:** eventually

`wire/thread.go:280` embeds the core type directly (`Plan core.ThreadApplyPlan`). Every other
envelope in `wire` projects through a wire-owned DTO. Two consequences, both verified against
`schema --json-schema` output:

1. The core struct now carries three simultaneous public contracts — the Go API, the durable YAML
   plan file (`yaml:` tags), and the published JSON schema (`json:` tags). A field rename in core is a
   silent breaking change to two published contracts at once, and `SchemaVersion` discipline cannot
   catch it because the golden simply regenerates alongside.
2. `ThreadApplyPlan`, `ThreadApplyThread` and `ThreadApplyDependency` are published with **no
   description**, because the schema-comment registry is wire-scoped and cannot reach core doc
   comments. `ThreadApplyJSON`, the wire-owned sibling, is documented. The compose envelope's most
   important payload is therefore the least documented part of the machine contract.

**Recommendation:** Mirror the plan into `wire.ThreadApplyPlanJSON` (plus thread/dependency) and
register its doc comments, leaving the core type free to evolve and the durable file format owned by
core alone.

**Resolution:** The compose envelope now projects through wire-owned
ThreadApplyPlanJSON, ThreadApplyThreadJSON, and ThreadApplyDependencyJSON DTOs.
Schema comments and golden contracts were regenerated.

#### L5. A completed apply becomes an opaque conflict once the Thread is legitimately used  · **Status:** fixed

**File:** `internal/core/thread_apply.go:502-508` | **Component:** core/thread-apply
**Effort:** S · **Urgency:** eventually

`equivalentPlannedThread` requires `Updated == "" && StartedAt == "" && EndedAt == ""` plus exact
metadata, membership and body equality. Applying a plan successfully, running `thread start`, then
re-applying the same plan exits 14 with
`planned Thread id 6g5b044hbc0e already exists with different content: conflict`. The message names
no differing field and does not distinguish "this plan already landed and the Thread has since moved
on" from "a foreign Thread squats your preallocated ID".

The rule itself is correct per the same-ID/different-content contract. The problem is that the plan
is documented to users as a retry token ("safe to retry", "retry the same materialized plan to
converge"), so a scripted or CI re-run of a *successful* apply fails with a diagnosis that does not
say the apply already succeeded.

**Recommendation:** When the existing Thread matches the plan's ID, slug, created date and membership
but has advanced lifecycle or `updated` state, say so — "Thread … already exists and has since been
modified; this plan appears to have been applied already". In the general case, name the first
differing field.

**Resolution:** Same-ID conflicts now name the first differing field, and a
definition-identical Thread with advanced lifecycle state explicitly says the
plan appears already applied and will not overwrite it. Focused tests cover both
diagnoses.

## Acceptance-criteria traceability

| # | Criterion (abbreviated) | Verdict | Evidence |
| --- | --- | --- | --- |
| 1 | Compose validates every task reference, member role, local key, edge union; no mutation | **Qualified** | References/keys/edges verified by live probes and `TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest`; no mutation confirmed. "Member role" is where M1 lands. |
| 2 | Apply revalidates repository health and planning identity inside the guard | Met | `store/threadapply.go:45-59, 108-125`; identity re-read again at `:156`. Wrong-ID, ID-less, wrong-root all fail closed (live + `TestThreadApplyIdentityAndDryRunFailClosed`). |
| 3 | Compose refuses an ID-less space with an actionable message; no silent path identity | Met | Live: `validation failed: planning repository has no durable id; run 'tskflwctl config migrate'…`, exit 11, for both compose and apply. |
| 4 | Every interrupted prefix graph-valid; retry converges without duplicates | Met | `TestThreadApplyEveryDurablePrefixRetriesToCompletion` asserts `GraphHealthy` after each injected failure; monotonicity argument in M3 proves it independently. |
| 5 | Omitted membership/dependencies never imply destructive removal | Met | Additive union at `core/thread_apply.go:382-393`; live probe preserved an unrelated hand-added edge, body, comment, unknown field and mode 0600. |
| 6 | Receipts distinguish creates, updates, skips, conflicts, completion | **Qualified** | Operation vocabulary and `complete`/`committed` are accurate in every case I produced; `changed` has the narrow race in L2. |
| 7 | Existing task edits between retries don't conflict merely from linkage | Met | Raw `<!-- raw edit -->` content survived a conflict-and-retry cycle (`TestThreadApplyImmediateTargetCASRejectsRawEditAndRetryPreservesIt`, plus live repair probe). |
| 8 | One compound capability, single guard, no nested narrower ports | Met | `MutateThreadApply` calls `materializeTaskGraphPlan` / `materializeThreadCreation` directly and never re-enters `MutateTaskGraph` or `MutateThreadCreation`; `TestThreadApplyPlannerCannotReenterStore` pins the callback sentinel. |
| 9 | Dependencies precede the Thread; prefixes sound; receipts report the durable prefix | Met | Thread create is last (`:178`); `TestThreadApplyInterruptedPrefixRetriesToCompletion` asserts the Thread file is absent after an interrupted prefix. |
| 10 | Caller-provided clock; idempotent skips neither rewrite nor advance timestamps | Met | `now` threaded through `materializeTaskGraphPlan`; `TestThreadApplyPersistsDependenciesThenThreadAndConverges` byte-compares all three files across an idempotent retry; live re-apply produced `changed:false` with no file churn. |
| 11 | No manufactured task-creation projection impacts | Met | No task-creation path exists in V1; inline `new_task` is rejected and every doc reference is framed as future work. |

## Settled concerns

These looked suspicious enough to chase and were disproved by code plus an adversarial reproduction.

1. **Body round-trip breaking idempotent retry.** I expected `buildFile` to normalize the rendered
   template body so that `equivalentPlannedThread`'s `body == planned.Body` would fail on the first
   retry, turning every completed apply into a permanent conflict. It does not: `assembleFile` writes
   the body verbatim after the closing fence and `splitFrontmatter` returns exactly that slice.
   Verified end-to-end, including `tags` (`sortedUnique` returns `nil` for an empty input, matching a
   tag-less parse — a plausible `[]string{}` vs `nil` `DeepEqual` trap that is not present) and
   `target_date`.
2. **Path escape or identity confusion through a hand-edited plan.** Ten probes — `slug:
   ../../../../tmp/pwned`, a non-canonical slug, `id: ../../evil`, an ID colliding with an existing
   task, `status: completed`, `created` ≠ `composed_at`, malformed `composed_at`, a foreign
   `planning_repo_id`, `schema: 0`, `schema: 2` — all rejected with the right sentinel. Duplicate
   members, duplicate edges and self-edges are rejected too.
3. **Manifest format smuggling.** Duplicate YAML keys, unknown fields, multi-document input,
   `schema: 2`, null/empty node lists and an empty document are all rejected. Anchors/aliases resolve
   to literal values with no interpolation; there is no shell expansion anywhere in the path.
4. **Additive intent damaging unrelated state.** A task carrying a leading YAML comment, an unknown
   `custom_field`, a hand-written body section and mode `0600` came through an apply with all four
   intact, `depends_on`/`updated_at` appended in canonical position. `writeFileAtomic` re-stats the
   destination, so the mode is preserved rather than reset to 0644.
5. **Prefix ordering unsafety.** Writes are sorted by task ID, not topologically, which looked wrong
   for a graph mutation. It is safe here because the plan is purely additive (see M3), and
   `TestThreadApplyEveryDurablePrefixRetriesToCompletion` confirms it empirically.
6. **"Serialization" being mere mutual exclusion rather than re-authorization.** Two concurrent CLI
   processes applying the same 40-node/39-edge plan produced exactly one Thread; the loser waited,
   re-read, and re-derived a *fresh* all-skip decision (`changed:false, complete:true,
   committed:false`) rather than replaying a stale one. `TestThreadApplySerializesWithDirectDependencyMutation`
   pins the same property against a direct `MutateTaskGraph` writer.
7. **Programmatic stores applying without authorization.** An `FS` built without
   `WithPlanningIdentityReader` fails closed with `ErrValidation` at `store/threadapply.go:205`
   rather than proceeding. The other three production `store.NewFS` call sites
   (`workspacestore`, `spacestore`, shell completion) are read-only and never reach apply.
8. **Guards that are decoration.** Probes 2, 3 and 4 each break the suite, so the final convergence
   re-read, its pending-dependency check, and the per-target CAS are all genuinely exercised. Only
   the whole-source CAS is not (M2).
9. **Unreproducible performance claims.** The ADR's 0.53 s / 3.46 s reproduce here as 522 ms /
   3.499 s. The claim that per-file atomic durability dominates the *real* apply also holds — 300
   writes × two fsyncs ≈ 10 ms each accounts for essentially all of the 3 s delta. (What the note
   gets wrong is the planning-phase attribution: M3.)
10. **Stale generated artifacts.** `just docs` produces no drift; every golden diff is the 1.56→1.57
    bump plus the new envelope definitions; `envelopes_test.go` adds real schema-validation cases for
    both new envelopes. `--json` arrays are non-nil (`"operations":[]` on a pre-plan failure) and the
    error envelope carries `thread_apply` recovery detail with the attempted Thread ID even when the
    store never accepted the plan.
11. **Dogfooding drift.** `6g3q4rtv8d0a` is an in-flight member of `complete-production-threads`;
    `planning/threads/` is untouched by this branch, and neither review-audit document is a Thread
    member.
12. **Stale inline-`new_task` promises.** None remain: ADR, spike record, task and CLI docs all frame
    it as explicitly rejected in V1 or as future work.
13. **Raw-editor over-claiming.** `docs/ARCHITECTURE.md:237-239` and ADR §418 both disclaim rollback and
    cross-process isolation, matching what the code provides. The wording is appropriately hedged.
14. **Cross-kind ID uniqueness.** The mint checks tasks and Threads but not epics/audits/research.
    Unchanged from the pre-existing `thread new` path, IDs are time-ordered mints, and the file lands
    in its own directory — not a regression introduced here.
