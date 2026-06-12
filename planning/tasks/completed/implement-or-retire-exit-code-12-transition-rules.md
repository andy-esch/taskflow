---
status: completed
epic: 17-pm-go-cli
description: ErrInvalidTransition and documented exit code 12 can never fire - implement a domain transition matrix or remove the dead contract
effort: Unknown
tier: 2
priority: medium
autonomy_level: 3
tags: [go, domain, contract]
created: "2026-06-12"
updated_at: "2026-06-12"
started_at: "2026-06-12"
completed_at: "2026-06-12"
---
# Implement or retire exit code 12 (transition rules)

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], finding M16). This is a
> *decision* task first: a dead documented contract is worse than no
> contract for the scripting agents it targets.

## Objective

`domain.ErrInvalidTransition` is defined (`internal/domain/errors.go:11`)
and mapped to exit 12 (`internal/cli/exit.go:19`), README and
ARCHITECTURE.md advertise it — but nothing ever returns it. `Service.Move`
(`core/service.go:65-67`) and `FS.Move` (`fsstore.go:97-99`) accept any
status→status.

Decide, then do one of:
- **Implement:** a transition matrix in `domain` (the natural home —
  `Status` already owns `IsActive`). Define which moves are illegal (e.g.
  `deprecated → in-progress`?) and whether a `--force` escape hatch exists.
  Beware breaking the dogfooded workflow — current usage moves tasks freely.
- **Retire:** delete the sentinel, the exit-code mapping, and the README/
  ARCHITECTURE rows; renumber nothing (keep 13/14 stable).

## Acceptance criteria

- [x] Either exit 12 is reachable and tested, or it no longer appears in
      code or docs.
- [x] Decision recorded in the task body (or a planning note) with the
      reasoning.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/domain/`, `internal/core/service.go`,
  `internal/cli/exit.go`, `README.md`, `docs/ARCHITECTURE.md`.
## Decision + closure (2026-06-12)

**Retired** (decision D3 in [[2026-06-12-pending-decisions]]): no transition
rules exist and the dogfooded workflow moves tasks freely, so the sentinel was
a dead documented contract. `ErrInvalidTransition`, the exit-code mapping, and
the README/ARCHITECTURE rows are gone; 13/14 keep their numbers and 12 stays
reserved (comments in domain/errors.go + cli/exit.go). Reinstate via a domain
transition matrix if a real rule ever emerges.
