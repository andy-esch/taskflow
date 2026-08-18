---
schema: 1
status: active
description: 'Explore multi-space planning: a home registry of planning repos, a --space handle for any command, and a cross-space TUI board (the atlas). Design-first — naming, schema, and scope are all open.'
priority: low
tags: [config, cli, tui, multi-repo]
created: "2026-08-15"
---

# Multi-space planning: a home registry and the atlas

**Goal.** Explore multi-space planning: a home registry of planning repos, a
`--space` handle for any command, and a cross-space TUI board (the atlas).
Design-first — naming, schema, and scope are all open.

> **Nothing is decided.** The sketch is
> [2026-08-15-multi-space-home-registry-and-the-atlas](../research/2026-08-15-multi-space-home-registry-and-the-atlas.md)
> — a directionally-correct direction, not a design. Names (`space`, `atlas`),
> the config location and schema, the registration policy, and whether the TUI
> board is built at all are open questions, enumerated at the end of that doc. An
> ADR gets written **if and when** the shape settles, not before.

## The itch

Discovery is cwd-anchored, so `internal/config/config.go` can say in its own doc
comment: *"One planning repo per product; no cross-product registry."* Two
questions therefore have no answer: **"what am I in the middle of, anywhere?"**
(N terminals or N `cd`s) and **"show me that other repo's planning, from here"**
(`-C` needs a path you must already know). k9s's context list is the reference
point, and the counter-example — the bar is to carry the TUI's visual language,
not print a table of paths.

## Why this is its own epic

It spans three areas that no sibling owns together: a **new config scope**
(`internal/config`, above the repo), a **CLI surface** (`space` verbs, a global
`--space`, a `doctor` section), and a **TUI screen** plus the `model.go` split it
would force. Epic 23 is the close sibling but is deliberately *product*-scoped —
it links the repos of one product, which is a different axis from "one user's
several products" — and stretching its keys to cover this would put an impl repo
in a peer product's config.

## Shape (all provisional)

- A home config at `$XDG_CONFIG_HOME/tskflwctl/config.toml` whose first citizen
  is a `[[space]]` registry — and which also gives `[theme]`/`[pager]` the
  user-level tier their own doc comments describe.
- The registry stays **advisory**: nothing in it may change what `Discover`
  resolves from a cwd. No home config ⇒ today's behavior exactly.
- `init` auto-registers best-effort (warns, never fails — the `LinkBack`
  pattern); `space add` for existing repos; an unregistered current repo shows in
  the board with a keypress to keep it.
- Broken entries are diagnosed and never auto-removed, via a `SpaceProblem`
  shared by `doctor` and the TUI — the `LinkProblem` pattern one scope up.
- The atlas is a dashboard-like screen (not an `entityTab`), cards rather than a
  list, per-card async summaries, landing only when >1 space is registered.

## Sequencing

The cheap independently-useful parts first, the expensive commitment last:

1. **Home config** — location, load/save, env override, theme/pager precedence.
   Useful on its own, commits to nothing.
2. **Registry + CLI** — `space list|add|forget`, `--space`, `init`
   auto-register, `doctor` section, `status --all`. No TUI risk.
3. **Re-decide.** Having lived on the CLI half, is the board still wanted? It is
   plausible that `--space` + `status --all` is most of the value.
4. Only if yes: the `Resolve() → Workspace` seam
   ([6fgcr2403sjn](../tasks/6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md),
   currently deferred), the `spaceSession` refactor, the atlas, the switcher.

## Out of scope

- **Any ADR or irreversible naming commitment** until step 3.
- The atlas's visual design — card layout, accent derivation, the cross-space
  rail. Sketched, not specified.
- `space scan` (walking the filesystem for planning repos).
- Persisted per-space stats. Skeleton cards that fill in are honest; cached
  numbers lie.
- Non-local spaces (a served/remote planning repo) — that is ADR-0004's `serve`
  territory, and it may eventually invalidate "a space is a local path."
- Per-space theming. `accent` would be a card identity cue, not a theme override.
