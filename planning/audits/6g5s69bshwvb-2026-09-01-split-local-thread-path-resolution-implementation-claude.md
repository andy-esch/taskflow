---
schema: 1
id: 6g5s69bshwvb
bucket: closed
area: split-local-thread-path-resolution-implementation-claude
date: "2026-09-01"
updated_at: "2026-09-01"
---

# Audit: Split local Thread path resolution implementation — Claude — 2026-09-01

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
> should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for
[`split-local-thread-path-resolution-from-portable-thread-reads`](../tasks/6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
in the main worktree, based at commit `89cfb85`. Review it against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially the 2026-08-31 portable-read
boundary and 2026-09-01 TUI-foundation amendment, plus
[`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

The patch is small, but assume its construction semantics may be subtly wrong despite green tests.
Concentrate on whether it establishes a durable adapter-neutral boundary for CLI, TUI, future web,
database, API, and cache implementations. Do not reward complexity or test count by itself, and do
not manufacture findings: settle a concern when code and a hostile reproduction disprove it.

## Review target

The implementation is uncommitted in a deliberately dirty worktree. Inspect `git status --short`,
the relevant portions of `git diff HEAD`, and the untracked test/task/audit files. Primary files are:

- `internal/core/store.go`, `service.go`, `service_thread.go`, and `workspace.go`;
- `internal/core/service_thread_test.go` and `workspace_test.go`;
- `internal/store/threadstore.go`, `paths.go`, and the untracked `paths_test.go`;
- `internal/workspacestore/fs.go` and `fs_test.go`;
- `internal/cli/thread.go`, `thread_test.go`, and existing integration goldens;
- `README.md`, `docs/ARCHITECTURE.md`, ADR-0006, and the implementation task; and
- any production construction site or `ThreadStore` implementation found through repository-wide
  search.

Ignore unrelated simultaneous planning edits, including the broader TUI task split, the completed
v0.18.0 release task change, and other new foundation task files. This audit file is review
scaffolding, not implementation evidence.

The intended contract is:

- `ThreadStore` is a portable semantic document-read port containing only `ListThreads` and
  `GetThread`. List/show/compose/projection/plan/graph behavior must not require filesystem methods.
- `ThreadPathSource` is a separate optional consumer-owned capability used only by explicit Thread
  path lookup. A service may have reads without paths or paths without reads.
- `NewService(completeStore)` remains ergonomic by discovering every capability the aggregate value
  actually implements. Explicitly replacing Thread reads must detach an aggregate-discovered path
  source unless `WithThreadPathSource` is also supplied. Independently supplied paths must survive
  either functional-option order; typed-nil options must neither panic nor counterfeit support.
- `WorkspaceSource` transports semantic Thread reads and local paths independently. A split
  workspace must never pair remote Thread contents with an unrelated complete local store's path
  resolver. The filesystem workspace adapter supplies both explicitly.
- Missing path capability returns a typed normal-domain error rather than a panic, empty successful
  path, invented URI, or misleading not-found result. Under a local source, existing resolver errors
  propagate without reclassification.
- The filesystem resolver remains filename-based and parse-free so malformed Thread frontmatter is
  still locatable for repair. Existing exact/prefix/substring resolution, ambiguity, stable filename
  identity, symlink-root spelling, repository-planner exclusion, human output, and JSON output remain
  unchanged.
- `Layout` continues to own watcher directories, guarded mutation ports continue to authorize
  writes, and neither concern leaks into the two read capabilities.

## Required hostile angles

1. Re-derive the interface boundary from consumers. Search every `ThreadStore` implementation and
   use. Prove no semantic Thread operation still calls or assumes `ResolveThreadPath`, and no path
   operation unnecessarily requires `ThreadStore` or `TaskGraphSource`. Look for hidden compile-time
   coupling in fakes, CLI completion, workspace/TUI setup, lint, compose, or generated graph reads.
2. Attack `NewService` functional-option semantics. Exercise aggregate-only construction,
   path-only and read-only construction, explicit read replacement, explicit path replacement, both
   option orders, repeated options, the same object supplied through multiple ports, typed-nil
   aggregate/read/path values, and sources whose methods return errors. Look for order dependence,
   stale fallback, surprising last-write behavior, or a path source that can remain paired with the
   wrong Thread records.
3. Attack workspace composition separately. Try a complete local aggregate plus remote Thread
   reads with no path source, distinct explicit read/path adapters, absent narrow ports that should
   preserve aggregate defaults, typed-nil ports, pointer workspaces, and malformed adapter results.
   Confirm `WorkspaceService` cannot silently cross-wire capabilities before handing the service to
   the TUI or a future primary adapter.
4. Attack local path recovery. Use malformed and missing frontmatter, malformed YAML, duplicate or
   ambiguous slugs, filename/frontmatter ID drift, invalid queries, missing directories, a symlinked
   planning root, non-regular files, and a repository-planner callback. Confirm path lookup performs
   no semantic parse and retains existing error identity and path bytes.
5. Inspect CLI and machine behavior. Verify ordinary local `thread path` human and JSON output are
   unchanged, missing capability is classified through the standard exit/error machinery, errors
   stay off successful stdout, and completion behavior does not accidentally make path resolution
   depend on portable Thread reads at execution time.
6. Review contract ownership and future portability. Challenge whether `ThreadPathSource` belongs in
   core, whether its name and promise overfit the filesystem, whether a future URL/remote locator is
   being precluded, and whether the anti-cross-wire rule is explicit enough for future composition
   roots. Keep general remote URLs and splitting every other entity path out of scope unless this
   patch makes them materially harder.
7. Assess test quality with mutation probes where useful. Remove the path detachment, reverse option
   order, restore `ResolveThreadPath` to `ThreadStore`, skip workspace path injection, make the path
   source typed nil, or force `GetThread` during path recovery; verify focused tests fail for the
   intended reason, then restore every probe. Flag assertions that only restate implementation
   details or leave a contract-bearing branch untested.
8. Compare code, Go comments, task acceptance criteria, ADR, architecture guidance, README, and
   existing generated CLI docs. Identify stale claims that Thread reads are filesystem-shaped,
   overclaims about absolute/local paths, or documentation that obscures the explicit override rule.
9. Look for proportionality and unnecessary machinery. In particular, question the persistent
   `threadPathsExplicit` construction flag and propose a smaller design only if it preserves
   order-independence, typed-nil safety, complete-store defaults, and cross-wire prevention more
   clearly. Treat mere aesthetic preference as a settled concern, not a finding.

Run proportionate validation: focused and full tests, race tests, vet/static analysis, planning
lint, integration goldens, generated-doc drift checks if relevant, `go mod tidy -diff`, and
`git diff --check`. Record exact commands and results. Do not install dependencies or change the
implementation to make a test possible. Restore all mutation probes and generated artifacts so the
worktree differs only by your edits to this audit file.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- the reviewed branch/base/worktree state and exact validation commands;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete minimum recommendation;
- a concise acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

If there are no findings, say so plainly and still document the hostile cases and settled concerns.
Do not edit implementation, task, ADR, Thread, generated artifacts, or any other planning file. Do
not create follow-up tasks or pre-resolve findings; the implementation owner will triage the report
using the established audit-finding lifecycle verbs.

---

## Reviewer report

### Executive verdict

**Ready with tracked follow-ups.**

This is a small patch that does exactly what it claims. The port split is real — I searched every
`ThreadStore` implementation and use in the repository and `ResolveThreadPath` now has precisely one
consumer, `Service.ThreadPath`. A pathless `ThreadStore` fake drives list, show, graph, compose, plan
and blocker reads with a nil aggregate store, and only `thread path` fails, with a typed
`ErrValidation`. The construction semantics survived a sixteen-case matrix: both option orders,
repeated options, typed-nil aggregates, typed-nil reads, typed-nil paths, and an embedded-nil `Store`
all behave correctly, and none panics. Five of six mutation probes on contract-bearing lines were
caught by a focused test failing for the intended reason; the sixth is expected to pass and I explain
why below rather than counting it as a gap.

Three low findings, none blocking. Two are undocumented asymmetries in the construction rules — the
anti-cross-wire guard runs in one direction only, and a source implementing both ports silently loses
path resolution unless passed to both options. The third is a one-word documentation overclaim.

On the brief's proportionality question: I tried to design `threadPathsExplicit` away and could not.
Any formulation that keeps order-independence, typed-nil safety, aggregate defaults, and detachment
needs to distinguish "paths came from the aggregate" from "paths were supplied", and a boolean is the
smallest honest way to hold that. I record it as settled, not a finding.

### Reviewed state and validation

Main worktree, based at `89cfb85`, deliberately dirty: `18 files changed, 402 insertions(+),
67 deletions(-)`, 27 `git status --short` lines including untracked. Machine: Apple M5, darwin 25.6.0.
Restored to exactly this state after every probe — `shasum -c` over the four patched sources, and all
four temporary probe files deleted and confirmed absent.

| Command | Result |
| --- | --- |
| `go build ./...` · `go vet ./...` | clean |
| `go test ./...` | all packages pass |
| `go test -race -count=1 ./internal/core/... ./internal/store/... ./internal/cli/... ./internal/workspacestore/... ./internal/tui/...` | pass, no race reports |
| `just lint` (golangci-lint) | 0 issues |
| `./bin/tskflwctl lint` · `audit lint` | both clean |
| `go mod tidy -diff` | clean |
| `git diff --check` | clean |
| `just docs` + `diff -rq` against a pre-run copy | no drift |
| Live `thread path` matrix (10 invocations on a scratch repo) | see settled concerns 4 and 5 |

Mutation probes, each applied then reverted and checksum-verified:

| Probe | Caught by |
| --- | --- |
| remove path detachment in `WithThreadStore` | `TestServiceThreadPathDefaultsAndExplicitSourcesDoNotCrossWire`, `TestWorkspaceService_ExplicitThreadReadsDoNotBorrowAggregatePaths` |
| reverse option order in `WorkspaceService.Open` | *nothing — and correctly so; see settled concern 2* |
| drop `WithThreadPathSource` from `Open` | `TestWorkspaceService_OpenAssemblesRuntimeAndPreservesSelection`, `TestFS_OpenWorkspaceResolvesDirectAndPointerEntries` |
| accept a typed-nil path source | `TestServiceThreadPathRejectsMissingAndTypedNilCapabilities` |
| remove aggregate path discovery in `NewService` | 1 core test, `TestGolden_MachineContract`, `TestThreadNewListShowPathAndFrontier` |
| make the FS resolver parse the document | `TestResolveThreadPathRemainsParseFreeForMalformedDocuments`, `TestResolveThreadPathPreservesSymlinkedPlanningRoot`, `TestFS_OpenWorkspaceResolvesDirectAndPointerEntries` |

## Findings

### Low

#### L1. The anti-cross-wire guard runs in one direction only  · **Status:** wontfix

**File:** `internal/core/service.go:123-134`, `internal/core/workspace.go:105-109` | **Component:** core/service-construction
**Effort:** XS · **Urgency:** soon

`WithThreadStore` detaches an aggregate-discovered path source, so explicitly replaced Thread reads
cannot borrow a local resolver. The mirror case is unguarded: explicitly supplied paths are happily
paired with aggregate-discovered reads.

**Reproduction** — one complete local aggregate implementing `Store` + `ThreadStore` +
`ThreadPathSource`, plus a foreign path resolver:

```
baseline: complete local everywhere            reads=LOCAL-THREAD   paths=/local/threads/x.md
GUARDED:  remote reads, no paths supplied      reads=REMOTE-THREAD  paths=nil                  ✓
REVERSE:  foreign paths, no reads supplied     reads=LOCAL-THREAD   paths=/FOREIGN/threads/x.md
REVERSE via NewService directly                reads=LOCAL-THREAD   paths=/FOREIGN/threads/x.md
```

The third row is reachable through `WorkspaceSource{Store: local, ThreadPaths: foreign}` with
`Threads` left nil — a shape `Open` accepts without complaint. `thread show X` then renders the local
Thread while `thread path X` prints a foreign file, which is precisely the hazard the detach rule
exists to prevent, only mirrored.

I am not claiming this is likely: supplying a path source without a matching read source is a
deliberate and odd act, and the production adapter (`workspacestore/fs.go`) supplies both. It matters
because the design's stated premise is that cross-wiring is prevented *by construction*, and here it
is prevented by convention in one direction and by nothing in the other. Neither `WorkspaceSource`'s
doc comment nor `docs/ARCHITECTURE.md` mentions the asymmetry.

**Recommendation:** State the invariant where the struct is defined — reads and paths must describe
the same Thread corpus — and have `Open` reject a source that supplies `ThreadPaths` while leaving
`Threads` nil and the aggregate `Store` capable of Thread reads. Symmetric detachment would be wrong
(it would break the legitimate local-plus-explicit-paths composition), so a validation is the smaller
correct move.

**Resolution:** An explicit path option is an affirmative override, unlike the
hidden aggregate fallback that must be detached after a read override. Rejecting
ThreadPaths without Threads would forbid legitimate path decorators over
aggregate-discovered reads, and neither narrow port exposes backend identity
with which to validate corpus equality. The intentional asymmetry and
composition-root compatibility obligation are now documented in Service,
WorkspaceSource, ADR-0006, and the architecture guide; a regression test pins
the legitimate aggregate-read plus explicit-path composition.

#### L2. A source implementing both ports silently loses path resolution unless passed to both options  · **Status:** fixed

**File:** `internal/core/service.go:123-147` | **Component:** core/service-construction
**Effort:** XS · **Urgency:** eventually

`WithThreadStore` unconditionally detaches paths when they were not explicitly supplied, including
when the newly supplied read source itself implements `ThreadPathSource`.

**Reproduction** — one fake implementing both interfaces:

```
source implementing BOTH via WithThreadStore only   threads=*bothPortsFake  paths=nil
        -> ThreadPath: ERR validation failed: thread path resolution is unavailable
source implementing BOTH via both options           threads=*bothPortsFake  paths=BOTH
        -> ThreadPath: OK BOTH
```

Safe-by-default is the right instinct here, and I would not change the behaviour. The problem is
discoverability: a future split adapter — a caching local reader, an offline mirror — that implements
both capabilities on one value and is wired through `WithThreadStore` alone loses `thread path` with
a message ("unavailable from this service") that points at the service rather than at the missing
second option. That is a slow diagnosis for exactly the consumers this patch is built for.

**Recommendation:** Document the rule on `WithThreadStore` and in the `WorkspaceSource` comment: a
value that implements both ports must be supplied to both options. One sentence at each site is
enough; no behaviour change.

**Resolution:** Documented on WithThreadStore and WorkspaceSource that one value
implementing both ports must be supplied through both options; the service
deliberately does not infer path capability from an explicit ThreadStore
argument.

#### L3. `ThreadStore` is documented as persistence-neutral while its problem channel is still filesystem-shaped  · **Status:** fixed

**File:** `docs/ARCHITECTURE.md:221` vs `internal/core/store.go:109-112` | **Component:** docs/architecture
**Effort:** XS · **Urgency:** soon

The architecture text now reads "`ThreadStore` is a separate persistence-neutral semantic read port."
The port is:

```go
type ThreadStore interface {
	ListThreads() ([]domain.Thread, []domain.FileProblem, error)
	GetThread(ref string) (thread domain.Thread, body string, err error)
}
```

`domain.FileProblem` is `{Path, Message}`, and the returned `domain.Thread` carries `Path`,
`FilenameID`, and `SourceVersion`. Compare the sibling port that *has* completed this work,
`TaskGraphSource`, which is now `ReadTaskGraph() (TaskGraphRead, error)`.

The gap itself is correctly scoped out of this patch — ADR-0006's 2026-09-01 amendment item 4 names it
explicitly ("`ThreadStore.ListThreads` still does"), and task `6g5rxq1ravd3` is filed and `next-up`.
So this is not deferred work masquerading as done; it is one word in the architecture guide that has
run ahead of the code, in the same document that a future adapter author will read to decide what the
port guarantees.

**Recommendation:** Say "semantic read port" and note that its diagnostic channel is not yet
persistence-neutral, referencing `6g5rxq1ravd3`. Promote the wording when that task lands.

**Resolution:** Corrected the architecture guide to call ThreadStore a semantic
read port and explicitly note that its filesystem-shaped list diagnostics remain
tracked by task 6g5rxq1ravd3.

## Acceptance-criteria traceability

| # | Criterion (abbreviated) | Verdict | Evidence |
| --- | --- | --- | --- |
| 1 | Pathless `ThreadStore` fake supports list, show, compose, projection, plan, graph without filesystem methods | Met | `var _ ThreadStore = (*threadReadFake)(nil)` compiles with no `ResolveThreadPath`. Probe on `NewService(nil, WithTaskGraphSource, WithThreadStore)`: `ListThreadViews`, `ShowThread`, `ShowThreadGraph`, `ComposeThreadApply`, `TaskBlockers` all returned `nil` error; only `ThreadPath` failed. |
| 2 | `thread path` byte-compatible locally; explicit typed failure when the capability is absent | Met | Human and `--json` output unchanged on a live repo (schema 1.58, `{"path": ...}`); no golden drift. Missing capability returns `%w: domain.ErrValidation` → exit 11 through the standard machinery. |
| 3 | Workspace tests prove independent injection and no cross-wiring | **Qualified** | `TestWorkspaceService_ExplicitThreadReadsDoNotBorrowAggregatePaths` and the `Open` probes pin the guarded direction. The reverse direction is neither guarded nor tested — L1. |
| 4 | Malformed local Threads remain path-resolvable when `GetThread` cannot parse | Met | Live: corrupting frontmatter left `thread path` at exit 0 with the correct path while `thread show` failed at exit 11. Probe forcing the resolver to parse fails three tests across two packages. |
| 5 | Typed-nil and missing-capability cases do not panic; aggregate defaults retained | Met | Sixteen-case construction matrix: typed-nil pointer aggregate, embedded-nil `Store`, typed-nil read option, and typed-nil path option all produced a typed error or a correct no-op. No panic anywhere. Aggregate-only construction still resolves paths. |
| 6 | Architecture and ADR distinguish semantic reads, optional paths, layout watches, guarded mutations | **Qualified** | All four are now named and separated in `docs/ARCHITECTURE.md` and the ADR amendment. Two construction rules are undocumented (L1, L2) and one claim overshoots the code (L3). |

## Settled concerns

Chased and disproved by code inspection plus a hostile reproduction.

1. **Residual coupling between semantic reads and paths.** A repository-wide search finds
   `ResolveThreadPath` referenced in exactly three non-test places: the interface declaration, the FS
   implementation, and `Service.ThreadPath`. No list, show, compose, projection, plan, graph, lint,
   or CLI-completion path touches it. `ThreadPath` no longer consults `s.threads` at all, so paths do
   not require reads either. The split is real in both directions.
2. **Option order inside `WorkspaceService.Open`.** Reversing `WithThreadStore` and
   `WithThreadPathSource` leaves the suite green — and that is the correct result, not a coverage
   hole. `threadPathsExplicit` makes the pair commutative by design, and
   `TestServiceThreadPathDefaultsAndExplicitSourcesDoNotCrossWire` asserts both orders explicitly at
   the `NewService` level. My probe confirmed the behaviour is genuinely identical rather than
   accidentally untested.
3. **`threadPathsExplicit` as unnecessary machinery.** I tried to remove it. Every alternative I
   constructed — resolving aggregate defaults last in a builder, comparing `s.threadPaths` against
   `s.store`, or dropping detachment and pushing the rule to composition roots — either reintroduces
   order dependence, breaks on a value that supplies paths through embedding, or gives up the
   cross-wire guarantee. Distinguishing "discovered" from "supplied" is irreducible here, and a
   boolean is the smallest way to hold it. Aesthetic preference only; not a finding.
4. **Local path recovery regressions.** Malformed YAML, an unterminated frontmatter block,
   filename/frontmatter ID drift, ambiguous slug prefixes, a missing `threads/` directory, a
   directory named `*.md`, an empty query, and two path-traversal attempts all behave exactly as
   before: parse-free filename identity, `ErrNotFound` for a frontmatter-only ID,
   `ErrAmbiguous` for a shared prefix, and a plain-name validation refusal for separators. The
   symlinked-root spelling is preserved.
5. **CLI and machine behaviour.** Human output is the bare absolute path; `--json` is
   `{"schema_version":"1.58","path":...}`; errors stay on stderr with stdout empty; exit codes are 10
   for not-found and 11 for validation. `completeThreadSlugs` globs `threads/*.md` directly
   (`internal/cli/completion.go:236-243`), so shell completion never depends on either read port at
   execution time.
6. **Guarded mutation and watcher leakage.** No mutation port changed; `Layout` still owns
   `WatchPaths` alone; neither read capability gained a write or watcher method. `thread path` cannot
   authorize anything — it has no mutation path.
7. **Future portability of the port name.** `ResolveThreadPath(ref) (string, error)` returns an
   opaque string, so a future remote locator returning a URL is not precluded by the signature; the
   README, ADR, and architecture all now scope it as explicitly *local*, which is the honest framing
   for what ships today. Splitting other entity paths and general remote URLs remain out of scope and
   are not made harder by this patch.
8. **Deferred neutrality work masquerading as complete.** `ThreadStore`'s filesystem-shaped
   `[]domain.FileProblem` is a real remaining gap, but it is named in ADR-0006's 2026-09-01 amendment
   item 4 and filed as `6g5rxq1ravd3` (`next-up`). Correctly out of scope here; only the architecture
   wording overreaches (L3).
