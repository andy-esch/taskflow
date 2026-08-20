---
schema: 1
id: 6g1sb9keb5c9
status: next-up
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: A worktree branch that edits planning_repo takes its value from the worktree but its base from the canonical checkout. Decide, or let layer C identity verification dissolve it.
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [config, multi-repo, design]
created: "2026-08-19"
---
# Decide how a relative `planning_repo` resolves from a worktree branch

## Objective

Settle a semantic the worktree work deliberately left open
(audit 2026-08-19-worktree-aware-resolution, M2).

`resolvePlanningRepo` takes the **value** from the worktree's own `.tskflwctl.toml` — which
may differ by branch — but joins it against the **canonical checkout's** directory. If a
branch changes the pointer to something created relative to the *worktree*, resolution
either errors confusingly (naming the main checkout's parent) or, worse, binds silently to
an unrelated directory of the same name sitting next to the main checkout.

## Why it was left open

The mix is right for the common case: a committed `../planning` means "sibling of the
repo", and the repo's canonical location is the main checkout — which is precisely what
makes the value mean the same thing from every worktree. It is only wrong when a branch
*changes* the pointer to something worktree-relative, which requires a deliberate and
unusual edit.

## Options

- **Document it and move on.** The behavior is defensible; the failure needs an unusual
  edit. Cheapest, and honest if written down.
- **Split on the prefix** (the reviewer's suggestion): `./`-relative resolves against the
  actual dir, `../` escapes anchor to the canonical checkout. Expressive, but adds a
  subtle rule to a file most users never read, and `./x` vs `x` would behave differently.
- **Wait for identity verification.** Layer C
  ([worktrees-and-relative-paths](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md))
  has pointers record the target's durable id and verify after resolving. The silent-wrong-
  bind — the only *dangerous* outcome here — becomes a loud error regardless of which base
  was used, which dissolves most of this finding without a new rule.

## Urgency

**Low, and coupled.** Requires a deliberate per-branch pointer edit, so it is rare; but the
bad outcome is a silent wrong-tree bind, which is the same class the branch just closed.
The efficient moment is **alongside the `space.id` identity decision**
([decide-space.id-identity](6g1m8mc8p46h-decide-space.id-identity-durable-minted-id-vs-path-keyed.md)),
because adopting a durable id likely removes the danger without needing the prefix rule.
Deciding it before that would risk shipping a rule that identity verification then makes
redundant.

## Acceptance criteria

- [ ] The behavior is decided and recorded with its rejected alternatives
- [ ] If unchanged: documented where a reader of `.tskflwctl.toml` will find it
- [ ] The silent-wrong-bind path is either impossible or produces a loud error

## Related

- Audit [2026-08-19-worktree-aware-resolution](../audits/6g1s9jpzr1yk-2026-08-19-worktree-aware-resolution.md) M2
- [decide-space.id-identity](6g1m8mc8p46h-decide-space.id-identity-durable-minted-id-vs-path-keyed.md)
