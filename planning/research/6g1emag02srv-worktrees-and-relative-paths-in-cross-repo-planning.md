---
schema: 1
id: 6g1emag02srv
created: "2026-08-19"
description: 'Two ways cross-repo resolution goes wrong: worktrees look like untracked repos (constant noise), and a committed relative planning_repo can silently resolve to an unrelated planning tree.'
tags: [config, multi-repo, git, worktree, correctness]
updated_at: "2026-08-20"
---
# Worktrees and relative paths in cross-repo planning

## Question

Running the newly-shipped `tskflwctl workspace` across a real multi-repo setup
(`desirelines` + `desirelines-planning`, plus four git worktrees) surfaced two ways
cross-repo resolution goes wrong. One is constant noise; the other silently writes to
the wrong planning tree. What causes them, and what would make resolution bullet-proof?

Both were found by accident — the observability shipped for audit
2026-07-24 H1 is what made them visible at all, which is itself a data point about
where to invest.

## Findings

### 1. Worktrees are invisible to the linkback model

Four directories in `~/git/andy-esch/` are **git worktrees** of `desirelines`. Each
carries the committed `.tskflwctl.toml` with its `planning_repo` pointer, while the
planning repo's `tracked_repos` lists only `["../desirelines", "../desirelines-deploy"]`.
Every worktree therefore looks like an untracked impl repo and emits
`⚠ one-sided link` **on every command**:

```
⚠ desirelines-docs-link-updates
⚠ desirelines-wt-chore-tidy-up-things
⚠ desirelines-wt-go-mod
⚠ desirelines-wt-tf-cleanup
```

**Root cause:** `CheckLinks` compares *physical directory paths*. A worktree is a
different directory but the **same repo** — the model has no notion of repo identity
distinct from directory identity.

**Tracking each worktree is the wrong fix.** They are created and deleted constantly,
it would make branching mutate the planning repo's config as a side effect, and the
list would never converge.

**The asset:** a worktree's `.git` is a *file* containing
`gitdir: /…/desirelines/.git/worktrees/<name>`. The canonical checkout is recoverable
with a plain file read — no git binary, no subprocess. `config.go` already
special-cases `.git`-as-a-file for the climb boundary; it simply never reads the
contents.

**Why this is worse than cosmetic:** a warning that fires constantly on a *healthy*
setup is one you train yourself to ignore, which costs you the one time it is real.

### 2. A committed relative path is a bet about directory layout

`planning_repo = "../desirelines-planning"` is **committed**, so it travels to every
clone, worktree, and CI checkout — and it resolves against *the config file's
location*, which is precisely the thing that varies.

Tested against a synthetic repo. A worktree placed outside the sibling layout:

```
error: validation failed: planning_repo "../planning-repo" points at
.../elsewhere/planning-repo, which is not a planning repo — run `tskflwctl init` there first
```

Loud and recoverable. **Then an unrelated planning repo was placed at that location:**

```
root:   .../wt/elsewhere/planning-repo      ← the WRONG tree
config: .../wt/elsewhere/impl-wt/.tskflwctl.toml
source: pointer
```

**Silent.** No warning, no error. Every mutation from that worktree lands in a
completely unrelated planning repo.

This is the audit-H1 wrong-repository write hazard arriving through a door H1 never
considered. `workspace` *reports* it correctly, but nothing *prevents* it.

Absolute paths are not the escape: a committed absolute path breaks on every other
machine and in CI.

### 3. The two compound

Worktrees are the main reason a checkout ends up somewhere other than the sibling
layout. So finding 1 is the mechanism that makes finding 2 likely, and the two should
not be reasoned about separately.

## Recommendation (as of 2026-08-19)

Four layers, cheapest first. A and B are independent and small; C is the durable fix
and must be decided *with* `space.id`; D falls out of the registry for free.

### A. Worktree-aware linkback

When `CheckLinks` meets a `.git` *file*, read its `gitdir:`, walk up from
`<main>/.git/worktrees/<name>` to `<main>`, and compare **that** against
`tracked_repos`. `desirelines-wt-go-mod` then resolves to `desirelines`, which is
tracked — the warning disappears *correctly* rather than by suppression.

Small, dependency-free, and it extends a special case the code already has.

### B. Resolve `planning_repo` from the canonical checkout

The same mechanism applied to resolution rather than checking: when a config lives in
a worktree, resolve a **relative** `planning_repo` against the canonical checkout's
directory. `../desirelines-planning` then means the same thing from every worktree —
which is what the author meant when committing it — and worktrees stop being a special
case at all.

### C. Identity, not paths — the real fix

Stop betting on layout. Mint a durable id into the planning repo's own committed
`.tskflwctl.toml`, and have pointers record **both**:

```toml
planning_repo    = "../desirelines-planning"   # a hint about where to look
planning_repo_id = "6g1abc…"                   # the assertion about what to accept
```

Resolution follows the path, then **verifies the id**. A mismatch is a loud error,
never a silent wrong tree. This is the only layer that closes finding 2 in the general
case rather than just for worktrees.

It is the same primitive already flagged for `space.id`, which now has **four**
consumers: linkback verification, `--space` selection, the `workspace` receipt, and
`moved` detection in the registry. Defining it once, deliberately, is the whole point —
inventing it per-feature is how the four drift apart.

The id belongs in the **committed** config, because identity is a property of the repo,
not of the machine. Existing repos need a backfill: `init` mints for new ones,
`lint --fix` or `doctor` backfills the rest.

### D. Recovery via the registry

Once the space registry exists, a failed path lookup can fall back to "find the space
whose id matches." The registry becomes the resolver of last resort — which earns it
keep for something other than the atlas.

### Rejected: suppress the warning for worktrees

The cheap stopgap. It hides finding 2's only visible symptom while leaving the hazard
live. A is barely more work and is actually correct.

### What would change the call

If `space.id` is decided as a *path-keyed* identity rather than a minted durable one,
C becomes unavailable and finding 2 stays open permanently — the registry would inherit
exactly the fragility this documents. That decision and this one are the same decision.

## Progress — shipped 2026-08-20

All four layers landed, in two merges (`feat/worktree-handling-for-spaces`,
`feat/workspace-ids`). Recording it here so the Recommendation above reads as history
rather than a plan.

| layer | status |
| --- | --- |
| **A** worktree-aware linkback | **shipped** — `anchorDir` in `internal/config/worktree.go` |
| **B** resolve out-of-tree paths from the canonical checkout | **shipped** — same helper, wired into `resolvePlanningRepo` and `resolveRepoPath` |
| **C** identity, not paths | **shipped** — `init` mints a durable `id`; a pointer recording `planning_repo_id` verifies after resolving |
| **D** registry recovery | **not built** — waits on the space registry (epic 29) |

**The rule that came out of implementing A/B** turned out to be cleaner than "worktrees are
special", and is worth carrying forward: *which side of the repo boundary a path points at
decides what it resolves against.* `taskflow_root` is in-tree and anchors to the config
file's own dir; `planning_repo` and `tracked_repos` point outward and anchor to the
canonical checkout. Now stated in `docs/ARCHITECTURE.md`.

**What testing changed about the plan:**

- `.git`-as-a-file has **three** causes and only one is a worktree. `--separate-git-dir`
  and submodules must be left alone; the discriminator is a `/worktrees/<name>` segment.
- A worktree of the **planning** repo must NOT be redirected — it edits its own tree. The
  original framing would have written to a different checkout than the user stood in.
- The main working tree of a `--separate-git-dir` repo is **not recorded anywhere**; `git
  worktree list` reports the git dir itself. Degrading is the only honest behavior
  (audit 2026-08-19 H1, wontfix).
- Verification had to fail on a **missing** target id, not only a mismatch: the decoy that
  motivated layer C is a repo with no id at all.

**Still open:** layer D, and whether a pointer changing `planning_repo` on a branch should
also have to update `planning_repo_id` (it must today, which forces intentionality but
nobody has hit it yet).

## Related

- The hazard this extends, and the observability that exposed it: audit
  [2026-07-24-ai-agent-cli-ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md) H1
- The cross-repo model at issue: epic
  [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)
- Where `space.id` gets decided — the same decision as C: epic
  [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- The multi-space sketch this feeds back into:
  [multi-space-home-registry-and-the-atlas](6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
- Stable-key identity, the philosophy C follows: [ADR-0003](../adrs/0003-stable-key-id-addressed-storage.md)
