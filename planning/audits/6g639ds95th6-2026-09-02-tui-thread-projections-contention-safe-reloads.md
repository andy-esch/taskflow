---
schema: 1
id: 6g639ds95th6
bucket: closed
area: tui-thread-projections-contention-safe-reloads
date: "2026-09-02"
updated_at: "2026-09-02"
---
# Audit: tui-thread-projections-contention-safe-reloads — 2026-09-02

> Adversarial implementation review of the uncommitted work for task
> `6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads`
> (TUI Thread projection cache + the shared read-contention retry policy).
> Findings are left **open** for implementation-owner triage; nothing here has been fixed.

## Verdict

No crash, deadlock, or data-loss defect in production code. Architecture non-negotiables all
hold: no globals, no fs/cobra in `core`, no I/O in `Update`/`View`, every read through
`core.Service` as a `tea.Cmd`, errors classified via `domain.Classify`. The real defects are
(a) one test-infrastructure hang that the *next* task's fixtures will trigger, (b) a
user-visible feature that is entirely unproven and survives inversion, (c) three generation
guards that no test pins, and (d) one state-shape omission that is cheap now and expensive
after the Threads tab lands.

**Validation reproduced:** `just build` · `go test ./...` · `golangci-lint run ./...` (0 issues) ·
`go test -race -count=5 ./internal/tui` · `-shuffle=on` ×3 · `-race -count=20` on the three
contention tests. All green.

**Method:** 23 targeted mutations of the production code, checking which tests actually fail.
Killed 11, **survived 9** — the survivors drive the findings below.

## Findings

#### H1. Test planner-window helper hangs the package when the guarded mutation fails early  · **Status:** fixed

**File:** internal/tui/read_retry_test.go:206 | **Component:** tui/tests
**Effort:** XS · **Urgency:** acute

`holdPlannerWindow` blocks on `<-inside` with no fallback. `MutateTaskGraph` returns *before*
`close(inside)` whenever `LoadTaskGraph` errors or `ValidateTaskGraphMutationSource` sees
`GraphBroken` — a dangling `depends_on`, a cycle, a duplicate id, or a flock failure. The
goroutine then exits, `inside` is never closed, and the whole package hangs until the
`go test` timeout. `mutateErr` holds the real reason and is never surfaced.

Reproduced: adding one dangling dependency to the `threadRepo` fixture hangs the suite until
`-timeout 30s` panics; the stack names `holdPlannerWindow` and the actual cause is invisible.
This is exactly the fixture shape the next task calls for ("missing members",
"completed-but-unsound", "broken evidence"), so it is a trap armed for the next author.

**Recommendation:** `select { case <-inside: case <-done: t.Fatalf("planner window never opened: %v", mutateErr) }`.

**Resolution:** Planner-window setup now selects between callback entry and
early mutation completion and reports the real failure. A broken
dependency-graph fixture pins the non-hanging path.

#### H2. The retained-detail feature is 0% covered and survives inversion  · **Status:** fixed

**File:** internal/tui/detail.go:195 | **Component:** tui/detail
**Effort:** S · **Urgency:** acute

Coverage: `SetRefreshError` **0.0%**, `showing` **0.0%**; the `case m.detail.refreshFailed():`
footer branch (view.go:473) is never taken by any test. Three mutations survive with the full
suite green:

1. delete the footer branch entirely;
2. delete the `m.detail.showing(msg.id)` guard (model.go:357);
3. **invert `SetRefreshError` to set `d.content = nil`** — i.e. make it blank the pane, the
   exact opposite of its documented contract.

The behaviour is also the weaker of the change's two scope calls: the acceptance criterion
governs the *conflict-retry* policy, while this is a separate UX decision about what a
*durable* (post-retry) conflict does to the detail pane, and it introduces a new user-visible
string (`⚠ detail refresh failed`).

**Recommendation:** Either pin it (a detail read that conflicts twice keeps `m.detail.content`
and renders the footer note; a conflict for a *different* id still takes `SetError`) or cut it
and let the durable conflict fall through to the existing error pane, deferring the retain to
`6g5rwjqr7rt8` where the Thread detail UX is decided.

**Resolution:** A repeated entity-detail conflict now has an end-to-end
regression proving same-record content and the footer warning survive, while a
different canonical loaded record takes the error path.

#### H3. Thread detail cache cannot say which Thread its resolved values belong to  · **Status:** fixed

**File:** internal/tui/thread_projection.go:16 | **Component:** tui/thread_projection
**Effort:** XS · **Urgency:** acute

`detailRef` is the ref the newest request **asked for**, set by `requestThreadDetail` before the
read resolves. `detail` / `detailBody` / `detailOK` / `detailErr` are the **last resolved**
values. Nothing links them.

Failure: select Thread A (loads) → select Thread B (conflict held for the 200ms quiet period).
The model now reports `detailRef == "B"`, `detailOK == true`, and `detail == A's view`, with no
field a renderer can consult to tell them apart. The same aliasing applies to `detailErr`: A's
error persists beside `detailRef == "B"`.

The precedent is in the file this change extends: `detailPane` carries `title` (the id its
`content` belongs to), and this very diff adds `showing(id)` to ask that question for entities —
then omits the equivalent for Threads. Cheap now; expensive once a renderer depends on it.

**Recommendation:** Add `detailLoadedRef`, set on the success branch of `threadDetailLoadedMsg`,
plus a `showingDetail()` helper. Mirror `SetError`/`SetRefreshError` in the reducer: retain the
previous projection only when the failure refreshes the *same* ref, otherwise clear it.

**Resolution:** Thread detail state now carries detailLoadedRef separately from
the requested detailRef; a failed new selection retains but cannot misidentify
the previous coherent payload.

#### M1. The dashboard's new generation stamp is unpinned  · **Status:** fixed

**File:** internal/tui/model.go:373 | **Component:** tui/model
**Effort:** XS · **Urgency:** soon

No test constructs a stale `dashLoadedMsg`; deleting the `msg.gen != m.dash.loadGen` guard leaves
the suite green. Compounding it, the three existing dashboard tests
(dashboard_test.go:415/438/456) inject `dashLoadedMsg{…}` with an implicit `gen: 0` and pass only
because `loadedDash` leaves `dash.loadGen == 0` — the moment anyone adds an `enterDash` or `r` to
those setups, their messages are silently dropped and the tests fail confusingly. The Thread
mirror of this test already exists (`TestStaleThreadResultsAreDropped`).

**Recommendation:** A stale-`dashLoadedMsg` test mirroring the Thread one, and stamp the existing
three with `m.dash.loadGen` so they stop depending on the counter being zero.

**Resolution:** Existing dashboard result fixtures use explicit current
generations, and a new regression proves stale success and stale error messages
cannot overwrite current rows or error state.

#### M2. `readRequestCurrent` is unpinned for three of its five surfaces  · **Status:** fixed

**File:** internal/tui/model.go:1200 | **Component:** tui/read_retry
**Effort:** XS · **Urgency:** soon

Surviving mutations: dropping `isCurrentSelection` from the `readEntityDetail` case; making
`readEntityList` always current; making `readDashboard` `return true`. Only the two Thread cases
are proven (`TestNewerGenerationSupersedesAScheduledRetry`). This function is the single
chokepoint for the whole retry policy and its doc comment promises "a retry can never resurrect a
load the model has already moved past".

Related: the `default: return false` arm silently loses retries for any surface added without a
case, and Go has no exhaustive-switch check here.

**Recommendation:** A table test over all five surfaces (current vs superseded). Consider a
trailing `numReadSurfaces` sentinel so the table's length can assert coverage of the enum.

**Resolution:** A table covers current and superseded identities for every read
surface and is tied to readSurfaceCount, including entity selection and
generation checks.

#### M3. Acceptance criteria were narrowed in the same change that ticked them  · **Status:** tracked by 6g5rwjqr7rt8

**File:** planning/tasks/6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md:52 | **Component:** planning
**Effort:** XS · **Urgency:** soon

Three narrowings landed in the same diff that ticked the criteria:

- AC1 dropped `ShowThreadGraph` / topology loaders. Defensible (the topology task owns it), but
  the criterion moved to meet the code.
- AC3 dropped **"and cursor/selection intent survives the refresh."** Nothing delivered addresses
  Thread cursor/selection intent, and the entity precedent for it (`restore`/`restoreGen`,
  entity.go:110) has no Thread analogue. The requirement was removed, not deferred to a destination.
- AC4 changed "bounded, **coalesced** retry" to "exactly one quiet-period retry". Coalescing is
  genuinely not implemented — five surfaces conflicting on one reload schedule five independent
  `tea.Tick`s. The behaviour is fine and proven; the word was deleted rather than satisfied.

"Out of scope" also gained "cursor and navigation behavior" in the same edit. This is the pattern
most likely to hide a real gap from a later reader.

**Recommendation:** Record the three narrowings as explicit deferrals with destinations (append to
the completed task, and carry the cursor/selection clause into `6g5rwjqr7rt8`'s scope) rather than
leaving them as silent edits to the criteria.

**Resolution:** The completed task now records all three scope decisions
explicitly. Cursor and selection ownership is carried by the tracked list/detail
task, graph loading remains in 6g5rwjr0dz4p, and per-surface rather than
cross-surface retry is documented as deliberate isolation.

#### M4. The parallel Thread cache will be largely deleted by the next task  · **Status:** tracked by 6g5rwjqr7rt8

**File:** internal/tui/thread_projection.go:16 | **Component:** tui/thread_projection
**Effort:** M · **Urgency:** eventually

`6g5rwjqr7rt8` says verbatim: add a Threads tab/list and detail route **through the existing
entity registry**, command palette, tab navigation, selection restore, filtering and sorting
conventions. That means an `entityTab{kind: entityThreads}` and a `threadDetail` implementing
`detailContent`. At that point `threads.list` / `loaded` / `loadErr` / `listGen` duplicate four
`entityTab` fields, the six `detail*` fields duplicate `detailPane` + `m.detailGen`, and
`readThreadList` / `readThreadDetail` collapse into `readEntityList` / `readEntityDetail` with
`kind: entityThreads`.

What genuinely does not fit the registry is small and identifiable: `ThreadListView.GraphHealth`
and `.GraphProblems` (repository-level, deliberately hoisted out of rows,
core/thread_projection.go:76) and `[]core.ThreadReadProblem` (a different type from
`entityTab.problems`'s `[]domain.FileProblem`). That is a ~3-field side-car, not a parallel cache
with its own generation counters.

Correctly out of scope for `6g5rwjqeh6a6` (which explicitly deferred the visible surface), but
`6g5rwjqr7rt8`'s scope and estimate should state that it *deletes* this structure rather than
building on it.

**Resolution:** The tracked task now requires one registry-owned Thread
list/detail state machine and explicitly deletes or collapses the temporary
parallel cache, retaining only repository-level diagnostics that do not belong
to a row.

#### M5. `readSurface` is a second "which surface" discriminator beside `entityKind`  · **Status:** tracked by 6g5rwjqr7rt8

**File:** internal/tui/messages.go:22 | **Component:** tui/read_retry
**Effort:** S · **Urgency:** eventually

The package already has one surface enum, deliberately extended with negative sentinels
(`entityDashboard = -1`, `entityAtlas = -2`, entity.go:29) precisely so non-tab screens can be
named without a tab slot. `readRequest` now carries **both** `surface` and `kind`, and for the two
Thread surfaces `kind` is silently `0` (= `entityTasks`) — a landmine for anyone who later adds
`kind` to a Thread comparison or a log line. `readRequest{kind entityKind, detail bool, id, gen}`
covers all five cases in existing vocabulary. Collapses together with M4.

**Recommendation:** Fold `readSurface` into `entityKind` when the Threads tab lands.

**Resolution:** The tracked registry-integration task explicitly folds the
temporary readSurface distinction into entity or screen identity where that
removes duplicate discriminators.

#### M6. The atlas is the one async surface left outside the shared policy  · **Status:** tracked by 6g63db3sdfrh

**File:** internal/tui/atlas.go:290 | **Component:** tui/atlas
**Effort:** S · **Urgency:** soon

`markAtlasStale` fires from `reloadAll` alongside every other surface and, when the atlas is on
screen, issues `loadAtlas(…)`. But `atlasLoadedMsg` is absent from `readMessageError`
(read_retry.go:42), and `handleAtlasLoaded` (atlas.go:158) stores any error durably. So the exact
scenario the flagship test simulates — one reload conflicting on every surface — would leave the
atlas showing a durable "repository mutation planner is active" while every other surface silently
retried and recovered. `TestPlannerWindowRetainsEveryLoadedSurface` misses it because its model is
constructed without `spaceOverviewSvc`.

Defensible against the acceptance criterion as finally worded (which lists four surfaces, not the
atlas) — see M3.

**Recommendation:** Bring `atlasLoadedMsg` into `readMessageError` and give the atlas a
`readSurface`, or state in the task record why the atlas is deliberately excluded.

**Resolution:** The proposed top-level atlasLoadedMsg wrapper cannot see
planner-window failures because Overview contains them per space. The tracked
task carries the stronger structured partial-failure and retained-summary design
and is sequenced in the production Threads DAG.

#### M7. `readMessageError` is a hand-maintained type list with no compile-time backstop  · **Status:** fixed

**File:** internal/tui/read_retry.go:42 | **Component:** tui/read_retry
**Effort:** S · **Urgency:** soon

Its own doc says "A read whose message type carries no error … simply opts out." Add a new
asynchronous read message, forget a `case`, and it silently loses the retry policy — invisible in
review, with no test failure. M6 is this defect already realised once.

Type-switching on the *result* is the least-bad option available, since each loader returns a
distinct concrete message and wrapping at the loader would mean touching eight loaders. The
mechanism, not the approach, is what needs hardening.

**Recommendation:** `type readResult interface{ readError() error }` implemented at each message's
definition site, so opting in is local to the type and an omission is visible where the type is
declared.

**Resolution:** Read-result messages now opt in beside their declarations
through readResult.readError; the centralized hand-maintained concrete-type
switch is gone and tests enumerate every participating result type.

#### M8. `sessionScope` is set only by `WithAtlas`, and this change raises the stakes  · **Status:** fixed

**File:** internal/tui/model.go:142 | **Component:** tui/session
**Effort:** XS · **Urgency:** soon

**Settled: no reachable path switches workspaces with `sessionScope` false**, in the shipped
composition root or any other composition. `sessionScope = svc != nil` is assigned only in
`WithAtlas`; the only `activateWorkspace` call is `handleWorkspaceOpened` (atlas.go:186); the only
`openWorkspace` call sites are `openAtlasSpace`/`openAtlasWork` (atlas.go:341), reachable only
from an atlas screen; `enterAtlas` hard-returns "atlas is unavailable in this session" when
`spaceOverviewSvc == nil` (atlas.go:263); `atlasInRing()` gates the `[`/`]` ring on the same field.
`internal/cli/ui.go:64` passes both options unconditionally.

The mechanism is real if that invariant ever breaks. `activateWorkspace`'s restored branch
(session.go:132) restores `tabs` (with `loadGen`), `dash` (now with `loadGen`), `detailGen`, and
now `threads` (`listGen`/`detailGen`/`detailRef`) **verbatim** — so on an A→B→A round trip a retry
held from the first visit to A would find byte-identical counters, pass `readRequestCurrent`, and
run against the *previous* `*core.Service`. `sessionGen` is monotonic and is the only thing
stopping it. This change adds two more restored counters and, uniquely, a service-capturing
closure deliberately held across a 200ms delay — longer than any pre-existing in-flight read —
without adding a guard.

**Recommendation:** Set `sessionScope` in `WithWorkspaceOpening` as well, or assert it at the top
of `activateWorkspace`, so the invariant is enforced where workspaces actually switch.

**Resolution:** Injecting WorkspaceService now enables session scoping directly,
independent of WithAtlas option presence or ordering, and a capability test pins
the invariant.

#### L1. The retry timing assertion pins "a delay", not "the quiet period"  · **Status:** fixed

**File:** internal/tui/read_retry_test.go:72 | **Component:** tui/tests
**Effort:** XS · **Urgency:** eventually

`tea.Tick` calls `time.NewTimer(d)` **outside** the returned closure (bubbletea v2.0.9
commands.go:154), so the 200ms starts when `Update` reduces the conflict, not when the command
runs. `start` is taken *after* `m.Update(conflict)`, so the assertion measures 200ms minus the gap
between those two statements; a >100ms deschedule there flakes it under CI load. Mutation check:
halving the delay to `fsDebounce/2` **survives** the test; 1ms is killed.

Also: the inline comment attributes the early firing to timer-boundary alignment, which describes
`tea.Every`, not `tea.Tick`.

**Recommendation:** Move `start := time.Now()` above `m.Update(conflict)` (free, strictly better,
and lets the bound tighten toward `fsDebounce`), and correct the comment.

Investigated and dismissed: a `Tick` created but never run does **not** leak — in production every
command returned from `Update` is executed by the runtime, and Go ≥1.23 garbage-collects
unreferenced timers whether stopped or not.

**Resolution:** The retry timing measurement now starts before Update creates
tea.Tick, uses the correct Tick semantics, and pins at least three quarters of
the configured quiet period.

#### L2. Test-helper sprawl, and one test reports only its first failure  · **Status:** fixed

**File:** internal/tui/read_retry_test.go:250 | **Component:** tui/tests
**Effort:** XS · **Urgency:** eventually

`TestPlannerWindowRetainsEveryLoadedSurface`'s four "did this surface blank" checks are arms of a
single `switch`, so **only the first failing surface is reported** — in the one test whose entire
purpose is "no surface may blank". They should be four `if`s. `len(retries) < 4` should be `!= 4`.

Separately there are now four drain helpers across three files: `drain` (model_test.go:49),
`drainBatch` (:64), `drainNested` (thread_projection_test.go), `drainConflicts`
(read_retry_test.go). `drainNested` strictly supersedes `drainBatch`, and `drainConflicts` is
`drainNested` plus one branch. Precedent is that shared helpers live in `model_test.go`.

**Recommendation:** Split the switch, tighten the count, and move `drainNested` to `model_test.go`
beside the helpers it generalises.

**Resolution:** The multi-surface retention test now uses independent assertions
and requires exactly four retry messages. The specialized drain helpers were
retained because they intentionally differ in whether follow-up commands and
held conflicts are consumed.

#### L3. `readConflictMsg` and `readRetryMsg` match only by coincidence  · **Status:** fixed

**File:** internal/tui/messages.go:41 | **Component:** tui/read_retry
**Effort:** XS · **Urgency:** eventually

The two are independent declarations that happen to have identical fields. Tests rely on the
conversion `readRetryMsg(conflict)` (which staticcheck S1016 suggested), and that conversion is
legal only as long as the two stay identical — adding a field to one silently breaks it.

**Recommendation:** `type readRetryMsg readConflictMsg`, which documents the relationship and
keeps the conversion valid by construction.

**Resolution:** readRetryMsg is now defined from readConflictMsg, documenting
and mechanically preserving the field-shape relationship used by conversions.

#### L4. Placement and documentation nits  · **Status:** fixed

**File:** internal/tui/detail.go:192 | **Component:** tui
**Effort:** XS · **Urgency:** eventually

- detail.go:192 — the comment cites `renderFooter`; no such function exists (it is `footer()`,
  view.go:363).
- messages.go:22 — `readSurface` and `readRequest` are neither messages nor used as messages;
  `read_retry.go` is their home. (The two message types are correctly placed.)
- docs/ARCHITECTURE.md:502 — a 105-character line in a section that otherwise wraps at 76–88.

**Resolution:** Read request types moved to read_retry.go, the footer comment
names the real footer path, and the architecture prose was wrapped consistently.

## Survived adversarial review

Recorded so these disproved suspicions are not re-investigated later.

- **`reloadAll`'s `|| m.onDash`** — in scope and load-bearing. Removing it kills
  `TestFirstLoadContentionIsNotAFalseEmptyState`. It mirrors the pre-existing active-tab carve-out
  verbatim and is the only thing making a contended first load recoverable. Keep.
- **The planner-window test is real and not deadlock-prone.** Reads take
  `rejectRepositoryPlannerCall` (an RWMutex read, store/lock.go:89) and never touch the write mutex
  the goroutine holds, so no deadlock path exists; `-race -count=20` clean. It is the only test
  that kills the three "unwrap the list / dashboard / detail read from the policy" mutations — the
  strongest test in the change. (Its hang risk is H1, a separate matter.)
- **Lazy Thread loading with no production caller** — defensible; the task explicitly deferred the
  visible surface, and `TestUnrequestedThreadProjectionsAreNeverRead` pins the discipline. The trap
  to name in the next task's criteria: `reloadThreadProjections` gates on `listGen > 0`, so if
  `6g5rwjqr7rt8` forgets to call `requestThreadList` on entering the surface, Threads will load
  once and then never live-reload, with no test failing.
- **The unwrapped retry as the bound** (read_retry.go:26) — correct and well tested; re-wrapping it
  kills `TestRepeatedConflictBecomesVisibleWithoutSpinning`.
- **Closure inside a message** — not an anti-pattern here. `tea.BatchMsg []Cmd` and
  `sequenceMsg []Cmd` (bubbletea v2.0.9 commands.go:21,30) are upstream precedent, and `session.go`
  already hand-unwraps `BatchMsg`. Re-deriving the read from `readRequest` instead would need
  non-bumping variants of five loaders plus a second five-way switch, for identical behaviour,
  because the criterion requires the retry to reuse the same generation. `scopeSession` handles the
  closure correctly: the conflict message is stamped, and the retry command returned from `Update`
  is re-wrapped at the then-current `sessionGen` (model.go:303). Costs worth knowing: the closure
  captures `*core.Service` across the 200ms window (safe only via `sessionGen` — see M8), and the
  structs are non-comparable, which is why tests compare `.request`.
- **`TestThreadReadsSurviveSplitWorkspaceCapabilities`** — a real AC2 proof; the fakes serve data
  absent from disk, so any fallback to the aggregate store would surface as the fixture's values.
- **Thread cache generation stamps** — solid; three separate mutations killed.
- **`return m, m.reloadDashboard()`** (value receiver plus pointer-receiver mutation in one
  `return`) — gc evaluates the call first, so the bump reaches the returned model; verified
  standalone, and the codebase already relies on this pervasively.

## Candidate tasks

- ⏳ `tskflwctl task new "Pin the TUI read-contention policy's unproven guards" --epic 18-tui-bubble-tea-interactive-planning-browser --tags tui,tests` — H1, H2, M1, M2, L1, L2: fix the helper hang, then pin the retained detail, the dashboard stamp, and `readRequestCurrent` for all five surfaces.
- ⏳ `tskflwctl task new "Close the shared TUI read policy's remaining gaps" --epic 18-tui-bubble-tea-interactive-planning-browser --tags tui,architecture` — M6, M7, M8, L3: atlas participation, a `readResult` opt-in interface, and the `sessionScope` invariant.
- ⏳ H3 — small enough to fold into the Thread list/detail task `6g5rwjqr7rt8`, whose renderer is the first consumer that can be misled by the aliasing.
- ⏳ M3 — record the three acceptance-criteria narrowings as deferrals on the completed task, and carry the cursor/selection clause into `6g5rwjqr7rt8`.
- ⏳ M4, M5 — no new task; state in `6g5rwjqr7rt8`'s scope that it *deletes* `threadProjectionState` and folds `readSurface` into `entityKind`.
