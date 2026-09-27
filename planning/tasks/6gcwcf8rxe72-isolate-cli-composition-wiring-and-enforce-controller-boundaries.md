---
schema: 1
id: 6gcwcf8rxe72
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Confine concrete adapter construction so dependency rules can prevent ordinary CLI controllers from bypassing core.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, cli, depguard, ports]
created: "2026-09-23"
depends_on: [6gcwcf8gzn50]
updated_at: "2026-09-27"
audited: "2026-09-27"
---

# Isolate CLI composition wiring and enforce controller boundaries

## Objective

Make the source-level architecture rule match the intended runtime architecture. Top-level
`internal/cli/*.go` currently contains both the composition root and ordinary Cobra controllers, so
the whole package is exempt from adapter dependency restrictions. Isolate concrete construction in a
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

- [ ] Only the documented composition boundary imports concrete store/configuration adapters to
      construct the application.
- [ ] A mutation that imports `internal/store`, `configstore`, `spacestore`, or `workspacestore` from
      an ordinary CLI controller fails the standard lint suite.
- [ ] CLI integration tests can still compose real adapters while focused command tests use ports.
- [ ] Architecture documentation and the executable dependency policy name the same exceptions.
- [ ] No command behavior, machine schema, or startup/discovery semantics change accidentally.

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
