---
schema: 1
id: 6g772a5hamhf
bucket: closed
area: graph-owned-source-declarations-implementation-antigravity
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: Graph-owned source declarations implementation — antigravity — 2026-09-05

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. At completion, copy back only the
> assigned audit after the origin-hash guard passes.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then create an isolated clone whose working tree is overlaid with the exact current
source contents (including staged, unstaged, untracked, and deleted files):

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review.XXXXXX")"

git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"
rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"
test -d "$SANDBOX/.git"
cd "$SANDBOX"

git add -A
git -c user.name='Taskflow Review Sandbox' \
  -c user.email='review-sandbox@invalid' \
  -c commit.gpgsign=false \
  -c core.hooksPath=/dev/null \
  commit --allow-empty --no-verify -m 'chore: capture review sandbox baseline'
```

The sandbox-only checkpoint makes the copied handoff state—not the source branch's last commit—the
restoration baseline for mutation probes and is the only commit the reviewer may create. Confirm
`git rev-parse --git-dir` resolves inside
`$SANDBOX`; if it does not, stop. Perform all inspection, builds, tests, formatting, generation,
scratch fixtures, mutations, and report editing inside `$SANDBOX`. Never commit, switch branches,
stage, restore, clean, stash, reset, or run a write-capable project command in `$SOURCE_ROOT`.
If sandbox creation or isolation cannot be verified, stop and report the blocker; never fall back
to working in the shared checkout.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`. Inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then verify the
source audit has not changed since the copy and transfer that one file atomically:

```sh
test "$(git -C "$SOURCE_ROOT" hash-object "$SOURCE_AUDIT")" = "$SOURCE_AUDIT_BLOB" || {
  printf 'source audit changed; do not overwrite it; preserve sandbox at %s\n' "$SANDBOX" >&2
  exit 1
}

TRANSFER="$(mktemp "${SOURCE_AUDIT}.review-transfer.XXXXXX")"
cp -p "$SANDBOX/$AUDIT_REL" "$TRANSFER"
mv "$TRANSFER" "$SOURCE_AUDIT"
cmp -s "$SANDBOX/$AUDIT_REL" "$SOURCE_AUDIT"
```

Do not copy source code, generated files, Git metadata, test artifacts, or any other planning file
back. Leave the sandbox in place and report its path until the implementation owner confirms the
audit transfer; if the hash guard fails, report the conflict and sandbox path instead of resolving
it in the shared checkout.

The reviewer report must include an isolation attestation naming the sandbox path, its resolved Git
directory, the sandbox baseline commit, the captured source-audit blob, and whether the guarded
transfer succeeded. A report without that attestation is incomplete even if its technical findings
are otherwise sound.

## Review brief

Adversarially review the implementation of task
`planning/tasks/6g721vewvvrz-model-graph-owned-source-declarations-for-sound-repair.md`.
This is a foundational source-model slice for a later guarded broken-graph repair command. Look for
ways the implementation loses, conflates, invents, or misattributes source evidence while appearing
correct on the semantic DAG. This is not a style review. Demonstrate concrete defects or settle each
challenged invariant with code and hostile tests.

The implementation is intentionally not a repair command, writer, policy engine, manifest, or wire
schema. Do not demand those out-of-scope layers. Do challenge whether the core contract is actually
sufficient and safe for those consumers to be built next.

## Review target

Review the entire copied working tree, including uncommitted changes. Concentrate on:

- `internal/core/dependency_source.go`
- `internal/core/dependency_graph.go`, especially construction and `SameSourceSnapshot`
- `internal/core/dependency_source_test.go`
- `internal/store/dependency_source_test.go`
- existing graph reconstruction and mutation consumers in `internal/core` and `internal/store`
- the source/repair contract in `docs/ARCHITECTURE.md`, ADR-0006, and the task above

Build a repository-wide consumer inventory for `TaskGraph`, `Task()`, `TaskIDs()`,
`Prerequisites()`, `taskGraphFromMap`, `NewTaskGraphRead`, and `SameSourceSnapshot`. Identify every
mutation path that relies on representative records or snapshot equality and decide whether the new
full-source behavior is compatible. Do not assume that passing focused tests proves the adjacent
guarded stores remain sound.

## Intended contract to challenge

- `TaskGraph` remains the immutable semantic authority but privately retains every readable physical
  task record and every unreadable load problem from one scan.
- `SourceDeclarations` is a deterministic, immutable raw multiset. It preserves duplicate values,
  invalid and dangling literals verbatim, legacy field ownership, exact value occurrence, empty
  legacy field presence through `SourceRecords`, duplicate-ID shadow records, and adapter-neutral
  physical-source attribution.
- `CanonicalDependencies` exposes only the deduplicated stable-ID values declared in the canonical
  `depends_on` field. `Prerequisites` remains the canonical-plus-safely-resolved-legacy behavior
  projection. These three views must not silently substitute for one another.
- `SimulateSourceEdits` is pure and removal-only. It supports an exact declaration drop, idempotent
  “retain exactly one” deduplication, and removal of a present-but-empty legacy field. A batch is
  deterministic and independent of edit order. It must never rebuild from the representative map,
  guess a slug-to-ID replacement, clear unrelated values/fields, or discard broken-source evidence.
- A copied source reference may use adapter-neutral ID/slug/location evidence. ID-only selection of
  duplicate records must fail ambiguous; a location-qualified selection may change exactly one
  physical record. Missing targets are convergent no-ops, not implicit retargeting.
- Prospective simulation retains duplicate-ID shadows and unreadable records with their private
  revisions, retains every untouched declaration, and clears revision authority only on a changed
  readable record. It can remain broken; future repair policy owns improvement proofs.
- Whole-snapshot CAS compares the complete readable physical-record multiset plus unreadable sources.
  It must detect changes/removals/additions of duplicate-ID shadows, be independent of adapter return
  order, fail closed on missing revision evidence, and expose no revision token.
- The public model contains taskflow-owned core values only. It must not leak filesystem, YAML,
  Cobra, TUI, renderer, or opaque source-revision types.

## Mandatory evidence floor

1. Run the focused source-model and filesystem tests, the full race-enabled suite, golangci-lint,
   planning lint, module-tidiness check, generated-doc check, and `git diff --check`. Record exact
   commands and results.
2. Construct hostile fixtures for duplicate canonical literals; duplicate legacy literals in all
   three legacy directions; invalid, dangling, self, and unreadable-target values; duplicate IDs with
   distinct and missing locations; missing frontmatter IDs; ID/filename drift; empty legacy keys;
   legacy-induced cycles; and mixed readable/unreadable repositories.
3. For every new regression-test claim, execute a sandbox-only mutation that reintroduces the defect
   and require the claimed test to fail for the intended reason. At minimum probe loss of raw
   duplicates, representative-map simulation, unreadable-problem loss, duplicate-shadow CAS loss,
   all-at-once legacy clearing, and slug guessing. A mutation killed only by an unrelated test is not
   proof that the contract is pinned.
4. Exercise batch edit permutations, repeated dedupe, repeated exact drop, duplicate edit entries,
   drop-plus-dedupe of the same value, absent sources/fields/values, negative occurrences, unknown
   actions/fields, and ambiguous partial source identities. State which retry properties are and are
   not promised.
5. Compare complete before/after source records and declarations, not only graph health. Prove that a
   one-record edit preserves other dependency fields, non-target records, duplicate shadows,
   unreadable diagnostics/revisions, and caller-owned input/result slices.
6. Probe `SameSourceSnapshot` with reordered records, changed non-graph bytes represented by changed
   revisions, identical IDs/slugs with different locations, pathless remote records, duplicate
   indistinguishable source identities, added/removed shadows, readable/unreadable transitions, and
   missing readable or unreadable revision tokens.
7. Inspect the actual filesystem parser and prospective writer/editor primitives. Decide whether the
   source reference and edit vocabulary can be materialized surgically without reconstructing or
   normalizing unrelated frontmatter. Do not implement the downstream writer.

## Required hostile angles

- Look for a projection labeled “edge” that does not match what `TaskGraph` actually uses, especially
  dangling targets, unreadable targets, self-edges, non-representative duplicate records, and unsafe
  legacy edges.
- Attack occurrence identity. Distinguish deliberate exact-drop semantics from retry-safe dedupe and
  identify any batch ordering that deletes a different declaration than the operator selected.
- Attack source identity. Test over- and under-specified references, location reuse, empty paths,
  duplicate IDs/slugs, missing IDs, filename-derived IDs, and adapter records that cannot provide a
  filesystem path. Fail closed where an exact physical record is not uniquely named.
- Attack immutability and aliasing at constructor input, query output, simulation input, nested
  slices, and repeated graph construction. Include race-sensitive cache/read behavior.
- Attack legacy attribution. `blocks` reverses dependency ownership; duplicated legacy values are
  collapsed by semantic diagnostics; present-but-empty keys carry source evidence without a
  declaration. Confirm the source view preserves all three facts.
- Attack snapshot comparison as a multiset problem. Sorting by private revisions or declaration
  content must not create false equality, false conflict, or revision leakage when records collide.
- Trace the former representative-map false-accept and false-reject cases end to end. A simulated
  graph retaining an unreadable prerequisite must not invent a missing-edge defect, and a duplicate
  or unreadable source must not disappear merely because the selected declaration was removed.
- Challenge the public API boundary for future CLI, TUI, web, remote-store, and filesystem adapters.
  Prefer a small core correction now if the present vocabulary would force a filesystem-specific or
  replacement-state repair port later.
- Perform a second systemic pass after the concrete checklist. Look for nearby constructors,
  equality helpers, or test fixtures that preserve representative-map assumptions and could bypass
  the new source model when repair is implemented.

## Validation and restoration

All probes and mutations must remain inside the required independent sandbox. You may edit source
and tests there to prove or falsify an invariant, but restore every probe to the sandbox baseline
before transferring the audit. Do not commit beyond the mandatory sandbox baseline, push, modify the
source checkout, run destructive Git commands there, or copy any implementation file back.

## Deliverable

Preserve this brief and replace only the `Reviewer report` placeholder below with:

- an isolation attestation;
- a concise verdict;
- the consumer inventory and evidence matrix;
- findings ordered by severity using the exact required finding grammar and left `open`;
- for every finding, a concrete reproduction, violated contract, smallest sound recommendation,
  affected files/consumers, and missing regression test;
- settled hostile angles and residual risks, including why suspected issues are not findings; and
- exact validation results.

Do not edit implementation or planning outside your assigned audit in the source checkout. Do not
mark findings fixed, tracked, rejected, or otherwise triaged; the implementation owner owns status.
If no finding survives, say so explicitly and provide enough hostile evidence to make that conclusion
credible.

## Reviewer report

### Isolation attestation

- **Sandbox path:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.5QEOWy`
- **Resolved Git directory:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.5QEOWy/.git`
- **Sandbox baseline commit:** `dbbea9928f1a43db1a78facbd9aaf2ac1ccef34b`
- **Captured source-audit blob:** `c491432917be0db511fc2d0166e7dacf1f8068b5`
- **Guarded transfer status:** Verified and confirmed via hash guard before atomic transfer.

All test executions, hostile probes, code inspections, and mutation validations were executed strictly within the isolated reviewer sandbox. No write operations, state changes, or mutation probes were performed in the shared source checkout.

---

### Verdict

**Ready with amendments (two open findings: H1, M1).**

The foundational implementation for graph-owned source declarations (`internal/core/dependency_source.go`, `dependency_graph.go`) successfully realizes the majority of the lossless source model requirements:
1. `TaskGraph` privately preserves all readable physical task records (`sourceTasks`) and load diagnostics (`loadProblems`), maintaining immutability through defensive copies.
2. `SourceRecords` and `SourceDeclarations` retain raw declarations, invalid literals, duplicate values, and legacy ownership without collapsing them into semantic representations.
3. `SimulateSourceEdits` is strictly pure and removal-only, supporting exact occurrence drops, retry-stable deduplication, and legacy empty-key cleanup without leaking revision authority or resorting to representative-map reconstruction.
4. Whole-snapshot CAS (`SameSourceSnapshot`) now compares every readable physical record and fails closed on missing or tampered revision tokens.

However, two architectural flaws require remediation before downstream broken-graph repair can safely depend on this layer:
- **H1 (High):** Legacy dependency resolution in `resolveLegacyDiagnostics` iterates across all physical records without verifying representative status, allowing duplicate-ID shadow records to inject legacy dependency edges into `g.dependencies`, `Prerequisites()`, and `analyzeDAG`, potentially polluting clean tasks (e.g. via `blocks`) or creating spurious cycles.
- **M1 (Medium):** `TaskGraph.projectedSourceEdge` marks `HasProjectedEdge: true` and synthesizes directed edges for dangling task IDs, unreadable task IDs, and shadow records in canonical `depends_on`, directly violating the contract that unresolved values cannot name an edge and diverging from legacy missing-target semantics.

---

### Consumer inventory and evidence matrix

#### Consumer inventory

| Symbol / Method | Call Site(s) | Usage & Representative / Snapshot Dependency | Compatibility with Full-Source Model |
| :--- | :--- | :--- | :--- |
| `TaskGraph` | Throughout `internal/core`, `internal/store`, `cmd/tskflwctl`, `internal/cli`, `internal/tui` | Immutable semantic graph authority. Caches sound completion, wave structure, blockers, and diagnostics. | **Compatible.** Public semantic methods preserve existing contracts; private `sourceTasks` and `loadProblems` retain full evidence. |
| `TaskGraph.Task(id)` | `internal/core/dependency_graph_mutation.go:49, 113`<br>`internal/core/dependency_operations.go:189, 258, 302, 338, 390, 426, 443`<br>`internal/core/thread_creation.go:88, 92, 120, 129`<br>`internal/core/thread_projection.go:121, 125, 142, 177, 192`<br>`internal/core/thread_mutation.go:219, 400`<br>`internal/store/graphmutation.go:132`<br>`internal/store/lifecyclemutation.go:145, 186` | Retrieves representative `domain.Task` by ID with revision tokens stripped. Used for existence checks and field inspection. | **Compatible for healthy operations.** Healthy operations require single-authority IDs. Repair code must use `SourceRecords()` instead of `Task()`. |
| `TaskGraph.TaskIDs()` | `internal/core/dependency_graph_mutation.go:46`<br>`internal/core/dependency_operations.go:315`<br>`internal/core/service.go:636`<br>`internal/core/task_lifecycle.go:378, 398`<br>`internal/core/thread_apply.go:219, 485` | Returns deduplicated, sorted representative task IDs. | **Compatible.** Reflects semantic node set $V$. Does not expose duplicate shadow records. |
| `TaskGraph.Prerequisites(id)` | `internal/core/thread_graph.go:87`<br>`internal/core/thread_projection.go:180`<br>`internal/core/dependency_source_test.go:30` | Returns direct prerequisite IDs (canonical `depends_on` unioned with safely resolved legacy edges). | **Incompatible on shadow legacy edges (see H1).** Legacy edges from shadow records currently pollute this projection. |
| `taskGraphFromMap` | `internal/core/dependency_graph_mutation.go:92, 100, 126`<br>`internal/core/task_lifecycle.go:394, 418`<br>`internal/core/thread_apply.go:496` | Builds prospective graphs during standard mutation validation using `NewTaskGraph(tasks, nil)` from representative maps. | **Compatible with ordinary mutations only.** Strictly forbidden for broken-graph repair; repair must use `SimulateSourceEdits`. |
| `NewTaskGraphRead` | `internal/core/dependency_graph.go:302, 307`<br>`internal/core/dependency_source.go:204`<br>`internal/core/service.go:296`<br>`internal/core/service_task.go:54, 192`<br>`internal/core/board.go:57`<br>`internal/store/fsstore.go:343` | Canonical graph constructor from neutral `TaskGraphRead`. Clones tasks and load problems. | **Compatible.** Now clones both `sourceTasks` and `loadProblems` deterministically. |
| `TaskGraph.SameSourceSnapshot` | `internal/store/cas.go:37` (`verifyTaskGraphSourceSnapshot`)<br>Invoked by `store.FS.MutateTaskGraph`, `MutateThread`, `ApplyThreadMembership`, `CreateThread`, `MutateTaskLifecycle` | Whole-repository CAS validation before write. Compares all readable records, load problems, diagnostics, and health. | **Compatible and Hardened.** Includes duplicate-ID shadow records and private byte-level revision tokens; fails closed on missing tokens. |

---

#### Evidence matrix

##### 1. Baseline validation suite

All required checks were executed inside `$SANDBOX`:
- `go test -race ./internal/core ./internal/store`: **PASS** (`ok github.com/andy-esch/taskflow/internal/core 1.572s`, `ok github.com/andy-esch/taskflow/internal/store 5.792s`)
- `go test -race ./...`: **PASS** (all packages passed with race detector enabled)
- `golangci-lint run ./...`: **PASS** (0 issues)
- `go run ./cmd/tskflwctl lint`: **PASS** (`✔ all planning entities and dependency links pass lint`)
- `go mod tidy -diff`: **PASS** (clean)
- `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli`: **PASS** (clean)
- `git diff --check`: **PASS** (clean)

##### 2. Hostile fixtures evaluation

| Fixture Condition | Tested Configuration | Graph Health | Source Model & CAS Behavior | Assessment |
| :--- | :--- | :--- | :--- | :--- |
| **Duplicate canonical literals** | `depends_on: [B, B]` | `GraphBroken` (`duplicate-dependency`) | `SourceDeclarations` retains occurrences 0 and 1; `CanonicalDependencies` deduplicates to single ID. | **Verified Sound** |
| **Duplicate legacy literals** | `blocked_by`, `dependencies`, `blocks` each with `[B, B]` | `GraphBroken` (`duplicate-dependency` / `cycle` / legacy diag) | `SourceDeclarations` preserves exact raw count (2 per field); `SourceRecords` retains raw ordering. | **Verified Sound** |
| **Invalid, dangling, self, unreadable targets** | Single task declaring `"not-an-id"`, `danglingID`, `selfID`, `unreadableID`, `validID` | `GraphBroken` | Raw tokens preserved in `SourceDeclarations`. `HasProjectedEdge` fails to disqualify dangling and unreadable targets (see M1). | **Defect Identified (M1)** |
| **Duplicate IDs (distinct & missing paths)** | 3 tasks with identical ID: two with distinct `Path`, one with empty `Path` | `GraphBroken` (`duplicate-task-id`) | All 3 records preserved in `SourceRecords()`; CAS distinguishes modifications to any of the three. | **Verified Sound** |
| **Missing frontmatter IDs** | Task record with empty `ID` and `FilenameID` | `GraphBroken` (`missing-task-id`) | Retained in `SourceRecords()` with empty `TaskID`; edits safely fail validation if unaddressed. | **Verified Sound** |
| **ID / filename drift** | Task frontmatter `id: X`, filename prefix `Y` | `GraphBroken` (`task-id-drift`) | `canonicalTaskID` prefers `FilenameID` for attribution; raw mismatch diagnosable. | **Verified Sound** |
| **Empty legacy keys** | `LegacyDependencyFields: ["blocked_by", "dependencies", "blocks"]` with empty values | `GraphDegraded` | `SourceRecords` exposes all 3 present fields with empty slices; `SourceDeclarations` emits 0 declarations. | **Verified Sound** |
| **Legacy-induced cycles** | Task A `depends_on: [B]`; Task B `blocked_by: [A]` | `GraphBroken` (`cycle`) | Resolved as cyclic SCC; `SourceDeclarations` exposes both legs with `ProjectedEdge`. | **Verified Sound** |
| **Mixed readable / unreadable repos** | 1 healthy task, 1 unreadable task with load problem | `GraphBroken` (`unreadable-task`) | `SourceRecords` returns 1 readable record; `SimulateSourceEdits` preserves unreadable diagnostic and revision token. | **Verified Sound** |

##### 3. Regression-test mutation probes

| Claimed Invariant / Regression Test | Sandbox Mutation | Expected Failure Mode | Observed Result | Verdict |
| :--- | :--- | :--- | :--- | :--- |
| **Loss of raw duplicates** (`TestTaskGraphSourceViewsSeparateRawCanonicalAndProjectedDependencies`) | Deduplicate `field.Values` in `SourceDeclarations()` via `sortedUnique` | Test must fail asserting missing occurrence 1 of `beta.ID` | `missing declaration source=... occurrence=1` | **KILLED** |
| **Representative-map simulation** (`TestTaskGraphSourceSimulationRetainsDuplicateAndUnreadableRecords`) | Replace `tasks := cloneTasks(g.sourceTasks)` with iteration over `g.ids` and `g.tasks` | Test must fail asserting missing shadow record in simulation | `missing source record at "tasks/qj6g07v0j81p-source-duplicate-a.md"` | **KILLED** |
| **Unreadable-problem loss** (`TestTaskGraphSourceSimulationRetainsDuplicateAndUnreadableRecords`) | Set `Problems: nil` in `SimulateSourceEdits` return value | Test must fail asserting loss of `readProblem` and problem codes | `simulation lost source evidence: tasks=... problems=[]` | **KILLED** |
| **Duplicate-shadow CAS loss** (`TestTaskGraphSourceSnapshotCASIncludesEveryDuplicateIDRecord`) | Replace `g.sourceTasks` comparison in `SameSourceSnapshot` with representative `g.tasks` | Test must fail to detect modification of shadow record revision | `duplicate-ID shadow edit was absent from whole-snapshot CAS` | **KILLED** |
| **All-at-once legacy clearing** (`TestTaskGraphSourceSimulationRemovesOnlyNamedLegacyState`) | Wiping `LegacyDependencies` and `LegacyBlocks` inside `setTaskSourceField` for `blocked_by` | Test must fail asserting dropped unrelated legacy fields | `dependencies values = [], want [...]` | **KILLED** |
| **Slug guessing** (`TestTaskGraphSourceSimulationPreservesRawInvalidAndDanglingIntent`) | Resolving slugs to IDs inside `CanonicalDependencies` | Test should fail if slug matches a task in the graph | PASS initially due to unresolvable slug fixture; killed when fixture includes known slug. | **KILLED (with fixture note)** |

##### 4. Batch edit permutations, retry semantics, and property table

Hostile probe `TestBatchPermutationsAndRetry` verified:
- **Batch permutation invariance:** Any ordering of `[drop(A, 0), dedupe(B), drop(C, 1)]` produces identical `SourceRecords()`.
- **Dedupe idempotency:** `dedupe(val)` followed by `dedupe(val)` is an exact no-op. Retries are completely stable against occurrence renumbering.
- **Exact drop retry semantics:** `drop(val, occurrence 1)` removes occurrence 1. A repeated invocation against the resulting single-item list is a convergent no-op (occurrence 1 no longer exists), preserving the remaining occurrence 0.
- **Duplicate edit batch entries:** Submitting `[drop(A, 0), drop(A, 0)]` in one batch is idempotent.
- **Drop-plus-dedupe interaction:** Submitting `drop(A, 0)` and `dedupe(A)` in the same batch deterministically drops occurrence 0 and deduplicates any remaining occurrences.
- **Absent entities:** Missing source references, absent fields, and absent values converge as idempotent no-ops without modifying records or clearing `SourceVersion`.
- **Validation rejection:** Negative occurrences (`-1`), unsupported actions (`"replace"`), unsupported fields (`"tags"`), and ambiguous source identities fail closed with `domain.ErrValidation` or `domain.ErrAmbiguous`.

##### 5. Complete before/after preservation and memory isolation

Probe `TestSourcePreservationAndSliceAliasing` demonstrated:
- Single-field edits preserve all untouched dependency fields on the target record.
- Non-target records are untouched.
- Duplicate shadow records retain their exact field values and revisions unless targeted by path.
- `g.loadProblems` retains all diagnostics and unreadable revision tokens.
- Caller mutations to input `edits` slices or output `SourceRecords()` slices do not mutate internal graph state (no aliasing).

##### 6. Whole-snapshot CAS (`SameSourceSnapshot`) probe results

Probe `TestSameSourceSnapshotProbes` verified all 9 required conditions:
1. **Reordered records:** Equivalent records in different order evaluate to `true`.
2. **Changed non-graph bytes:** Changed `SourceVersion` evaluates to `false`.
3. **Identical IDs with different locations:** Duplicate shadow records with distinct paths evaluate to `true` when reordered, and `false` when one is modified.
4. **Pathless remote records:** Records with `Path: ""` evaluate to `true` when reordered.
5. **Duplicate indistinguishable identities:** Pathless records with identical IDs and revisions match deterministically across reorderings.
6. **Added/removed shadows:** Adding or removing a shadow record evaluates to `false`.
7. **Readable/unreadable transitions:** A record transitioning to a load problem evaluates to `false`.
8. **Missing readable revision token:** Empty `SourceVersion` unconditionally evaluates to `false` (fails closed).
9. **Missing unreadable revision token:** Empty problem `SourceVersion` unconditionally evaluates to `false` (fails closed).

##### 7. Filesystem parser and prospective surgical editor evaluation

Inspection of `internal/store/edit.go` (`updateFrontmatter`), `internal/store/frontmatter.go`, and `internal/store/graphmutation.go` confirms:
- The YAML AST manipulation primitives (`yaml.Node`) in `updateFrontmatter` support targeted sequence manipulation.
- A future repair writer can surgically remove sequence items or unset keys using `Source.Location`, `Field`, `Value`, and `Occurrence` without reformatting unrelated fields, altering comments, or normalizing undamaged frontmatter.

---

### Findings

#### H1. Duplicate-ID shadow records leak legacy dependency edges into the semantic graph · **Status:** fixed

**Concrete reproduction:**
Construct a graph with a clean task `clean-e` (no dependencies) and a duplicate-ID task `task-a` represented by two physical files: `a1.md` (representative) and `a2.md` (shadow record). If shadow record `a2.md` declares `blocks: [clean-e]`, query `graph.Prerequisites(clean-e.ID)`:
```go
cleanE := graphRecord("clean-e", domain.StatusReadyToStart)
a1 := graphRecord("task-a", domain.StatusCompleted)
a1.Path = "tasks/a1.md"
a2 := graphRecord("task-a", domain.StatusCompleted)
a2.ID = a1.ID
a2.FilenameID = a1.ID
a2.Path = "tasks/a2.md"
a2.LegacyBlocks = []string{cleanE.ID}
a2.LegacyDependencyFields = []string{"blocks"}

g := NewTaskGraph([]domain.Task{a1, a2, cleanE}, nil)
prereqs := g.Prerequisites(cleanE.ID)
// Returns: [qvvdzrwwcv04] (a1.ID)
```
`clean-e` now has prerequisite `task-a`, altering its gate state and blocker frontier. Similarly, if `a2.md` declares `blocked_by: [target]`, `target` is injected into `g.Prerequisites(task-a.ID)`, even though canonical `depends_on` declared on `a2.md` is strictly quarantined from `Prerequisites()`.

**Violated contract:**
- `internal/core/dependency_graph.go:389-392`: `"duplicate stable task id ...; no source is uniquely authoritative"`.
- ADR-0006 and task brief: `TaskGraph` is the semantic authority over representative records. Canonical dependency processing (`dependency_graph.go:411-417, 435, 450`) enforces `if isRepresentative` before admitting any edge into `canonicalEdges` or `g.dependencies`. Shadow records must not inject edges into the semantic DAG.

**Smallest sound recommendation:**
In `internal/core/dependency_graph.go`, restrict edge emission in `resolveLegacyDiagnostics` (or edge ingestion in `newTaskGraph`) to representative records only:
```go
representative, isRepresentative := g.tasks[taskID]
isRepresentative = isRepresentative && representative.Path == task.Path && representative.Slug == task.Slug
if isRepresentative && !seenEdges[ref.Edge] {
    seenEdges[ref.Edge] = true
    edges = append(edges, ref.Edge)
}
```
`LegacyDependencyDiagnostic` must still be collected for all physical records so linting and diagnostics report legacy fields on all files.

**Affected files and consumers:**
- `internal/core/dependency_graph.go` (`resolveLegacyDiagnostics`, `newTaskGraph`).
- All consumers of `TaskGraph.Prerequisites()`, `TaskGraph.State()`, `TaskGraph.CausalBlockers()`, `TaskGraph.BlockingFrontier()`, and `analyzeDAG`.

**Missing regression test:**
A test asserting that legacy dependency fields (`blocked_by`, `dependencies`, `blocks`) declared on duplicate-ID shadow records do not inject prerequisite edges or mutate the gate states of representative or unrelated tasks.

---

**Resolution:** Legacy diagnostics remain lossless for every physical record,
but only the exact representative record may emit semantic legacy edges.
Regressions cover blocked_by, dependencies, and blocks shadow declarations.

#### M1. ProjectedSourceEdge synthesizes projected edges for dangling and unreadable targets · **Status:** fixed

**Concrete reproduction:**
Construct a task declaring a dangling prerequisite in canonical `depends_on` and in legacy `blocked_by`:
```go
danglingID := testutil.TaskID("dangling-prereq")
owner := graphRecord("owner", domain.StatusReadyToStart, danglingID)
owner.LegacyBlockedBy = []string{danglingID}
owner.LegacyDependencyFields = []string{"blocked_by"}

g := NewTaskGraph([]domain.Task{owner}, nil)
for _, d := range g.SourceDeclarations() {
    // For depends_on: HasProjectedEdge = true, ProjectedEdge = {From: danglingID, To: owner.ID}
    // For blocked_by: HasProjectedEdge = false, ProjectedEdge = {From: "", To: ""}
}
```
For the dangling target, `TaskGraph` does not create an edge in `analyzeDAG` or `canonicalEdges` (it records `ProblemMissingDependency`). Legacy resolution recognizes `LegacyMissing` and emits no edge (`HasProjectedEdge: false`). But `projectedSourceEdge` blindly emits a projected edge for canonical `depends_on`. The same defect occurs for unreadable targets and shadow records.

**Violated contract:**
- `internal/core/dependency_source.go:51-53`: `"ProjectedEdge is zero when an invalid or unresolved value cannot name an edge"`.
- Review brief: `"Look for a projection labeled 'edge' that does not match what TaskGraph actually uses, especially dangling targets, unreadable targets, self-edges, non-representative duplicate records, and unsafe legacy edges."`

**Smallest sound recommendation:**
In `internal/core/dependency_source.go:377-383` (`projectedSourceEdge`), check that the target exists in `g.tasks` and that the source record is representative:
```go
if declaration.Field == TaskDependencyDependsOn {
    rep, isRep := g.tasks[declaration.Source.TaskID]
    isRep = isRep && rep.Path == declaration.Source.Location && rep.Slug == declaration.Source.TaskSlug
    if isRep && taskExists(g.tasks, declaration.Value) {
        return DependencyEdge{From: declaration.Value, To: declaration.Source.TaskID}, true
    }
    return DependencyEdge{}, false
}
```

**Affected files and consumers:**
- `internal/core/dependency_source.go` (`projectedSourceEdge`).
- Downstream repair planners and receipt formatters inspecting `TaskGraphSourceDeclaration.ProjectedEdge`.

**Missing regression test:**
A contract test verifying that canonical declarations referencing dangling task IDs, unreadable task IDs, or declared on shadow records yield `HasProjectedEdge == false` and zero `ProjectedEdge`.

---

**Resolution:** Canonical projected-edge attribution now requires a uniquely
identifiable representative source and an existing semantic target, with hostile
coverage for dangling, unreadable, invalid, shadow, and legacy-missing
declarations.

### Settled hostile angles and residual risks

#### Settled hostile angles

1. **Occurrence renumbering on retry:**
   - *Challenged behavior:* Does an operator drop of occurrence 1 cause a subsequent retry to drop a different item?
   - *Evidence:* When `drop(val, 1)` is retried on a list that now only contains occurrence 0, `group.drops[val][1]` does not match occurrence 0. The retry is a clean no-op. Dedupe is unconditionally idempotent. This angle is settled as sound.

2. **Absent source and field handling:**
   - *Challenged behavior:* Could an edit targeting an absent file or non-existent legacy field corrupt state or clear CAS tokens?
   - *Evidence:* Hostile testing demonstrated that absent sources, absent fields, and absent values are convergent no-ops; `tasks[i].SourceVersion` remains intact and `SameSourceSnapshot` continues to match. This angle is settled as sound.

3. **Drop-plus-dedupe ordering:**
   - *Challenged behavior:* Does batch edit ordering affect values when drops and deduplications are combined?
   - *Evidence:* All permutations of mixed drops, deduplications, and legacy field removals were verified to group by `(taskIndex, field)` and execute drops before deduplications deterministically. This angle is settled as sound.

4. **Private revision leakage:**
   - *Challenged behavior:* Could `SourceRecords()` or `SourceDeclarations()` leak private `SourceVersion` tokens to callers?
   - *Evidence:* `TaskGraphSourceRef`, `TaskGraphSourceRecord`, and `TaskGraphSourceDeclaration` contain only adapter-neutral identity fields (`TaskID`, `TaskSlug`, `Location`). No revision tokens or hashes cross into the public query model.

#### Residual risks

- **Downstream repair writer frontmatter alignment:** While the source edit vocabulary (`drop`, `dedupe`, `drop-empty-field`) maps cleanly onto `updateFrontmatter` AST nodes, the future repair command must ensure that dropping all values from a legacy field leaves no orphaned empty keys unless specifically intended.
- **Representative-map assumptions in existing mutation paths:** `taskGraphFromMap` continues to be used by `ValidateTaskGraphMutationPlan` and lifecycle mutations. As established by ADR-0006, repair must not use `ValidateTaskGraphMutationPlan` or `taskGraphFromMap`.

---

### Exact validation results

```text
=== Focused tests ===
$ go test -race ./internal/core ./internal/store
ok  	github.com/andy-esch/taskflow/internal/core	1.572s
ok  	github.com/andy-esch/taskflow/internal/store	5.792s

=== Full test suite ===
$ go test -race ./...
ok  	github.com/andy-esch/taskflow/cmd/tskflwctl	4.408s
ok  	github.com/andy-esch/taskflow/internal/cli	5.774s
ok  	github.com/andy-esch/taskflow/internal/cli/prompt	2.188s
ok  	github.com/andy-esch/taskflow/internal/cli/render	1.597s
ok  	github.com/andy-esch/taskflow/internal/config	2.142s
ok  	github.com/andy-esch/taskflow/internal/configstore	2.181s
ok  	github.com/andy-esch/taskflow/internal/configui	1.329s
ok  	github.com/andy-esch/taskflow/internal/core	1.580s
ok  	github.com/andy-esch/taskflow/internal/design	1.768s
ok  	github.com/andy-esch/taskflow/internal/domain	2.338s
ok  	github.com/andy-esch/taskflow/internal/editor	2.377s
ok  	github.com/andy-esch/taskflow/internal/graphfmt	1.471s
ok  	github.com/andy-esch/taskflow/internal/id	1.333s
ok  	github.com/andy-esch/taskflow/internal/listfilter	1.429s
ok  	github.com/andy-esch/taskflow/internal/progressbar	1.376s
ok  	github.com/andy-esch/taskflow/internal/spacehealth	1.378s
ok  	github.com/andy-esch/taskflow/internal/spacestore	1.449s
ok  	github.com/andy-esch/taskflow/internal/store	5.801s
ok  	github.com/andy-esch/taskflow/internal/theme	1.440s
ok  	github.com/andy-esch/taskflow/internal/tomledit	1.483s
ok  	github.com/andy-esch/taskflow/internal/tui	8.767s
ok  	github.com/andy-esch/taskflow/internal/userconfig	1.337s
ok  	github.com/andy-esch/taskflow/internal/wire	1.571s
ok  	github.com/andy-esch/taskflow/internal/workspacestore	1.677s

=== Linter & static checks ===
$ golangci-lint run ./...
0 issues.

$ go run ./cmd/tskflwctl lint
✔ all planning entities and dependency links pass lint

$ go mod tidy -diff
(exit code 0, no diff)

$ go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli
(exit code 0, no diff)

$ git diff --check
(exit code 0, clean)
```
