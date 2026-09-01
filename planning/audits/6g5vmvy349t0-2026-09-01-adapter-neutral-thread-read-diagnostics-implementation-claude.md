---
schema: 1
id: 6g5vmvy349t0
bucket: closed
area: adapter-neutral-thread-read-diagnostics-implementation-claude
date: "2026-09-01"
updated_at: "2026-09-01"
---
# Audit: Adapter-neutral Thread read diagnostics implementation — Claude — 2026-09-01

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
> should update.

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

## Reviewer report

### Executive verdict

**Ready with tracked follow-ups.**

The port change is honest. `ThreadStore` now carries `ReadThreads() (ThreadRead, error)` and nothing
else filesystem-shaped; the FS adapter recovers filename identity at its own boundary and refuses to
guess when the filename is not id-led; core never reads `Location` for meaning, which I confirmed by
feeding it a pathless record whose location deliberately *looks* like `<id>-<slug>.md`. Guarded
creation, membership, Thread lifecycle, task lifecycle, and bulk apply all keep their native
`ListThreads` scan and their `domain.FileProblem` whole-source CAS, converting only the validation
vocabulary — so no guarded path gained a scan or lost exact source evidence. Determinism is real:
twelve shuffles of both records and problems produced byte-identical output, and the service now
clones before sorting rather than sorting the adapter's own slice in place, which the previous code
did.

Three findings, none blocking. One is a genuine duplicated diagnostic in the machine contract that a
TUI or web consumer will see. Two are coverage gaps on contract lines this patch itself wrote — most
notably, three of the four guarded conversion sites can be replaced with `nil` and the whole suite
stays green, even though that would make guarded writes fail *open* on unreadable Thread documents.
The behaviour is correct today; I verified all four paths refuse. Nothing would catch it regressing.

### Reviewed state and validation

Main worktree, based at `c4b3e63`, deliberately dirty: `63 files changed, 636 insertions(+),
117 deletions(-)`, 66 `git status --short` lines including untracked. Machine: Apple M5, darwin
25.6.0. Restored to exactly this state after every probe — `shasum -c` plus `cmp` over all seven
patched sources, and the temporary probe file deleted and confirmed absent.

| Command | Result |
| --- | --- |
| `go build ./...` · `go vet ./...` | clean |
| `go test ./...` | all packages pass |
| `go test -race -count=1 ./internal/core/... ./internal/store/... ./internal/cli/... ./internal/wire/... ./internal/workspacestore/... ./internal/tui/...` | pass, no race reports |
| `just lint` (golangci-lint) | 0 issues |
| `./bin/tskflwctl lint` · `audit lint` | both clean |
| `go mod tidy -diff` | clean |
| `git diff --check` | clean |
| `just docs` + `diff -rq` against a pre-run copy | no CLI-doc drift |
| `schema --json-schema` inspection | `ThreadReadProblemJSON` requires only `message`; `thread_id`/`thread_slug`/`location` optional; `FileProblem` no longer referenced by `ThreadsEnvelope` |
| Live hostile repos (4 scratch trees, ~25 invocations) | see findings and settled concerns |

Mutation probes, each applied then reverted and verified:

| Probe | Caught by |
| --- | --- |
| FS converter stops recovering filename identity | `TestReadThreadsRecoversFilesystemProblemIdentityAtAdapterBoundary`, `TestThreadListReportsIdentityAwareUnreadableRecords` |
| stop sorting read problems | `TestServiceThreadListPreservesAndOrdersAdapterNeutralProblems` |
| duplicate IDs stay healthy (`markDuplicateThreadIDs` removed) | `TestServiceThreadListFailsDuplicateIDsButAllowsDuplicateSlugs` |
| human render prefers location over identity | `TestThreadProblemsHumanPrefersIdentityAndKeepsOptionalLocation` |
| **`ReadThreads` performs a second native scan** | **nothing — M2** |
| **guarded conversion → `nil` (creation / thread mutation / task lifecycle)** | **nothing — M1** |
| guarded conversion → `nil` (bulk apply) | inconclusive: the patch left an unused variable and failed to build |

## Findings

### Medium

#### M1. Three guarded conversion sites can be made fail-open with the entire suite green  · **Status:** fixed

**File:** `internal/store/threadcreation.go:55`, `internal/store/threadmutation.go:54`, `internal/store/lifecyclemutation.go:58` | **Component:** testing/guarded-mutation
**Effort:** S · **Urgency:** soon

Each guarded flow now converts its native problems before validating:

```go
if err := core.ValidateThreadCreationSource(graph, threads, threadReadProblemsFromFiles(problems)); err != nil {
```

Replacing `threadReadProblemsFromFiles(problems)` with `nil` at any one of those three sites leaves
`go test ./internal/store/ ./internal/core/ ./internal/cli/` **fully green**. The fourth site
(`threadapply.go`, two occurrences) could not be probed the same way — the patch left an unused
variable and failed to compile — so its coverage is unproven rather than disproven.

The behaviour is correct as written. I verified all four guarded paths refuse while a Thread document
is unreadable, on a scratch repository containing one valid Thread and one malformed
`6g5w000000q1-unreadable.md`:

```
thread new Second …    exit=11  validation failed: current Thread record unreadable (6g5w000000q1) …
thread add <t> <task>  exit=11  validation failed: current Thread record unreadable (6g5w000000q1) …
thread start <t>       exit=11  validation failed: current Thread record unreadable (6g5w000000q1) …
task start <task>      exit=11  validation failed: current Thread record unreadable (6g5w000000q1) …
```

So this is purely a coverage gap — but on a fail-closed safety property, at the exact call sites this
patch rewrote. The failure mode of a regression is not a wrong message: it is a guarded write
committing while the repository contains Thread documents the tool cannot read, producing a receipt
that claims authority over a snapshot it never validated. That is the one thing the brief's angle 4
asks to confirm cannot happen silently, and today only manual inspection stands behind it.

**Recommendation:** Add one focused store test per guarded path — creation, Thread mutation, task
lifecycle, and apply — asserting `ErrValidation` and no write when an unreadable Thread file is
present. The scratch reproduction above is directly translatable. Fix the apply probe's unused
variable while you are there so that site is provably covered too.

**Resolution:** Added focused store regressions for Thread creation,
membership/lifecycle mutation, task lifecycle mutation, and bulk apply. Every
path now proves an unreadable Thread source fails closed before any durable
write; apply also covers concurrent unreadable evidence at final snapshot
verification.

#### M2. The "exactly one native scan per portable read" contract line is untested  · **Status:** fixed

**File:** `internal/store/threadstore.go:18-30` | **Component:** testing/store
**Effort:** XS · **Urgency:** soon

The contract states that the filesystem adapter "performs exactly one native `ListThreads` scan for
each portable read". Inserting a redundant `s.ListThreads()` call inside `ReadThreads` before the real
one leaves the whole suite green.

The code is already factored to make this trivially testable — `readThreads(scan func() (...))` takes
an injectable scan, and `internal/store/paths_test.go:102` already uses that seam for a *different*
assertion. Nothing counts invocations.

Two things regress silently without this pin: directory-scan cost doubles on every `thread list`, and
more importantly the read stops being one snapshot. Two scans can disagree, so a Thread deleted
between them would appear in `Threads` while its problem vanished, or vice versa — a torn portable
read presented as a canonical repository view, which is precisely what `ThreadRead`'s doc comment
says the type exists to prevent.

**Recommendation:** Extend the existing `readThreads` seam with a counting scan function and assert
exactly one call. Roughly three lines beside the test that already uses that helper.

**Resolution:** Pinned the concrete FS.ReadThreads wrapper to exactly one native
ListThreads invocation with a per-instance counting scan seam, so an accidental
second scan fails the focused adapter test.

### Low

#### L1. `completed-thread-unhealthy-evidence` is emitted twice when a completed Thread has a duplicate ID and any other broken evidence  · **Status:** fixed

**File:** `internal/core/service_thread.go` `markDuplicateThreadIDs` vs `internal/core/thread_projection.go` completed block | **Component:** core/thread-projection
**Effort:** XS · **Urgency:** eventually

`ProjectThread` appends `ThreadProblemCompletedUnhealthyEvidence` whenever a completed Thread's
projection is already unhealthy. `markDuplicateThreadIDs` then appends the same code again,
unconditionally, for every duplicate-ID member. When both conditions hold, the view carries the
identical problem twice.

**Reproduction** — two completed Threads sharing frontmatter id `6g5w000000c1`, where the second also
references a missing member:

```
comp-a  inconsistent=True  codes=[completed-thread-undrained, duplicate-thread-id,
                                  completed-thread-unhealthy-evidence]
comp-b  inconsistent=True  codes=[invalid-thread-document, thread-id-drift, missing-thread-member,
                                  completed-thread-unhealthy-evidence, completed-thread-undrained,
                                  duplicate-thread-id, completed-thread-unhealthy-evidence]
        >>> REPEATED CODES: {'completed-thread-unhealthy-evidence': 2}
```

`comp-a` — duplicate ID with no other defect — correctly gets exactly one, so the duplication only
appears in the compound case. The verdict fields (`ProjectionHealth`, `Inconsistent`, frontier
suppression) are all correct; what is wrong is the diagnostic list itself, which is part of the
machine contract the TUI and future web adapters are about to bind to. A consumer counting problems,
grouping by code, or rendering a list shows the same sentence twice.

**Recommendation:** Have `markDuplicateThreadIDs` add the completed-evidence problem only when the
view does not already carry that code, or set the duplicate problem before the completed-consistency
pass runs so a single site owns that conclusion.

**Resolution:** Deduplicated completed-thread-unhealthy-evidence when duplicate
identity is layered onto an already-broken completed projection, with a compound
missing-member plus duplicate-ID regression asserting one occurrence.

## Acceptance-criteria traceability

| # | Criterion (abbreviated) | Verdict | Evidence |
| --- | --- | --- | --- |
| 1 | A pathless fake attributes an unreadable Thread by ID and slug; list and machine output retain it without fabricated filesystem data | Met | Probe with a pathless `ThreadRead` carrying explicit identity and a *misleading* `<id>-<slug>.md` location: core preserved the explicit values and did not backfill from `Location`. JSON omits `location` when empty (`omitempty`). |
| 2 | FS adapter preserves readable Threads, exact diagnostics, and optional repair paths with no duplicate scan | **Qualified** | Identity recovery correct on four hostile filenames; original messages and absolute paths intact; measured exactly one `ReadThreads` per `ListThreadViews`. The single-native-scan half is unpinned — M2. |
| 3 | Core projections and the TUI-facing boundary depend only on the neutral contract; no path parsing or filesystem type crosses it | Met | `ThreadStore` no longer mentions `domain.FileProblem`; `ThreadsEnvelope.Unreadable` is `[]ThreadReadProblemJSON` and the generated schema no longer references `FileProblem`; `internal/cli/render/thread.go` dropped its `domain` import. |
| 4 | Missing, invalid, duplicate, drifted and unreadable identities order deterministically and do not collide with readable Threads or task IDs | Met | Live matrix produced `thread-task-id-collision`, `duplicate-thread-id`, `thread-id-drift` (and drift+duplicate together) with `proj=broken` and `frontier=0` for every affected row. Twelve shuffles of records and problems gave byte-identical output. |
| 5 | Lint, guarded creation/membership/lifecycle/apply validation, CLI human output, and JSON fixtures retain fail-closed behavior | **Qualified** | All four guarded paths refuse (exit 11) with an unreadable Thread present; `thread list` exits 11 with clean JSON stdout; local lint keeps exact paths and messages. Retained but unpinned — M1. |
| 6 | Schema comments, generated schema, architecture docs, and compatibility notes name the wire change | Met | Schema 1.59; `ThreadReadProblemJSON` requires only `message`; problem-code vocabulary extended in both the tag and the golden; README preview notice names 1.59 and the path-optional change; ADR and `docs/ARCHITECTURE.md` updated; no CLI-doc drift. |

## Settled concerns

Chased and disproved by code inspection plus a hostile reproduction.

1. **Core silently re-deriving identity from `Location`.** The obvious way to make this port "work"
   is to parse the location string in core. It does not: a pathless problem whose `Location` is
   `6g5w000000f9-misleading.md` came back with `ThreadID=""` and `ThreadSlug=""`. Only
   `internal/store/threadstore.go` interprets filenames, and it uses `splitFlatName`, which validates
   the Crockford-base32 id rather than pattern-matching a prefix.
2. **Guessed identity for filenames that merely resemble the shape.** Four hostile files:
   `6g5w000000a1-malformed.md` → `id=6g5w000000a1 slug=malformed`;
   `6g5w000000a2-multi-part-slug-here.md` → full multi-segment slug preserved;
   `NOTANID00000-badid.md` → **unattributed**; `plainname.md` → **unattributed**. Invalid ids stay
   unattributed rather than being guessed, and the original error text (including the Crockford
   explanation) survives verbatim.
3. **A second scan or a torn snapshot at the service layer.** A counting fake showed exactly one
   `ReadThreads` call per `ListThreadViews`. Records and problems travel in one `ThreadRead` value, so
   the service cannot split one logical snapshot. (The FS-level scan count is the separate M2 gap.)
4. **Adapter aliasing and leaked persistence versions.** The previous code sorted the slice returned
   by `ListThreads()` in place. The new code does `cloneThreads(read.Threads)` and
   `append([]ThreadReadProblem(nil), read.Problems...)` before sorting; across twelve shuffles the
   adapter's own slices were never reordered. `cloneThreads` also clears `SourceVersion`, so the
   persistence token never reaches the projection or the wire.
5. **Unstable sorting leaking scan order.** `threadLess` compares a fourteen-field key and
   `threadReadProblemLess` a four-field key, so `sort.Slice` instability is only reachable for records
   equal in every compared field — and those are indistinguishable in output, since `SourceVersion`
   (the one uncompared field) is stripped before rendering. `markDuplicateThreadIDs` iterates a map,
   but each view belongs to exactly one id bucket, so map order cannot affect any view's problem
   order.
6. **Weakened drift validation.** `ValidateThreadCreationSource` changed from
   `FilenameID == "" || FilenameID != ID` to `FilenameID != "" && FilenameID != ID`, which reads like a
   loosened guard. It is required for pathless adapters and safe for the filesystem: `parseThread`
   rejects any non-id-led name outright, so a readable local Thread can never carry an empty
   `FilenameID`. Live drift is still caught in both the guarded path and the projection
   (`thread-id-drift`).
7. **Guarded flows accidentally calling the portable wrapper.** They do not. Creation, thread
   mutation, task lifecycle and apply all still call native `ListThreads`, keep `[]domain.FileProblem`
   for `sameThreadSourceSnapshot`, and convert only for validation vocabulary. The whole-source CAS
   and repository-planner exclusion are untouched.
8. **Machine-contract drift beyond the intended change.** Every golden diff outside the Threads and
   schema fixtures is the `1.58` → `1.59` bump alone. `thread list --json` retains non-null arrays
   (dogfood shows `unreadable=[]`), and the error envelope goes to stderr while valid JSON stays on
   stdout with exit 11.
9. **The lint adaptation hiding a present defect.** For the filesystem — the only adapter that lints
   today — `ThreadReadProblem{Location, Message}` → `domain.FileProblem{Path, Message}` is lossless,
   and live `lint` output keeps every absolute path and message. Identity is lost only for a pathless
   adapter, which cannot reach `Service.Lint` today. Follow-up `6g5vm4efjcdv` is filed,
   `ready-to-start`, and scoped to the shared multi-entity bucket rather than to Threads — correctly
   separated, not an excuse.
10. **Human output regressing to path-first naming.** `ThreadProblemsHuman` leads with
    `slug (id)`, falls back to id, then to `unidentified Thread record`, and prints `location:` as
    secondary context only when present. Verified without color; a probe that prefers location fails a
    focused render test.
11. **Proportionality.** `ThreadRead` bundles records and problems in one value specifically so a
    future source-revision token can qualify the whole read — a smaller `[]ThreadReadProblem` return
    would foreclose that. The exported surface is four fields and two types; I found no smaller
    contract preserving the same adapter freedom. Aesthetic preference only.
12. **Dogfood agreement.** `thread list` on this repository is clean (`10/16 done`, `2 eligible`,
    `healthy/healthy`, `unreadable=[]`, schema 1.59) and planning plus audit lint both pass.
