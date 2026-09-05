---
schema: 1
id: 6g725zn6r92x
bucket: closed
area: version-unreadable-task-graph-sources-for-repair-safe-cas-implementation-claude
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: unreadable task graph source revisions — Claude — 2026-09-05

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
> should update.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Codes must match
> `[A-Z]+[0-9]+`; do not put status on a separate line or pre-resolve a finding.
>
> Required second pass: after completing the checklist, review again as a devil's advocate for
> systemic failure modes. Challenge the source/projection boundary, compatibility adapters,
> normalization and equality assumptions, test helpers, and future-repair claims. Prefer one
> demonstrated systemic issue over several speculative findings.
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
AUDIT_REL="planning/audits/6g725zn6r92x-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation-claude.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review-claude.XXXXXX")"

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

Assume the change can be wrong despite green tests. Re-derive the data path from exact source bytes
through the filesystem scan, neutral `TaskGraphRead`, private immutable `TaskGraph` evidence,
`SameSourceSnapshot`, and the pre-write conflict helper. Do not edit implementation or other
planning files, create follow-up tasks, change finding statuses, close this audit, push, or install
anything globally.

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

### Executive verdict — `ready with tracked follow-ups`

The slice does what it claims. The filesystem hashes the exact bytes it hands the parser, the
token is opaque to core and invisible on every public surface, `SameSourceSnapshot` normalizes
adapter ordering before comparing complete load-problem evidence, and the pre-write verifier is
genuinely on the production path returning a typed `ErrConflict`. All thirteen hostile filesystem
fixtures I built behave correctly, all six required mutations are killed by shipped tests, and
eight CLI surfaces are **byte-identical** to the pre-change binary on a deliberately degraded
repository. No behavioural defect survived verification.

Four follow-ups are recorded. One is a demonstrated coverage hole in the shipped tests (M1: the
fail-closed rule is load-bearing but nothing asserts it). One is the systemic second-pass result
(M2: the Thread half of the very same CAS expression still has the exact blind spot this task
closes for tasks — currently unreachable, but it is the premise the repair path will remove).
Two are consistency/accuracy items (L1, L2). None blocks landing.

### Branch, base, checkpoint, runtime, isolation attestation

| | |
|---|---|
| Branch | `feat/unreadable-graph-source-revisions` |
| Base | `7835374` (`docs(planning): harden broken graph repair design`), on `main` at `c921099` |
| Reviewed artifact | the complete uncommitted working tree, captured as sandbox checkpoint `159cb7b0cb6ae4e65e8e39f3de0dad4bf276ffd9` |
| Runtime | go1.26.6 darwin/arm64, golangci-lint 2.12.2, macOS 26.6.1 |

**Isolation.** The handoff checkout was read exactly twice: to read this brief, and as the `git clone
--no-hardlinks` + `rsync -a --delete --exclude=.git` source for the sandbox. Sandbox at
`/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review-claude.l3MSYV`;
`git rev-parse --absolute-git-dir` resolves to `<sandbox>/.git`; `.git/objects/info/alternates` is
absent, so no object is shared with the source. It is not a worktree and not a symlink. Every
build, test, benchmark, mutation probe, fixture, CLI comparison and report edit ran there. No
`git worktree`, and no staging, commit, restore, clean, stash, reset, branch switch, push, global
install, or write-capable project command was run in `$SOURCE_ROOT`. Both binaries used for the
compatibility comparison were built from `git archive` extracts into the scratchpad, never by
mutating a checkout. Before transfer, sandbox `git status --short` listed only this audit.

### Source-to-CAS data flow

```
tasks/<id>-<slug>.md
  └─ os.ReadDir → markdownDoc (regular .md only; symlinks rejected)
     └─ README.md carveout skipped silently
        └─ os.ReadFile(path) ─────────────► content []byte   (read error = FATAL, unchanged)
           ├─ hashContent(content)  ← the SAME slice later handed to parse; SHA-256, store/cas.go
           └─ parse(path, content) = parseTask
              ├─ ok   → domain.Task{ SourceVersion: hashContent(content) }   [readable]
              └─ err  → sourceFileProblem{ FileProblem{Path,Message}, sourceVersion }  [unreadable]
                        (parse failure stays RESILIENT, unchanged)
```

`FS.ReadTaskGraph` then converts each `sourceFileProblem` to a `core.TaskGraphLoadProblem`, deriving
identity via `splitFlatName(base)`; `core.NewTaskGraphRead` → `newTaskGraph` defensively clones the
problems, sorts them on `(TaskID, TaskSlug, Path, Message, SourceVersion)`, and retains them
privately as `g.loadProblems`. `SameSourceSnapshot` compares health, ids, per-task path+version,
then `loadProblems`, `problems`, `legacy`. `verifyTaskGraphSourceSnapshot` wraps the result in
`domain.ErrConflict`.

**Producer / consumer inventory.**

| Symbol | Role | Token? |
|---|---|---|
| `scanDirSources(dir, parse, versionProblems)` | new shared scan body | conditional |
| `scanDir` | `ListTasks`/`ListEpics`/`ListAudits`/… projection | **no** (projects down to `FileProblem`) |
| `scanDirWithSourceVersions` | sole token-bearing scan | **yes** |
| `FS.ListTasks` | ordinary list surface | **no** — shape unchanged |
| `FS.ReadTaskGraph` | the only production `TaskGraphSource` in the wired app | **yes** |
| `core.TaskGraphReadFromFiles` | legacy `[]domain.FileProblem` adaptation | **no** |
| `taskStoreGraphSource` | wired fallback at `service.go:192` for stores lacking `ReadTaskGraph` | **no** |
| `core.NewTaskGraph` | body-aware/lint path (`service.go:433`, `dependency_graph_mutation.go:131`) | **no** |
| `core.NewTaskGraphRead` | strict-snapshot constructor (5 production call sites) | passthrough |
| `core.LoadTaskGraph` | canonical loader used by every guarded store | passthrough |
| `taskGraphFileProblems` | back-projection for board/status/list (`board.go:48`, `service.go:295`, `service_task.go:51`) | **strips** the token |
| `SameSourceSnapshot` consumers | `graphmutation.go:87` (via helper), `threadmutation.go:92`, `threadapply.go:123`, `threadcreation.go:88`, `lifecyclemutation.go:110` | all load both sides via `core.LoadTaskGraph(*FS)` → tokens present |

`*FS` satisfies `TaskGraphSource`, so the wired application always takes the token-bearing path;
`taskStoreGraphSource` and `NewTaskGraph` are the token-less producers and neither reaches a CAS.

### Findings

#### M1. The fail-closed rule for missing unreadable-source revisions is load-bearing but unasserted · **Status:** fixed

`sameTaskGraphLoadProblem` (`internal/core/dependency_graph.go:908`) requires
`left.SourceVersion != ""`. That clause is the *only* thing stopping two token-less unreadable
snapshots from comparing equal — the brief's explicit requirement that missing revision evidence be
"CAS-ineligible: it must fail closed rather than claim two unreadable snapshots are equal".

The shipped test's `missing` case sets **one** side's token to `""` and compares it against a side
holding `"opaque-revision-a"`. Two different strings are unequal with or without the guard, so that
assertion is already covered by the preceding `changed` case and does not exercise the rule.

Demonstrated in the sandbox: deleting `left.SourceVersion != "" &&` leaves the entire suite green —
`go test ./internal/core/... ./internal/store/...` reports `ok` for both packages. A probe
constructing two graphs whose single unreadable record carries no revision at all flips
`SameSourceSnapshot` from `false` (as shipped) to `true` (with the clause removed), and only that
probe fails. Token-less producers are live in the wiring (`taskStoreGraphSource` at
`service.go:192`, `NewTaskGraph` at `service.go:433`), so the guard protects a reachable
composition even though no CAS consumes one today.

**Failure scenario:** a future adapter or refactor routes a token-less `TaskGraphRead` into a
guarded mutation. The guard is silently dropped or inverted during an unrelated edit; every test
still passes; two structurally identical but byte-different broken snapshots compare equal and a
repair writes over a concurrent raw edit. Suggested fix: assert the both-empty case directly.

**Resolution:** Added a direct both-empty regression so two otherwise identical
unreadable records without source revisions remain CAS-ineligible.

#### M2. The Thread half of the same CAS expression retains the exact blind spot this task closes for tasks · **Status:** tracked by 6g72ncs4xjdm

This is the systemic second-pass result. Four production pre-write checks evaluate the task-side and
Thread-side comparisons in **one boolean expression**:

```go
if !graph.SameSourceSnapshot(currentGraph) || !sameThreadSourceSnapshot(threads, problems, currentThreads, currentProblems) {
```
(`threadmutation.go:92`, `threadapply.go:123`, `threadcreation.go:88`, `lifecyclemutation.go:110`)

After this change the left operand detects a raw edit to an unreadable *task* file that reproduces
the same parse error. The right operand cannot do the same for an unreadable *Thread* file:
`sameThreadSourceSnapshot` (`internal/store/threadcreation.go:179-181`) compares unreadable Threads
with `slices.Equal(leftProblems, rightProblems)` over `[]domain.FileProblem`, i.e. `{Path, Message}`
only — while comparing *readable* Threads with the same fail-closed non-empty `SourceVersion` rule
the task side uses (`:189`). `core.ThreadReadProblem` (`internal/core/store.go:112`) still carries
the placeholder comment "so a future source revision token can qualify the complete read" — the
same wording this task just removed from `TaskGraphRead`.

**Reachability is settled for today, and the evidence is:** `ValidateThreadCreationSource`
(`internal/core/thread_creation.go:66`) returns `ErrValidation` whenever `len(unreadable) > 0`, and
every one of the four sites calls it (directly or via `ValidateThreadMutationSource`) before
planning. A snapshot containing an unreadable Thread therefore aborts before the verify, so the
weak comparison cannot currently be reached with unreadable records on both sides. I confirmed the
task-side equivalent the same way: `ValidateTaskGraphMutationSource` / `ValidateTaskLifecycleSource`
refuse a broken graph, which is why this task's new comparison is likewise unreachable in today's
write paths and is correctly scoped as preparation.

**Failure scenario:** the guarded repair path lands and, by design, removes the broken-source
refusal so it can operate on damaged state. A repair transaction that also touches Threads (the
`lifecyclemutation.go` pairing already couples the two) then runs a half-versioned CAS: a
concurrent raw edit to an unreadable Thread document that reproduces its parse error compares equal
and the repair proceeds on stale evidence — the precise hazard this task exists to eliminate.
Fixing it is out of this slice's scope ("designing a universal revision scheme for every entity
type"), so this is recorded as a source-supported forward risk for a tracked follow-up, not a
defect in the change under review.

**Resolution:** Confirmed the Thread-side blind spot is unreachable in ordinary
guarded mutations but must be closed before repair authorizes partial Thread
evidence; a focused prerequisite is now sequenced in the production Thread.

#### L1. Only one of five production pre-write CAS sites uses the new named verifier · **Status:** fixed

`verifyTaskGraphSourceSnapshot` (`internal/store/graphmutation.go:117`) documents itself as "the
pre-write whole-repository CAS **shared by** healthy graph mutations and the forthcoming
broken-graph repair path", but only `graphmutation.go:87` calls it. The four sites listed in M2
still inline `!graph.SameSourceSnapshot(currentGraph)` with their own `ErrConflict` wrapping.

Behaviour is equivalent today — I verified all four load *both* the expected and current graph
through `core.LoadTaskGraph(s)` on `*FS`, so all four already inherit the new unreadable-source
evidence, and all four wrap `domain.ErrConflict`. **Failure scenario:** a later change tightens the
helper (structured conflict detail, a repair-authorization assertion, telemetry) and four
production write paths silently keep the old, weaker check because they never routed through it.

**Resolution:** Moved the graph snapshot verifier into the shared CAS module and
routed dependency, Thread creation/mutation/apply, and task lifecycle pre-write
checks through it while preserving operation-specific errors.

#### L2. The "conversion is confined here" comment is now inaccurate, and the filename→identity rule has two copies · **Status:** fixed

`core.TaskGraphReadFromFiles` still carries "Filename parsing for unreadable identity is
deliberately confined here" (`internal/core/service_task.go:145-148`), and ADR-0006 says the
filesystem "performs the filename conversion at its adapter boundary". `FS.ReadTaskGraph` now
derives identity itself via `splitFlatName` (`internal/store/fsstore.go:126`), so there are two
implementations: `store.splitFlatName` (`internal/store/flatname.go:28`) and
`core.taskIdentityFromPath` (`internal/core/dependency_graph.go:537`).

I compared them line by line: identical length gate (`id.Length+2`), identical separator check
(`stem[id.Length] != '-'`), identical `id.Valid` gate, identical slug slice — they cannot disagree
today, which is why nothing fails. **Failure scenario:** the id grammar changes (a prefix, a second
separator, a length change) and only one copy is updated; the same unreadable file then gets one
identity on the ordinary list path and another on the graph path, so `sameGraphProblem` sees a
`TaskID` change that never happened and every guarded mutation on a degraded repo returns a
permanent, unexplainable `ErrConflict`. Either delete the stale "confined here" claim and note the
two-copy invariant, or have the store call one shared helper.

**Resolution:** Added one core file-diagnostic conversion helper used by both
the legacy list adapter and the filesystem graph reader, eliminating duplicate
filename identity parsing.

### Acceptance-criteria traceability

| AC | Verdict | Evidence |
|---|---|---|
| Two scans of byte-identical unreadable content compare equal | **met** | fixture 1 below; plus `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS`'s byte-restoration leg |
| Changed bytes differ despite identical path, identity, code and message | **met** | fixtures 2–3; the shipped store test asserts message/path/TaskID equality *before* asserting the revisions differ |
| Readable↔unreadable transitions always invalidate | **met** | fixtures 7–8 (both directions, real filesystem) and `TestTaskGraphSameSourceSnapshotRejectsReadableUnreadableTransition` (both call orders) |
| Pathless adapters supply their own revision; token never becomes domain/wire data | **met** | `TestTaskGraph…ComparesOpaqueUnreadableRevisions` uses pathless records; `json:"-" yaml:"-"`; `Task()` clears the readable token (`dependency_graph.go:806`); eight CLI surfaces byte-identical; no 64-hex token in any output |
| Existing reads/mutations retain behaviour; focused race tests prove `ErrConflict` before any write | **met, with a scope note** | full base-vs-head CLI comparison below; `-race -count=25` focused runs. The `ErrConflict` proof is necessarily at the helper level (`verifyTaskGraphSourceSnapshot` called directly) because no repair path exists yet to drive it end-to-end — today every write path refuses a broken graph before reaching the verify (see M2). That is the correct reading of "would receive", not a gap. |

### Hostile-evidence ledger

Thirteen fixtures driven through the real `FS` adapter and `core.LoadTaskGraph`, comparing
`before.SameSourceSnapshot(after)`. **All thirteen behaved correctly.**

| # | Fixture | same= | want | |
|---|---|---|---|---|
| 1 | unchanged unreadable bytes | true | true | ✅ |
| 2 | changed bytes, same parse error | false | false | ✅ |
| 3 | changed body *outside* malformed frontmatter | false | false | ✅ |
| 4 | unreadable file renamed, byte-identical | false | false | ✅ |
| 5 | unreadable file added | false | false | ✅ |
| 6 | unreadable file removed | false | false | ✅ |
| 7 | unreadable → readable | false | false | ✅ |
| 8 | readable → unreadable | false | false | ✅ |
| 9 | non-id-led stray (`notes.md`), unchanged | true | true | ✅ |
| 10 | non-id-led stray, bytes changed | false | false | ✅ |
| 11 | duplicate unreadable identity (one id, two slugs), unchanged | true | true | ✅ |
| 12 | duplicate unreadable identity, one file's bytes changed | false | false | ✅ |
| 13 | `README.md` carveout added beside an unreadable file | true | true | ✅ |

Fixtures 9–10 confirm a non-id-led stray participates in the snapshot with `TaskID=""` and is still
byte-versioned. Fixtures 11–12 confirm the sort key is canonical over all five compared fields, so
duplicate identities pair correctly. Fixture 13 confirms the README carveout is untouched by the
refactor.

**Compatibility proof.** A binary built from base `7835374` and one from the checkpoint were run
against an identical degraded planning repository (the real 58-task tree plus an injected
no-frontmatter file, an injected bad-`tier` file, a non-id-led `notes.md`, and a `README.md`):

```
IDENTICAL (exit 1)  :: task list --json
IDENTICAL (exit 1)  :: task list -o table -c slug,status,description
IDENTICAL (exit 1)  :: lint --json
IDENTICAL (exit 1)  :: status --json
IDENTICAL (exit 1)  :: epic list --json
IDENTICAL (exit 1)  :: schema --json-schema
IDENTICAL (exit 1)  :: board --json
IDENTICAL (exit 1)  :: task list --json -c slug,status,description
```

Byte-identical on stdout, stderr and exit code; the `unreadable` array reports all three bad files
with the unchanged `{path,message}` shape; `grep -oE '[0-9a-f]{64}'` finds no token in any output.
Real lifecycle writes were also compared: `task ready` + `task start` on a healthy copy produced
byte-identical task files and JSON differing only in the temp-directory `planning_root`; on the
degraded copy both binaries refused identically.

### Restored-mutation ledger

Every mutation was applied in the sandbox, tested, then reverted with `git checkout -- internal/`.

| # | Mutation | Killed by |
|---|---|---|
| M-a | omit the unreadable revision comparison from `SameSourceSnapshot` | `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions` **and** `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` |
| M-b | derive the revision from path+message instead of bytes | `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` |
| M-c | return an empty FS token | `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` |
| M-d | expose the token through JSON (drop `json:"-"`) | `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions` |
| M-e | ignore adapter ordering (drop the normalizing sort) | `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions` |
| M-f | drop unreadable problems from the retained snapshot | both new tests |
| **M-g** | **drop the fail-closed `SourceVersion != ""` clause** | **SURVIVED → finding M1** |
| M-h | ignore `Path` in `sameTaskGraphLoadProblem` | survived — **redundant, not a gap** (see settled #3) |
| M-i | ignore `Message` in `sameTaskGraphLoadProblem` | survived — redundant (settled #3) |
| M-j | ignore `TaskSlug` in `sameTaskGraphLoadProblem` | survived — redundant (settled #3) |
| M-k | hash *after* parse instead of before | survived — no behavioural difference with today's read-only parsers (settled #2) |

All six required mutations are killed. M-g is reported as M1; M-h/M-i/M-j are demonstrably redundant
rather than uncovered; M-k is a deliberate defensive ordering with no observable effect today.

### Concurrency reasoning

**Whole-snapshot boundary.** `MutateTaskGraph` takes the repository lock, loads the authoritative
graph, plans, materializes, then re-loads and calls `verifyTaskGraphSourceSnapshot` *before the
first write* (`graphmutation.go:80-89`). The helper is genuinely on the production path, returns
`fmt.Errorf(... %w, domain.ErrConflict)` — preserving the typed sentinel and exit code 14 — and I
confirmed by mutation that removing the unreadable-revision term makes the store test fail.

**Per-target boundary.** Each planned write re-checks `verifyUnchanged(s.resolvePath, taskID, path,
ifVersion, ...)` immediately before its atomic replacement, bounding the raw-editor window to one
replacement after a durable prefix.

**What this change closes.** Previously, an unreadable source contributed only
`GraphProblem{Code, TaskID, Path, Message}` to the comparison, so a raw edit that changed the bytes
while reproducing the same parse error at the same path was invisible to the whole-snapshot CAS.
It is now caught (fixtures 2–3).

**What remains open, stated precisely.**
1. The verify→write window is unchanged: files that are *read but not written* are covered only by
   the whole-snapshot check, so a raw edit landing after the verify is not detected for them. The
   per-target CAS covers only planned writes. This is pre-existing and documented in the code.
2. There is no per-target CAS for an *unreadable* file, because unreadable files are never written
   today. The repair path will need one — `verifyUnchanged` resolves by canonical slug, which is
   exactly what a broken file may not have.
3. The Thread half of the paired comparison is still unversioned for unreadable records (M2).
4. The lock is advisory: it excludes cooperating writers, not a raw editor. The hashes are the only
   guard against the latter, which is precisely why M1's fail-closed clause matters.

### Settled concerns

1. **Does `scanDir`'s `make([]domain.FileProblem, len(...))` change `nil` to `[]` on the wire?**
   No. It does now return a non-nil empty slice where it previously returned `nil`, but every list
   envelope tags `Unreadable` `omitempty` (identical for nil and empty), and the two envelopes that
   do not (`ToLintEnvelope`, `ToFixEnvelope`) already normalize `nil` → `[]` explicitly. Confirmed
   by the byte-identical `lint --json` comparison.
2. **Does hashing every scanned file (not just failures) violate the O(bytes) / no-second-scan
   constraint?** No second scan, and still O(total bytes) — but readable files *are* hashed twice
   (once in `scanDirSources`, again inside `parseTask`). I measured it rather than guessed: at
   2000 files × 8 KB, `ReadTaskGraph` is ~74.7 ms as shipped vs ~71.0 ms with lazy hashing and
   ~69.9 ms for `ListTasks`, i.e. **~5%**, dominated by I/O and YAML parsing. The eager placement is
   a documented defensive choice (hash the exact slice before handing it to a parser). Not worth a
   finding.
3. **Why did ignoring `Path`/`Message`/`TaskSlug` survive mutation?** Because they are redundantly
   covered. Every load problem also produces a `GraphProblem{Code, TaskID, Path, Message}`, and
   `sameGraphProblem` (`dependency_graph.go:911`) already compares `Path` and `Message`; `TaskSlug`
   is derived from the same filename as `Path`. `SourceVersion` is the only genuinely new
   discriminator, and it is covered.
4. **Can the sort key including `SourceVersion` cause false equality or duplicate-pairing errors?**
   No. The key is a total function of exactly the five fields `sameTaskGraphLoadProblem` compares,
   so sorted order is canonical for the multiset: two multisets that compare pairwise-equal after
   sorting are the same multiset. Fixtures 11–12 exercise duplicate identities.
5. **Is the caller's slice aliased?** No. `cloneTaskGraphLoadProblems` copies before the in-place
   sort, and `TaskGraphLoadProblem` is all-string, so the shallow copy is a deep copy.
6. **Did the `ReadTaskGraph` rewrite lose `ListTasks` behaviour?** No. It re-issues
   `rejectRepositoryPlannerCall` itself (planner re-entry still rejected — covered by
   `TestMutateTaskGraphRejectsPlannerStoreCallsAndNestedMutationWithoutHanging`), uses the same
   scanner and the same `parseTask` closure, keeps read errors fatal and parse errors resilient,
   and preserves the README and symlink/non-regular-file carveouts (`markdownDoc` untouched;
   fixture 13). One scan, as before.
7. **Does the new ordering change user-visible output?** `g.problems` for unreadable records is now
   emitted in sorted rather than adapter order. For the filesystem the sort key leads with `TaskID`
   and `os.ReadDir` already returns id-led filenames sorted, so the orders coincide; the degraded-repo
   comparison came back byte-identical, including the `notes.md` stray that sorts under `TaskID=""`.
8. **Do the token-less producers cause false conflicts?** No. `taskStoreGraphSource` and
   `NewTaskGraph` are the only token-less producers and neither feeds a CAS; every
   `SameSourceSnapshot` consumer loads both sides through `core.LoadTaskGraph(*FS)`. Verified
   end-to-end: `task ready`/`task start` on a healthy repo behave identically to base.

### Separation of concern classes

- **Demonstrated defects:** none. No behavioural defect survived verification.
- **Demonstrated test-coverage defect:** M1 — mutation survives the full suite; probe attached.
- **Source-supported risks (verified in source, not currently reachable):** M2, L1, L2.
- **Unverified concerns:** none. Every hypothesis I raised was either demonstrated or settled with
  evidence above.

### Validation performed (all inside the sandbox)

| Command | Result |
|---|---|
| `go build ./...` | clean |
| `go vet ./...` | clean |
| `go test -race ./...` (uncached, `go clean -testcache` first) | **24 packages ok**, 0 failures |
| `go test -race -count=25 ./internal/core/ -run 'TestTaskGraphSameSourceSnapshot\|TestTaskGraphRead'` | ok 1.534s |
| `go test -race -count=25 ./internal/store/ -run 'TestUnreadableTaskSourceRevision\|TestMutateTaskGraph'` | ok 34.145s |
| `go test -race -count=10 ./internal/store/ -run 'Concurrent\|Race\|Conflict\|Lock\|Guard'` | ok 3.651s |
| `just lint` (golangci-lint 2.12.2) | `0 issues.` |
| `just tidy-check` (`go mod tidy -diff`) | clean |
| `just docs-check` | clean |
| `gofmt -l ./internal ./cmd` | no output |
| `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| `./bin/tskflwctl audit lint 6g725zn6r92x` | `✔ all audit findings pass lint` |
| `git diff --check 7835374 HEAD` | clean |
| base-vs-head CLI comparison, 8 surfaces, degraded repo | byte-identical |
| hostile fixture matrix, 13 cases | 13/13 correct |
| mutation probes, 11 applied and reverted | 6/6 required killed; 1 survivor reported as M1 |

All probes, fixtures, benchmarks and generated artifacts were reverted; sandbox
`git status --short` was empty before this audit was edited, and lists only this file at transfer.
