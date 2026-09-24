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
updated_at: "2026-09-23"
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
