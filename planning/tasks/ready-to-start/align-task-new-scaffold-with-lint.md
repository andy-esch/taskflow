---
status: ready-to-start
epic: 17-pm-go-cli
description: 'A fresh task new scaffold fails tskflwctl lint (tags: missing, exit 11) - create and lint must agree on required fields'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, cli, ux, dogfooding]
created: "2026-06-12"
---
# Align `task new` scaffold with `lint`

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], finding H4). Hand-verified
> live: in a fresh `init` repo, `epic new` + `task new --epic … --description d`
> + `lint` exits 11 with `tags: missing`.

## Objective

The tool's own create verb produces a file its own linter rejects: `task new`
(`internal/cli/task.go:34-68`) neither requires nor defaults `--tags`, while
`lint` demands non-empty tags on active tasks. The documented
create → lint-clean workflow is broken out of the box.

Decide the contract, then make `new` and `lint` agree:
- **Option A:** require `--tags` at creation (clear, but adds ceremony).
- **Option B:** default a placeholder tag / derive from the epic (magic).
- **Option C:** relax lint — tags become a warning, not a failure, on
  freshly created tasks.

Whichever way, the same audit should sweep the other `new`-vs-`lint` field
pairings so no second mismatch is lurking.

## Acceptance criteria

- [ ] Fresh `init` → `epic new` → `task new` → `lint` exits 0.
- [ ] An integration-style test pins the create→lint-clean invariant.
- [ ] Help text documents whatever `--tags` contract is chosen.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/cli/task.go`, `internal/core/service.go` (lint rules),
  docs/README examples.