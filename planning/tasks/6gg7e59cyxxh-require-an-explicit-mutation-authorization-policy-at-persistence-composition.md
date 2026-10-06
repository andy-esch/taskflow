---
schema: 1
id: 6gg7e59cyxxh
status: completed
epic: 21-code-quality-architecture-hardening
description: Require explicit fail-closed persistence authorization and preserve policy through late opening.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, ports, safety]
created: "2026-10-03"
updated_at: "2026-10-05"
depends_on: [6gcwcf8rxe72]
started_at: "2026-10-05"
completed_at: "2026-10-05"
---
# Require an explicit mutation authorization policy at persistence composition

## Objective

Make omission of a mutation policy explicit at persistence composition rather than silently
granting write access to an accidentally unguarded adapter.

## Evidence

Finding M2 in [the hexagonal-boundaries audit](../audits/6gefs7wbbcyt-2026-09-28-arch-hexagonal-boundaries.md)
documented that `store`, `configstore`, `spacestore`, and `workspacestore` authorized mutations when
their optional callback was absent. The pre-change binary supplied one invocation-scoped authorizer
to all four; composition regressions pinned this. This task closes the omission risk for future
callers, not a newly demonstrated bypass in the previously shipped CLI.

## Scope and design checkpoint

Inventory persistence constructors and their production/test callers, then choose a small explicit
policy contract: required authorization, read-only construction, and any deliberate unrestricted
fixture/library mode must be distinguishable. The compatibility/opt-out checkpoint is accepted
below. Keep Cobra vocabulary out of secondary adapters and core.

## Pre-change constructor and caller inventory (2026-10-05)

Started on `feat/explicit-persistence-authorization`, based on main after PRs #282 and #281.
PR #282 bookkeeping is bundled with this task, not opened as a separate planning PR.
The user approved both constructor migration and deliberate unrestricted mode on 2026-10-05,
before constructor changes. The policy below is accepted for implementation.

| Boundary | Current behavior | Caller/migration implication |
| --- | --- | --- |
| `store.NewFS` | Optional callback; missing or nil callback authorizes mutation. | 26 guarded mutation entries already cover callback denial; extend to missing/zero/read-only modes and both dry-run and write. |
| `configstore.New` | Missing or nil callback authorizes migration and preferences. | Binary supplies the guard; tests include both read-only loads and intentional writes. |
| `spacestore.New` | Missing or nil callback authorizes registry add/forget. | Binary supplies the guard; registry-write fixtures need an explicit writable choice. |
| `spacestore.OpenPlanningStore` | Constructs `store.NewFS(root)` without a guard. | Atlas summary source is read-only by contract; construct with an explicit read-only policy even if the registry adapter may write. |
| `workspacestore.New` | Carries an optional callback into subsequently opened planning stores. | Policy must survive direct/pointer workspace opening and every later TUI store; never choose a writable fallback. |
| `workspacestore.NewPlanningStore` | Requires config/discovery but accepts nil authorization. | Used by ordinary opening and workspace opening; preserve one-observation identity/watcher behavior while requiring policy. |
| `appwiring.compose` / `openPlanning` | Supply the real invocation callback to all writable adapter families. | Preserve lazy discovery and late command-safety binding; validate presence at construction, not the callback's authorization result. |

The production construction sites are in `appwiring/wiring.go`,
`workspacestore/{fs,planning}.go`, and `spacestore/fs.go`; no controller fallback remains.
All four adapter types are exported **within internal packages**, not a public importable SDK.
A constructor signature change therefore has a repository-local migration cost, not a published
external Go API promise. A text inventory finds 385 `NewFS(` occurrences across 70 test files;
that is a migration-size indicator, not an AST-resolved caller count. Most churn will be fixtures.

Baseline tests for store/configstore/spacestore/workspacestore/appwiring pass. Current tests
establish supplied-guard denial and propagation, but intentionally unguarded writable fixtures
do not establish safe omission. Registry/configuration denial tests currently emphasize preview;
extend them to real writes and prove no filesystem effects or callback execution after refusal.

### Accepted policy (2026-10-05)

- Require one explicit, framework-neutral policy argument at each persistence constructor.
  Prefer a small opaque value with guarded, read-only, and deliberately unrestricted modes over
  an optional callback, booleans, interfaces with typed-nil surprises, or a generic permissions
  framework. Zero policy and nil guarded callbacks are invalid, never aliases for unrestricted.
- Validate policy presence without invoking the authorizer or touching filesystem state.
  The invocation callback must still run at each mutation, including dry runs; construction
  occurs before the CLI binds command safety, so caching an allow decision would be incorrect.
- Make explicit read-only stores readable but reject every mutation API, including previews,
  before filesystem effects, planner callbacks, or editor/transform callbacks. A directly
  constructed zero-value adapter must also fail closed, not bypass constructor validation.
- Carry the same guarded policy through late workspace and ordinary store opening. Narrow
  summary readers use read-only policy, not the registry's write privilege.
- Accept repository-local constructor signature changes and update callers in this slice;
  do not preserve an implicit-allow compatibility route. Error-returning construction versus
  named validated helpers can follow the selected policy without introducing panics on normal
  operation paths. `ports.Bindings.Compose` currently has no error result: explicitly handle
  construction refusal rather than exposing a partial usable bundle or swallowing the error.
  Help/schema construction must remain repository-independent and metadata-only bindings valid.
- Expose a conspicuously named unrestricted mode for intentional writable fixtures
  and trusted embedding/tooling. Production CLI/TUI composition remains guarded; no automatic
  test/environment opt-out. Classify fixtures rather than making every test unrestricted, and
  keep test helpers local enough to avoid a `store` ↔ `testutil` import cycle.

### Decisions

1. Require explicit policy arguments and migrate all internal callers now. No compatibility
   signature or implicit read-only default remains.
2. Permit a conspicuously named unrestricted mode for deliberate trusted callers and writable
   fixtures; missing policies and nil guarded callbacks still fail closed.

Construction returns errors for invalid policies, without invoking the authorizer. CLI composition
propagates refusal without publishing partial services; metadata-only construction remains valid.

## Implementation and verification (2026-10-05)

- `core.MutationPolicy` is an opaque value with `GuardedMutations`, `ReadOnlyMutations`, and
  `UnrestrictedMutations`. Validation invokes neither the guard nor I/O; authorization retains
  guard errors and checks each mutation boundary, including previews. No nil interface is accepted.
- `store.NewFS`, `configstore.New`, `spacestore.New`, `workspacestore.New`, and the observed-planning
  helper require policy and return construction errors. Optional authorization options are removed;
  direct zero-value adapters also deny writes. Existing locks, CAS, identity readers, and domain
  validation remain unchanged.
- Binary composition uses one invocation-scoped guarded policy. Atlas summary stores always select
  read-only policy; direct/pointer workspace stores retain the original guarded/read-only choice.
  Failed `ports.Bindings.Compose` results publish no partial services. Operational commands refuse
  the cause; help/version/schema/completion scripts remain usable, and entity completion stays quiet.
- Constructor fixtures now make deliberate read-only or writable choices. The generic test-only
  `testutil.Must` unwraps known-valid fixture construction without importing adapters or supplying
  a default policy. Invalid-construction tests inspect errors directly.
- Regression coverage includes constructor omission/nil guards, all 26 FS mutation entries across
  deny/read-only/zero modes and preview/write, configuration and registry effects, widened Atlas
  capabilities, changed late-opening guards, and CLI partial-bundle/metadata/fallback behavior.
- `go test -race -timeout 180s ./...`, `just lint`, and `just build` pass. Generated CLI docs and
  schema comments match checked-in output. Planning lint is clean; wire/schema revision stays 1.81.
- Fresh-binary throwaway smoke passed: init without registration, epic/tasks, dependency preview
  and commit, Thread creation, prerequisite completion/dependent start, frontier, lint, registration
  into an isolated home, and `status --all`. No real home registry or planning lifecycle was altered.

### Review reconciliation (2026-10-05)

Both independent reviews are reconciled and closed with no new in-scope defects. Codex's
32 compiler-valid probes and independent populated-record/full-bundle checks supplied stronger
hostile evidence; Antigravity independently verified policy propagation and fallback refusal.
Permanent tests now cover populated no-ops/callbacks across all 26 mutation entries, invalid
constructor options, and failed composition with every real service, original error classes,
isolated repo/home snapshots, metadata/completion, and subsequent invocation independence.
Report limitations and inventory corrections are recorded in the audits rather than overstated.

The pre-existing markerless recovery instruction remains in
[its CLI-contract followup](6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md),
with the second executed reproduction added as evidence. No duplicate or new Thread blocker.
Full normal/race tests, lint, build, generated CLI/schema-comment comparisons, and planning/audit
lint pass after the added tests. The task is completed locally, not yet merged or released.
The final guarded-contract task remains the next Thread member, outside this slice; bookkeeping
for merged PR #282 stays bundled with this work.

## Acceptance criteria

- [x] The chosen policy and compatibility treatment are recorded before constructor changes.
- [x] Missing/typed-nil policy cannot silently authorize a production persistence operation.
- [x] Legitimate read-only callers and intentionally writable tests have explicit supported modes.
- [x] The policy reaches stores opened later through workspace/configuration/registry services.
- [x] Tests cover denied dry-run and committing calls, and current CLI/TUI behavior remains intact.

## Related

- [Composition isolation](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md)
- [Adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- Implementation review briefs: [Codex](../audits/6ggw69vdfk1z-2026-10-05-explicit-persistence-authorization-implementation-codex.md)
  and [Antigravity](../audits/6ggw69vpds7h-2026-10-05-explicit-persistence-authorization-implementation-antigravity.md).

## Out of scope

- A new global service locator or application-wide permissions framework.
- Replacing repository guards, CAS, or domain eligibility policy with authorization callbacks.
