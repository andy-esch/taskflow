---
schema: 1
area: consumer-data-flow-architecture
date: "2026-06-27"
---
# Architectural audit — multi-consumer data flow (CLI · TUI · future Web)

Scope: how data flows from `core.Service` out to its consumers — the CLI (`internal/cli` + `internal/cli/render`), the TUI (`internal/tui`), and the planned **web** adapter (`tskflwctl serve`, epic 19) — and where there are hacks, anti-patterns, code smells, inconsistencies, and non-idiomatic Go. Goal: long-term maintainability and robust, idiomatic Go before a third consumer arrives to triple the existing patterns.

Method: four independent reviews (read-path/presentation duplication; write-path/core-boundary; serialization/web-readiness; Go-idioms/robustness). Findings are deduplicated and grounded in `file:line`.

Overall health is strong — the core/store boundary is clean, writes are atomic + compare-and-swap, async loads are generation-guarded, errors wrap domain sentinels with `%w`, and the CLI exit-code mapping is a single `errors.Is` table. The debt concentrates **one level up**: the presentation *composites* and the JSON *wire contract* have no shared seam, so a web adapter would re-implement several hand-rolled patterns a third time; and a few cross-cutting concerns (the verb→state vocabulary, error classification, config/maintenance, request context) are trapped inside `package cli` or duplicated into `package tui`. See "What is solid" at the end for what was checked and deliberately excluded.

## High

#### H1. The JSON wire contract is trapped inside a sibling primary adapter  · **Status:** fixed
**Component:** cli/render · **Effort:** L · **Urgency:** soon

The entire machine wire format — envelopes, DTOs, `SchemaVersion` (`render.go:67`), `JSONSchema()` (`envelopes.go:257`), the embedded `schema_comments.json` — lives in `internal/cli/render`, the CLI's presentation adapter, and the DTO types/mappers (`taskJSON`, `auditToJSON`, `toFindingsRollup`) are **unexported** (`dto.go:13,81,126`). A future `tskflwctl serve` that wants the same JSON has only bad options: import `internal/cli/render` from the web adapter (a primary adapter importing a sibling primary adapter — the textbook hexagonal layering violation), or re-declare every envelope and drift from `SchemaVersion`/the goldens. The emit funcs are also `io.Writer`-shaped and version-stamping (`SummaryJSON(w, …)`), so an HTTP handler can't get a *value* to wrap in its own response without re-encoding through a buffer. The package doc claims render "is the only place that knows about presentation" — but a machine wire contract is an API, not presentation. Fix: extract the wire contract to a neutral leaf package (`internal/wire` / `internal/api`) depending only on `core`/`domain` — exported envelopes, DTO mappers, `SchemaVersion`, `JSONSchema()` — leaving the human renderers (`*Human`, `Style`, lipgloss) in `render`. Both `cli` and `web` then import `wire`.

#### H2. `core.Service.Summary()` re-reads every audit file from disk a second time  · **Status:** fixed
**Component:** core · **Effort:** M · **Urgency:** soon

`Summary()` calls `ListAudits` (which `os.ReadFile`s + `domain.ParseFindings` every audit body to compute the tally bands — `store/auditstore.go:180-204`) and THEN calls `s.QueryFindings(...)` (`service.go:150`), which `GetAuditByPath`s + re-parses every one of those same bodies again (`finding.go:170`). That's `2 × (#audits)` disk reads where one would do; the bodies `ListAudits` already parsed are discarded by the store before `Summary` sees them. The inline comment (`service.go:145-149`) is honest about a "second pass" but understates it as "the same parse the TUI already runs" — it is a second `os.ReadFile`, on the single hottest read path (`status`, the TUI dashboard re-run on every fsnotify reload, and the future web dashboard handler). This was introduced with the findings-rollup feature and should be tracked debt, not a resting state. Fix: have the store expose audit bodies alongside `ListAudits` (or cache parsed `[]domain.Finding` on `domain.Audit`) so the rollup reuses already-read data; or a single combined-pass `core` helper returning both the tally and the rollup.

#### H3. The lifecycle verb→destination vocabulary is duplicated per adapter with no shared registry  · **Status:** fixed
**Component:** core / cli / tui · **Effort:** M · **Urgency:** soon

"Verb name ↔ destination state" — the central lifecycle vocabulary — is written out twice with nothing tying the copies together. The CLI declares it inline as constructor args (`cli/task.go:54-66` `newTransitionCmd(…,domain.StatusInProgress)`, `cli/audit.go:22-24`); the TUI declares the same mappings as separate tables (`tui/action.go:27-34` `taskTransitions`, `:39-43`, `:50-54`). The `destructive` flag and verb labels exist *only* in the TUI table; the CLI has no equivalent. `core` already centralizes `AllStatuses()`/`AllAuditBuckets()`/`AllEpicStatuses()` — the verb mapping is the missing peer, and a web adapter needs exactly this table to render action buttons / map `POST /tasks/{slug}/transitions/{verb}`, becoming a third hand-written copy. Fix: promote one `[]Transition{Verb, To, Destructive}` per entity into `core`/`domain`; the CLI builds its command loop from it, the TUI tables become thin views of it, web reads it for routing.

#### H4. Error classification is centralized for the CLI but reinvented as raw text in the TUI  · **Status:** fixed
**Component:** core / tui · **Effort:** M · **Urgency:** soon

The CLI has one good sentinel→outcome mapper: `cli/exit.go:31-44` maps `ErrNotFound/ErrValidation/ErrAmbiguous/ErrConflict` to exit codes + stable JSON `code` names via `errors.Is`. The TUI has **no** `errors.Is` classification at all — it flashes `msg.err.Error()` verbatim (`tui/model.go:235`) and reconstructs the inline-edit reason by string-stripping the sentinel prefix: `strings.TrimPrefix(msg.err.Error(), domain.ErrValidation.Error()+": ")` (`model.go:232`). That hack couples the TUI to the exact wrapping format and breaks the moment the message doesn't start with `"validation failed: "`. More importantly, the TUI cannot distinguish `ErrConflict` (the CAS "changed on disk during move", `store/fsstore.go:173,239` — which should prompt a *reload*) from `ErrValidation` (fix-in-place) from `ErrNotFound`; it paints them all red. The good mapper is locked in `package cli` where neither TUI nor web can import it without pulling cobra. Fix: move an `ErrorClass(err) → {NotFound, Validation, Ambiguous, Conflict, Unknown}` (over `errors.Is`) into `core`/`domain`; CLI maps class→exit code, web→HTTP status (404/409/422), TUI→{red flash, reload-and-retry, inline field error}. Delete the `TrimPrefix`.

#### H5. The "bar + percent + done/total" progress composite is hand-assembled 7-10× with drifting formats  · **Status:** fixed
**Component:** theme / render / tui · **Effort:** M · **Urgency:** soon

The shared seams (`internal/theme`, `internal/progressbar`) cover the low-level primitives (glyph+color tokens, the bar render, `RelativeDate`) but stop *below* the composite line "`<bar> <pct%> <done>/<total>`", which is rebuilt by `fmt.Sprintf` at 7-10 sites that disagree: percent is `st.Percent()`→`%d%%` in the CLI (`style.go:130`), `%3d%%` in TUI rows/dashboard (`item.go:139,183`, `dashboard.go:128`), `%d%%` in TUI detail (`detail.go:450,483`) — so the same epic shows `7%` here and `  7%` there; done/total is inline `%d/%d` in the CLI but funnels through the column-measured `rollupCounts` in the TUI (`item.go:33`); bar width is a bare magic number differing by site (`10/8/10/8/8/10` CLI, `8/8/8/12/12` TUI). It is the single most-repeated read→render pattern and has no seam — three surfaces today, a fourth with web. Fix: one shared rollup-composite formatter beside the primitives (e.g. `theme.Rollup`/`progressbar.Progress(done,total,width)` owning the `%d%%` vs `%3d%%` and done/total justification); surfaces keep only color application (ANSI vs lipgloss).

**Resolution (2026-06-28, fixed):** on inspection the differences were *small* — the
bar (`progressbar.Render`) and the percent *color* (`theme.Percent`) were already
shared; only the number *formats* drifted. Unified those into one place
(`theme.PercentLabel` / `PercentLabelPadded` / `Counts`, `internal/theme/progress.go`),
consumed at all ~10 sites; each surface keeps its own color (ANSI Style vs lipgloss)
and the assembly is left per-context (a CLI table cell vs a `progress:` field vs a TUI
list row — genuinely different layouts, not drift). Output is byte-identical (goldens
unchanged), so non-breaking. Bar *widths* (8/10/12) are a deliberate per-context choice
(roomier in a detail field than a status row), not drift, and intentionally left.

#### H6. No `context.Context` on any core or store method  · **Status:** deferred
**Component:** core / store · **Effort:** L · **Urgency:** soon

No method on the `Store` ports, the ~27 `*Service` methods, or the `*FS` impls takes a `context.Context` (grep returns nothing). Fine for a CLI/TUI; a real liability for a web adapter, where an HTTP handler has no way to propagate request cancellation, deadlines, or tracing into a read/write — every call is unconditionally synchronous and uncancellable. The retrofit is broad but mechanical (thread `ctx` through the ports + Service + FS, gating `ReadFile`/`ReadDir` loops via `ctx.Err()` between files); the cost grows with every new call site, so doing it *before* web exists keeps the diff additive rather than a flag-day rewrite. Fix: add `ctx context.Context` as the first parameter to the port interfaces and `Service` methods now, even if the CLI passes `context.Background()` and the FS only checks `ctx.Err()` at loop boundaries.

**Resolution (2026-06-28, deferred):** this is pure web-readiness — the only non-web
upside today is a Ctrl-C-cancels-a-scan nicety, and the retrofit is additive +
mechanical, so deferring is cheap and waiting lets the real `serve` handler shape the
seam rather than guessing at it now. Tracked as epic-21 task
`thread-context.context-through-the-core-and-store-ports` (tags: architecture, web),
**deferred** pending the web effort (epic 19, `tskflwctl serve`); revisit when that
adapter is actually scoped. Deferred, not dropped: the diff grows only slowly with new
call sites, so the cost of waiting stays low.

#### H7. `init` / `doctor` / `lint --fix` + repo discovery bypass `core.Service` and have no reusable seam  · **Status:** deferred
**Component:** cli / config · **Effort:** L · **Urgency:** eventually

Real maintenance use cases live in the CLI adapter and `config`, not behind a port: `init` calls `config.Init/AddTrackedRepo/InitPointer/LinkBack` directly (`cli/init.go:103,109,152,159`); `doctor` calls `config.CheckLinks` (`cli/doctor.go`, duplicated in `cli/root.go:170` `warnLinks`); `lint --fix` calls `app.Fixer.FixFrontmatter` (`cli/lint.go:59`) over a port wired straight to the FS (`root.go:156`). Repo discovery + service construction also live CLI-side in `App.resolve()` (`root.go:142-159`), and the TUI is handed a built `Svc`+`Layout`, never the `*config.Config`. So a web adapter can reuse none of init/doctor/fix/discovery — it must re-import `config`, re-wire the `Fixer` port, and replicate `resolve()`. `doctor`'s double-sourcing is proof the logic wants a single home. Fix: lift "discover config → build store → build service" into a small reusable `Resolve(startDir) → Workspace{Cfg,Svc,Layout,Fixer}` both adapters call; promote `Doctor()`/`FixFrontmatter()` to `core.Service`; keep `init` (the repo-less case) as a cobra-free `config` function web can call.

**Resolution (2026-06-28, deferred):** the present-day win — collapsing `doctor`'s
double-sourced linkback check (`cli/doctor.go` vs `cli/root.go` `warnLinks`) — is
small; the bulk (a reusable `Resolve()`→`Workspace` seam + promoting `Doctor()`/
`FixFrontmatter()` to `core.Service`) only earns its keep once a second adapter exists
to reuse it. Tracked as epic-21 task
`reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli` (tags:
architecture, web), **deferred** pending the web effort (epic 19); revisit when
`tskflwctl serve` is scoped.

## Medium

#### M1. The dashboard `setSummary` hand-rolls aligned columns that `writeTable` already provides  · **Status:** fixed
**Component:** tui · **Effort:** M · **Urgency:** eventually

`tui/dashboard.go` `setSummary` (~145 lines) builds aligned multi-column rows (glyph·date·slug; bar·pct·counts·date·id) by pre-measuring the widest cell (`dateW`/`countsW`) and padding with `%-*s` by hand (`:78-91,118-134`), then concatenating into one `dashRow.text`. The CLI renders the same in-progress/epic widgets through `writeTable` (`render.go:257-272`), which does measurement/alignment/truncation generically; the same `countsW` pre-measure recurs in the TUI list loaders (`commands.go:92-95,134-137`). Column-alignment logic thus exists as a reusable writer on one side and open-coded per-widget on the other — exactly what drifts (see H5). Fix: factor the dashboard's rows into structured cells (or a shared column/alignment helper) so the layout is described once and rendered per-surface.

**Resolution (2026-06-28, fixed):** the per-widget measure-then-pad logic is now two
shared generic helpers in `internal/tui/column.go` — `relDateCells` (the aligned,
dimmed relative-date column) and `countsWidth` (the done/total column width). The
dashboard's in-progress + epics widgets AND the epic/audit list loaders (`commands.go`)
all call them instead of hand-rolling `dateW`/`countsW` + `%-*s`, so the alignment is
described once. Byte-identical render (dashboard tests unchanged); unit-tested
(`TestRelDateCells`/`TestCountsWidth`). A full writeTable-style cell framework was
deliberately not built — the dashboard rows are heterogeneous (fixed glyph/bar/pct
cells, a right-justified counts column, a trailing id), so factoring the two
genuinely-duplicated measured columns is the right scope; layout/separators stay
per-widget.

#### M2. The dashboard's epic ordering diverges from CLI `status` for identical data  · **Status:** fixed
**Component:** core / cli / tui · **Effort:** S · **Urgency:** soon

`core.Summary.Epics` arrives in store order. The TUI dashboard re-sorts it most-recently-touched-first (`dashboard.go:115` `epicsByRecent`, using `EpicSummary.LastUpdated`); CLI `status` renders the raw order (`render.go:264-272`). The stated contract is "the CLI `status` and the in-app dashboard show the same thing" (`dashboard.go:16-19`, `render.go` "so the two surfaces agree"), and `LastUpdated` was added to the core aggregate specifically for this lens — but only one of the two dashboards uses it. A web dashboard would have to guess which ordering is canonical. Fix: decide where recency ordering lives — either `core.Summary` returns `Epics` in dashboard order (entity tabs re-sort), or `epicsByRecent` becomes a shared helper both dashboards call.

**Resolution (2026-06-28, fixed):** recency ordering now lives in the aggregate —
`core.Summary.Epics` is returned most-recently-updated first (`epicsByRecent` in
`internal/core/service_epic.go`, applied once in `Summary()`), so CLI `status` and the
TUI dashboard read ONE order instead of each re-sorting. The TUI's local `epicsByRecent`
is deleted; the entity list / `epic list` keep their own store order (via `rollupEpics`
directly). Pinned by `TestService_Summary_EpicsByRecent`. The CLI `status` golden was
unchanged (its fixture epics were already recency-ordered), so non-breaking. The rest of
this finding's task cluster (M3, M9, L1) remains in epic-21 task
`let-core-own-the-dashboard-aggregates-adapters-re-derive`.

#### M3. Epic rollup percent is re-derived in the show/detail paths instead of `EpicSummary.Percent()`  · **Status:** fixed
**Component:** core / cli / tui · **Effort:** S · **Urgency:** eventually

`EpicSummary.Percent()` (`service_epic.go:89-94`) encapsulates the zero-guarded `Done*100/Total` rule, and the list/status paths use it — but the two show/detail paths bypass it: `render.go:498-503` (`EpicShowHuman`) and `tui/detail.go:444-450` (`renderEpicMeta`) both re-run `core.TaskRollup(tasks)` and re-implement the percent inline, because `ShowEpic` returns raw `(Epic, []Task)` not an `EpicSummary`. The "what counts as done / how percent is computed" domain rule now lives in three places and agrees only by discipline; a web epic-detail copies it a fourth time. Fix: have `ShowEpic` return an `EpicSummary` (or a small rollup struct) so all paths consume `Percent()`/`Done`/`Total` and `TaskRollup` is called once.

**Resolution (2026-06-28, fixed):** `ShowEpic` now returns an `EpicSummary` (its
rollup), assembled in one place — a new `rollupEpic(epic, tasks)` helper that both
`rollupEpics` (list/status) and `ShowEpic` (show/detail) call, so `TaskRollup` runs
once per path and the percent rule isn't re-implemented. `EpicShowHuman` and the TUI
`renderEpicMeta` consume `es.Percent()`/`es.Done`/`es.Total`/`es.Deprecated` instead of
re-deriving inline. Output byte-identical (goldens unchanged); pinned by the updated
`TestService_ShowEpic`.

#### M4. `defer` (move + revisit date) is special-cased structurally in all three layers; `DeferTask` is non-atomic  · **Status:** open
**Component:** core / cli / tui · **Effort:** M · **Urgency:** eventually

"defer is a move that also carries an optional revisit date" is hardcoded per adapter: core's `DeferTask` is Move + SetFields with a documented non-atomic two-write hazard (`service_task.go:116-141`); the CLI forks a bespoke `newDeferCmd` separate from the generic `newTransitionCmd` (`cli/task.go:466-509`); the TUI hardcodes `tr.to == string(domain.StatusDeferred)` interception in two places (`tui/model.go:711,732`). The transition registry (H3) has no way to say "this transition collects an extra parameter," so each adapter branches on the literal `deferred`; web will add a fourth special-case. Fix: model the extra parameter on the transition descriptor (`Param *ParamSpec`) so adapters render "this verb wants a date" generically; longer term let the store record the revisit date *within* the Move write (it already builds the `updates` map at `fsstore.go:122-133`) so `DeferTask` is one atomic write.

**Progress (2026-06-28):** the *atomicity* half is fixed — `DeferTask` is now ONE
atomic write via a new `Store.Defer` port method; `internal/store/fsstore.go`'s shared
`moveTask` records `revisit_at` inside the same relocation write (a re-defer rewrites it
in place), replacing the Move-then-SetFields two-write hazard. The core also validates the
date up front (the guard the old SetFields path gave for free). Pinned by `TestFS_Defer`
+ `TestDeferTask_*`. STILL OPEN (so this finding stays open): the structural special-casing
of `defer` across the three adapters — that rides the shared transition registry's optional
param spec, tracked in epic-21 task
`promote-a-shared-transition-registry-to-core-verb-to-state-destructive-params`.

#### M5. `deprecate` confirmation exists only in the TUI; relative revisit dates use wall-clock, not the injected clock  · **Status:** fixed
**Component:** cli / tui · **Effort:** S · **Urgency:** soon

Two small consistency leaks. (a) "deprecate is destructive" lives only in a TUI struct field (`action.go:33` `destructive:true`, gated by a y/n at `model.go:705-708`); the CLI `deprecate`/`complete` apply with no confirmation and no `--force`/`--yes`. (b) `core.Service` deliberately exposes an injectable clock `svc.Now()` (`service.go:61`) and the TUI uses it for due-ness (`commands.go:43`), but the revisit-date relative-offset parse reaches for wall-clock `time.Now()` at three sites — `cli/task.go:497`, `tui/edit.go:318`, `tui/model.go:661` — so "2w from now" is computed against a *different* clock than core stamps dates with, and a `WithClock` test or future "as-of" mode governs stamps but not offsets. Fix: carry `destructive` on the shared transition descriptor (H3); thread `svc.Now()` into the revisit-date parse at all three sites.

#### M6. `model.go` is a 1537-line god-file mixing reducer, key dispatch, layout, palette, and all rendering  · **Status:** fixed
**Component:** tui · **Effort:** M · **Urgency:** eventually

`tui/model.go` (1537 lines; next-largest source is 803) holds `Model`, `Update`/`update`, `handleKey` (189 lines — the longest TUI function), the palette builders, sort/view cycling, layout math, and the whole `View`/`footer`/`tabStrip`/`pane` render stack. It's coherent and well-commented but is the file every TUI change touches. The seams already exist (sibling concerns are split into `overlay.go`/`nav.go`/`edit.go`/`dashboard.go`/`commands.go`). Fix: mechanically lift the render half (`View`/`renderBody`/`footer`/`tabStrip`/`pane`/`windowTitle`, ~250 lines) into `view.go` and the palette/command-dispatch cluster into `command_dispatch.go`. No behavior change.

**Resolution (2026-06-28, fixed):** split mechanically, no behavior change. `model.go`
dropped 1659 → 1065 lines: the render/layout half (`View`, `renderBody`, `footer`,
`tabStrip`, `pane`, `recomputeLayout`, the detail-pane + help-scroll helpers, the
footer builders) moved verbatim into `view.go` (380 lines), and the `:` command +
`ctrl+p` palette cluster (`dispatchCommand`, the palette builders/handlers, command
completion) into `command_dispatch.go` (235 lines). The reducer (`Update`/`handleKey`),
navigation, selection, and sort/view cycling stay in `model.go`. Same package, so the
cross-file calls are unchanged; the full TUI suite stays green and there's no golden
churn.

#### M7. `ListAudits` doesn't validate an unknown bucket — asymmetric with `ListTasks`, a latent web trap  · **Status:** fixed
**Component:** core · **Effort:** XS · **Urgency:** soon

`ListTasks` validates `f.Status` up front so an unknown status returns `ErrValidation` rather than a silently-empty list that agents routing on exit codes can't distinguish (its own doc, `service_task.go:26-31`). `ListAudits(bucket, all)` (`service_audit.go:65-83`) does no such check — an unrecognized bucket falls through the `switch` and returns an empty slice with `nil` error. Unreachable today (callers pass boolean-derived buckets), but a web handler taking a `?bucket=` query param hits exactly the empty-vs-invalid ambiguity `ListTasks` was hardened against. Fix: mirror `ListTasks` — `if bucket != "" { if _, err := domain.ParseAuditBucket(bucket); err != nil { return nil,nil,err } }`.

#### M8. The `Update` default fall-through routes any untagged async message to the active tab — invariant by comment only  · **Status:** fixed
**Component:** tui · **Effort:** S · **Urgency:** eventually

The reducer's default case forwards any unhandled `tea.Msg` to `m.cur().list` only (`tui/model.go:261-271`), documented as an INVARIANT ("a future background component with its own ticks must be tab-tagged, or this fall-through changed to broadcast"). The whole stale-load/gen architecture's safety rests on a contract enforced only by a comment, and the failure mode (a background tab's tick applied to the active tab) is silent and intermittent. Fix: make the contract executable — type-switch known passthrough messages explicitly and drop (`return m, nil`) on a genuinely unrecognized type, or route by embedded kind; at minimum add a test asserting an untagged list message never reaches a non-active tab.

**Resolution (2026-06-28, fixed):** the invariant is now executable, not comment-only —
`TestModel_UntaggedMsgRoutesToActiveTabOnly` (`internal/tui/model_test.go`) sends an
untagged probe (an empty `list.FilterMatchesMsg`) and asserts only the *active* tab's list
processes it, with background tabs' visible sets untouched. Mutation-tested: the assertion
fails if the fall-through is changed to broadcast. The audit's heavier option —
type-switching the known passthrough types and dropping the rest — was deliberately NOT
taken: enumerating Bubble Tea's internal list/cursor/filter message types is fragile and
would risk swallowing a blink/filter tick the active list legitimately needs, so the
executable guard (plus the documented comment) is the right scope. The separate `model.go`
split (M6) remains in epic-21 task
`maintainability-and-latent-edge-hardening-model.go-split-routing-invariant`.

#### M9. The "ready to close" / settled aggregate is duplicated across adapters over raw `OpenAudits`  · **Status:** fixed
**Component:** core / cli / tui · **Effort:** S · **Urgency:** eventually

`core.Summary` is a grab-bag of raw lists (`InProgress []Task`, `OpenAudits []Audit`) AND pre-baked aggregates (`Epics`, `FindingsRollup`). Because the raw side is exposed, both surfaces re-derive the same rollup off it: `settledCount` / `countSettled` (counting `OpenAudits` with `a.Settled()`) is implemented verbatim twice — `render.go:336-344` and `dashboard.go:233-242` — and the JSON path re-walks `OpenAudits` for per-audit `ready_to_close` (`render.go:376-387`). A summary that mixes "here are the rows, you aggregate" with "here is the aggregate" invites exactly this. Fix: push the settled/ready-to-close count into `Summary`/`FindingsRollup` so all three surfaces read a number; keep raw lists only where a surface genuinely renders per-item.

**Resolution (2026-06-28, fixed):** the settled count is now an aggregate —
`Summary.ReadyToClose`, computed once in `Summary()` from the same audit sweep — and
both the CLI (`render.go`) and TUI (`dashboard.go`) read it; the verbatim
`settledCount`/`countSettled` helpers are deleted. The raw `OpenAudits` list stays (the
surfaces still render per-audit rows), and the JSON per-audit `ready_to_close` walk is
left as-is — a genuinely per-item render, not the duplicated count. Pinned by
`TestService_Summary_ReadyToClose`; goldens unchanged.

#### M10. The `[]CountBy` breakdown line is re-formatted in three adapters (four with web)  · **Status:** fixed
**Component:** cli / tui · **Effort:** XS · **Urgency:** eventually

Core correctly owns the *tally* (`FindingsRollup.ByUrgency`/`ByComponent` as `[]CountBy`, `finding.go:36-62`) — the right boundary — but the "N key · N key" join is re-implemented: `render.go:326` `countByLine` (plain, dim separators) vs `dashboard.go:203,220` `urgencyLine`/`componentLine` (acute red / soon yellow, `+N more` cap). The divergence is legitimately presentational, so don't over-merge; but factor the *structure* (iterate, format each `CountBy`, join, optional cap) into one helper taking a per-segment formatter + separator, so CLI/TUI/web supply only styling. Keep the structured JSON emit (`toFindingsRollup`) as-is.

**Resolution (2026-06-28, fixed):** the iterate/format/join/cap STRUCTURE is now one
generic `theme.Breakdown[T](items, sep, max, seg, more)`; `countByLine` (CLI),
`urgencyLine` and `componentLine` (TUI) supply only their per-segment format, separator,
and optional "+N more" cap — so the styling divergence the audit flagged as legitimate
stays per-surface while the loop isn't re-rolled. The generic signature keeps `theme`
free of a `core` import. Output unchanged (goldens green); the cap behavior is
unit-tested (`theme.TestBreakdown`). The structured JSON emit is untouched.

## Low

#### L1. `FindingsRollup` is a presentation-shaped aggregate living as an intrinsic `Summary` field  · **Status:** deferred
**Component:** core · **Effort:** S · **Urgency:** eventually

`FindingsRollup`'s ordering is display-driven — `ByUrgency` "canonical triage order (acute, soon, eventually)", `ByComponent` "most-findings-first", plus a hand-picked `Acute []AuditFinding` call-out (`finding.go:17-31,53-60`) — and it's a fixed field on `Summary` (`service.go:92`). Defensible (the tally is worth more than purity, and triage order is arguably domain logic), but `Summary` risks becoming "whatever the current dashboards need," and a web findings page wanting pagination/another sort would either reuse this fixed shape or call the composable `QueryFindings` and re-roll. Fix: keep it, but make it a `Service.FindingsRollup()` view-model that `Summary` *composes*, so web can roll up with its own filter — which also naturally enables the single-sweep fix in H2.

**Resolution (2026-06-28, deferred):** defensible as-is — the tally is worth more than
purity, triage order is arguably domain logic, and the H2 single-sweep it would enable
already landed independently. Recasting `FindingsRollup` as a composed
`Service.FindingsRollup()` view-model only earns its keep when a web findings page wants
its own pagination/sort, so it's **deferred** pending the web effort (epic 19). Tracked as
the deferred sub-item of epic-21 task
`let-core-own-the-dashboard-aggregates-adapters-re-derive`; revisit when `tskflwctl serve`
is scoped.

#### L2. Schema field descriptions come from two sources chosen by export-visibility  · **Status:** wontfix
**Component:** cli/render · **Effort:** S · **Urgency:** eventually

Descriptions come from inline `jsonschema:"description=…"` struct tags on the unexported projection DTOs (`dto.go:14-122`) AND the generated Go-doc map `schema_comments.json` for the exported envelopes/domain types — because `AddGoComments` skips unexported types. Two sources of truth for "the description of a schema field," selected by visibility; a known foot-gun guarded by `TestTaskJSONDescriptionTagMatchesCap`/`TestEpicStatusDescriptionMatchesVocab` precisely because a tag literal like `"<=200 chars"` can drift from `domain.MaxDescriptionLen`. Fix: when extracting the wire package (H1), make the envelope DTOs **exported** there so `AddGoComments` reads their doc comments — collapsing to one description source and retiring the drift-guard tests.

**Resolution (2026-06-28, wontfix):** investigated post-H1. The two "sources" are
actually the right mechanism per scope — `jsonschema:"description=…"` tags for field
descriptions (the reflector's intended mechanism, yielding clean machine strings) and
doc comments for *type* descriptions. Verified the **tag deterministically wins** over
a field's doc comment, so a maintainer note never leaks into the schema; there is no
live conflict. Migrating field descriptions to doc comments was tested and *degrades*
the output — invopop does not strip the field name, so a Go-idiomatic `// Slug is the…`
renders as `"Slug is the…"` (verbose, prefixed) vs the tag's clean `"task identifier…"`;
byte-identical output would require *non-idiomatic* comments, defeating the point. The
drift-guard tests guard a real literal-vs-constant risk and are kept. Convention now
documented in `internal/wire/dto.go`.

#### L3. `actionMenu.selected()` / `editMenu.cur()` index without a bounds guard  · **Status:** fixed
**Component:** tui · **Effort:** XS · **Urgency:** eventually

`move`/`optMove` are all guarded `if n := len(...); n > 0`, but `selected()` (`action.go:146` `a.options[a.cursor]`) and `cur()` (`edit.go:192` `e.fields[e.cursor]`) index directly. Safe today only as an emergent property of the data (every entity's transition table survives the no-op filter non-empty; edit forms always have ≥3 fields). A future entity whose only transition equals its current state makes `validTransitions` return `nil`, and the first `enter` panics (`model.go:686,704`). Fix: `open()` refuses to activate on empty options, or `selected()` returns `(transition, bool)`.

#### L4. `setMapNode` unconditionally carries the old value node's comments onto its replacement  · **Status:** fixed
**Component:** store · **Effort:** XS · **Urgency:** eventually

On a surgical update, `setMapNode` copies the old node's `Head/Line/FootComment` onto the new node "so comments survive" (`store/frontmatter.go:186-197`) — correct for the common case, but an unconditional overwrite of the new node's (today empty) comment fields. If a field is set twice in one `updates` pass, or a future caller builds a value node *with* a comment, the carry clobbers it. Fix: only copy when the new node's corresponding field is empty, making "preserve, don't overwrite" explicit.

#### L5. `scrollToCurrent` uses `> 0` where `>= 0` is intended  · **Status:** fixed
**Component:** tui · **Effort:** XS · **Urgency:** eventually

`detail.go:324`: `if target := …line - 2; target > 0 { SetYOffset(target) } else { GotoTop() }`. A match on line 2 yields `target == 0` and takes `GotoTop` — identical result, so harmless, but the boundary should be `>= 0` to match intent. Cosmetic; flagged only because it's the `> 0` vs `>= 0` class the audit scanned for (all other bounds math — the dashboard `scrollTo`, the `((x%n)+n)%n` wraps, every `Percent()` div guard — is correct).

## What is solid (checked, deliberately not findings)

So the fixes stay scoped, these were reviewed and found genuinely sound: the **atomic-write + compare-and-swap + temp cleanup** path (`store/atomic.go`, `fsstore.go:106-248` — staged temp → fsync → rename, CAS re-resolve to defeat lost-update-via-concurrent-move, partial-file cleanup on every error path); the **stale-load generation guards** (`loadGen`/`detailGen`/`dirtyGen` ride per-message, checked on every async landing, race-tested in `model_test.go`); the **fsnotify watcher lifecycle** (shared `*watcher`, no goroutine leak, dead-watcher avoided); **interface design** (three narrow `*Store` ports + split `Fixer`/`Layout`, defined at the consumer; `Column[T]`/the descriptor registry are justified de-duplication, not over-abstraction); **path-traversal** defenses (`validQueryName`, `IsRegular()` symlink rejection, physical `EvalSymlinks` containment); **frontmatter parsing** edge cases (unterminated fences, CRLF, conflict markers, non-mapping rejection, duplicate-key attribution); and the **schema/version machinery** (single `SchemaVersion` with a documented changelog + golden snapshots + drift-guard tests). `go vet ./...` is clean, the suite is green, there are zero panics or TODO/FIXME markers in library code, and errors wrap sentinels with `%w` throughout. The headline debt is structural seams for a third consumer, not correctness.
