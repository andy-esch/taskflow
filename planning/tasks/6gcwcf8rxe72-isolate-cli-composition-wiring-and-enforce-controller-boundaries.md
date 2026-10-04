---
schema: 1
id: 6gcwcf8rxe72
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Confine concrete adapter construction so dependency rules can prevent ordinary CLI controllers from bypassing core.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, cli, depguard, ports]
created: "2026-09-23"
depends_on: [6gcwcf8gzn50]
updated_at: "2026-10-03"
audited: "2026-09-27"
audit_sources: [2026-09-28-arch-hexagonal-boundaries]
started_at: "2026-10-03"
---

# Isolate CLI composition wiring and enforce controller boundaries

## Objective

Make the source-level architecture rule match the intended runtime architecture. At the baseline,
`internal/cli/*.go` contained both the composition root and ordinary Cobra controllers, so
the whole package was exempt from adapter dependency restrictions. Isolate concrete construction in a
small wiring boundary, then make direct controller imports of persistence/configuration adapters a
lint failure.

## Scope

- Choose and document one composition location, such as `cmd/tskflwctl` or a narrowly scoped wiring
  package, without introducing a service locator.
- Move concrete adapter construction and framework launch wiring there while keeping command
  definitions testable with injected application capabilities.
- Extend depguard or an equivalent executable architecture test over ordinary CLI controllers.
- Inventory and explicitly classify any remaining primary-to-secondary edges.

## Acceptance criteria

- [x] Only the documented composition boundary imports concrete store/configuration adapters to
      construct the application.
- [x] Known persistence adapters and an unknown internal adapter fail standard
  lint from ordinary CLI controllers; nested source packages inherit the rule,
  and exact allowances do not admit children of permitted packages.
- [x] CLI integration tests can still compose real adapters while focused command tests use ports.
- [x] Architecture documentation and the executable dependency policy name the same exceptions.
- [x] No command behavior, machine schema, or startup/discovery semantics change accidentally.

Implementation criteria verified locally; both external reviews are reconciled. Lifecycle stays
**in-progress** pending merge; checked criteria and closed audits do not claim the work has shipped.

## Out of scope

- Splitting every large CLI source file or redesigning the `App` object for aesthetic reasons.
- Changing Cobra, Bubble Tea, or configuration formats.
- Forbidding concrete adapters from integration tests.

The existing import-graph task verifies that documentation matches the actual graph and explicitly
does not change policy. This task changes the policy boundary and its enforcement, so the two are
complementary rather than duplicate.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [CLI planning-data port migration](6gcwcf8gzn50-route-cli-planning-data-operations-through-application-ports.md)
- [Executable import-graph task](6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md)

## Sweep verification (2026-09-27)

Historical baseline; the implementation below closes the broad exception.

Automated weekly sweep. The premise holds and nothing was rescoped; no acceptance
criterion was ticked. Refs verified accurate (2026-09-27):

- *"the whole package is exempt from adapter dependency restrictions"* — confirmed
  verbatim in `.golangci.yml`, whose `primary-adapters-use-application-seams` rule covers
  `internal/tui`, `internal/configui`, `internal/cli/render` and `internal/cli/prompt`,
  and whose comment states the exemption explicitly: *"internal/cli/*.go is also
  deliberately absent: that package is today's composition root and is allowed to
  construct adapters and launch the TUIs."* The deny list there already names all four
  adapters this task's AC 2 wants enforced (`store`, `configstore`, `spacestore`,
  `workspacestore`) — so the rule body largely exists, it just does not cover
  `internal/cli/*.go`.
- AC 4 (*"documentation and the executable dependency policy name the same exceptions"*)
  is still unmet and still worth its own criterion: `docs/ARCHITECTURE.md:89` records the
  `cli -> store` exception as a prose table row, `.golangci.yml` records it as an absent
  file glob. Two different shapes, no cross-check.

### The surface is narrower than "the whole package" suggests

Measured today, only **two** of the 35 non-test files in `internal/cli/*.go` actually
construct a concrete persistence/configuration adapter:

- `root.go` — `spacestore.New` (`:261`), `configstore.New` (`:264`), `workspacestore.New`
  (`:269`), `store.NewFS` (`:446`).
- `completion.go` — `store.NewFS(root)` (`:222`, `:259`).

`completion.go`'s two sites are exactly what `depends_on`
([route-cli-planning-data-operations-through-application-ports](6gcwcf8gzn50-route-cli-planning-data-operations-through-application-ports.md))
removes, so once that lands this task's move is **`root.go` only**. The dependency edge is
therefore doing real sequencing work rather than being a nominal ordering, and the two
tasks compose without overlap. Recorded as an observation only — `effort: 1-2 days` is
left alone, since the depguard rule extension and AC 4's documentation reconciliation are
untouched by this narrowing.

## Progress Log

- 2026-09-27: automated weekly sweep — premise re-verified against `.golangci.yml` and the docs exception table; noted that concrete adapter construction is already confined to `root.go` + `completion.go`, so this task's move is `root.go` alone once its dependency lands.
- 2026-10-03: dependency implementation removes all completion filesystem imports/globs and direct
  repair/link calls. Preserve the deferred `App.CompletionService` factory when extracting wiring:
  Cobra parses the completed command's `-C`/`--space` after its initial hook. Portable completion
  injection and real `__complete` tests pin this boundary. `root.go` retains concrete composition;
  local watcher `Layout` and init/doctor/discovery exceptions still need explicit classification.

Reinforced by audit 2026-09-28-arch-hexagonal-boundaries: M1 and L1. The baseline's enumerated
deny-list and single-level file globs admitted unknown adapters/subpackages. The implementation
uses default-deny internal imports, exact allowances for primary adapters, and both direct and
recursive file patterns (the glob matcher does not treat the extra slash as zero directories).

## Implementation notes (2026-10-03)

- `cmd/tskflwctl` selects `appwiring.LocalBindings`; `cli/ports` owns the explicit invocation
  services, lazy planning opener, neutral presentation readers, and framework-free launch requests.
  The composition package never depends on the CLI root, and controllers cannot import it.
- Every command tree receives fresh services and its own authorization closure. Ordinary opening
  preserves the planning identity reader, source-set checking, and optional watcher layout. No
  discovery is added to command-tree construction; generators use empty metadata-only bindings.
- Completion still opens only after Cobra parses the completed command's target flags. Missing,
  failed, or incomplete injected capabilities never trigger local fallback; missing named services
  return explicit validation failures rather than panicking.
- Concrete TUI/config-editor options and execution live in `appwiring`; command gates remain in
  controllers. The only direct `config` imports retained in production CLI files are init topology
  and workspace checkout receipts, each still default-deny for persistence construction.
- The same lint edit closes unknown-adapter and nested-package gaps in TUI/config UI/render/prompt,
  domain, core, and wire policies. Integration fixtures explicitly select real binary wiring.

### Validation

`go test ./...`, `go test -race ./...`, `just lint`, and `just docs-check` pass. Generator output
does not change; planning lint and the Thread frontier remain healthy. A fresh binary opened and
quit both `ui` and `config edit` in an isolated throwaway planning tree, and rendered config/status
JSON successfully. This is a launch smoke test, not a claim to have retested every interactive edit.
Composition tests cover direct/pointer corpora, scoped authorization, identity
replacement before Thread apply, lazy opaque completion, populated-cwd missing bindings, failed
opening without partial publication, and missing named services.

In an independent checkout, formatted and compiler-valid probes made standard `just lint` report
**16 depguard violations**: all four known persistence adapters, `appwiring`, an unknown adapter,
a child of allowed `core`, nested CLI/TUI/domain/core/wire source packages, neutral CLI ports, and
both named exception files. No type-check or formatting failure stood in for boundary enforcement.
Probe files are temporary evidence, not production fixtures.

### Separately tracked hardening

- [Workspace Thread-apply identity parity](6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md):
  reproduced direct dry-run success versus workspace validation failure; fails closed and predates
  this extraction. Own the behavior change independently of this preservation-focused slice.
- [Explicit persistence authorization policy](6gg7e59cyxxh-require-an-explicit-mutation-authorization-policy-at-persistence-composition.md):
  audit M2; current composition is guarded, but constructor omission elsewhere still authorizes.
- [Durable dependency-policy ADR](6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md):
  audit M3; proposed for explicit user acceptance, sequenced in the documentation Thread.

## External review handoff

Historical handoff: each reviewer captured the uncommitted snapshot in a separate sandbox and was
instructed to transfer only their assigned audit. Both reports now have owner dispositions and are
closed; their original evidence remains tied to the captured pre-fix snapshots.

- [Codex](../audits/6gg7h65sfjrb-2026-10-03-cli-composition-boundary-implementation-codex.md): runtime
  identity, authorization, and compatibility lens.
- [Antigravity](../audits/6gg7h6621wwt-2026-10-03-cli-composition-boundary-implementation-antigravity.md):
  compiler-valid boundary escapes, incomplete bindings, and deceptive test-helper lens.

## Review reconciliation (2026-10-03)

- Codex M1: require all four browser services before either landing route; test each omitted service
  in repository and atlas starts. The normal full binary already supplied them; incomplete injections
  now fail consistently instead of launching a partially functioning UI.
- Codex M2: deny direct Bubble Tea imports in controllers and both named exceptions; invocation
  contracts allow only stdlib and exact core/design. Ten additional compiler-valid probes produced
  ten depguard errors under standard lint, including nested sources and Cobra/Lipgloss contracts.
- Codex L1 / Antigravity M1: observe real startup readers and actual adapter composition, use distinct
  populated direct/pointer/cwd corpora and exact watcher roots, and test chrome fallback against
  non-default repository/home themes. Failed explicit opening also cannot expose populated cwd tasks.

Seven runtime mutants now fail their named regressions: repository UI service-check bypass, eager
real composition opening, duplicate discovery, cwd store selection, repository/home chrome fallback,
and failed-opener cwd fallback. Planning/workspace authorization-loss probes also fail for the
missing authorizer itself after making their task fixtures valid. All probes ran in an independent
sandbox, were restored, and passed clean focused tests/lint afterward. Full uncached race tests and
generated-doc checks pass; both audit files record bounded evidence and owner resolutions.

The startup observation seam is per binding, not global; file-scoped lint prevents direct
config/userconfig imports in other production composition files. It protects declared startup
readers, not every syscall inside secondary adapters. Allocating repo-independent adapters remains
intentional. No new design decision or out-of-scope finding was needed beyond the three followups above.
