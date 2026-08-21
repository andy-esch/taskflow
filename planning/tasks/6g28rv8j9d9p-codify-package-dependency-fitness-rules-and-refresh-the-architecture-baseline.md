---
schema: 1
id: 6g28rv8j9d9p
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Architecture direction is documented but mostly unenforced; add focused dependency guardrails and give epic 21 a current scope, rationale, and trigger map.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, lint, documentation]
created: "2026-08-21"
---

# Codify package dependency fitness rules and refresh the architecture baseline

## Objective

Turn the project’s current architectural claims into a small, enforceable baseline. The
dependency graph is healthy today, but most rules in `docs/ARCHITECTURE.md` are prose and
`.golangci.yml` enforces only the repo-config versus home-config boundary. Document the
actual package map, encode the stable inward-dependency rules as focused fitness checks,
and refresh epic 21 so future work has a current rationale and trigger map.

## Acceptance criteria

- [ ] Generate and review the current internal import graph; document the intended layers,
  consumer-owned port rule, and explicit composition-root exceptions using actual package
  names rather than an aspirational diagram.
- [ ] Add narrowly scoped dependency checks for at least `internal/domain`,
  `internal/core`, `internal/wire`, and the TUI boundary, while retaining the existing
  `internal/config` prohibition on importing home-scoped `userconfig`.
- [ ] Keep `internal/cli`’s legitimate composition-root imports explicit; do not forbid the
  root from constructing stores or launching the TUI, and do not make primary adapters
  import one another outside that documented composition path.
- [ ] Audit current direct primary-adapter to secondary-adapter imports. Classify each as a
  composition concern, an intentional narrow fs/text port, or tracked debt; link existing
  tasks instead of filing duplicates.
- [ ] Replace epic 21’s placeholder “why” and “out of scope” sections with its current
  architectural goal, completed foundation, live work, and the concrete atlas/web triggers
  for deferred workspace/context changes.
- [ ] `docs/ARCHITECTURE.md`, lint configuration, and planning references agree; `just
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
