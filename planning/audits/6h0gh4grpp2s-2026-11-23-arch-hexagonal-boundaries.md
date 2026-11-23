---
schema: 1
id: 6h0gh4grpp2s
bucket: open
area: arch-hexagonal-boundaries
date: "2026-11-23"
updated_at: "2026-11-23"
---
# Weekly Architecture Audit: hexagonal-boundaries — 2026-11-23

> Create findings through `tskflwctl audit finding new`; update them through
> `tskflwctl audit finding` so identity, status, resolution, and managed candidate
> metadata stay synchronized. Never hand-edit those fields.
> Architecture audits are propose-only — no code, docs, or ADR edits.

Routine: `weekly-architecture-audit` · lens `hexagonal-boundaries` · ISO week
`2026-W48` (mod 4 = 0).
ADRs consulted: 0001, 0003, 0005, 0007, 0008 (and the recorded absence of a layering ADR) ·
Files read: `internal/core/store.go`, `internal/core/service.go`, `internal/core/finding.go`,
`internal/cli/root.go`, `internal/cli/ports/runtime.go`, `internal/appwiring/wiring.go`,
`internal/appwiring/launch.go`, `cmd/tskflwctl/main.go`, `internal/domain/entity.go`,
`internal/domain/finding.go`, `internal/domain/resolution.go`, `internal/domain/layout.go`,
`.golangci.yml`, `docs/architecture/boundaries.md`, plus edge probes in
`internal/store/edit.go`, `internal/userconfig/userconfig.go`, `internal/cli/edit.go`,
`internal/cli/render/render.go` · Sources cited: 5.

`just` is not installed in this runner, so the build used the spec's documented fallback
(`go build -o bin/tskflwctl ./cmd/tskflwctl`) and validation used the equivalent `go`
commands. Five research hosts (`jerf.org`, `dev.to`, `www.martinfowler.com`, `go.dev`,
`oreilly.com`) are outside this session's network policy and did not resolve; the
comparison below is anchored only to sources that were actually fetched and quoted.

## Executive summary

This lens is in good structural shape and the inward-facing half of it is genuinely
enforced: `domain`, `core`, `wire`, the primary adapters and `cli/ports` each carry a
default-deny import rule, composition is explicit in one binary-selected package, and no
production package violates the documented direction today. The two findings are both
about *who is told when that stops being true*. M1: the catch-all deny pattern that makes
the inward layers default-deny was never extended outward — fifteen production packages,
including the entire seven-package secondary-adapter family, match no depguard rule at
all, so the hexagon's outermost rule is protected only by a descriptive import map whose
documented remedy is to edit the map. M2: `internal/core` re-spells the finding-status
"actionable" partition twice in one file while `domain` already owns the complement via
`TerminalFindingStatus`, and because finding status is a bare `string` rather than a named
type, no linter can see the duplication. Read M1 first; it is the one with real
compounding cost, and its fix is a configuration pattern the repo already uses elsewhere.

## State of the architecture (this lens)

**The seam itself.** `internal/core/store.go` is the port surface, and it is unusually
disciplined for a 479-line file: twenty-odd capability interfaces, each with a comment
stating why it is *separate* from `Store` rather than part of it
(`TaskGraphMutationStore` at `internal/core/store.go:131-135` — "a sibling capability
rather than part of Store so read-only/test adapters do not acquire a mutation method they
cannot implement"; `ThreadPathSource` at `:253-257` — "a database, API, or cache adapter
can provide portable Thread records without inventing a filesystem path"). The aggregate
`Store` (`:430-440`) is deliberately only task/epic/audit/research persistence; every
optional workflow is selected beside it.

**Identity vocabulary across the seam.** Mutation entry points split cleanly into two
shapes and the split is principled, not accidental: reference-based entries take a user
string and resolve it (`EditTask(slug string, …)` at `:68`, implemented at
`internal/store/edit.go:152` via `s.resolve(slug)`), while already-selected actions take
an adapter-established target (`SetFields(target TaskMutationTarget, …)` at `:59`). The
`TaskMutationTarget` comment (`:21-27`) names the rule — "It is not a user reference: an
adapter must match ID exactly" — and no port method mixes the two.

**Capability discovery.** `core.NewService` (`internal/core/service.go:318-387`) discovers
nineteen optional capabilities by type-asserting the aggregate store, then lets functional
options override them, then validates the result with `validateSourceSets`
(`:405-429`). In production the discovery path does all the work: `appwiring.openPlanning`
(`internal/appwiring/wiring.go:58-74`) calls `core.NewService(fs)` with *no options at
all*, so one concrete type satisfies every capability and the segregation is exercised only
by tests and by the workspace route. That is a real asymmetry, but a deliberate one — see
Tensions.

**Composition and DI.** `cmd/tskflwctl/main.go:23-24` is the whole composition root:
`appwiring.LocalBindings()` then `cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr, bindings)`.
`internal/cli/ports/runtime.go` carries no framework types (enforced by a `list-mode:
strict` depguard rule allowing only `$gostd`, core and design). `App.openPlanning` is a
closure supplied by the binary (`internal/cli/root.go:57`), opened lazily in `resolveFrom`
(`:443-467`) after flags are parsed, and a failed composition is held in `compositionErr`
(`:60`) so metadata-only commands stay usable while operational ones refuse. There is no
service locator and no package-global dependency default in any controller.

**Where the injected-stream contract is not total.** `NewRootCmd`'s comment promises "All
I/O flows through the injected streams" (`internal/cli/root.go:258-266`), and width,
color and TTY detection honour it (`:101`, `isTerminal(a.ErrOut)`). Three terminal-
background probes do not: `internal/cli/root.go:543`, `internal/cli/theme.go:101` and
`internal/tui/tui.go:26` each call `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)`
directly. The practical blast radius is nil today (production passes exactly those
streams, and a non-TTY returns the default without querying), so this is recorded as an
observed inconsistency rather than inflated into a finding; the adjacent palette decision
is already owned by task `6g50ry3dhgw7`.

**Entity-kind vocabularies agree.** `domain.entities` (`internal/domain/entity.go:53`) and
`core.EntityKind` (`internal/core/entity_read.go:13-20`) both carry exactly five kinds —
task, epic, audit, research, thread — with no drift. Adding a sixth still touches several
layers (a `Dir` constant, a registry entry, an `EntityKind` const, an entity-specific port,
a CLI command file, a TUI registry row), but `boundaries.md` states that position
explicitly and the bounded-reuse work is queued as `6g63hjmb3d8s`.

**Enforcement topology.** Ten depguard rules, five catch-all `deny: …/internal` entries
(one shared by three rules through the `&cli-controller-deny` anchor), and no `$all`-scoped
rule anywhere in `.golangci.yml`. Eleven production packages match at least one rule;
seventeen match none, of which `cmd/tskflwctl` is the legitimate composition root and
`internal/testutil` is test support. See M1.

## ADR reconciliation

### ADR-0007 §2 — words that mean the same thing in two places are spelled once

**Quoted:** "`domain/resolution.go` holds the shared pool — `deferred`, `wontfix`,
`tracked` — and each entity declares its own full set from that pool plus its own
additions. … A test fails if a shared word is spelled differently in either set."
**Implementation:** drift at one site, follows elsewhere. The shared *pool* is honoured:
`internal/domain/resolution.go:23-35` holds the three words, `SharedResolutionWords()`
(`:39-41`) publishes them, and `TestSharedResolutionWordsAgree` pins them. The *partition*
over the finding-status vocabulary is not: `internal/domain/finding.go:483` owns the
canonical disposition map and `:478` publishes `TerminalFindingStatus` as its complement,
while `internal/core/finding.go:38` declares `actionableStatuses = []string{"open",
"in-progress"}` with its own claim to be "the single definition of 'actionable'", and
`internal/core/finding.go:55-59` re-switches on the same two literals a second time in the
same file. Nothing ties the three together.
**Class if divergent:** ADR gap — §2 decided the mechanism (a cross-entity word pool plus
an agreement test) and is silent on a *partition over one entity's own closed vocabulary*
being re-derived in another layer. See M2. Note this sharpens last cycle's reconciliation,
which recorded finding status as "derived, not transcribed"; that holds for the status
*set* and its glyphs, but not for this partition.

### ADR-0003 §4 — id-led, flat, one directory per entity

**Quoted:** "`<id>-<slug>.md`, in one flat directory per entity (`tasks/`, `epics/`,
`audits/`); no status, bucket, or epic subdirectories."
**Implementation:** follows. Directory *names* are domain knowledge in one place
(`internal/domain/layout.go:8-15`, whose comment states the ADR-0003 §4 basis); absolute
paths stay in the store; no primary adapter rebuilds the layout. `ProjectsDir` survives
only as legacy-cleanup vocabulary (`internal/config/config.go:673,701-703`), not as a live
entity directory — correct, not drift. The 2026-09-28 completion-path drift on this clause
remains fixed.
**Class if divergent:** n/a.

### ADR-0005 — repo-scoped discovery must not be influenced by home scope

**Quoted (implementation status note):** "home config, the separate `spaces.toml`
registry, `space list|add|forget`, shared list/doctor health diagnosis, and global
`--space`/`TSKFLW_SPACE` selection are shipped."
**Implementation:** follows, and the critical half is executable. `config-must-not-read-home-scope`
(`.golangci.yml:211-220`) denies `internal/userconfig` from `internal/config`, and the
rule's own comment records the audit finding that motivated it. Selection precedence is
one function (`App.startDir`, `internal/cli/root.go:383-414`) shared by the fatal and
forgiving paths, and an explicit `--space` fails closed rather than falling back to cwd
(`registeredSpaceStart`, `:426-438`). `startupHome` (`internal/userconfig/userconfig.go:121`)
is a package global but is a documented test-isolation snapshot read only by
`guardTestIsolation`, not a dependency default.
**Class if divergent:** n/a. Note the enforcement is a *narrow* deny, not a default-deny —
`internal/config` could import a primary adapter and pass lint. Folded into M1.

### ADR-0008 — the neutral machine contract

**Quoted (as implemented):** wire is "the neutral machine contract shared by adapters"
(`.golangci.yml:51-53`).
**Implementation:** follows structurally. `wire-stays-adapter-neutral` (`:54-67`) holds
wire to core and domain only, and nothing in `internal/wire` reaches a UI or a filesystem
adapter. The contract exposure of M2 is semantic, not structural: `FindingsRollupJSON`
(`internal/wire/dto.go:331-337`) publishes `open` and `in_progress` as the machine answer,
and those counters are produced by the second of core's two independent spellings.
**Class if divergent:** n/a for this lens; M2 carries the semantic consequence.

### ADR-0001 — ADRs are the durable home for "why"

**Quoted:** "On **`accepted`**, the decision content … **freezes**. New information is
added **append-only** as dated entries under **`## Amendments`**."
**Implementation:** follows — nothing in this run edited an ADR, and M2's proposed clause
is written here for a human to accept or reject, not into ADR-0007.
**Class if divergent:** n/a.

### No ADR governs the layering itself

Unchanged for a third consecutive cycle. The cli/tui/core/store layering, consumer-owned
ports, the composition-root exception and the DI rules live only in
`docs/architecture/boundaries.md` and `.golangci.yml`; boundaries.md:160-162 says so itself
("this extraction does not enact a new decision"). Owned by `6gg7e59mm68g`
(ready-to-start). M1 is the enforcement half of the same gap and is cross-referenced to it
rather than re-filed as an ADR-gap finding.

## Best-practice comparison

### Default-deny beats an enumerated denylist, and the tool says how to get it

**Source:** https://pkg.go.dev/github.com/OpenPeeDeeP/depguard/v2
**Guidance:** "Strict, at its roots, is everything is denied unless in allowed." "Lax, at
its roots, is everything is allowed unless it is denied." A file may match several lists
and "an import must be allowed by all of them."
**This codebase:** partial — `internal/cli/ports` uses `list-mode: strict` with `$gostd`
(`.golangci.yml:182-186`), and nine other rules achieve effective default-deny for
repository-internal imports by pairing `lax` with a catch-all `deny:
github.com/andy-esch/taskflow/internal`. Both patterns are correct. They are simply not
applied to fifteen production packages.
**Justified?** No. The mechanism is already in the file twice; the gap is coverage, not
design.

### Put the rules every file shares in one `$all` list

**Source:** https://raw.githubusercontent.com/OpenPeeDeeP/depguard/master/README.md
**Guidance:** "Put the rules every file shares in one list (for example with `$all`), and
add separate lists for the extra rules of a subset of files (for example `$test`)." The
`$all` variable "matches all go files"; "If the files list is empty, then all files will
match that list."
**This codebase:** diverges — every one of the ten rules is scoped to an explicit package
glob and `$all` appears nowhere. A new package is therefore unconstrained by construction
rather than denied by construction, which is the exact failure mode the 2026-09-28 M1
finding fixed one layer in.
**Justified?** No. This is the concrete shape of M1's recommendation.

### An architectural rule with no executable check is a convention

**Source:** https://github.com/fe3dback/go-arch-lint
**Guidance:** the linter "compares the actual (code) and desired (config) dependency
graph" and "if it got a non-empty DIFF, then project has some issues" — i.e. the desired
graph is declared policy that code is measured against.
**This codebase:** partial, and honestly labelled. `boundaries.md` publishes a versioned
`internal-import-map:v1` block checked against `go list` by
`internal/tools/architecturecheck/imports_test.go`, and warns in its own prose that
"**observed imports are not an allowlist**". That is the right caveat, and it is exactly
the problem: for the fifteen unruled packages the map is the *only* check, and its drift
message names "added code facts", whose natural remedy is to add the row. A descriptive
tripwire whose fix is a documentation edit is not a policy.
**Justified?** No for the secondary-adapter family (M1); yes for the pure leaves (`id`,
`tomledit`, `editor`, `listfilter` have no internal imports at all, so there is nothing to
constrain yet).

### Name the complete set of optional capabilities where the consumer can find it

**Source:** https://raw.githubusercontent.com/golang/go/master/src/net/http/responsecontroller.go
**Guidance:** `NewResponseController`'s doc comment enumerates the optional methods in one
place — "If the ResponseWriter implements any of the following methods, the
ResponseController will call them as appropriate" — and absence is a typed sentinel: "If
the ResponseWriter does not support a method, ResponseController returns an error matching
[ErrNotSupported]."
**This codebase:** partial. The naming half is excellent — `core.Store`'s comment
(`internal/core/store.go:430-440`) states which capabilities are in the aggregate and which
are selected separately, and each optional port says why. The typed-absence half is the
open M1 of the 2026-10-26 run and is not re-reported here.
**Justified?** The naming discipline needs no defence; the sentinel gap is already filed.

### A closed vocabulary a linter cannot see must be derived, not retyped

**Source:** https://pkg.go.dev/github.com/nishanths/exhaustive/passes/exhaustive
**Guidance:** the analyzer applies when "the type of the switch expression is an enumerated
type T", and "A given case expression must be a constant to satisfy a corresponding value
in the value set of T." Also: "The presence of a default case will not satisfy missing
cases."
**This codebase:** diverges at the M2 site, and the divergence is structural rather than
careless. `domain.Finding.Status` is a bare `string`
(`internal/domain/finding.go:20`) and `findingStatusSpecs` holds plain string literals
(`:199-207`), so finding status is not an enumerated type in this analyzer's sense. Task
`6gcqz5aqt2sg` plans to adopt `exhaustive`; that adoption cannot cover these sites.
**Justified?** No — but the fix is a domain accessor, not a linter, which is precisely why
M2 is not foldable into the existing vocabulary task.

## Tensions and trade-offs

**One concrete store satisfies twenty ports.** Production passes `core.NewService(fs)` with
no options, so the capability segregation in `store.go` buys nothing at runtime today. It
is not dead weight, though: the segregation is what lets the workspace route compose a
read-only bundle, what lets focused tests inject one capability without counterfeiting
nineteen, and what makes `validateSourceSets` able to refuse an incoherent composition. The
honest reading is that this project pays a documentation and review cost for a seam whose
*second* implementation is the test suite — and `boundaries.md#why-these-boundaries` makes
exactly that argument rather than inventing a future HTTP adapter. Not a finding.

**Thirty-one mutable package globals in the store adapter.** `internal/store` carries 31
`var testHook… func(…)` globals across eleven production files, checked on every guarded
write path and nil in production. By the letter of "no package-global dependency defaults"
this looks like the L1 the 2026-10-26 run filed against the TUI's one global. It is a
different thing: these are crash-window instrumentation for failure modes that are
otherwise unreachable, the convention is named and reasoned in `6fes83r03s90`, and six
prior audits rely on it for their recovery evidence. The real cost is narrow and currently
unpaid — process-global hooks structurally prevent `t.Parallel()` in the repo's largest
adapter package, and no store test uses `t.Parallel()` today. Recorded here rather than
filed, because re-litigating a settled convention with no observed harm is the kind of
invented finding this routine exists to avoid.

**Three terminal probes bypass the injected streams.** Covered in State of the
architecture. Genuinely small, genuinely an inconsistency with a stated contract, and
genuinely not worth a Low that would only create churn next to two Mediums about something
else.

## Findings

#### M1. Fifteen production packages match no import rule, so the hexagon's outward direction is guarded only by a descriptive map · **Status:** open

**File:** .golangci.yml:12 | **Component:** build / dependency policy
**Effort:** S · **Urgency:** soon

**Class:** unanchored (no ADR governs the layering; best-practice-anchored. The ADR half of
this gap is already owned by `6gg7e59mm68g` and is not re-filed here.)
**Anchored to:** https://pkg.go.dev/github.com/OpenPeeDeeP/depguard/v2 ·
https://raw.githubusercontent.com/OpenPeeDeeP/depguard/master/README.md

`.golangci.yml` declares ten depguard rules. Every one is scoped to an explicit package
glob (`**/internal/domain/*.go`, `**/internal/tui/*.go`, …) and `$all` appears nowhere in
the file. Because depguard is per-list and "Lax, at its roots, is everything is allowed
unless it is denied", a package that matches no list has no deny entry and is
unconstrained — not denied by default, but unruled by default.

Measured against `go list ./internal/... ./cmd/...`, eleven production packages match at
least one rule (`domain`, `core`, `wire`, `tui`, `configui`, `cli`, `cli/render`,
`cli/prompt`, `cli/ports`, `appwiring`, `config`) and seventeen match none. Discounting
`cmd/tskflwctl` (the legitimate composition root) and `internal/testutil` (test support),
fifteen production packages have no import policy at all:

- the whole secondary-adapter family: `store`, `configstore`, `spacestore`,
  `workspacestore`, `spacehealth`, `userconfig`
- presentation leaves: `design`, `theme`, `progressbar`, `themepreview`, `graphfmt`
- foundations: `id`, `tomledit`, `editor`, `listfilter`

Two of the eleven "covered" packages are also only narrowly covered: `config` carries a
single deny of `internal/userconfig` (`.golangci.yml:211-220`) and `appwiring` a single
deny of `config`/`userconfig` outside `local_sources.go` (`:192-202`). Neither has a
catch-all. So `internal/store` importing `internal/cli`, or `internal/config` importing
`internal/tui`, would pass `golangci-lint` today. Neither creates an import cycle, so the
compiler does not object either.

The only thing that would notice is `TestArchitectureImportMap`, which diffs the
`internal-import-map:v1` block in `docs/architecture/boundaries.md` against `go list` and
reports "added code facts". That is a documentation-sync tripwire, not a policy — as
boundaries.md itself warns, "**observed imports are not an allowlist**". Its failure
message names a row that is missing from a Markdown table, and the obvious way to make it
green is to add the row. For an agent working to a green suite that is not a guardrail,
it is a prompt.

This is the same shape as finding M1 of the 2026-09-28 run ("the primary-adapter fitness
rule enumerates denied packages, so a new adapter package escapes it"), which was fixed in
PR #277 by moving primary adapters to a catch-all deny. The fix was applied to the layer
that prompted it and never extended outward, so the inward half of the hexagon is
default-deny and the outward half is default-unruled.

The pure leaves are the justified part: `id`, `tomledit`, `editor` and `listfilter` have no
internal imports at all today, so there is nothing yet to constrain. The secondary-adapter
family is not — it has grown from one package to six over the life of epic 29 and Thread
work, and each new one arrived unguarded.

**Compounding cost:** every future secondary adapter and presentation package starts
outside the policy, so the cost is paid per package added rather than once. The
`serve`/web adapter contemplated by ADR-0004 §3 and epic 19 is the next one, and it is
precisely the kind of adapter that would be tempted to reach sideways into `cli` for an
exit-code or rendering helper. Fixing it now is one rule plus one allow-list for an
existing, stable six-package family whose real imports are already written down in the map;
fixing it after a violation lands means arguing about whether the existing edge is a named
exception, which is how `appwiring -> config/userconfig` and `cli/init.go -> config`
reached the exceptions table. The ADR task `6gg7e59mm68g` is also blocked behind this in
practice: it is hard to record a dependency *policy* when two-thirds of the packages are
outside the mechanism that would enforce it.

**Recommendation:** Add one $all-scoped depguard rule denying the repository internal prefix with no allowances, so a package with no named rule is denied by default; then give the secondary-adapter family its own rule allowing core/domain/id and the config readers it genuinely needs.

#### M2. Core re-spells the actionable finding-status partition twice while domain already owns its complement · **Status:** open

**File:** internal/core/finding.go:38 | **Component:** core / domain vocabulary
**Effort:** XS · **Urgency:** eventually

**Class:** ADR gap (ADR-0007 §2 decided the mechanism for cross-entity shared *words* and
is silent on a partition over one entity's own closed vocabulary being re-derived in
another layer)
**Anchored to:** ADR-0007 §2 ·
https://pkg.go.dev/github.com/nishanths/exhaustive/passes/exhaustive

`internal/domain/finding.go` owns the finding-status vocabulary and its meaning.
`findingStatusSpecs` (`:199-207`) is the ordered table, `tallyFindingStatus` (`:483`) is
documented as "the single source of 'what each status means for progress'", and
`TerminalFindingStatus` (`:478`) publishes the complement of "actionable" as
`t.Done+t.Dropped == 1`.

`internal/core/finding.go` then spells the same partition twice more:

1. `actionableStatuses = []string{"open", "in-progress"}` (`:38`), whose own comment claims
   it is "the single definition of 'actionable'" — a second package asserting sole
   ownership of the same concept. `isActionableFinding` (`:43-45`) matches against it.
2. `rollupFindings` (`:55-59`) switches on the bare literals `"open"` and `"in-progress"`
   again, in the same file, to produce the `Open`/`InProgress` counters.

No test pins any of the three spellings to another: `actionableStatuses` and
`isActionableFinding` appear in no `_test.go` file in the repository.

The failure is not hypothetical arithmetic. `Service.Summary` filters with
`isActionableFinding` (`internal/core/service.go:610`) and counts with `rollupFindings`
(`:668`), and `wire.ToFindingsRollup` (`internal/wire/dto.go:340`) publishes the result as
`findings.open` and `findings.in_progress` — described in the schema as "actionable
findings with status in-progress". Add an eighth actionable status (say `blocked`) to
domain and to `actionableStatuses` but not to the rollup switch, and findings pass the
filter, land in `ByUrgency`/`ByComponent`, and are counted in neither published counter:
`status --json` reports a findings rollup whose components do not sum to the set it was
built from. Add it to domain only and the dashboard silently stops seeing the status
altogether, while `audit lint` continues to accept it. The two layers fail differently
because they fail from different lists.

Task `6gcqz5aqt2sg` plans to adopt the `exhaustive` analyzer for closed-vocabulary
switches, and that adoption structurally cannot reach these sites: the analyzer requires
"the type of the switch expression is an enumerated type T" and that "A given case
expression must be a constant", but `domain.Finding.Status` is a bare `string`
(`internal/domain/finding.go:20`) and the spec table holds plain string literals. The same
task's 2026-11-17 note already records that `render.findingStatusOrder` "needs a new domain
accessor, not a coverage test" and was left out of scope for that reason — this is the
same remedy for a different site, and unlike `findingStatusOrder` (whose set is pinned both
ways by `internal/cli/render/render_test.go:28-51`) this one has no pin at all.

**Compounding cost:** small per month but paid at the worst moment. The cost lands when the
finding-status vocabulary next changes — which epic 27 ("agent code review on tasks") makes
likely, since it introduces review findings with their own lifecycle and explicitly lists
"finding identity across passes/lenses" as an open piece. At that point the editor must
discover three unlinked spellings across two packages with no test to lead them to the
second and third, and the symptom of missing one is a machine-contract payload that is
internally inconsistent rather than an error. Fixing it now is one exported predicate and
two call sites.

**Proposed ADR amendment:** extend ADR-0007 §2 with a clause stating that a *classification
over* a closed vocabulary — terminal vs actionable, done vs dropped — is owned by the same
package that owns the vocabulary and is consumed through an exported predicate, not
re-derived from literals by a consuming layer; and that where the vocabulary is not a named
Go type (so `exhaustive` cannot see it), the agreement must be pinned by a test the way
`TestSharedResolutionWordsAgree` pins the shared word pool. Written here for a human to
accept or reject; ADR-0007 is accepted and was not edited.

**Recommendation:** Export the partition from domain (a FindingStatusActionable predicate, or reuse TerminalFindingStatus) and have isActionableFinding and rollupFindings derive from it instead of spelling the two literals.

## What audited clean

- **`--dry-run` is refused, never ignored, on all four interactive edit verbs.**
  `internal/cli/edit.go:41`, `epic.go:149`, `audit.go:646`, `research.go:107` each return a
  validation error pointing at the scriptable preview, with the reason stated in-band ("a
  safety flag must never be a no-op"). The pattern holds at every site — this is the kind
  of edge check that usually finds a holdout, and there isn't one.
- **The composition root is two lines and holds no policy.** `cmd/tskflwctl/main.go:23-24`
  selects bindings and injects the three real streams; everything else in `main.go` is
  presentation (fang chrome) or the machine-path gate. Satisfies
  `boundaries.md#composition-and-local-exceptions`.
- **Invocation contracts are framework-free, and that is enforced rather than asserted.**
  `internal/cli/ports/runtime.go` carries no Cobra, Bubble Tea or Lipgloss type, held there
  by the one `list-mode: strict` rule in the file (`.golangci.yml:175-187`).
- **The TUI does no I/O in `Update`/`View`.** No service call appears inside any
  `Update`/`View` method, and the only process launch is the documented clipboard exception
  (`internal/tui/clipboard.go:26`). Satisfies the AGENTS.md TUI rule and the
  `tui -> editor/os/exec` named exception.
- **`core` and `domain` carry no filesystem vocabulary.** The only `filepath` calls in core
  are pure string formatting (`internal/core/dependency_graph.go:672`,
  `space_registry.go:225`) and `domain` has none. Matches what task `6ggxymjbf86r` predicted
  it would find, rather than the violation it guards against.
- **Reference-based and selected-identity mutations never mix on the port.** Every
  `core.Store` method takes either a user reference it resolves or an adapter-established
  `*MutationTarget`, consistent with `internal/core/store.go:21-27`.
- **Entity-kind vocabularies agree across domain and core**, five kinds each, no drift
  (`internal/domain/entity.go:53`, `internal/core/entity_read.go:16-20`).
- **The one hand-ordered status list in presentation is set-pinned to the registry.**
  `render.findingStatusOrder` (`internal/cli/render/render.go:823`) is checked both ways by
  `internal/cli/render/render_test.go:28-51`, so only its *order* is hand-held — and order
  is a presentation decision presentation may own. Already cross-referenced by the
  2026-11-17 run.

## Related-task observations (propose-only)

- Scope note: `planning/tasks/6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md`
  carries `depends_on: [6gg7e59mm68g]`, so the only queued import-enforcement work is gated
  behind the layering ADR (`6gg7e59mm68g`, ready-to-start since 2026-10-03, itself gated
  behind the completed `6gcwcf8rxe72`). The gate is right for *that* task, which proposes a
  new policy about standard-library and framework imports. It is arguably wrong for M1: a
  catch-all `$all` deny over packages that already comply adds no new decision and changes
  no production import, so it could land ahead of the ADR and would give the ADR author a
  mechanism to point at rather than describe. Proposed only — the sequencing call is the
  owner's, and no dependency was added or changed by this run.
- No task appears demotable or already-shipped from this lens. The three open findings of
  the 2026-10-26 run (M1 absent-capability sentinels, M2 composable guarded writes, M3
  source-set pairing coverage) were re-checked against the current tree and all three still
  reproduce at the cited sites; none is re-reported here.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — tskflwctl task new "Make every production package import-ruled by default with an $all-scoped deny" --epic 21-code-quality-architecture-hardening --tags architecture,lint,build --effort S --priority medium
- ○ M2 · open — tskflwctl task new "Derive the actionable finding-status partition from domain instead of re-spelling it in core" --epic 21-code-quality-architecture-hardening --tags architecture,core,domain --effort XS --priority medium
