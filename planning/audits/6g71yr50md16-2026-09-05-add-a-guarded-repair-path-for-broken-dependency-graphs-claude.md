---
schema: 1
id: 6g71yr50md16
bucket: closed
area: add-a-guarded-repair-path-for-broken-dependency-graphs-claude
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: add-a-guarded-repair-path-for-broken-dependency-graphs-claude — 2026-09-05

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial design pass for task
[`6g4g8gatbnrs`](../tasks/6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md)
(epic `30-threads-and-task-dependency-graphs`, Thread `6g503c6pfqeb`), read against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and the landed graph substrate.
No implementation was written and no product behaviour was changed.

**Method and isolation.** The repository was copied to a private scratch directory
(`rsync`, excluding `.git`) and every counterexample below was executed there as
throwaway `zz_probe*_test.go` files under `internal/core`, then deleted. The working
tree was untouched throughout: `git status --porcelain` reported the same two modified
planning files before and after, and this audit file is the only repository write.
Claims marked **[probe]** were run, not reasoned about.

## Executive verdict

**The capability is necessary and the task's instincts are right, but three of its seven
acceptance criteria cannot be satisfied as written, and the substrate it plans to reuse is
unsound on exactly the repositories it targets.** The task should be re-scoped before
implementation, not merely estimated.

The core problem is that repair has been framed as a variant of the existing guarded
mutation, and it is not. Ordinary mutation starts healthy and must end healthy; repair
starts broken and — in any realistic broken repository — ends *less broken*. Every
validation rule in `ValidateTaskGraphMutationPlan` is built on the first shape, and three
of them are actively wrong for the second:

1. **The prefix rule is unsatisfiable.** It demands every durable prefix leave a
   non-broken graph. A multi-file repair's prefixes are broken by construction — that is
   what "multi-file repair" means. Confirmed by probe: a legitimate two-write repair of
   two overlapping cycles is rejected at prefix 1 (H1).
2. **The simulator lies about broken repositories.** It rebuilds the prospective graph
   from `graph.Task()` values, discarding unreadable records and duplicate-ID shadow
   records. It therefore both *rejects* plans that correctly preserve an edge into an
   unreadable task and *accepts* plans while reporting a health the repository does not
   have (H2).
3. **The only legacy primitive is field-coarse.** `ClearLegacy` deletes `blocked_by`,
   `dependencies` and `blocks` together. A task with one cyclic `blocked_by` and one
   valid `dependencies` entry cannot be repaired without destroying the valid one, and
   validation raises no objection (H3). That is acceptance criterion 4 failing on the
   first fixture anyone will try.

The fix is not more machinery. It is a smaller, sharper contract: **repair is
removal-only, keyed on source declarations rather than edges, proved by a lexicographic
defect measure plus explicit set containment, and permitted to leave the graph broken.**
Removal-only turns out to buy a theorem — every removal-only plan is monotone in the
measure, so *every* permutation of it is prefix-safe — which dissolves the ordering
problem the task expects to be hard, and simultaneously gives a machine check for
"you deleted a constraint you did not need to delete."

On the user surface: the preference for an explicit user-authored plan is correct **for
cycles and for legacy references only**. Requiring a YAML manifest to delete a self-edge
is ceremony that pushes people back to hand-editing frontmatter, which is the behaviour
this feature exists to replace. Recommend one command with three input modes.

Classifications: **[blocks implementation]**, **[fix during implementation]**,
**[monitor]**.

## Recommended design

### The repair invariant

Define over an immutable snapshot `S`:

- **`C(S)`, the declaration multiset.** Every graph-owned *source occurrence*, as a
  quadruple `(owner task ID, field, raw value, occurrence index)` where field is one of
  `depends_on`, `blocked_by`, `dependencies`, `blocks`. This is deliberately not an edge
  set: it survives duplicates, retains invalid and slug-shaped tokens verbatim, and can
  name a legacy `blocks` declaration whose owner is the *prerequisite* rather than the
  dependent.
- **`D(S)`, the defect multiset.** Every `GraphProblem` keyed by
  `(Code, TaskID, RelatedTaskID, Field, Path)`, plus every legacy reference whose
  resolution is not `resolved`.

A repair plan `P` is an ordered list of removals. With `S₀` the authoritative snapshot
and `Sᵢ` the snapshot after the *i*-th durable file write, the invariant is:

1. **Removal-only.** `C(Sₙ) ⊆ C(S₀)`. No declaration is added, rewritten, reordered or
   reinterpreted. The only other permitted delta is the `updated_at` stamp on a file the
   plan actually changes.
2. **Authorized removal.** Every removed declaration is either *inferable* — defective in
   isolation, with exactly one non-inventing repair — or *operator-selected* by name.
3. **Preservation.** `C(S₀) \ removals(P) ⊆ C(Sₙ)`, checked explicitly. This, not the
   measure, is what protects unrelated valid constraints.
4. **Monotone improvement at every durable prefix.** For every `i ≥ 1`:
   `μ(Sᵢ) <ₗₑₓ μ(Sᵢ₋₁)` and `D(Sᵢ) ⊆ D(S₀)`. No new defect of any class, on any task, at
   any prefix.
5. **Bounded blast radius.** Only files owning a removed declaration are written, and
   every task whose *derived* state changes is reported.
6. **No health promise.** `Sₙ` need not be healthy or even degraded. Residual defects and
   their owners are receipt data, not a failure.

Point 6 is the deliberate departure from the task text. Real broken repositories are
mixed-class, and a repair that refuses until every unrelated identity defect is also
fixed is a repair nobody can run (H4).

### The problem measure and why it is a tuple

`μ(S)` is a six-component lexicographic tuple of non-negative integers, all derivable
from one strict snapshot:

| # | Component | Counts |
|---|---|---|
| 1 | `unreadable` | unreadable task records |
| 2 | `identity` | `missing-task-id`, `task-id-drift`, `duplicate-task-id`, `invalid-status` |
| 3 | `cyclic` | `cycle` problems (one per SCC member) + `self-dependency` problems |
| 4 | `legacy_unsafe` | legacy references resolved but structurally unsafe |
| 5 | `declaration` | `duplicate-dependency`, `invalid-dependency-id`, `missing-dependency`, `legacy-reference-missing`, `legacy-reference-ambiguous` |
| 6 | `legacy_present` | legacy declarations still present, including present-but-empty keys |

Components 1 and 2 are the classes repair *cannot* touch. Putting them first in the
lexicographic order makes "repair never traded an edge fix for an identity regression" a
free consequence of the comparison rather than a separate rule.

**Monotonicity theorem.** If `P` is removal-only over graph-owned declarations, then
`D(S′) ⊆ D(S)` and `μ(S′) ≤ₗₑₓ μ(S)`.

*Proof sketch.* Components 1–2 depend only on record readability, filename/frontmatter
identity and status — none of which a declaration removal touches, so they are unchanged.
Removing a declaration deletes a set of projected edges and adds none; strongly connected
component membership is monotone non-increasing under edge deletion, so component 3 is
non-increasing. `LegacyUnsafe` is defined in `markUnsafeLegacy` by shared cyclic component
or self-edge, so it is non-increasing with component 3. Each defect counted in component 5
is attributed to a specific declaration; removing declarations removes attributions and
creates none. Component 6 counts declarations directly. ∎

Two corollaries do real work:

- **Order independence.** Every permutation of a removal-only plan is prefix-safe.
  Write order is presentation, not recovery data — the exact opposite of
  `planLegacyDependencyMigration`, which must write dependents before legacy `blocks`
  owners. This is the honest answer to design question 3: "every durable prefix must
  improve" *is* achievable for cycles and multi-file repairs, but because of the algebra
  of removal, not because of clever sequencing.
- **A minimality check.** If a removal does not strictly decrease `μ` at its prefix, it
  deleted a declaration that was neither defective nor necessary. Reject the plan and name
  the declaration. This is the machine check behind acceptance criterion 4, and it is
  cheap: rebuilding the snapshot per removal is the cost the existing validator already
  pays for non-additive plans.

**State the limit honestly.** A `μ` decrease at a prefix is necessary but not sufficient
for "the operator removed the *right* edge": for a cycle with several breakable edges,
every choice decreases `μ` identically. That is precisely why cycle-breaking must be
operator-selected and why the preview must enumerate the alternatives it did not take.

### What may be inferred, and what requires a choice

**Inferable** — offered by a bare `repair`, applied by `repair --auto`:

- **duplicate occurrence** — the semantic set is unchanged; this is the only class with
  literally zero constraint loss.
- **self-edge** — exactly one legal repair exists.
- **present-but-empty legacy key** — declares no constraint.
- **invalid ID token** — no non-inventing alternative. The receipt must record the removed
  literal verbatim (L3).
- **dangling reference** (valid ID, no such record) — no non-inventing alternative, but it
  carries real semantic weight: see M1.

**Operator-selected** — never inferred:

- **cycle breaking** — several valid choices, no defensible ranking.
- **legacy reference unsafe only because of a cycle** — same.
- **legacy reference missing or ambiguous** — the value is a slug, i.e. human intent that
  cannot be recovered from the ID space. ADR-0006 already forbids best-guess resolution;
  deleting is likewise a judgement about intent.
- **any removal that fails the minimality check** — refused outright, not downgraded to a
  prompt.

### Core types

```go
// DeclarationRef is the unit of graph-owned repair: one source occurrence, not
// an edge. It can name a legacy `blocks` declaration whose owner is the
// prerequisite, a second duplicate occurrence, or a token that is not an ID.
type DeclarationRef struct {
    OwnerTaskID string // the task file that declares it
    Field       string // depends_on | blocked_by | dependencies | blocks
    Value       string // the declared token, verbatim
    Occurrence  int    // 0-based index among identical Values in that field
}

type RepairAction string // drop | dedupe

type RepairReason string // self-edge | duplicate | invalid-id | dangling |
                         // legacy-missing | legacy-ambiguous | legacy-empty-field |
                         // cycle-break | legacy-unsafe

type RepairOperation struct {
    Action     RepairAction
    Target     DeclarationRef // Occurrence ignored for dedupe
    Reason     RepairReason
    Selected   bool           // true when the operator named it
}

type TaskGraphRepairPlan struct{ Operations []RepairOperation }

// GraphMeasure is the six-component tuple, named so a --json consumer can
// assert monotonicity itself rather than trusting the receipt's verdict.
type GraphMeasure struct {
    Unreadable, Identity, Cyclic, LegacyUnsafe, Declaration, LegacyPresent int
}
```

The receipt, following the lifecycle slice's durability contract:

```go
type GraphRepairReceipt struct {
    DryRun, Changed, Committed bool
    Workspace                  WorkspaceIdentity

    Operations []RepairOperationOutcome // removed | already-absent | pending | refused

    Before, After          GraphMeasure
    BeforeHealth, AfterHealth GraphHealth
    ResolvedProblems       []GraphProblem // D(S₀) \ D(Sₙ)
    ResidualProblems       []GraphProblem // D(Sₙ), each tagged repairable | needs-direct-edit
    PreservedDeclarations  int            // the preservation witness

    StateImpacts  []TaskGraphStateImpact    // before/after role, gate, soundness
    ThreadImpacts []ThreadProjectionImpact  // see M1

    PlannedTaskIDs, AppliedTaskIDs, RemainingTaskIDs []string
}
```

`Changed` must be derived from the materialized write set, not from plan length (M4).
`Committed` must exist here as it does on `TaskLifecycleMutationResult` (M5).

### Ports versus adapter

**Core owns**, purely and without a filesystem:

- `TaskGraph.Declarations() []DeclarationRef` and a `CanonicalDependencies(taskID)` that
  is explicitly *not* the legacy-unioned `Prerequisites()` (L1).
- `Measure(*TaskGraph) GraphMeasure` and its lexicographic comparison.
- `SimulateRepair(TaskGraphRead, TaskGraphRepairPlan) *TaskGraph` — reconstructing from
  the full read (records **and** load problems), which is the H2 fix.
- `PlanGraphRepair(*TaskGraph, RepairRequest) (TaskGraphRepairPlan, RepairPreview, error)`.
- `ValidateGraphRepairPlan(*TaskGraph, TaskGraphRepairPlan) (TaskGraphRepairPlan, error)`
  enforcing invariants 1–5.
- `RepairFingerprint(*TaskGraph) string` — a digest over the ordered declaration multiset
  plus the sorted defect multiset, for manifest staleness (L2).

**A new port**, a sibling capability rather than a flag:

```go
type TaskGraphRepairPlanner func(*TaskGraph) (TaskGraphRepairPlan, error)

type TaskGraphRepairStore interface {
    RepairTaskGraph(now time.Time, dryRun bool, planner TaskGraphRepairPlanner) (TaskGraphRepairResult, error)
}
```

Separate from `TaskGraphMutationStore` on purpose. The source precondition is *inverted* —
repair requires a broken graph, mutation requires a non-broken one — and a boolean on the
existing entry point is exactly the general escape hatch ADR-0006 and this task both
forbid. A distinct interface also keeps read-only and test adapters from acquiring it by
accident and makes every call site greppable.

**The filesystem adapter owns** the canonical-root guard, the authoritative load, the
whole-snapshot and immediate-target CAS, surgical frontmatter removal, atomic replacement,
durable-prefix reporting, and planning-identity re-verification under the guard (the
pattern `MutateThreadApply` already established). It gains one obligation: a source token
on unreadable records, so the CAS stops being blind to them (H6).

**The CLI adapter owns** manifest decoding. Core takes decoded, typed operations, never a
path — the rule bulk apply already follows.

### Validation and materialization

```
ValidateGraphRepairPlan(graph, plan):
    if plan.Operations is empty: return ErrValidation "no repair operations"
    declarations := graph.Declarations()                   # source-level, occurrence-indexed
    seen := {}
    for op in plan.Operations:
        if op.Target not in declarations:                  # idempotent retry lands here
            mark op already-absent; continue
        if op.Target in seen: return ErrValidation "declaration named twice"
        seen += op.Target
        if reasonRequiresSelection(op.Reason) and not op.Selected:
            return ErrValidation "reason %s requires an explicitly named declaration"

    current := graph
    measure := Measure(current)
    for op in applyOrder(plan):                            # order is presentation only
        next := SimulateRepair(sourceRead, plan.prefixThrough(op))
        m := Measure(next)
        if not lexLess(m, measure):
            return ErrValidation
                "removing %s is not necessary to reach the repaired state; it would "
                "delete a valid constraint — use `task depend remove` after repair"
        if not defectsContained(next, graph):              # invariant 4, second half
            return ErrValidation "repair would introduce a new defect: %s"
        current, measure = next, m

    if not preserved(graph, current, plan):                # invariant 3
        return ErrValidation "repair would drop declarations it did not name: %s"
    return normalized(plan)
```

`SimulateRepair` is the piece that must not be `taskGraphFromMap`. It rebuilds from the
adapter read so unreadable records keep suppressing `missing-dependency`, duplicate-ID
shadow records keep their defects, and the simulated health is the health the repository
will actually have.

Materialization mirrors `materializeTaskGraphPlan` with three changes:

```
materializeRepair(graph, plan, now):
    group plan.Operations by OwnerTaskID              # one atomic write per file
    for each owner in stable ID order:
        path := resolvePath(owner)                    # ErrAmbiguous here means duplicate
                                                      # task ID — classify it as that,
                                                      # not as a resolution failure (H2)
        content := read(path)
        for each field touched:
            newValues := removeOccurrences(currentValues, ops)   # surgical, order-preserving
            updates[field] = newValues or UnsetField{} if empty
        newContent := updateFrontmatter(content, updates)
        if newContent == content: skip                # never counts toward Changed
        updates["updated_at"] = now
        reparse and assert the exact surviving declaration list  # preservation, again
        emit write{owner, path, ifVersion: hash(content), content: newContent}
```

The re-parse assertion is deliberately stronger than the existing
`slices.Equal(parsed.DependsOn, planned.DependsOn)`: it asserts the *surviving* list,
including duplicates that were meant to survive and legacy fields that were meant to be
left alone.

### Command and manifest contract

One command, three input modes. `lint --fix --graph` is rejected outright (L4).

```
tskflwctl task depend repair                     # diagnose: what is wrong, what is inferable,
                                                 # what needs a choice. Read-only, always safe.
tskflwctl task depend repair --auto              # apply the inferable class only
tskflwctl task depend repair --drop <task>:<field>=<value>[#occurrence]   # repeatable
tskflwctl task depend repair --dedupe <task>:<field>=<value>
tskflwctl task depend repair --plan repair.yaml  # manifest, for multi-operation repairs
```

`--auto` composes with `--drop`/`--dedupe`. `--dry-run` and `--json` are the existing
persistent flags and behave as they do for every other guarded mutation: same guard, same
validation, no replacement, and therefore no promise that a later real invocation sees the
same repository.

The bare form as the default is the important choice. Someone reaching for this command is
looking at a repository that just refused a mutation; the first thing they need is a
diagnosis, not a write.

**Manifest** (schema 1, unknown fields rejected, read from a durable path):

```yaml
schema: 1
planning_repo: 6g3a1wtx4zrr9k2m          # verified under the guard, as bulk apply does
graph_fingerprint: sha256:9f2c1a…        # optional; a mismatch is ErrConflict, not a silent apply
operations:
  - drop: {task: 6g4g8gatbnrs, field: depends_on, value: 6g6scc9jgxae}
    because: cycle-break
  - dedupe: {task: 6g5f1d23jy1b, field: depends_on, value: 6g3q4rt7mgjn}
  - drop: {task: 6g5fthzwbeq1, field: blocked_by, value: portable-thread-reads}
    because: legacy-unsafe
```

The fingerprint is a *semantic* staleness check over declarations and defects, not a byte
hash: reformatting an unrelated task should not invalidate a hand-authored repair plan,
whereas any change to any declaration must. It sits beside — not instead of — the
whole-snapshot byte CAS the guard already performs before the first write.

**Human preview**, for a two-node cycle:

```
Repository graph: broken → would become degraded
  cyclic 2 → 0 · declaration 1 → 0 · legacy_present 1 → 1

Inferable (apply with --auto)
  drop   6g5f1d23jy1b · depends_on · "6g3q4rt7mgj"        invalid task id
         planning/tasks/6g5f1d23jy1b-….md

Needs a choice — dependency cycle across 2 tasks
  6g4g8gatbnrs → 6g6scc9jgxae → 6g4g8gatbnrs
  breaking it requires removing exactly one of:
    --drop 6g4g8gatbnrs:depends_on=6g6scc9jgxae   (cut v0.19.0 no longer gates repair)
    --drop 6g6scc9jgxae:depends_on=6g4g8gatbnrs   (repair no longer gates cut v0.19.0)

Not repairable here — 1 problem needs a direct file edit
  6g5rwjqeh6a6  missing stable task id in frontmatter (field id)

Nothing was written. Re-run with --auto and/or --drop to apply.
```

Enumerating the alternatives is the honest answer to "multiple valid ways to break a
cycle", and it is cheap: for each edge inside the SCC, test whether its removal reduces
cyclic membership. Cap the enumeration and say so when capped.

`--json` carries the whole receipt: both measures component-by-component, per-operation
outcome, resolved and residual problems with an owner tag, state and Thread impacts,
workspace identity, and the durable prefix. `changed` under `--dry-run` means "would
change", as it does for bulk apply.

### Lint and status guidance

Today one function, `taskGraphHealthDetail`, produces the remedy for mutation refusals,
`board`, `status`, the TUI overview and the atlas, and it appends the same sentence
regardless of defect class: *"repair the graph-owned frontmatter directly, then run
`tskflwctl lint`"* — or, whenever any legacy occurrence remains, *"run `tskflwctl task
depend migrate`"*, which on a broken graph refuses (M3).

Replace the single string with a per-class remedy:

| Defect class | Remedy once repair exists |
|---|---|
| self-edge, duplicate, invalid id, dangling, empty legacy key | `task depend repair --auto` |
| cycle, legacy reference unsafe because of a cycle | `task depend repair` (preview, then choose) |
| legacy reference resolved | `task depend migrate` (unchanged) |
| unreadable, missing/drifting/duplicate id, invalid status | direct edit — name the file, and say repair cannot fix it |

Where a repository needs both, the detail should say so in order: *repair first, then
migrate*. That two-step is the common real case and the current message actively
mis-routes it. `lint --fix` keeps its documented promise to never touch graph-owned
fields and gains a pointer to `repair`, satisfying the task's last acceptance criterion.

## Non-goals

- No general `--force` for graph writes, and no `allowBroken` boolean on
  `MutateTaskGraph`.
- No automatic resolution of slug-shaped or ambiguous references to task IDs.
- No addition, rewriting or reordering of declarations — removal only.
- No lifecycle or Thread membership changes. Repair *reports* projection impact; it never
  writes a status or a `tasks:` list.
- No promise of an all-files transaction, and no rollback. The durable-prefix contract
  plus removal monotonicity is the recovery story.
- No repair of identity-class defects (unreadable, missing/drifting/duplicate id, invalid
  status). Those belong to `lint --fix` and direct edits.
- No `lint --fix --graph`.

## Failure and recovery matrix

| Event | Classification | Durable state | Receipt | Recovery |
|---|---|---|---|---|
| Guard acquisition fails | `ErrConflict` | none | none | retry |
| Snapshot load fails | wrapped I/O error | none | none | fix the filesystem |
| Manifest decode fails | `ErrValidation` (exit 11) | none | none | fix the manifest |
| Fingerprint mismatch | `ErrConflict` (exit 14) | none | names changed declarations | re-preview, re-author |
| Operation names an absent declaration | success, `already-absent` | none for that op | outcome recorded | none needed — this is retry convergence |
| Removal fails the minimality check | `ErrValidation` | none | names the declaration | drop it, or use `task depend remove` after repair |
| Plan would add a declaration | `ErrValidation` | none | names it | reject — removal-only is structural |
| Whole-snapshot CAS fails pre-write | `ErrConflict` | none | none | retry from a fresh snapshot |
| Per-file CAS fails at write *k* | `ErrConflict` | writes `1…k-1` | applied prefix + remaining | re-run the same plan; applied ops become `already-absent` |
| Write *k* fails (I/O, permissions) | wrapped error | writes `1…k-1` | applied prefix + remaining | fix the cause, re-run the same plan |
| Process killed after write *k* | — | writes `1…k` | none emitted | re-run the same plan; monotonicity guarantees the partial state is strictly better than the start |
| All writes succeed, guard release fails | typed failure, `Committed: true` | all | complete | **never auto-retry**; report and reload |
| Retry of a fully applied plan | success, `Changed: false` | unchanged | every op `already-absent` | none |
| Concurrent cooperating writer | serialized by the guard | consistent | — | second operation re-authorizes from fresh state |
| Concurrent raw editor of a task file | `ErrConflict` at that file's CAS | prefix only | applied prefix | re-preview — the plan may no longer be minimal |
| Concurrent raw editor of an *unreadable* file | **undetected today** | prefix | wrong | H6 — needs the load-problem source token |

Two entries carry the load. The killed-process row is where removal monotonicity pays for
itself: with an additive or rewriting plan, a partial application can be worse than the
start; with removal-only it provably cannot be. The last row is the one real correctness
gap in the existing substrate.

## Test strategy

**Pure core, per defect class.** One fixture each for dangling reference, invalid token,
duplicate declaration, self-edge, simple cycle, two overlapping cycles sharing a node,
legacy `blocked_by` cycle, legacy `blocks` cycle (owner ≠ dependent), mixed canonical and
legacy on one task, and a task carrying both a defective and a valid legacy declaration
(the H3 fixture). Each asserts the preview, the plan, the measure delta, and the exact
surviving declaration list.

**Property tests for the theorem.** Generate random small graphs with injected defects and
random removal subsets; assert `D(S′) ⊆ D(S)` and `μ(S′) ≤ₗₑₓ μ(S)` for every subset, and
that every permutation of a plan produces the same final snapshot. This is the cheapest
possible proof that prefix safety is order-independent, and it will catch a future
"helpful" normalization that quietly makes a repair non-monotone.

**Preservation tests.** For every plan, assert that each declaration not named survives
byte-identically in the reparsed file — including a duplicate that was meant to survive
and an unrelated legacy field on the same task.

**Simulator soundness (H2 regression).** A repository with one unreadable file plus one
cycle: assert the plan preserving an edge into the unreadable task validates, and assert
the simulated health is `broken`, not `healthy`. Same shape for duplicate task IDs, plus
an assertion that the write-time failure is classified as a duplicate-ID defect rather
than a bare `ErrAmbiguous`.

**Durable-prefix tests** through the existing hooks: `testHookAfterGraphWrite` to inject
failure after each prefix and assert the receipt's applied/remaining split and that
re-running converges; `testHookBeforeGraphVerify` and `testHookBeforeGraphWrite` to
interleave a raw edit before the whole-snapshot CAS and immediately before one target's
CAS.

**Adversarial concurrency**, two real services over distinct store instances, per the
2026-08-29 amendment's cooperating-writer rule:

- Two repairs targeting the same SCC. The second must wait on the guard, re-authorize from
  fresh state, and report `already-absent` — never double-delete, never claim a second
  cycle break.
- Repair racing `task depend add`. The add must fail closed on the broken source; after
  repair commits, a fresh add must authorize.
- Repair racing `task start` / a Thread membership mutation, asserting the guard serializes
  all three.
- **Repair racing a hand-fix of an unreadable file.** This test must *fail* against the
  current `SameSourceSnapshot` and pass after the load-problem source token lands. It is
  the acceptance test for H6 and should be written first.

**Mutation testing on the two new mechanisms**, matching the bar the 2026-08-27 audits set:
delete the minimality check and assert a targeted test fails; make `SimulateRepair` fall
back to `taskGraphFromMap` and assert the H2 regression test fails.

**Parity.** Human and `--json` name the same operations, measures, residuals and impacts;
the JSON validates against `schema --json-schema`.

## Alternatives considered and rejected

- **`lint --fix --graph`.** Rejected. `lint --fix` is run casually and in CI and is
  documented as never normalizing graph-owned fields; ADR-0006's "every first-party write
  honors graph ownership" names lint-repair paths explicitly. Reachability from `--fix` is
  the door this whole design exists to keep shut.
- **A boolean on `MutateTaskGraph` (`allowBroken`).** Rejected. One entry point with two
  inverted source preconditions is unreviewable, and the boolean *is* the general escape
  hatch the task lists as out of scope. A separate capability costs one interface.
- **Extending `task depend remove` to work on a broken graph.** Rejected. Same verb, two
  preconditions, and the failure messages would have to explain which mode the user was in.
  This is the "misleading error classification" trap.
- **Heuristic cycle breaking** (drop the newest edge, the lowest-priority task's edge, the
  edge added last). Rejected. No defensible ranking exists, and `updated_at` is stamped by
  unrelated writes, so it does not even identify the newest edge.
- **A desired-state plan** (`depends_on: [...]` per task, as the migration planner uses).
  Rejected for repair. Preservation becomes unprovable: a typo in a hand-authored manifest
  silently deletes constraints and no validator can tell that from intent. Removal-only
  makes preservation a set-containment check.
- **Transactional multi-file apply** (staging directory plus swap). Rejected. It cannot be
  atomic across N files without replacing the flat layout's directory, it breaks
  cooperation with raw editors, and removal monotonicity already makes every partial state
  safe. Rollback would buy tidiness, not correctness.
- **A graph library** (`dominikbraun/graph`, Gonum). Rejected. The algorithms are not the
  gap — Tarjan plus Kahn already give exact SCC membership and deterministic
  representatives in O(V+E). The gap is a *source-level declaration model* underneath the
  graph, which no library provides. This is consistent with the 2026-08-30 amendment's
  rule: research a package when a named operation exceeds the owned algorithms, not
  because another graph command shipped.
- **Manifest-only input.** Rejected as the sole door, kept as one of three. Requiring YAML
  to delete a self-edge is friction that returns users to hand-editing frontmatter.
- **Auto-resolving slug-shaped invalid tokens to task IDs.** Rejected — already out of
  scope in the task, and it invents data. Worth a regression test, not a feature flag.

## Findings

#### H1. The prefix predicate is unsatisfiable for every multi-file repair · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/dependency_graph_mutation.go:94 | **Component:** core/plan-validation
**Effort:** M · **Urgency:** acute

`ValidateTaskGraphMutationPlan` rebuilds the graph after each planned write and rejects the
plan if that prefix graph is `GraphBroken`. A repair begins from a broken graph, so every
prefix before the last is broken by construction.

**[probe]** Two overlapping cycles sharing node X (`X→Y→X` and `X→Z→X`, three cycle
problems). The correct two-write repair — clear `Y.depends_on`, then `Z.depends_on` — is
rejected at prefix 1:

```
validation failed: planned write prefix ending at task …mgjy would leave a broken graph:
dependency cycle: …mgjx -> …mgjz -> …mgjx … (1 additional problem(s))
```

The prefix was in fact a strict improvement — three cycle problems fell to two — which is
what acceptance criterion 2 asks for and what the implemented predicate cannot express.

**Recommendation:** repair must not reuse `ValidateTaskGraphMutationPlan`. Give it a
predicate of strict lexicographic decrease in `μ` plus defect-set containment, per the
invariant above. Leave the existing validator untouched for add/remove/migrate, where
"prefix stays non-broken" is both correct and cheap.

**Resolution:** Repair gets a dedicated validator: durable prefixes preserve
authorized subtractive intent and the full plan strictly improves structural
defects without requiring intermediate or final health.

#### H2. Plan simulation is unsound on unreadable and duplicate-ID repositories · **Status:** tracked by 6g721vewvvrz

**File:** internal/core/dependency_graph_mutation.go:126 | **Component:** core/plan-validation
**Effort:** M · **Urgency:** acute

`taskGraphFromMap` builds the prospective graph with `NewTaskGraph(tasks, nil)` from
`graph.Task()` values. That discards two things the live snapshot models: unreadable
records (tracked in `unreadableIDs`, not in `tasks`) and duplicate-ID shadow records
(collapsed to one representative in `g.tasks`). Three consequences, all probed:

1. **False reject.** With `…mgjb` unreadable, a plan that *preserves* `…mgja`'s existing
   edge to it fails: `planned dependency …mgjb for task …mgja does not exist`. The live
   graph deliberately suppresses `missing-dependency` for unreadable prerequisites; the
   simulation does not. Repair would be forced to delete a valid constraint to pass
   validation — acceptance criterion 4, inverted.
2. **False accept.** In the same repository, the plan that *drops* the edge validates with
   no error, because the prospective graph — having discarded the unreadable record —
   reports `healthy`. Identically, a repository with two files claiming one task ID
   validates a plan cleanly while remaining `broken`. Any repair receipt built on
   `finalGraph.MutationReady()` would claim health the repository does not have.
3. **Misclassified write failure.** A plan targeting a duplicated task ID passes
   validation, then `materializeTaskGraphPlan` calls `s.resolvePath`, which returns
   `ErrAmbiguous` (exit 13). The user is told their reference was ambiguous; the truth is
   that two files claim one ID.

**Recommendation:** simulate from the full `TaskGraphRead` (records plus load problems),
not from `graph.Task()`. Classify a duplicate-ID resolution failure as that defect.

**Resolution:** The prerequisite owns a full source projection and prospective
simulator that retain unreadable and duplicate-ID records.

#### H3. `ClearLegacy` destroys unrelated valid legacy constraints, unchecked · **Status:** tracked by 6g721vewvvrz

**File:** internal/core/store.go:66 | **Component:** core/ports
**Effort:** M · **Urgency:** acute

`TaskDependencyWrite.ClearLegacy` is a single boolean that unsets `blocked_by`,
`dependencies` and `blocks` together. It is the only primitive that can remove a legacy
declaration, and legacy declarations are exactly where the deadlocked repair cases live.

**[probe]** Task `…mgjb` carries `blocked_by: [alpha]` — cyclic, resolution `unsafe`,
must go — and `dependencies: [gamma]` — resolution `resolved`, structurally fine. The
plan `{TaskID: …mgjb, DependsOn: nil, ClearLegacy: true}` validates with `err=nil` and
silently deletes the `gamma` constraint. Nothing in the validator objects; nothing in the
receipt would name it.

**Recommendation:** repair operates on `DeclarationRef` values, never whole fields. Leave
`ClearLegacy` to `task depend migrate`, where clearing everything is the intended
semantics because every reference has just been made canonical.

**Resolution:** Repair will target individual source declarations;
migration-only ClearLegacy will not be reused.

#### H4. Requiring a healthy final graph makes repair impossible in mixed-defect repositories · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/dependency_graph_mutation.go:103 | **Component:** core/plan-validation
**Effort:** S · **Urgency:** acute

`ValidateTaskGraphMutationPlan` ends with `finalGraph.MutationReady()`, which is true only
for `GraphHealthy`. Repositories that need repair are rarely single-class.

**[probe]** One task missing its frontmatter `id` plus an unrelated two-node cycle. The
cycle repair is rejected — not for anything about the cycle, but because the identity
defect repair cannot touch survives:

```
validation failed: planned write prefix ending at task …mgjc would leave a broken graph:
missing stable task id in frontmatter … (field id)
```

The same rule blocks the natural staged recovery `broken → degraded → task depend migrate`,
since a graph with any remaining legacy occurrence is `degraded`, not healthy.

**Recommendation:** repair succeeds on strict improvement. Report residual defects with an
owner tag (`repairable` versus `needs-direct-edit`) and let the operator work the classes
in whatever order suits them. Reframe the task's third acceptance criterion from
"converge to the intended healthy state" to "converge to the intended repaired state."

**Resolution:** Selected repairs may commit with explicitly reported residual
broken state when the structural state strictly improves.

#### H5. Edge-keyed repair cannot name a legacy `blocks` declaration · **Status:** tracked by 6g721vewvvrz

**File:** internal/core/dependency_graph.go:596 | **Component:** core/graph
**Effort:** M · **Urgency:** acute

`blocks: [x]` on task P projects the edge P→Q but is declared in **P's** file. ADR-0006's
"the dependent task is the single source of truth" holds for canonical edges only; the
legacy vocabulary breaks it. Any plan shape keyed on "dependent plus its complete
`depends_on` set" — which is what `TaskDependencyWrite` is — structurally cannot express
its removal.

This is not hypothetical: it is the deadlock class. **[probe]** A canonical edge plus a
legacy `blocked_by` slug reference forms a cycle; `resolveLegacyDiagnostics` marks the
reference `unsafe`, `MigrationReady()` is false, so `task depend migrate` refuses — and
`ValidateTaskGraphMutationSource` refuses the whole guarded path anyway because health is
`broken`. Removing the canonical edge instead is rejected with *"planned dependency state
is degraded … run `tskflwctl task depend migrate`"*, advice that cannot work (see M3).
Only clearing the legacy declaration resolves it.

**Recommendation:** key repair on declarations. Widen the task's scope line from "canonical
cycles" to cycles in the *projected* graph, and name legacy `blocks` explicitly.

**Resolution:** The source model explicitly preserves legacy blocks ownership
independently from projected edge direction.

#### H6. Whole-snapshot CAS is blind to concurrent edits of unreadable files · **Status:** tracked by 6g721tvf4crh

**File:** internal/core/dependency_graph.go:848 | **Component:** store/cas
**Effort:** M · **Urgency:** acute

`SameSourceSnapshot` compares health, the readable ID list, per-task paths and
`SourceVersion`, then the problem and legacy diagnostic lists. Unreadable records
contribute only a `GraphProblem{Code, TaskID, Path, Message}` — `TaskGraphLoadProblem`
carries no source token — so two snapshots taken either side of an arbitrary rewrite of an
unreadable file compare **equal** whenever the file still fails to parse with the same
message. **[probe]** confirmed: `SameSourceSnapshot` returns `true` across exactly that.

For ordinary mutation this was tolerable, since a broken graph is refused outright. For
repair it is not: the operator's own editor is very likely open on precisely those files,
and the guard is advisory.

**Recommendation:** add an opaque `SourceVersion` to `TaskGraphLoadProblem` — the same
contract as `domain.Task.SourceVersion`, supplied by the adapter and never parsed by core
— and include it in `SameSourceSnapshot`. This is a port change; consider landing it ahead
of repair so the concurrency test can be written first.

**Resolution:** A prerequisite adds opaque source revisions to unreadable load
problems and covers same-error content races.

#### M1. Removing a dangling edge silently changes soundness, eligibility and Thread completability · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/dependency_graph.go:686 | **Component:** core/graph
**Effort:** M · **Urgency:** soon

A dangling or cyclic prerequisite forces `GateBroken` and `broken` soundness. Deleting the
declaration flips the owner's gate to clear; if that task is `completed` it becomes
*soundly* completed, which propagates: downstream tasks can become eligible, and a Thread
that `thread complete` refused can become completable. The dangling reference was, in
effect, holding a constraint open.

Excluding lifecycle and membership *writes* from repair is right. But the task has no
acceptance criterion for **reporting** derived-state or Thread impact, and the 2026-08-30
amendment already requires guarded task lifecycle receipts to report every Thread whose
projection changes. Repair mutates the same projections through a different door.

**Recommendation:** load Threads under the repair guard, as `MutateTaskLifecycle` does, and
carry `StateImpacts` and `ThreadImpacts` in the receipt and preview. Add an acceptance
criterion for it. Decide separately whether a malformed Thread should *block* a repair —
see the decisions section.

**Resolution:** Repair previews and receipts report task-state and readable
Thread projection impacts plus incomplete Thread evidence; malformed Threads do
not block graph recovery.

#### M2. Acceptance criterion 2 conflates progress with preservation · **Status:** tracked by 6g4g8gatbnrs

**File:** planning/tasks/6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md:33 | **Component:** planning
**Effort:** S · **Urgency:** soon

The criterion asks for one documented measure that both proves monotone progress *and*
proves no unrelated valid constraint was discarded. No measure can do both: removing an
innocent edge from a cycle reduces every cycle-based measure exactly as much as removing
the guilty one does.

**Recommendation:** split it. `μ` proves progress and minimality; explicit set containment
(`C(S₀) \ removals ⊆ C(Sₙ)`) proves preservation. Both are cheap; conflated, neither is
implementable as written.

**Resolution:** The acceptance contract now proves structural progress and
declaration preservation independently.

#### M3. Broken-graph remedies route the user to a command that will refuse · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/dependency_graph_mutation.go:151 | **Component:** cli
**Effort:** S · **Urgency:** soon

`taskGraphHealthDetail` appends *"run `tskflwctl task depend migrate`"* whenever legacy
occurrences remain and no `GraphProblem` outranks them. **[probe]** on a
legacy-induced-cycle repository, rejecting a canonical repair produced exactly that advice
— and `migrate` refuses on that repository, because `ValidateTaskGraphMutationSource`
rejects a `broken` source. The user is handed a loop.

The same function feeds `board`, `status`, the TUI overview and the atlas, so the wrong
remedy is repeated on every surface.

**Recommendation:** make the remedy a function of the defect class, not of which list is
non-empty, per the table above; and when a repository needs both, state the order —
repair, then migrate.

**Resolution:** The command must provide defect-specific remedies and name
repair-before-migrate ordering where migration would refuse.

#### M4. `Changed` is derived from plan length, not from materialized writes · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/dependency_operations.go:166 | **Component:** core/graph
**Effort:** S · **Urgency:** soon

`dependencyReceipt` sets `Changed: len(planned) > 0`, while `materializeTaskGraphPlan`
silently drops byte-identical writes (`if bytes.Equal(content, newContent) { continue }`).
For add/remove the planner returns an empty plan when nothing changes, so the gap is
currently masked. For repair, "did this actually change anything" is the primary question a
receipt answers — especially on a retry after a partial apply, where the honest answer is
often "no, it was already converged."

**Recommendation:** derive `Changed` from the materialized write set. A planned task whose
write was a byte-level no-op should not appear in `RemainingTaskIDs` on failure either.

**Resolution:** Changed and remaining work are now required to derive from
actual materialized writes.

#### M5. The graph mutation result has no durability outcome · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/store.go:80 | **Component:** core/ports
**Effort:** S · **Urgency:** soon

`TaskLifecycleMutationResult` carries `Committed`, and the 2026-08-29 amendment makes it a
reusable constraint: a guard-release or cleanup failure after the atomic write returns a
typed failure carrying the committed receipt and must never enter a retry loop.
`TaskGraphMutationResult` has no equivalent — callers infer commit from a non-empty
`AppliedTaskIDs`, which conflates "committed everything" with "committed a prefix."

**Recommendation:** `TaskGraphRepairResult` carries `Committed` explicitly from the start,
and `runDependencyMutation`'s retry guard keys on it rather than on prefix length.

**Resolution:** The typed repair result and partial failure must carry an
explicit Committed outcome.

#### M6. Retry semantics for duplicate declarations are undefined · **Status:** tracked by 6g721vewvvrz

**File:** internal/domain/task.go:45 | **Component:** core/graph
**Effort:** S · **Urgency:** soon

Duplicates are the one defect class where a single `(owner, field, value)` names more than
one declaration. **[probe]** confirms the parser preserves them
(`Task().DependsOn == ["…mgja", "…mgja"]`) while `Prerequisites()` collapses them, so the
defect is visible only through the raw record. A removal keyed on value alone is ambiguous;
a removal keyed on occurrence index is wrong after a partial apply, because the surviving
occurrence renumbers.

**Recommendation:** a distinct `dedupe` operation meaning "keep exactly one", idempotent by
construction. Reserve occurrence indexing for `drop`, where the operator is naming a
specific occurrence deliberately.

**Resolution:** The declaration model defines idempotent dedupe as retain
exactly one, avoiding occurrence-index retries.

#### L1. There is no canonical-only prerequisite projection · **Status:** tracked by 6g721vewvvrz

**File:** internal/core/dependency_graph.go:913 | **Component:** core/graph
**Effort:** S · **Urgency:** eventually

`Prerequisites()` deliberately returns the union of canonical and safely-resolved legacy
edges, deduplicated. `Task().DependsOn` is the only canonical view, and it is the raw list
including duplicates and invalid tokens. A planner reading `Prerequisites` would treat a
legacy-projected constraint as canonical and plan a removal that writes the wrong file.

**Recommendation:** add a named `CanonicalDependencies(taskID)` and document the three
views — raw record, canonical set, projected union — where they are defined.

**Resolution:** The prerequisite explicitly separates raw declarations,
canonical edges, and the canonical-plus-legacy projected union.

#### L2. `Task()` clears `SourceVersion`, so a pure planner cannot fingerprint its snapshot · **Status:** wontfix

**File:** internal/core/dependency_graph.go:773 | **Component:** core/graph
**Effort:** S · **Urgency:** eventually

Deliberate and correct — persistence tokens are not planner data — but it means a
user-authored manifest cannot be checked against "the repository I looked at" using
anything the planner can see.

**Recommendation:** core computes a semantic `RepairFingerprint` over the declaration and
defect multisets. It is more useful than a byte hash anyway: an unrelated reformat should
not invalidate a hand-authored plan, and any declaration change must.

**Resolution:** A strict semantic plan fingerprint would reject a valid retry
after a durable prefix. The accepted design uses convergent intent reauthorized
against each current guarded snapshot, with whole-snapshot and per-file CAS.

#### L3. Removing an invalid dependency token deletes recoverable human intent · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/dependency_graph.go:417 | **Component:** cli
**Effort:** XS · **Urgency:** eventually

`depends_on: [some-slug]` is almost always someone typing a slug where an ID belongs.
**[probe]** confirms the token survives into `Prerequisites()` verbatim, so it is
recoverable right up to the moment repair deletes it.

**Recommendation:** allow inferable removal, but record the removed literal verbatim in the
receipt so it survives in `--json` output and scrollback. Keep "never resolve a slug-shaped
token to an ID" as an explicit regression test, not just a scope line.

**Resolution:** Invalid and dangling values require explicit drop and remain
verbatim in diagnosis, preview, and receipts.

#### L4. Repair must not be reachable from `lint --fix` · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/cli/lint.go:21 | **Component:** cli
**Effort:** XS · **Urgency:** eventually

`lint --fix` is documented as never normalizing graph-owned fields, and ADR-0006's
"every first-party write honors graph ownership" names lint-repair paths specifically.
The task lists `lint --fix --graph` as a live CLI option to decide between.

**Recommendation:** close it explicitly in the task's scope rather than leaving it open.
`lint` should *point at* repair; it should never perform one.

**Resolution:** The task and ADR now explicitly prohibit repair through lint
--fix; lint only points to diagnosis.

## Decisions that need maintainer input

1. **Does repair accept leaving the graph broken?** The design above says yes and treats
   residual defects as receipt data. The task's third acceptance criterion implies no. This
   is the single decision everything else hangs off.
2. **Is inferable removal of dangling and invalid references acceptable**, or must every
   removal be operator-named? The recommendation is inferable with a verbatim receipt; the
   conservative alternative is that `--auto` covers only duplicates, self-edges and empty
   legacy keys.
3. **Should repair load Threads under the guard and refuse when one is malformed**, as
   `MutateTaskLifecycle` does for lifecycle writes? Correctness-first says yes; it also
   means a broken Thread document blocks a graph repair, which may be the wrong trade for a
   recovery command.
4. **Is the minimality check a hard rejection or an override?** Hard rejection is
   recommended — the follow-up removal is an ordinary `task depend remove` once the graph
   is mutable — but it will occasionally require two commands where the operator wanted one.
5. **Does the `TaskGraphLoadProblem` source token land as a prerequisite task or inside
   this one?** It is a port change with its own test, and writing the concurrency test
   first is the strongest sequencing.
6. **Is the `μ` tuple frozen wire vocabulary or an implementation detail?** Exposing the
   six named components lets an agent verify monotonicity itself; it also freezes the
   defect taxonomy against future problem codes.
7. **Manifest threshold.** Always available, or required above some operation count? The
   recommendation is always available and never required.

## Candidate tasks

- ⏳ `tskflwctl task new "Add a source revision token to task graph load problems" --epic 30-threads-and-task-dependency-graphs --tags graph,storage` — H6: close the CAS blind spot before repair depends on it.
- ⏳ `tskflwctl task new "Model graph-owned declarations as a source-level projection" --epic 30-threads-and-task-dependency-graphs --tags graph` — H2/H5/L1/M6: the `DeclarationRef` model plus a sound simulator, underneath everything else.
- ⚠️ Re-scope `6g4g8gatbnrs` against H1, H3, H4 and M2 before estimating: removal-only, declaration-keyed, improvement-not-health, and progress and preservation as two separate proofs.
- ⏳ `tskflwctl task new "Make graph health remedies specific to the defect class" --epic 30-threads-and-task-dependency-graphs --tags cli,graph` — M3: independently shippable, and it removes a live mis-routing today.
- ⏳ `tskflwctl task new "Report Thread projection impact from dependency repair" --epic 30-threads-and-task-dependency-graphs --tags threads,graph` — M1, if decision 3 lands as yes.
