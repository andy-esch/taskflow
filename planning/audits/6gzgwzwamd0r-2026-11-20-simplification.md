---
schema: 1
id: 6gzgwzwamd0r
bucket: open
area: simplification
date: "2026-11-20"
updated_at: "2026-11-20"
---

# Audit: simplification — 2026-11-20

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

#### M1. Five creation controllers duplicate one receipt-interpretation policy, and it has already drifted · **Status:** open

**File:** internal/tui/task_create_commands.go:120 | **Component:** tui/creation
**Effort:** M · **Urgency:** soon

Each of the five TUI creation kinds carries its own `handle<Kind>Created` controller
that interprets the typed receipt after an issued write: displacement, the
committed/landed path, the finalization-uncertainty classification, the locked
"inspect the original attempt" state, and the plain-refusal fallback. The five
copies are not five policies — mechanically, three of them are the same bytes.

Normalizing each function (strip the entity noun, collapse whitespace) and diffing:

- `handleEpicCreated` (`internal/tui/epic_create.go:184`),
  `handleAuditCreated` (`internal/tui/audit_create.go:156`) and
  `handleResearchCreated` (`internal/tui/research_create.go:170`) are
  **byte-identical** — 27 normalized lines each, `diff` clean.
- `handleThreadCreated` (`internal/tui/thread_create.go:174`) differs in exactly
  two lines: the `errors.As` target is `*core.ThreadCreationMutationFailure`
  rather than `*core.CreateFinalizationError`. See `L1`, which is the consequence.
- `handleTaskCreated` (`internal/tui/task_create_commands.go:120`) has **drifted
  structurally** — 35 normalized lines. It hoists the uncertainty test above the
  success test, gives the no-durable-confirmation case its own arm with a
  different message ("Create returned no durable confirmation.") and, unlike the
  other four, does **not** call `reloadAll()` on that arm. Its lock message also
  reads "Inspect the original attempt before doing anything else" where the
  siblings read "Inspect the original <noun> attempt".

That is the point of the finding. This is not a defect report — no user is harmed
today — it is a report that a load-bearing write-outcome policy exists in five
places and the five have already diverged in wording, in control flow and in
whether a reload is issued. The next edit to one of them will not reach the
others, which is exactly how `L1` happened.

The same five-way duplication repeats through the rest of the family and should be
costed with it, not separately:

| Shape | Copies | Divergence |
| --- | --- | --- |
| `handle<Kind>Created` | 5 | 3 identical, 1 type-only, 1 structural |
| `handleKey` saving/locked prologue | 5 | identical modulo receipt accessor; task uses a `copyRecovery()` wrapper |
| `open<Kind>Create` guard prologue | 5 | identical modulo noun and `cur().kind`; task adds parent-epic checks |
| `close()` | 5 | identical modulo task's extra `f.query.Blur()` |
| `updateInput` | 4 | byte-identical modulo the `<kind>CreateInput` predicate |
| `recoveryText()` | 5 | identical modulo noun and ID accessor |
| `view()` | 4 | identical modulo noun, title row, subtitle and receipt accessors |
| `creationRoute()` / `applyCreation()` | 5 | trivial three-liners |
| `creationHost.open<Kind>` | 5 | identical modulo constructor and focus row |

Roughly 350 of the family's ~1,200 non-test lines are mechanical repetition of
this kind.

This is deliberate up to a point, and the audit should say so. The forms' own
comments assert that field definitions, typed params, default policy and receipt
semantics stay per kind — and that part is correct and should not be collapsed
into a universal create schema. `newEpicCreateForm`'s explicit status/priority
defaults genuinely differ from the task form's core-defaulting, and
`submitThreadCreate` genuinely must not offer members or `DryRun`. None of that
belongs in a shared helper.

What is duplicated is not entity policy; it is the post-write state machine, which
is identical by construction because it reads the same four receipt facts
(`Committed`, `Local.CommittedPath`, `DryRun`, the minted ID) plus one error
classification. Completed task
`planning/tasks/6gpb4y5yqmjp-harden-the-shared-tui-creation-host-and-interaction-contracts.md`
drew its boundary at "no universal creation port, form-schema registry, generic
receipt" when there were **two** forms, and closed with "Any genuinely wider
finding needs a scoped follow-up rather than silently expanding this task." There
are now five. This finding is that follow-up, not a reversal of that decision.

One implementation constraint worth recording, because it rules out the obvious
shape: these are methods on `*Model`, and a Go method cannot introduce a type
parameter beyond its receiver's. The consolidation therefore has to be a generic
free function taking the model plus a small per-kind accessor set, not a generic
method and not an interface returning `any`.

**Failing scenario:** none claimed; this is a reduction finding, not a defect.
The drift it predicts is already observable: `handleTaskCreated` does not issue
the `reloadAll()` that its four siblings issue on the equivalent arm, and nothing
in the suite objects to the two behaviours coexisting.

**Why tests didn't catch it:** the suite asserts each kind's receipt policy
independently and deliberately — task 6gpb4y5yqmjp lists "independent durable
assertions" as an acceptance criterion. Independent assertions prove each copy
works; they cannot notice that five copies drifted apart, because no test
compares them.

**Tightening (adjacent):** `epicCreateForm.dirty()`
(`internal/tui/epic_create.go:81`) hand-enumerates its five inputs where
`researchCreateForm.dirty()` and `threadCreateForm.dirty()` loop over
`inputs[:cancel]`; a sixth epic field would not register as dirty. Unify on the
loop form while in the file.

**Follow-up:** whether the five `creationRoute`/`applyCreation` pairs and the five
`creationHost.open<Kind>` methods should collapse into the same table as `M2` is a
design call on the host's seam, not part of this extraction.

**Recommendation:** Extract the shared post-write policy into one generic free function over the receipt (a method cannot add its own type parameter), keeping per-kind params, receipts and field definitions where they are.

#### M2. Creatable-entity capability is spelled out in five parallel switches that nothing keeps in agreement · **Status:** open

**File:** internal/tui/creation_host.go:35 | **Component:** tui/creation-host
**Effort:** M · **Urgency:** eventually

`creationHost` is the right shape — a value-owned host with an active-kind index
that delegates to concrete per-kind slots is the standard Bubble Tea composition,
and the host correctly avoids retaining a mutable draft pointer across `Model`
copies. The finding is not about the design; it is that the set of creatable
entity kinds is written out longhand five separate times inside it, as `switch`
statements, with nothing that keeps the five lists in agreement.

In `internal/tui/creation_host.go` alone:

- `pageCreationAction` (`:35`) — switch over 5 kinds returning the palette word,
  title and filter string.
- `openPageCreation` (`:75`) — switch over the same 5 kinds returning the opener.
- `runCreationAction` (`:93`) — switch over the 5 palette *words* (string keys
  this time) returning the same openers.
- `draft` (`:115`) — switch over the same 5 kinds returning the form slot.
- `begin` (`:142`) — a bare `case entityTasks, entityEpics, entityAudits,
  entityResearch, entityThreads:` whitelist with `default: return false`.

Plus `creationHost.open<Kind>` ×5 (`:156`–`:200`) and a `Model.open<Kind>Create`
guard prologue in each of the five form files. Adding a sixth creatable noun —
which epic `28-first-class-entities-new-planning-nouns` is actively doing — means
nine coordinated edits across six files.

The lens's test is whether a table expresses this better than a switch, and here
it does, because the five switches are keyed on the same enum and return
components of one record. A single `map[entityKind]creationKindSpec{word, title,
filter, slot, open}` would make `pageCreationAction`, `openPageCreation`, `draft`
and `begin` into lookups, and `runCreationAction` a reverse lookup on `word`.
That is fewer branches, one place to add a noun, and no loss of per-kind policy:
the spec's `open` field still points at the kind's own constructor and focus row.

**Failing scenario:** no live defect — all five lists agree today. What the
structure guarantees is that a future mismatch fails *silently* rather than
loudly, and that is worth recording concretely. If a sixth kind is added to
`pageCreationAction` but missed in `begin`'s whitelist, pressing `c` on that page
calls `openPageCreation` → `m.open<Kind>Create()` → `h.open<Kind>()` →
`if !h.begin(kind) { return nil }`, which returns `nil` with **no flash and no
error**. The key simply does nothing, the palette still advertises the action, and
no test fails, because the omission is a missing `case` rather than a wrong value.
The compiler cannot help: `begin` has a `default:` arm, and `draft` returns a nil
interface for an unlisted kind, which `draft`'s own nil check then reads as
"inactive".

**Why tests didn't catch it:** `creation_host_test.go` exercises the four kinds
that exist, which is the correct scope for the code as written. There is no
exhaustiveness assertion tying `entityKind`'s creatable members to the five
lists, so the tests can only ever cover the kinds someone remembered to add.

**Tightening (adjacent):** the `default:` arm in `openPageCreation` (`:87`) sets a
"creation is not available on this page yet" flash, while `runCreationAction`
(`:112`) and `begin` (`:149`) fall through to a bare `nil`/`false` with no
diagnostic. If the table lands, route all three refusals through the one flash so
an unlisted kind is visibly unsupported rather than inert.

**Follow-up:** an `exhaustive`-style lint rule over `entityKind` switches would
catch this class repo-wide rather than only here, but enabling a new linter is a
`.golangci.yml` policy change and belongs to the architecture audit, not this one.

**Recommendation:** Replace the five switches with one table keyed by entityKind holding the palette word/title/filter, the form slot accessor and the open func, and derive begin()'s whitelist from the table's keys.

#### L1. Thread creation is the one create path whose cleanup failure is untyped, so callers cannot classify it · **Status:** open

**File:** internal/store/threadcreation.go:45 | **Component:** store/thread-creation
**Effort:** S · **Urgency:** soon

`core.CreateFinalizationError` exists specifically so a caller can tell a failed
create *cleanup* apart from an ordinary refusal. Its doc comment
(`internal/core/creation_receipt.go:5`) is explicit:

> A caller must not interpret a conflict elsewhere in the error chain as a safe ID
> collision to retry: the adapter could not finish the original create attempt
> cleanly.

`NewResearch` honours that exactly (`internal/core/service_research.go:97`), and so
do four of the five TUI creation controllers, which test
`errors.As(msg.err, &finalization)` before deciding whether a form may be
resubmitted.

Thread creation cannot participate, because the typed error is never produced on
its path. Of the nine guard-release sites in `internal/store`, exactly one —
`create.go:103`, the shared `createEntityFile` used by task/epic/audit/research —
wraps the release failure in `&core.CreateFinalizationError{...}`. The Thread
path has its own guarded writer, and `threadcreation.go:45` emits a plain
`fmt.Errorf("release repository Thread creation guard: %w", releaseErr)` instead.

The visible asymmetry is in `handleThreadCreated`
(`internal/tui/thread_create.go:188`), which is why `M1` and this finding are the
same story told twice. It tests `errors.As(msg.err, &failure)` for
`*core.ThreadCreationMutationFailure`, and `runThreadCreationMutation`
(`internal/core/service_thread.go:117`) produces that wrapper **only when
`result.Committed` is true**. So for an uncommitted attempt whose cleanup failed,
no typed marker reaches the controller at all.

**Failing scenario:** a Thread create whose plan is refused and whose guard
release then fails. `MutateThreadCreation` returns, e.g.,
`errors.Join(fmt.Errorf("...retry: %w", domain.ErrConflict), releaseGuardErr)`
with `Committed == false`. In `handleThreadCreated`:
`errors.As(err, &*ThreadCreationMutationFailure)` is false (not committed);
`creationKnownRefusal(err)` is **true**, because `errors.Is` on a joined error
matches if *any* branch matches, and the `ErrConflict` branch does; therefore
`uncertain` is false; and with `Committed` false and `CommittedPath` empty the
whole guard is false. The form falls through to `f.err = domain.Reason(msg.err)`
and stays editable, inviting the resubmission the contract says must not be
offered. The four sibling forms, given the same shape of error, set `locked` and
show "inspect the original attempt".

This is reported as Low rather than Medium deliberately, because the harm stops
there and the audit should not overstate it:

- `Committed == false` with an empty `CommittedPath` means the write provably did
  not land — `writeNewFileUnlocked` is the final step and sets `Committed` only
  on success — so the offered retry cannot duplicate a record.
- The lock is `flock(LOCK_EX)` on the repo root
  (`internal/store/lock_unix.go:21`), released by closing the descriptor and by
  process exit. A failed `LOCK_UN` therefore does not strand a lock that would
  block or corrupt the retry.

So the live cost is a misleading recovery message on one of five forms, not a
durability bug. The reason to fix it anyway is `M1`'s reason: the typed-marker
contract is documented, four paths implement it, one silently does not, and the
gap is invisible at the call site. If Thread creation ever acquires an
ID-mint-on-collision retry of the kind `NewResearch` already has, that loop would
inherit precisely the unsafe retry the comment at
`internal/core/creation_receipt.go:5` was written to prevent.

**Why tests didn't catch it:** the Thread creation tests cover the committed
cleanup-failure case, which is the dangerous one and is correctly wrapped in
`*ThreadCreationMutationFailure`. No test constructs the *uncommitted* release
failure, and no test asserts that the five create paths classify an equivalent
error chain the same way — the cross-kind comparison `M1` describes as missing.

**Tightening (adjacent):** `threadcreation.go:46-50` hand-rolls
`if err == nil { err = wrapped } else { err = errors.Join(err, wrapped) }`;
`errors.Join` already drops nils, so this is the one-line
`err = errors.Join(err, wrapped)` that `create.go`, `fix.go` and `rename.go` use.

**Follow-up:** the other seven untyped release sites (graph, lifecycle, rename,
threadapply, threadmutation, fix, graphrepair) are mutation paths with their own
typed receipts and are out of scope here; whether they need an analogous marker is
a separate question for the correctness-and-errors lens.

**Recommendation:** Wrap the Thread creation guard-release failure in &core.CreateFinalizationError{...} as store/create.go:103 already does, and check that type in handleThreadCreated alongside *ThreadCreationMutationFailure.

#### L2. max1 is a hand-written clamp for the max builtin the same package already uses 118 times · **Status:** open

**File:** internal/tui/view.go:278 | **Component:** tui/layout
**Effort:** XS · **Urgency:** eventually

`internal/tui/view.go:278` defines:

```go
func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
```

That is `max(n, 1)`. The builtin has been available since Go 1.21, `go.mod`
targets 1.26.0, and `internal/tui` already calls `max`/`min` 118 times in
non-test code — including three lines above the first `max1` call, where
`recomputeLayout` writes `min(max(listOuterW, minListOuterW), maxListOuterW)`
(`:52`) and then `max1(listOuterW - m.st.paneHFrame)` (`:55`). The package uses
two idioms for the same clamp within four lines.

21 call sites, all in `internal/tui`: `view.go` (7), `atlas.go` (7),
`dashboard.go` (3), `item.go` (2), `detail.go` (2).

The lens rejects line-shaving that costs a useful name, so the question is
whether `max1` is one. It is not: `max1(outerW - m.st.paneHFrame)` hides its own
bound, while `max(outerW - m.st.paneHFrame, 1)` states it, and the "1" is the
load-bearing part — every one of these call sites exists to keep lipgloss from
receiving a negative frame dimension. The replacement is shorter, more explicit,
and consistent with the surrounding code.

**Failing scenario:** none — `max1(n)` and `max(n, 1)` agree on every `int`,
including `math.MinInt`. This is a pure consistency finding with no behaviour
change.

**Why tests didn't catch it:** nothing to catch; the function is correct.

**Tightening (adjacent):** `recomputeLayout`'s `if bodyH < 4 { bodyH = 4 }`
(`internal/tui/view.go:27`) is the same hand-rolled clamp and becomes
`bodyH = max(bodyH, 4)` in the same pass.

**Recommendation:** Delete max1 and replace its 21 call sites with max(n, 1).

#### L3. Five listRow literals are positional in an all-string struct, so a field reorder would silently miswire sort keys · **Status:** open

**File:** internal/core/service_list_query.go:26 | **Component:** core/list-query
**Effort:** XS · **Urgency:** eventually

`internal/core/service_list_query.go` is otherwise the cleanest file read this
run — six `Query*` methods that each delegate to the shared generic `queryList`,
with only the irreducible typed glue left per method. One detail is worth an XS
fix.

`listRow` is `struct{ id, name, created, updated, text string }`
(`internal/core/list_query.go:60`) — five fields, all the same type. Five of the
six row builders construct it positionally:

```go
return listRow{r.Source.ID, t.Slug, t.Created, t.Updated, strings.Join(...)}
```

`QueryFindingPage` (`:145`) already uses the named form,
`listRow{id: ..., name: ..., text: ...}`, so the file is internally inconsistent
about it.

`created` and `updated` are the sort keys `validateListQuery` accepts
(`internal/core/list_query.go:82`), and `queryList` sorts on them.

**Failing scenario:** conditional on a future edit, which is why this is Low and
not Medium — the five literals are correct today. Because every field is
`string`, swapping the `created` and `updated` fields in the `listRow` definition
compiles cleanly and silently inverts both sort keys at all five positional call
sites, while `QueryFindingPage` is unaffected because it names its fields. The
observable result is `task list --sort -created` ordering by `updated`, with no
compile error and no panic — ordering regressions of that shape are easy to miss
in review and are not what the existing tests assert hardest.

**Why tests didn't catch it:** nothing to catch yet. The hazard is latent, and a
reorder is the kind of change a reviewer waves through precisely because the
compiler stays quiet.

**Tightening (adjacent):** `QueryEpics` (`:44`) and `QueryThreadViews` (`:110`)
hand-roll the same `status == "" || value == status` filter loop over different
element types; a small generic filter helper would cover both, though the saving
is five lines and does not on its own justify the churn.

**Recommendation:** Give the five positional listRow literals field names, matching QueryFindingPage's existing named form.

#### L4. footer() builds a detail hint line it then throws away in three of four cases · **Status:** open

**File:** internal/tui/view.go:517 | **Component:** tui/render
**Effort:** XS · **Urgency:** eventually

In `footer()`, the detail-focus branch assigns `hints` and then immediately
re-assigns it from a `switch` whose three arms each rebuild the whole line:

```go
if m.focus == focusDetail {
	hints = strings.Join([]string{ ..., m.detailFooterBody(), ... }, " · ")
	switch {
	case m.detail.immersive():
		hints = strings.Join([]string{ ..., m.detailFooterBody(), ... }, " · ")
	case m.zoom:
		hints = strings.Join([]string{ ..., m.detailFooterBody(), ... }, " · ")
	case !m.twoPane:
		hints = strings.Join([]string{ ..., m.detailFooterBody(), ... }, " · ")
	}
}
```

The first assignment is the switch's default arm written before the switch, so it
is discarded whenever any of the three cases matches. `detailFooterBody`
(`internal/tui/view.go:362`) is not trivial — it builds an 8-capacity slice
through nine conditionals and several `m.detail` capability probes — and it runs
twice per frame on the immersive, zoom and single-pane paths, once for a line
that is thrown away.

Moving the literal into `default:` makes the exclusivity explicit, runs
`detailFooterBody` once, and removes one `strings.Join`. `View` is called on
every frame, so this is also the rare simplification that is marginally cheaper.

**Failing scenario:** none — the rendered output is identical; the discarded line
is never displayed. This is wasted work and a misleading shape, not a defect.

**Why tests didn't catch it:** the render tests assert the displayed footer, which
is correct in all four cases. A dead intermediate assignment is invisible to them,
and `go vet` does not flag a conditionally-overwritten variable.

**Follow-up (not a recommendation):** the two atlas footer lines (`:453` and
`:462`) share eight of their tokens and look like an obvious candidate for a
shared tail, but their *order* differs — the work screen renders
`View, Sort, SortRev, tabs, Atlas, …` and the cards screen
`View, tabs, Atlas, Sort, SortRev, …`. Factoring a common tail would reorder
displayed keys on one of the two screens, which is a visible UX change and
disqualified under this lens's Risk rule. Either deliberately unify the order as a
UX decision or leave both literals alone; do not fold this into a mechanical pass.

**Recommendation:** Fold the pre-switch assignment into the switch as its default: arm so detailFooterBody() and its Join run once.

## Punch list

Order: Critical → High → Medium → Low. No Critical or High findings this run.

- `M1`. Five creation controllers duplicate one receipt-interpretation policy, and it has already drifted  (effort: M · urgency: soon)
- `M2`. Creatable-entity capability is spelled out in five parallel switches that nothing keeps in agreement  (effort: M · urgency: eventually)
- `L1`. Thread creation is the one create path whose cleanup failure is untyped, so callers cannot classify it  (effort: S · urgency: soon)
- `L2`. max1 is a hand-written clamp for the max builtin the same package already uses 118 times  (effort: XS · urgency: eventually)
- `L3`. Five listRow literals are positional in an all-string struct, so a field reorder would silently miswire sort keys  (effort: XS · urgency: eventually)
- `L4`. footer() builds a detail hint line it then throws away in three of four cases  (effort: XS · urgency: eventually)

**Nothing here is a live defect.** `M1`, `M2`, `L2` and `L4` are reductions;
`L1` and `L3` are latent hazards whose live cost today is a misleading message and
nothing respectively. That is the honest result of this lens, and it matches the
last simplification run's: the question is "can this be simpler", not "is this
broken". `L1` is the closest to a defect and is scored Low on purpose — the body
says why, including the two facts that bound its harm.

Routine: `code-quality-audit` · lens `simplification` · ISO week `2026-W47`,
slot `Fri` (index 5 = (47 × 2 + 1) mod 6).

## Files audited

The lens's surface is "the two largest packages by line count, plus whatever the
churn signal surfaces". Measured today (non-test Go lines) those are
`internal/tui` (16,478) and `internal/core` (13,824); `internal/cli` (7,950),
which the routine spec still names from an older measurement, is third. The
measurement is unchanged from the 2026-10-30 run.

30-day churn inside that surface is overwhelmingly concentrated in one feature
family: all 20 of the top-churned non-test files in the surface are in
`internal/tui`, and 10 of them are the staged-creation files (`task_create.go`, `staged_creation.go`,
`epic_create.go`, `creation_host.go`, `thread_create.go`, `research_create.go`,
`audit_create.go` and their view/command partners). Signal =
log(churn₃₀d + 1) × log(blast_radius + 1): `tui/task_create.go` 10.3,
`tui/view.go` 8.9, `tui/epic_create.go` 8.8, `tui/staged_creation.go` 8.5,
`tui/creation_host.go` 7.5.

- **Signal**: `internal/tui/task_create.go` (278 lines) — top score in the surface; the original form the other four were derived from.
- **Signal**: `internal/tui/view.go` (579 lines) — highest raw churn in the surface (5 commits/30d) and the package's render/layout half.
- **Adjacency** (from `task_create.go`): `internal/tui/creation_host.go` (218 lines) — the shared host the form plugs into; the hop from a per-kind controller to the seam that owns its lifetime.
- **Adjacency** (from `task_create.go`): `internal/tui/epic_create.go` (243 lines) — the sibling near-duplicate, chosen to measure whether the derived forms diverged.
- **Random**: `internal/core/service_list_query.go` (148 lines) — drawn from the 132-file eligible pool (≥50 lines, non-test, non-generated, non-fixture).

Read in full as supporting context rather than as picks:
`internal/tui/task_create_commands.go`, `internal/tui/staged_creation.go`,
`internal/tui/audit_create.go`, `internal/tui/research_create.go`,
`internal/tui/thread_create.go`, `internal/store/threadcreation.go`,
`internal/store/create.go`, `internal/core/creation_receipt.go`,
`internal/core/service_thread.go`, `internal/core/list_query.go`,
`internal/store/lock_unix.go`, `internal/cli/creation_failure.go`.

Two notes on method, both limiting:

- The spec's "sibling untouched for 6+ months" adjacency option was unavailable.
  This repository's git history is 199 commits deep and its oldest file touch is
  2026-10-06, so no file in the tree qualifies. The stale-neighbour hop cannot be
  exercised here until the history is older than the window; the sibling-duplicate
  hop was used instead.
- `internal/tui` is the largest package and also where this run spent almost all
  its attention. `internal/core` was reached only through the random pick and
  through following `L1`'s error chain. That is this run's coverage gap, not a
  statement that `core` is clean — and it is the complement of the 2026-10-30 run,
  which read no `internal/tui` file at all.

## Commands run

| Command | Result |
| --- | --- |
| `go build -o bin/tskflwctl ./cmd/tskflwctl` | ok — `just` is not installed in this container, so the spec's documented fallback was used |
| `go test -race ./...` | 27 packages ok, **2 FAIL — both pre-existing and environmental** (detail below) |
| `golangci-lint run ./...` | **not run** — refuses: "the Go language version (go1.25) used to build golangci-lint is lower than the targeted Go version (1.26.0)". The installed binary is 2.5.0 built with go1.25.1; `go.mod` targets 1.26.0. Unchanged since the 2026-10-30 run |
| `./bin/tskflwctl -C . lint` | ok — "all planning entities and dependency links pass lint" |
| `./bin/tskflwctl -C . audit lint 2026-11-20-simplification` | ok — "all audit findings pass lint" |
| `git status --porcelain` | one path, `planning/audits/6gzgwzwamd0r-2026-11-20-simplification.md` — checked before and after the run |
| `git log --since="30 days ago" --name-only` over the lens surface | the churn ranking in "Files audited" |
| `grep -rln <ExportedName> internal cmd` per candidate | the blast-radius term of the signal score |
| normalized `awk`/`sed`/`diff` of the five `handle<Kind>Created` functions | `M1`'s evidence: three byte-identical, one type-only difference, one structural divergence |
| `grep -rn "releaseErr := unlock()" internal/store/*.go` | `L1`'s evidence: 9 guard-release sites, 1 typed |
| `grep -rn "max1(" internal/` · `grep -rn "\bmax(\|\bmin(" internal/tui/*.go` | `L2`'s counts: 21 call sites against 118 builtin uses |

The audited packages are green under the race detector: `internal/tui` 170.2s,
`internal/domain` 35.5s, `internal/store` 13.4s, `internal/cli` 11.2s,
`internal/wire` 3.8s, `internal/core` 1.7s.

**Pre-existing test failures — two packages, neither touched by this run.** Both
are the same failures the 2026-10-30 simplification audit recorded, reproduced
unchanged:

- `internal/tools/reviewworkspace` — 50 assertions failing with `error: rsync is
  required`. `rsync` is not installed in this container (`which rsync` finds
  nothing).
- `cmd/tskflwctl` — `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior`:
  `filesystem failure = exit 0, want 1`. The test makes a directory unwritable to
  force an I/O failure; this container runs as root, which ignores the permission
  bits, so the write succeeds. **Already tracked**: open task
  `planning/tasks/6gppr2jd4amw-tighten-residual-test-rigour-gaps-from-the-september-and-october-audits.md`
  (`ready-to-start`) carries exactly this as its first acceptance criterion —
  "injects a filesystem failure that holds for every uid, or skips with a reason
  when the chmod does not bite".

Per the routine spec these were red before this run and no attempt was made to fix
them.

## What audited clean

Four of the five picks produced no Medium-or-higher finding, and two produced
nothing at all. Recording why, since "clean" is a result:

- **`internal/tui/staged_creation.go` (supporting read) — clean, and the reason
  `M1` is scoped the way it is.** The genuinely shared mechanics are *already*
  extracted here and extracted well: `creationState` holds the orthogonal
  active/saving/locked/displaced axes, `creationControls` borrows a form's widget
  array for one synchronous call rather than retaining it, `updateCreationInput`
  defends against the `textinput.Model` shared-rune-buffer trap with a comment
  explaining why, and `routeCreationWidget` captures only routing tokens so a
  delayed clipboard command cannot retain a `*core.Service`. This is careful code.
  `M1` is not "nothing was factored out" — it is "the post-write half was not".

- **`internal/tui/creation_host.go` — the host's design is right; only its
  longhand is reportable (`M2`).** The value-owned slots with a synchronous
  `draft()` borrow are the correct answer to aliasing mutable form state across
  `Model` copies, and the comment says so. `applyResult`'s four-way match on
  kind/gen/session/saving before dispatch is the right guard and is not
  duplicated anywhere.

- **The `errors.As` disjunct in the four non-Thread receipt handlers — checked for
  redundancy, and it is load-bearing.** `uncertain := errors.As(err,
  &finalization) || (err != nil && !creationKnownRefusal(err))` looks like the
  first term can never decide anything the second does not. It can: the reachable
  case is `store/create.go:103` joining an `ErrConflict`-classified refusal with a
  `*CreateFinalizationError` from the failed guard release. `creationKnownRefusal`
  then returns true, the second disjunct is false, and only `errors.As` keeps the
  form locked. Deleting it would be a regression. This is exactly the protection
  `L1` shows the Thread path lacks.

- **`internal/core/service_list_query.go` (the random pick) — clean apart from the
  XS `L3`.** Six `Query*` methods, each reduced to validate → list → `queryList`
  with a per-type row projection. The shared machinery is already a generic
  function, and what is left per method is the typed glue Go cannot abstract
  without method-level type parameters. There is no reduction available here; the
  one note is a literal style that the file itself is already inconsistent about.

- **`internal/tui/task_create.go` — no reportable finding of its own.** Its picker
  mode, epic-catalog generation counter and `enumValue` hook for "default" → `""`
  are real per-kind policy, correctly kept out of the shared host. Its
  contribution to this audit is as `M1`'s drifted copy, which is reported against
  the controller file, not here.

- **`internal/tui/view.go` — the layout half is clean; `L2` and `L4` are both
  surface.** `recomputeLayout` subtracts borders before sizing children and
  reserves the `bubbles/list` pagination line, with comments naming both traps.
  The modal registry in `bodyView` is already table-driven — "a new overlay needs
  no case added here" — which is precisely the shape `M2` asks `creation_host.go`
  to adopt, and is evidence the codebase already knows the pattern.

## External research

Two Medium findings, so under the routine's threshold the lens got light
research. Three queries; all four points below changed or constrained a
recommendation rather than decorating it.

- **A Go method cannot introduce a type parameter beyond its receiver's.** This
  rules out the obvious shape for `M1` — the five handlers are methods on
  `*Model`, so the consolidation has to be a generic *free function* plus a small
  per-kind accessor set, not a generic method and not an interface returning
  `any`. The proposal to allow method-level type parameters (golang/go#77273) is
  accepted but was not shipped as of early 2026, with Go 1.27 discussed as a
  possible target; nothing should be designed against it yet.
  <https://dev.to/gabrielanhaia/go-generics-in-2026-what-finally-works-and-what-still-doesnt-3i31>,
  <https://go.googlesource.com/website/+show/HEAD/_content/blog/generic-methods.md>
- **The threshold guidance argues for `M1` and against over-reach.** The common
  rule of thumb is to reach for generics when the type parameter removes a type
  assertion and the constraint stays narrow, and to skip them when the constraint
  becomes a paragraph or the duplication is "two five-line functions". Five
  27-to-35-line policy functions reading the same four receipt fields clear that
  bar; the per-kind field definitions and param conversion do not, which is why
  `M1` explicitly leaves them alone.
  <https://dev.to/gabrielanhaia/go-generics-4-years-in-the-3-cases-where-theyre-the-right-answer-5bip>
- **The host's existing shape matches the recommended Bubble Tea composition.** A
  parent model embedding child models with an active-index and delegated
  `Update`/`View` is the documented pattern, and it is said to scale to many
  embedded models "just keep track of which one is active". That is what
  `creationHost` does, so `M2` is narrowed to replacing the five hand-written
  switches with a table — not to restructuring the host.
  <https://charmbracelet.pages.dev/en/bubbletea/tutorials/composable-views>
- **`errors.Is`/`As` on a joined error match if *any* branch matches**, and
  `Unwrap` does not traverse a `Join` at all. This is the mechanism behind `L1`:
  a joined (refusal + cleanup-failure) error satisfies `creationKnownRefusal`
  because the refusal branch does, so a sentinel test alone cannot see the
  cleanup failure. The standard mitigation is exactly what
  `core.CreateFinalizationError` is — a typed marker pulled out with `errors.As`
  — which makes `L1` a gap in an existing correct design rather than a new one.
  <https://pkg.go.dev/errors>,
  <https://dev.to/gabrielanhaia/errorsjoin-vs-multi-return-when-to-aggregate-when-to-wrap-3b5o>

No antipattern search produced a finding the code actually commits.

## Related-task observations (propose-only)

The step-10 cross-reference found **no open task owning any finding in this
audit**, so no finding was marked tracked and no task was annotated. One
observation is worth a human's attention:

- **The two natural owners are both completed, and one of them explicitly asked
  for this follow-up.**
  `planning/tasks/6gpb4y5yqmjp-harden-the-shared-tui-creation-host-and-interaction-contracts.md`
  (`completed`) drew its Boundaries at "no universal creation port, form-schema
  registry, generic receipt" and closed with "Any genuinely wider finding needs a
  scoped follow-up rather than silently expanding this task". It made that call
  with two forms in the tree; there are now five.
  `planning/tasks/6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md`
  (`completed`) added the other three and was scoped to "without pretending their
  required inputs or mutation policies are identical" — which remains the right
  constraint and is not what `M1` proposes relaxing. Neither should be reopened;
  `M1` and `M2` are the scoped successors those tasks invited.
- No task status, tier, priority, scope or `depends_on` was changed by this run,
  and no judgment call here is acted on.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — `tskflwctl task new "Consolidate the TUI creation post-write receipt policy into one shared path" --epic 21-code-quality-architecture-hardening --tags tui,authoring,maintainability --tier 3 --priority medium --description "Extract the five handle<Kind>Created controllers' shared displacement/committed/uncertain/locked policy into one generic free function, keeping per-kind params and receipts; reconcile the drift in handleTaskCreated."`
- ○ M2 · open — `tskflwctl task new "Table-drive the creatable entity kinds in creationHost" --epic 21-code-quality-architecture-hardening --tags tui,authoring,maintainability --tier 3 --priority low --description "Replace the five parallel entityKind switches in creation_host.go with one spec table so adding a creatable noun is a single edit and an omission is not a silent no-op."`
- ○ L1 · open — `tskflwctl task new "Type the Thread creation guard-release failure as CreateFinalizationError" --epic 21-code-quality-architecture-hardening --tags store,threads,errors --tier 3 --priority medium --description "Wrap threadcreation.go's release failure in core.CreateFinalizationError and classify it in handleThreadCreated so all five create paths treat a cleanup failure alike."`
- ○ L2 · open — `tskflwctl task new "Mechanical TUI simplification pass: drop max1 and the dead footer hint line" --epic 21-code-quality-architecture-hardening --tags tui,render,maintainability --tier 4 --priority low --description "Delete max1 in favour of the max builtin across its 21 call sites and fold footer()'s pre-switch detail hint into the switch default arm."`
- ○ L4 · open — Same task as `L2` — mechanical TUI simplification pass
- ○ L3 · open — `tskflwctl task new "Name the fields in the listRow literals" --epic 21-code-quality-architecture-hardening --tags core,maintainability --tier 4 --priority low --description "Give the five positional all-string listRow literals field names so a struct field reorder cannot silently miswire the created/updated sort keys."`
