---
schema: 1
id: 6g6x7e2ef37r
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Revise the oversized architecture document into a smaller navigable structure with clear ownership and durable cross-links.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [docs, architecture, maintainability]
created: "2026-09-04"
depends_on: [6g9cz5saenme]
updated_at: "2026-09-13"
---

# Restructure the architecture documentation into focused guides

## Objective

Make the architecture documentation easier to navigate and maintain by replacing the oversized
single-file structure with a concise entry point and smaller, clearly owned guides organized around
stable architectural concerns. Use the source-of-truth matrix from the documentation-ownership task
to decide what belongs in the index, a focused guide, generated output, package documentation, an
ADR, or planning history.

## Acceptance criteria

- [ ] Define a simple target information architecture before moving content.
- [ ] Break the current architecture document into focused guides with clear scopes and useful
  navigation from a central architecture entry point.
- [ ] Preserve meaningful architectural decisions and context while removing obsolete or duplicate
  prose deliberately rather than accidentally.
- [ ] Repair repository links and contributor guidance that reference the current document layout.
- [ ] Documentation generation/checks, link validation where available, and planning lint pass.

## Out of scope

- Changing product or code architecture merely to match the new document layout.
- Re-litigating accepted ADR decisions unless the restructuring exposes a concrete contradiction
  that should be tracked separately.
- A comprehensive rewrite of unrelated user or CLI documentation.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- Follows [Establish documentation ownership and agent routing](6g9cz5saenme-establish-documentation-ownership-and-agent-routing.md)
