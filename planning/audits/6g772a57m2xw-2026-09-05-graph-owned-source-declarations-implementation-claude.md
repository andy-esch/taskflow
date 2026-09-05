---
schema: 1
id: 6g772a57m2xw
bucket: closed
area: graph-owned-source-declarations-implementation-claude
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: Graph-owned source declarations implementation — claude — 2026-09-05

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

| Item | Value |
| --- | --- |
| Sandbox path | `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.cucmDT` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.cucmDT/.git` |
| Sandbox baseline commit | `3556386c743008e0d4d1d55ea8021fa37ffb71f7` — `chore: capture review sandbox baseline` (the only commit created) |
| Captured source-audit blob | `9db7a56185ac061ba3185f8fd8b561710683ddfa` |
| Guarded transfer | Succeeded — the pre-transfer `git hash-object` of the source audit still equalled the captured blob, and the post-copy `cmp -s` was clean |

The sandbox is an independent `git clone --no-hardlinks` overlaid with `rsync -a --delete --exclude='.git'`,
so it carries the staged, unstaged, untracked and deleted state of the handoff checkout. `.git` is a real
directory and `.git/objects/info/alternates` is absent, so no object store is shared with
`$SOURCE_ROOT`. Every build, test, lint, generator, fixture, mutation probe and report edit ran inside
the sandbox. `$SOURCE_ROOT` saw exactly two operations: reading this brief, and the final guarded
single-file copy. No branch switch, stage, restore, clean, stash, reset, or write-capable project
command was run there. All probe files were deleted and `git checkout -- .` restored the tree, after
which `git status --short` and `git diff --stat HEAD` were both empty and
`go test ./internal/core/ ./internal/store/` was green again; at transfer time `git status --short`
lists only the assigned audit.

### Verdict

**The core model is sound in its load-bearing claims, and its central regression tests genuinely
pin them** — every mutation I aimed at raw-duplicate loss, representative-map simulation,
unreadable-evidence loss, duplicate-shadow CAS loss, all-at-once legacy clearing, dedupe retry
instability, and a coordinated constructor downgrade was killed by the intended test for the
intended reason. Ordering is adapter-order independent, immutability holds at every seam I attacked,
`SameSourceSnapshot` is a correct fail-closed multiset comparison with no revision leakage, and
source identity fails closed (`ErrAmbiguous`) in every under-specified case I could construct.

**Six findings survive.** Two are high: `ProjectedEdge` asserts an edge for declarations the graph
projects nowhere — provably so for a non-representative duplicate-ID record, which is the exact
state a repair command exists to fix; and `TaskGraph` now carries two silently different
source-fidelity classes, because the four `taskGraphFromMap` reconstruction sites build graphs whose
`sourceTasks` is the representative map and whose unreadable set is empty, with no marker and no
failure. The remainder are contract gaps that the current suite does not pin at all: two behavioural
mutations — implicit legacy-key removal, and outright slug-to-ID guessing in the edited-field write
path — pass `go test ./...` in full.

None of these break a shipping consumer today; all of them are load-bearing for the repair command
this slice exists to enable, and H2 in particular reintroduces the false-accept class the task
claims to have eliminated.

### Consumer inventory

Every non-test consumer of the touched surface, and whether the new full-source behaviour is
compatible.

| Consumer | Site | Uses | Compatible? |
| --- | --- | --- | --- |
| `store.verifyTaskGraphSourceSnapshot` | `internal/store/cas.go:36` | `SameSourceSnapshot` | **Yes.** Both sides come from `core.LoadTaskGraph`, i.e. a complete adapter read. |
| graph mutation CAS | `internal/store/graphmutation.go:87` | `SameSourceSnapshot` | **Yes.** `expected` and `current` are both full `LoadTaskGraph` scans. |
| thread mutation / creation / apply / lifecycle CAS | `threadmutation.go:92`, `threadcreation.go:87`, `threadapply.go:123`, `lifecyclemutation.go:110` | `SameSourceSnapshot` | **Yes.** Same full-scan pairing. New shadow-record comparison is a strict superset; `ValidateTaskGraphMutationSource` already refuses the broken snapshots where shadows can exist. |
| `taskGraphFromMap` | `dependency_graph_mutation.go:126` | constructor | **No — see H2.** Feeds one representative per ID and `nil` problems into `newTaskGraph`, so the new `sourceTasks` is silently lossy. |
| `taskGraphWithStatus` / `taskGraphWithTask` | `task_lifecycle.go:371,377` | `Task()`, `TaskIDs()`, `taskGraphFromMap` | **No — see H2.** Also inherits `Task()`'s cleared `SourceVersion`. |
| `taskGraphAfterDependencyPlan` | `task_lifecycle.go:397` | `Task()`, `TaskIDs()`, `taskGraphFromMap` | **No — see H2.** |
| `graphAfterTaskWrites` | `thread_apply.go:484` | `Task()`, `TaskIDs()`, `taskGraphFromMap` | **No — see H2.** |
| prefix/final validation graphs | `dependency_graph_mutation.go:92,100` | `taskGraphFromMap` | Health-only use today, so behaviour-neutral; same latent trap as H2. |
| `NewTaskGraphRead` | `board.go:57`, `service.go:296`, `service_task.go:54,192`, `dependency_source.go:204` | constructor | **Yes.** All are complete adapter reads or the simulator's own full clone. |
| `Prerequisites` | `thread_graph.go`, `thread_projection.go`, `service.go:586`, blockers/state/impacts | projection | **Yes.** Unchanged semantics; see L2 for its newly written doc description. |
| `Task()` / `TaskIDs()` | lifecycle, thread apply, mutation planners, CLI/TUI renderers | representative projection | **Yes.** Unchanged; `SourceVersion` still cleared. |
| `SourceRecords` / `SourceDeclarations` / `CanonicalDependencies` / `SimulateSourceEdits` | none outside tests | new | No shipping consumer yet — which is why the contract must be right now. |

### Evidence matrix

Mutation probes: each defect was reintroduced in the sandbox, the suite run, and the sandbox restored.

| # | Reintroduced defect | Killed by | Verdict |
| --- | --- | --- | --- |
| M1 | `sourceFieldsForTask` dedupes raw values | `…SourceViewsSeparateRawCanonicalAndProjected…`, `…PreservesRawInvalidAndDangling…`, `…DedupeIsRetryStable…`, `store/…LosslessDependencySourceProjection` | Killed, intended reason |
| M2 | `SimulateSourceEdits` rebuilds from `g.tasks` | `…SimulationRetainsDuplicateAndUnreadableRecords` | Killed, intended reason |
| M3 | Simulation drops `loadProblems` | `…SimulationRetainsDuplicateAndUnreadableRecords` | Killed, intended reason |
| M4 | `SameSourceSnapshot` reverted to representative-only | `…SnapshotCASIncludesEveryDuplicateIDRecord` | Killed, intended reason |
| M5 | `setTaskSourceField` clears all three legacy fields | `…SimulationRemovesOnlyNamedLegacyState` | Killed, intended reason |
| M6 | **Slug guessing**: non-ID literals in an edited field rewritten to `id.New()` | *nothing* | **SURVIVED `go test ./...`** → M2 finding |
| M7 | `dedupe` made occurrence-sensitive | `…DedupeIsRetryStableAndExactDropNamesAnOccurrence` | Killed, intended reason |
| M8 | `projectedSourceEdge` always returns no edge | `…SourceViews…`, `…ProjectEveryLegacyCycleDirection` | Killed (true-branch only) |
| M9 | `projectedSourceEdge` always fabricates `{Value → Source.TaskID}` | same two tests (wrong `blocks` direction) | Killed for legacy direction only; **no test asserts `HasProjectedEdge == false`** → H1 |
| M10 | **Last-value drop keeps the present legacy key** | *nothing* | **SURVIVED `go test ./...`** → M1 finding |
| M11 | Coordinated: constructor stores only representatives *and* fills `sourceTasks` from them | `…SimulationRetainsDuplicateAndUnreadableRecords`, `…SnapshotCASIncludesEveryDuplicateIDRecord` | Killed — the *constructor* path is pinned; `taskGraphFromMap` reaches the same state unpinned → H2 |

Hostile fixtures constructed and exercised: duplicate canonical literals; duplicate legacy literals in
all three directions; invalid, empty-string, whitespace, dangling, self and unreadable-target values;
duplicate IDs with distinct, identical and missing locations; a fully indistinguishable twin pair;
missing frontmatter IDs; ID/filename drift; present-but-empty legacy keys; legacy-induced cycles in
`blocked_by`/`dependencies`/`blocks`; pathless remote records; and mixed readable/unreadable
repositories.

---

### Findings

#### H1. ProjectedEdge asserts edges the graph projects nowhere, most sharply for non-representative duplicate-ID records · **Status:** fixed

**Reproduction.** Two records collide on stable ID `X`. `newTaskGraph` sorts by
`(canonicalID, Path, Slug)` and keeps the first as representative
(`dependency_graph.go:351-361,393-396`); the loser contributes *no* dependency state, because both
`g.dependencies[X]` and the canonical edge list are guarded by `isRepresentative`
(`dependency_graph.go:406-414,445-448`). `projectedSourceEdge` applies no such guard: for the
canonical field it returns an edge whenever `id.Valid(Source.TaskID) && id.Valid(Value)`
(`dependency_source.go:377-383`). Sandbox probe, two records with ID `ev7w12f8cht3`:

```
src=ev7w12f8cht3|probe-shadow-a|tasks/ev7w12f8cht3-probe-shadow-a.md  value="kd2qr678y1yf"
    hasEdge=true edge={From:kd2qr678y1yf To:ev7w12f8cht3}
    inOutgoing=false  inDependencies=false        <-- claimed edge exists in NEITHER projection
representative path for ev7w12f8cht3 = tasks/1c5v76cx4nm8-probe-shadow-b.md
```

The shadow's declaration is attributed to the representative's semantic node and claims an edge the
graph carries in neither `outgoing` (the DAG) nor `dependencies` (the `Prerequisites` projection).
The same probe shows the second, weaker class: a canonical `depends_on` value that is dangling or
names an unreadable target also reports `HasProjectedEdge=true` while being absent from `outgoing`
(2 of 4 declarations mismatched). That branch is also **internally inconsistent with the legacy
branch**, which resolves through `g.legacy` and correctly returns `false` for a `LegacyMissing`
reference — so the identical defect class answers differently depending only on which field declared
it.

**Violated contract.** The type's own comment ("ProjectedEdge is zero when an invalid or unresolved
value cannot name an edge", `dependency_source.go:49-53`) and the brief's requirement that a
projection labelled "edge" match what `TaskGraph` actually uses, specifically for dangling targets,
unreadable targets and non-representative duplicate records.

**Why it matters downstream.** The next slice is a repair command that must prove progress. A
planner that reads `HasProjectedEdge` to answer "does dropping this declaration remove an edge?" or
"which declarations feed cycle C?" gets a false yes for every shadow-record and dangling declaration
— exactly the broken states repair operates on, since duplicate IDs are themselves a
`ProblemDuplicateTaskID`.

**Smallest sound recommendation.** Resolve the canonical branch the way the legacy branch already
resolves: only claim an edge when the declaration's owning record is the representative for its
canonical ID *and* the value resolves to a node the graph projects. Keeping the raw value in
`Value` already preserves the evidence, so nothing is lost by refusing to name an edge. If instead
the intent is "the edge this value would name if repaired", rename the field
(`IntendedEdge`) and document that it is not a graph edge — but the current name plus
`HasProjectedEdge` reads as the former.

**Affected files/consumers.** `internal/core/dependency_source.go:49-61,103-123,377-396`;
`internal/core/dependency_graph.go:393-396,406-414`. No shipping consumer yet; the guarded repair
command is the intended one.

**Missing regression test.** There is no assertion anywhere that `HasProjectedEdge == false`.
`assertSourceDeclaration` only ever asserts a *present* edge, which is why mutation M9 (fabricate an
edge for every declaration) is caught only incidentally, by the `blocks` direction. Add a table
asserting `HasProjectedEdge == false` for: a shadow record's declaration, a dangling canonical ID, an
unreadable-target canonical ID, and a legacy missing reference — plus a property test that every
`ProjectedEdge` is contained in the graph's own projected edge set.

---

**Resolution:** Projected-edge attribution now requires a uniquely identifiable
representative source and an existing semantic target; regressions cover shadow,
dangling, unreadable, invalid, and legacy-missing declarations plus subset
containment in the semantic graph.

#### H2. `taskGraphFromMap` builds a `TaskGraph` whose source model is silently lossy, with no marker and no failure · **Status:** fixed

**Reproduction.** `newTaskGraph` now populates `sourceTasks` from whatever slice it is handed
(`dependency_graph.go:313`). `taskGraphFromMap` (`dependency_graph_mutation.go:126-132`) hands it one
representative per ID, obtained from `graph.Task()` — which also clears `SourceVersion`
(`dependency_graph.go:816`) — and `nil` unreadable problems. Sandbox probe over a repository with one
duplicate-ID shadow and one unreadable record:

```
full graph:                          records=3  loadProblems=1  health=broken
taskGraphWithStatus rebuild:         records=2  loadProblems=0  health=healthy
taskGraphAfterDependencyPlan rebuild: records=2 loadProblems=0
graphAfterTaskWrites rebuild:         records=2 loadProblems=0
simulating on a rebuilt graph keeps  records=2  loadProblems=0   (evidence silently gone)
Task() projection SourceVersion=""  (rebuilt graphs inherit this)
```

A `*TaskGraph` that answers `SourceRecords`, `SourceDeclarations` and `SimulateSourceEdits` with
confidence has lost a physical record and an unreadable diagnostic, and reports the repository
**healthy when it is broken**. There is no type, flag or error distinguishing it from a
full-fidelity graph, and `SimulateSourceEdits` returns the same `*TaskGraph` type, so the two
classes are freely interchangeable at every call site.

**Violated contract.** ADR-0006 as amended by this change: "Existing representative-map graph
reconstruction remains valid only behind ordinary mutation's healthy-source precondition and is
**forbidden** for broken-graph repair simulation." The code states the rule in prose but provides no
mechanism to enforce it. It is also the precise false-accept class acceptance criterion 4 claims is
demonstrated unable to recur — the constructor route is genuinely pinned (coordinated mutation M11
is killed), but the reconstruction route reaches the same state untested.

**Smallest sound recommendation.** Give the lossy class a marker rather than relying on discipline.
One private boolean (e.g. `derived bool`, set by `taskGraphFromMap` / cleared by `NewTaskGraphRead`)
and a fail-closed guard returning `ErrValidation` from `SourceRecords`, `SourceDeclarations` and
`SimulateSourceEdits` on a derived graph. That costs ~10 lines, changes no shipping behaviour
(no current consumer calls those from a rebuilt graph), and converts a silent wrong answer into a
loud one before the repair command is written against it. `SameSourceSnapshot` already fails closed
here for a different reason — `Task()` clears `SourceVersion` — but that is accidental protection,
not a contract.

**Affected files/consumers.** `internal/core/dependency_graph_mutation.go:92,100,126`;
`internal/core/task_lifecycle.go:371,377,394,397,418`; `internal/core/thread_apply.go:484,496`;
`internal/core/dependency_graph.go:313,816`.

**Missing regression test.** No test asserts what a reconstructed graph does with the source views.
Add one that builds a full graph containing a duplicate-ID shadow plus an unreadable record, passes
it through `taskGraphAfterDependencyPlan`, and requires the source queries to fail closed rather
than answer from the representative map.

---

**Resolution:** Representative-map graph reconstructions are now explicitly
source-incomplete. Source records, source declarations, simulation, and
whole-snapshot equality fail closed, with a regression through the existing
prospective graph helper.

#### M1. Dropping a legacy field's last value silently performs `drop-empty-field` as well, and the same intent is rejected when batched · **Status:** fixed

**Reproduction.** `SimulateSourceEdits` sets `keepPresent = false` whenever a field had values and
has none left (`dependency_source.go:195-197`), so a single `drop` also deletes the key from
`LegacyDependencyFields`:

```
before fields = [{Field:blocked_by Values:[ha3tdykf6bsv]}]
after  fields = []
LegacyDependencyFields now=[]
```

The operator asked for one declaration drop and got a second, distinct edit that the vocabulary
already models separately as `TaskGraphSourceDropEmptyField`. Worse, the two spellings of the same
intent disagree:

```
ONE batch  {drop, drop-empty-field} -> err=validation failed: graph-owned field blocked_by is not empty
TWO batches drop then drop-empty-field -> ok  (health degraded -> healthy -> healthy)
```

because the `dropEmpty` guard at `dependency_source.go:190-192` inspects the *pre-edit* values.

**Violated contract.** ADR-0006 as amended: the simulator "changes only selected graph-owned
declarations". Also the brief's "It must never … clear unrelated values/fields", and the model's own
deliberate preservation of "empty legacy field presence through `SourceRecords`" — that evidence is
destroyed by an edit that never named it.

**Smallest sound recommendation.** Keep `keepPresent = present` for legacy fields on a value drop and
require the explicit `drop-empty-field` edit to remove the key; then make the same-batch pair legal
by evaluating the `dropEmpty` guard against `remaining` rather than `values`. If the implicit removal
is instead deliberate (it does let a degraded snapshot converge in one step), say so in the type
comment and the ADR and pin it — right now the code, the ADR and the tests disagree.

**Affected files/consumers.** `internal/core/dependency_source.go:144-147,188-202`; the future
materializer, which must decide between writing `blocked_by: []` and emitting
`domain.UnsetField{}` (`internal/store/frontmatter.go:139-149`).

**Missing regression test.** Mutation M10 — deleting the `len(values) > 0 && len(remaining) == 0`
clause so the key survives as present-but-empty — passes the entire repository suite. Add a test
pinning the chosen semantics for a last-value drop, and one asserting that `{drop, drop-empty-field}`
in one batch behaves the same as the two-batch sequence.

---

**Resolution:** Dropping a final legacy value now retains a non-nil
present-but-empty field. Explicit empty-field cleanup works in the same batch or
a later batch with equivalent results.

#### M2. Nothing pins the "never guess a replacement ID" acceptance criterion in the write path · **Status:** fixed

**Reproduction.** Sandbox mutation adding, at the top of `setTaskSourceField`
(`dependency_source.go:309-311`):

```go
for i, v := range values {
    if !id.Valid(v) && v != "" { values[i] = id.New() }
}
```

Every non-ID literal in an edited field is replaced by a freshly minted stable ID:

```
after dropping the empty literal: [{Field:depends_on Values:[6m77xysvx2fy 6g775ge3nkv8]}]
                                   (the surviving "  " literal became a fabricated ID)
```

`go test ./...` is fully green under this mutation.

**Violated contract.** Acceptance criterion 2 of task `6g721vewvvrz`: "Invalid and dangling raw values
survive verbatim for diagnosis and receipts; **the model never guesses a replacement ID from a
slug-shaped value**", and the brief's requirement that the simulator "must never … guess a
slug-to-ID replacement". The shipped code is correct; the criterion is simply unproven, and this is
the one claim whose violation would silently fabricate dependency edges into a repaired file.

**Why the existing tests miss it.** `TestTaskGraphSourceSimulationPreservesRawInvalidAndDanglingIntent`
drops both non-ID literals, so nothing malformed ever reaches `setTaskSourceField`; the assertion is
about what was removed, not about what survived the write path unchanged.

**Smallest sound recommendation.** No production change. Add the missing assertion.

**Affected files/consumers.** `internal/core/dependency_source.go:309-331`;
`internal/core/dependency_source_test.go:59-84`.

**Missing regression test.** Simulate an edit on a record whose field retains at least one invalid
literal (e.g. `depends_on: [<valid-id>, "human-authored-slug", ""]`, dropping only the empty string)
and assert the surviving raw values are byte-identical, including the slug-shaped and whitespace
literals. Do the same for a legacy field carrying slug values, which is the direction where a
slug→ID rewrite is most tempting.

---

**Resolution:** Regression tests now preserve untouched canonical and legacy
invalid literals verbatim while other declarations in those fields are removed.

#### M3. `SimulateSourceEdits` returns no per-edit outcome, so an unmatched edit is indistinguishable from an applied one · **Status:** wontfix

**Reproduction.** Four edits that match nothing — an absent source, an undeclared value, an
out-of-range occurrence, and a dedupe on an absent field — return a graph and `nil`:

```
4 wholly unmatched edits -> err=<nil>
identical to origin: true (caller gets no per-edit outcome and no error)
```

An absent source is skipped at `dependency_source.go:156-157`; an absent value or occurrence simply
never matches in `applySourceDeclarationEdits` (`:220-241`). The signature is
`([]TaskGraphSourceEdit) (*TaskGraph, error)`, so the only way for a caller to learn which edits took
effect is to diff `SourceRecords()` before and after and re-derive the attribution the simulator
already computed.

**Violated contract.** Not a stated invariant — convergent no-ops for missing targets are explicitly
intended — but it collides with the sequencing note in the task: "the downstream task owns policy,
authorization, materialization, **receipts**". A receipt and a progress proof both need per-edit
outcomes, and the fail-closed posture here is also inconsistent within the same function:
`drop-empty-field` against a non-empty field is a hard `ErrValidation`, while `drop` against a
non-existent occurrence is silent.

**Smallest sound recommendation.** Return the applied set alongside the graph — e.g.
`SimulateSourceEdits(edits) (*TaskGraph, []TaskGraphSourceEditOutcome, error)` with one
`{Edit, Applied bool, RemovedOccurrences int}` per input edit. It is pure information the function
already has, it keeps missing targets convergent, and it lets the repair command distinguish "no-op,
already converged" from "your selector matched nothing" without a second projection pass. Doing it
now avoids a signature change once the repair port exists.

**Affected files/consumers.** `internal/core/dependency_source.go:128-207`; the future repair
receipt/progress consumer.

**Missing regression test.** None can exist for behaviour the API cannot express. Once outcomes are
returned, pin: absent source, absent value, out-of-range occurrence, dedupe on an already-unique
value, and a duplicate edit entry that resolves to the same removal.

---

**Resolution:** The simulator intentionally treats missing selections as
convergent no-ops and proves aggregate source effects. The downstream repair
planner validates selected intent, while the materializer owns actual
per-operation outcomes and receipts; architecture and ADR text now state that
boundary.

#### L1. A present-but-empty legacy field yields `Values == nil`, so field presence is carried only by the entry's existence · **Status:** fixed

**Reproduction.** `sourceFieldsForTask` builds `Values: append([]string(nil), item.values...)`
(`dependency_source.go:289`), which is `nil` — not `[]string{}` — when the legacy list is empty:

```
present-but-empty blocks: Values==nil? true len=0 (presence carried only by the entry existing)
```

`internal/store/dependency_source_test.go:44` asserts `{Field: blocks, Values: []string{}}` but
compares with `slices.Equal`, which treats `nil` and `[]string{}` as equal — so the test reads as if
it pins an empty non-nil slice and does not.

**Violated contract.** None outright: presence *is* faithfully represented, by the entry appearing in
`Fields`. The risk is that the distinction the model exists to preserve is encoded in the container
rather than in the value, and the one test that looks like it pins the representation does not.
A materializer keying on `len(f.Values) == 0` to emit `domain.UnsetField{}` would delete a key the
model says is present.

**Smallest sound recommendation.** Either normalise to a non-nil empty slice and assert it with
`reflect.DeepEqual` (or `f.Values != nil`), or add an explicit `Present bool` to
`TaskGraphSourceField`. One line either way; the type comment already claims present-but-empty
fields "remain observable", so make the observation unambiguous.

**Affected files/consumers.** `internal/core/dependency_source.go:34-39,272-292`;
`internal/store/dependency_source_test.go:40-50`.

**Missing regression test.** An assertion that distinguishes nil from empty for a present-but-empty
legacy field, and one that a present-but-empty field survives an unrelated edit to the same record.

---

**Resolution:** Present-but-empty legacy fields now expose an explicit non-nil
empty values slice, pinned through both core preservation and filesystem adapter
tests.

#### L2. The new `Prerequisites` documentation understates what the projection contains · **Status:** fixed

**Reproduction.** Both the ADR paragraph and the `CanonicalDependencies` comment added by this change
describe `Prerequisites` as "the canonical-plus-safely-resolved-legacy projection". It is not
filtered: `g.dependencies[taskID] = sortedUnique(dependencies)` copies every raw `depends_on` value
including malformed ones (`dependency_graph.go:406-408`). Probe:

```
prerequisites(owner) = [1pgmz9tqjebw 7r7dfy23vyzn not-an-id vrccqkqvc9kh]
                                      ^^^^^^^^^ not a stable task id at all
```

so `Prerequisites` returns non-IDs, dangling IDs and unreadable IDs, while the newly added
`CanonicalDependencies` filters them via `id.Valid`.

**Violated contract.** Documentation only — the runtime behaviour is pre-existing and fails closed
(`computeSound` marks an unknown prerequisite broken, so the gate is `GateBroken`). But this change
is precisely the one that tells future consumers to "choose the view explicitly rather than
reconstructing source intent from semantic edges", and it describes one of the three views
inaccurately.

**Smallest sound recommendation.** One sentence in `dependency_graph.go:951-955` and in the ADR
paragraph: `Prerequisites` is the fail-closed behaviour projection and deliberately retains invalid
and unresolvable literals so gating stays broken; `CanonicalDependencies` is the stable-ID-only view.

**Affected files/consumers.** `internal/core/dependency_graph.go:951-975`;
`planning/adrs/0006-adopt-threads-as-task-dags.md` (added paragraph); `docs/ARCHITECTURE.md` (added
paragraph).

**Missing regression test.** An assertion that `Prerequisites` retains an invalid literal while
`CanonicalDependencies` omits it, pinning the documented difference between the two views.

---

**Resolution:** Code, ADR, and architecture now describe Prerequisites as the
fail-closed behavior view retaining all canonical literals plus resolved legacy
values; a regression distinguishes it from stable-ID-only CanonicalDependencies.

### Settled hostile angles

Each of these was attacked and is **not** a finding; the evidence is recorded so the conclusion is
checkable.

- **Raw multiset losslessness.** Duplicates, invalid tokens, empty-string and whitespace literals,
  dangling IDs and legacy ownership all survive verbatim through `SourceRecords`,
  `SourceDeclarations` and a round trip through `SimulateSourceEdits`. Mutation M1 is killed by four
  separate tests including the filesystem one.
- **Occurrence identity.** `SourceDeclarations` numbers occurrences per `(record, field)` in source
  order (`dependency_source.go:107-113`) and `applySourceDeclarationEdits` numbers them identically
  (`:222-230`), so a declaration's `Occurrence` is exactly the selector `drop` consumes. Record-level
  sorting in `SourceRecords` cannot perturb it because numbering is within a field.
- **Batch order independence.** Resolution runs to completion before any mutation, and group keys are
  applied in a sorted `(taskIndex, fieldOrder)` order (`:177-186`). Forward and reversed batches of
  `{drop occ1, dedupe}` produce identical sources; `dedupe`-twice, duplicate identical drop entries,
  and `drop+dedupe` of the same value are all deterministic. No batch ordering deletes a different
  declaration than selected.
- **Retry properties — stated explicitly.** `dedupe` is idempotent and occurrence-independent
  (verified: re-running it on its own output is a byte-identical no-op). `drop` is deliberately
  **not** idempotent: re-running `drop(value, 0)` against the already-edited state removes a second
  equal value (`[b,a,a] → [b,a]`). That is the documented trade (`dependency_source.go:74-83`), and
  retry safety is correctly delegated upward — `SimulateSourceEdits` clears `SourceVersion` on every
  changed record (`:202`), so a repair command holding whole-snapshot CAS cannot replay a `drop`
  against a shifted source. Callers must not treat `drop` as retry-safe on its own.
- **Source identity fails closed.** Empty `TaskGraphSourceRef` fields act as wildcards, which sounds
  dangerous but cannot mis-target: a record always matches its own reference, so a unique match is
  necessarily the intended record. ID-only selection across three colliding records returns
  `ErrAmbiguous`; a fully indistinguishable twin pair returns `ErrAmbiguous`; a wholly empty ref
  returns `ErrValidation`; an absent target is a silent convergent no-op with no retargeting; a
  location-qualified reference changes exactly one physical record and leaves its shadow untouched.
- **Whole-snapshot CAS.** `sameReadableTaskSources` sorts both sides by exactly the tuple it then
  compares (`taskSourceSnapshotKey` vs `sourceRefForTask` + `SourceVersion`), so it is a correct
  multiset comparison, not a content-sorted heuristic — no false equality or false conflict when
  records collide. Verified: reordered records equal; a changed shadow revision, a removed shadow, an
  added shadow, a changed location, a pathless variant and a changed unreadable revision all detected;
  a missing revision on either side, on both sides, or on an unreadable record fails closed; no
  revision token is exposed on any public type. `SimulateSourceEdits` clears revisions only on
  changed records, so an empty simulation correctly compares equal and any real edit does not.
- **Legacy attribution, all three facts.** `blocks` reverses ownership and the source view keeps the
  declaration on the declaring record while `ProjectedEdge` points `From` that record
  (`dependency_source.go:388-393`, verified in all three legacy cycle directions). Duplicate legacy
  values are collapsed by `sortedUnique` in `resolveLegacyDiagnostics`
  (`dependency_graph.go:623`) yet `sourceFieldsForTask` preserves both occurrences raw. Present-but-empty
  keys appear in `SourceRecords` with no declarations (subject to L1).
- **Immutability and aliasing.** Constructor input (task slices, nested dependency slices, and the
  `Problems` slice), query output (`SourceRecords` fields and value slices), simulation input
  (the `edits` slice is never written), and the origin graph after simulation were all mutated or
  inspected post-hoc; nothing leaked. `SimulateSourceEdits` clones once and `NewTaskGraphRead` clones
  again, so the simulated graph does not alias its origin.
- **Race behaviour.** 32 goroutines concurrently running `SourceRecords`, `SourceDeclarations`,
  `CanonicalDependencies`, `Prerequisites`, `BlockingFrontier` (which writes the mutex-guarded caches)
  and `SimulateSourceEdits` on one graph are clean under `-race`, as is the full suite.
- **Core purity.** The public source vocabulary is strings, typed core enums and `DependencyEdge`;
  no `os`, `yaml`, `cobra`, TUI or renderer type crosses the seam, and `SourceVersion` never appears
  on a public type. `golangci-lint` is clean.
- **Materialization feasibility.** `updateFrontmatter` (`internal/store/frontmatter.go:116-152`) edits
  a `yaml.Node` mapping in place, so unknown fields, comments and key order survive, and
  `domain.UnsetField{}` removes a key — the `drop`/`dedupe`/`drop-empty-field` vocabulary maps onto
  it surgically without reconstructing unrelated frontmatter. One prerequisite the downstream slice
  must not overlook: today's writer addresses files through `s.resolvePath(taskID)`
  (`graphmutation.go:134`), which goes through `resolveID` and returns `ErrAmbiguous` for exactly the
  duplicate-ID case repair targets. `TaskGraphSourceRef.Location` already carries the needed physical
  address, so this is a new path-addressed store primitive, not a core-vocabulary gap.

### Residual risks (not findings)

- **Pathless duplicates are unaddressable.** Two records sharing an ID and slug where one has no
  location cannot be disambiguated, because an empty `Location` is a wildcard rather than a match on
  emptiness. It fails closed with `ErrAmbiguous`, and no filesystem repository can produce the state,
  but a future remote adapter could. Worth an explicit note in the type comment.
- **`CanonicalDependencies` returns `nil` both for an unknown task and for a known task with no
  canonical dependencies**, unlike `Task()`'s `(value, ok)`. Consistent with `Prerequisites`, so not a
  regression, but a repair planner asking "does this task declare anything?" gets an ambiguous answer.
- **`SourceRecords` ordering is keyed partly on `SourceVersion`** (`taskSourceSortKey`,
  `dependency_source.go:357-363`), so the public record order depends on a store-private token. I
  could not make this observable: the only records that tie on the public part of the key are fully
  indistinguishable ones, whose public projections are identical (verified by swapping revisions and
  comparing the public order). Declaration order is unaffected, since
  `sourceDeclarationSortKey` excludes the revision. Low risk, but it means the ordering contract is
  not purely a function of public data.
- **`setTaskSourceField` sorts `LegacyDependencyFields`** (`:329`) on any touched record. Harmless
  today — the slice is only ever read through `slices.Contains` — but it is a normalisation of state
  the model otherwise treats as verbatim.
- **`SimulateSourceEdits` on a `nil` receiver returns `ErrValidation`** but every other method
  (`SourceRecords`, `SourceDeclarations`, `SameSourceSnapshot`) would panic on `nil`. Inconsistent,
  though no call site can produce it.
- **The `drop`/`dedupe` write rewrites the whole target list** in the eventual materialization, so
  any inline YAML comment inside a `depends_on` sequence would be lost. Unrelated keys are safe.
  Out of scope here, but the downstream writer should be tested for it.

### Validation results

All commands run inside `$SANDBOX` at baseline `3556386`, with probe files removed.

| Command | Result |
| --- | --- |
| `go test ./internal/core/ -run 'TestTaskGraphSource' -count=1 -v` | PASS — 7 tests, 7 subtests, `ok … 0.205s` |
| `go test ./internal/store/ -run 'TestReadTaskGraphPopulatesLosslessDependencySourceProjection' -count=1 -v` | PASS — `ok … 0.161s` |
| `go test -race ./...` | PASS — every package `ok`, exit 0 |
| `just lint` (`golangci-lint run ./...`) | `0 issues.` |
| `just tidy-check` (`go mod tidy -diff`) | clean, exit 0 |
| `just docs-check` (`docgen` + `git diff --exit-code docs/cli`) | clean, exit 0 |
| `just build` then `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| `git diff --check` | clean, exit 0 |
| Post-restore `go test ./internal/core/ ./internal/store/ -count=1` | PASS — `ok … 0.607s`, `ok … 3.554s` |
| Post-restore `git status --short` / `git diff --stat HEAD` | both empty |

Findings are left `open`; status triage belongs to the implementation owner.
