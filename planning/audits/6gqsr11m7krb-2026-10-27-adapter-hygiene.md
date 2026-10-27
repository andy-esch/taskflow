---
schema: 1
id: 6gqsr11m7krb
bucket: open
area: adapter-hygiene
date: "2026-10-27"
updated_at: "2026-10-27"
---

# Audit: adapter-hygiene — 2026-10-27

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

#### M1. Core composes CLI-specific remedy prose that every adapter and the JSON contract inherit · **Status:** open

**File:** internal/core/task_lifecycle.go:144 | **Component:** core
**Effort:** M · **Urgency:** soon

The 2026-09-15 adapter-hygiene audit's M1 correctly moved recovery guidance out of
`cli/render` and made it core-owned (shipped 2026-10-04, PR #282, task
`6ggdkztshmzz`). The ownership move was right and this finding does not relitigate
it. What moved, though, was the CLI's *prose*, verbatim — so the application layer
now authors strings in CLI vocabulary that two other consumers inherit.

Three spellings, all in core:

- `internal/core/task_lifecycle.go:144` and `:162` — `"pass --force to bypass only
  the blocked dependency gate"`. `--force` is a Cobra flag registered only in
  `internal/cli/task.go`.
- `internal/core/task_lifecycle.go:167` — ``"inspect the task with `tskflwctl task
  blockers <task>` before retrying"``: the CLI binary name, a literal `<task>`
  placeholder, and Markdown backticks.
- `internal/core/service_task.go:718` — ``fmt.Sprintf("`tskflwctl task blockers
  %s`", taskRef)``: the same, with the id interpolated.

The type comment at `task_lifecycle.go:120` states the intent plainly — "its Error
text remains a deterministic, actionable **CLI** diagnosis" — so this is a known
shortcut rather than an accident. The cost is that it is no longer CLI-only:

1. **The machine contract publishes it.** `internal/wire/dependency.go:345`
   describes the field as "core-owned inspection or recovery guidance", so the
   backticks and the CLI invocation ship inside a JSON string value. Verified (see
   Commands run): `task depend add --json` returns `remedy` = "inspect each
   affected task with `` `tskflwctl task blockers 6gqsq39trwh9` `` and restore sound
   prerequisites or update its dependencies". Epic 19 (`tskflwctl serve`, active)
   and any agent consumer gets Markdown decoration in a data field.
2. **The TUI replays it unfiltered.** `internal/tui/model.go:522` appends
   `receipt.Remedy` to the flash line, so a TUI user sees raw backticks and a CLI
   command in a status flash.
3. **The TUI is told to pass a flag it does not have.** `internal/tui/entity.go:360`
   always passes `TaskLifecycleOverrideNone` and exposes no force affordance, yet a
   refused start surfaces `TaskEligibilityError.Error()` — which ends in "pass
   --force to bypass only the blocked dependency gate".

This is the inverse of the usual adapter-hygiene leak: not an adapter re-deriving a
core decision, but core authoring one adapter's presentation. The remedy *decision*
(what is wrong, what would fix it) is core's; the *wording* is not.

**Recommendation:** Give the remedy/diagnosis a neutral core-owned shape (a stable reason code plus data: task id, gate, whether a dependency-gate override is permitted) and move the adapter wording — flag names, command invocations, backticks — into cli/render and the TUI's styles, so each adapter renders the same decision in its own vocabulary.

#### M2. Clipboard PATH probing runs on the Bubble Tea event loop inside Update · **Status:** open

**File:** internal/tui/clipboard.go:21 | **Component:** tui
**Effort:** XS · **Urgency:** soon

`AGENTS.md` states the non-negotiable: "TUI I/O runs through application services in
`tea.Cmd`, never in `Update`/`View`." `copyToClipboard` honours that for the write
but not for the probe that precedes it.

`internal/tui/clipboard.go:21-23` calls `clipboardArgv()` **eagerly**, in the
caller's goroutine, before returning any `tea.Cmd`:

```go
func copyToClipboard(text string) tea.Cmd {
	argv := clipboardArgv()            // <- runs on the caller's goroutine
	if argv == nil {
		return tea.SetClipboard(text)
	}
	return func() tea.Msg { ... }      // only the exec is deferred
}
```

`clipboardArgv` (`:39-53`) walks five candidates — `pbcopy`, `wl-copy`, `xclip`,
`xsel`, `clip.exe` — calling `exec.LookPath` on each until one resolves. On Linux
with no macOS/Wayland tooling that is at minimum two full `PATH` sweeps, each
`stat`-ing every `PATH` entry.

Both production call paths reach it from `Update`:

- `internal/tui/model.go:1623` (`yank`) ← `handleKey` (`model.go:810`) ← `update`
  (`:314`) ← `Update` (`:284`). Bound to `y` and `Y` (`keys.go:83-84`).
- `internal/tui/model.go:439`, directly inside `update`.

The third call site is written correctly and shows the intended pattern:
`copyCreationRecovery` (`internal/tui/staged_creation.go:314`) returns
`func() tea.Msg { return copyToClipboard(text)() }`, which defers the probe along
with the exec. Two sites eager, one deferred — an oversight, not a deliberate
asymmetry.

**Failing scenario:** a `PATH` containing a directory on an unresponsive network
mount (NFS/SSHFS over a dropped link, a stale automount). Pressing `y` on any row
enters `Update`, which blocks in `exec.LookPath`'s `stat` until the filesystem
timeout. Bubble Tea's event loop is single-threaded through `Update`, so the whole
TUI freezes — no repaint, no input, no spinner — for the duration, and the user has
no signal that a clipboard probe is the cause. Off the pathological case the cost is
small but still unbounded-by-contract work on the render loop.

**Why tests didn't catch it:** the clipboard tests
(`6ffdv9g03dxv`, completed) assert *which* argv is chosen and that the fallback
fires, never *where* the probe executes; no test asserts that `Update` performs no
filesystem access.

**Recommendation:** Move the clipboardArgv() probe inside the returned tea.Cmd closure (one func() tea.Msg that decides native-vs-OSC-52 and then acts), the way copyCreationRecovery at internal/tui/staged_creation.go:314 already defers it.

#### L1. The in-flight member partition is unpublished, so four adapter sites recompute it · **Status:** open

**File:** internal/cli/render/thread.go:206 | **Component:** cli/render + tui
**Effort:** S · **Urgency:** eventually

`core.ThreadView` (`internal/core/thread_projection.go:65`) publishes `Frontier` as
a first-class derived slice but leaves the complementary "in flight" partition to
callers. Four adapter sites recompute it from the primitive role field:

- `internal/cli/render/thread.go:206` — `ThreadFrontierHuman` filters `Members`
- `internal/tui/item.go:239` — `activityForThread`'s `▶` count
- `internal/tui/detail.go:1310` — the detail pane's "In flight" section
- `internal/tui/thread_spatial.go:429` — spatial anchor preference

All four spell it `member.State.Role == core.RoleInFlight` and therefore agree
today; this is an ownership observation, not a wrong-output defect, and it is
labelled Low accordingly. No failing scenario exists on the current tree.

What makes it worth recording is the asymmetry: `Frontier` is derived once in core
and consumed four times, while its sibling partition is derived four times from raw
state. Core already has the vocabulary — `isPendingWorkRole`
(`internal/core/dependency_graph.go:180`) is unexported precisely because the
partition is core's to own.

Adjacent (not duplicated here): `internal/cli/render/thread.go:263` re-sorts the
frontier by task id while `internal/tui/detail.go:1316` renders core's persisted
member order and `wire` publishes that same order — one projection, three
orderings. Sort order is a legitimate presentation choice, so this is noted only as
context for whoever consolidates the partition.

**Recommendation:** Publish the partition once — either an InFlight slice beside Frontier on core.ThreadView, or an exported predicate over TaskGraphState — and have the four sites consume it; fold into 6gppr2kfa1dz, which already consolidates the neighbouring marker predicates in three of the same files.

#### L2. The TUI's archived-task filter negates two statuses where the domain owns the predicate · **Status:** tracked by 6gcqz5aqt2sg

**File:** internal/tui/commands.go:322 | **Component:** tui
**Effort:** XS · **Urgency:** eventually

`dropArchivedRecords` (`internal/tui/commands.go:319-328`) decides the TUI's default
working view with a negated status chain:

```go
if record.Value.Status != domain.StatusCompleted && record.Value.Status != domain.StatusDeprecated {
```

`internal/domain/status.go:81` owns the lifecycle partition and says so — "The
lifecycle invariant lives on the domain type, not in a use-case" — and
`ActiveStatuses` (`:38`) adds "so 'the active set' has one definition and one order,
owned here in the domain rather than re-derived per use case". Every other consumer
honours that: `core/board.go:51`, `core/service.go:524`/`:828`,
`core/service_task.go:72`, `cli/schema.go:88`. This is the one site that hand-rolls
it.

The TUI's intent is legitimately *not* `IsActive()` — the working view deliberately
keeps deferred tasks visible as reminders (`commands.go:60-64`), so "not archived"
is a third, wider partition than "active". The problem is only that it is spelled as
a negation rather than owned as a predicate: `IsActive()`'s `default: false` fails
closed on an unknown status, while a `!=` chain fails open.

No failing scenario on the current tree — the six statuses partition exactly as
intended, so this is Low. The drift is latent: adding a terminal status to
`domain.allStatuses` would silently admit it to the TUI's default view (and
`rankOf`, `commands.go:310`, would sort it below deferred), while the domain's own
predicate would have excluded it.

Tracked: `6gcqz5aqt2sg` already carries this exact item in scope — "replace the
archived-task filter's two-status negation with a domain predicate (added in
2026-10-23 audit triage)".

**Recommendation:** Already in scope for 6gcqz5aqt2sg — replace the negation with a domain predicate (an Archived()/IsArchived() beside Status.IsActive()) so the partition has one owner and a new status fails closed.

**Resolution:** Already carried in that task's scope as 'replace the
    archived-task filter's two-status negation with a domain predicate' (added
    in 2026-10-23 audit triage); this audit only names the exact site and the
    domain owner to reuse.

Routine: `code-quality-audit` · lens `adapter-hygiene` · ISO week `2026-W44`,
slot `Tue` (index 4).

## Punch list

- M1. Core composes CLI-specific remedy prose that every adapter and the JSON contract inherit  (effort: M · urgency: soon)
- M2. Clipboard PATH probing runs on the Bubble Tea event loop inside Update  (effort: XS · urgency: soon)
- L1. The in-flight member partition is unpublished, so four adapter sites recompute it  (effort: S · urgency: eventually)
- L2. The TUI's archived-task filter negates two statuses where the domain owns the predicate  (effort: XS · urgency: eventually) — tracked by `6gcqz5aqt2sg`

No Critical and no High. Both Mediums are seam-ownership defects with concrete
consequences; neither corrupts data.

## Files audited

- **Signal**: `internal/core/service_task.go` — 9 commits in 30 days and the highest
  blast radius in the lens surface (`core` is referenced by 103 production files).
- **Signal**: `internal/wire/wire.go` — 14 commits in 30 days, the most-churned file
  in the surface; `wire` is referenced by 32 production files.
- **Adjacency** (from `service_task.go`): `internal/cli/task.go` — across the
  hexagonal seam to the primary adapter that drives those use cases, where an
  adapter would most plausibly re-derive lifecycle policy.
- **Adjacency** (from `wire.go`): `internal/cli/render/thread.go` — imports `wire`
  and was one of the hits from this lens's high-yield `core.Role|Gate|Status` grep.
- **Random**: `internal/core/service_research.go` — passed the quality filter
  (251 lines, not generated, not a fixture, not a test).

Read as supporting context, not as picks: `core/task_lifecycle.go`,
`core/thread_projection.go`, `core/dependency_graph.go`, `domain/status.go`,
`wire/thread.go`, `wire/dependency.go`, `tui/item.go`, `tui/commands.go`,
`tui/detail.go`, `tui/thread_spatial.go`, `tui/clipboard.go`, `tui/model.go`,
`tui/entity.go`, `cli/root.go`, `cli/theme.go`, `cli/render/markdown.go`.

M2 came out of this lens's "does `tui` do I/O in `Update`/`View`" question rather
than from one of the five picks — the `exec`/`os` sweep over `internal/tui` led to
`clipboard.go`. Recording that so the discovery path is auditable.

## Commands run

- `go build -o bin/tskflwctl ./cmd/tskflwctl` — OK (`just` is not installed in this
  container, so the documented fallback was used).
- `git log --since="30 days ago" --name-only` over the lens paths — churn ranking.
- `grep -rn "core\.Role\|core\.Gate\|core\.Status" internal/cli internal/tui internal/wire`
  — 18 production hits; triaged, most are legitimate display of a core decision.
- Mutation probe in a throwaway planning space outside this checkout
  (`tskflwctl init --no-register` in a scratch dir, one epic, two tasks, one
  dependency). This is how M1's claims were verified rather than inferred:
  - `task depend add dependent-work --on prereq-work` printed
    ``remedy: inspect each affected task with `tskflwctl task blockers 6gqsq39trwh9` and restore sound prerequisites or update its dependencies``
    — literal Markdown backticks on a terminal line.
  - `task depend add ... --json` returned that same string, backticks included, as
    the `remedy` JSON string value.
  - `task start dependent-work` exited 11 with
    `...; pass --force to bypass only the blocked dependency gate`.
  - M1's TUI consequences (points 2 and 3) are **read-verified only** — traced
    through `tui/model.go:522` and `tui/entity.go:360`; the TUI was not driven
    headlessly.
- `./bin/tskflwctl -C . lint` — green.
- `./bin/tskflwctl -C . audit lint 2026-10-27-adapter-hygiene` — green.
- `git status --porcelain` — only `planning/` paths, before and after.
- `go test -race ./...` — two packages fail, **both pre-existing and
  environmental, neither caused by this run** (this run changed only planning
  Markdown):
  - `internal/tools/reviewworkspace` — every failure is `error: rsync is required`;
    `rsync` is not installed in this container.
  - `cmd/tskflwctl` `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` —
    `main_test.go:180` does `os.Chmod(tasksDir, 0o500)` and expects exit 1; the
    suite runs as uid 0, which ignores the write bit, so the write succeeds and the
    assertion fails. Everything else, `internal/tui` and `internal/wire` included,
    is green.
- `golangci-lint run ./...` — **could not run**: "the Go language version (go1.25)
  used to build golangci-lint is lower than the targeted Go version (1.26.0)". The
  Justfile already documents this requirement. No Go source changed in this run.

## What audited clean

- `internal/core/service_research.go` (random pick) — clean, and a good model for
  the seam: no fs, no cobra, every write through the store port, `updated_at`
  stamped in core "so every adapter gets it", protected-field refusals carrying the
  domain's own reason so all adapters explain them identically. The
  regenerate-on-collision loop is bounded and documented.
- `internal/wire/wire.go` (signal pick) — clean. It is declaration plus one
  encoder: the revision changelog, the policy constants and `EncodeJSON`. No
  decision is recomputed here; the `ToXEnvelope` mappers in sibling files copy core
  values rather than re-deriving them, which is exactly what the neutral-leaf role
  requires.
- `internal/cli/task.go` (adjacency pick) — clean after a closer look that expected
  otherwise. `lifecycleOverride` (`:931`) maps one `--force` flag onto a typed
  override and looks like adapter-owned policy, but core validates the
  override/destination pairing at `core/task_lifecycle.go:316-324` and rejects a
  mismatch with `ErrValidation`. The adapter translates; it does not decide, and a
  wrong translation fails closed.
- `internal/core/service_task.go` (signal pick) — clean apart from M1's one line.
  No fs, no cobra, no path handling; `activeOnly` filtering goes through
  `domain.Status.IsActive()` rather than a local spelling.
- Terminal-capability detection — clean, and deliberately so. `cli/theme.go:101`
  and `cli/root.go:543` pass `os.Stdin, os.Stdout` to `HasDarkBackground`, which
  looks like a process global inside a DI'd adapter, but the OSC-11 query must
  reach the real terminal, and both call sites are gated on the **injected** writer
  (`wantColor(..., app.Out)`, `st.on`) and evaluated lazily, so a captured writer
  never triggers it.
- `core` filesystem purity — `filepath` appears in `core/dependency_graph.go:672`
  (`ToSlash`) and `core/space_registry.go:225` (`Base`/`Clean`), both pure string
  work, no I/O. The stronger import ban is already tracked by `6ggxymjbf86r`.
- Adapter persistence construction — clean. The only `config` imports in `cli` are
  `init.go` and `workspace.go`, both named exceptions in the boundaries guide; no
  adapter imports `store` or the `*store` families.

## External research

Done because this run surfaced 2 Medium-or-higher findings, below the step-11
threshold of 3. Three WebSearch queries. **WebFetch could not be used**: every
external host failed DNS resolution under this environment's network policy
(`charm.sh`, `charm.land`, `mintlify.com` all returned `ENOTFOUND`), so the notes
below rest on search-result summaries rather than on pages read end to end. Treat
them as corroborating direction, not as citations verified against primary text.

- **Bubble Tea: no I/O in `Update`/`View`** — corroborates M2 as an established
  rule rather than a local preference. The Charm community guidance is that
  commands exist for I/O and that `Update` must return fast, with blocking work
  (HTTP, disk, sleep) moved into a `tea.Cmd` the runtime executes on its own
  goroutine; `View` must be side-effect free. Sources:
  <https://charm.sh/blog/commands-in-bubbletea/>,
  <https://www.mintlify.com/charmbracelet/bubbletea/concepts/commands>,
  <https://www.mintlify.com/charmbracelet/bubbletea/guides/async-operations>.
- **Hexagonal Go: translate errors at the edge** — no source addressed the exact
  question of CLI flag names inside application-layer error text, which is worth
  stating plainly. The general principle lines up with M1: adapters translate
  errors for external interfaces, the inner layers stay free of interface concerns,
  and flag handling belongs to the command layer (one project enforces this with an
  architecture test forbidding `flag` imports in core/app code — a mechanism this
  repo could mirror in `architecturecheck`). Sources:
  <https://alamrafiul.com/posts/go-hexagonal-architecture/>,
  <https://github.com/illud/gohex>, <https://tomevault.io/tome/sufield/stave>.
- **Machine-readable error contracts: a code *plus* prose, not instead of it** —
  this tempered M1's recommendation. The recurring practitioner pattern is a stable
  machine-readable code with an optional human-readable suggestion and help
  reference in the same envelope, with prose retained for humans; AWS CLI keeps a
  human-oriented default and offers a JSON error form for automation. There is no
  settled standard mandating structured remedy codes. So M1's fix should **add** a
  neutral reason code that adapters render, not delete the prose field — removing
  `remedy` would be a NOT-ADDITIVE break under ADR-0008 for no gain. Sources:
  <https://macmyths.com/cli-errors-are-part-of-your-agent-api/>,
  <https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-error-format.html>,
  <https://instagit.com/iofficeai/OfficeCLI/officecli-structured-error-codes-self-healing-workflows/>.

An antipattern this codebase does **not** have, worth recording as a negative
result: no adapter reaches around `core.Service` to construct persistence, and no
`core` file imports cobra or does file I/O — the two failures this lens most often
finds in Go CLI/TUI codebases.

## Related-task observations (propose-only)

- Not a re-find, and not a disagreement:
  `planning/tasks/6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md`
  (completed 2026-10-05, PR #282) deliberately moved recovery guidance from
  `cli/render` into core, closing the 2026-09-15 adapter-hygiene audit's M1. M1
  here accepts that ownership decision and reports only that the migrated text kept
  its CLI vocabulary. Whoever picks M1 up should read that task first.
- Partial overlap, left open:
  `planning/tasks/6gppr2kfa1dz-stop-branding-readable-broken-gate-thread-members-as-unreadable.md`
  (ready-to-start) already consolidates the readable/gate-sound marker predicates in
  three of L1's four files. L1's predicate is a different one, so folding it in
  would widen those acceptance criteria — a scoping call for a human, not something
  this run should decide. Cross-referenced on both sides.
- Fully tracked: `planning/tasks/6gcqz5aqt2sg-make-closed-vocabulary-switches-fail-closed-and-exhaustively-checked.md`
  (ready-to-start) already carries L2 verbatim in scope. L2 is marked
  `tracked by 6gcqz5aqt2sg` rather than becoming a new task.
- No demotion or obsolescence candidates surfaced this run.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- → L2 · tracked — Tracked in planning/tasks/6gcqz5aqt2sg-make-closed-vocabulary-switches-fail-closed-and-exhaustively-checked.md
- ○ M1 · open — `tskflwctl task new "Make lifecycle and dependency remedies adapter-neutral" --epic 21-code-quality-architecture-hardening --tags architecture,core,recovery,json --tier 2 --priority medium --description "Give core-owned remedies a neutral shape (reason code plus data) and move CLI flag/command wording into the adapters."`
- ○ M2 · open — `tskflwctl task new "Defer clipboard PATH probing out of the TUI Update loop" --epic 18-tui-bubble-tea-interactive-planning-browser --tags tui,correctness --tier 3 --priority medium --description "Move clipboardArgv() inside the returned tea.Cmd so no filesystem probe runs in Update, and pin it with a test."`
- ○ L1 · open — `tskflwctl task new "Publish the in-flight Thread member partition once in core" --epic 30-threads-and-task-dependency-graphs --tags threads,core,tui --tier 4 --priority low --description "Derive the in-flight partition beside Frontier in core and consume it from the four adapter sites; consider folding into 6gppr2kfa1dz."`
