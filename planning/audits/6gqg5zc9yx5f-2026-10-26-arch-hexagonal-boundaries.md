---
schema: 1
id: 6gqg5zc9yx5f
bucket: open
area: arch-hexagonal-boundaries
date: "2026-10-26"
updated_at: "2026-10-26"
---
# Weekly Architecture Audit: hexagonal-boundaries — 2026-10-26

> Create findings through `tskflwctl audit finding new`; update them through
> `tskflwctl audit finding` so identity, status, resolution, and managed candidate
> metadata stay synchronized. Never hand-edit those fields.
> Architecture audits are propose-only — no code, docs, or ADR edits.

Routine: `weekly-architecture-audit` · lens `hexagonal-boundaries` · ISO week
`2026-W44` (mod 4 = 0). On-schedule Monday run.
ADRs consulted: ADR-0006 §5/§9 and amendment 2026-08-31 §1/§3/§4/§5/§6, ADR-0008 §5 and
its exit-code taxonomy, ADR-0007 §2, ADR-0001 (amendment policy) ·
Files read: `internal/core/store.go`, `internal/core/service.go`, `internal/core/source_set.go`,
`internal/core/workspace.go`, `internal/core/dependency_source.go`, `internal/cli/ports/runtime.go`,
`internal/appwiring/wiring.go`, `internal/appwiring/local_sources.go`, `internal/store/fsstore.go`,
`internal/workspacestore/fs.go`, `internal/cli/root.go`, `internal/cli/exit.go`,
`internal/tui/action.go`, `.golangci.yml`, `docs/architecture/boundaries.md` ·
Sources cited: 5

`just` is not installed in this runner, so the build was `go build -o bin/tskflwctl
./cmd/tskflwctl` and validation used the recipes' underlying commands directly
(`go build ./...`, `go test -race ./...`, `golangci-lint run ./...`). Recorded because
the spec names `just` first.

## Executive summary

The hexagon's **direction** is sound, and this audit re-verified it rather than assuming
it: `internal/core` imports only `internal/domain` and `internal/id`, nothing inward
reaches a primary adapter, output writers are injected on every production path, and
every optional port the filesystem adapter satisfies carries a compile-time witness.
Last cycle's findings landed — the enumerated depguard allowlist and the single-level
file globs are both fixed, and the mutation-policy guard is now a required constructor
parameter.

All three Medium findings this cycle are one seam: **the capability set is enumerated in
four hand-maintained places and declared in none.** `core.Service` lists 21 capability
fields (service.go:17), `validateSourceSets` checks 20 of them (service.go:408),
`WorkspaceSource` exposes 7 of them as independently injectable (workspace.go:29), and
the construction test enumerates 19 options (source_set_test.go:63). Nothing ties those
lists together, and the 34 "capability unavailable" guards they imply split 16/18 on
whether they wrap a domain sentinel.

M1 is the one with a reproduced consumer-visible consequence: the same failure class
exits 11 from one guard family and 1 — published as "unclassified" — from the other.
M2 has the largest compounding cost: guarded write capabilities cannot be composed
independently on the workspace route, which is the one ADR-0006's own 2026-08-31
amendment removed for reads. M3 is XS and is the finding that would have caught the
class. Read M1 first, then M2.

The two Low findings are separate and small: a mutable package global in the TUI that two
production constructors read (L1), and the one closed domain vocabulary still spelled by
hand in an adapter (L2, already owned by task `6gcqz5aqt2sg`).

## State of the architecture (this lens)

**Direction is enforced, not aspirational.** `go list` over the production graph matches
the documented map in boundaries.md exactly: `domain -> id`, `core -> domain, id`,
`wire -> core, domain`. Nothing under `core`, `domain`, `wire`, or `store` transitively
reaches `internal/cli`, `internal/tui`, or `internal/configui`; the TUI could be deleted
without disturbing core. Four depguard rules cover the pure layers and both primary-adapter
families, and since last cycle's fix they are default-deny: `list-mode: lax` plus a blanket
`github.com/andy-esch/taskflow/internal` deny, with `$`-anchored exact allowances
(.golangci.yml:74-104, :118-152). A brand-new internal package is therefore denied to an
adapter by default instead of waiting to be added to a blacklist.

**Ports are consumer-owned and genuinely segregated.** store.go:1-3 states the rule in the
package comment ("Interfaces are defined here, at the consumer") and the file delivers it:
the aggregate `Store` (store.go:435) composes only the four entity persistence ports, while
graph reads, graph mutation, graph repair, lifecycle mutation, Thread reads, Thread
creation/mutation/apply, lint reads, audit snapshots, completion, repair, link checks,
layout, and five per-entity path sources are each their own interface. This is interface
segregation taken seriously, and it is why a read-only or split adapter is expressible at
all. It is also the source of the bookkeeping the findings below describe: 20 narrow ports
is 20 things a composition root may or may not supply.

**Capabilities are discovered, not declared.** The production composition root is three
lines: `workspacestore.NewPlanningStore(...)` then `core.NewService(fs)` with no options
(appwiring/wiring.go:61-72). Every optional capability arrives by type assertion against
the aggregate store (service.go:325-383) — 19 consecutive `if x, ok := store.(Port); ok`
blocks. That is the stdlib's own feature-detection idiom, and taskflow pairs it correctly
with compile-time witnesses: `store/fsstore.go:84-99` declares `var _ core.X = (*FS)(nil)`
for all twelve otherwise-dynamic ports, plus the guarded mutation ports beside their
implementations (store/threadapply.go:12, store/graphmutation.go:15, and six more). A
signature drift becomes a compile error instead of a silently absent capability.

**The second composition root is shaped differently from the first.** `WorkspaceService.Open`
(workspace.go:107-115) builds its Service with seven explicit options — graph reads and the
five entity path sources plus Thread reads/paths — because `WorkspaceSource` has a typed field
for each (workspace.go:29-41). The remaining thirteen capabilities have no field and can only
arrive through `WorkspaceSource.Store`. The production adapter sets every field to the same
`*store.FS` value (workspacestore/fs.go:42-46), so nothing is missing today; the asymmetry is
in what a *different* adapter could supply.

**Absent capabilities are reported two different ways.** Core guards every optional capability
before use, and the guards split cleanly by family. The optional local-navigation and
maintenance family wraps `domain.ErrValidation` (16 guards: selected_entity_path.go:16-68,
service_task.go:378, service_epic.go:450, service_audit.go:113, service_research.go:161,
service_thread.go:411, completion.go:52, selected_task_creation.go:60,
planning_maintenance.go:60-93). The portable-read and guarded-mutation family returns a bare
`fmt.Errorf` with no sentinel (18 guards: service_thread.go:88,191,245,385,
service_thread_apply.go:14,17,50, service_task.go:40,353,652, finding.go:156,301,362,
service.go:581,713,716, dependency_operations.go:141, dependency_repair.go:629).

**Primary-adapter discipline holds, with one exception.** No production code writes to
`os.Stdout`/`os.Stderr` outside `cmd/tskflwctl/main.go:24,58,69` and three terminal-capability
probes (`lipgloss.HasDarkBackground`); every other write goes through an injected `io.Writer`
(cli/root.go:28-31). No `View` method calls a service. Package-level state in `internal/cli`,
`internal/tui`, and `internal/configui` is immutable lookup tables — except
`internal/tui/staged_creation.go:20`, a mutable `creationBlinkSpeed` global read by two
production constructors (staged_creation.go:25-27, entity.go:436-438) and written by a test
`init()`.

**Core's one remaining filesystem vocabulary is bounded and tracked.** `path/filepath` appears
twice in core: `displayPath` applies `filepath.ToSlash` to `source.LocalPath`
(dependency_graph.go:476,668-672) — the explicitly local field, not the neutral `Location` —
and `defaultSpaceID` derives a registry label from a directory basename
(space_registry.go:224-226). Neither parses a neutral location for identity, so ADR-0006's
amendment holds. One nuance worth recording without inflating it into a finding:
`TaskGraphSourceRef` (dependency_source.go:28-33) carries `Location` and `LocalPath` but no
`LocationIsPath`, so core infers path-ness by string equality (dependency_source.go:228)
where `RecordSource` (entity_read.go:26-32) and `ThreadReadProblem` (store.go:183) take it
as adapter-declared. The broader stdlib-and-framework guard for pure packages is already
owned by task `6ggxymjbf86r`.

## ADR reconciliation

### ADR-0006 §5 — no third-party graph type crosses the owned analysis interface

**Quoted:** "No graph-library type may cross the taskflow-owned analysis interface into
domain, persistence, or wire contracts."
**Implementation:** follows — `go list -deps ./internal/core` resolves to
`domain`, `id` and the standard library only; no graph package was adopted, so the clause
is satisfied by construction. `TaskGraph` and its projections are taskflow-owned values
(dependency_graph.go:50-75).
**Class if divergent:** n/a

### ADR-0006 §9 — dependency analysis is pure; narrow consumer-owned ports

**Quoted:** "Dependency analysis belongs in pure core/domain-facing code over task IDs and
statuses. Filesystem code parses and atomically mutates documents but does not decide
frontier, sound completion, or topological semantics. Introduce narrow consumer-owned
dependency and Thread persistence ports."
**Implementation:** follows — the store owns YAML surgery and the repository guard and calls
a pure core planner through control inversion (`TaskGraphPlanner`, store.go:118-131); the
planner "must not call a Store method" and the production adapter honours that.
`TaskDependencyWrite` names only a task and its canonical set (store.go:97-104).
**Class if divergent:** n/a

### ADR-0006 amendment 2026-08-31 §1 — task DAG reads use an independently injectable capability

**Quoted:** "`core.Service` accepts the narrow task snapshot source independently from
`ThreadStore`. Complete adapters default both capabilities from their aggregate store,
preserving ordinary CLI construction, while workspace and future composition roots may
provide distinct implementations."
**Implementation:** follows for reads — `WithTaskGraphSource` / `WithThreadStore` exist
(service.go:88-100, 258-273), the aggregate default is retained (service.go:360-364), and
`WorkspaceSource` exposes both as fields. The amendment's stated motivation — that a split
or read-only adapter should not "implement unrelated entity methods merely to project a
Thread" — is **not** carried to the guarded write capabilities on the workspace route: they
have no `WorkspaceSource` field and reach the Service only through the aggregate `Store`
(workspace.go:29-41 vs. 107-115). See M2.
**Class if divergent:** ADR gap — the amendment split reads deliberately and is silent on
whether write capabilities are independently composable.

### ADR-0006 amendment 2026-08-31 §4 — paired adapters must address compatible sources

**Quoted:** "Paired adapters must guarantee that the later task snapshot is at least as new
as the Thread snapshot; a lagging replica or split backend that cannot provide that order
must coordinate a compatible snapshot behind the two ports."
**Implementation:** follows, with an enforcement hole. `SourceSetID` plus
`SourceSetProvider` (source_set.go:14-30) is the composition-time witness, and
`validateSourceSet` (source_set.go:41-67) rejects a missing, empty, unstable, or foreign
witness. The check runs over a hand-written list of 20 capability names
(service.go:408-430); a 21st capability field already exists outside it. See M3.
**Class if divergent:** ADR gap — the ADR mandates compatible paired sources and is silent
on what guarantees every capability is actually paired-checked.

### ADR-0006 amendment 2026-08-31 §3/§5 — neutral projections; `graphfmt` imports no framework

**Quoted:** "Cobra, Bubble Tea, HTTP, filesystem, and third-party graph-library types do not
cross the core or wire contracts." · "`internal/graphfmt` owns pure deterministic Mermaid and
DOT formatting and format-specific escaping … and imports no CLI/TUI/HTTP framework."
**Implementation:** follows — `go list -deps ./internal/graphfmt` resolves to `core`,
`domain`, `id` and the standard library. Worth noting that nothing *enforces* it: the
depguard rules cover `domain`, `core`, `wire`, the primary adapters, `cli/ports`,
`appwiring` and `config`, but not the presentation or secondary-adapter packages. The
documented import map (boundaries.md:46-75) catches any new edge as drift, which is a
change detector plus human review rather than a policy gate — and boundaries.md says so
explicitly ("observed imports are not an allowlist").
**Class if divergent:** n/a (compliant; enforcement gap recorded, not filed — the detector
does fire)

### ADR-0006 amendment 2026-08-31 §6 — portable core readers never parse location

**Quoted:** "Kind and recoverable identity are authoritative; location is optional context
and is never parsed by core."
**Implementation:** follows — see the filesystem-vocabulary paragraph above. Core's two
`filepath` uses touch an explicitly local field and a registry label, never a neutral
`Location`.
**Class if divergent:** n/a

### ADR-0008 §5 — consumer-facing output remains agent-usable

**Quoted:** "Consumer-facing output remains agent-usable."
**Implementation:** drift at the capability boundary. The published exit-code taxonomy
(cli/exit.go:33-47) is a closed vocabulary in which `1` means "an unclassified command,
filesystem, or flag-parsing failure occurred" (exit.go:44), and classification is driven
entirely by `domain.Classify` over wrapped sentinels (exit.go:66-78). Because 18 of 34
absent-capability guards wrap nothing, the same failure class reaches an agent as two
different codes. Reproduced: `TaskPathByID` on a storeless service exits **11**,
`ListThreadViews` on the same service exits **1**. See M1.
**Class if divergent:** ADR gap — ADR-0008 publishes the taxonomy and the agent obligation
but does not say how an absent optional capability must classify.

### ADR-0007 §2 — words that mean the same thing in two places are spelled once

**Quoted:** "`domain/resolution.go` holds the shared pool … A test fails if a shared word is
spelled differently in either set."
**Implementation:** follows within its scope. ADR-0007 scopes itself to two body-level
vocabularies (finding status, criterion state) and those are derived, not transcribed. The
closed epic-status vocabulary is domain-owned (`domain.AllEpicStatuses()`, epic.go:66-69)
and every consumer derives from it — CLI completion (cli/epic.go:192,310), `schema`
(cli/schema.go:108), the wire DTO description (pinned by
wire/schema_descriptions_test.go:49-72), and the TUI's own epic creation picker
(tui/epic_create.go:51) — except the TUI action menu, which spells the three words inline
(tui/action.go:63-67). See L2.
**Class if divergent:** ADR gap — §2 decided the principle for two body fields and is silent
on frontmatter-field vocabularies consumed by adapters.

### No ADR governs the layering itself

Still true, and unchanged since last cycle. The cli/tui/core/store layering, consumer-owned
ports, the composition-root exception and the DI rules live only in
`docs/architecture/boundaries.md` and `.golangci.yml`. boundaries.md:160-162 says so
itself ("this extraction does not enact a new decision"). Owned by `6gg7e59mm68g`
(ready-to-start), whose scope already names "source-set checks, and explicit invocation DI"
— which makes it the natural home for the policy halves of M2 and M3.

## Best-practice comparison

### Optional capabilities: name the complete set, make absence typed, make wrapping explicit

**Source:** https://raw.githubusercontent.com/golang/go/master/src/net/http/responsecontroller.go
**Guidance:** The standard library's canonical optional-capability surface does three things
in one doc comment: it names the complete optional method set in a single place ("If the
ResponseWriter implements any of the following methods, the ResponseController will call
them as appropriate: Flush() / FlushError() error / Hijack() … / EnableFullDuplex() error"),
it defines an explicit protocol so a wrapper cannot silently hide a capability ("The
ResponseWriter should be the original value passed to the Handler.ServeHTTP method, or have
an Unwrap method returning the original ResponseWriter"), and it makes absence one typed
outcome ("If the ResponseWriter does not support a method, ResponseController returns an
error matching [ErrNotSupported]"), produced through one helper
(`func errNotSupported() error { return fmt.Errorf("%w", ErrNotSupported) }`,
responsecontroller.go:143-147).
**This codebase:** partial. Feature detection by assertion matches the idiom and is
compile-pinned, which is better than the stdlib's own position. The other two legs are
missing: there is no single declaration of the complete capability set (four partial lists),
and absence is 34 ad-hoc strings of which 18 carry no sentinel at all.
**Justified?** No. None of the three stdlib legs costs anything a single-user CLI can't
afford; the typed-absence leg is one sentinel and a mechanical edit.

### A typed "unsupported" sentinel is the documented Go convention

**Source:** https://pkg.go.dev/errors
**Guidance:** "ErrUnsupported indicates that a requested operation cannot be performed,
because it is unsupported." Callers test it with `errors.Is`, and "Functions and methods
should not return this error but should instead return an error including appropriate
context that satisfies `errors.Is(err, errors.ErrUnsupported)`, either by directly wrapping
ErrUnsupported or by implementing an Is method." It adds that functions "should document the
cases in which an error wrapping this will be returned."
**This codebase:** diverges for 18 of 34 guards. The wrapped 16 satisfy the shape (context
plus a sentinel), using `domain.ErrValidation` rather than `errors.ErrUnsupported` — a
reasonable local choice given that `domain.Classify` drives the exit contract. The bare 18
satisfy neither clause.
**Justified?** No. This is the finding (M1) and its recommendation.

### Guarantee dynamic interface satisfaction with a compile-time blank witness

**Source:** https://raw.githubusercontent.com/golang/website/master/_content/doc/effective_go.html
("Interface checks")
**Guidance:** Effective Go describes exactly taskflow's risk shape and its remedy: when a
type is detected by run-time assertion, "there are no static conversions that would cause
the compiler to verify this automatically. If the type inadvertently fails to satisfy the
interface, the JSON encoder will still work, but will not use the custom implementation." The
fix is a package-level blank declaration — `var _ json.Marshaler = (*RawMessage)(nil)` —
after which "Should the json.Marshaler interface change, this package will no longer compile
and we will be on notice that it needs to be updated."
**This codebase:** follows, thoroughly. store/fsstore.go:84-99 carries the blank witness for
all twelve dynamically-detected ports and fsstore.go:99 for selected-epic creation; eight
more sit beside the guarded mutation implementations. Every capability
`NewService` detects by assertion is compile-pinned on the production adapter.
**Justified?** n/a — this is the counterweight to M1/M2/M3: the mechanism taskflow *did*
build is the one Effective Go prescribes, and it is complete.

### Import policy should be default-deny, not an enumerated blacklist

**Source:** https://raw.githubusercontent.com/OpenPeeDeeP/depguard/master/README.md
**Guidance:** "Strict, at its roots, is everything is denied unless in allowed. Lax, at its
roots, is everything is allowed unless it is denied." And on matching: "Every `allow` and
`deny` entry is a **prefix** … End an entry with a dollar sign `$` to match that exact
package only."
**This codebase:** follows. The four boundary rules pair `list-mode: lax` with a blanket
`…/internal` deny and `$`-anchored allowances (.golangci.yml:74-104, :118-152), and the
file globs now cover `**/internal/<pkg>/**/*.go` as well as the direct level — the two
fixes last cycle's M1 and L1 recommended, verified present.
**Justified?** n/a — compliant.

### Exhaustiveness checking does not see a second hand-written list

**Source:** https://pkg.go.dev/github.com/nishanths/exhaustive
**Guidance:** The analyser "checks that switch statements in Go source code are exhaustive"
for a switch "in which the switch expression type is an enumerated type". Its unit of
analysis is the switch over a named type; a second, independently written list of the same
values — a slice literal, a map, a transition table — is outside what it can see.
**This codebase:** partial. The closed vocabularies are domain-owned and derived nearly
everywhere, and the one hand-spelled copy has a *validity* test
(tui/action_test.go:48-64 runs each epic transition through `domain.ValidateEpicStatus`).
Nothing checks *coverage*, which is the half a linter cannot supply and a three-line test
can. See L2.
**Justified?** Partly — a three-value vocabulary that has never changed is a small risk, and
the validity half is already pinned. The asymmetry is still worth one test.

## Tensions and trade-offs

**Twenty narrow ports is the right call and it has a bill.** taskflow is a single-user,
local-first, filesystem-backed CLI with one production adapter that implements every
capability, so all twenty could collapse into one `Store` with no user-visible loss today.
The project deliberately didn't, for a stated reason: the CLI and TUI are two real primary
consumers, portable test adapters exercise use cases without a filesystem, and epic 19's
`serve` is a third consumer in flight. Interface segregation is the correct answer for that
shape. The bill is bookkeeping — twenty things a composition root may forget — and the
findings below are that bill, not an argument against the ports. Each recommendation is a
coverage check or a sentinel, never a port count.

**"Latent" is the honest word for M1 and M2.** Because `*store.FS` implements everything
and `appwiring` passes it whole, no shipped command can reach a missing write capability
today. The one storeless Service the CLI really builds
(`core.NewBuiltinTemplateService()`, cli/template.go:42) serves only `template list/show`.
That is why neither finding is High. It is also why both are "soon" rather than
"eventually": the composition root that makes them simultaneously reachable —
a read-only or served adapter — is epic 19 work, and the cost of fixing them *after* that
adapter exists is paid in its error contract, not in a one-file edit.

**Enforcement has an outer edge, and it is a reasonable one.** Nothing stops
`internal/store` from importing `internal/cli`, or `internal/graphfmt` from importing Bubble
Tea; the depguard rules guard the pure layers and the adapters, not the presentation or
secondary-adapter packages. The documented import map turns any such edge into a test
failure, so the gap is "policy vs. detector", not "unguarded". Extending depguard to every
package would be cheap, but it is a decision for the policy ADR (`6gg7e59mm68g`) rather than
an audit finding, because boundaries.md deliberately distinguishes an executable policy from
a descriptive graph.

## Findings

#### M1. Half the absent-capability guards carry no domain sentinel, so the same failure reaches an agent as two different exit codes · **Status:** open

**File:** internal/core/service_thread.go:245 (system-wide: 18 of 34 guards) | **Component:** core capability guards + cli/exit
**Effort:** S · **Urgency:** soon

Core guards every optional capability before use, and the guards split cleanly into two
conventions by capability family rather than randomly.

The optional local-navigation and maintenance family wraps `domain.ErrValidation` — 16
guards at `selected_entity_path.go:16,29,42,55,68`, `service_task.go:378`,
`service_epic.go:450`, `service_audit.go:113`, `service_research.go:161`,
`service_thread.go:411`, `completion.go:52`, `selected_task_creation.go:60`, and
`planning_maintenance.go:60,65,68,93`.

The portable-read and guarded-mutation family returns a bare `fmt.Errorf` with no sentinel
at all — 18 guards at `service_thread.go:88,191,245,385`,
`service_thread_apply.go:14,17,50`, `service_task.go:40,353,652`,
`finding.go:156,301,362`, `service.go:581,713,716`, `dependency_operations.go:141`, and
`dependency_repair.go:629`.

Exit classification is driven entirely by `domain.Classify` over wrapped sentinels
(`cli/exit.go:66-78`), and an unclassified error falls through to `exitError` — code `1`,
published in the closed taxonomy as "an unclassified command, filesystem, or flag-parsing
failure occurred" (`cli/exit.go:44`). So the bare 18 are indistinguishable, at the process
boundary, from a flag-parsing mistake.

Reproduced in an isolated copy of this checkout against the storeless Service the CLI
really constructs (`core.NewBuiltinTemplateService()`, `cli/template.go:42`):

    task path resolution  -> exit 11 | validation failed: task path resolution is unavailable from this service
    thread reads          -> exit 1  | thread reads are unavailable from this store

Two guards, one failure class ("this service does not have that capability"), two process
contracts. ADR-0008 §5 requires consumer-facing output to remain agent-usable and agents
route on the code rather than parsing text (`cli/exit.go:90-93`); AGENTS.md requires
wrapping domain error sentinels rather than inventing adapter-specific error policy. Go's
own convention is one typed sentinel for exactly this case: `errors.ErrUnsupported`
documents that a function "should instead return an error including appropriate context
that satisfies `errors.Is(err, errors.ErrUnsupported)`", and `net/http` funnels every
absent optional capability through one helper
(`func errNotSupported() error { return fmt.Errorf("%w", ErrNotSupported) }`).

No shipped command can reach a bare guard today, because `*store.FS` implements every
capability and `appwiring` passes it whole — this is latent, not live. That is why it is
Medium rather than High.

**Compounding cost:** every new capability adds another guard, and with the convention
split 16/18 the next author has no majority to copy. The fix is one sentinel plus a
mechanical edit across 34 call sites while they are all in `internal/core`; after epic 19's
served or read-only adapter lands, every bare guard becomes simultaneously reachable and
the same change has to be made inside a published error contract, with consumers already
routing on the wrong code. Epic 28 adds a read, path, and mutation port per new noun, so
the guard count grows with the noun count.

**Recommendation:** Add one core sentinel (e.g. ErrCapabilityUnavailable wrapping domain.ErrValidation), route all 34 guards through a single helper that wraps it with the capability name, and add a test asserting every guard's error is classified.

#### M2. Guarded write capabilities are not independently composable on the workspace route · **Status:** open

**File:** internal/core/workspace.go:29-41,107-115 | **Component:** core/workspace + workspacestore
**Effort:** M · **Urgency:** soon

`core.Service` accepts 19 capabilities as explicit options. `WorkspaceSource` — the
adapter-neutral bundle the second composition root builds a Service from — has a typed field
for seven of them: `TaskGraphs`, the five entity `*Paths` sources, and `Threads`/`ThreadPaths`
(`workspace.go:29-41`). `WorkspaceService.Open` passes exactly those seven
(`workspace.go:107-115`).

The other twelve have no field and can only arrive by type assertion against
`WorkspaceSource.Store`: `TaskGraphMutationStore`, `TaskGraphRepairStore`,
`TaskLifecycleMutationStore`, `ThreadCreationMutationStore`, `ThreadMutationStore`,
`ThreadApplyMutationStore`, `SelectedEpicTaskCreationStore`, `LintSource`,
`AuditSnapshotSource`, `CompletionSource`, `Fixer`, and `Linter`.

Consequence: on the workspace route, an adapter that wants to supply only guarded Thread
apply must implement the whole aggregate `Store` — every task, epic, audit, and research
method — to get it through. That is verbatim the cost ADR-0006's 2026-08-31 amendment
removed for reads, in its own words: the pre-amendment shape "made a read-only service,
split persistence adapter, or future served interface implement unrelated entity methods
merely to project a Thread". The amendment split the read side and said nothing about the
write side, so the write side still has the defect the read side was fixed for.

Two further consequences fall out of the same asymmetry. First, the two composition roots
have different shapes for no stated reason: `appwiring` passes `core.NewService(fs)` with
zero options (`appwiring/wiring.go:70`) and relies wholly on discovery, while
`WorkspaceService.Open` is half explicit and half discovery — so "how does a Service get its
capabilities" has two answers. Second, `WorkspaceSource`'s own doc comment overstates what it
delivers: "Its capabilities stay separate even when one concrete filesystem value implements
all of them" (`workspace.go:17-19`) is true of the seven read fields and false of the twelve
write and aggregate-read capabilities.

Latent today: `workspacestore/fs.go:42-46` sets every field to the same `*store.FS`, which
implements all twenty capabilities, so nothing is missing in the shipped binary.

**Compounding cost:** the workspace route is the one epic 29's atlas uses for cross-space
opening and the one epic 19's served adapter would reuse, and the TUI's staged creation
already drives guarded Thread/task/epic/audit/research creation through a workspace-opened
Service. Each new guarded capability (epic 28 adds creation and mutation ports per new noun)
widens the aggregate-only path by one, and each is cheap to add as a field now and expensive
later: once a served adapter exists, making a capability independently injectable means
changing the bundle type every caller already constructs, rather than adding a nil-defaulted
field to a type with one producer.

**Proposed ADR amendment:** extend ADR-0006's 2026-08-31 split to the write side — "every
capability `core.Service` accepts as an independent option is also independently supplied by
the workspace/served composition bundle; the aggregate store remains the documented default
for any capability a bundle leaves unset." A human decides whether the symmetry is worth the
field count; nothing in `planning/adrs/` was edited.

**Recommendation:** Give WorkspaceSource a typed field for each guarded write and aggregate-read capability the Service accepts as an option, keep the aggregate Store as the documented default for any left nil, and assert field-for-option parity in a construction test so a new port cannot be added to NewService alone.

#### M3. The source-set pairing check is a hand-maintained list, so a new capability is exempt by default · **Status:** open

**File:** internal/core/service.go:408-430 | **Component:** core composition checks
**Effort:** XS · **Urgency:** soon

`SourceSetID` is the composition-time witness that every independently injected capability
addresses the same planning corpus (`source_set.go:14-30`), and `validateSourceSet` correctly
rejects a missing, empty, unstable, or foreign witness (`source_set.go:41-67`). What it checks
is a literal slice of 20 `{name, value}` pairs typed out by hand (`service.go:408-430`). A
capability field that is not in that slice is silently exempt from the entire check.

This is the same defect shape as last cycle's M1 — a guard that works exactly and only for
the entries someone remembered to list — one layer in, where it protects a data-integrity
invariant rather than import hygiene.

Two probes in an isolated copy of this checkout, with no change to `internal/core` for the
first:

Reflecting over `Service`'s interface-typed fields finds 21; `validateSourceSets` lists 20.
The unlisted one is `templates` (`TemplateSource`). Injecting a `TemplateSource` that
publishes a *different* `SourceSetID` alongside an aggregate store is accepted, while the
identical move on a listed capability is refused:

    control (checked capability, foreign corpus): REFUSED -> incompatible planning capabilities: aggregate store and task graph reads address different source sets
    control (checked capability, same corpus):    ACCEPTED
    probe   (templates, foreign corpus):          ACCEPTED

`templates`' exemption is *correct today*: it defaults to `builtinTemplates{}`
(`service.go:379`), which is the compiled-in domain registry and addresses no corpus at all,
so demanding a witness of it would break `template list` outside a planning repo
(`service.go:488-493`). The problem is that nothing distinguishes "deliberately exempt" from
"forgotten", and epic 22's repo-local template source — "a repo-local source layers over it
later" (`core/template.go:22`) — is exactly a corpus-addressing implementation that would
land in the exempt slot.

Second probe: adding a 22nd capability port the ordinary way — a port, a `With…` option
shaped like its 19 siblings, and an aggregate-store discovery block — but omitting the
`validateSourceSets` row compiles and passes `go test ./internal/core/...` and
`go test ./internal/tools/architecturecheck/...` with no failure. The construction test that
would be the natural place to catch it enumerates the 19 options by hand too
(`source_set_test.go:63-96`), so it grows the same blind spot.

**Compounding cost:** XS now — one reflection-based test over `reflect.TypeOf(Service{})`,
which is a mechanism this repo already uses for exactly this purpose elsewhere
(`wire/schema_descriptions_test.go` pins a reflected schema against
`domain.AllEpicStatuses()`). The capability count has grown from a handful to 21 over the
audited history and epics 19, 22, and 28 each add more, so without the check the odds of a
silent exemption compound with every capability — and an exemption is only observable the
day a split composition root pairs the wrong corpus, which is the failure the witness exists
to make impossible.

**Proposed ADR amendment:** record in the composition policy (the ADR task `6gg7e59mm68g`
already scopes "source-set checks") that the witness requirement applies to every
independently injectable planning-data capability, that exemptions are explicit and
justified, and that coverage is checked rather than maintained. Nothing in `planning/adrs/`
was edited.

**Recommendation:** Make validateSourceSets derive its rows by reflecting over Service's interface-typed fields with an explicit, documented exemption list, or keep the literal list and add a test that reflects over those fields and fails when one is neither listed nor exempted.

#### L1. A mutable package global in the TUI is read by production constructors · **Status:** open

**File:** internal/tui/staged_creation.go:20 | **Component:** tui composition
**Effort:** XS · **Urgency:** eventually

`var creationBlinkSpeed time.Duration` (`staged_creation.go:20`) is a mutable package-level
variable read by two production constructors — `newCreationInput` (`staged_creation.go:25-27`)
and the entity-registry input builder (`entity.go:436-438`) — and written only by tests,
including a package-wide `init()` in `staged_creation_helpers_test.go:11-13`.

boundaries.md's composition section makes a positive claim this contradicts: "No
package-global dependency defaults, `cmd.Context()` service locator or fallback local store
exists in controllers." The sentence names controllers, and this is the TUI, which is why
this is Low and not Medium — but the TUI is the other primary adapter, and every other
package-level `var` in `internal/cli`, `internal/tui`, and `internal/configui` is an
immutable lookup table (transition tables, key maps, label arrays, compiled regexps).
This is the single mutable one.

It is not a live bug: no `internal/tui` test calls `t.Parallel()`, so there is no race today,
and the production default is the documented zero value that preserves Bubbles' ~530ms
cadence — which `TestNewCreationInputPreservesProductionCursorDefaults`
(`staged_creation_test.go:17-45`) pins deliberately.

**Compounding cost:** small but real and one-directional. It is the package's first
test-seam-as-global, so it is the precedent the next widget-timing seam copies; and it is a
standing obstacle to `t.Parallel()` in `internal/tui`, the repo's largest test package, since
adopting parallelism there would require finding and fixing this first under `-race`. Fixing
it now is a two-field change with two call sites.

**Recommendation:** Carry the blink speed on the creation/entity widget struct (or a textinput factory held by the Model) and have tests construct it with the fast value, instead of a package-level var the production constructors read.

#### L2. The TUI epic lifecycle table is the one hand-spelled copy of a closed domain vocabulary, and only its validity is pinned · **Status:** tracked by 6gcqz5aqt2sg

**File:** internal/tui/action.go:63-67 | **Component:** tui/action + domain vocabulary
**Effort:** XS · **Urgency:** eventually

Task and audit lifecycle tables in the TUI are derived from the shared domain registry —
`fromDomain(domain.TaskTransitions())` and `fromDomain(domain.AuditTransitions())`
(`action.go:48,56`) — so they cannot drift. `epicTransitions` (`action.go:63-67`) spells its
three destinations as string literals instead, with a documented reason: epics "have no CLI
verb vocabulary to share" (`action.go:24-26`), so there is no `domain.EpicTransitions()` to
read.

The *statuses* are domain-owned regardless (`domain.AllEpicStatuses()`, `epic.go:66-69`), and
every other consumer derives from that function: CLI completion (`cli/epic.go:192,310`),
`schema` (`cli/schema.go:108`), the wire DTO description — pinned by a test that exists for
precisely this drift class, "Pin it to domain.AllEpicStatuses() so the drift fails loudly
here" (`wire/schema_descriptions_test.go:49-72`) — and the TUI's own epic *creation* picker
(`tui/epic_create.go:51`). The action menu is the one place the words are retyped.

Half the exposure is already guarded: `TestTransitionTablesValidStates`
(`tui/action_test.go:48-64`) runs every epic transition destination through
`domain.ValidateEpicStatus`, so a typo or a renamed status fails loudly. The unguarded half
is coverage — if `domain` gains a fourth epic status, it reaches CLI completion, `schema`,
the wire schema, and the TUI creation picker automatically, and silently never appears in the
TUI action menu. No test fails.

This is the gap a linter cannot close: `exhaustive` analyses switches over an enumerated
type, not a second hand-written slice literal, so a derived list or a matching test is the
only thing that sees it.

**Compounding cost:** genuinely small — a three-value vocabulary that has not changed, with
validity already pinned. It is filed because the TUI entity registry generalizes the pattern
("epics declare `epicTransitions`/`moveEpic`", `entity.go:159,205`), so the next entity whose
status is a frontmatter field rather than a directory inherits a hand-spelled table and the
same half-guard, and because the fix is three lines in a test that already iterates the
table.

**Recommendation:** Extend TestTransitionTablesValidStates with a coverage assertion that every domain.AllEpicStatuses() value appears as some epicTransitions destination, or derive the table from a verb map over AllEpicStatuses() the way taskTransitions derives from domain.TaskTransitions().

**Resolution:** Already owned by task 6gcqz5aqt2sg (ready-to-start, epic 21),
    whose scope and final acceptance criterion name exactly this work: pin the
    TUI's hand-typed epic transition and view tables to
    domain.AllEpicStatuses/AllAuditBuckets by coverage tests. No new work
    proposed; this audit only adds the coverage-vs-validity distinction and the
    wire/schema_descriptions_test.go precedent as evidence.

## What audited clean

- **Inward dependency direction, re-verified by `go list`.** `domain -> id`,
  `core -> domain, id`, `wire -> core, domain`; nothing inward reaches a primary adapter.
  Satisfies ADR-0006 §9 and boundaries.md's package map.
- **Default-deny import policy with exact-match allowances.** Last cycle's M1 and L1 fixes
  are present and correctly shaped per depguard's documented `lax` + prefix/`$` semantics;
  a new internal package is denied to an adapter by default.
- **Compile-time witnesses for every dynamically detected port.** store/fsstore.go:84-99 and
  eight siblings implement exactly Effective Go's "Interface checks" remedy, which is what
  keeps capability discovery-by-assertion safe here.
- **Control inversion at the guarded-mutation seam.** The store owns the repository guard
  and calls a pure core planner that "must not call a Store method" (store.go:118-131),
  satisfying ADR-0006 §9's split between parsing/atomicity and graph semantics.
- **Consumer-owned, segregated ports.** 20 capability interfaces instead of one god-Store,
  each documented with why it stays separate (store.go:133-137, 231-239) — ADR-0006
  amendment 2026-08-31 §1's stated goal.
- **Required mutation policy on every persistence constructor.** `NewFS` refuses a missing
  policy rather than defaulting to writable (store/fsstore.go:101-104) — last cycle's M2,
  verified fixed.
- **Framework-neutral invocation contracts.** `cli/ports` imports only the standard library,
  `core` and `design`, enforced by `list-mode: strict` (.golangci.yml:173-186); `ports.Browser`
  carries no Bubble Tea type.
- **Injected writers everywhere.** No production `fmt.Print*` or `os.Stdout` write outside
  `cmd/tskflwctl/main.go` and three terminal-capability probes; `View` methods call no service.
- **Lazy planning opening after flag parsing.** `OpenPlanning` is a closure on the bundle
  (ports/runtime.go:22-30) and completion defers composition until Cobra has parsed the
  completed command (cli/root.go:66-69) — boundaries.md's composition contract.
- **Core parses no neutral location.** Both `path/filepath` uses in core touch an explicitly
  local field or a registry label — ADR-0006 amendment 2026-08-31 §6.

## Related-task observations (propose-only)

- Policy home: `planning/tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md`
  already scopes "source-set checks, and explicit invocation DI". The policy halves of M2 and
  M3 belong there rather than in new design tasks; the code halves do not. No status change
  proposed.
- Adjacent, not overlapping: `planning/tasks/6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md`
  owns the stdlib/framework guard for pure packages, which is the general form of the two
  `path/filepath` uses recorded above. This audit files no finding against it.
- Already shipped, did not cover this surface: `6fgcr2402rbj` (centralize domain error
  classification) and `6gbn4g1bb7wr` (publish the complete exit-code taxonomy) are both
  `completed`. They built the mechanism M1 relies on and classified command failures; neither
  reached the absent-capability guards. Recorded so M1 is not read as a regression of either.
- FULL overlap, marked `tracked`: L2 is owned by
  `planning/tasks/6gcqz5aqt2sg-make-closed-vocabulary-switches-fail-closed-and-exhaustively-checked.md`
  (ready-to-start), whose final acceptance criterion already reads "The TUI epic and audit view
  tables and epic transitions are pinned to domain.AllEpicStatuses/AllAuditBuckets and the
  EpicStatus constants by coverage tests". Its `audit_sources` and body were annotated
  additively; no status, tier, priority, or scope was touched.
- PARTIAL overlap, L1 kept open: `planning/tasks/6gpxc9zegzzd-eliminate-ci-test-suite-bottlenecks-and-parallelize-slow-package-tests.md`
  (in-progress) is the task whose already-checked criterion introduced `creationBlinkSpeed`, and
  its still-open criterion — stateless `internal/tui` and `internal/domain` tests use
  `t.Parallel()` — is exactly the collision L1 names. The task owns the performance goal, not the
  seam: nothing in it says the test hook should stop being a package global. Worth the
  implementer's attention when that criterion is picked up; no status change proposed.
- Route to `code-quality-audit` if preferred: `internal/tui/staged_creation.go:20` is filed here
  as L1 because boundaries.md makes a positive claim about package-global state. If the reviewer
  disagrees that it is architectural, it is a clean XS code finding instead.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — tskflwctl task new "Classify absent optional capabilities through one core sentinel" --epic 21-code-quality-architecture-hardening --tags architecture,core,agents --effort S --priority high
- ○ M2 · open — tskflwctl task new "Make guarded write capabilities independently composable in WorkspaceSource" --epic 21-code-quality-architecture-hardening --tags architecture,core,ports --effort M --priority medium
- ○ M3 · open — tskflwctl task new "Make the source-set capability check cover every Service capability field" --epic 21-code-quality-architecture-hardening --tags architecture,core,tests --effort XS --priority medium
- ○ L1 · open — tskflwctl task new "Replace the TUI creation blink-speed global with constructor state" --epic 18-tui-bubble-tea-interactive-planning-browser --tags tui,tests,architecture --effort XS --priority low
- → L2 · tracked — Tracked in planning/tasks/6gcqz5aqt2sg-make-closed-vocabulary-switches-fail-closed-and-exhaustively-checked.md
