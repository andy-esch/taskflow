---
schema: 1
id: 6g726063yv11
bucket: closed
area: version-unreadable-task-graph-sources-for-repair-safe-cas-implementation-antigravity
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: unreadable task graph source revisions — Antigravity — 2026-09-05

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer
> should update.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Codes must match
> `[A-Z]+[0-9]+`; do not put status on a separate line or pre-resolve a finding.
>
> Required second pass: after completing the checklist, review again as a devil's advocate for
> systemic failure modes. Challenge the source/projection boundary, compatibility adapters,
> normalization and equality assumptions, test helpers, and future-repair claims. Prefer one
> demonstrated systemic issue over several speculative findings. A green suite and a restatement of
> the implementation are not a review; execute restored hostile mutations and try to falsify each
> load-bearing claim.
>
> Shared-worktree isolation is mandatory. Treat the handoff checkout as a read-only source. Before
> inspecting implementation, running tests or generators, or making mutation probes, create the
> independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement whose `.git`
> metadata points back to the shared checkout.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may use the handoff checkout concurrently. Reading
this brief and performing the initial copy are the only operations allowed there until the final
guarded audit transfer. Capture the exact current contents, including staged, unstaged, untracked,
and deleted files:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g726063yv11-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review-antigravity.XXXXXX")"

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

Confirm `git rev-parse --git-dir` resolves inside `$SANDBOX`. Perform all inspection, builds,
formatting, generation, fixtures, mutation probes, and report editing there. Never commit again,
switch branches, stage, restore, clean, stash, reset, or run a write-capable project command in
`$SOURCE_ROOT`. If isolation cannot be verified, stop and report the blocker.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`; inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then transfer
only the audit, guarded against concurrent source edits:

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

Do not copy anything else back. Leave the sandbox in place and report its path until the
implementation owner confirms receipt.

## Review brief

Perform an independent adversarial implementation review of
[version unreadable task graph sources for repair-safe CAS](../tasks/6g721tvf4crh-version-unreadable-task-graph-sources-for-repair-safe-cas.md)
on branch `feat/unreadable-graph-source-revisions`, based on `main` at `c921099`. Planning is in
commits `b4271b5` and `7835374`; the implementation under review is the complete uncommitted working
tree captured by the sandbox checkpoint. Review its diff against `7835374` and judge it against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md),
[the architecture guide](../../docs/ARCHITECTURE.md), the task acceptance criteria, and existing
graph mutation/OCC contracts.

Assume the change can be systemically wrong despite green tests. Re-derive the data path from exact
source bytes through the filesystem scan, neutral `TaskGraphRead`, private immutable `TaskGraph`
evidence, `SameSourceSnapshot`, and the pre-write conflict helper. Do not edit implementation or
other planning files, create follow-up tasks, change finding statuses, close this audit, push, or
install anything globally.

## Intended contract to challenge

- `TaskGraphLoadProblem.SourceVersion` is an opaque adapter-owned revision of the exact unreadable
  source. Core may copy, normalize, and compare it but may not parse it or expose it through domain,
  planner, YAML, JSON, schema, CLI, TUI, or wire contracts.
- The filesystem hashes the exact bytes already read during its resilient task scan, before parse
  success is known. `ReadTaskGraph` retains that token for unreadable records; ordinary `ListTasks`
  still returns the established `domain.FileProblem` shape without another scan or token leak.
- `TaskGraph` privately retains a defensive, deterministic copy of load problems. Adapter ordering
  is irrelevant. Snapshot equality compares identity, optional location, message, and a non-empty
  equal revision for each unreadable source in addition to existing readable-task versions and
  semantic evidence.
- Missing revision evidence is diagnostic-compatible but CAS-ineligible: it must fail closed rather
  than claim two unreadable snapshots are equal. A pathless/remote adapter can supply any stable
  opaque revision without inventing a filesystem path.
- Readable-to-unreadable and unreadable-to-readable transitions, additions, removals, renames,
  identity changes, and byte changes that reproduce the same parse error invalidate equality.
  Restoring byte-identical source restores the filesystem revision.
- Existing healthy graph reads and guarded mutations retain behavior. A named pre-write snapshot
  verifier returns `ErrConflict` on stale evidence and can be reused by the forthcoming dedicated
  broken-graph repair store. This task does not implement or weaken repair authorization.
- Hashing remains O(total source bytes already read), uses the existing content-CAS primitive, and
  creates no second task scan, filesystem watch, persisted version field, or graph-library need.

## Mandatory evidence floor

A `ready` verdict is not credible without:

1. A producer/consumer inventory for `scanDir`, `scanDirWithSourceVersions`, `ListTasks`,
   `ReadTaskGraph`, `TaskGraphReadFromFiles`, every `TaskGraphSource` production composition path,
   `NewTaskGraphRead`, `SameSourceSnapshot`, and every guarded store calling it.
2. Hostile fixtures for unchanged unreadable bytes, changed bytes with the same error, changed body
   outside malformed frontmatter, readable/unreadable transitions, add/remove/rename, non-ID-led
   files, duplicate unreadable identities, pathless records, reordered problems, and missing tokens.
3. Proof that ordinary list, human/JSON lint/status, generated schema, planner-facing task values,
   and public wire envelopes cannot reveal the token or change compatibility.
4. Concurrency reasoning at both the whole-snapshot and immediate per-target CAS boundaries. State
   precisely what this closes and what raw-editor verify-to-rename window remains.
5. Restored mutation probes that at minimum: omit the unreadable revision comparison; derive the
   revision from path/message rather than bytes; return an empty FS token; expose the token through
   JSON; ignore adapter ordering; and drop unreadable problems from the prospective snapshot.
   Name the test that kills each mutation and flag any surviving mutation as a coverage finding.
6. Repeated focused tests under `-race`, an uncached full `go test -race ./...`, static analysis,
   module tidiness, planning/audit lint, and `git diff --check`, with exact commands and results.

## Required adversarial angles

1. **Snapshot identity.** Look for false equality, needless false conflicts, asymmetric comparison,
   duplicate-record pairing errors, aliasing of caller slices, empty-token acceptance, or hash input
   that differs from the bytes the parser actually consumed.
2. **Portability and compatibility.** Challenge pathless sources, legacy `TaskStore` adaptation,
   fake stores, split read/write adapters, serialization tags, schema generation, and assumptions
   that every source is a local Markdown file.
3. **Scan and error semantics.** Ensure read failures remain fatal where intended, parse failures
   remain resilient, directory ordering does not become identity, and the refactor does not change
   README/symlink/non-regular-file carveouts or perform duplicate reads.
4. **Concurrency truthfulness.** Verify the named helper is actually on the production pre-write
   path, preserves typed `ErrConflict`, and does not imply atomicity beyond current lock/CAS limits.
5. **Regression and scope.** Exercise healthy/degraded/broken reads, lifecycle/Thread/dependency
   guards, lint, status, and graph projections. Reject repair policy, source-declaration modeling,
   graph-library adoption, or generalized entity revisioning smuggled into this slice.
6. **Systemic second pass.** Seek a shared equality, cloning, resilient-read, or adapter-composition
   convention that makes the local patch appear correct while another production path remains blind.
   Demonstrate the issue or record the evidence that settles it.

## Validation and restoration

Run every probe only inside the independent sandbox. Restore all implementation mutations and
generated artifacts to the sandbox checkpoint. At finish, sandbox `git status --short` must show
only this assigned audit before the guarded one-file transfer.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- branch/base/checkpoint, runtime, exact validation, and isolation attestation;
- a compact source-to-CAS data-flow and production consumer inventory;
- findings grouped by severity with exact grammar and `**Status:** open`;
- acceptance-criteria traceability plus hostile-evidence and restored-mutation ledgers;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but still provide the required evidence. Do not
pre-resolve findings; the implementation owner will triage them with `tskflwctl audit finding`.

## Reviewer report

### 1. Executive Verdict

**Verdict:** `ready with tracked follow-ups`

The implementation on branch `feat/unreadable-graph-source-revisions` satisfies all acceptance criteria, closes the whole-snapshot compare-and-swap (CAS) blind spot for unreadable task sources, and adheres strictly to ADR-0006 and the architecture guide. Exact source bytes are hashed before parsing; the resulting opaque `SourceVersion` is retained by `TaskGraphRead` and evaluated by `TaskGraph.SameSourceSnapshot` while remaining completely isolated from domain records, planner callbacks, human CLI/TUI outputs, wire DTOs, and generated JSON schemas. All 10 hostile fixtures passed, and all 6 required mutation probes were definitively killed by focused tests under the race detector.

Four non-blocking follow-up findings are recorded: one test coverage hole for the fail-closed assertion on tokenless unreadable records (M1), one systemic second-pass finding regarding Thread unreadable source CAS equality (M2), and two consistency/cleanliness findings regarding pre-write verifier adoption and duplicate filename parsing (L1, L2). None blocks landing the task-graph CAS slice.

---

### 2. Environment, Validation, and Isolation Attestation

- **Branch / Base / Checkpoint:**
  - Feature branch: `feat/unreadable-graph-source-revisions`
  - Base commit: `c921099` (on `main`); planning commits in `b4271b5` and `7835374`
  - Sandbox baseline checkpoint commit: `47323ae3697afc863dae484c01520a7aa7b42a0b`
- **Runtime Environment:**
  - Go version: `go version go1.26.6 darwin/arm64`
  - Git version: `git version 2.54.0`
  - OS / Arch: macOS (Darwin 25.3.0, arm64)
- **Isolation Attestation:**
  - Sandbox directory: `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T//taskflow-review-antigravity.gSbywS`
  - Created via `git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"` followed by `rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"`.
  - Verified that `git rev-parse --git-dir` resolves to `.git` inside the sandbox directory.
  - No write-capable operations, commits, branch switches, or file edits were performed in the shared `$SOURCE_ROOT` worktree during the audit.
  - All test runs, mutation probes, hostile fixtures, builds, and linting were executed exclusively within the sandbox.
- **Validation Suite and Results:**
  - Repeated focused tests under race detector:
    `go test -race -count=5 -run "TestTaskGraphSameSourceSnapshot|TestUnreadableTaskSourceRevision" ./internal/core ./internal/store` -> **PASS** (core: 1.242s, store: 1.374s)
  - Full uncached race test suite:
    `go test -race -count=1 ./...` -> **PASS** (all 31 packages green; no data races detected)
  - Static analysis:
    `golangci-lint run ./...` (`just lint`) -> **PASS** (0 issues)
  - Go module consistency:
    `go mod tidy -diff` (`just tidy-check`) -> **PASS** (clean)
  - Planning and audit lint:
    `./bin/tskflwctl lint` -> **PASS** (all planning entities and dependency links pass lint)
    `./bin/tskflwctl audit lint` -> **PASS** (all audit findings pass lint)
  - Git whitespace and diff hygiene:
    `git diff --check` -> **PASS** (0 whitespace/syntax defects)
  - Wire and schema leak verification:
    `./bin/tskflwctl schema --json-schema | grep -i "sourceversion"` -> **PASS** (not found)
    `./bin/tskflwctl task list --json | grep -i "sourceversion"` -> **PASS** (not found)
    `./bin/tskflwctl lint --json | grep -i "sourceversion"` -> **PASS** (not found)
    `./bin/tskflwctl status --json | grep -i "sourceversion"` -> **PASS** (not found)

---

### 3. Source-to-CAS Data Flow and Production Consumer Inventory

#### 3.1 Data Flow Architecture

1. **Filesystem Scan (`internal/store/resolve.go`):**
   `scanDirSources` iterates over regular `.md` files in `tasksDir` (skipping `README.md` and symlinks). It executes a single `os.ReadFile(path)`.
2. **Byte Fingerprinting Before Parse:**
   When `versionProblems` is enabled (`scanDirWithSourceVersions`), `sourceVersion = hashContent(content)` computes the SHA-256 hex digest of the raw byte slice *before* calling `parseTask`.
3. **Resilient Problem Capture:**
   If `parseTask` fails (due to absent frontmatter, YAML decode error, invalid filename structure, or missing status), the error is caught and wrapped into `sourceFileProblem{problem: domain.FileProblem{Path: path, Message: err.Error()}, sourceVersion: sourceVersion}`. Read errors remain fatal; parse errors remain resilient.
4. **Adapter Boundary Translation (`internal/store/fsstore.go`):**
   `FS.ReadTaskGraph()` consumes `sourceProblems`. For each problem, it extracts `taskID` and `taskSlug` using `splitFlatName` and populates `core.TaskGraphLoadProblem{TaskID, TaskSlug, Path, Message, SourceVersion}`. The returned `core.TaskGraphRead` binds parsed valid `Tasks` with unreadable `Problems`.
5. **Deterministic Graph Normalization (`internal/core/dependency_graph.go`):**
   `newTaskGraph` clones `unreadable` problems via `cloneTaskGraphLoadProblems` (guaranteeing caller slices are not aliased) and sorts them with `sort.SliceStable` using the composite key:
   `strings.Join([]string{left.TaskID, left.TaskSlug, left.Path, left.Message, left.SourceVersion}, "\x00")`.
   This ensures adapter scan order is completely decoupled from graph snapshot identity. The sorted problems are preserved as unexported `loadProblems` in `*TaskGraph`.
6. **Authoritative Whole-Snapshot CAS (`SameSourceSnapshot`):**
   `g.SameSourceSnapshot(other)` compares `g.health`, `g.ids`, task `SourceVersion`s, and load problems via `slices.EqualFunc(g.loadProblems, other.loadProblems, sameTaskGraphLoadProblem)`.
   `sameTaskGraphLoadProblem` requires:
   `left.TaskID == right.TaskID && left.TaskSlug == right.TaskSlug && left.Path == right.Path && left.Message == right.Message && left.SourceVersion != "" && left.SourceVersion == right.SourceVersion`.
   If any unreadable problem lacks a `SourceVersion`, the comparison immediately fails closed.
7. **Pre-Write Conflict Gate (`internal/store/graphmutation.go`):**
   In `FS.MutateTaskGraph`, after planning and materializing file contents under advisory flock, `core.LoadTaskGraph(s)` re-reads the repository from disk into `currentGraph`. `verifyTaskGraphSourceSnapshot(graph, currentGraph)` evaluates `SameSourceSnapshot`. Any discrepancy returns `domain.ErrConflict` (exit code 14) before a single file write is initiated.
8. **Boundary Isolation:**
   `TaskGraphLoadProblem.SourceVersion` carries `json:"-" yaml:"-"` tags. Planner callbacks querying `graph.Task(id)` receive cloned domain tasks with `SourceVersion` explicitly cleared (`task.SourceVersion = ""`). `ListTasks()` routes through `scanDirSources(dir, parse, false)`, stripping internal revision tokens from `domain.FileProblem`.

#### 3.2 Producer and Consumer Inventory

| Component / Function | Role / Type | Producers | Consumers |
| :--- | :--- | :--- | :--- |
| `scanDirSources` | Internal generic directory scanner | `internal/store/resolve.go:32` | `scanDir`, `scanDirWithSourceVersions` |
| `scanDir` | Resilient scan without problem hashing (`versionProblems: false`) | `internal/store/resolve.go:87` | `FS.ListTasks`, `FS.ListTasksWithBodies`, `FS.ListEpics`, `FS.ListAudits`, `FS.ListAuditsWithFindings`, `FS.ListResearch`, `FS.ListThreads`, `FS.readThreadApplyDocuments`, `descriptor_scan_test.go` |
| `scanDirWithSourceVersions` | Resilient scan with byte hashing (`versionProblems: true`) | `internal/store/resolve.go:78` | `FS.ReadTaskGraph` (exclusive production consumer) |
| `ListTasks` | Primary resilient task list query | `FS.ListTasks` (`internal/store/fsstore.go:103`) | `Service.ListTasks` (CLI `task list`, TUI), `internal/cli/completion.go:222`, `internal/cli/fill.go`, `taskStoreGraphSource.ReadTaskGraph` (fallback adapter), store/core test suites |
| `ReadTaskGraph` | Port capability method (`core.TaskGraphSource`) | `*store.FS` (`fsstore.go:117`), `taskStoreGraphSource` (`service_task.go:137`), test doubles (`taskGraphReadFake`) | `core.loadTaskGraphRecords` -> `core.LoadTaskGraph` |
| `TaskGraphReadFromFiles` | Legacy adapter boundary translator | `internal/core/service_task.go:149` | `taskStoreGraphSource.ReadTaskGraph()` (non-conforming store fallback), `NewTaskGraph` (lint/body-aware callers), tests |
| `TaskGraphSource` | Secondary read port interface | `*store.FS`, `taskStoreGraphSource`, `taskGraphReadFake` | `Service.taskGraphs`: `Board`, `Summary`, `TaskList` (`--unblocked`), `ShowThreadGraphDetail`, `ListThreadViews`, `ComposeThreadApply`, `dependency_operations` |
| `NewTaskGraphRead` | Canonical graph constructor from `TaskGraphRead` | `internal/core/dependency_graph.go:306` | `core.LoadTaskGraph`, `NewTaskGraph`, test fixtures |
| `SameSourceSnapshot` | Whole-repository snapshot equality predicate | `TaskGraph.SameSourceSnapshot` (`dependency_graph.go:889`) | `verifyTaskGraphSourceSnapshot` (`internal/store/graphmutation.go:117`) |
| `verifyTaskGraphSourceSnapshot` | Pre-write repository CAS conflict helper | `internal/store/graphmutation.go:117` | `FS.MutateTaskGraph` (line 87); designated for reuse by the forthcoming broken-graph repair store |

---

### 4. Findings

#### M1. The fail-closed rule for missing unreadable-source revisions is load-bearing but unasserted · **Status:** fixed

`sameTaskGraphLoadProblem` (`internal/core/dependency_graph.go:908`) requires `left.SourceVersion != ""`. That clause is the only barrier stopping two tokenless unreadable snapshots from comparing equal—the brief's explicit requirement that missing revision evidence be "CAS-ineligible: it must fail closed rather than claim two unreadable snapshots are equal".

In `internal/core/dependency_graph_test.go:772`, the `missing` test case sets only one side's token to `""` and compares it against a side holding `"opaque-revision-a"`. Two different strings are unequal regardless of whether `SourceVersion != ""` is checked. Mutating `left.SourceVersion != "" && left.SourceVersion == right.SourceVersion` to `left.SourceVersion == right.SourceVersion` leaves the entire test suite green.

**Failure scenario:** A future refactor or adapter passes tokenless unreadable records into snapshot comparison. If `left.SourceVersion != ""` is accidentally omitted or inverted, two broken snapshots with identical error messages will falsely compare as equal across scans. A test comparing two records where both have `SourceVersion: ""` should be added to ensure the fail-closed behavior is asserted directly.

**Resolution:** Added a direct both-empty regression so two otherwise identical
unreadable records without source revisions remain CAS-ineligible.

#### M2. Pre-write CAS for Thread documents retains the unreadable-source blind spot · **Status:** tracked by 6g72ncs4xjdm

Four production pre-write checks evaluate the task-side and Thread-side comparisons in a single boolean expression:
`if !graph.SameSourceSnapshot(currentGraph) || !sameThreadSourceSnapshot(threads, problems, currentThreads, currentProblems)` (`internal/store/threadmutation.go:92`, `threadapply.go:123`, `threadcreation.go:88`, `lifecyclemutation.go:110`).

While this task successfully adds SHA-256 byte tokens to unreadable task files, `sameThreadSourceSnapshot` (`internal/store/threadcreation.go:179-181`) still compares unreadable Threads using `slices.Equal(leftProblems, rightProblems)` over `[]domain.FileProblem` (`{Path, Message}` only). Furthermore, `core.ThreadReadProblem` (`internal/core/store.go:112`) still carries the placeholder comment `"so a future source revision token can qualify the complete read"`.

**Reachability and forward risk:** Currently, `ValidateThreadCreationSource` refuses broken thread sources before planning, so this is unreachable in today's healthy write paths. However, once broken-graph repair lands, a repair transaction that also touches Thread files would run a half-versioned CAS where concurrent raw edits to unreadable Thread files reproducing the same error would compare equal. This should be addressed when Thread repair is implemented.

**Resolution:** Confirmed the Thread-side blind spot is unreachable in ordinary
guarded mutations but must be closed before repair authorizes partial Thread
evidence; a focused prerequisite is now sequenced in the production Thread.

#### L1. Newly introduced verifyTaskGraphSourceSnapshot helper is bypassed by four other production CAS sites · **Status:** fixed

`verifyTaskGraphSourceSnapshot` (`internal/store/graphmutation.go:117`) is introduced as the named pre-write whole-repository CAS verifier. However, only `graphmutation.go:87` calls it. Four other pre-write sites in `internal/store` (`threadmutation.go:92`, `threadapply.go:123`, `threadcreation.go:88`, `lifecyclemutation.go:110`) inline `!graph.SameSourceSnapshot(currentGraph)` directly with duplicate `domain.ErrConflict` wrapping.

While behavior is equivalent today because all four sites call `core.LoadTaskGraph(s)`, any future enhancement to `verifyTaskGraphSourceSnapshot` (such as structured conflict diagnostics or authorization checks) will not be picked up by the other four call sites unless they route through the shared helper.

**Resolution:** Moved the graph snapshot verifier into the shared CAS module and
routed dependency, Thread creation/mutation/apply, and task lifecycle pre-write
checks through it while preserving operation-specific errors.

#### L2. Filename-to-identity parsing is duplicated between fsstore and service_task · **Status:** fixed

`core.TaskGraphReadFromFiles` carries the comment: `"Filename parsing for unreadable identity is deliberately confined here"` (`internal/core/service_task.go:145-148`). However, `FS.ReadTaskGraph` now also extracts unreadable identity directly via `splitFlatName` (`internal/store/fsstore.go:126`).

This creates two implementations of the filename-to-identity extraction rule: `store.splitFlatName` (`internal/store/flatname.go:28`) and `core.taskIdentityFromPath` (`internal/core/dependency_graph.go:537`). While identical today, drift between them could cause mismatched `TaskID` attribution on unreadable records between ordinary listing and graph analysis. The comment in `service_task.go` should be updated and the extraction logic unified.

---

**Resolution:** Added one core file-diagnostic conversion helper used by both
the legacy list adapter and the filesystem graph reader, eliminating duplicate
filename identity parsing.

### 5. Acceptance-Criteria Traceability and Evidence Ledgers

#### 5.1 Acceptance Criteria Traceability

| Acceptance Criterion | Status | Implementation Evidence | Verification Tests |
| :--- | :--- | :--- | :--- |
| **AC1:** Two scans of byte-identical unreadable task content compare as the same source snapshot. | **Met** | `internal/store/resolve.go:59`, `fsstore.go:120`, `dependency_graph.go:325,901` | `internal/store/graphmutation_test.go:287-293`, hostile fixture probe `unchanged unreadable bytes` |
| **AC2:** Changing an unreadable task's bytes makes the snapshots differ even when its path, recovered identity, problem code, and error message are unchanged. | **Met** | `internal/store/resolve.go:59`, `dependency_graph.go:907-911` | `internal/store/graphmutation_test.go:261-280`, `internal/core/dependency_graph_test.go:765-769`, hostile fixture probe `changed bytes with same error` |
| **AC3:** Transitions between readable and unreadable representations always invalidate the snapshot. | **Met** | `internal/core/dependency_graph.go:890,901` (`health` and `ids` mismatch, plus load problem length mismatch) | `internal/core/dependency_graph_test.go:777-789`, hostile fixture probes `readable to unreadable transition` and `unreadable to readable transition` |
| **AC4:** Pathless adapters can supply their own opaque revision without fabricating a local path, and core never exposes the token as task-domain or public wire data. | **Met** | `internal/core/service_task.go:109` (`json:"-" yaml:"-"`), `dependency_graph.go:806` (`task.SourceVersion = ""`), `internal/wire/dto.go:45` | `internal/core/dependency_graph_test.go:740-756`, `schema --json-schema` inspection, hostile fixture probe `pathless records and reordering` |
| **AC5:** Existing ordinary graph reads and mutations retain their behavior; focused race tests prove a stale broken-graph repair would receive `ErrConflict` before any write. | **Met** | `internal/store/graphmutation.go:87,117` (`verifyTaskGraphSourceSnapshot`) | `internal/store/graphmutation_test.go:278-281`, uncached `go test -race ./...`, repeated `-race` runs |

#### 5.2 Hostile Evidence Ledger

| # | Hostile Scenario | Test Action & Condition | Observed Behavior | Verdict |
| :- | :--- | :--- | :--- | :--- |
| 1 | **Unchanged unreadable bytes** | Two consecutive `LoadTaskGraph(fs)` reads of the same unreadable file (`# broken frontmatter`). | `g1.SameSourceSnapshot(g2)` returned `true`; `verifyTaskGraphSourceSnapshot` returned `nil`. | **PASS** |
| 2 | **Changed bytes with same error** | Unreadable file modified from `"broken 1"` to `"broken 2"` (same parse error and diagnostic). | `g1.SameSourceSnapshot(g2)` returned `false`; `verifyTaskGraphSourceSnapshot` returned `domain.ErrConflict`. | **PASS** |
| 3 | **Changed body outside malformed frontmatter** | Malformed YAML frontmatter (`bad: [`) kept identical while markdown body changed (`# Body 1` -> `# Body 2`). | Diagnostic error was identical; `hashContent` differed; `SameSourceSnapshot` returned `false`; `ErrConflict` returned. | **PASS** |
| 4 | **Readable to unreadable transition** | Valid task file replaced with corrupt content lacking frontmatter. | `g1.health` (`Healthy`) vs `g2.health` (`Broken`), `ids` length changed; `SameSourceSnapshot` returned `false`; `ErrConflict` returned. | **PASS** |
| 5 | **Unreadable to readable transition** | Corrupt task file repaired with valid task content. | `g1.health` (`Broken`) vs `g2.health` (`Healthy`); `SameSourceSnapshot` returned `false`; `ErrConflict` returned. | **PASS** |
| 6 | **Add unreadable file** | Added new corrupt task file to existing healthy repository. | `loadProblems` length increased from 0 to 1; `SameSourceSnapshot` returned `false`. | **PASS** |
| 7 | **Remove unreadable file** | Removed corrupt task file from repository. | `loadProblems` length decreased from 1 to 0; `SameSourceSnapshot` returned `false`. | **PASS** |
| 8 | **Rename unreadable file** | Renamed `tasks/alpha-old.md` to `tasks/alpha-new.md` with identical corrupt content. | Path changed; `sameTaskGraphLoadProblem` failed; `SameSourceSnapshot` returned `false`. | **PASS** |
| 9 | **Non-ID-led files** | Added `tasks/stray.md` without ID prefix; modified bytes in subsequent scan. | Non-ID file produced `TaskGraphLoadProblem` with `TaskID: ""`; byte modification altered `SourceVersion`; `SameSourceSnapshot` returned `false`. | **PASS** |
| 10 | **Duplicate unreadable identities** | Two corrupt task files with identical ID prefix: `tasks/dup-alpha.md` and `tasks/dup-beta.md`. | Deterministically sorted by composite key; byte edit in `dup-alpha.md` invalidated snapshot. | **PASS** |
| 11 | **Pathless records & reordering** | In-memory `TaskGraphRead` with `Path: ""` supplied in reverse order `[B, A]` vs `[A, B]`. | `newTaskGraph` sorted problems stably; `SameSourceSnapshot` returned `true`; altering `probB.SourceVersion` returned `false`. | **PASS** |
| 12 | **Missing tokens fail closed** | Two identical problem records provided with `SourceVersion: ""`. | `sameTaskGraphLoadProblem` checked `left.SourceVersion != ""` and returned `false`; `SameSourceSnapshot` failed closed. | **PASS** |

#### 5.3 Restored Mutation Ledger

| Mutation ID | Target File & Code Modification | Intended Defect | Killing Test | Observed Failure Output | Restored Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **M1** | `internal/core/dependency_graph.go:909`<br>Replaced `left.SourceVersion != "" && left.SourceVersion == right.SourceVersion` with `true`. | Omit unreadable revision comparison in snapshot equality check. | `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions`<br>`TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` | `dependency_graph_test.go:768: changed unreadable source revision compared as the same snapshot`<br>`graphmutation_test.go:279: stale unreadable-source snapshot error = <nil>, want ErrConflict` | **Restored** |
| **M2** | `internal/store/resolve.go:59`<br>Replaced `hashContent(content)` with `hashContent([]byte(path))`. | Derive revision from path/message rather than exact source bytes. | `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` | `graphmutation_test.go:258: unreadable source version = "c8df74d09415..."` (failed hash expectation on modified content) | **Restored** |
| **M3** | `internal/store/resolve.go:59`<br>Replaced `sourceVersion = hashContent(content)` with `sourceVersion = ""`. | Filesystem adapter returns an empty token for unreadable sources. | `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` | `graphmutation_test.go:258: unreadable source version = ""` | **Restored** |
| **M4** | `internal/core/service_task.go:109`<br>Removed `json:"-"` struct tag from `TaskGraphLoadProblem.SourceVersion`. | Expose source revision token through JSON serialization. | `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions` | `dependency_graph_test.go:755: opaque source revision leaked through JSON: {"TaskID":"...","SourceVersion":"opaque-revision-a"}` | **Restored** |
| **M5** | `internal/core/dependency_graph.go:326-331`<br>Removed `sort.SliceStable` over `g.loadProblems`. | Ignore adapter ordering and rely on raw scan sequence. | `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions` | `dependency_graph_test.go:762: identical pathless unreadable sources compared as different snapshots` | **Restored** |
| **M6** | `internal/core/dependency_graph.go:325`<br>Replaced `g.loadProblems = cloneTaskGraphLoadProblems(unreadable)` with `g.loadProblems = nil`. | Drop unreadable problems from graph snapshot. | `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS`<br>`TestTaskGraphSameSourceSnapshotComparesUnreadableIdentity` | `graphmutation_test.go:279: stale unreadable-source snapshot error = <nil>, want ErrConflict`<br>`dependency_graph_test.go:738: different unreadable task sets compared as the same source snapshot` | **Restored** |

---

### 6. Separation of Defects, Risks, and Concerns

- **Demonstrated Defects:**
  None. The implementation functions as specified with no regressions across the test suite.
- **Source-Supported Risks:**
  1. *Uncooperative Raw Editor Race Window:*
     In `FS.MutateTaskGraph`, the whole-snapshot preflight CAS (`verifyTaskGraphSourceSnapshot`) verifies that neither healthy tasks nor unreadable sources have changed since planning. Before each individual file modification, `verifyUnchanged` validates the per-target content hash immediately prior to `writeFileAtomic`. An uncooperative process that directly writes or truncates a file without acquiring flock could theoretically race between `verifyUnchanged` (file read) and `writeFileAtomic` (atomic rename). This window is inherent to POSIX filesystems and bounded to the single atomic file replacement; it is documented accurately in `ARCHITECTURE.md`.
  2. *Fallback Adapter CAS Ineligibility:*
     Non-conforming stores wrapped by `taskStoreGraphSource` produce `TaskGraphRead` values via `TaskGraphReadFromFiles`, leaving `SourceVersion: ""` on unreadable records. Consequently, `SameSourceSnapshot` fails closed if unreadable records are present. This behavior is deliberate per ADR-0006, ensuring that stores without byte-level CAS capabilities cannot authorize broken-graph repair writes.
- **Unverified Concerns:**
  1. *Directory Scan Performance Overhead:*
     Calculating SHA-256 for task files during `scanDirWithSourceVersions` adds negligible overhead (microseconds for typical markdown files under 100 KB). Furthermore, ordinary `ListTasks` invocations explicitly use `scanDirSources(..., versionProblems: false)`, completely bypassing hash calculation for skipped files.

---

### 7. Settled Concerns

1. **Concurrency and OCC Boundary Truthfulness:**
   The implementation places `verifyTaskGraphSourceSnapshot` squarely on the production pre-write path of `FS.MutateTaskGraph` (line 87) and returns `domain.ErrConflict` on mismatch. This eliminates the vulnerability where concurrent edits to unreadable files went undetected if the diagnostic message remained unchanged.
2. **Wire Compatibility and Token Leakage:**
   `SourceVersion` is explicitly unexported from machine contracts via `json:"-" yaml:"-"` on `TaskGraphLoadProblem`, while `TaskGraph.Task(id)` zeroes `SourceVersion` before returning projections. CLI commands (`task list --json`, `lint --json`, `status --json`) and `schema --json-schema` were empirically verified to emit no revision tokens.
3. **Portability of Pathless Adapters:**
   `TaskGraphLoadProblem` treats `Path` as optional diagnostic context. Non-filesystem adapters can provide arbitrary opaque revisions without inventing synthetic filesystem paths.
4. **Adapter Ordering Independence:**
   Deterministic composite sorting (`TaskID \x00 TaskSlug \x00 Path \x00 Message \x00 SourceVersion`) ensures that variations in directory walk or remote query ordering do not trigger spurious CAS conflicts.
5. **Fail-Closed Missing Revision Policy:**
   `sameTaskGraphLoadProblem` explicitly requires `left.SourceVersion != ""`, ensuring unversioned problems cannot produce false equality.
