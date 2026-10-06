---
schema: 1
id: 6gg7e59mm68g
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Give the settled executable adapter boundary a durable decision record without duplicating the package map.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, docs, adr]
created: "2026-10-03"
depends_on: [6gcwcf8rxe72]
updated_at: "2026-10-03"
---
## Objective

Give the settled hexagonal dependency policy a durable decision record; keep the architecture guide
as the current package map instead of asking it to be both map and historical authority.

## Evidence

Finding M3 in [the hexagonal-boundaries audit](../audits/6gefs7wbbcyt-2026-09-28-arch-hexagonal-boundaries.md)
identifies layering/composition decisions documented only in `docs/ARCHITECTURE.md` and lint config.
The composition extraction now provides a concrete baseline rather than a broad CLI exception.

## Scope

- Record inward dependencies, consumer-owned ports, source-set checks, and explicit invocation DI.
- Name `internal/appwiring` and binary selection as concrete composition; document narrowly named
  local topology/checkout and UI embedding exceptions, not a blanket primary-adapter exemption.
- Distinguish an executable dependency policy from a descriptive import graph.
- Reconcile the guide and `.golangci.yml`; propose an ADR for user acceptance rather than marking
  it accepted automatically or introducing a new architecture framework.
- Resolve the enforcement floor for direct I/O and external framework imports: today's
  core/domain rules restrict repository-internal imports, while current production imports comply
  with the broader pure-layer intent. The [focused guard followup](6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md)
  follows this decision; harmless pure utilities are not automatically forbidden.

## Acceptance criteria

- [ ] A concise proposed ADR records the settled policy and explicitly identifies remaining choices.
- [ ] The guide links to the ADR and retains useful current ownership/import information.
- [ ] Its permitted edges and named exceptions agree with standard lint enforcement.
- [ ] Boundary tests and the import-graph task are linked as complementary evidence, not duplicated.

## Related

- [Composition isolation](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md)
- [Architecture guide restructuring](6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md)
- [Executable import-graph task](6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md)
- [Documentation Thread](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- [Planning-data architecture checklist](../../docs/ARCHITECTURE.md#planning-data-change-checklist)

## Out of scope

- A package/module rewrite or new product feature.
- Accepting an ADR without the user's sign-off.
