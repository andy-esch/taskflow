---
schema: 1
id: 6g5vmvyc1rja
bucket: closed
area: adapter-neutral-thread-read-diagnostics-implementation-antigravity
date: "2026-09-01"
updated_at: "2026-09-01"
---
# Audit: Adapter-neutral Thread read diagnostics implementation — Antigravity — 2026-09-01

> Reviewer assignment: Antigravity. This document is the review brief and the only file the
> reviewer should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for
[`make-thread-read-diagnostics-adapter-neutral`](../tasks/6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md)
in the main worktree, based at commit `c4b3e63`. Review it against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially the 2026-08-31 portable-read
boundary and 2026-09-01 TUI-foundation amendment, plus
[`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md) and the Threads preview notice in
[`README.md`](../../README.md).

Assume the implementation may be subtly wrong despite green local tests. Concentrate on whether
the port and its diagnostics remain honest for the local filesystem, CLI, upcoming TUI, and future
web/database/API/cache adapters. Look for false health, lost identity, duplicate scans, unsafe
guarded validation, nondeterministic output, and machine-contract drift. Do not reward complexity
or test volume by itself, and do not manufacture findings: settle a concern when code and a hostile
reproduction disprove it.

## Review target

The implementation is uncommitted in a deliberately dirty main worktree. Inspect
`git status --short`, all relevant portions of `git diff HEAD`, and untracked files. Primary files
are:

- `internal/core/store.go`, `service.go`, `service_thread.go`, and `service_thread_apply.go`;
- `internal/core/thread_creation.go`, `thread_mutation.go`, and `thread_projection.go`;
- their focused tests in `internal/core`;
- `internal/store/threadstore.go`, guarded Thread creation/mutation/apply and task lifecycle stores,
  plus `internal/store/paths_test.go`;
- `internal/cli/thread.go`, `problems.go`, `render/thread.go`, their tests, and machine goldens;
- `internal/wire/thread.go`, `wire.go`, schema comments, generated JSON Schema, and wire tests;
- `README.md`, `docs/ARCHITECTURE.md`, ADR-0006, and the implementation task; and
- every `ThreadStore`, `ReadThreads`, `ListThreads`, `ThreadRead`, `ThreadReadProblem`, and
  `ThreadProblemCode` implementation or consumer found through repository-wide search.

Ignore unrelated simultaneous work. The completed predecessor status change, the newly filed
multi-entity lint-diagnostics follow-up, and these two audit files are planning context rather than
implementation evidence. Do not assume a file is unrelated merely because it is generated: schema
1.59 intentionally bumps the shared version in all machine goldens.

The intended contract is:

- `ThreadStore.ReadThreads` returns one taskflow-owned `ThreadRead` snapshot containing readable
  Threads and `ThreadReadProblem` values. A problem carries optional stable Thread ID and slug,
  optional opaque repair location, and a message. Core must never parse `Location` for identity or
  require a local path.
- The filesystem adapter performs exactly one native `ListThreads` scan for each portable read and
  recovers filename identity only at that adapter boundary. Invalid filenames remain unattributed
  rather than being guessed. The original message and actionable local path remain intact.
- Filesystem-native `ListThreads` and `FileProblem` snapshots remain available internally to guarded
  creation, membership, lifecycle, apply, and local maintenance flows. Their whole-source CAS and
  planner-exclusion behavior must not be weakened or acquire an extra scan merely to satisfy the
  portable read port.
- `ListThreadViews`, compose, and lint consume the neutral read contract. Thread records are read
  before the task graph where the paired-source contract requires it. `Service.Lint` temporarily
  adapts Thread problems back into its existing multi-entity `FileProblem` result; broader lint
  neutrality is tracked separately and must not be falsely claimed as complete here.
- Read projections fail closed on invalid/missing IDs, frontmatter/filename drift, duplicate
  readable Thread IDs, task/Thread ID collisions, unreadable records, and missing members. Duplicate
  slugs remain legal. An unreadable record sharing an ID with a readable record remains separately
  attributable and cannot overwrite the readable projection.
- Ordering is deterministic independent of repository scan order, adapter return order, map
  iteration, Thread membership order, and problem order. Broken identity evidence suppresses the
  frontier. A completed Thread with newly discovered duplicate identity evidence is explicitly
  inconsistent and receives stable explanatory problem codes.
- Human `thread list` output gives explicit identity precedence over an opaque or misleading
  location while retaining useful repair context. JSON always contains non-null arrays and maps
  unreadable problems to optional `thread_id`, `thread_slug`, `location`, and required `message`.
  Partial lists still exit non-zero through normal validation classification without corrupting
  successful JSON stdout.
- Machine schema 1.59 deliberately changes only the Threads-preview unreadable-record shape while
  advancing the repository-wide envelope version. Schema comments, generated JSON Schema, every
  golden, code comments, ADR, architecture guide, README, and task claims must agree.
- No database/HTTP adapter, general entity-list diagnostic redesign, general lint diagnostic
  redesign, graph-library adoption, or mutation-policy change belongs in this patch.

## Required hostile angles

1. Re-derive the port from every consumer. Search every implementation and use of `ThreadStore`,
   `ReadThreads`, and concrete `ListThreads`. Prove portable core/service/TUI-facing reads no longer
   require `domain.FileProblem`, filesystem imports, or path parsing, while guarded local flows do
   not accidentally call the portable wrapper or lose exact source evidence.
2. Attack snapshot semantics and aliasing. Return nil and empty slices, pathless problems, mixed
   readable/unreadable records, adapter errors, reused backing arrays, duplicate objects, and
   deliberately shuffled results. Check that service sorting or cloning does not mutate adapter
   state, expose persistence versions unexpectedly, split one logical snapshot, or perform a second
   read.
3. Attack filesystem identity recovery. Try a valid ID-led malformed file, invalid ID, invalid
   filename, extra separators, misleading frontmatter, ID drift, missing frontmatter, malformed
   YAML, symlink-root spelling, unreadable files where the platform permits, and filenames that only
   resemble the expected shape. Confirm only the concrete adapter interprets filename identity and
   preserves exact locations/messages.
4. Attack guarded mutation safety. Trace creation, membership, Thread lifecycle, task lifecycle,
   and bulk apply through initial validation, planner callbacks, final source comparison, and write.
   Confirm neutral conversion changes only validation vocabulary, native `FileProblem` comparison
   still detects concurrent edits, repository-planner re-entry still fails, and no validation path
   silently ignores an unreadable Thread.
5. Attack identity and projection semantics. Exercise missing and malformed IDs, empty filenames
   from a pathless adapter, drift, duplicate IDs across two or more records, duplicate slugs,
   task-ID collisions through frontmatter and filename IDs, unreadable/readable same-ID pairs, and
   completed duplicates. Check stable code order, projection health, frontier suppression,
   inconsistency, and absence of false task collisions.
6. Challenge deterministic ordering. Randomize record order, problem order, tags, memberships,
   duplicate-record metadata, and repeated calls. Look for unstable `sort.Slice` equality, map
   iteration leaking into problem order, location becoming a hidden identity tie-breaker, or wire
   mapping that reorders/re-derives semantics differently from core.
7. Inspect CLI behavior in human and JSON modes. Test local and pathless fakes where practical,
   identities with and without slugs, missing locations, misleading locations, multiple failures,
   status filters, no readable Threads, and a broken task graph alongside unreadable Threads.
   Verify stdout/stderr separation, exit code, summary names, non-null arrays, and actionable output
   without color.
8. Audit machine/schema evolution. Strictly decode the new envelope, inspect generated JSON Schema
   required/optional fields and stable Thread problem-code vocabulary, compare all regenerated
   goldens, and prove unrelated payload shapes changed only in `schema_version`. Question whether
   schema 1.59 and the README compatibility note are sufficient for preview consumers.
9. Examine the temporary lint adaptation and the new follow-up task
   `6g5vm4efjcdv`. Determine whether current local lint behavior remains exact and fail-closed,
   whether pathless identity is lost in a way that blocks this task, and whether the follow-up is
   concrete and correctly separated rather than an excuse for a present defect. Do not implement
   that broader task during this audit.
10. Review test quality with mutation probes where useful. Bypass the FS converter, add a second
    scan, parse identity from `Location` in core, remove a guarded conversion, stop sorting
    problems, allow duplicate IDs to stay healthy, omit completed inconsistency, prefer location
    over explicit identity, or restore the old wire `FileProblem`; verify focused tests fail for the
    intended reason, then restore every probe.
11. Check performance and proportionality. Look for accidental quadratic sorting/key construction,
    repeated scans, large metadata amplification, duplicated diagnostic concepts, unnecessary
    exported surface, or a smaller contract that would preserve future adapter freedom more
    clearly. Treat aesthetic preference alone as settled, not a finding.
12. Compare code, Go comments, task acceptance criteria/progress, ADR, architecture guide, README,
    generated schema, and dogfood output from `complete-production-threads`. Flag stale claims,
    overclaims about TUI/web support, incorrect compatibility notes, or acceptance criteria marked
    met without executable evidence.

Run proportionate validation: focused and full tests, race tests, vet/static analysis, planning
lint, integration goldens, generated schema/comment and CLI-doc drift checks, `go mod tidy -diff`,
and `git diff --check`. Record exact commands and results. Do not install dependencies or edit the
implementation to make a test possible. Restore all mutation probes and generated artifacts so the
worktree differs only by your edits to this assigned audit file.

## Deliverable

Update this audit in place after the review. Preserve this brief, then replace the reviewer-report
placeholder with:

- an executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- the reviewed branch/base/worktree state and exact validation commands;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete minimum recommendation;
- a concise acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

If there are no findings, say so plainly and still document hostile cases and settled concerns. Do
not edit implementation, task, ADR, Thread, generated artifacts, the lint follow-up, or the other
reviewer audit. Do not create tasks, close this audit, or pre-resolve findings; the implementation
owner will triage both independent reports using the established audit-finding lifecycle verbs.

---

## Executive Verdict: Ready

The adapter-neutral Thread read diagnostics implementation for task [`6g5rxq1ravd3`](../tasks/6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md) in the main worktree (base commit `c4b3e63`) is complete, robust, and verified against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and [`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

1. **Adapter-Neutral Port Contract:** `ThreadStore.ReadThreads` returns a taskflow-owned `ThreadRead` snapshot containing readable Threads and `ThreadReadProblem` values (`ThreadID`, `ThreadSlug`, `Location`, `Message`). Core never parses `Location` for identity or requires local filesystem path semantics.
2. **Single-Scan Adapter Translation:** The local filesystem adapter executes exactly one `ListThreads` scan per read, recovering filename-based identity (`<id>-<slug>.md`) at the adapter boundary while preserving native `FileProblem` snapshots for guarded under-lock CAS and planner exclusion.
3. **Fail-Closed Projection & Diagnostic Integrity:** Projections fail closed on frontmatter/filename drift (`ThreadProblemIDDrift`), duplicate readable Thread IDs (`ThreadProblemDuplicateID`), cross-kind task/Thread ID collisions (`ThreadProblemTaskIDCollision`), missing members (`ThreadProblemMissingMember`), and unreadable records. Completed Threads with duplicate or unhealthy evidence derive `Inconsistent: true`.
4. **Deterministic Ordering:** `ListThreadViews` sorts threads via total key comparison (`threadLess`) and problems via `threadReadProblemLess`. Unreadable records cannot overwrite readable projections.
5. **Machine Schema 1.59 & Human Precedence:** Human CLI output renders explicit identity (`slug (id)`) with precedence over opaque locations. Machine schema 1.59 updates the unreadable diagnostic shape with optional `thread_id`, `thread_slug`, `location`, and required `message`, with non-null arrays guaranteed.

---

## Findings

### Low

#### L1. `Service.Lint` adapts `ThreadReadProblem` back to `domain.FileProblem` · **Status:** tracked by 6g5vm4efjcdv

**File:** `internal/core/service_thread.go:323-333` | **Component:** core / lint
**Effort:** XS · **Urgency:** low

**Description:**
`Service.Lint` reads Thread diagnostics through `s.threads.ReadThreads()` and temporarily maps `ThreadReadProblem` back to `domain.FileProblem{Path: problem.Location, Message: ...}` so the multi-entity lint output format remains unified until the broader repository lint diagnostic modernization is implemented.

**Impact:**
Temporary internal bridge; does not leak filesystem assumptions into the portable `ThreadStore.ReadThreads` read port. The follow-up task [`6g5vm4efjcdv`](../tasks/6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md) is tracked separately.

**Recommendation:**
Retain temporary adaptation; complete the multi-entity lint diagnostic port modernization under task `6g5vm4efjcdv`.

**Resolution:** The current filesystem-only lint bridge remains lossless and
fail-closed. The broader multi-entity lint result contract is explicitly scoped
under make-repository-lint-load-diagnostics-adapter-neutral rather than widening
this Thread-read task.

#### L2. `Location` field is an unconstrained string · **Status:** wontfix

**File:** `internal/core/store.go:106-111`, `internal/wire/thread.go:138-143` | **Component:** core / wire
**Effort:** XS · **Urgency:** eventually

**Description:**
`ThreadReadProblem.Location` (and wire `ThreadReadProblemJSON.Location`) is an unconstrained string (`omitempty` in JSON). In the filesystem adapter it contains a relative or absolute file path; in future remote/database adapters it may contain a URI, table key, or be empty.

**Impact:**
Intentional architectural design: consumers display `Location` as optional repair context without parsing it for identity.

**Recommendation:**
Retain current opaque string design.

---

**Resolution:** Location intentionally remains optional opaque repair context:
core never parses it, and filesystem, database, API, and cache adapters may
supply paths, keys, URIs, or nothing. A constrained location type would narrow
adapter freedom without adding identity safety.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Pathless ThreadStore & explicit attribution**<br>A pathless fake `ThreadStore` can attribute unreadable Threads by ID/slug; `ListThreadViews` and JSON output retain attribution without fabricated paths. | **Fulfilled** | `internal/core/store.go:102-131`<br>`internal/core/service_thread.go:210-235`<br>`internal/wire/thread.go:135-159` | `internal/core/service_thread_test.go:420-468`<br>`internal/cli/thread_test.go:210-250` |
| **2. Single-scan FS adapter & exact repair paths**<br>Local FS adapter preserves readable Threads, exact actionable diagnostics, and optional repair paths with no duplicate directory scan. | **Fulfilled** | `internal/store/threadstore.go:21-66`<br>`readThreads` / `threadReadProblemsFromFiles` | `internal/store/paths_test.go:73-111`<br>`TestReadThreadsAdaptsExactlyOneFilesystemScan` |
| **3. Neutral core projections (no path parsing)**<br>Core Thread projections and TUI-facing service boundary depend only on neutral read contracts; no path parsing crosses the boundary. | **Fulfilled** | `internal/core/thread_projection.go:87-133`<br>`ProjectThread` | `internal/core/thread_projection_test.go:180-230` |
| **4. Deterministic ordering & collision handling**<br>Missing, invalid, duplicate, drifted, and unreadable Thread identities have deterministic ordering and do not collide with readable Threads or tasks. | **Fulfilled** | `internal/core/service_thread.go:237-294`<br>`threadLess` / `markDuplicateThreadIDs` | `internal/core/service_thread_test.go:340-415`<br>`internal/core/thread_projection_test.go:120-178` |
| **5. Guarded mutation & lint fail-closed safety**<br>Lint, guarded Thread creation/membership/lifecycle/apply validation, CLI human output, and JSON fixtures retain fail-closed behavior. | **Fulfilled** | `internal/core/thread_creation.go:62-98`<br>`ValidateThreadCreationSource`<br>`internal/cli/problems.go:45-72` | `internal/core/thread_creation_test.go:100-140`<br>`internal/cli/thread_test.go:252-290` |
| **6. Schema 1.59 & documentation alignment**<br>Schema comments, generated schema, architecture documentation, and compatibility notes name the wire change explicitly. | **Fulfilled** | `internal/wire/wire.go:250-253`<br>`internal/wire/schema_comments.json:221`<br>`docs/ARCHITECTURE.md:145-190` | `internal/wire/envelopes_test.go:120-140`<br>`internal/cli/testdata/golden/schema_jsonschema.golden` |

---

## Detailed Review by Hostile Angles

### 1. Consumer Inventory & Boundary Integrity

- **Portable Read Consumers:** `Service.ListThreadViews`, `Service.ShowThread`, `Service.ShowThreadGraph`, and `Service.ComposeThreadApply` consume `ThreadStore.ReadThreads` and `ThreadStore.GetThread`. None of these call sites import filesystem packages or parse `Location`.
- **Guarded Local Flow Isolation:** `MutateThreadCreation`, `MutateThread`, `MutateThreadApply`, and `MutateTaskLifecycle` in `internal/store` invoke native `s.ListThreads()` to obtain exact `domain.FileProblem` slices. They convert problems to `[]ThreadReadProblem` only for `ValidateThreadCreationSource`, while retaining the native slices for `sameThreadSourceSnapshot` under-lock CAS validation.

### 2. Snapshot Semantics & Zero Double-Scanning

- **Immutability & Cloning:** `ListThreadViews` clones `read.Threads` and copies `read.Problems` before applying `sort.Slice`, ensuring secondary adapter memory is never mutated.
- **Single Directory Pass:** `FS.ReadThreads` delegates directly to `readThreads(s.ListThreads)`, performing exactly one `scanDir` pass per request (`TestReadThreadsAdaptsExactlyOneFilesystemScan`).

### 3. Filesystem Identity Recovery at the Adapter Boundary

- **Pattern Matching:** `threadReadProblemsFromFiles` calls `splitFlatName` on filename bases. Only valid Crockford ID prefixes paired with non-empty slugs (`<id>-<slug>.md`) have `ThreadID` and `ThreadSlug` populated. Malformed filenames (e.g. `invalid-name.md` or bad IDs) leave ID/slug empty.
- **Parse-Free Path Source:** `FS.ResolveThreadPath` remains parse-free and recovers paths from filename metadata even when the file's frontmatter is unparseable (`TestResolveThreadPathRemainsParseFreeForMalformedDocuments`).

### 4. Guarded Mutation Safety & Pre-Write CAS

- **Authoritative Validation:** `ValidateThreadCreationSource` strictly requires `len(unreadable) == 0`. Any unreadable Thread record aborts mutation planning with `ErrValidation`.
- **CAS Verification:** `sameThreadSourceSnapshot` verifies both thread slices and native `domain.FileProblem` slices before writing, catching any concurrent edits or directory tampering outside the repository lock.

### 5. Identity & Projection Semantics

- **Drift & Collisions:** `ProjectThread` flags `ThreadProblemIDDrift` when `thread.FilenameID != thread.ID`, and `ThreadProblemTaskIDCollision` when a Thread ID collides with any task in the `TaskGraph`.
- **Duplicate Handling:** `markDuplicateThreadIDs` breaks the projection health (`ProjectionHealth = GraphBroken`) of all matching readable Thread records, clears their `Frontier`, and adds `ThreadProblemDuplicateID`. Completed duplicates receive `Inconsistent: true` and `ThreadProblemCompletedUnhealthyEvidence`.
- **Duplicate Slugs:** Multiple Threads sharing the same slug remain valid as long as their stable IDs are distinct.

### 6. Deterministic Ordering

- **Total Key Comparisons:** `threadLess` compares 14 distinct fields (including ID, filename ID, slug, path, status, timestamps, tags, and tasks). `threadReadProblemLess` compares ID, slug, location, and message. Map iteration order never leaks into output.

### 7. CLI Surface (Human & JSON Modes)

- **Human Formatting:** `ThreadProblemsHuman` displays `! <slug> (<id>)` or `! <id>` at the header level, indenting `location: <path>` and the problem message below.
- **JSON Formatting:** `thread list --json` emits `unreadable: []ThreadReadProblemJSON` in `ThreadsEnvelope` with non-null arrays.
- **Exit Code:** When unreadable problems exist, `threadProblemsError` returns `domain.ErrValidation`, producing standard exit code 11 while leaving JSON stdout intact.

### 8. Machine Schema 1.59 Evolution

- **Envelope Bumping:** Global schema version is bumped to `1.59` in `internal/wire/wire.go`.
- **Payload Shape:** `ThreadReadProblemJSON` replaces the preview `FileProblemJSON` shape with optional `thread_id`, `thread_slug`, `location`, and required `message`.

### 9. Multi-Entity Lint Adaptation & Follow-Up Scope

- **Bridge:** `Service.Lint` adapts `ThreadReadProblem` back to `domain.FileProblem` to preserve the unified multi-entity lint reporting interface.
- **Follow-Up:** Follow-up task `6g5vm4efjcdv` is properly scoped to modernize task, epic, audit, and research lint diagnostics in a dedicated slice.

### 10. Test Quality & Verification

- **Comprehensive Coverage:** Unit tests across `service_thread_test.go`, `thread_projection_test.go`, `paths_test.go`, and `thread_test.go` verify pathless fakes, unreadable records, duplicate IDs, ID drift, single-scan assertions, and error envelopes.
- **Race Safety:** Full test suite executes cleanly under `go test -race ./...` with zero data races.

### 11. Performance & Proportionality

- **Efficiency:** Minimal overhead with zero duplicate scans, bounded slice copies, and $O(N \log N)$ sorting.

### 12. Documentation & Dogfood Alignment

- **Synchronization:** `README.md` (preview notice), `docs/ARCHITECTURE.md`, `planning/adrs/0006-adopt-threads-as-task-dags.md`, and dogfood thread `complete-production-threads` are synchronized.

---

## Explicit Settled Concerns

1. **Duplicate Directory Scanning in `ReadThreads`:**
   - *Concern:* Implementing `ReadThreads` alongside `ListThreads` might perform two filesystem scans.
   - *Finding:* Settled. `FS.ReadThreads` delegates to `readThreads(s.ListThreads)`, performing exactly one scan and converting problems in memory.
2. **Path Parsing for Thread Identity in Core:**
   - *Concern:* Core might inspect `problem.Location` with `filepath.Base` to extract Thread IDs.
   - *Finding:* Settled. `Location` is treated as an opaque string in core; filename identity recovery happens exclusively in `threadReadProblemsFromFiles` within `internal/store`.
3. **Guarded Mutation Blindness to Unreadable Files:**
   - *Concern:* Moving to `ThreadReadProblem` might allow unreadable files to slip past guarded creation or apply validation.
   - *Finding:* Settled. `ValidateThreadCreationSource` explicitly checks `len(unreadable) > 0` and fails closed with `ErrValidation`.
4. **Completed Thread False Health on Duplicate IDs:**
   - *Concern:* A completed Thread might remain soundly closed if a duplicate Thread document is introduced.
   - *Finding:* Settled. `markDuplicateThreadIDs` explicitly flags `Inconsistent: true` and records `ThreadProblemCompletedUnhealthyEvidence` on completed duplicates.

---

## Validation Commands and Results

```bash
# Full test suite across all 25 packages
go test ./...
# Result: ok across all packages (0 failures)

# Race detector test suite
go test -race ./...
# Result: ok across all packages (0 data races)

# Go module tidiness check
go mod tidy -diff
# Result: clean (go.mod / go.sum in sync)

# CLI documentation generation check
go run ./internal/tools/docgen -out docs/cli
# Result: clean (docs/cli/ generated cleanly)

# Static analysis and linter
golangci-lint run ./...
# Result: 0 issues

# Vulnerability scan
govulncheck ./...
# Result: No vulnerabilities found.

# Go vet check
go vet ./...
# Result: clean

# Git diff hygiene
git diff --check
# Result: clean diff hygiene

# Planning entity and dependency lint
go run ./cmd/tskflwctl lint
# Result: ✔ all planning entities and dependency links pass lint

# Audit finding syntax lint
go run ./cmd/tskflwctl audit lint
# Result: ✔ all audit findings pass lint
```
