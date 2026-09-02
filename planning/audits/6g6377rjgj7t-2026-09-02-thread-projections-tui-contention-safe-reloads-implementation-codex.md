---
schema: 1
id: 6g6377rjgj7t
bucket: closed
area: thread-projections-tui-contention-safe-reloads-implementation-codex
date: "2026-09-02"
updated_at: "2026-09-02"
---
# Audit: Thread projections in the TUI with contention-safe reloads — codex — 2026-09-02

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

## Review brief

Perform an independent adversarial implementation and architecture review of the uncommitted work for task `6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads`. Assume the implementation may be subtly wrong even when the happy-path tests pass. Seek concrete defects, boundary erosion, stale-result races, misleading UI states, and missing hostile tests; do not manufacture findings when inspection and reproduction settle a concern.

## Review target

Review the full uncommitted diff from `main` at `f6096d3`, especially `internal/tui/read_retry.go`, `thread_projection.go`, their tests, and changes to the TUI model, messages, entity/detail/dashboard/session/view plumbing and `docs/ARCHITECTURE.md`. Compare it to the completed task contract, ADR-0006, the core Thread projection/read ports, the workspace/session boundary, watcher debounce/reconciliation, and every existing async list/detail/dashboard consumer. Build a repository-wide consumer inventory before judging whether the shared retry abstraction is complete or correctly scoped.

## Intended contract to challenge

- The TUI retains `core.ThreadListView`, `core.ThreadView`, and `ThreadReadProblem` without deriving competing lifecycle, readiness, gate, frontier, health, or diagnostic semantics.
- `TaskGraphSource` and `ThreadStore` remain independently injectable through workspace construction.
- Thread projections are lazy until a future visible consumer asks for them, then both task and Thread changes invalidate every requested Thread projection.
- Only a read-side `domain.ErrConflict` is transient: retain last coherent state, delay exactly one retry by one quiet period, drop it when its generation/session/selection is stale, and expose a repeated conflict without spinning.
- The policy behaves consistently for entity lists, entity detail, dashboard, and Thread list/detail, without unrelated reload amplification or hidden first-load false emptiness.
- Workspace switches cannot accept commands, retry closures, or cached Thread projections from the previous service/session.
- Graph loading remains deferred until a topology consumer exists.

## Mandatory evidence floor

1. Inspect the complete diff and trace every producer and reducer consumer for each retryable message. Produce a consumer inventory, including session scoping and nested `tea.BatchMsg` behavior.
2. Run focused tests repeatedly and under `-race`; run the full relevant package suite. Do not treat existing tests as proof without checking assertion strength.
3. Exercise real planner-window contention, first-load and loaded-refresh states, repeated conflict, newer generations, selection changes, workspace switches, task-only invalidation, Thread-only invalidation, unrequested/lazy projections, and a durable non-conflict failure.
4. Use at least one mutation-testing probe: temporarily break a key stale guard, retry bound, retained-state rule, split-port choice, or reload trigger; identify whether the suite fails for the intended reason, then restore the file exactly.
5. Inspect resource/concurrency behavior: timers, held closures/services, repeated watcher bursts, nested batches, and whether retries can survive or amplify mutations and workspace changes.
6. Verify no graph projection is loaded invisibly and no TUI-local graph traversal or readiness calculation was added.

## Required hostile angles

- Generation identity collisions across separate entity kinds, Thread refs, dashboards, and restored workspace sessions.
- A conflict message queued before a workspace switch and a retry tick delivered after it; prove both outer session scoping and inner request identity are sufficient.
- Multiple simultaneous surface conflicts from one watcher reload; prove each gets at most one isolated retry and no retry invokes `reloadAll`.
- First-load conflict followed by another event/manual refresh before the timer; prove there is neither a false empty result nor an orphaned permanent loading state.
- Detail retention when the selected identifier aliases or resolves differently from the displayed canonical title; check slugs versus stable IDs and duplicate/ambiguous refs.
- Failed Thread detail for a new ref after a previously successful ref; ensure stale data cannot be mistaken for the selected Thread by the next UI slice.
- Session caching and first-visit/restored-workspace paths, including zero generations, unseen Thread surfaces, and old cached errors.
- Test helpers that synchronously drain commands: ensure they do not conceal production ordering, timer, batch, or scoping defects.
- Documentation claims versus actual lazy loading, invalidation, error visibility, and portability boundaries.
- Supported-platform behavior only; distinguish an actual defect from the documented advisory-lock/raw-writer limit.

## Validation and restoration

Keep the implementation read-only except for temporary adversarial probes. Restore every probe before reporting and verify `git diff` contains only the pre-existing implementation plus this audit. Run `go test ./internal/tui`, repeated focused tests, `go test -race ./internal/tui`, `go test ./...`, `go vet ./...`, `git diff --check`, and planning lint as evidence permits. Do not fix findings during the review.

## Deliverable

Replace the placeholder below with a concise executive verdict, then findings ordered High/Medium/Low. Each finding must include exact file/line evidence, component, effort, urgency, a concrete reproduction or reasoning chain, and a bounded recommendation. Use the repository finding grammar exactly: `#### M1. <title> · **Status:** open` (or H1/L1). Leave every finding open for implementation-owner triage. Include survived-adversarial-review claims and exact validation evidence so disproved suspicions are not rediscovered later.

## Reviewer report

**Verdict:** not ready to treat the projection foundation as settled. The bounded read-conflict
policy, split-port composition, stale-result guards, and core-projection fidelity are sound. Two
state-model seams remain before the read-only Thread tab should build on this work: loaded detail
identity is not modeled independently from request/presentation identity, and the new Thread cache
does not yet connect to the entity registry it is intended to serve.

## Findings

### Medium

#### M1. Detail caches conflate requested, loaded, and presentation identity · **Status:** fixed

**File:** `internal/tui/thread_projection.go:17-29,74-81`;
`internal/tui/model.go:340-360,399-410`; `internal/tui/detail.go:148-167,200-204` ·
**Component:** TUI detail state / stale-content retention · **Effort:** M · **Urgency:** soon

The Thread cache has one `detailRef`, which is changed as soon as a request starts, plus a
`detailOK` bit and payload left over from the last successful request. On a current non-conflict
failure the reducer changes only `detailErr`. A direct adversarial reproduction loaded `delivery`,
then requested `does-not-exist`; the settled state was `detailRef="does-not-exist"`,
`detailOK=true`, and a payload whose Thread ID was still `6g503c6pfqeb`. A future renderer cannot
interpret `detailOK` as "the selected Thread is loaded," and the cache no longer carries the ref to
which the retained payload belongs. Retaining old content during an in-flight refresh is reasonable;
claiming that it belongs to the new selection after a durable failure is not.

The shared entity-detail path has the sibling identity problem. `detailPane.showing` decides whether
a repeated conflict may retain content by comparing `detailPane.title` (the display value returned by
`detailContent.Title`) with `detailErrMsg.id` (the request/selection key). That happens to work while
tasks, audits, and research still use slugs for both. The already-queued
`make-tui-entity-navigation-use-stable-identities` task explicitly changes internal request keys to
stable IDs while keeping slugs as detail titles. Without another identity field, that change makes
every repeated refresh conflict take `SetError` and discard the last coherent body, violating this
task's retention guarantee. Duplicate display slugs also make title equality an unsafe proxy for
same-record scroll retention.

**Recommendation:** give both detail state holders an explicit loaded canonical reference separate
from the latest requested reference and from the display title. Stamp it from the accepted message,
retain content only when loaded and requested canonical references match, and clear or mark content
as belonging to another selection after a durable failure. Add regressions for successful A then
failed B, stable-ID request with a different display slug, duplicate slugs, same-record conflict
retention, and selection changes while a retry is pending. Coordinate this with the existing stable-
identity task rather than adding a second identity migration later.

**Resolution:** Entity detail panes now store canonical loaded identity
separately from display title, and Thread detail state separates requested from
loaded refs. Regressions cover stable IDs, duplicate display slugs, and a
successful A load followed by failed B.

#### M2. The Thread cache is a parallel state machine with no production registry consumer · **Status:** tracked by 6g5rwjqr7rt8

**File:** `internal/tui/thread_projection.go:17-29,65-97`; `internal/tui/model.go:232-257,384-410`;
`internal/tui/entity.go:65-92,297-355`; `planning/tasks/6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md` ·
**Component:** TUI architecture / next-slice integration · **Effort:** M · **Urgency:** soon

The production consumer inventory contains no call to `requestThreadList` or
`requestThreadDetail`; their only direct consumers are tests. `reloadThreadProjections` is wired into
`reloadAll`, but its zero-value generation gates mean it remains a no-op until one of those otherwise
unreachable request methods has been called. This is consistent with "no invisible prefetch," but it
also means the completed task has not exercised a production entry path.

More importantly, the intended consumer already has a different list/detail state machine.
`entityTab` owns list generations, loaded/error/problem state, cursor restoration, filtering, and
its `loadList`/`loadItem` commands; `handleListLoaded` owns active-detail chaining. The downstream task
requires Threads to join that existing registry. The side cache independently owns another pair of
generations, errors, loaded bits, detail selection, messages, reducer branches, session storage, and
reload fan-out. Adding the Thread tab as currently scoped therefore requires either duplicating each
Thread result into both owners, adding Thread-only registry branches, or discarding much of this
foundation and reimplementing the loaders as ordinary entity loaders. Each option makes two sources
of truth or immediate rewrite the default path.

**Recommendation:** before implementing the visible tab, make one state owner explicit. Either let
the entity registry own Thread list/detail async state while its row/detail adapters retain complete
`core.ThreadView` values, or define a small registry projection seam through which the typed Thread
cache supplies loaded/error/generation/selection state without duplicating it. Add an integration
test that enters the future registry route and proves one request, one generation, one error owner,
and one reload path. Keep graph traversal and semantic projection in core; this is a TUI state-
ownership correction, not a reason to flatten the core model.

**Resolution:** The visible Thread task now requires the entity registry to
become the single owner of list/detail generations, errors, canonical selection,
and reloads, deleting or collapsing the temporary parallel cache.

#### M3. Shared-policy claims exceed the entity and dashboard regressions · **Status:** fixed

**File:** `internal/tui/read_retry_test.go:39-197,238-354`;
`internal/tui/dashboard_test.go:405-462`; `internal/tui/model.go:347-360,367-378,1200-1214` ·
**Component:** TUI contention/staleness tests · **Effort:** S · **Urgency:** soon

The strongest tests establish the retry bound with a Thread-list message and exercise a real
planner window across several loaded surfaces, but they do not pin the final-error and stale-request
branches for every surface the task now claims shares the policy. A coverage run reports
`detailPane.SetRefreshError` and `detailPane.showing` at 0%; no test lets an already-loaded entity
detail conflict both initially and on its one retry, so the newly added retained-body/footer outcome
can be inverted or removed without an assertion. The existing dashboard error tests inject
`dashLoadedMsg` with implicit generation zero, and no test sends an older dashboard success or error
after a newer generation. `TestNewerGenerationSupersedesAScheduledRetry` proves the Thread-list and
Thread-detail cases, not the entity-list, entity-detail, and dashboard arms of
`readRequestCurrent`.

This matters because the policy is implemented through two closed type switches. A later surface or
identity change can silently skip conflict extraction or stale-request rejection unless tests enumerate
the supported surfaces. The package's high aggregate coverage does not expose that branch-level gap.

**Recommendation:** add a table that exercises current and superseded requests for all five
`readSurface` values, a stale dashboard success/error regression with explicit generations, and an
already-loaded entity-detail case whose first and retry reads both conflict and therefore retain the
body with the visible footer warning. Make the table fail when a new `readSurface` is added without a
case/test (for example with a count sentinel). These tests should accompany M1's canonical loaded-ref
fix so the identity and retention contract are tested together.

**Resolution:** Tests now enumerate current and superseded requests for all five
read surfaces, pin stale dashboard success and error handling, and exercise
repeated entity-detail conflict retention through the visible footer outcome.

### Low

#### L1. The real planner-window helper can hang instead of reporting setup failure · **Status:** fixed

**File:** `internal/tui/read_retry_test.go:200-229` · **Component:** TUI contention test harness ·
**Effort:** XS · **Urgency:** soon

`holdPlannerWindow` launches `MutateTaskGraph` and then unconditionally waits on `<-inside`.
`inside` is closed only from the planner callback. If graph loading, graph validation, or guard setup
fails before the callback—for example after a future broken-graph fixture is reused—the goroutine
closes `done` with `mutateErr`, but the test goroutine can never observe either and waits until the
global Go test timeout. That converts an ordinary fixture/setup error into a package hang and hides
the useful cause. The upcoming Thread view task explicitly calls for missing/broken projection
fixtures, so this is a plausible evolution of the test corpus rather than a theoretical malformed
call.

**Recommendation:** wait for either `inside` or `done`; if `done` wins, fail immediately with
`mutateErr`. Keep the cleanup idempotent, and add a small broken-graph fixture test proving the helper
reports rather than hanging.

**Resolution:** Planner-window setup waits for either callback entry or mutation
completion, reports the underlying setup error, and has a broken-graph
regression proving the helper returns instead of hanging.

## Survived adversarial review

- `ListThreadViews` and `ShowThread` results are retained as the core types; no TUI-local readiness,
  frontier, gate, health, or topology calculation was added, and `ShowThreadGraph` is not loaded.
- Workspace construction preserves distinct aggregate, `TaskGraphSource`, and `ThreadStore`
  capabilities. The split fake uses values absent from the aggregate store, so the assertion is
  non-vacuous.
- The first conflict is replaced by a delayed message carrying the unwrapped original command. A
  repeated conflict therefore lands as the ordinary surface error and cannot schedule a third read.
  Repeated focused execution was stable across 20 runs.
- List, detail, dashboard, and Thread result types all enter the same classifier. Generations and
  selection/ref checks drop old successes, old failures, and superseded timers independently per
  surface; one retry invokes its own read rather than `reloadAll`.
- An explicit session probe queued a retry under generation 4, advanced the workspace session to 5,
  and confirmed the scoped timer result was dropped without executing its held old-service command.
- A mutation probe disabled the Thread-list generation check; `TestStaleThreadResultsAreDropped`
  failed on the intended overwrite assertion. The probe was restored before this report.
- The real planner-window fixture demonstrates first-load and loaded-state contention, while task-
  only and Thread-document edits both reach already-requested Thread list/detail caches through the
  common reload path. The store watcher already includes `threads/`.

## Validation

- `go test ./internal/tui` — passed.
- Focused contention/staleness/reload tests with `-count=20` — passed.
- `go test -race ./internal/tui` — passed.
- `go test ./...` — passed.
- `go vet ./...` — passed.
- `tskflwctl lint` — passed.
- `git diff --check` — passed.
- `go test -coverprofile=/tmp/taskflow-tui-review.cover ./internal/tui` — passed at 88.1%; the
  function report supplied the branch-coverage evidence in M3.
