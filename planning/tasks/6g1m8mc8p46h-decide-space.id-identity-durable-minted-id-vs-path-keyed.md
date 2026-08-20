---
schema: 1
id: 6g1m8mc8p46h
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Settle what identifies a planning repo before the registry schema is written. Gates --space's write guard, linkback verification, the workspace receipt, and moved detection.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [design, config, multi-repo]
created: "2026-08-19"
updated_at: "2026-08-19"
completed_at: "2026-08-19"
---
# Decide `space.id` identity: durable minted id vs path-keyed

## Objective

Settle what identifies a planning repo, **before** any registry schema is written. This
started as one line of the registry task and has since acquired four consumers and a
safety property, which is why it is now its own decision.

## Why this is not a detail

| consumer | what it needs identity for | what breaks if identity is a path |
| --- | --- | --- |
| `--space <id>` selection | the wrong-repo write guard that replaced `--expect-root` | the guarantee "naming a tree cannot resolve to the wrong one" stops holding when a path drifts |
| linkback verification (epic 23) | proving a `planning_repo` pointer reached the intended repo | a committed relative path can **silently** resolve to an unrelated planning repo — demonstrated |
| `wire.WorkspaceJSON` on receipts | proving which tree a mutation changed | the receipt names a path that may since have moved |
| `moved` detection (fork 6) | recognising a repo that changed location | reduces to a filename heuristic |

Evidence for row 2:
[worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
— a worktree outside the sibling layout resolved to a *different* planning repo with no
warning and no error.

## The options

- **Durable minted id.** A stable id written into the planning repo's own **committed**
  `.tskflwctl.toml` (`id = "6g1…"`, minted by `init`). Registry entries and pointers
  reference it; path is a hint, id is the assertion. Follows
  [ADR-0003](../adrs/0003-stable-key-id-addressed-storage.md)'s philosophy of addressing
  by a durable key rather than a location. Costs a backfill for existing repos.
- **Path-keyed label.** `space.id` is a local nickname for a path; nothing is written to
  the target repo. Zero migration, no coordination — but every row above stays fragile,
  and the wrong-tree hazard is permanent.

## What to settle

- Durable or path-keyed? If path-keyed, **record explicitly that the wrong-tree hazard
  is accepted**, so it reads as a decision rather than an omission.
- If durable: where is it minted (`init` only, or `lint --fix`/`doctor` backfill for the
  two existing repos)? Is a repo without one an error, a warning, or silently tolerated?
- Does `wire.WorkspaceJSON` carry it, so the receipt and the registry agree? (Adding a
  field there is cheap now, awkward once consumers pin the shape.)
- Is the id per **planning repo** only, or does an impl repo get one too? Only the
  planning repo needs identity for the four consumers above — resist widening.

## Decision (2026-08-19)

**Two identities, not one.** Chosen by andy-esch. The original framing — durable *vs*
path-keyed — was a false dichotomy: a committed id is shared by every worktree of a repo,
so it identifies the **repo**, never the **checkout**. They do different jobs and both are
adopted.

| | what it is | what it answers |
| --- | --- | --- |
| **path** | the registry entry's location | *which working tree* — unique per checkout |
| **durable id** | minted into the planning repo's **committed** `.tskflwctl.toml` | *which repo* — survives moves, shared by worktrees |

**Path is the address. Id is the assertion.**

### 1. `--space <X>` takes the LOCAL LABEL

```toml
[[space]]
id        = "taskflow"                  # local label — the address
path      = "~/git/andy-esch/taskflow"
verify_id = "6g1durable0aa"             # the assertion, checked after resolving
```

The label addresses a **checkout**, which the durable id cannot: every worktree of a repo
carries the same committed id, so `--space <durable-id>` would be ambiguous the moment a
worktree exists. Labels are also typable; ids are not.

### 2. Verification is OPT-IN per pointer, and strict once opted in

```
pointer has id + target matches   → proceed
pointer has id + target differs   → exit 14 (ErrConflict)
pointer has id + target has NONE  → exit 14
pointer has NO id (legacy)        → today's behavior, unchanged
```

**"Missing also fails" is the load-bearing case, not an edge case.** The decoy scenario
that motivated this — an unrelated planning repo sitting where a committed relative path
lands — is a repo with *no id at all*. Tolerating a missing target id would leave the
original hazard wide open while appearing to close it.

Legacy pointers resolving unchanged is what keeps this non-breaking: nothing fails until a
pointer opts in by recording an id.

### 3. Settled without needing a call

- **Planning repos only.** All four consumers (linkback verification, `--space`, the
  `workspace` receipt, `moved` detection) concern the planning tree. An impl repo would
  carry a committed id with no consumer — resist widening.
- **`init` mints; `doctor` / `lint --fix` backfills.** The established pattern here, and
  only two repos need it.
- **`wire.WorkspaceJSON` carries the durable id**, so the receipt and the registry agree.
  Additive, and `workspace` is marked EXPERIMENTAL precisely so its shape can still move —
  cheap now, awkward once consumers pin it.

### What this unblocks and dissolves

- The registry schema can now be written ([space-registry-model](6g0fzhbtk0gy-space-registry-model-and-the-space-verbs-list-add-forget.md)).
- `--space` becomes a real guard rather than a convention
  ([global-space-flag](6g0fzk8mazrc-global-space-flag-run-any-command-against-a-registered-space.md)).
- Audit 2026-08-19 **M2 largely dissolves**: whichever base a relative `planning_repo`
  resolves against, a wrong target now fails loudly instead of binding silently
  ([decide-how-a-relative-planning-repo-resolves](6g1sb9keb5c9-decide-how-a-relative-planning-repo-resolves-from-a-worktree-branch.md)).
- `moved` detection becomes honest rather than a filename heuristic.

### Accepted cost

`init` writes an id into a **committed** file, so it appears in diffs for everyone working
in that repo, and the two existing planning repos need a backfill. Accepted knowingly: it
is the only mechanism that closes the silent wrong-tree bind in the general case.

## Acceptance criteria

- [x] The choice is recorded here with its rejected alternative and the accepted cost
- [x] If durable: minting, backfill, and missing-id behavior are specified
- [x] The registry task's schema section is updated to match
- [x] Whether `WorkspaceJSON` carries it is decided either way

## Out of scope

- Implementing the registry — this is the decision that unblocks it.
- Layers A/B of the worktree fix, which are independent of this
  ([make-cross-repo-resolution-worktree-aware](6g1kxj5br3z1-make-cross-repo-resolution-worktree-aware.md)).

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- The evidence: [worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
- Consumer 1: [global-space-flag-run-any-command-against-a-registered-space](6g0fzk8mazrc-global-space-flag-run-any-command-against-a-registered-space.md)
- The schema this gates: [space-registry-model-and-the-space-verbs-list-add-forget](6g0fzhbtk0gy-space-registry-model-and-the-space-verbs-list-add-forget.md)

### 2026-08-19 — "durable vs path-keyed" is a false dichotomy

Prompted by asking whether a user should be able to select *which worktree* of a planning
repo to view. Tested: three worktrees of one planning repo, three branches, and the
**same committed id** in all three.

```
CHECKOUT         COMMITTED id     BRANCH
planning         6g1durable0aa    main
planning-wt-a    6g1durable0aa    branch-a
planning-wt-b    6g1durable0aa    branch-b
```

A committed id necessarily identifies the **repo**, not the **checkout**. So a single
durable id cannot serve as the registry key the moment worktrees exist — `--space <id>`
would be ambiguous across them.

**The two identities answer different questions, and this task should decide both:**

| identity | answers | properties | consumers |
| --- | --- | --- | --- |
| **repo identity** (durable, committed) | "is this the planning repo I meant?" | survives moves; **not** unique per checkout | linkback verification, the wrong-tree guard, `moved` detection |
| **checkout identity** (path, local label) | "which working tree am I looking at?" | unique per checkout; **breaks when the path moves** | the registry key, `--space` selection |

Framed as either/or, each option loses something real. Held as two fields, they compose:
the registry keys on a local label + path, and *verifies* against the committed repo id
before writing. Path is the address, id is the assertion — the same split
[worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
layer C proposes for pointers.

**On registering worktrees as spaces — leaning against.** Worktrees are ephemeral:
`desirelines-planning` already carries a **prunable** worktree registration
(`/workspace/dl-sweepport`, a directory that no longer exists). A registry that admits
them accumulates corpses and amplifies the `missing`/`moved` diagnosis problem. Selecting
a checkout stays `-C <path>`, which already works — a planning-repo worktree correctly
resolves to itself.

**But note taskflow cannot follow ADR-0004's trunk-only guidance.** It self-hosts
(`taskflow_root = "./planning"`), so *every* branch and worktree of this repo carries its
own planning tree — the divergence ADR-0004 set out to escape is this repo's default
state, not an exception. Whatever is decided here should be honest about that rather than
assuming one checkout per planning repo.
