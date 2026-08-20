---
schema: 1
id: 6g1sb9keb5c9
status: deprecated
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: A worktree branch that edits planning_repo takes its value from the worktree but its base from the canonical checkout. Decide, or let layer C identity verification dissolve it.
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [config, multi-repo, design]
created: "2026-08-19"
updated_at: "2026-08-20"
deprecated_at: "2026-08-20"
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

### 2026-08-19 — largely dissolved by the identity decision

The durable-id decision landed in the "adopt" direction
([decide-space.id-identity](6g1m8mc8p46h-decide-space.id-identity-durable-minted-id-vs-path-keyed.md)),
which was the coupling this task was waiting on.

With a pointer recording `planning_repo_id` and verification failing on **mismatch or a
missing target id**, the *dangerous* outcome here — a relative path silently binding to an
unrelated planning repo next to the main checkout — becomes a loud `exit 14` regardless of
which base the path resolved against.

What remains is only the **confusing-error** case: a worktree branch whose pointer is
worktree-relative still resolves against the canonical checkout and fails, now with a clear
id-mismatch message rather than a silent bind. That is a documentation question, not a
correctness one, so the `./` vs `../` prefix rule is very likely unnecessary.

**Recommendation: close as wontfix once id verification ships**, documenting the base-anchoring
behavior where a reader of `.tskflwctl.toml` will find it. Left open until then so the
dependency is visible.

### 2026-08-20 — closing: the premise arrived

Identity verification shipped (epic 23, merged as `feat/workspace-ids`), which was the
condition this task was waiting on. Re-checking the finding against the shipped behavior:

- **The dangerous outcome is gone for any opted-in pointer.** A relative `planning_repo`
  that resolves to an unrelated repo now fails with `ErrConflict` (exit 14) — on a
  mismatched id *and* on a target carrying no id, which is the shape a decoy actually has.
  It cannot silently bind, regardless of which base the path resolved against.
- **What remains is a confusing error, not a correctness bug.** A worktree branch whose
  pointer is worktree-relative still resolves against the canonical checkout and fails —
  now with an id-mismatch message that names both ids, rather than quietly using the wrong
  tree.
- **So the `./` vs `../` prefix rule is unnecessary.** It would add a subtle distinction to
  a file most users never open, to disambiguate a case that is now loud anyway. `./x` and
  `x` behaving differently would be its own trap.

**Documented rather than changed:** `pointerConfigTOML` now states the anchoring in the
config it writes — *"In a git worktree a RELATIVE value resolves against the canonical
checkout, not the worktree, so it means the same thing from every checkout of this repo."*
That is where a reader of `.tskflwctl.toml` will actually look. The rule is also in
`docs/ARCHITECTURE.md` and in `internal/config/worktree.go`'s package comment.

**One new behavior worth knowing**, surfaced by the fix rather than by this finding:
changing `planning_repo` on a branch now also requires updating `planning_repo_id`, or
resolution fails. That is arguably correct — it forces the change to be intentional — but
nobody has hit it yet. If it proves annoying, the fix is a better error message, not
weaker verification.
