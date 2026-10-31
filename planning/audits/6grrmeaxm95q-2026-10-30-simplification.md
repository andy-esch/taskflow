---
schema: 1
id: 6grrmeaxm95q
bucket: open
area: simplification
date: "2026-10-30"
updated_at: "2026-10-30"
---

# Audit: simplification — 2026-10-30

> Create findings with `audit finding new`; update them with `audit finding`.

Routine: `code-quality-audit` · lens `simplification` · ISO week `2026-W44`,
slot `Fri` (index 5 = (44 × 2 + 1) mod 6).

The single question this lens asks: **can this be simpler without losing
functionality and without resorting to tricks?** Two of the four findings below
are about surface that outlived its consumer; one is about a layer that never had
one; one is about reductions the toolchain will perform on request. None is a
defect, and none is reported as one.

## Punch list

- `M1`. Transitional task-graph compatibility surface is unreachable in production  (effort: M · urgency: eventually)
- `M2`. go fix modernizers find mechanical simplifications the lint gate never asks for  (effort: S · urgency: soon)
- `L1`. Private setCriteriaState is a zero-value pass-through layer  (effort: XS · urgency: eventually)
- `L2`. Three exported FS read methods are test-only, and one cites a consumer it lost  (effort: XS · urgency: eventually)

## Files audited

The lens's surface is "the two largest packages by line count, plus whatever the
churn signal surfaces". Measured today (non-test Go lines) those are
`internal/tui` (16,478) and `internal/core` (13,824) — not `internal/cli` (7,950),
which the routine's spec still names from an earlier measurement.

Signal = log(churn₃₀d + 1) × log(blast_radius + 1), computed within that surface:
`core/store.go` 12.6, `core/service_task.go` 11.5, `tui/entity.go` 11.4,
`tui/model.go` 10.1, `tui/view.go` 9.5, `tui/edit.go` 9.5.

- **Signal**: `internal/core/store.go` (479 lines) — highest churn in the surface (14 commits/30d) and the consumer-owned port file every adapter is measured against.
- **Signal**: `internal/core/service_task.go` (1,047 lines) — second-highest signal; the largest single file in `core` and the busiest write path (9 commits/30d).
- **Adjacency** (from `core/store.go`): `internal/store/fsstore.go` (443 lines) — the hexagonal-seam hop the lens prefers: the one production type asserted against those ports.
- **Adjacency** (from `core/service_task.go`): `internal/store/edit.go` (329 lines) — cross-seam to the store side of the editor/CAS write path `service_task.go` orchestrates.
- **Random**: `internal/core/selected_entity_path.go` (72 lines) — drawn from the 86-file eligible pool (≥50 lines, non-test, non-generated, non-fixture).

`tui/entity.go` scored within noise of the second signal pick (11.41 vs 11.46) and
lost on the tie-break to `core`'s structurally higher blast radius, so **no
`internal/tui` file was read this run** despite it being the largest package. That
is a gap in this run's coverage, not a statement that the TUI is clean.

## Commands run

| Command | Result |
| --- | --- |
| `go build -o bin/tskflwctl ./cmd/tskflwctl` | ok — `just` is not installed in this container, so the spec's documented fallback was used |
| `go test -race ./...` | **FAIL, pre-existing and environmental** — see below |
| `golangci-lint run ./...` | **not run** — the installed binary is built with Go 1.25 while `go.mod` targets 1.26.0, which `golangci-lint` refuses. The `Justfile` already warns that this recipe "needs a v2.x binary BUILT WITH Go ≥ go.mod's target". |
| `./bin/tskflwctl -C . lint` | ok — "all planning entities and dependency links pass lint" |
| `./bin/tskflwctl -C . audit lint 2026-10-30-simplification` | ok (see PR) |
| `go fix -diff ./internal/core/... ./internal/store/...` | 906-line diff across 10 production + 13 test files; **no files written** (`git status` confirmed clean after) — this is M2's evidence |
| `go tool fix help` | enumerated the 22 registered modernizers quoted in M2 |

**Pre-existing test failures (two packages, both environmental, neither touched by
this run):**

- `internal/tools/reviewworkspace` — 22 failures, every one `create failed: exit
  status 2 / error: rsync is required`. `rsync` is not installed in this container.
- `cmd/tskflwctl` — `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior`:
  `filesystem failure = exit 0, want 1`. The test makes a directory unwritable to
  force an I/O failure; this container runs as `root`, which ignores the
  permission bits, so the write succeeds. This is **already tracked** — open task
  `tighten-residual-test-rigour-gaps-from-the-september-and-october-audits` carries
  "Make the smoke test's filesystem probe honest as root" as its first item.

Everything else was green, including `internal/core` (1.9s), `internal/store`
(12.8s), `internal/tui` (199.8s) and `internal/domain` (38.1s) under `-race`.

## Findings


#### M1. Transitional task-graph compatibility surface is unreachable in production · **Status:** open

**File:** internal/core/service_task.go:210 | **Component:** core/task-graph-read
**Effort:** M · **Urgency:** eventually

`TaskGraphRead` carries three mutually-redundant representations of the same task
set — `GuardedRecords`, `Records`, and `Tasks` — and in production exactly one of
them is ever populated.

**Reachability, as verified:**

- `*store.FS` is the only type asserted as `core.Store` (`internal/store/fsstore.go:85`),
  and it also asserts `core.TaskGraphSource` (`:88`). Its `ReadTaskGraph`
  (`internal/store/fsstore.go:194`) populates `GuardedRecords` and `Problems` only.
- The `taskStoreGraphSource` fallback (`internal/core/service.go:358`) is therefore
  unreachable in production: it is selected only for a `Store` that does *not*
  implement `TaskGraphSource`, and no such production store exists. It is the only
  producer of `Records`.
- `Tasks` has no producer at all outside tests. The literal `TaskGraphRead{... Tasks: ...}`
  appears at four test sites (`internal/tui/thread_projection_test.go:186`,
  `internal/core/dependency_repair_test.go:123`, `internal/core/space_overview_test.go:87`,
  and via `NewTaskGraph`).
- `NewTaskGraph` (`internal/core/dependency_graph.go:382`) — the wrapper over
  `TaskGraphReadFromFiles` — has **zero** non-test callers; ~100 test call sites.
  `TaskGraphReadFromFiles`' only non-test caller is `NewTaskGraph` itself, so it is
  transitively test-only too.

**What the redundancy costs:** a three-branch fallback cascade in `taskGraphRecords`
(`internal/core/service_task.go:210-228`), the `taskGraphGuardedRecords` re-wrapping
layer (`:230`), `taskGraphTasks` on top of that (`:242`), and a defensive
`read.Tasks = nil // compatibility values cannot reintroduce a rejected record`
in `validatedTaskGraphRead` (`:274`). Two of the three branches are dead weight that
every reader of this load-bearing snapshot type must still reason about, and the
type's own doc comment describes `Records` and `Tasks` as "transitional read-only
compatibility projections" without saying that the transition is already complete.

**This is drift from a decision the project already closed.** Task
`6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities`
(status: `completed`) classifies these exact three symbols in its crossing
inventory: "`taskStoreGraphSource`, `TaskGraphReadFromFiles`, and
`NewTaskGraph(...FileProblem)` | Transitional local compatibility | Native sources
return identity-bearing/versioned graph records. **Retain only a clearly named
test/legacy adapter if still needed**; no new core path inference." Task
`6gcwcf88z57p` (also `completed`) separately records a checkpoint review of "the
compatibility `NewTaskGraph`/`TaskGraphRead.Tasks` conversions". Both owning tasks
are closed; the named follow-through was never done and no open task carries it.

This is a smell, not a defect — there is no input that makes current code behave
wrongly, which is why no failing scenario is given below.

**Recommendation:** Delete `TaskGraphRead.Tasks`, `taskStoreGraphSource`, and the `Records`/`Tasks` branches of `taskGraphRecords`; move `NewTaskGraph`/`TaskGraphReadFromFiles` to a clearly named test helper (e.g. `internal/testutil`) as task 6gcwcf7rgxef's inventory prescribed, migrating the ~100 test call sites.

#### M2. go fix modernizers find mechanical simplifications the lint gate never asks for · **Status:** open

**File:** internal/core/service_task.go:895 | **Component:** toolchain/core+store
**Effort:** S · **Urgency:** soon

`go.mod` already targets `go 1.26.0`, so Go 1.26's rewritten `go fix` and its
modernizer analyzers are available on this repo's own declared minimum today.
Run read-only against the two packages this lens audited, the tool proposes
mechanical simplifications in **10 production files**:

```
$ go fix -diff ./internal/core/... ./internal/store/...
internal/core/configuration.go, internal/core/dependency_graph.go,
internal/core/service_research.go, internal/core/service_task.go,
internal/core/thread_apply.go, internal/store/diagnose.go, internal/store/fix.go,
internal/store/frontmatter.go, internal/store/frontmatter_preserve.go,
internal/store/threadapply.go              (plus 13 test files)
```

Two concrete examples from the audited surface:

- `internal/core/service_task.go:895` — `splitList` loops over `strings.Split(s, ",")`;
  the `stringsseq` modernizer rewrites it to `strings.SplitSeq`, dropping the
  intermediate slice allocation.
- `internal/core/thread_apply.go:488` — `sliceContainsExact` is a hand-rolled
  five-line linear scan that the `slicescontains` modernizer replaces with
  `slices.Contains(values, target)`. This is the same shape as
  `taskImpactsNeedRepair` / `threadImpactsNeedRepair`
  (`internal/core/service_task.go:728`, `:737`), two six-line any-predicate loops
  that `slices.ContainsFunc` expresses directly — `slices` is already imported in
  that file.

These are exactly the lens's target: real reduction, no behaviour change, and no
trick — the rewrites come from the toolchain's own analyzers, which `go tool fix`
documents as carrying "fixes that may be safely applied".

**Why the existing lint gate didn't surface this.** `.golangci.yml` runs the
`standard` set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus
`depguard`. None of those modernize, and `unused` additionally cannot find M1's
dead surface: staticcheck's `unused` works per-package, counts a test-only caller
as a live caller, and does not report exported declarations at all. A whole-program
reachability pass (`golang.org/x/tools/cmd/deadcode`, rooted at `./cmd/tskflwctl`)
is the check that *would* have found M1 mechanically — it builds a call graph by
Rapid Type Analysis from `main` entry points, so exported-but-unreached symbols do
get reported. Its documented caveats are worth carrying into any gate: results are
per-GOOS/GOARCH/tags, it is blind to assembly and `go:linkname` aliasing, and its
`-test` flag changes what counts as reachable.

This is a smell, not a defect; no failing scenario is given.

**Recommendation:** Apply `go fix` to the production files it flags (reviewing each hunk, tests separately), then decide whether `go fix -diff --exit-code`-style drift and a `deadcode` reachability pass belong in the lint recipe.

#### L1. Private setCriteriaState is a zero-value pass-through layer · **Status:** open

**File:** internal/core/service_task.go:474 | **Component:** core/criteria-writes
**Effort:** XS · **Urgency:** eventually

The criterion write path runs through four layers where two of them add nothing:

```
(*Service).SetCriterionState   :412  →  SetCriteriaState(slug, []int{n}, …)
(*Service).SetCriteriaState    :422  →  s.setCriteriaState(slug, ns, state, reason, dryRun)
(*Service).setCriteriaState    :474  →  s.transformTaskBody(slug, dryRun, func(body) {…})
(*Service).transformTaskBody   :488     the real retry + CAS + transform path
```

`setCriteriaState` (`:474`) is private, has exactly **one** caller — the exported
method two lines above it at `:423` — and takes an identical parameter list. The
exported method's entire body is `return s.setCriteriaState(slug, ns, state, reason, dryRun)`.
This is the lens's named smell verbatim: a two-level abstraction with one
implementation, where the private layer neither narrows the signature, adds a
guard, nor gives the operation a name the caller does not already have.

For contrast, the two sibling methods on the same path go straight to
`transformTaskBody` with no intermediary: `SetAcceptanceCriterion` (`:401`) and
`EditCriteria` (`:484`). `EditCriteria`'s own doc comment makes the intent
explicit — "naming it here rather than adding three near-identical methods keeps
the write path … in one place" — so the extra hop is inconsistent with the
convention this file already states for itself.

`SetCriterionState` (`:412`) is a different matter and is deliberately **not**
part of this recommendation: it has no caller anywhere in `internal/` or `cmd/`
today, but open task `6g3cp32y9149-let-the-tui-flip-an-acceptance-criterion`
(ready-to-start, tier 2) names it in its implementation constraints as one of the
"existing core write paths" the TUI slice will use. Deleting it would break queued
work. See Related-task observations.

This is a smell, not a defect; no failing scenario is given.

**Recommendation:** Inline `setCriteriaState` into `SetCriteriaState` so the exported method calls `transformTaskBody` directly, matching `SetAcceptanceCriterion` and `EditCriteria`.

#### L2. Three exported FS read methods are test-only, and one cites a consumer it lost · **Status:** open

**File:** internal/store/fsstore.go:214 | **Component:** store/read-surface
**Effort:** XS · **Urgency:** eventually

Three exported read methods on `*store.FS` are on no `core` port and have no
production caller — only tests reach them. Adjacent to M1: the same "compatibility
read surface outliving its consumer" pattern, one layer out in the adapter.

| Method | Port? | Non-test callers | Test callers |
| --- | --- | --- | --- |
| `FS.ListTasks` (`fsstore.go:137`) | no — `TaskStore` declares `ReadTasks` | 0 | ~8 |
| `FS.GetTask` (`fsstore.go:227`) | no — `TaskStore` declares `ReadTask` | 0 | ~12 |
| `FS.ListTasksWithBodies` (`fsstore.go:214`) | no — `LintSource` declares `ReadLintTasks` | 0 | 2 |

`ListTasksWithBodies` additionally carries a justification that is no longer true.
Its doc comment reads: "ListTasks' scan with each task's body kept alongside (one
pass), so lint's acceptance-criteria checks read every file once — the task twin of
ListAuditsWithFindings." Lint does not call it. `FS.ReadLintTasks`
(`internal/store/lintsource.go:12`) projects `scanTaskDocuments` directly, which is
where the one-pass property actually lives. The comment names a consumer that moved
away, so a reader auditing the one-read contract is sent to the wrong method.

Test-support helpers on an adapter are legitimate; three of them indistinguishable
from port methods, one of them citing a consumer it no longer has, is the part
worth fixing. The cheapest honest fix is the comment plus a one-line marker that
these are test conveniences, not the read contract — the deletion question can
ride along with M1's test migration, which touches many of the same call sites.

This is a smell, not a defect; no failing scenario is given.

**Recommendation:** Correct `ListTasksWithBodies`' doc comment to point at `ReadLintTasks`/`scanTaskDocuments`, and say in one line that these three methods are test conveniences rather than the read contract.

## What audited clean

Two of the five files produced nothing worth reporting, and one produced a
candidate I am rejecting rather than filing. That is a real result.

- `internal/store/edit.go` — **clean, and the lens's positive example.** The
  parse-before-accept editor loop is already factored once into the generic
  `editFile[T]` (`:69`) and shared by `EditTask`/`EditAudit`/`EditEpic`/`EditResearch`;
  `acceptEdited[T]` (`:44`) and `asValidation` (`:62`) are each used more than once and
  each carry a comment explaining why they exist. There is no duplication left to
  collapse here. The one thing I looked at twice — `defer unlock()` inside the `for`
  loop at `:124` — is not a leak: every path after the lock acquisition returns, so
  the loop cannot iterate with the lock held. It is a correctness-lens question
  about readability at worst, and belongs to that lens, not this one.
- `internal/store/fsstore.go` — **clean on the lens's own question.** The three
  task-read projections (`ListTasks`, `ListTasksWithBodies`, `ReadTaskGraph`) all
  funnel through one `scanTaskDocuments` (`:158`), which is the non-duplicated
  shape, and `resolve`/`resolvePath` differ by exactly the fuzzy-vs-exact matcher
  with a comment saying so. Its exported-but-test-only read surface is L2; the file
  itself is not the problem.
- `internal/core/store.go` — **a rejected simplification candidate.** It holds five
  near-identical two-method interfaces (`TaskPathSource`, `ThreadPathSource`,
  `EpicPathSource`, `AuditPathSource`, `ResearchPathSource`, `:92`–`:282`) and four
  one-field structs that differ only in name (`TaskMutationTarget`,
  `AuditMutationTarget`, `ResearchMutationTarget`, `EpicMutationTarget`, `:27`–`:42`).
  Collapsing either set would shave real lines — and would delete the type safety
  that stops an epic ID being passed where a task ID is required, which is the
  whole point of the `MutationTarget` design the file documents at length. The lens
  says a change that leans on fewer names at the cost of comprehension is not a
  simplification. **Rejected; not filed as a finding.**
- `internal/core/selected_entity_path.go` — **five 9-line methods that differ in
  three tokens, and I am still not filing it.** `TaskPathByID` / `EpicPathByID` /
  `AuditPathByID` / `ResearchPathByID` / `ThreadPathByID` are mechanically identical
  modulo `(EntityKind, port field, resolve method)`. A shared helper taking a
  resolver closure would cut roughly 25 lines. But open task
  `6g63hjmb3d8s-collapse-the-zero-variance-cross-entity-cli-verbs-behind-the-existing-factory`
  names "Collapsing the `core.Service` per-entity method fan-out
  (`TaskPath`/`ResearchPath`/…) behind a generic seam" in its **Out of scope**
  section — i.e. a decision was already taken to leave this fan-out alone. A fresh
  finding would be re-litigating it. Routed to Related-task observations instead.

## External research

Required by step 11 (fewer than three Medium-or-higher findings). Four queries;
`WebFetch` could not reach `go.dev` or `go.googlesource.com` from this session —
both are outside its allowed network domains — so the Go-tooling claims below rest
on search-result summaries plus **locally verified behaviour** (`go tool fix help`
and `go fix -diff`, both run in this checkout and quoted in M2). The claims that
matter for M2 were confirmed locally, not taken on the web's word.

- **Go 1.26's rewritten `go fix`** (the basis of M2). Alan Donovan's 2026-02-17 Go
  blog post describes a fully rewritten `go fix` whose modernizer analyzers "create
  many opportunities to simplify existing code", and states the motive plainly:
  LLM coding assistants reproduce the idioms dominant in their training corpus even
  when better ones exist, so the project is pushing current idioms into the
  open-source corpus. Verified locally: 22 analyzers registered, `-diff` is
  read-only, and `go.mod` already declares `go 1.26.0`.
  <https://go.dev/blog/gofix>
- **`deadcode` vs staticcheck's `unused`** (why the gate missed M1). `golang/go`
  issue 66073 records that staticcheck's unused analysis "only works on a
  per-package basis and doesn't consider exported declarations", whereas
  `golang.org/x/tools/cmd/deadcode` builds a call graph by Rapid Type Analysis from
  `main` entry points, so exported-but-unreachable symbols are reported. The same
  thread notes an `-exported` flag was proposed and resisted on precision grounds.
  Documented caveats: per-GOOS/GOARCH/tags results, blindness to assembly and
  `go:linkname` aliasing, and `-test` changing reachability.
  <https://golang.org/issue/66073> ·
  <https://go.googlesource.com/website/+show/HEAD/_content/blog/deadcode.md>
- **An antipattern this code does *not* have.** I went looking for writing on
  production types carrying representations that only test fakes populate — which
  is exactly M1's shape — expecting to find it named. It is not, really: the
  nearest established frame is DHH's "test-induced design damage", which
  Thoughtworks summarises as the dispute between Fowler, Beck and DHH over whether
  testability pressure damages design. That literature is about abstraction added
  *for* testing; M1 is the gentler case of compatibility scaffolding *left behind
  after* a migration, with tests as its last remaining caller. Worth recording that
  the pattern is under-named, because it means no linter ships looking for it and a
  reachability pass is the only mechanical defence.
  <https://www.thoughtworks.com/insights/blog/test-induced-design-damage-fallacy-or-reality>
- **Proactive, not a finding:** Go 1.26 also ships `//go:fix inline` directives,
  which rewrite callers of a function marked for inlining. That is the sanctioned
  mechanism for retiring a pass-through like L1's `setCriteriaState` or M1's
  compatibility constructors *with* their call sites, rather than deleting them and
  hand-fixing ~100 test references. Not yet used anywhere in this repo.
  <https://go.dev/blog/gofix>

## Related-task observations (propose-only)

No task status, tier, priority, or scope was changed. These are judgment calls for
the owner.

- **Unowned follow-through** (M1's root cause): `6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities`
  (`completed`) classified `taskStoreGraphSource`, `TaskGraphReadFromFiles` and
  `NewTaskGraph(...FileProblem)` as "Transitional local compatibility" with a named
  target, and `6gcwcf88z57p` (`completed`) recorded a checkpoint review of the same
  conversions. Both are closed; the retirement they prescribed was never scheduled
  and no open task carries it. The design task did its job — the inventory row is
  what let this audit verify the drift in minutes — but a completed design task is
  not a completed migration. **Possible scope conflict with how "completed" was
  applied**, worth the owner's eye.
- **Scope boundary worth revisiting**: `6g63hjmb3d8s-collapse-the-zero-variance-cross-entity-cli-verbs-behind-the-existing-factory`
  (`ready-to-start`) excludes the `core.Service` per-entity path fan-out from its
  scope. That exclusion is defensible for the *reference-based* methods
  (`TaskPath`/`EpicPath`/…), which have per-entity help, completions and error
  nouns in the CLI above them. It fits less well for the *selected-identity* twins
  in `internal/core/selected_entity_path.go`, which have no CLI surface of their own
  and genuinely vary in only three tokens. Proposal: when that task is picked up,
  re-read the exclusion against `selected_entity_path.go` specifically rather than
  inheriting it wholesale.
- **Partial overlap, both sides cross-referenced** (M2): `6gppr2jh0j3x-adopt-go-1.24-1.26-idioms-in-error-mapping-tests-and-dependencies`
  (`ready-to-start`) covers two of the 22 modernizers by hand — `new(expr)` and
  `wg.Go`, both of which `go fix` automates as `newexpr` and `waitgroup` — and
  derives its site list from the 2026-09-14 modernization audit. It does not cover
  `stringsseq`, `slicescontains`, `mapsloop`, `rangeint`, `minmax`, `fmtappendf`,
  `stringscut`, or the rest, and a hand-typed inventory drifts in a way
  `go fix -diff` cannot. Left `open`; the finding proposes coordinating rather than
  duplicating. Note also that `6gppr2jwqnpm` (the Go 1.27 bump) explicitly puts
  "Go 1.24-1.26 idioms" out of its own scope, so neither open task owns this.
- **Not dead, do not delete** (L1's neighbour): `(*core.Service).SetCriterionState`
  (`internal/core/service_task.go:412`) has zero callers in `internal/` or `cmd/`
  and zero tests of its own. It reads as dead export, and a mechanical sweep would
  remove it. It must not be: `6g3cp32y9149-let-the-tui-flip-an-acceptance-criterion`
  (`ready-to-start`, tier 2) names it in **Implementation constraints** as one of
  the "existing core write paths" that slice will build on. Flagging it here so the
  next reachability sweep (M2's `deadcode` proposal) does not propose deleting
  queued work — and so that whoever picks up `6g3cp32y9149` knows they will be the
  method's first caller and its first test.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — `tskflwctl task new "Retire the transitional task-graph compatibility surface" --epic 21-code-quality-architecture-hardening --tags refactor,core,simplification --tier 3 --priority low --description "Delete TaskGraphRead.Tasks, taskStoreGraphSource and the dead taskGraphRecords branches; move NewTaskGraph/TaskGraphReadFromFiles to a named test helper, completing task 6gcwcf7rgxef inventory follow-through."`
- ○ M2 · open — `tskflwctl task new "Adopt the go fix modernizers across production packages" --epic 21-code-quality-architecture-hardening --tags modernization,go,toolchain --tier 3 --priority medium --description "Apply go fix to the production files it flags repo-wide, coordinating with 6gppr2jh0j3x which hand-enumerates only new(expr) and wg.Go, and decide whether modernizer drift plus a deadcode reachability pass join the lint recipe."`
- ○ L1 · open — `tskflwctl task new "Flatten the criterion-write pass-through layers in core" --epic 21-code-quality-architecture-hardening --tags refactor,core --tier 3 --priority low --description "Inline setCriteriaState into SetCriteriaState so criteria writes reach transformTaskBody directly, matching SetAcceptanceCriterion and EditCriteria."`
- ○ L2 · open — Roll into M1 candidate task (same test call sites); the doc-comment correction is independent and XS.
