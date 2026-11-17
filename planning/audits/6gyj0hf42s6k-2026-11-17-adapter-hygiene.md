---
schema: 1
id: 6gyj0hf42s6k
bucket: open
area: adapter-hygiene
date: "2026-11-17"
updated_at: "2026-11-17"
---

# Audit: adapter-hygiene — 2026-11-17

> Create findings with `audit finding new`; update them with `audit finding`.

## Findings

<!-- Example grammar; let `audit finding new` allocate and render real findings: -->

```
#### H1. <title>  · **Status:** open

**File:** <path:line> | **Component:** <component>
**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>

<what's wrong, why it matters, evidence>

**Recommendation:** <minimum fix>

**Resolution:** <how it was resolved — written by `audit finding --note`, not by hand>
```

#### M1. Core refuses to default epic creation fields, so every adapter replicates them · **Status:** open

**File:** internal/core/service_epic.go:25 | **Component:** core + cli + tui
**Effort:** S · **Urgency:** soon

`prepareTaskCreation` (`internal/core/service_task.go:966-982`) states this repo's rule for
creation defaults, and states *why*:

> Defaults for zero-valued fields live here so EVERY caller — not just the CLI flags —
> produces a valid, lint-clean task; the CLI flag defaults stay as help-text hints.
> (A second adapter calling NewTask without replicating them would otherwise get
> ErrValidation or scaffold a lint-failing file.)

`NewEpic` (`internal/core/service_epic.go:25-47`) does the opposite. It defaults nothing and
validates immediately — `domain.ValidatePriority(p.Priority)` at `:35` and
`domain.ValidateEpicStatus(p.Status)` at `:38` both reject `""` — so the two required
non-title fields must be supplied by whoever calls it. Both current primary adapters
therefore carry their own copy of the same two values:

- `internal/cli/epic.go:252` — `cmd.Flags().StringVar(&p.Status, "status", "active", ...)`
- `internal/cli/epic.go:253` — `cmd.Flags().StringVar(&p.Priority, "priority", "medium", ...)`
- `internal/tui/epic_create.go:69` — `f.inputs[epicCreateStatus].SetValue(domain.EpicStatusActive)`
- `internal/tui/epic_create.go:70` — `f.inputs[epicCreatePriority].SetValue("medium")`
- `internal/tui/epic_create.go:83-84` — `dirty()` compares against both values a *second* time,
  so the TUI alone spells the default twice.

Three of those five are bare string literals where `domain.EpicStatusActive` exists and is used
one line above.

This is not a relitigation of the TUI form's design. Task `6gp1ytfvm652` (completed 2026-10-22)
deliberately accepted "bounded selectors start at active and medium ... matching the CLI", and
`docs/architecture/tui.md` records it as "visible initial form choices matching the CLI, not
invented core defaults". That decision is sound *given* a core that refuses to default. The
finding is about the core side: the asymmetry between `NewTask` and `NewEpic` is what forces the
replication, and it is undocumented — no ADR or type comment explains why epics are the
exception.

**Failing scenario:** verified by probe (see Commands run), in a throwaway space outside this
checkout.

- `tskflwctl task new "Probe Task" --epic 01-probe-epic --tags probe --description probe` with no
  priority/tier/autonomy/effort flags → exit 0, and core fills `effort: Unknown`, `tier: 3`,
  `priority: medium`, `autonomy_level: 3`.
- The exact analog for epics, a caller that does not replicate the defaults —
  `tskflwctl epic new "P4" --description probe --priority ""` → **exit 11**,
  `validation failed: priority must be high|medium|low, got ""`. With `--status ""` instead:
  `invalid epic status "" (valid: active, retired, deprecated)`.

So `svc.NewEpic(NewEpicParams{Title: t, Description: d})` — the shape that succeeds for
`NewTask` — fails for epics. Epic 19 (`tskflwctl serve`, active) is a third adapter over the
same service that will either replicate these literals a third time or ship an ErrValidation on
a request that looks complete.

**Why tests didn't catch it:** `TestService_NewEpic` is a *failure* table (it is the subject of
task `6gppr2j8g1h4`, which notes its cases do not reliably assert the predicate they name).
There is no test asserting what a blank-status/blank-priority `NewEpicParams` *should* do, so
neither behaviour is pinned and the asymmetry with `NewTask` is invisible to the suite.

**Context:** scored Medium/soon rather than higher because no current write is wrong — today's
two adapters do replicate correctly, so no epic is created with a bad status. The urgency comes
from the third adapter in epic 19 and from the literals, not from a live corruption.

**Tightening (adjacent):** replace the three bare `"active"`/`"medium"` literals with
`domain.EpicStatusActive` and a domain-side priority default constant; collapse
`epic_create.go`'s second spelling in `dirty()` to one source.

**Follow-up:** whether the `domain` package should publish the per-field defaults it already
declares in the descriptor registry (`internal/domain/entity.go:62` and `:98` both record
`"medium"` as the priority default, and `schema` publishes it) so core reads the default from
one declaration instead of a literal. That is epic 26 territory (a declared frontmatter schema
that lint, `schema` and the JSON contract all derive from), not this fix.

**Recommendation:** Default Status to domain.EpicStatusActive and Priority to "medium" inside NewEpic when blank — the way prepareTaskCreation already does for the task analog — and reduce the CLI flag defaults and the TUI's staged initial choices to help-text/presentation hints over a core-owned default.

#### L1. Five domain vocabulary accessors hand adapters a mutable alias of the closed set · **Status:** open

**File:** internal/domain/epic.go:69 | **Component:** domain (consumed by cli/tui/core)
**Effort:** XS · **Urgency:** eventually

The domain publishes its closed vocabularies through nine accessors, and they split cleanly into
two groups that disagree about ownership:

Return a fresh slice (caller cannot reach the domain's state):

- `internal/domain/thread.go:31` — `AllThreadStatuses()` → `append([]ThreadStatus(nil), threadStatuses...)`
- `internal/domain/finding_create.go:36,39,42` — `FindingBands()`, `FindingEfforts()`, `FindingUrgencies()` → same shape
- `internal/domain/status.go:38` — `ActiveStatuses()` builds a new slice while filtering
- `internal/domain/finding.go:230` — `FindingStatuses()` builds from a map

Return the package-level slice itself:

- `internal/domain/status.go:32` — `func AllStatuses() []Status { return allStatuses }`
- `internal/domain/audit.go:20` — `func AllAuditBuckets() []AuditBucket { return auditBuckets }`
- `internal/domain/epic.go:69` — `func AllEpicStatuses() []string { return epicStatuses }`
- `internal/domain/transition.go:63,66` — `TaskTransitions()`, `AuditTransitions()`

The second group hands every caller a writable window onto the domain's own closed vocabulary.
`epicStatuses` is declared at `internal/domain/epic.go:66` and is also what `ValidateEpicStatus`
(`:73-78`) iterates to decide legality, so a caller that sorted or overwrote the returned slice
would be editing the validator's table for the rest of the process.

The adapters do receive the raw alias. `internal/tui/epic_create.go:51` returns
`domain.AllEpicStatuses()` straight out of `epicCreateEnumOptions` into
`creationControls.options`; `internal/cli/epic.go:310` returns it straight to Cobra as a flag
completion result; `internal/cli/schema.go:108` stores it as
`render.SchemaContract.EpicStatuses`, which `internal/wire/schema.go:95` publishes as the
`epic_statuses` field of the machine contract.

**No failing scenario on the current tree — this is a smell, not a defect, and is labelled Low
accordingly.** I checked every caller rather than assuming:

- the staged-creation host only reads `options` (`len`, index, `slices.Index`) and never writes it;
- `internal/cli/epic.go:192` appends *from* the alias into a separate slice, which copies elements out;
- every other caller ranges over it or builds a new slice;
- `internal/cli/completion.go:39`'s `sort.Strings(out)` sorts a locally-built slice, not the alias;
- `cobra@v1.10.2`'s `completions.go` contains no `sort` or `slices.Sort` call, so the completion
  slice handed to it at `epic.go:310` is not reordered in place by the framework.

What makes it worth recording is that the drift is *already decided*: the newer accessors
(Thread statuses, finding bands/efforts/urgencies) copy, so the intended idiom exists in the same
package — the five older ones simply predate it. The cost of the fix is five `append` calls on
slices of three to eight elements, evaluated outside any hot path.

Had a caller sorted `AllEpicStatuses()`, the damage would be visible rather than silent: the
declared order is `active, retired, deprecated` (`epic.go:66`) and the sorted order is
`active, deprecated, retired`, and the accessor's own doc comment promises "in declared order"
— a promise `cli/render/schema_render.go:61` and `tui/epic_create.go:51` both render directly.

**Discovery path, recorded so it is auditable:** this came from reading `domain.AllEpicStatuses`
as the correct, domain-owned counter-example to M1's bare `"medium"` literal, not from one of
the five picks.

**Recommendation:** Return a defensive copy from AllStatuses, AllAuditBuckets, AllEpicStatuses, TaskTransitions and AuditTransitions, matching AllThreadStatuses and the Finding* accessors that already do.

#### L2. Adapters re-declare two domain vocabularies the domain owns but does not publish usably · **Status:** open

**File:** internal/tui/edit.go:95 | **Component:** cli/render + tui
**Effort:** S · **Urgency:** eventually

Two adapter-local tables restate a vocabulary the domain owns. Both exist for the same reason:
the domain validates the vocabulary but publishes no accessor in the shape the adapter needs.

**1. `internal/tui/edit.go:95-104` — the priority and tier/autonomy choice sets.**

```go
func editEnumOptions(field string) []string {
	switch field {
	case "priority":
		return []string{"high", "medium", "low"}
	case "tier", "autonomy_level":
		return []string{"1", "2", "3", "4", "5"}
	...
```

The domain owns both: `validPriorities` (`internal/domain/validate.go:17`) is the legality table
and `ValidateTier`/`ValidateAutonomy` (`:23-37`) own the 1–5 range. Neither is enumerable from
outside — `validPriorities` is unexported and the tier bound exists only inside a comparison —
so the TUI hardcodes the parsed form. The same literals reappear as prose in
`internal/domain/entity.go:62` and `:98` ("One of: high | medium | low."), which is what
`schema task` publishes.

The asymmetry is visible inside one function: `epicCreateEnumOptions`
(`internal/tui/epic_create.go:48-56`) gets its status choices from `domain.AllEpicStatuses()` —
domain-owned, correct — and its priority choices from `editEnumOptions("priority")` — the local
literal. Same function, same kind of closed vocabulary, two owners.

**2. `internal/cli/render/render.go:823` — the finding-status lifecycle order.**

```go
var findingStatusOrder = []string{"open", "in-progress", "fixed", "tracked", "deferred", "superseded", "wontfix"}
```

This is an exact copy of `domain.findingStatusSpecs` (`internal/domain/finding.go:199-207`),
which declares the vocabulary *in that order* with its glyphs. The domain does export an
accessor — but `FindingStatuses()` (`:230-237`) builds from a map and `sort.Strings` the result,
explicitly "sorted (for help/schema)". So the only published form has the declared lifecycle
order removed, and the adapter that needs that order has no choice but to re-declare it. The
ordering decision itself is domain-owned and carries a documented rationale on the adapter side
("active work first, terminal states last", `render.go:819-822`) — rationale that lives nowhere
near the table it describes.

**No failing scenario on the current tree — labelled Low as a smell, not a defect.** Both copies
agree with the domain today; I diffed them element by element. The drift is latent and
one-directional: adding a status to `findingStatusSpecs` makes `FindingStatusGlyph`,
`ValidFindingStatus`, `audit lint` and `schema` all accept it, while `audit show` sorts it into
the "outside the vocabulary or missing entirely" position — a legal status rendered as a
malformed one. Adding a priority or widening the tier range leaves the TUI's pickers offering
the old set while the CLI's `--priority` flag accepts the new one.

**Relationship to existing work (not a re-find):** task `6gcqz5aqt2sg` already owns the
neighbouring half of this — its scope names "Pin the TUI's hand-typed epic/audit view and epic
transition tables to the domain vocabularies" and its AC 7 names
`domain.AllEpicStatuses`/`AllAuditBuckets` and the `EpicStatus` constants. That covers
`statusview.go`'s view tables and `action.go:63`'s `epicTransitions`, which is why neither is
reported here. The two sites above are outside it, and differ in kind: they cannot be "pinned to
the domain vocabulary" at all until the domain publishes an accessor to pin them to. That is the
actual work this finding proposes, and it is a prerequisite shape change rather than a coverage
test — so the finding stays open and cross-referenced rather than folded in.

**Recommendation:** Publish the two missing shapes from the domain — an ordered priority/tier choice set beside ValidatePriority/ValidateTier, and a declared-order finding-status accessor beside the alphabetical FindingStatuses() — and have the two adapter tables consume them.

#### L3. The TUI's post-create landing hardcodes the status core assigns instead of reading the receipt · **Status:** open

**File:** internal/tui/task_create_commands.go:156 | **Component:** tui
**Effort:** XS · **Urgency:** eventually

`landCreatedTask` decides which status view can keep a newly created task selected:

```go
func (m *Model) landCreatedTask(receipt core.TaskCreationReceipt) tea.Cmd {
	i := indexOfKind(m.tabs, entityTasks)
	view := m.tabs[i].statusView
	if view != "all" && view != string(domain.StatusReadyToStart) {
		view = ""
	}
	...
```

The adapter is restating a decision core owns. `prepareTaskCreation`
(`internal/core/service_task.go:1002-1006`) picks the initial status — `StatusReadyToStart`,
or `StatusNextUp` when `NewTaskParams.Next`, or `StatusInProgress` via the lifecycle plan when
`NewTaskParams.Start` (`:944-952`). The receipt handed to this function already carries the
answer as `receipt.Task.Status`, and the function already dereferences `receipt.Task` on the
next line for the ID and slug.

**No failing scenario on the current tree — a latent trap, labelled Low.** `submitTaskCreate`
(`internal/tui/task_create_commands.go:85-87`) builds `NewTaskParams` with neither `Next` nor
`Start` set, so every TUI-created task really is `ready-to-start` and the literal is correct
today. The three reachable cases all behave:

- `statusView == "all"` → kept; the new row is in view.
- `statusView == "ready-to-start"` → kept; the new row is in view.
- anything else → `""`, the "working" default (`internal/tui/statusview.go:18`), which spans
  active plus deferred and therefore includes `ready-to-start`.

What makes it worth a line is that core already supports the two creation modes the TUI lacks,
and `docs/architecture/tui.md` frames the remaining creation work as adding per-kind intent. The
first time a "create as next-up" affordance is added, this comparison silently stops matching:
with `statusView == "ready-to-start"` the view is *kept* while the created task lands in
`next-up`, so the row is filtered out and `tab.restore` (set by `landCreatedEntity`,
`internal/tui/staged_creation.go:286-287`) has nothing to select — a successful creation that
appears to have done nothing. Reading `receipt.Task.Status` costs nothing and fails correctly
for all three modes.

Recorded as adjacent to M1: same class of seam error — an adapter holding a copy of a decision
the core layer already made and already published — differing only in that here the published
value is sitting in the struct the function receives.

**Recommendation:** Compare against receipt.Task.Status, which the receipt already carries, instead of the literal domain.StatusReadyToStart.

Routine: `code-quality-audit` · lens `adapter-hygiene` · ISO week `2026-W47`,
slot `Tue` (index 4).

## Punch list

- M1. Core refuses to default epic creation fields, so every adapter replicates them  (effort: S · urgency: soon)
- L1. Five domain vocabulary accessors hand adapters a mutable alias of the closed set  (effort: XS · urgency: eventually)
- L2. Adapters re-declare two domain vocabularies the domain owns but does not publish usably  (effort: S · urgency: eventually)
- L3. The TUI's post-create landing hardcodes the status core assigns instead of reading the receipt  (effort: XS · urgency: eventually)

No Critical and no High. One Medium: a core-side ownership gap that two adapters
currently paper over correctly, with a third adapter (epic 19) queued behind it.
The three Lows are one coherent class — a decision the domain or core owns,
re-spelled in an adapter because the owner publishes no usable accessor — which is
why they are kept rather than dropped as churn.

## Files audited

- **Signal**: `internal/tui/staged_creation.go` — score 9.85 (4 commits in 30 days ×
  452 referencing-file sum). The shared staging mechanics every creation form routes
  through; highest-scoring production file in the lens surface.
- **Signal**: `internal/tui/task_create.go` — score 9.55 (4 commits × 377). The
  reference creation controller the other four kinds were extracted against.
- **Adjacency** (from `task_create.go`): `internal/tui/task_create_commands.go` — the
  across-the-seam hop. `docs/architecture/tui.md` names this file as owner of "the
  selected-parent request, typed receipt and task-list landing", so it is where the
  adapter meets `core.Service` and where a re-derived core decision would live. L3
  came from here.
- **Adjacency** (from `staged_creation.go`): `internal/tui/creation_host.go` — the file
  that references it most heavily (score 8.52), owning active-kind lifetime, attempt
  generations and result routing.
- **Random**: `internal/core/service_epic.go` — passed the quality filter (454 lines,
  not generated, not a fixture, not a test). M1 came from here.

Read as supporting context, not as picks: `core/service_task.go` (the
`prepareTaskCreation` defaulting rule), `cli/epic.go`, `tui/epic_create.go`,
`tui/audit_create.go`, `tui/research_create.go`, `tui/edit.go`, `tui/action.go`,
`tui/statusview.go`, `tui/watch.go`, `tui/atlas.go`, `cli/completion.go`,
`cli/schema.go`, `cli/render/render.go`, `cli/render/schema_render.go`,
`wire/schema.go`, `domain/epic.go`, `domain/status.go`, `domain/validate.go`,
`domain/finding.go`, `domain/finding_create.go`, `domain/transition.go`,
`domain/entity.go`, plus `cobra@v1.10.2/completions.go` from the module cache.

Churn in this surface is concentrated almost entirely in `internal/tui` — the
staged-creation wave (`d3b53ff`, `68b29a1`, `7300ef1`, `f3979b0`, `0962cfb`,
`02608ca`) landed five entity creation forms in the last month. The two signal picks
and both adjacency hops therefore sit inside that wave by construction, which is the
churn signal working as intended rather than a sampling accident.

## Commands run

- `just build` — **unavailable**: `just: command not found` in this container, so the
  spec's documented fallback was used: `go build -o bin/tskflwctl ./cmd/tskflwctl` — OK
  (`tskflwctl version v0.0.0-20261115121649-1268c727d349`).
- `./bin/tskflwctl -C . audit list --json` — 7 audits in the `open` bucket, below the
  backpressure threshold of 10, so this is an authoring run.
- `TZ=America/New_York date "+%A %Y-%m-%d %V"` — `Tuesday 2026-11-17 week=47`;
  index = (47 × 2 + 0) mod 6 = 4 → `adapter-hygiene`.
- `git log --since="30 days ago" --name-only` over `internal/cli internal/tui
  internal/wire internal/core/service*.go` — churn ranking (top production file
  `internal/tui/view.go`, 5 commits).
- Blast-radius sweep: per-file declared identifiers × referencing production files,
  for the seven top-churn candidates.
- `grep -rn "core\.Role\|core\.Gate\|core\.Status\|core\.Liveness" internal/cli
  internal/tui internal/wire` — 18 production hits. All are either legitimate display
  of a core decision or already owned by the 2026-10-27 audit's L1 and task
  `6gppr2kfa1dz`; nothing new reported.
- `grep -rn "os\.Stdout\|os\.Stderr\|fmt\.Print" internal/cli internal/tui` — 4 hits,
  all the already-triaged `HasDarkBackground` terminal-capability probes.
- `grep -rn "os\.ReadFile\|os\.WriteFile\|os\.Stat\|exec\.\|LookPath" internal/tui` —
  traced every hit to its call site (see What audited clean).
- **Mutation probe**, in a throwaway planning space outside this checkout
  (`tskflwctl init --no-register` in a scratch directory, one epic, one task). This is
  how M1 was verified rather than inferred:
  - `epic new "Probe Epic" --description probe` → `created epics/01-probe-epic.md`,
    frontmatter `status: active`, `priority: medium` (the CLI's flag defaults).
  - `epic new "P4" --description probe --priority ""` → **exit 11**,
    `error: validation failed: priority must be high|medium|low, got ""`.
  - `epic new "Probe Three" --description probe --status ""` → `error: validation
    failed: invalid epic status "" (valid: active, retired, deprecated)`.
  - `task new "Probe Task" --epic 01-probe-epic --tags probe --description probe` with
    no priority/tier/autonomy/effort → **exit 0**, frontmatter `effort: Unknown`,
    `tier: 3`, `priority: medium`, `autonomy_level: 3` — core's defaults.
- `grep -n "sort\.\|slices\.Sort" $(go env GOMODCACHE)/github.com/spf13/cobra@v1.10.2/completions.go`
  — no matches, which is what rules out a live mutator for L1 via the completion path.
- `./bin/tskflwctl -C . lint` — green (see Validation).
- `./bin/tskflwctl -C . audit lint 2026-11-17-adapter-hygiene` — green.
- `git status --porcelain` — only `planning/` paths, before and after.
- `go test -race ./...` — see Validation for the two pre-existing environmental
  failures, neither caused by this run.
- `golangci-lint run ./...` — **could not run** (toolchain gate, see Validation).

## What audited clean

- `internal/tui/staged_creation.go` (signal pick) — clean, and a good model for the
  seam. It is mechanics only: focus, enum staging, dirty confirmation, widget routing.
  `routeCreationWidget` (`:124-139`) captures routing tokens and never the form or its
  `*core.Service`, recursing into `tea.BatchMsg` so a batched child cannot escape the
  generation guard. `updateCreationInput` (`:109-116`) detaches the Bubbles rune buffer
  through the public `SetValue` API with the documented reason. `creationKnownRefusal`
  (`:308-311`) tests the domain sentinels with `errors.Is` rather than matching error
  text. The one place it could have interpreted a domain value — `enumValue` — is
  explicitly delegated to the per-kind controller instead (`:150-153`).
- `internal/tui/task_create.go` (signal pick) — clean apart from the landing literal
  reported as L3, which actually lives in its command sibling. Notably it does *not*
  attribute fields from error text: `handleTaskCreated` ends at
  `f.err = domain.Reason(msg.err)` with the comment "no field attribution guessed from
  error strings". The free-text `Effort` row initially read as a closed vocabulary
  rendered as free input, but `internal/domain/entity.go:60` declares effort as
  "Rough size estimate (free-form)" with default `"Unknown"` — so text is the correct
  widget and the enum/text split across the four optional rows matches the domain.
- `internal/tui/creation_host.go` (adjacency pick) — clean. Concrete value slots rather
  than a stored draft interface, with the aliasing hazard named in the type comment;
  `begin` (`:142-154`) refuses a second active kind and bumps the generation; `applyResult`
  (`:211-218`) matches kind, attempt *and* origin session before dispatching. It routes
  and never interprets a receipt.
- `internal/tui/task_create_commands.go` (adjacency pick) — clean on the question that
  matters most here. Every core call is inside a `tea.Cmd` closure over copied inputs
  and a captured origin `*core.Service` (`:43-46`, `:105-108`); the uncommitted/committed/
  uncertain classification at `:130-151` fails closed three separate ways, including a
  success with no durable confirmation. The string→int conversion for tier/autonomy
  (`:92-101`) is adapter-side input conversion that `docs/architecture/tui.md` explicitly
  blesses, and core re-validates the result via `domain.ValidateTier`/`ValidateAutonomy`,
  so a wrong conversion fails closed. (Its `strconv.Atoi` error branch does appear
  unreachable — tier and autonomy are enum-only rows that `taskCreateInput` excludes
  from text editing and from paste routing, so only `"1"`–`"5"` or `""` can reach it.
  That is a dead-branch observation for the simplification lens, not an adapter-hygiene
  defect, and it fails safe; not reported.)
- `internal/core/service_epic.go` (random pick) — clean apart from M1's defaulting gap.
  No filesystem, no cobra, no path handling; every write goes through the store port;
  validation delegates to `domain`. Its derived values are owned deliberately and say
  so: `rollupEpic` (`:190`) is "the single place an EpicSummary is assembled ... so the
  done/total/percent rule can't drift", `epicsByRecent` (`:205`) notes the order "lives
  here, in the aggregate, rather than being re-derived per adapter", and `dashboardEpics`
  (`:221-231`) fails *open* on an unknown status rather than hiding the record. This is
  the shape the lens wants and the counter-example the three Lows are measured against.
- The TUI watcher — clean, and checked specifically because `os.Stat` in a primary
  adapter is exactly what this lens looks for. `watcher.reconcile`
  (`internal/tui/watch.go:90`) does a `stat` sweep under a mutex, but all four call
  sites defer it correctly: `waitForFS` (`:336`) and `reconcileWatcher` (`:355`) wrap it
  in a `tea.Cmd` closure, `debounceTick` (`:362`) in a `tea.Tick` callback, and
  `newWatcherWithPathAdder` (`:74`) runs at construction — itself reached only from
  `tui.go:47` (startup) and inside the `openWorkspace` command closure
  (`internal/tui/atlas.go:177-184`). Nothing reaches it from `Update` or `View`.
- The other three creation forms — clean on their declared contracts.
  `audit_create.go:44` and `research_create.go:50` place the date default in the
  *placeholder* ("blank uses today at creation") and pass a blank date through to core
  rather than freezing a value at form-open time, which is what
  `docs/architecture/tui.md` promises; `NewAudit`/`NewResearch` keep date defaulting,
  identity minting and bucket policy.
- Adapter persistence construction and `core` purity — nothing new. No adapter imports
  `store` or constructs persistence; the only `config` imports in `cli` remain the two
  named exceptions in the boundaries guide.
- Already-known, deliberately not re-found: the clipboard `exec.LookPath` probe running
  in `Update` (2026-10-27 M2, still open), the four-site in-flight partition recompute
  (2026-10-27 L1), the archived-task filter negation (tracked by `6gcqz5aqt2sg`), and
  `action.go:63`'s hand-typed `epicTransitions` table (already in `6gcqz5aqt2sg`'s
  scope and AC 7). Verified each against the open audits before writing.

## External research

Done because this run surfaced 1 Medium-or-higher finding, below the step-11 threshold
of 3. Three WebSearch queries (one re-run in extended mode after the first came back
off-target). **WebFetch was not used**, so the notes below rest on search-result
summaries rather than on pages read end to end — corroborating direction, not citations
verified against primary text.

- **Returning a package-level slice from an exported accessor is a recognized Go
  encapsulation bug** — corroborates L1 as an established antipattern rather than a
  local preference. The returned slice shares its backing array with the package
  variable, so a caller's write reaches the original; the recommended fixes are an
  accessor that returns a defensive copy and, specifically, replacing a package-level
  slice or map with a function that returns a fresh value per call. Three-index slicing
  (`s[i:j:k]`) is only a partial mitigation — it bounds `append` but one append that
  fits still writes through. Tailscale's `types/views` package is cited precedent, with
  `AsSlice` returning a copy. The Java CERT rule OBJ05-J ("do not return references to
  private mutable class members") is the closest established formulation. The same
  sources note the real cost — a defensive copy allocates — which is why L1 stays Low
  and XS: these are three- to eight-element slices outside any hot path. Sources:
  <https://dev.to/bala_paranj_059d338e44e7e/your-go-structs-are-leaking-6-encapsulation-fixes-from-a-security-cli-4mbl>,
  <https://algomaster.io/learn/go/slice-gotchas>,
  <https://pkg.go.dev/tailscale.com@v1.66.1/types/views>,
  <https://securecoding.cert.org/confluence/display/java/OBJ05-J.+Do+not+return+references+to+private+mutable+class+members>,
  <https://github.com/golang/go/issues/75620>.
- **"Defaults belong in one place, below the flag parser" is the general Go
  recommendation** — corroborates M1's direction. Because Go has no declaration-level
  default parameters, the library has to apply them itself; the recommended shape is a
  constructor or a `FillDefaults`-style step that fills still-zero fields, with the
  defaults documented, so a frontend passes only what the user actually set rather than
  repeating the default values. That is precisely what `prepareTaskCreation` does and
  what `NewEpic` does not. The sources also name the real hazard — zero-value defaulting
  can swallow a *deliberate* zero — which is worth stating for M1 and does not apply to
  it: epic `status` and `priority` are strings whose zero value `""` is already rejected
  as invalid, so there is no meaningful blank to preserve. Sources:
  <https://thelinuxcode.com/how-to-assign-default-values-for-struct-fields-in-go-without-surprises/>,
  <https://codetrail.io/questions/go/default-function-parameter-values-in-go>,
  <https://www.golinuxcloud.com/golang-default-parameter/>.
- **A negative result, recorded honestly:** a targeted search for where defaults belong
  in a ports-and-adapters design ("hexagonal architecture ... CLI flag defaults
  duplication") returned nothing on the actual question — the results discussed folder
  structure and wiring placement and did not address defaults, flags or duplication at
  all. The inference that config-shaped values sit adapter-side while the core receives
  plain values through its ports is mine from general sources, not a finding from them.
  M1 does not rest on it; it rests on the probe and on this repo's own stated rule at
  `internal/core/service_task.go:966-970`. Sources:
  <https://codeartify.substack.com/p/folder-structures>,
  <https://www.netguru.com/blog/hexagonal-architecture-ports-adapters-guide>,
  <https://www.thoughtworks.com/insights/blog/architecture/hexagonal-architecture-explained-practical-example>.

An antipattern this codebase does **not** have, worth recording as a negative result:
across five creation controllers and the shared host, every single core call is issued
from inside a `tea.Cmd` over copied inputs and a captured origin service, and no form
interprets a receipt it does not own. The "do I/O in Update because it is easier"
failure that this lens most often finds in Bubble Tea codebases is absent from the
entire creation wave — the one live instance of it in `internal/tui` is the clipboard
probe the 2026-10-27 run already reported.

## Related-task observations (propose-only)

- Partial overlap, left open: `planning/tasks/6gppr2j8g1h4-refuse-unsetting-required-epic-fields-and-pin-the-epic-unset-path.md`
  (ready-to-start, tier 2) will be editing `TestService_NewEpic`'s failure table —
  exactly where M1's regression belongs — and its AC already says those cases must
  "assert the predicate each case names". It does not own M1's defaulting change, and
  its own Out of scope defers descriptor-declared required fields to epic 26. Whether
  to land M1's test inside that task or alongside the core change is a scoping call for
  a human. Cross-referenced on both sides.
- Partial overlap, left open: `planning/tasks/6gcqz5aqt2sg-make-closed-vocabulary-switches-fail-closed-and-exhaustively-checked.md`
  (ready-to-start) already owns pinning the TUI's epic/audit view tables and
  `epicTransitions` to the domain vocabularies (scope item 5, AC 7), which is why those
  sites are absent from L2. L2's two sites are outside that scope and differ in kind —
  they cannot be pinned until the domain publishes an accessor to pin them to. Folding
  L2 in would widen that task from "add coverage tests" to "add domain API", so it is
  left as a separate proposal.
- Not a re-find, and not a disagreement: `planning/tasks/6gp1ytfvm652-create-epics-in-the-tui-with-shared-staged-form-mechanics.md`
  (completed 2026-10-22) deliberately chose the TUI's `active`/`medium` initial choices
  — "bounded selectors start at active and medium ... matching the CLI". M1 accepts that
  decision as correct given today's core and reports the core-side asymmetry that forces
  it. Whoever picks M1 up should read that task first, along with
  `6gpb4y5yqmjp` and `6gmshzwphadw` (both completed), which accepted the same
  per-kind explicit-defaults contract.
- No demotion, obsolescence or promotion candidates surfaced this run.

## Validation

- `git status --porcelain` — only `planning/` paths, before and after: the new audit
  file plus the two additive task annotations from step 10. No `internal/`, `cmd/`,
  `docs/` or `*.go` path appeared at any point.
- `just build` — **not available** (`just: command not found`); the spec's fallback
  `go build -o bin/tskflwctl ./cmd/tskflwctl` succeeded, exit 0.
- `./bin/tskflwctl -C . lint` — green: "all planning entities and dependency links
  pass lint".
- `./bin/tskflwctl -C . audit lint 2026-11-17-adapter-hygiene` — green: "all audit
  findings pass lint".
- `go test -race ./...` (the fallback for `just test`) — exactly two packages fail,
  **both pre-existing and environmental, neither caused by this run**, which changed
  only planning Markdown. Both were reported identically by the 2026-10-27
  adapter-hygiene run, so this is unchanged pre-existing red, not new breakage:
  - `internal/tools/reviewworkspace` — every failure is `error: rsync is required`;
    confirmed `rsync` is not installed in this container.
  - `cmd/tskflwctl` `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` —
    `main_test.go:188` expects a filesystem failure to exit 1, but the suite runs as
    uid 0 (confirmed `id -u` = 0), so the test's `chmod 0o500` does not block the write
    and `task new` succeeds with exit 0.
  - Everything else is green, `internal/tui`, `internal/core`, `internal/cli`,
    `internal/wire` and `internal/domain` included — the packages this lens read.
- `golangci-lint run ./...` — **could not run**: "the Go language version (go1.25) used
  to build golangci-lint is lower than the targeted Go version (1.26.0)". The Justfile
  already documents this toolchain requirement. No Go source changed in this run.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — `tskflwctl task new "Own epic creation defaults in core like task creation defaults" --epic 21-code-quality-architecture-hardening --tags core,validation,architecture --tier 2 --priority medium --description "Default blank epic status/priority inside NewEpic (mirroring prepareTaskCreation) so no adapter has to replicate them, and pin it with a NewEpic defaulting test."`
- ○ L1 · open — `tskflwctl task new "Return defensive copies from the domain vocabulary accessors" --epic 21-code-quality-architecture-hardening --tags domain,correctness --tier 4 --priority low --description "Make AllStatuses/AllAuditBuckets/AllEpicStatuses/TaskTransitions/AuditTransitions copy before returning, matching AllThreadStatuses and FindingBands/Efforts/Urgencies, and pin the no-aliasing contract with a test."`
- ○ L2 · open — `tskflwctl task new "Publish the priority/tier and finding-status orders from the domain" --epic 21-code-quality-architecture-hardening --tags domain,tui,cli,correctness --tier 4 --priority low --description "Add domain accessors for the ordered priority/tier choice sets and the declared finding-status lifecycle order, and replace the editEnumOptions and findingStatusOrder literals with them."`
- ○ L3 · open — `tskflwctl task new "Land a created task against the receipt status rather than an assumed one" --epic 18-tui-bubble-tea-interactive-planning-browser --tags tui,correctness --tier 4 --priority low --description "Use receipt.Task.Status in landCreatedTask instead of the hardcoded ready-to-start, so a future next-up/start creation affordance cannot land on an invisible row."`
