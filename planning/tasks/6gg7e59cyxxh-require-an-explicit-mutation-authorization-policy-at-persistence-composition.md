---
schema: 1
id: 6gg7e59cyxxh
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Prevent forgotten adapter authorization from silently granting persistence access; decide explicit caller modes first.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, ports, safety]
created: "2026-10-03"
updated_at: "2026-10-03"
depends_on: [6gcwcf8rxe72]
---
## Objective

Make omission of a mutation policy explicit at persistence composition rather than silently
granting write access to an accidentally unguarded adapter.

## Evidence

Finding M2 in [the hexagonal-boundaries audit](../audits/6gefs7wbbcyt-2026-09-28-arch-hexagonal-boundaries.md)
shows that `store`, `configstore`, `spacestore`, and `workspacestore` authorize mutations when their
optional callback is absent. The current binary supplies one invocation-scoped authorizer to all
four; composition regressions pin this. The risk is a future caller or constructor forgetting it,
not a newly demonstrated bypass in the shipped CLI.

## Scope and design checkpoint

Inventory persistence constructors and their production/test callers, then choose a small explicit
policy contract: required authorization, read-only construction, and any deliberate unrestricted
fixture/library mode must be distinguishable. Ask the user before choosing the public constructor
compatibility/opt-out behavior. Keep Cobra vocabulary out of secondary adapters and core.

## Acceptance criteria

- [ ] The chosen policy and compatibility treatment are recorded before constructor changes.
- [ ] Missing/typed-nil policy cannot silently authorize a production persistence operation.
- [ ] Legitimate read-only callers and intentionally writable tests have explicit supported modes.
- [ ] The policy reaches stores opened later through workspace/configuration/registry services.
- [ ] Tests cover denied dry-run and committing calls, and current CLI/TUI behavior remains intact.

## Related

- [Composition isolation](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md)
- [Adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)

## Out of scope

- A new global service locator or application-wide permissions framework.
- Replacing repository guards, CAS, or domain eligibility policy with authorization callbacks.
