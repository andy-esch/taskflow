---
schema: 1
id: 6g28rv8j9d9p
status: completed
epic: 21-code-quality-architecture-hardening
description: Architecture direction is documented but mostly unenforced; add focused dependency guardrails and give epic 21 a current scope, rationale, and trigger map.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, lint, documentation]
created: "2026-08-21"
started_at: "2026-08-21"
updated_at: "2026-08-21"
completed_at: "2026-08-21"
---

# Codify package dependency fitness rules and refresh the architecture baseline

## Objective

Turn the project’s current architectural claims into a small, enforceable baseline. The
dependency graph is healthy today, but most rules in `docs/ARCHITECTURE.md` are prose and
`.golangci.yml` enforces only the repo-config versus home-config boundary. Document the
actual package map, encode the stable inward-dependency rules as focused fitness checks,
and refresh epic 21 so future work has a current rationale and trigger map.

## Acceptance criteria

- [x] Generate and review the current internal import graph; document the intended layers,
  consumer-owned port rule, and explicit composition-root exceptions using actual package
  names rather than an aspirational diagram.
- [x] Add narrowly scoped dependency checks for at least `internal/domain`,
  `internal/core`, `internal/wire`, and the TUI boundary, while retaining the existing
  `internal/config` prohibition on importing home-scoped `userconfig`.
- [x] Keep `internal/cli`’s legitimate composition-root imports explicit; do not forbid the
  root from constructing stores or launching the TUI, and do not make primary adapters
  import one another outside that documented composition path.
- [x] Audit current direct primary-adapter to secondary-adapter imports. Classify each as a
  composition concern, an intentional narrow fs/text port, or tracked debt; link existing
  tasks instead of filing duplicates.
- [x] Replace epic 21’s placeholder “why” and “out of scope” sections with its current
  architectural goal, completed foundation, live work, and the concrete atlas/web triggers
  for deferred workspace/context changes.
- [x] `docs/ARCHITECTURE.md`, lint configuration, and planning references agree; `just
  test`, `just lint`, planning lint, and `git diff --check` pass.

## Out of scope

- Repackaging the repository, renaming layers, introducing a framework, or rewriting
  working services merely to match a textbook diagram.
- Activating web-only context propagation or the reusable workspace resolver before a
  second consumer is committed.
- Combining this audit/guardrail task with the separate space-registry application-boundary
  implementation.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- [2026-06-27 consumer data-flow architecture audit](../audits/6fgcr24022z8-2026-06-27-consumer-data-flow-architecture.md)
- [Reusable workspace discovery seam](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)
- [`context.Context` through core/store ports](6fgcr24023yr-thread-context.context-through-the-core-and-store-ports.md)

## Implementation (2026-08-21)

Generated the production internal import graph with `go list` and recorded the actual
foundation/domain/application/wire/primary/secondary package map in
`docs/ARCHITECTURE.md`. The document now names the consumer-owned port rule, the Cobra
composition-root exception, the focused TUI-with-configui composition, and every current
primary-to-secondary edge with its disposition.

Added four production-only `depguard` rules: domain may reach only `id`; core only
`domain`/`id`; wire only `core`/`domain`; and the TUI/configui/CLI presentation packages
cannot reach persistence, discovery, home-registry, or sibling-primary-adapter packages.
The existing repo-config versus home-config rule remains. Temporary negative probes
confirmed all four new boundaries fail lint with the intended named diagnostic; the probes
were removed afterward. `store.FS` now also carries the missing compile-time `core.Linter`
assertion, matching its three narrow fs/text ports.

The direct-import audit found no new task: CLI adapter construction and the narrow
fix/lint/layout/process capabilities are intentional; direct registry orchestration is
already tracked by the reusable space-registry boundary; workspace discovery and request
context remain correctly deferred until atlas/web triggers. Epic 21 now records that
foundation, current work, trigger map, and explicit exclusions instead of placeholders.

Validation: `just test` passed under the race detector; `just lint` reported zero issues;
`golangci-lint config verify`, planning lint, and `git diff --check` passed.
