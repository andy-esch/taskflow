---
schema: 1
id: 6ga93gvffkpc
bucket: open
area: adapter-hygiene
date: "2026-09-15"
---
# Code Quality Audit: adapter-hygiene — 2026-09-15

> Edit findings through `tskflwctl audit finding` so status and resolution
> metadata stay queryable. Never hand-edit a `**Status:**` line.

Routine: `code-quality-audit` · lens `adapter-hygiene` · ISO week `2026-W38`,
slot `Tue` (index 4).

## Punch list

One line per finding for fast triage. Order: Critical → High → Medium → Low.

- `M1. Dependency mutation receipts carry no remedy; the CLI adapter authors core's recovery prose  (effort: S · urgency: soon)`
- `M2. MarkerUnreadable brands readable tasks whose prerequisite is broken  (effort: S · urgency: soon)`
- `L1. The "newly unsafe impact" predicate is spelled three times, twice inside the adapter  (effort: S · urgency: eventually)`
- `L2. The spatial anchor guard is narrower than its sibling render guards  (effort: XS · urgency: eventually)`

## Files audited

- **Signal**: `internal/wire/wire.go` — churn 8 × blast radius 23 (score ≈ 6.98), the
  highest in the lens surface; it is the neutral machine contract every adapter imports.
- **Signal**: `internal/tui/thread_spatial.go` — churn 8 × blast radius 7 (score ≈ 4.57);
  2,997 lines and the densest concentration of `core.Role`/`core.Gate` reads in the tree.
- **Adjacency** (from `internal/wire/wire.go`): `internal/cli/render/dependency.go` — the
  hop across the hexagonal seam, from the wire contract to the primary adapter that
  publishes the dependency receipt to humans.
- **Adjacency** (from `internal/tui/thread_spatial.go`): `internal/tui/detail.go` — sibling
  in the same package sharing the Thread member/node rendering derivation.
- **Random**: `internal/core/service_research.go`

Read for context while evaluating the above: `internal/cli/render/render.go`,
`internal/core/service_task.go`, `internal/core/dependency_graph.go`,
`internal/core/dependency_operations.go`, `internal/core/thread_graph.go`,
`internal/wire/dependency.go`, `internal/theme/theme.go`.

## Commands run

- `go build -o bin/tskflwctl ./cmd/tskflwctl` → exit 0. (`just` is not installed in the
  routine's container; this is the fallback the routine spec names.)
- `go test ./...` → all packages ok, no FAIL. Not pre-existing red.
- `./bin/tskflwctl lint` → `✔ all planning entities and dependency links pass lint`.
- `git log --since="30 days ago" --name-only --pretty=format: -- internal/cli/ internal/tui/
  internal/wire/ internal/core/service*.go` → churn ranking, with generated goldens and
  `_test.go` excluded per the lens quality filter.
- `grep -rn "core\.Role\|core\.Gate\|core\.Status" internal/cli internal/tui internal/wire`
  → the lens's high-yield probe; 20 production hits, which is where M2/L1/L2 came from.
- Contract check for M1, run against the committed published schema rather than inferred
  from the Go types:
  `python3 -c "…" internal/cli/testdata/golden/schema_jsonschema.golden` →
  `DependencyMutationEnvelope -> NO REMEDY`, `DependencyMutationJSON -> NO REMEDY`,
  `TaskLifecycleJSON -> remedy`, `TaskEligibilityFailureJSON -> remedy`.

## Findings

### Critical

(none)

### High

(none)

### Medium

#### M1. Dependency mutation receipts carry no remedy; the CLI adapter authors core's recovery prose  · **Status:** open

**File:** `internal/cli/render/dependency.go:223` (and `internal/core/dependency_operations.go:42`) | **Component:** cli/render + core
**Effort:** S · **Urgency:** soon

Every other mutation receipt in `core` publishes its own recovery guidance as data:
`TaskLifecycleMutationResult.Remedy` (`service_task.go:396`, built by `taskLifecycleRemedy`),
`ThreadMutationReceipt.Remedy` (`service_thread.go:200`), `TaskRenameReceipt.Remedy`
(`service_task.go:682`), plus `TaskEligibilityError.Remedy()`. `DependencyMutationReceipt`
is the one exception — it has no `Remedy` field — and `DependencyMutationHuman` fills the
gap by writing the prose itself:

```go
// internal/cli/render/dependency.go:223
fmt.Fprintf(w, "  %s inspect with `tskflwctl task blockers %s`; restore a sound prerequisite or remove the edge\n", ...)
```

That is a primary adapter deciding what a user should do about a core-derived condition —
the decision core owns everywhere else in this package. Two consequences follow, and the
second is the one that bites.

First, the advice has forked. Core's text for the same condition is "inspect each affected
task with `tskflwctl task blockers <task>` and restore sound prerequisites or update its
dependencies"; the adapter's says "…or remove the edge". Neither is wrong; they are simply
two strings nobody can keep in step, for one condition.

Second, the remedy exists only on the human path. `wire.DependencyMutationJSON`
(`internal/wire/dependency.go:307`) has no remedy field, so `--json` consumers get the
impacts and no guidance — on a tool whose stated audience is agents.

**Failing scenario:** in a planning repo, `tskflwctl task depend add <prereq> <dependent>`
where the new edge drives a third task's gate from `clear` to `blocked` (or newly
`Inconsistent`). Human output prints the `remedy:` line. The same command with `--json`
emits `{"operation":"add",…,"impacts":[…]}` with **no** `remedy` key — verified against the
committed JSON Schema, which lists `remedy` on `TaskLifecycleJSON` and
`TaskEligibilityFailureJSON` and not on `DependencyMutationEnvelope`. An agent that runs
`task start --json` gets a remedy for a newly-unsafe impact; the same agent running
`task depend add --json` against the same condition gets nothing, and must either parse the
human stream it was told not to parse or re-derive the advice itself.

**Why tests didn't catch it:** the goldens assert the envelope shape that exists, which is
exactly the shape with the gap. There is no cross-receipt invariant test asserting that
every receipt carrying `impacts` also carries `remedy`, so the omission is only visible by
comparing two schema definitions side by side. Changelog entry 1.52 in `wire.go` records
the asymmetry in prose — lifecycle rows got "an explanatory remedy", dependency receipts got
"the same before/after impact shape" — without flagging it as a gap.

**Recommendation:** add `Remedy string` to `core.DependencyMutationReceipt` and populate it
in core from the existing `taskImpactsNeedRepair` predicate (see L1), map it in
`wire.ToDependencyMutationJSON`, and reduce `DependencyMutationHuman` to printing
`receipt.Remedy`. Additive on the wire — a `remedy,omitempty` field is a minor
`SchemaVersion` bump under the rule stated at the top of `wire.go`.

**Tightening (adjacent):** while the remedy moves into core, settle the two prose variants
into one sentence. "Remove the edge" is genuinely useful advice that core's version omits;
it should survive the merge rather than be dropped in favour of the existing core wording.

**Follow-up:** the general invariant — every core receipt that reports `impacts` also
reports a `remedy` — would be better enforced than remembered. A table-driven wire test over
the receipt DTOs would catch the next one; that is its own task, not part of this fix.

#### M2. MarkerUnreadable brands readable tasks whose prerequisite is broken  · **Status:** open

**File:** `internal/tui/detail.go:1322` (also `detail.go:1467`, `thread_spatial.go:2774`, `thread_spatial.go:2792`, `thread_spatial.go:2896`) | **Component:** tui
**Effort:** S · **Urgency:** soon

`theme.MarkerUnreadable` is declared as a cross-surface token with one meaning. The
`internal/theme` doc comment is explicit that markers exist "so the row delegates, the
dashboard, and the `?` legend draw ONE glyph+color per concept instead of re-typing them —
and so the legend can't drift from the rows", and names this one "the unreadable-file !".
`dashboard.go:209` uses it in exactly that sense: `"%d unreadable file(s) (run lint)"`.

The Thread surfaces apply it to a different concept. Five sites brand a node red-and-`!`
whenever `State.Role == core.RoleUnknown || State.Gate == core.GateBroken`:

```go
// internal/tui/detail.go:1322
if member.Task.Slug == "" || member.State.Role == core.RoleUnknown || member.State.Gate == core.GateBroken {
        marker = s.fg(theme.ColorRed, theme.MarkerUnreadable.Glyph)
```

`GateBroken` does not mean the record was unreadable. `TaskGraph.gate`
(`dependency_graph.go:765`) returns `GateBroken` when **any prerequisite** fails
`computeSound` — the task's own file parsed perfectly. So a task with a real status, slug and
description is displayed with the glyph reserved for a file the tool could not read, and the
user is pointed at the wrong document.

**Failing scenario:** task `A` has `status: next-up` and `depends_on: [B]` where `B` is
missing or unreadable. `deriveState(A)` returns `{Role: queued, Gate: broken}` — `roleForStatus`
maps `next-up` to `RoleQueued`, and `gate` returns `GateBroken` from `B`. Open `tskflwctl ui`,
select a Thread containing `A`, and `A` renders in the members list (`detail.go:1322`) and in
the spatial canvas (`thread_spatial.go:2774`, `:2792`) with the red `!` that the dashboard
and the `?` legend both define as "unreadable file". `A` is entirely readable. The broken
record is `B`, which may not even be a Thread member and so may not be on screen at all.

**Why tests didn't catch it:** the TUI suite asserts on `View()` substrings and model state,
and there is a drift test pinning the marker *glyphs* (`theme_test.go:174`) — but nothing
pins a marker's *meaning* to the predicate that selects it. A test can only catch this by
constructing a readable task with a dangling prerequisite and asserting the absence of the
unreadable glyph, which no fixture does.

**Context:** severity sits above what a display nit would earn because the glyph actively
misdirects diagnosis — it tells the user to `run lint` on a file that is fine — and because
the repo has an explicit standing rule against exactly this class of claim ("an adapter
cannot claim that a missing checkout is healthy", `docs/ARCHITECTURE.md`). This is the same
error with the sign flipped.

**Recommendation:** separate the two concepts at all five sites. Keep
`theme.MarkerUnreadable` for genuinely unreadable records (`Task.Slug == ""`,
`Role == RoleUnknown`) and give a broken **gate** its own marker — `theme.MarkerWarn` (⚠,
yellow) already exists and reads correctly for "this is sound as a record, unsound as a
dependency". The cleanest shape is one helper in `internal/tui` returning the marker for a
`core.TaskGraphState`, called by all five sites, so the concept is chosen once.

**Tightening (adjacent):** `thread_spatial.go:2896` colours the `role/gate` inspector line
by the same conflated predicate. Once the helper exists it should drive that switch too, so
the inspector's colour and the node's marker cannot disagree.

**Follow-up:** whether `core` should publish this distinction as a named derived boolean on
`TaskGraphState` (it already publishes `Eligible`, `Drained`, `Inconsistent`,
`SoundlyCompleted`) is a design call worth its own task — see L1, which is the same question
arriving from the CLI side.

### Low

#### L1. The "newly unsafe impact" predicate is spelled three times, twice inside the adapter  · **Status:** open

**File:** `internal/cli/render/render.go:291` (also `internal/cli/render/dependency.go:215`, `internal/core/service_task.go:414`) | **Component:** cli/render + core
**Effort:** S · **Urgency:** eventually

One rule — "this impact went newly unsafe" — is written out three times:

```go
// core/service_task.go:414, inside the unexported taskImpactsNeedRepair
if (!impact.Before.Inconsistent && impact.After.Inconsistent) ||
        (impact.Before.Gate != impact.After.Gate && impact.After.Gate != GateClear) {

// cli/render/dependency.go:215  — against the typed core value
newlyUnsafe := (impact.Before.Gate != impact.After.Gate && impact.After.Gate != core.GateClear) || …

// cli/render/render.go:291  — against the wire DTO, where Gate is a string
if (impact.Before.Gate != impact.After.Gate && impact.After.Gate != string(core.GateClear)) || …
```

Core already owns the predicate; it is simply unexported, so the adapter re-types it to pick
a `⚠` over a `•`. The third spelling additionally lives in a different type world — the wire
DTO stringifies `Gate`, hence the `string(core.GateClear)` conversion — so a future change to
`GateState` will be caught by the compiler at two sites and not at the third.

No failing scenario: all three spellings agree today, which is why this is Low rather than
Medium. It is a drift risk, not a live defect, and it is listed because it is the mechanism
behind M1 — the same rule being unexported is what left the dependency receipt without a
remedy in the first place.

**Recommendation:** export the predicate from `core` over `TaskGraphStateImpact` (for
example `func (i TaskGraphStateImpact) NeedsRepair() bool`), have `taskImpactsNeedRepair`
call it, and have both render sites call it too. `render.go` can only do so once it holds
the core value rather than the wire DTO; the cheaper interim is for wire to carry the
decision as a field, which M1's fix makes natural.

**Tightening (adjacent):** fold this into M1's change rather than shipping it alone — M1
needs the predicate reachable from core's dependency path regardless, so the two land as
one edit.

#### L2. The spatial anchor guard is narrower than its sibling render guards  · **Status:** open

**File:** `internal/tui/thread_spatial.go:428` (also `:443`) | **Component:** tui
**Effort:** XS · **Urgency:** eventually

Anchor selection skips members with `taskID == "" || member.Task.Slug == "" ||
member.State.Role == core.RoleUnknown` — the same guard as M2's render sites, minus the
`Gate == GateBroken` clause. The nearby comment states the intent: "A readable external
prerequisite is a better last-resort anchor than an unreadable member record."

So in M2's scenario, task `A` (readable, gate broken) is eligible to become the view's focal
anchor at `:428` while being drawn with the unreadable marker at `:2774`. The view can open
focused on a node it simultaneously brands unreadable.

Worth noting for honesty: `:428` is arguably the *correct* one of the two. A readable task
with a broken gate genuinely is anchorable — it can be opened, inspected and fixed — which
is precisely M2's argument that the render sites are the sites that are wrong. The finding
here is narrower: the two guards are written as if they answered the same question, and
nothing in the code says which question each is asking.

**Recommendation:** once M2's helper exists, give the two questions two names —
"is this record readable" (anchorable) and "is this task's gate sound" (marker colour) — and
call the right one at each site. No behaviour change is intended at `:428`/`:443`.

## What audited clean

- `internal/wire/wire.go` — **clean, and the lens's best result.** The file is
  `SchemaVersion`, `EncodeJSON`, and a 60-entry changelog. The package doc states the
  neutrality rule ("no cobra, no lipgloss… a machine wire contract is an API, not
  presentation") and the file honours it exactly: no core decision is re-derived, no
  presentation leaks in, and the value-constructor/emit-func split it describes is real in
  the code. The changelog is maintained to a standard that made M1 findable — entry 1.52
  is what records that the dependency receipt got impacts without the lifecycle's remedy.
- `internal/core/service_research.go` — **clean.** The random pick, and nothing to report.
  No `os`/`filepath`/cobra import; errors wrap `domain.ErrValidation` and `domain.ErrConflict`
  so the CLI's exit-code mapping holds; the newest-first ordering and the bounded
  id-mint retry live in core rather than in a renderer; and protected-field refusals come
  from `domain.ProtectedResearchField` so every adapter explains them identically. The one
  place it could plausibly have leaked — a default-view carve like `ListTasks` has — is
  deliberately absent and the comment says why.
- `internal/cli/render/dependency.go` — **clean on the machine path.** `DependencyMutationJSON`
  is a three-line wrapper over `wire.ToDependencyMutationEnvelope`, which is precisely the
  structure `docs/ARCHITECTURE.md` prescribes; a web adapter can obtain the same value. Its
  human path is M1 and L1, but the seam itself is right.
- `internal/tui/detail.go` and `internal/tui/thread_spatial.go` — **clean on the
  non-negotiables the lens checks first.** Despite 4,758 combined lines, neither does I/O in
  `Update`/`View`, neither imports `store` or any config package, and every read arrives as a
  `tea.Msg`. The marker defects (M2, L2) are display-semantics bugs inside a boundary that is
  otherwise correctly held.

## External research

Two Medium findings is under the routine's threshold of three, so light research on the lens
follows. **Caveat on citations:** the routine's container permits `WebSearch` but its egress
proxy blocked `WebFetch` on all three articles below (`EGRESS_BLOCKED`), so what is reported
is the search-result summary, not the fetched text. Treat these as pointers to read, not as
verified quotations.

- **The canonical antipattern for this lens has a name, and M2/L1 are it.** The recurring
  failure mode in ports-and-adapters write-ups is that "people tend not to take the 'lines'
  in the layered drawing seriously, letting the application logic leak across the layer
  boundaries", with the standing recommendation that adapters be thin translation layers
  rather than places where domain logic is re-implemented. That is exactly L1 (a core
  predicate re-typed in `render`) and M2 (an adapter deciding what a derived state *means*).
  <https://alistair.cockburn.us/hexagonal-architecture> ·
  <https://dev.to/buarki/hexagonal-architectureports-and-adapters-clarifying-key-concepts-using-go-14oo>
- **Remediation-as-data is now the stated norm for agent-facing CLIs, which is M1's case
  made externally.** Current writing on designing CLIs for agents argues that a tool serving
  both humans and agents from one binary needs structured JSON output *and* remediation
  fields, with the explicit formulation that agents need remediation **declared, not
  inferred** — a `suggestion`/next-step field giving a concrete recovery path. taskflow
  already does this on three receipts; M1 is the fourth that doesn't.
  <https://blog.arcjet.com/designing-a-cli-for-ai-agents/> ·
  <https://zircote.com/blog/2026/04/cli-error-messages-are-a-dual-consumer-problem/>
- **Antipatterns this codebase does NOT have** — worth recording, since the lens asks for
  proactive recommendations and not only defects. The common Go adapter-hygiene failures are
  package-level globals standing in for DI, writing to `os.Stdout` instead of an injected
  writer, and a core that imports its persistence library. This tree has none: DI runs
  through one `*cli.App` in `PersistentPreRunE`, output goes through injected `io.Writer`s,
  and the `core`/`wire` import restrictions are enforced executably by `depguard` in
  `.golangci.yml` rather than by convention. The lens's own high-yield grep
  (`core.Role|core.Gate|core.Status` in adapters) returned 20 production hits and **all 20
  are reads of core-derived state**, not re-derivations of it from primitive fields — the
  failure this lens exists to catch is largely absent. M2 and L1 are the residue.
- **On single-sourcing a predicate (L1's fix).** The standard argument is that one tested
  definition removes the risk of duplicated predicates drifting apart in their treatment of
  edge cases; the taskflow-specific version of that argument is already written down in
  `docs/ARCHITECTURE.md` — "A word that means the same thing in two places is spelled once"
  — with `domain/resolution.go` as the worked example and a drift test enforcing it. L1 and
  M2 are two more words that are currently spelled twice and five times respectively.
  <https://en.wikipedia.org/wiki/Single_source_of_truth>

## Candidate tasks (human to triage)

Cross-reference (step 10) found **no** open task covering any of these findings. Every active
task body was grepped for `DependencyMutationReceipt`, `taskImpactsNeedRepair`, `GateBroken`,
`RoleUnknown`, `newlyUnsafe`, `dependency_mutation` and `remedy`: zero matches outside
completed work. No finding was marked `tracked`, and no task was annotated.

- `tskflwctl task new "Let core own the dependency-mutation remedy" --epic 21-code-quality-architecture-hardening --tags core,cli,graph --tier 3 --priority medium --description "Add Remedy to core.DependencyMutationReceipt (from the shared impact predicate), map it in wire, and reduce DependencyMutationHuman to printing it — closing the --json remedy gap. Covers M1 + L1."`
- `tskflwctl task new "Stop branding broken-gate tasks as unreadable in the TUI" --epic 18-tui-bubble-tea-interactive-planning-browser --tags tui,threads,design --tier 3 --priority medium --description "Split 'unreadable record' from 'broken gate' across the five Thread marker sites; one helper over core.TaskGraphState chooses the marker. Covers M2 + L2."`

## Related-task observations (propose-only)

- Precedent, not overlap: `planning/tasks/6fgcr2403nbs-let-core-own-the-dashboard-aggregates-adapters-re-derive.md`
  (status `completed`, epic 21) resolved this exact class of finding in June — it deduplicated
  a settled-count predicate out of `render.go` + `dashboard.go` into core, under the heading
  "Adapters should read aggregates, not re-derive off raw Summary lists." M1/L1 are the same
  shape one surface over, which is evidence the pattern recurs rather than evidence it was
  missed. Worth considering whether the fitness rule deserves a lint check rather than a
  recurring audit finding.
- No task looks obsolete, mis-scoped, or promotable as a result of this run.
