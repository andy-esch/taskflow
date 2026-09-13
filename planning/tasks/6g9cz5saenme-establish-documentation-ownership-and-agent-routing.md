---
schema: 1
id: 6g9cz5saenme
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Define canonical owners, precedence, and task-based navigation so agents load the right project guidance without duplicating mutable facts.
effort: 4-6 hours
tier: 3
priority: medium
autonomy_level: 4
tags: [documentation, agents, architecture, dx]
created: "2026-09-12"
updated_at: "2026-09-13"
---
# Establish documentation ownership and agent routing

## Objective

Give contributors and coding agents one short, tool-neutral entry point that explains which source owns each kind of project truth, how conflicts are handled, and which deeper document to read for a given change.

The current root has a substantial `CLAUDE.md` but no `AGENTS.md`. It duplicates architectural and planning behavior that is also present in the README, `docs/ARCHITECTURE.md`, generated CLI reference, schema output, and code. One copy already says Thread membership and lifecycle mutations have not landed, while the current command surface says they have.

## Acceptance criteria

- [ ] A documented source-of-truth matrix assigns canonical ownership for build commands, architectural policy, package contracts, CLI behavior, entity authoring conventions, compatibility promises, decisions, and current planning status.
- [ ] A concise root `AGENTS.md` provides tool-neutral non-negotiables, source precedence, validation commands, and a task-to-document routing table without copying large mutable inventories.
- [ ] `CLAUDE.md` becomes a thin agent-specific adapter or a checked/generated projection of the shared guidance; stale Thread command guidance is removed.
- [ ] Conflicts are handled explicitly: current code, executable checks, and generated schemas describe implemented behavior; ADRs describe accepted decisions; planning documents describe status and history.
- [ ] Any nested agent guidance is limited to genuinely local constraints and links to package or subsystem documentation for explanation.
- [ ] README and contributor entry points route readers to the new guide without presenting another competing architecture summary.

## Out of scope

- Encoding every product behavior in agent instructions.
- Assuming that every coding agent supports the same tool-specific instruction filename or import syntax.
- Rewriting architecture guides, package documentation, or CLI reference content owned by sibling tasks.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
