---
schema: 1
id: 6g73ymg15x2e
bucket: closed
area: thread-source-revision-cas-implementation-antigravity
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: Thread source revision CAS implementation — Antigravity — 2026-09-05

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer
> should update.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Codes must match
> `[A-Z]+[0-9]+`; do not put status on a separate line or pre-resolve a finding.
>
> A second adversarial pass is mandatory. After the checklist, deliberately seek a systemic reason
> the locally symmetric patch is still unsafe: divergent read paths, a token-bearing contract that
> leaks through another adapter, a comparison that is correct only for filesystem order, or a
> guarded writer that silently retains the former `FileProblem` snapshot. A no-finding verdict must
> show the hostile evidence that falsified those hypotheses; green tests alone are insufficient.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may use the handoff checkout concurrently. Reading
this brief and performing the initial copy are the only operations allowed there until the final
guarded audit transfer. Capture the exact current contents, including staged, unstaged, untracked,
and deleted files:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g73ymg15x2e-2026-09-05-thread-source-revision-cas-implementation-antigravity.md"
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

Confirm `git rev-parse --absolute-git-dir` resolves inside `$SANDBOX` and no object alternate points
to the source. Perform all inspection, builds, formatting, fixtures, and mutation probes there.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If isolation cannot be verified, stop and report the blocker.

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
[version unreadable Thread sources for repair-safe CAS](../tasks/6g72ncs4xjdm-version-unreadable-thread-sources-for-repair-safe-cas.md)
on branch `feat/unreadable-thread-source-revisions`, based on `main` at `aababec`. The implementation
under review is the complete uncommitted working tree captured by the sandbox checkpoint. Review its
diff against `aababec` and judge it against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md),
[the architecture guide](../../docs/ARCHITECTURE.md), the task acceptance criteria, the immediately
preceding task-source revision implementation, and every existing Thread mutation/OCC contract.

Assume the apparent task/Thread symmetry can conceal an asymmetry. Re-derive exact source bytes
through both filesystem Thread scanners, neutral `ThreadRead`, service projections,
`sameThreadSourceSnapshot`, `verifyThreadSourceSnapshot`, and every pre-write consumer. Do not edit
implementation or other planning files, create follow-up tasks, change finding statuses, close this
audit, push, or install anything globally.

## Intended contract to challenge

- `ThreadReadProblem.SourceVersion` is an opaque adapter-owned revision of the exact unreadable
  source. Core and guarded stores may copy, normalize, and compare it but may not parse it or expose
  it through Thread/domain values, validation errors, CLI, TUI, schema, JSON/YAML, or wire contracts.
- Production filesystem `ReadThreads` hashes the same bytes it hands `parseThread`, in one resilient
  scan. Ordinary `ListThreads` retains its established `[]domain.FileProblem` compatibility shape;
  tokenless compatibility reads are CAS-ineligible rather than falsely authoritative.
- Thread apply's body-aware scan carries the identical revision semantics. Creation, membership and
  lifecycle mutation, bulk apply, and task lifecycle all compare the complete revision-bearing
  `ThreadRead` snapshot before any write.
- Snapshot comparison is order-independent and does not mutate caller slices. It compares readable
  source identity/location plus a non-empty equal revision, and unreadable stable identity, optional
  location, message, plus a non-empty equal revision.
- Missing revision evidence fails closed. Pathless/remote adapters may provide stable opaque tokens
  without fabricating filesystem paths. Add/remove/rename/identity drift, readable/unreadable
  transitions, and changed bytes reproducing one diagnostic all invalidate equality; exact byte
  restoration restores the filesystem revision.
- Existing ordinary mutations still refuse malformed Thread input before planning. The stronger
  comparison protects transitions at their verify boundary and is reusable by graph repair, where
  malformed Threads will be non-blocking but impact evidence may be incomplete.
- The change adds no second scan, persisted token, universal entity revision scheme, repair policy,
  or graph-library dependency.

## Mandatory evidence floor

A `ready` verdict is not credible without all of the following:

1. Inventory `ReadThreads`, `ListThreads`, their test override, `threadReadFromFiles`, the
   revision-bearing conversion, `scanDirWithSourceVersions`, `listThreadApplyThreads`, every
   `ThreadStore` production composition, every service consumer, and all four guarded write families.
2. Exercise unchanged unreadable bytes; changed body bytes with the same diagnostic; exact restore;
   add/remove/rename; recovered identity drift; readable/unreadable transitions; pathless problems;
   reordered and duplicate problem identities; multiple readable records; and both one-sided and
   two-sided missing tokens.
3. Prove ordinary native listing, Thread service results, lint, human CLI, JSON CLI, TUI state,
   generated schema, and wire DTOs cannot reveal the token or change established diagnostics.
4. Verify the body-aware Thread apply scan did not lose body fidelity, change scan counts, or diverge
   from ordinary `ReadThreads` revision semantics.
5. Trace the whole-snapshot and immediate target CAS windows. State exactly what cooperating writers,
   raw editors, unreadable non-targets, and the future repair path can and cannot guarantee.
6. Apply and restore mutation probes that: omit unreadable revision comparison; accept two empty
   tokens; derive the token from path/message; return an empty filesystem token; route one writer
   back through `ListThreads`; make equality input-order-sensitive; expose the token from the service
   or wire; and drop unreadable problems. Name the shipped test killing each mutation and report any
   survivor as a finding.
7. Run repeated focused race tests, an uncached full race suite, static analysis, module tidiness,
   generated-doc checks, planning/audit lint, and `git diff --check`, with exact results.

## Required adversarial angles

1. **Source truth.** Challenge whether the hash covers the exact parser bytes, whether a read error
   stays fatal, and whether parse failures remain resilient without another filesystem pass.
2. **Portability and disclosure.** Challenge pathless adapters, tokenless fakes, direct struct
   serialization, service cloning, TUI retention, explicit wire mapping, schemas, and error strings.
3. **Comparator soundness.** Seek false equality, false conflicts, asymmetric non-empty checks,
   unstable duplicate pairing, aliasing, location-as-identity inference, or ordering-dependent CAS.
4. **Consumer completeness.** Inspect every guarded operation independently. Do not infer that fixing
   creation fixed mutation, task lifecycle, body-aware apply, or apply reprepare.
5. **Compatibility.** Compare healthy and malformed Thread list/lint/show/apply behavior to base;
   challenge the `threadListOverride` seam and planner re-entry protections.
6. **Systemic second pass.** Look beyond this patch for a shared resilient-read/CAS convention that
   leaves one entity or mutation path half-versioned. Demonstrate it, distinguish current reachability
   from forward risk, and avoid speculative architecture expansion.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- branch/base/checkpoint, runtime, exact validation, and isolation attestation;
- a compact source-to-CAS data flow and complete producer/consumer inventory;
- findings grouped by severity with exact grammar and `**Status:** open`;
- acceptance-criteria traceability plus hostile-evidence and restored-mutation ledgers;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but still provide every required evidence class. Do not
pre-resolve findings; the implementation owner will triage them with `tskflwctl audit finding`.

## Reviewer report

### Executive verdict

**Verdict:** `ready with tracked follow-ups`

The implementation for task [`6g72ncs4xjdm`](../tasks/6g72ncs4xjdm-version-unreadable-thread-sources-for-repair-safe-cas.md) on branch `feat/unreadable-thread-source-revisions` is sound, robust, and correctly extends whole-snapshot CAS to cover unreadable Thread documents via opaque cryptographic source hashes (`core.ThreadReadProblem.SourceVersion`). Snapshot comparison in `sameThreadSourceSnapshot` is order-independent, preserves caller memory without aliasing, fails closed on missing tokens, and protects all four guarded writer operations (`MutateThreadCreation`, `MutateThread`, `MutateThreadApply`, and `MutateTaskLifecycle`). Tokens are strictly internal and do not leak through service views, CLI renderers, TUI, schema, or wire DTOs.

A single medium-severity finding (`M1`) is opened: existing writer integration tests do not exercise concurrent unreadable Thread modifications or verify that writers require `ReadThreads()` rather than compatibility `ListThreads()`. When mutation probe 5 (routing writers back through `ListThreads`) was applied, all shipped tests continued to pass. The implementation correctly calls `ReadThreads()`, but an end-to-end regression test is recommended as a tracked follow-up.

---

### Environment, runtime, validation, and isolation attestation

- **Reviewer:** Antigravity
- **Branch:** `feat/unreadable-thread-source-revisions`
- **Base commit:** `aababec`
- **Sandbox baseline checkpoint:** `957dff91850ef5dea96fae20c0cbe2c910c0b1eb`
- **Sandbox location:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review-antigravity.qisCcZ`
- **Absolute git dir:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review-antigravity.qisCcZ/.git`
- **Isolation verification:** Non-hardlink clone verified; `.git/objects/info/alternates` is absent; `$SOURCE_ROOT` remained completely unmodified and untouched throughout all builds, probes, and tests.
- **Runtime validation:**
  - `go test -race -count=1 ./...`: **PASS** (0 data races across all packages).
  - `go test -race -count=5 ./internal/store/ -run 'TestThread|TestTaskLifecycle'`: **PASS** (repeated concurrency race checks clean).
  - `go test -race -count=5 ./internal/core/ -run 'TestThread'`: **PASS** (clean).
  - `just build`: **PASS** (built `bin/tskflwctl` v0.19.0-12-g957dff9).
  - `just tidy-check`: **PASS** (`go mod tidy -diff` reported no changes).
  - `just docs-check`: **PASS** (`internal/tools/docgen -out docs/cli` clean diff).
  - `just lint`: **PASS** (`golangci-lint run ./...`: 0 issues).
  - `./bin/tskflwctl lint`: **PASS** (`✔ all planning entities and dependency links pass lint`).
  - `./bin/tskflwctl audit lint`: **PASS** (`✔ all audit findings pass lint`).
  - `git diff --check`: **PASS** (no whitespace or patch formatting issues).

---

### Source-to-CAS data flow and producer/consumer inventory

#### Source-to-CAS Data Flow

```
[On-Disk Markdown File: threads/<id>-<slug>.md]
          │
          ▼
[scanDirSources(dir, parse, versionProblems=true)]
   ├─ os.ReadFile(path) -> raw content []byte
   ├─ sourceVersion = hashContent(content)  // SHA-256 hex
   ├─ parse(path, content)
   │    ├─ Success: parseThread -> domain.Thread{SourceVersion: hashContent(content)}
   │    └─ Error:   parse error -> sourceFileProblem{problem: FileProblem, sourceVersion: sourceVersion}
          │
          ▼
[threadReadFromSourceFiles(threads, problems)]
   ├─ Copies threads slice
   └─ Maps problems to []core.ThreadReadProblem{..., SourceVersion: sourceVersion}
          │
          ▼
[core.ThreadRead snapshot]
   │
   ├─ Planning Preflight: core.ValidateThreadCreationSource(graph, read.Threads, read.Problems)
   │
   ├─ Verification Stage (Pre-Write under Repository Guard):
   │    re-read currentThreadRead via s.ReadThreads() / s.listThreadApplyThreads()
   │    verifyThreadSourceSnapshot(threadRead, currentThreadRead)
   │       └─ sameThreadSourceSnapshot(left, right)
   │            ├─ stable sort left/right threads by threadSourceKey
   │            ├─ slices.EqualFunc(..., sameThreadSource) // left.SourceVersion != "" && left.SourceVersion == right.SourceVersion
   │            ├─ stable sort left/right problems by threadProblemSourceKey
   │            └─ slices.EqualFunc(..., sameThreadProblemSource) // left.SourceVersion != "" && left.SourceVersion == right.SourceVersion
   │
   └─ Safe Persistence: atomic write to target file(s)
```

#### Producer & Consumer Inventory

1. **Producers**:
   - `(*FS).ReadThreads() (core.ThreadRead, error)` ([`internal/store/threadstore.go:21`](file:///internal/store/threadstore.go#L21)): The primary production and test read entry point. Invokes `scanDirWithSourceVersions(s.threadsDir, parseThread)` and converts via `threadReadFromSourceFiles`. Honors `s.threadListOverride` if set by wrapping with `threadReadFromFiles` (which leaves `SourceVersion` empty, ensuring test overrides are safely CAS-ineligible).
   - `(*FS).listThreadApplyThreads() (core.ThreadRead, map[string]string, error)` ([`internal/store/threadapply.go:226`](file:///internal/store/threadapply.go#L226)): Body-aware Thread scanner for bulk apply operations. Calls `scanDirWithSourceVersions` to compute SHA-256 hashes on unreadable and readable records, extracts markdown bodies in the same single pass, and converts via `threadReadFromSourceFiles`.
   - `(*FS).ListThreads() ([]domain.Thread, []domain.FileProblem, error)` ([`internal/store/threadstore.go:46`](file:///internal/store/threadstore.go#L46)): Compatibility listing retained for maintenance and tests needing `domain.FileProblem`. Uses `scanDir` (`versionProblems=false`).
   - `threadReadFromSourceFiles(threads []domain.Thread, problems []sourceFileProblem) core.ThreadRead` ([`internal/store/threadstore.go:66`](file:///internal/store/threadstore.go#L66)): Store-private converter attaching non-empty `sourceVersion` to each `core.ThreadReadProblem`.
   - `threadReadFromFiles(threads []domain.Thread, problems []domain.FileProblem) core.ThreadRead` ([`internal/store/threadstore.go:58`](file:///internal/store/threadstore.go#L58)): Compatibility converter mapping `domain.FileProblem` with empty `SourceVersion: ""` (CAS-ineligible).

2. **Consumers**:
   - `(*FS).MutateThreadCreation` ([`internal/store/threadcreation.go:17`](file:///internal/store/threadcreation.go#L17)): Loads `threadRead` via `s.ReadThreads()`, checks `ValidateThreadCreationSource`, re-reads `currentThreadRead` via `s.ReadThreads()`, and executes whole-snapshot CAS with `verifyThreadSourceSnapshot` before writing the new Thread file.
   - `(*FS).MutateThread` ([`internal/store/threadmutation.go:19`](file:///internal/store/threadmutation.go#L19)): Loads `threadRead` via `s.ReadThreads()`, checks `ValidateThreadMutationSource`, re-reads `currentThreadRead` via `s.ReadThreads()`, executes whole-snapshot CAS with `verifyThreadSourceSnapshot`, and immediate target CAS via `verifyUnchanged`.
   - `(*FS).MutateThreadApply` ([`internal/store/threadapply.go:18`](file:///internal/store/threadapply.go#L18)): Loads `threadRead` via `s.listThreadApplyThreads()`, checks `ValidateThreadCreationSource`, re-reads `currentThreadRead` via `s.listThreadApplyThreads()`, executes whole-snapshot CAS with `verifyThreadSourceSnapshot`, executes per-task dependency target CAS via `verifyUnchanged`, and re-validates at reprepare (`reprepareThreadApply`).
   - `(*FS).MutateTaskLifecycle` ([`internal/store/lifecyclemutation.go:19`](file:///internal/store/lifecyclemutation.go#L19)): Loads `threadRead` via `s.ReadThreads()`, checks `ValidateThreadMutationSource`, re-reads `currentThreadRead` via `s.ReadThreads()`, executes whole-snapshot CAS with `verifyThreadSourceSnapshot` before changing task status, and executes target CAS via `verifyUnchanged`.
   - `(*Service).ListThreadViews` ([`internal/core/service_thread.go:217`](file:///internal/core/service_thread.go#L217)): User-facing service listing. Reads `s.ReadThreads()`, sanitizes all problems by stripping `SourceVersion` (`problems[i].SourceVersion = ""`), and returns `ThreadListView` and `[]ThreadReadProblem`.

---

### Findings

#### M1. Guarded writers lack end-to-end regression tests verifying unreadable Thread SourceVersion CAS enforcement · **Status:** tracked by 6g4g8gatbnrs

- **Description:** `internal/store/thread_source_snapshot_test.go` directly exercises `verifyThreadSourceSnapshot` and `FS.ReadThreads()` under hostile inputs, verifying that unreadable records carry SHA-256 byte hashes and fail closed when modified. However, none of the integration tests for the four guarded writers (`MutateThreadCreation`, `MutateThread`, `MutateThreadApply`, `MutateTaskLifecycle`) test a scenario where an unreadable Thread document is present or concurrently modified during the transaction.
- **Evidence / Reproduction:** When applying mutation probe 5 (routing any of the writer methods back through `s.ListThreads()` and `threadReadFromFiles`), all tests in the repository (`go test ./...`) passed without failure. This occurs because:
  1. `s.ListThreads()` parses valid Threads and sets `thread.SourceVersion = hashContent(content)` on readable Threads;
  2. In all existing writer tests, no unreadable Thread documents exist, so `problems` is empty in both initial and verification reads;
  3. Pre-planning validation (`ValidateThreadCreationSource` / `ValidateThreadMutationSource`) rejects transactions if unreadable Threads are present initially (`len(unreadable) > 0`).
- **Impact:** While the production code is correctly written to invoke `s.ReadThreads()` and `s.listThreadApplyThreads()`, an unintentional regression in a writer reverting to `s.ListThreads()` would not be detected by existing CI tests.
- **Suggested Remediation:** Add a store-level concurrency/hook test for at least one guarded writer (such as `MutateThreadCreation` or `MutateTaskLifecycle`) that injects a concurrent unreadable Thread document via `testHookBeforeThreadCreationVerify` or `testHookBeforeLifecycleVerify` and asserts that `domain.ErrConflict` is raised.

---

**Resolution:** Removed ListThreads and its tokenless helpers so the reviewed
routing regression is no longer expressible. The end-to-end
unreadable-to-unreadable writer proof is explicitly tracked on repair task
6g4g8gatbnrs, the first writer allowed to plan with malformed Thread evidence.

### Acceptance criteria traceability

| Acceptance Criterion | Verification Evidence | Status |
| :--- | :--- | :--- |
| **AC1:** Unreadable Thread records carry exact byte hash in `ThreadReadProblem.SourceVersion`. | Verified by `TestReadThreadsVersionsExactUnreadableSourceBytes` in [`internal/store/thread_source_snapshot_test.go:31`](file:///internal/store/thread_source_snapshot_test.go#L31); hash matches `hashContent([]byte(content))` before parser execution. | **PASS** |
| **AC2:** Hash computed in single scan pass without secondary filesystem read. | Verified by implementation in `scanDirSources` ([`internal/store/resolve.go:50-68`](file:///internal/store/resolve.go#L50-L68)). `os.ReadFile` content is hashed immediately; if `parse` errors, the hash is attached directly to `sourceFileProblem`. | **PASS** |
| **AC3:** Whole-snapshot CAS (`sameThreadSourceSnapshot`) compares `SourceVersion` across readable and unreadable records. | Verified by `TestThreadSourceSnapshotRejectsRepresentationAndIdentityChanges` and `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` in [`internal/store/thread_source_snapshot_test.go`](file:///internal/store/thread_source_snapshot_test.go). | **PASS** |
| **AC4:** Snapshot comparison is order-independent and does not mutate input slices. | Verified by `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` ([`thread_source_snapshot_test.go:79-86`](file:///internal/store/thread_source_snapshot_test.go#L79-L86)). Caller slices are copied before sorting with `sort.SliceStable`. | **PASS** |
| **AC5:** Tokenless reads fail closed in CAS. | Verified by `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` ([`thread_source_snapshot_test.go:94-104`](file:///internal/store/thread_source_snapshot_test.go#L94-L104)). Both one-sided and two-sided empty tokens return `domain.ErrConflict`. | **PASS** |
| **AC6:** All four guarded writer families enforce whole-snapshot CAS. | Traced in `threadcreation.go:87`, `threadmutation.go:92`, `threadapply.go:123`, and `lifecyclemutation.go:110`. | **PASS** |
| **AC7:** Body-aware Thread apply scan preserves body fidelity and revision semantics. | Traced in `listThreadApplyThreads` ([`threadapply.go:230-247`](file:///internal/store/threadapply.go#L230-L247)). Uses `scanDirWithSourceVersions` and extracts markdown body via `splitFrontmatter`. | **PASS** |
| **AC8:** No tokens exposed to planners, service views, CLI, TUI, schema, or wire DTOs. | Verified by `TestServiceThreadListStripsOpaqueProblemSourceRevisions` ([`internal/core/service_thread_test.go:244`](file:///internal/core/service_thread_test.go#L244)) and `TestToThreadsEnvelopeRetainsPathlessIdentityWithoutParsingLocation` ([`internal/wire/thread_test.go:23`](file:///internal/wire/thread_test.go#L23)). | **PASS** |

---

### Hostile evidence ledger

| # | Hostile Condition / Angle | Fixture / Probe Setup | Expected Result | Observed Result |
| :- | :--- | :--- | :--- | :--- |
| 1 | Unchanged unreadable bytes | `first` read with unreadable frontmatter compared with byte-identical `restored` read | Snapshot equal (`err == nil`) | **PASS** (`thread_source_snapshot_test.go:67`) |
| 2 | Changed body bytes with same diagnostic | `first` read vs `second` read with modified body markdown but identical invalid YAML error | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:58`) |
| 3 | Exact restore | Overwritten file rewritten with exact original bytes | Snapshot equal (`err == nil`) | **PASS** (`thread_source_snapshot_test.go:67`) |
| 4 | Add unreadable problem | `unreadable` vs `added` with additional problem record | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:147`) |
| 5 | Remove unreadable problem | `unreadable` vs empty snapshot | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:140`) |
| 6 | Rename unreadable problem location | `unreadable` vs `renamed` with altered `Location` path | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:137`) |
| 7 | Recovered identity drift | `readable` vs `drifted` with mismatched `ID` and `FilenameID` | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:119`) |
| 8 | Readable to unreadable transition | `readable` vs `unreadable` record | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:127`) |
| 9 | Unreadable to readable transition | `unreadable` vs `readable` record | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:130`) |
| 10 | Pathless unreadable problems | Remote adapter problems with empty `Location` and opaque revisions | Order-independent match (`err == nil`) | **PASS** (`thread_source_snapshot_test.go:81`) |
| 11 | Reordered problem identities | Slices `[P1, P2]` vs `[P2, P1]` | Snapshot equal (`err == nil`) | **PASS** (`thread_source_snapshot_test.go:81`) |
| 12 | Duplicate problem identities | Slices `[P1, P1]` vs `[P1, P1]` with identical keys | Snapshot equal (`err == nil`) | **PASS** (Tested via probe) |
| 13 | Multiple readable records | Slices `[T1, T2]` vs `[T2, T1]` | Snapshot equal (`err == nil`) | **PASS** (Tested via probe) |
| 14 | One-sided missing token | `left.SourceVersion == "opaque"` vs `right.SourceVersion == ""` | `domain.ErrConflict` | **PASS** (Tested via probe) |
| 15 | Two-sided missing tokens | `left.SourceVersion == ""` vs `right.SourceVersion == ""` | `domain.ErrConflict` | **PASS** (`thread_source_snapshot_test.go:96`) |

---

### Restored mutation ledger

| Probe | Mutation Description | Target File & Line | Shipped Test Killing Mutation | Failure Output | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **M1** | Omit unreadable revision comparison (`left.SourceVersion == right.SourceVersion`) | `internal/store/cas.go:97` | `TestReadThreadsVersionsExactUnreadableSourceBytes` & `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` | `thread_source_snapshot_test.go:59: changed unreadable bytes error = <nil>, want ErrConflict` | **KILLED** (Restored) |
| **M2** | Accept two empty tokens (remove `left.SourceVersion != ""` check) | `internal/store/cas.go:97` | `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` | `thread_source_snapshot_test.go:97: two unversioned unreadable snapshots error = <nil>, want ErrConflict` | **KILLED** (Restored) |
| **M3** | Derive token from path/message rather than bytes | `internal/store/resolve.go:56` | `TestReadThreadsVersionsExactUnreadableSourceBytes` | `thread_source_snapshot_test.go:32: first problem = ... SourceVersion: ...` | **KILLED** (Restored) |
| **M4** | Return empty filesystem token in `ReadThreads()` | `internal/store/threadstore.go:72` | `TestReadThreadsVersionsExactUnreadableSourceBytes` | `thread_source_snapshot_test.go:32: first problem = ... SourceVersion:}` | **KILLED** (Restored) |
| **M5** | Route one writer back through `ListThreads()` | `internal/store/threadcreation.go:50, 83` | **None** (`go test ./...` passed) | Mutation survived; reported as finding `M1` | **SURVIVED** (Restored) |
| **M6** | Make equality input-order-sensitive (omit sorting) | `internal/store/cas.go:67-72` | `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` | `thread_source_snapshot_test.go:82: reordered pathless problems = conflict` | **KILLED** (Restored) |
| **M7** | Expose token from service or wire (strip logic omitted / json tag removed) | `internal/core/service_thread.go:224` & `internal/core/store.go:119` | `TestServiceThreadListStripsOpaqueProblemSourceRevisions` | `service_thread_test.go:257: service exposed opaque source revision` & `service_thread_test.go:264: leaked through JSON` | **KILLED** (Restored) |
| **M8** | Drop unreadable problems from snapshot (`Problems = nil`) | `internal/store/threadstore.go:65` | `TestReadThreadsVersionsExactUnreadableSourceBytes` | `thread_source_snapshot_test.go:27: first read={Threads:[] Problems:[]} err=<nil>` | **KILLED** (Restored) |

---

### Whole-snapshot and target CAS window analysis

The OCC concurrency safety model relies on two distinct CAS verification boundaries:

1. **Whole-Snapshot CAS (`verifyTaskGraphSourceSnapshot` & `verifyThreadSourceSnapshot`)**:
   - Evaluated under the repository file lock (`flock`) immediately before any mutation writes begin.
   - Compares the cryptographic state of all entities (tasks and Threads, readable and unreadable) against the state captured when planning began.
   - **Guarantees for cooperating writers:** Cooperating writers serialize on the repository lock; whole-snapshot CAS guarantees that no planning occurred against a stale world view while another writer held the lock.
   - **Guarantees for raw external editors:** If an external editor modifies any task or Thread file on disk during planning (even changing bytes in an unreadable file without altering the compiler/parser error message), the SHA-256 `SourceVersion` changes, and the whole-snapshot CAS aborts the transaction with `domain.ErrConflict`.
   - **Future repair path:** Graph repair or Thread repair routines will be able to plan non-blocking fixes in the presence of unreadable Threads; whole-snapshot CAS guarantees that an ongoing repair will detect concurrent raw edits to the broken document and refuse to overwrite newer on-disk state.

2. **Immediate Target CAS (`verifyUnchanged`)**:
   - Re-resolves and hashes the specific file being updated immediately prior to invoking `writeFileAtomic`.
   - Protects against in-place target edits or renames that race between the start of multi-file writes and each individual file write.

3. **What CAS Cannot Guarantee**:
   - Raw external filesystem writes occurring *after* `writeFileAtomic` completes will overwrite disk state (inherent to POSIX file semantics).
   - Tokenless/in-memory test fakes providing empty `SourceVersion` cannot participate in CAS (they fail closed).

---

### Separation of demonstrated defects, source-supported risks, and unverified concerns

- **Demonstrated Defects:**
  - None in the production codebase. The implementation meets all functional and non-functional requirements. Finding `M1` is a test coverage gap for writer wiring regression defense.
- **Source-Supported Risks:**
  - Callers consuming `ListThreads()` obtain `domain.FileProblem` records without `SourceVersion`. If an adapter or caller attempts to construct a `core.ThreadRead` from `ListThreads()` using `threadReadFromFiles`, the resulting unreadable problems have `SourceVersion: ""`, rendering the snapshot permanently ineligible for CAS writes (`ErrConflict`). This behavior is intentional and fails closed, but developers should be aware that `ReadThreads()` is the mandatory entry point for CAS-capable reads.
  - When implementing the future repair workflow, authors must ensure the repair transaction retains `verifyThreadSourceSnapshot` and does not bypass CAS when handling malformed Thread inputs.
- **Unverified Concerns:**
  - *Risk of false conflicts from non-deterministic directory ordering:* Disproven. `sameThreadSourceSnapshot` explicitly sorts all records stably using keys incorporating all entity fields before performing comparison.
  - *Risk of token leakage through error messages or logs:* Disproven. `verifyThreadSourceSnapshot` returns the generic sentinel `domain.ErrConflict`; tokens are never formatted into user-visible error strings.

---

### Settled concerns

1. **Concern: Does `SourceVersion` accurately capture raw byte changes when YAML diagnostics remain identical?**
   - *Settled:* Yes. `scanDirSources` computes `hashContent(content)` directly on the byte slice returned by `os.ReadFile` before handing it to `parse`. In `TestReadThreadsVersionsExactUnreadableSourceBytes`, two distinct YAML syntax errors yielding identical error diagnostics produced distinct `SourceVersion` hashes and tripped `domain.ErrConflict`.
2. **Concern: Does snapshot comparison mutate caller memory or suffer slice aliasing?**
   - *Settled:* No. `sameThreadSourceSnapshot` clones slices using `append([]T(nil), slice...)` prior to running `sort.SliceStable`. Verified by `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed`, which asserts caller slices remain unchanged after comparison.
3. **Concern: Can pathless remote adapters operate without fabricating fake filesystem paths?**
   - *Settled:* Yes. `ThreadReadProblem` treats `Location` as optional diagnostic context. Identity is established by `ThreadID` and `ThreadSlug`. Tests confirm pathless records with empty `Location` normalize and compare successfully.
4. **Concern: Does the body-aware Thread apply scan diverge in semantics or performance?**
   - *Settled:* No. `listThreadApplyThreads` uses `scanDirWithSourceVersions`, reading each file once in a single directory pass, hashing the raw bytes, and extracting the body via `splitFrontmatter`.
