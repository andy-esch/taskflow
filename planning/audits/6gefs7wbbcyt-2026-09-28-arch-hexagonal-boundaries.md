---
schema: 1
id: 6gefs7wbbcyt
bucket: open
area: arch-hexagonal-boundaries
date: "2026-09-28"
updated_at: "2026-09-28"
---
# Weekly Architecture Audit: hexagonal-boundaries — 2026-09-28

> Create findings through `tskflwctl audit finding new`; update them through
> `tskflwctl audit finding` so identity, status, resolution, and managed candidate
> metadata stay synchronized. Never hand-edit those fields.
> Architecture audits are propose-only — no code, docs, or ADR edits.

Routine: `weekly-architecture-audit` · lens `hexagonal-boundaries` · ISO week
`2026-W40` (mod 4 = 0). On-schedule Monday run.
ADRs consulted: ADR-0003 §2/§4, ADR-0005 §5, ADR-0006 §5/§9, ADR-0001 (amendment policy) ·
Files read: `internal/core/store.go`, `internal/core/service.go`, `internal/core/source_set.go`,
`internal/cli/root.go`, `internal/cli/command_safety.go`, `internal/cli/completion.go`,
`internal/store/fsstore.go`, `internal/workspacestore/fs.go`, `.golangci.yml`,
`docs/ARCHITECTURE.md` · Sources cited: 5

## Executive summary

The hexagon itself is in good shape, and this audit found no defect in the running
dependency direction. `internal/core` imports *nothing* outside `internal/domain` and
`internal/id` — not even a third-party package — ports are consumer-owned and segregated
into narrow capabilities, and no inward package reaches a primary adapter. The findings
are all one step removed from the code: they are about the **enforcement** of the boundary
rather than the boundary.

The headline is M1, and it is empirically proven rather than argued. The depguard rule
that guards the primary adapters enumerates forbidden packages by name under
`list-mode: lax`, so a **brand-new** internal package is permitted by default. A probe in
an isolated copy of the repo confirmed this three ways: `internal/tui` importing a new
`internal/fakestore` package reports `0 issues`, while the same import from
`internal/core` is correctly denied and `internal/tui` importing the *enumerated*
`internal/store` is correctly denied. The rule works exactly and only for the packages
someone remembered to list.

M2 is the same shape one layer down: the mutation-authorization guard defaults to *allow*
when a composition root omits the option. M3 records that the repo's central architectural
rule — the cli/tui/core/store layering — has never been ratified as an ADR and lives only
in a freely-editable doc. M4 is tracked drift already owned by task `6gcwcf8gzn50`.

Read M1 first; it is XS effort and it is the finding that would have caught the others.

## State of the architecture (this lens)

**Direction is real, not aspirational.** `go list` over the production graph confirms the
documented inward direction exactly: `internal/domain` imports only `internal/id`;
`internal/core` imports only `internal/domain` and `internal/id`; `internal/wire` imports
`core`, `domain`, and `github.com/invopop/jsonschema`. Nothing under `core`, `domain`,
`wire`, or `store` transitively reaches `internal/cli`, `internal/tui`, or
`internal/configui`. The TUI could be deleted without disturbing core.

**Ports are consumer-owned and genuinely segregated.** `internal/core/store.go:1-3` states
the rule in its package comment ("Interfaces are defined here, at the consumer"), and the
file delivers it: the aggregate `Store` (store.go:311-316) composes only the four
entity-use-case ports, while capabilities that a read-only or test adapter cannot honestly
provide are separate interfaces — `TaskGraphMutationStore` (store.go:96-98),
`TaskGraphRepairStore` (store.go:105-107), `TaskLifecycleMutationStore` (store.go:115-117),
`ThreadStore` (store.go:150-153), `ThreadPathSource` (store.go:160-162), and the three
Thread mutation stores (store.go:166-190). `Fixer`, `Linter`, and `Layout`
(store.go:325-333) are deliberately outside `Store` because they are fs/text operations
rather than use cases.

**The composition root is one function.** `internal/cli/root.go:255-325` (`newRootCmd`)
plus `resolveFrom` (root.go:427-460) are where every concrete adapter is constructed:
`spacestore.New`, `configstore.New`, `workspacestore.New`, and `store.NewFS`. `App`
(root.go:32-95) is the container, populated lazily in `PersistentPreRunE` because the
dependencies need parsed flags. The narrow ports are assigned as core interface types
(`a.Fixer = fs; a.Layout = fs; a.Linter = fs`, root.go:456-458), so consumption sites
depend on `core.Fixer`, not on `*store.FS`.

**Capability discovery is runtime type assertion.** `core.NewService`
(service.go:228-277) discovers optional capabilities from a complete adapter by asserting
interfaces — `LintSource`, `AuditSnapshotSource`, `TaskGraphSource`, five mutation stores,
`ThreadStore`, `ThreadPathSource`. This is a deliberate ergonomics trade (a complete
adapter supplies one value; a split adapter supplies narrow ports through options), and it
is guarded: `validateSourceSet` (source_set.go:44-68) requires every selected planning-data
capability to publish the same non-zero, stable `SourceSetID` and returns
`ErrIncompatibleCapabilities` otherwise. `NewService` returns that error before exposing
the service.

**Safety crosses the hexagon as a neutral callback.** `commandSafetyState`
(command_safety.go:22-25) reads a Cobra annotation in the primary adapter and hands the
secondary adapters a `func() error`. The adapters never learn the `read-only`/`mutating`
vocabulary — `store.WithMutationAuthorization` (fsstore.go:80-84) takes only the callback.
Twenty-six filesystem mutation entries call it, and
`TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects`
(`internal/store/mutation_authorization_test.go:14`) pins that.

**The enforcement layer is `.golangci.yml`, and it is uneven.** Four depguard rules encode
the direction. Three of them (`domain-stays-inward`, `core-owns-ports-not-adapters`,
`wire-stays-adapter-neutral`) pair a narrow `allow` with a `deny` on the whole
`github.com/andy-esch/taskflow/internal` prefix, which under lax mode is effectively
deny-by-default. The fourth (`primary-adapters-use-application-seams`,
`.golangci.yml:69-104`) instead enumerates ten specific denied packages. That asymmetry is
M1.

## ADR reconciliation

### ADR-0006 §5 — no graph-library type may cross the taskflow-owned analysis interface

**Quoted:** "No graph-library type may cross the taskflow-owned analysis interface into
domain, persistence, or wire contracts."

**Implementation:** follows — provably, and more strongly than the clause requires.
`go list -f '{{range .Imports}}...'` over `./internal/core` returns only
`internal/domain` and `internal/id`: `internal/core` has **zero** third-party imports, so
there is no graph library to leak. The bake-off in §5 was resolved toward the owned
implementation and `internal/core/dependency_graph.go` carries it.

**Class if divergent:** n/a.

### ADR-0006 §9 — dependency analysis is pure core; filesystem code does not decide semantics

**Quoted:** "Dependency analysis belongs in pure core/domain-facing code over task IDs and
statuses. Filesystem code parses and atomically mutates documents but does not decide
frontier, sound completion, or topological semantics. Introduce narrow consumer-owned
dependency and Thread persistence ports."

**Implementation:** follows. The control inversion is explicit: `TaskGraphPlanner`
(`internal/core/store.go:89-93`) is a pure core callback the *store* invokes while holding
the repository guard, and its contract comment states the planner "must not call a Store
method or begin another mutation." `TaskDependencyWrite` (store.go:59-63) carries only a
task ID and its complete canonical dependency set — the store owns YAML surgery, core owns
semantics. `TaskGraphSource`, `ThreadStore`, and `ThreadPathSource` are the narrow
consumer-owned ports the clause asks for.

**Class if divergent:** n/a.

### ADR-0003 §2 — the `status == directory` invariant is retired

**Quoted:** "The `status == directory` / `bucket == directory` invariant — a stated
non-negotiable in CLAUDE.md / ARCHITECTURE.md — is **retired** (those docs update *with*
the implementation, not before)."

**Implementation:** drift, in contract comments rather than behaviour.
`internal/cli/completion.go:206-207` still documents `taskCompleter` as completing "task
slugs whose status (== their directory)", `completion.go:245-247` says the same of audit
buckets, and `slugsFromGlobs` (completion.go:181-182) states "Because status/bucket *is*
the directory, the caller selects which dirs to glob to filter by state". The executable
code was migrated correctly — `taskCompleter` parses frontmatter via
`store.NewFS(root).ListTasks()` precisely because status is no longer the directory, and
says so at completion.go:213-216 — but three contract comments still describe the retired
model.

**Class if divergent:** implementation drift (M4).

### ADR-0003 §4 — id-led, flat, one directory per entity

**Quoted:** "`<id>-<slug>.md`, in one flat directory per entity (`tasks/`, `epics/`,
`audits/`); no status, bucket, or epic subdirectories."

**Implementation:** follows in the store, duplicated in the CLI. `store.WatchPaths`
(`internal/store/fsstore.go:128-133`) owns the flat layout and exposes it through the
`core.Layout` port — a port that exists *because* this exact knowledge escaped once before
and was pulled back (task `6fbj87001q7p`, completed 2026-06-14, "Put storage-layout
knowledge back behind the port"). The completion path re-derives it anyway, globbing
`filepath.Join(root, domain.TasksDir, "*.md")` at completion.go:229, `domain.ThreadsDir` at
:239, and `domain.AuditsDir` at :266.

**Class if divergent:** implementation drift (M4).

### ADR-0005 §5 — the registry is advisory, never authoritative

**Quoted:** "**Nothing in the registry may change what `Discover` resolves from a given
cwd.**"

**Implementation:** follows. `startDir` (`internal/cli/root.go:345-385`) keeps registry
selection an explicit opt-in branch: `--space`/`TSKFLW_SPACE` resolves a label through
`SpaceSvc.Resolve` and discovers from the recorded checkout, while the default path is
plain `os.Getwd()`. `registeredSpaceStart` (root.go:405-412) fails loudly on a broken
entry and never falls back to cwd, matching the clause's second paragraph. The
`config-must-not-read-home-scope` depguard rule (`.golangci.yml:106-118`) makes the
direction executable.

### No ADR governs the cli/tui/core/store layering itself

This lens's central subject has **no** governing ADR. `planning/adrs/` contains eight
ADRs; none establishes the primary-adapter / core / secondary-adapter layering, the
consumer-owned port rule, the composition-root exception, or the DI discipline. Those
rules live in `docs/ARCHITECTURE.md` and `.golangci.yml` only. ADR-0006 §5/§9 decide a
package boundary for *dependency analysis* specifically; they presuppose the layering
rather than establish it.

**Class:** ADR gap (M3).

## Best-practice comparison

### Deny-by-default is the mode for architectural boundaries

**Source:** https://github.com/OpenPeeDeeP/depguard/blob/v2/README.md

**Guidance:** "Strict, at its roots, is everything is denied unless in allowed." "Lax, at
its roots, is everything is allowed unless it is denied."

**This codebase:** partial — and the split is the finding. Three of the four boundary rules
reach deny-by-default despite using `list-mode: lax`, because their `deny` names the whole
`github.com/andy-esch/taskflow/internal` prefix and their `allow` is narrower and more
specific. The fourth, `primary-adapters-use-application-seams`, enumerates ten packages,
so lax means what it says: a package not on the list is allowed.

**Justified?** No. The three sibling rules show the repo already knows the stronger
pattern, and the two rules cost the same to write.

### Interfaces belong in the consuming package; implementations return concrete types

**Source:** https://github.com/golang/wiki/blob/master/CodeReviewComments.md

**Guidance:** "Go interfaces generally belong in the package that uses values of the
interface type, not the package that implements those values." "The implementing package
should return concrete (usually pointer or struct) types: that way, new methods can be
added to implementations without requiring extensive refactoring."

**This codebase:** follows, deliberately and with the rationale written down.
`internal/core/store.go:1-3` cites the principle in its package comment. `store.NewFS`
returns `*store.FS`, a concrete type; the interfaces live in `core` beside their consumer;
compile-time assertions (`fsstore.go:92-95`) pin satisfaction without inverting ownership.

**Justified?** n/a — no divergence.

### Information leakage: the same design decision reflected in two modules

**Source:** https://github.com/alysivji/notes/blob/main/software-engineering/philosophy_of_software_design.md
(notes on Ousterhout, *A Philosophy of Software Design*)

**Guidance:** Information leakage is the "Same knowledge used in multiple places"; it
"occurs when a design decision is reflected in multiple modules" and "creates dependencies
where changes require touching many different modules."

**This codebase:** partial. The flat id-led layout is known in `internal/store` (which owns
it), exposed through `core.Layout` (the port built for it), and *re-derived* in
`internal/cli/completion.go`. The cost is not hypothetical: the completion copy kept a
stale model across the ADR-0003 migration while the store copy was updated.

**Justified?** Partly. Completion must be fast and must offer records whose frontmatter
will not parse, which the semantic port does not serve today. That is a real constraint,
and task `6gcwcf8gzn50` proposes the right answer — a parse-free completion capability —
rather than forcing completion through the semantic read.

### Explicit wiring should be checkable without running the program

**Source:** https://github.com/google/wire

**Guidance:** "Dependencies between components are represented in Wire as function
parameters, encouraging explicit initialization instead of global variables. Because Wire
operates without runtime state or reflection, code written to be used with Wire is useful
even for hand-written initialization."

**This codebase:** partial. The wiring is explicit and reflection-free, which is the
substance of the recommendation. But two composition obligations are expressed as
*optional* functional options rather than required parameters:
`store.WithMutationAuthorization` and its three siblings. Omitting one is not a compile
error and not a test failure; it silently removes the guard (M2). The repo already models
the alternative — `core.NewService` validates its `SourceSetProvider` witness at
construction and returns a typed error.

**Justified?** No, though the cost is currently latent rather than realised.

### A composition root should be a single, unique location

**Source:** https://blog.ploeh.dk/2011/07/28/CompositionRoot/ (Mark Seemann). *Direct fetch
was blocked by this environment's egress proxy; the wording below is as surfaced by search
and is the canonical definition, but it was not read first-hand.*

**Guidance:** "A Composition Root is a (preferably) unique location in an application
where modules are composed together."

**This codebase:** partial. `newRootCmd` + `resolveFrom` is the intended single location and
`docs/ARCHITECTURE.md:75-81` documents the exception explicitly. But the depguard exemption
is granted to the whole `internal/cli` package — 36 production files — while only six
import a secondary adapter (`root.go` imports all of them; `completion.go` takes `store`
and `config`; `init.go`, `ui.go`, and `workspace.go` take `config`). One of those,
`completion.go:222` and `:259`, constructs a second `store.NewFS(root)` outside the
composition root and without either option `root.go:449-452` supplies. This is already
owned by task `6gcwcf8rxe72`.

**Justified?** The exception is justified; its *breadth* is an artifact of package
granularity rather than a decision, and the owning task says so.

## Tensions and trade-offs

**Capability discovery by type assertion versus compile-time completeness.**
`core.NewService` discovers a dozen optional capabilities by asserting interfaces against
the aggregate store. Judged against a "make illegal states unrepresentable" standard this
looks like a smell, but it is the same pattern the standard library uses (`io.Copy`
probing for `WriterTo`), and it buys the thing the project actually needs: one complete
filesystem adapter stays ergonomic while a future split or remote adapter can supply
narrow ports independently. The `SourceSetID` witness is the right mitigation — it catches
the dangerous failure (capabilities addressing *different* corpora) while tolerating the
benign one (a capability legitimately absent). No finding.

**A whole-repo guard instead of row-level concurrency.** taskflow is single-user and
local-first; a repository-wide critical section with control-inverted planners is a
correct simplification, not an antipattern. Scoring it against a multi-tenant service's
concurrency model would be scoring it on a scale it deliberately isn't on.

**The `internal/cli` exemption is a cost of Go package granularity.** depguard scopes rules
by file glob, so the composition root and 30 ordinary controllers share one exemption
because they share one package. Splitting the package to satisfy a linter would be
tail-wagging; `6gcwcf8rxe72` proposes the right order — move the wiring first, then narrow
the rule.

**Five heterogeneous entities resist further collapse.** `docs/ARCHITECTURE.md` argues at
length that the residual per-entity fan-out (a typed domain struct, a parser, thin port
methods, use cases, a command, render/TUI delegates) is the price of a typed domain, not
evidence of a boundary defect. This audit agrees: the generic seams (`scanDir[T]`,
`resolveID`, `Column[T]`, the entity `Descriptor`) already remove the mechanics.

## Findings

#### M1. The primary-adapter fitness rule enumerates denied packages, so a new adapter package escapes it · **Status:** open

**File:** .golangci.yml:69-104 | **Component:** build/lint policy
**Effort:** XS · **Urgency:** soon

**Class:** unanchored (no ADR governs the fitness-function mechanism; best-practice-only)
**Anchored to:** depguard v2 README — "Lax, at its roots, is everything is allowed unless it is denied."

`.golangci.yml` carries four depguard rules that make `docs/ARCHITECTURE.md`'s dependency
direction executable. Three of them — `domain-stays-inward` (:19-31),
`core-owns-ports-not-adapters` (:35-48), `wire-stays-adapter-neutral` (:52-65) — pair a
narrow `allow` with a `deny` on the whole `github.com/andy-esch/taskflow/internal` prefix.
Under lax mode that is deny-by-default: an import is allowed only when it matches the more
specific allow rule.

`primary-adapters-use-application-seams` (:69-104) is built the other way. Its `deny`
enumerates ten packages by name (`cli`, `tui`, `config`, `configstore`, `spacehealth`,
`spacestore`, `store`, `tomledit`, `userconfig`, `workspacestore`). Lax mode then means
exactly what the README says: anything not on that list is allowed.

Verified empirically in an isolated copy of the repo at `5da70d1` (golangci-lint 2.5.0,
`--enable-only depguard`, probe removed afterward, working tree untouched):

```
# a brand-new secondary-adapter package, internal/fakestore

internal/tui/action.go imports internal/fakestore   -> 0 issues
internal/core/board.go imports internal/fakestore   -> DENIED
    "internal/core may depend only on internal/domain and internal/id"
internal/tui/action.go imports internal/store       -> DENIED
    "primary adapters must use core services and ports, not the Markdown store"
```

The control proves the rule is live and the contrast proves the prefix deny works. The
rule guards exactly the ten packages someone remembered to list, and nothing else.

This is not hypothetical drift. `internal/graphfmt` and `internal/workspacestore` were both
added after the rule was written; `workspacestore` was explicitly added to the deny list,
`graphfmt` was not. Nothing in the build would have said either way.

**Compounding cost:** the rule's job is to survive the author's attention, and it does not.
Every new `internal/*` package is admitted to the primary adapters by default and must be
noticed by a human reviewer instead — the failure mode a fitness function exists to remove.
Epic 19 (`serve`, web companion apps over a shared core) and epic 29 (atlas) are both
actively adding packages, and epic 21's own follow-up task `6gcwcf8rxe72` proposes an
acceptance criterion that enumerates four package names, which would extend the same
fail-open shape into the CLI controllers. Fixing the rule now is an XS config edit; fixing
it after a third primary adapter exists means auditing whatever crossed in the meantime.

**Follow-up:** the four file globs are single-level (`**/internal/tui/*.go`), so a future
subpackage of a guarded package is also unguarded — see L1.

**Recommendation:** Give primary-adapters-use-application-seams the same shape as its three siblings: keep list-mode lax, replace the ten enumerated deny entries with a single deny on github.com/andy-esch/taskflow/internal, and move the packages primary adapters legitimately use (configui, id, domain, core, wire, editor, listfilter, graphfmt, theme, design, progressbar) into allow. Re-run the fakestore probe as a negative test.

#### M2. Mutation authorization is fail-open when a composition root omits the option · **Status:** open

**File:** internal/store/fsstore.go:86-90 | **Component:** store + configstore + spacestore + workspacestore
**Effort:** S · **Urgency:** soon

**Class:** ADR gap (ADR-0008 publishes the safety capability in the machine contract but is silent on what enforces it)
**Anchored to:** google/wire — "encouraging explicit initialization instead of global variables … without runtime state or reflection"; and `core.NewService`'s own precedent in this repo.

Four secondary adapters accept the command-safety guard as an optional functional option,
and all four default to *allow* when it is absent:

```
internal/store/fsstore.go:86-90        if s.mutationAuthorization == nil { return nil }
internal/spacestore/fs.go:40-45        if f.mutationAuthorization == nil { return nil }
internal/configstore/fs.go:39-44       if f.mutationAuthorization == nil { return nil }
internal/workspacestore/fs.go:39       passes its own possibly-nil hook straight through
```

The guarantee the tests pin is "every mutation entry consults the hook **if one is
installed**" (`TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects`,
`internal/store/mutation_authorization_test.go:14`), not "a store without a hook cannot
mutate." Nothing — not a compile error, not a lint rule, not a test — fires when a
construction site omits the option.

To be fair to the work already done: the CLI path is wired correctly and thoroughly tested.
Task `6g63hhk3eddf` (completed 2026-09-22) installed the authorizer across all four
families, and `commandSafetyState.authorizeMutation` (`internal/cli/command_safety.go:52-60`)
genuinely fails closed on an *unbound* classification. This finding is about the adapter
default and about construction sites the CLI does not own.

One such site already exists. `internal/cli/completion.go:222` and `:259` build
`store.NewFS(root)` with neither `WithMutationAuthorization` nor
`WithPlanningIdentityReader`, both of which `root.go:449-452` supplies. Those two calls are
reads, so nothing is wrong today — but they demonstrate that the pattern is reachable and
unflagged. The same task's own adversarial review found the workspace boundary by
inspection ("Antigravity … independently identified the workspace boundary as residual
risk"), which is the review method this finding proposes to replace with a structural one.

The repo already models the fix. `core.NewService` (`internal/core/service.go:228-277`)
validates every selected planning-data capability against a `SourceSetProvider` witness and
returns `ErrIncompatibleCapabilities` **before** exposing the service
(`internal/core/source_set.go:44-68`). Two composition-time validation policies sit in the
same codebase pointing opposite ways: one refuses to construct, one silently degrades.

**Compounding cost:** `schema --json` revision 1.71 publishes 100 command records with an
`enforced` read-only/mutating capability, and CLAUDE.md tells agents the tag is machine-
readable. That claim is currently a property of one composition root's wiring, not of the
adapters. Epic 19 (`serve`) is an in-flight second composition root; each new one must
remember the option in four places, and a miss is invisible until a read-only path writes.
The cost of fixing it rises with the number of construction sites, which is the number the
epic exists to increase.

**Proposed ADR amendment:** to ADR-0008, under `## Amendments` — that the `safety`
capability published in the `--json` command surface asserts an adapter-enforced invariant,
and that any adapter capable of persisting planning data must therefore refuse construction
without an explicit authorization decision (a guard, or a recorded opt-out). Human decision;
nothing was edited.

**Recommendation:** Make the guard a required construction parameter rather than an optional functional option, in the shape core.NewService already uses: have NewFS return (*FS, error) and reject a nil authorizer, with an explicit store.WithoutMutationAuthorization() opt-out for the read-only cases that genuinely need it (completion). Minimum viable alternative: a test that walks every store.NewFS construction site in internal/ and asserts each supplies an authorizer.

#### M3. The cli/tui/core/store layering is the repo's central architectural rule and has no ADR · **Status:** open

**File:** planning/adrs/ (absent); docs/ARCHITECTURE.md:11-76 | **Component:** planning/adrs
**Effort:** S · **Urgency:** soon

**Class:** ADR gap
**Anchored to:** ADR-0001 §"Append-only after acceptance (amendments, not edits)"; this routine's own reconciliation mandate.

`planning/adrs/` holds eight ADRs. None of them establishes the layering this lens exists to
audit. ADR-0003 decides storage identity and layout; ADR-0004 sync; ADR-0005 the registry;
ADR-0006 Threads; ADR-0007 vocabularies; ADR-0008 the machine contract. ADR-0006 §5 and §9
come closest — they fix a package boundary for *dependency analysis* ("Dependency analysis
belongs in pure core/domain-facing code…", "Introduce narrow consumer-owned dependency and
Thread persistence ports") — but they presuppose the hexagon rather than establish it.

So the repo's most-cited architectural rule lives in exactly two places, and neither is a
decision record:

- `docs/ARCHITECTURE.md` — an ordinary document, freely editable, with no append-only
  discipline and no `supersedes`/`superseded_by` chain.
- `.golangci.yml` — the enforcement, which M1 shows is narrower than the rule.

CLAUDE.md calls these the "non-negotiables" and points at `docs/ARCHITECTURE.md` as "the one
-screen orientation". ADR-0001 §"Append-only after acceptance" exists precisely so that a
settled decision cannot be quietly revised; the layering currently gets none of that
protection.

The symptom is already visible in the document. `docs/ARCHITECTURE.md:83-96` carries a
dated classification table — "The 2026-08-21 audit of direct primary-to-secondary edges
classified the remaining ones as follows" — with per-edge dispositions and the conclusion
"No new architecture task is needed for these edges." That is a decision record in
substance: a judgement, made on a date, that later work is expected to respect. It sits in a
file any change can rewrite without leaving a trace, and task `6g63hjm7cp6w` is already
filed against a *different* stale date-stamp in the same document.

**Compounding cost:** the gap costs nothing while one person holds the model in their head
and one adapter exists. Epic 19 splits the module into "core + bundled apps" and adds a
third primary adapter; epic 28 adds entity kinds; epic 21 has two in-flight tasks
(`6gcwcf8gzn50`, `6gcwcf8rxe72`) that move the boundary itself. Each of those will need to
cite an authority for what the boundary *is*, and will instead cite a paragraph that the
previous task may have edited. Writing the ADR while the rules are uncontested is cheap;
writing it during a module split means re-litigating settled questions with a second adapter
already depending on the answers.

**Proposed ADR amendment:** none to an existing ADR. Proposing a **new** ADR — provisionally
ADR-0009, "Primary adapters over a shared core; the filesystem is a secondary adapter" —
recording: (1) the layer roles and the inward dependency direction; (2) ports are defined at
the consumer in `internal/core` and implementations return concrete types; (3) the
composition-root exception and its intended narrowing; (4) DI through one per-invocation
container with no package-level state and all output through injected writers; (5) that the
direction is enforced executably, with `.golangci.yml` named as the mechanism. Human decision
— nothing in `planning/adrs/` was created or edited by this run.

**Recommendation:** Write one ADR that ratifies what docs/ARCHITECTURE.md already describes and .golangci.yml already enforces: the primary-adapter / core / secondary-adapter layering, consumer-owned ports, the composition-root exception, and DI with no package-level state. It is a consolidation of settled practice, not a new decision, so it can be accepted on the same day it is written. ARCHITECTURE.md then links to it and keeps only the current package map.

#### M4. Shell completion re-derives the flat entity layout and still documents the retired status-as-directory model · **Status:** tracked by 6gcwcf8gzn50

**File:** internal/cli/completion.go:181-182,206-207,222,229,239,245-247,259,266 | **Component:** cli/completion
**Effort:** S · **Urgency:** eventually

**Class:** implementation drift
**Anchored to:** ADR-0003 §2 and §4; Ousterhout, *A Philosophy of Software Design* — information leakage is the "Same knowledge used in multiple places".

The flat id-led layout is owned by the store and exposed through a port built for exactly
this purpose. `store.WatchPaths` (`internal/store/fsstore.go:128-133`) satisfies
`core.Layout`, whose doc comment states the intent plainly: "The store owns the layout
convention, so consumers recover desired paths instead of rebuilding the flat
entity-directory shape themselves." That port exists *because* this knowledge escaped once
already and was pulled back — task `6fbj87001q7p`, "Put storage-layout knowledge back
behind the port", completed 2026-06-14, when the TUI watcher was reconstructing
`<root>/tasks/<status>` itself.

The completion path rebuilds it anyway:

```
completion.go:229   filepath.Glob(filepath.Join(root, domain.TasksDir, "*.md"))
completion.go:239   filepath.Glob(filepath.Join(root, domain.ThreadsDir, "*.md"))
completion.go:266   filepath.Glob(filepath.Join(root, domain.AuditsDir, "*.md"))
completion.go:222   store.NewFS(root).ListTasks()    — a second FS, outside the composition root
completion.go:259   store.NewFS(root).ListAudits()
```

The predicted cost of a duplicated model is that the copies diverge, and they did. ADR-0003
§2 retired the old invariant explicitly — "The `status == directory` / `bucket ==
directory` invariant … is **retired** (those docs update *with* the implementation, not
before)" — and the executable code was migrated correctly: `taskCompleter` parses
frontmatter precisely because status is no longer the directory, and says so at
completion.go:213-216. But three contract comments still assert the retired model:

- `completion.go:206-207` — "completes task slugs whose status (== their directory)"
- `completion.go:245-247` — "completes audit slugs whose bucket (== their directory)"
- `completion.go:181-182` — "Because status/bucket *is* the directory, the caller selects
  which dirs to glob to filter by state"

The third is the most misleading, because `slugsFromGlobs` is now reached only from epic
completion (`completion.go:288`), and epics were always flat — so the comment describes a
mechanism the function no longer implements for a caller that never needed it.

No behavioural defect follows from any of this today; `flatCompletions`
(`completion.go:34-74`) handles id-led stems and duplicate-slug disambiguation correctly.

**Compounding cost:** low and already bounded. The layout knowledge is duplicated in one
adapter, and the divergence it produced was in comments rather than behaviour. It is
recorded here because the trajectory matters — epic 28 adds entity kinds, and each new kind
must be taught to the completion copy as well as the store — and because it is the concrete
evidence behind this lens's "adapter re-deriving what core owns" question. Task
`6gcwcf8gzn50` already owns the structural fix and proposes the right shape (a parse-free
completion *capability*, not forcing completion through the semantic read, which would lose
the ability to complete a file whose frontmatter will not parse).

**Recommendation:** Already owned by task 6gcwcf8gzn50, whose acceptance criterion is exactly this: completion must offer malformed id-led local records without filesystem globs in the Cobra adapter. No new work proposed. The three stale contract comments are a smaller, separable cleanup that should not wait for that task — route to the next code-quality-audit run.

#### L1. Boundary rule file globs are single-level, so a future subpackage is unguarded · **Status:** open

**File:** .golangci.yml:21-22,37-38,54-55,71-76 | **Component:** build/lint policy
**Effort:** XS · **Urgency:** eventually

**Class:** unanchored (best-practice-only; adjacent to M1)
**Anchored to:** same depguard semantics as M1.

Every depguard rule scopes itself with a single-level glob:

```
domain-stays-inward                    **/internal/domain/*.go
core-owns-ports-not-adapters           **/internal/core/*.go
wire-stays-adapter-neutral             **/internal/wire/*.go
primary-adapters-use-application-seams **/internal/tui/*.go, **/internal/configui/*.go,
                                       **/internal/cli/render/*.go, **/internal/cli/prompt/*.go
```

`*.go` does not match nested directories, so a package added at `internal/core/graph/` or
`internal/tui/panels/` would be governed by no rule at all — including the strong
prefix-deny rules that M1 holds up as the correct pattern.

There is no violation today: `find internal/{cli,tui,core,domain,store,wire} -mindepth 1
-type d` returns only `internal/cli/prompt`, `internal/cli/render`, and test data
directories, and both subpackages are named explicitly in the primary-adapter rule. The
existing `internal/cli/render` and `internal/cli/prompt` entries are in fact evidence that
the single-level glob was already worked around once by hand.

**Compounding cost:** minimal in isolation, which is why this is Low. It is recorded only
because it is the same edit as M1 — one file, one review — and because M1 shows what
happens to a boundary rule that depends on someone remembering to extend it.

**Recommendation:** Change the guarded file globs from **/internal/<pkg>/*.go to **/internal/<pkg>/**/*.go so subpackages inherit their parent's rule. Fold into M1's edit — same file, same review.

## What audited clean

- **Zero third-party dependencies in `internal/core`** — satisfies ADR-0006 §5's "no
  graph-library type may cross the taskflow-owned analysis interface" more strictly than
  the clause requires.
- **Inward direction holds transitively** — no package under `core`/`domain`/`wire`/`store`
  reaches `cli`, `tui`, or `configui`; the TUI is deletable without touching core.
- **Ports are segregated by honest capability** — `TaskGraphMutationStore`,
  `TaskLifecycleMutationStore`, `ThreadPathSource`, and the three Thread mutation stores
  sit outside the aggregate `Store` so read-only and test adapters do not claim semantics
  they cannot provide (interface-segregation; `internal/core/store.go:96-190`).
- **Consumer-owned interfaces, concrete return types** — matches the Go Code Review
  Comments rule verbatim, and cites it in `internal/core/store.go:1-3`.
- **No presentation vocabulary in core** — a scan for render/colour/style/width/terminal/
  viewport/keybinding terms across `internal/core/*.go` returns only lifecycle uses of
  "terminal" and comments describing what adapters do. `service.go:359` states the split
  explicitly: "Colour and layout stay with the adapters; the partition does not."
- **The safety callback is framework-neutral** — secondary adapters receive `func() error`
  and never learn Cobra's `read-only`/`mutating` vocabulary
  (`internal/cli/command_safety.go:19-25`, `internal/store/fsstore.go:77-84`).
- **Composition-time validation exists where it matters most** — `validateSourceSet`
  (`internal/core/source_set.go:44-68`) rejects a missing, unstable, or mismatched
  source-set witness before the service is exposed, and `WorkspaceService.Open` propagates
  it. This is the fail-closed pattern M2 asks to extend.
- **DI with no package-level state** — `App` is constructed per invocation and threaded
  explicitly; all output flows through injected writers, including Cobra's own
  (`internal/cli/root.go:278-281`).
- **The registry stays advisory** — ADR-0005 §5 holds in `startDir`, and a depguard rule
  keeps `internal/config` from reading home scope.

## Related-task observations (propose-only)

- Scope note: `planning/tasks/6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md`
  vs finding M1 — its acceptance criterion reads "A mutation that imports `internal/store`,
  `configstore`, `spacestore`, or `workspacestore` from an ordinary CLI controller fails the
  standard lint suite." That enumerates four packages by name, which would reproduce the
  fail-open shape M1 identifies inside the very task meant to tighten the boundary.
  Proposing only that the criterion be satisfied with a prefix deny plus an allow-list, so
  a fifth adapter added later is denied by default. Not edited.
- Possibly reinforcing, not duplicating:
  `planning/tasks/6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md`
  would independently catch M1's probe scenario, because a new `tui -> <newpkg>` edge is
  absent from the documented graph block. Its "Out of scope" explicitly declines to change
  depguard policy, so the two are complementary; M1 stays open rather than being folded in.
- Evidence note: `planning/tasks/6g28rv8j9d9p-codify-package-dependency-fitness-rules-and-refresh-the-architecture-baseline.md`
  (completed 2026-08-21) records "Temporary negative probes confirmed all four new
  boundaries fail lint with the intended named diagnostic." That is accurate and M1 does
  not contradict it — a probe using an *enumerated* package passes for the primary-adapter
  rule. The gap is only visible with a package that is not on the list, which is why it
  survived verification. No status change proposed.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — ./bin/tskflwctl task new "Make the primary-adapter dependency rule deny-by-default" --epic 21-code-quality-architecture-hardening
- ○ M2 · open — ./bin/tskflwctl task new "Require a mutation authorizer at store construction instead of defaulting to allow" --epic 21-code-quality-architecture-hardening
- ○ M3 · open — ./bin/tskflwctl task new "Ratify the hexagonal layering and consumer-owned port rule as an ADR" --epic 21-code-quality-architecture-hardening
- → M4 · tracked — Tracked in planning/tasks/6gcwcf8gzn50-route-cli-planning-data-operations-through-application-ports.md
