---
schema: 1
id: 6gd1wee31vfx
bucket: closed
area: attribute-dependency-lint-diagnostics-implementation-antigravity
date: "2026-09-23"
updated_at: "2026-09-24"
---
# Audit: Portable dependency-lint record attribution implementation — Antigravity — 2026-09-23

> Reviewer assignment: Antigravity/Gemini. This document is the review brief and the only source
> file the reviewer may update.
>
> Play devil's advocate. Reconstruct the attribution flow independently, look for systemic failure
> modes and architectural leakage, and prefer a runnable counterexample or surviving mutation over
> broad approval. A no-findings verdict must be earned with the evidence described below.

## Mandatory isolated workspace

The implementation owner is actively using the handoff checkout. Treat it as read-only. Resolve
this audit, then create an independent sandbox before inspecting implementation, running tests, or
performing mutation probes:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_ABS="$(tskflwctl audit path 2026-09-23-attribute-dependency-lint-diagnostics-implementation-antigravity)"
AUDIT_REL="${AUDIT_ABS#"$SOURCE_ROOT"/}"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

Do all inspection, builds, tests, temporary fixtures, and mutations in the sandbox. Do not stage,
commit, switch branches, restore, clean, stash, reset, or run write-capable project commands in the
source checkout. Restore every probe in the sandbox so only the assigned audit differs, then verify
and transfer it:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Include the helper attestation and transfer result in the report. Preserve the sandbox until the
implementation owner confirms receipt.

## Review target

Adversarially review task `6gcwcf77tvgq`,
`attribute-dependency-lint-diagnostics-by-portable-task-record-identity`, on branch
`feat/portable-lint-graph-attribution` against `main`. Review the complete working-tree diff,
including the planning update. Do not implement fixes or edit any file other than this audit.

The implementation replaces path-keyed dependency-lint attribution with an opaque,
snapshot-local readable-record reference; carries that reference through graph problems, legacy
diagnostics, cycles, and representative lifecycle findings; clears it from public graph query
results; and stops core from parsing an ID-shaped location to recover unreadable-task identity.

## Claims to falsify

- Every graph-owned lint finding for a readable task is joined back to the exact `TaskWithBody`
  record that produced it, even when its path is empty, duplicated, opaque, or contradictory.
- Duplicate canonical task IDs do not merge findings. A defect declared by one duplicate record is
  not attributed to another, including when ID, slug, and location collide.
- The snapshot-local reference remains aligned when the analyzer sorts records internally, chooses
  a deterministic representative, emits SCC/cycle diagnostics, resolves legacy declarations, and
  derives lifecycle inconsistency.
- Unreadable records remain separate load diagnostics with portable ID/slug identity. Dependencies
  on a pathless unreadable record are still diagnosed, while core never parses `Path` to invent an
  identity.
- Filesystem-backed human and JSON lint output is unchanged except where the task deliberately
  enables previously dropped pathless findings. Graph health, severity, repair guidance, and
  guarded mutation behavior are unchanged.
- The correlation handle is private, ephemeral query plumbing: it does not leak into public graph
  DTOs, equality/CAS semantics, serialized output, persisted Markdown, or adapter contracts.

## Required independent work

1. Build a complete producer/consumer inventory for `GraphProblem`,
   `LegacyDependencyDiagnostic`, lifecycle-consistency findings, `TaskGraphRead`,
   `TaskGraphLoadProblem`, `TaskWithBody`, and `dependencyLintIssues`. Identify which findings are
   record-owned, global, representative-only, or deliberately rendered elsewhere.
2. Trace one readable record from `ReadLintTasks` through `Service.Lint`, graph construction,
   internal sorting, diagnostic production, and final `LintResult`. Prove the positional reference
   is scoped to the same immutable slice and cannot be joined against a reordered or re-read slice.
3. Exercise pathless cases for malformed/duplicate/missing dependencies, self-edges, cycles,
   resolved and unresolved legacy fields, and inconsistent in-progress/completed lifecycle states.
   Test duplicates with distinct slugs, identical slugs, identical empty locations, and the same
   non-empty opaque location; ensure defects do not cross records.
4. Exercise unreadable task records with: explicit identity and no location; explicit identity
   contradicting an ID-shaped location; location only; invalid identity; and filesystem-backed
   malformed YAML. Verify only the filesystem adapter owns filename parsing and that dependency
   blocker semantics remain correct.
5. Compare filesystem-backed `lint` human and JSON output against `main` using the same fixtures.
   Check ordering, labels, duplicate issue counts, severity, repair commands, and exit codes—not
   merely semantic containment.
6. Mutation-test at least these seams: restore path-keyed joining; replace the record key with task
   ID; assign references after internal sorting; omit the reference from cycle diagnostics; omit it
   from legacy diagnostics; omit representative lifecycle attribution; stop clearing it from public
   queries; and restore core location parsing. Name the focused test that kills each mutation or
   report the surviving gap.
7. Run focused tests, `go test -race ./...`, golangci-lint, repository planning lint, audit lint, and
   `git diff --check`. Use `/tmp` Go and golangci caches if host cache permissions require it.

## Required hostile angles

- Challenge the ordinal-reference design rather than assuming it is safe. Look for a copy, filter,
  sort, retry, secondary graph construction, or future helper that can silently break positional
  alignment.
- Check whether duplicate records that render the same `LintResult.Slug` are truly attributable to
  callers, or merely kept separate internally but indistinguishable in human/machine output.
- Look for graph problem constructors without a readable-record reference and for new problem codes
  that would be silently dropped by the `recordRef == 0` branch.
- Check whether public `Problems()`/`LegacyDiagnostics()` clearing is comprehensive and whether
  direct internal slice access creates a fragile special channel or mutable alias.
- Challenge deterministic behavior under every permutation of input records, especially exact
  duplicates whose sort keys otherwise tie and duplicate IDs whose representative owns cycle or
  lifecycle output.
- Verify local-path actionability was not weakened when identity parsing moved. Test malformed
  filenames and scanner errors, not only syntactically valid `<id>-<slug>.md` names.
- Look for overreach: changes to graph validity, repair authorization, lifecycle policy, source CAS,
  wire contracts, or adapters beyond what record attribution requires.

## Finding and evidence rules

Findings must be evidence-backed and use the exact repository grammar, preferably via
`tskflwctl audit finding new`: `#### H1. <title> · **Status:** open` (or M1/L1). Verify every named
symbol, file, test, and command. Distinguish implemented behavior from planning intent. Do not mark
findings tracked, settled, or done; the implementation owner will adjudicate them.

If no defect survives, provide a substantive settled verdict containing the producer/consumer
inventory, fixture matrix, hostile mutation table, exact validation commands/results, and isolation
attestation. “Tests pass” or a checklist without attempted falsification is not sufficient.

## Reviewer report

### Verdict & summary

**Verdict:** Adversarially verified with one low-severity coverage finding (L1).

Task `6gcwcf77tvgq` on branch `feat/portable-lint-graph-attribution` successfully decouples dependency lint diagnostics from local filesystem paths. The snapshot-local `taskGraphRecordRef` reliably attributes graph problems, cycles, legacy dependency diagnostics, and lifecycle-consistency findings to the originating readable task record without path dependency, even across duplicate canonical IDs, identical or opaque locations, and identical slugs. Core no longer parses filenames to recover unreadable task identity, cleanly delegating that responsibility to the filesystem store scanner while preserving full path actionability.

Mutation testing verified that 7 of 8 critical seams are killed by focused tests. However, mutation testing identified a surviving gap: `TaskGraph.LegacyDiagnostics()` clears `recordRef` at [`internal/core/dependency_graph.go:839`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L839), but no existing test fails if that clearing is omitted.

---

### Mandatory isolation attestation

```ini
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/.git
baseline_commit=4503a2aed46c185169c2e43d56579645b9c73e89
source_blob=e2d612501973f732b2cb5a95fbac4b8f70c3e8c9
source_fingerprint=b868037970da898cf860574c189ccdd8b98c3bc1
deliverable=planning/audits/6gd1wee31vfx-2026-09-23-attribute-dependency-lint-diagnostics-implementation-antigravity.md
```

#### Verification gates
- `go test -race ./...`: **PASS** (34 packages passed, 0 race conditions).
- `golangci-lint run ./...`: **PASS** (0 issues).
- `tskflwctl lint --json`: **PASS** (`{"schema_version":"1.68","unreadable":[],"issues":[]}`).
- `tskflwctl audit lint 2026-09-23-attribute-dependency-lint-diagnostics-implementation-antigravity`: **PASS** (`✔ all audit findings pass lint`).
- `git diff --check`: **PASS** (clean working tree).

---

### 1. Producer / consumer inventory

| Type / Finding | Producer | Consumers | Scope & Ownership | Rendering & Clearing Behavior |
| :--- | :--- | :--- | :--- | :--- |
| `GraphProblem` (structural: duplicate, self, invalid, missing dep) | `newTaskGraph` in [`internal/core/dependency_graph.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L464) | `dependencyLintIssues` via `graph.problems` | Record-owned (`recordRef = record.ref`) | Emitted as `depends_on` (or legacy field) issue on originating record; `recordRef` cleared in public `Problems()` |
| `GraphProblem` (duplicate task ID) | `newTaskGraph` in [`internal/core/dependency_graph.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L428) | `dependencyLintIssues` via `graph.problems` | Record-owned (`recordRef = record.ref`) | Emitted as `id` issue on each duplicate record; `recordRef` cleared in public `Problems()` |
| `GraphProblem` (cycle) | `newTaskGraph` in [`internal/core/dependency_graph.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L538) | `dependencyLintIssues` via `graph.problems` | Representative-only (`recordRef = g.representativeRecord[taskID]`) | Emitted as `depends_on` issue on representative record; `recordRef` cleared in public `Problems()` |
| `GraphProblem` (unreadable, missing ID, ID drift, invalid status) | `newTaskGraph` in [`internal/core/dependency_graph.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L375) | Public `Problems()` callers | Global or domain-lint owned | Skipped in `dependencyLintIssues` problem loop; unreadable rendered as `LintLoadProblem`, status/id via domain lint |
| `LegacyDependencyDiagnostic` | `resolveLegacyDiagnostics` in [`internal/core/dependency_graph.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L662) | `dependencyLintIssues` via `graph.legacy`, public `LegacyDiagnostics()` | Record-owned (`recordRef = record.ref`) | Emitted as grouped legacy field issue on originating record; `recordRef` cleared in public `LegacyDiagnostics()` |
| Lifecycle consistency findings | `dependencyLintIssues` in [`internal/core/service.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/service.go#L764) | Joined into `Service.Lint` results | Representative-only (`recordRef = graph.representativeRecord[taskID]`) | Emitted as `status` issue on representative record |
| `TaskGraphRead` | `TaskGraphReadFromFiles` or `Service.Lint` | `NewTaskGraphRead` in [`internal/core/dependency_graph.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L329) | Ephemeral transport DTO | Carries `Tasks []domain.Task` and `Problems []TaskGraphLoadProblem` into strict snapshot constructor |
| `TaskGraphLoadProblem` | `TaskGraphLoadProblemFromFile` in [`internal/core/service_task.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/service_task.go#L163) | `newTaskGraph` | Entity diagnostic | Treats `Path` as opaque repair context; copies `EntityID` and `EntitySlug` populated by scanner |
| `TaskWithBody` | `s.lintReads.ReadLintTasks()` in [`internal/core/service.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/service.go#L470) | `Service.Lint` | Readable record snapshot | Feeds acceptance criteria lint, domain lint, and `taskRecords` 1:1 positional indexing |
| `dependencyLintIssues` | `Service.Lint` in [`internal/core/service.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/service.go#L497) | `Service.Lint` task loop | Query plumbing | Maps `taskGraphRecordRef` to `[]domain.Issue`; joined back via `taskGraphRecordRefAt(taskIndex)` |

---

### 2. End-to-end record reference trace

A readable record traverses the pipeline through strictly scoped immutable slices:

1. **Read:** `s.lintReads.ReadLintTasks()` produces `tasks []TaskWithBody`.
2. **Projection:** `Service.Lint` creates `taskRecords := make([]domain.Task, len(tasks))` and shallow-copies `taskRecords[i] = tasks[i].Task`. No filtering, sorting, or reordering occurs. Position `i` corresponds directly to `tasks[i]`.
3. **Graph construction:** `TaskGraphRead{Tasks: taskRecords}` is passed to `NewTaskGraphRead` -> `newTaskGraph`.
4. **Correlation handle assignment:** Before any sorting, `newTaskGraph` initializes `ordered := make([]taskGraphRecord, len(tasks))` where `ordered[index] = taskGraphRecord{task: task, ref: taskGraphRecordRefAt(index)}`. For slice index `i`, `ref = i + 1`.
5. **Deterministic sort:** `sort.SliceStable(ordered, ...)` orders by `(canonicalTaskID, Path, Slug, ref)`. The `ref` is preserved inside each `taskGraphRecord`.
6. **Diagnostic generation:**
   - Record-owned structural defects (duplicate dep, self dep, missing dep, invalid ID, duplicate ID) and legacy diagnostics attach `record.ref`.
   - DAG cycle diagnostics and lifecycle consistency findings attach `g.representativeRecord[taskID]`, capturing the deterministic representative for that canonical ID.
7. **Query boundary:**
   - Package-internal `dependencyLintIssues(graph *TaskGraph)` reads `graph.problems` and `graph.legacy` directly to populate `map[taskGraphRecordRef][]domain.Issue`.
   - Public queries `graph.Problems()` and `graph.LegacyDiagnostics()` clear `out[i].recordRef = 0`, preventing leakage into callers.
8. **Final join:** In `Service.Lint`, the loop `for taskIndex, tb := range tasks` appends `graphIssues[taskGraphRecordRefAt(taskIndex)]`. Because `taskIndex` indexes the original `tasks` slice, the join precisely matches `ordered[index].ref = index + 1`.

Because `tasks` and `taskRecords` are allocated within the function stack of `Service.Lint`, the slice cannot be re-read or mutated concurrently during graph construction or joining.

---

### 3. Hostile fixture & defect matrix

Adversarial testing was performed across 18 pathless, duplicate, and unreadable permutations:

| Case | Fixture Configuration | Expected Behavior | Observed Result |
| :--- | :--- | :--- | :--- |
| **Pathless malformed dep** | `Path: ""`, `DependsOn: ["not-a-valid-id"]` | Diagnosed on record's `depends_on` | **PASS** — `"not a stable task id"` attributed to record |
| **Pathless duplicate dep** | `Path: ""`, `DependsOn: [tgt.ID, tgt.ID]` | Diagnosed on record; target clean | **PASS** — `"repeats dependency"` on record, target unaffected |
| **Pathless missing dep** | `Path: ""`, `DependsOn: ["6g9999999999"]` | Diagnosed on record's `depends_on` | **PASS** — `"depends on missing task"` attributed to record |
| **Pathless self-edge** | `Path: ""`, `DependsOn: [ownID]` | Diagnosed on record's `depends_on` | **PASS** — `"cannot depend on itself"` attributed to record |
| **Pathless cycle** | `A -> B -> A`, both `Path: ""` | Diagnosed on both cycle participants | **PASS** — `"dependency cycle"` attributed to both records |
| **Pathless resolved legacy** | `Path: ""`, `blocked_by: [existingSlug]` | Diagnosed on record's `blocked_by` | **PASS** — resolves to target ID |
| **Pathless unresolved legacy** | `Path: ""`, `blocked_by: ["nonexistent"]` | Diagnosed on record's `blocked_by` | **PASS** — `"has no exact task ID or slug match"` |
| **Pathless in-flight gate** | `Path: ""`, status `in-progress` on not-started | Diagnosed on dependent's `status` | **PASS** — `"blocked dependency gate"` on dependent |
| **Pathless completed gate** | `Path: ""`, status `completed` on in-flight | Diagnosed on dependent's `status` | **PASS** — `"persisted nominally-complete task has a blocked dependency gate"` |
| **Dup IDs, distinct slugs, empty paths** | Same ID, distinct slugs, both `Path: ""` | Duplicate ID on both; isolated dep defects | **PASS** — defects isolated, no cross-record leakage |
| **Dup IDs, distinct slugs, same opaque path** | Same ID, distinct slugs, both `Path: "opaque://same"` | Duplicate ID on both; isolated dep defects | **PASS** — defects isolated, no cross-record leakage |
| **Dup IDs, same slug, empty paths** | Same ID, same slug, both `Path: ""` | Distinct `LintResult` rows, isolated defects | **PASS** — exactly one row carries `bad-ref` defect |
| **Dup IDs, same slug, same opaque path** | Same ID, same slug, both `Path: "opaque://same"` | Distinct `LintResult` rows, isolated defects | **PASS** — exactly one row carries `bad-ref` defect |
| **Unreadable: explicit ID, no path** | `EntityID: id`, `Location: ""` | Load problem + lifecycle diagnosis | **PASS** — dependent diagnoses unreadable ID blocker |
| **Unreadable: explicit ID vs path ID** | `EntityID: idA`, `Location: "tasks/idB-slug.md"` | Explicit ID takes precedence over path | **PASS** — dependent diagnoses `idA`; `idB` not inferred |
| **Unreadable: location only** | `EntityID: ""`, `Location: "tasks/id-slug.md"` | Core does NOT parse location for identity | **PASS** — dependent diagnoses missing task ID |
| **Unreadable: invalid ID** | `EntityID: "invalid-id"`, `Location: ""` | Preserved in problems; does not break graph | **PASS** — load problem preserved without panic |
| **Filesystem malformed YAML** | File on disk with invalid YAML frontmatter | Scanner parses flatname ID; path preserved | **PASS** — actionable location and recovered ID rendered |

---

### 4. Filesystem-backed parity with `main`

The compiled binary from `main` (`/tmp/tskflwctl-main`) and the current branch (`/tmp/tskflwctl-feat`) were evaluated against:
1. The production taskflow repository.
2. A synthetic multi-defect fixture repository containing missing dependencies, self-dependencies, duplicate dependencies, cycles, legacy fields, duplicate IDs across files, unreadable files, and lifecycle gate violations.

**Results:**
- **Machine contract (`lint --json`):** `diff -u` produced 0 diff lines across all fixtures.
- **Human diagnostics (`lint --no-color`):** `diff -u` produced 0 diff lines across all fixtures.
- **Ordering & labeling:** Identical entity row ordering, issue grouping, and field labels.
- **Severity & exit codes:** Identical exit codes (exit code `11` on defects, `0` on clean repos) and identical advisory/blocking issue tallies.
- **Repair guidance:** Exact match on strings such as `; run \`tskflwctl task depend repair\` for exact source-level diagnosis`.

---

### 5. Hostile mutation testing matrix

| Seam | Mutation Description | Target File & Line | Killing Test / Probe | Observed Failure Output |
| :---: | :--- | :--- | :--- | :--- |
| **1** | Restore path-keyed joining (`t.Path` map key) | `internal/core/service.go:564` | `internal/core/dependency_graph_test.go:688` (and `TestLintAttributesPathlessGraphDiagnosticsToReadableRecords`) | Compile failure: `cannot use taskGraphRecordRefAt(1) ... as string value in map index`; pathless tasks drop all graph issues |
| **2** | Replace record key with task ID | `internal/core/dependency_graph.go:383` | `TestLintRecordAttributionDoesNotCollideOnIDOrLocation` | `lint_source_test.go:144: second record dependency defect leaked onto first record` |
| **3** | Assign references after internal sorting | `internal/core/dependency_graph.go:398` | `TestLintAttributesPathlessGraphDiagnosticsToReadableRecords/dependency_and_lifecycle` | `lint_source_test.go:105: missing depends_on issue containing "not a stable task id" for pathless-invalid ... leaked onto pathless-prerequisite` |
| **4** | Omit reference from cycle diagnostics | `internal/core/dependency_graph.go:538` | `TestLintAttributesPathlessGraphDiagnosticsToReadableRecords/cycle` | `lint_source_test.go:116: missing depends_on issue containing "dependency cycle" for pathless-cycle-left` |
| **5** | Omit reference from legacy diagnostics | `internal/core/dependency_graph.go:662` | `TestLintAttributesPathlessGraphDiagnosticsToReadableRecords/legacy_declaration` | `lint_source_test.go:128: missing blocked_by issue containing "legacy dependency field" for pathless-legacy-owner` |
| **6** | Omit representative lifecycle attribution | `internal/core/service.go:764` | `TestLintAttributesPathlessGraphDiagnosticsToReadableRecords/dependency_and_lifecycle`<br>`TestLintUsesPathlessUnreadableIdentityInLifecycleDiagnosis` | `lint_source_test.go:106: missing status issue containing "dependency gate" for pathless-in-flight`<br>`lint_source_test.go:164: missing status issue containing "6g0000000005"` |
| **7a** | Stop clearing `recordRef` from `Problems()` | `internal/core/dependency_graph.go:830` | `TestTaskGraphHealthAndDeterministicStructuralProblems` | `dependency_graph_test.go:65: seed 0 changed diagnostics ... want recordRef:1 got recordRef:3` |
| **7b** | Stop clearing `recordRef` from `LegacyDiagnostics()` | `internal/core/dependency_graph.go:839` | **None (SURVIVED)** | Entire test suite passed (`go test ./...` exit 0). Recorded as Finding L1. |
| **8** | Restore core location parsing in `TaskGraphLoadProblemFromFile` | `internal/core/service_task.go:163` | `TestTaskGraphReadFromFilesPreservesListDiagnosticsAndAdaptsGraphMessage` | `dependency_graph_test.go:862: core inferred identity from an opaque location: {TaskID:0b3f2t3akr84 ...}` |

---

### 6. Hostile architectural angles

1. **Fragility of ordinal reference design:**
   `taskGraphRecordRef` relies on 1:1 positional indexing `taskGraphRecordRefAt(index) = index + 1` between `tasks []TaskWithBody` and `ordered []taskGraphRecord`. While safe today because `taskRecords` is constructed as a direct loop copy without filtering or reordering, this creates an implicit architectural invariant. Any future intermediary filter (e.g. skipping archived tasks before graph build) or secondary graph construction would silently misalign diagnostics.
2. **Duplicate records with identical slugs:**
   When two conflicting records share both canonical task ID and slug, the analyzer isolates defect attribution to each record's internal reference. However, `LintResult` exposes only `Slug string` and `Issues []domain.Issue`. On the CLI and wire JSON (`wire.LintTaskJSON`), callers receive multiple entries with the identical `Slug` label. Callers cannot correlate entries back to distinct source files from `Slug` alone unless they inspect the duplicate ID issue message paths. This preserves the v1.68 wire schema without breaking API consumers.
3. **Deterministic tie-breaking on duplicate records:**
   In `newTaskGraph`, ties on `(canonicalTaskID, Path, Slug)` fall back to `ordered[i].ref < ordered[j].ref`, making representative selection among complete ties dependent on input slice order. However, duplicate IDs mark `g.hardBroken[taskID] = true` and snapshot health as `GraphBroken`, ensuring the graph fails closed for all mutations, board dispatch, and scheduling decisions.
4. **Actionability of malformed filenames:**
   When a file name does not match the `<id>-<slug>.md` convention (e.g. `notes.md` or scanner errors), `splitFlatName` returns `ok = false`. The store sets `problem.Path` with empty `EntityID`. Core renders this as an unreadable task file with the exact path preserved, ensuring repairability is not weakened.
5. **Architectural scope & non-overreach:**
   The diff is strictly bounded to correlation plumbing. It introduces no changes to graph validity semantics, repair authorization, lifecycle policy, source CAS tokens, wire schemas, or storage adapters.

---

## Findings

#### L1. Public `TaskGraph.LegacyDiagnostics()` does not verify suppression of snapshot-local `recordRef` · **Status:** fixed

In [`internal/core/dependency_graph.go:839`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph.go#L839), `LegacyDiagnostics()` clears the internal correlation handle before returning public diagnostics:

```go
func (g *TaskGraph) LegacyDiagnostics() []LegacyDependencyDiagnostic {
	out := make([]LegacyDependencyDiagnostic, len(g.legacy))
	for i, diagnostic := range g.legacy {
		out[i] = diagnostic
		out[i].recordRef = 0 // snapshot-local correlation is not query data
```

During hostile mutation testing (Seam 7b), omitting this clearing line (`out[i].recordRef = 0`) allowed the snapshot-local `recordRef` to leak into public `LegacyDependencyDiagnostic` query objects without failing a single test in the repository (`go test ./...` passed with exit code 0).

In contrast, omitting `out[i].recordRef = 0` from `Problems()` (line 830) was immediately caught and killed by `TestTaskGraphHealthAndDeterministicStructuralProblems` in [`internal/core/dependency_graph_test.go:65`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9qcY17/internal/core/dependency_graph_test.go#L65) due to stringified struct comparison across shuffled input permutations.

**Recommendation:** Add an explicit assertion in `internal/core/dependency_graph_test.go` (e.g. in `TestTaskGraphResolvesLegacyDependencyFields`) verifying that every `LegacyDependencyDiagnostic` returned by `graph.LegacyDiagnostics()` has `recordRef == 0`.

**Resolution:** Added a boundary assertion that internal legacy diagnostics
retain their readable-record reference while LegacyDiagnostics clears it before
returning public query data. Removing the clearing line now fails
TestTaskGraphLegacyResolutionHealthAndDirection.
