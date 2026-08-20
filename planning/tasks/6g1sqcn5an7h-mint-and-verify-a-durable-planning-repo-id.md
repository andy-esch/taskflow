---
schema: 1
id: 6g1sqcn5an7h
status: ready-to-start
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: 'Implement the identity decision: init mints a durable id into a planning repo''s committed config; a pointer that records it verifies after resolving, failing loudly on mismatch or missing.'
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [config, multi-repo, correctness]
created: "2026-08-19"
---
# Mint and verify a durable planning-repo id

## Objective

Implement the identity decision
([decide-space.id-identity](6g1m8mc8p46h-decide-space.id-identity-durable-minted-id-vs-path-keyed.md)):
a planning repo carries a durable id in its **committed** `.tskflwctl.toml`, and a pointer
that records it verifies after resolving.

**This is independently valuable and does not wait for the registry.** It closes the silent
wrong-tree bind for `planning_repo` pointers — demonstrated in
[worktrees-and-relative-paths](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
— before any space registry exists, and it is the primitive epic 29 then reuses.

## Scope

- **Mint.** `init` writes `id = "<minted>"` into a scaffolded planning repo's config.
  Use `internal/id` (the same generator tasks/audits/research use); no new id machinery.
- **Backfill.** `doctor` or `lint --fix` adds one to a planning repo that lacks it. Only two
  repos need it today, but the path must exist for anyone else.
- **Record.** `init --planning-repo` writes `planning_repo_id` alongside `planning_repo`,
  read from the target at init time.
- **Verify.** `resolvePlanningRepo` compares after resolving:

```
pointer has id + target matches   → proceed
pointer has id + target differs   → exit 14 (ErrConflict)
pointer has id + target has NONE  → exit 14
pointer has NO id (legacy)        → today's behavior, unchanged
```

- **Expose.** `wire.WorkspaceJSON` carries the durable id, so the receipt and (later) the
  registry agree. Schema bump; `workspace` is EXPERIMENTAL so its shape may still move.

## Notes

- **"Missing also fails" is the whole point**, not an edge case. The decoy that motivated
  this is a planning repo with *no id*; tolerating a missing target id would leave the hazard
  open while appearing to close it. Do not soften this to a warning.
- **Legacy stays silent.** A pointer with no `planning_repo_id` must behave exactly as today
  — that is what makes this non-breaking. Do not warn about its absence; that would recreate
  the always-firing noise this epic just removed.
- **Every worktree shares the id** (it is in a committed file). That is correct: the id
  identifies the repo, and worktrees are addressed by path.
- Surgical TOML writes, as ever — preserve comments, key order, unknown keys.
- Exit 14 (`ErrConflict`) is the right code: the world is not what the caller assumed, the
  same sense the CAS path uses.

## Acceptance criteria

- [ ] `init` mints an id into a scaffolded planning repo's config
- [ ] A planning repo without one can be backfilled, and the backfill is idempotent
- [ ] `init --planning-repo` records `planning_repo_id` from the target
- [ ] All four verification rows above are covered by tests, including **missing → exit 14**
- [ ] A legacy pointer (no id) resolves exactly as today, silently — regression test
- [ ] The decoy scenario from the research doc now fails loudly instead of binding
- [ ] `WorkspaceJSON` carries the id; schema bumped, goldens regenerated
- [ ] `just test` + `just lint` green

## Out of scope

- The space registry and `--space` (epic 29) — this is the primitive they consume.
- Any id for impl repos: planning repos only.

## Related

- The decision: [decide-space.id-identity](6g1m8mc8p46h-decide-space.id-identity-durable-minted-id-vs-path-keyed.md)
- The hazard: [worktrees-and-relative-paths](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
- Dissolves: [decide-how-a-relative-planning-repo-resolves](6g1sb9keb5c9-decide-how-a-relative-planning-repo-resolves-from-a-worktree-branch.md)
- Consumer: [global-space-flag](6g0fzk8mazrc-global-space-flag-run-any-command-against-a-registered-space.md)
