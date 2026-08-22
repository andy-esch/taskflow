---
schema: 1
status: retired
description: 'Shipped multi-space planning: a home registry and CLI plus a navigable TUI atlas, with one logical space per durable planning identity.'
priority: low
tags: [config, cli, tui, multi-repo]
created: "2026-08-15"
updated_at: "2026-08-22"
---

# Multi-space planning: a home registry and the atlas

**Goal.** Multi-space planning: a home registry of planning repos, a `--space`
handle for any command, and a cross-space TUI board (the atlas).

> **Vocabulary locked 2026-08-18: `space` (one durable planning identity) ·
> `atlas` (the whole set).** See
> [decide-the-multi-space-vocabulary-blocks-slice-2](../tasks/6g1erb0p5893-decide-the-multi-space-vocabulary-blocks-slice-2.md)
> for the rejected alternatives, so it is not relitigated.
>
> **Direction settled and shipped 2026-08-22:** registration, the CLI slice, and the
> atlas are implemented. `status --all` is useful, but it cannot provide the
> desired in-process navigation from the cross-space overview into a space. The sketch is
> [multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md);
> its remaining forks are enumerated at the end. An ADR gets written if and when
> the whole shape settles, not before.

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

## Shape

- A home config under `$XDG_CONFIG_HOME/tskflwctl/`, in **two files**: a
  hand-edited `config.toml` the tool only ever reads (which gives
  `[theme]`/`[pager]` the user-level tier their own doc comments describe —
  **shipped**, slice 1), plus a tool-owned `spaces.toml` holding the `[[space]]`
  registry. Registry mutations are atomic and surgical so comments, unknown keys, and
  untouched entries survive.
- The registry stays **advisory**: nothing in it may change what `Discover`
  resolves from a cwd. No home config ⇒ today's behavior exactly.
- Registration is explicit for existing checkouts through `space add`; fresh scaffold and
  pointer `init` runs auto-register best-effort (the `LinkBack` pattern), with
  `--no-register` / `TSKFLW_NO_REGISTER=1` as the opt-out. An
  unregistered current repo on a future board remains a design possibility.
- Broken entries are diagnosed and never auto-removed, via a `SpaceProblem`
  shared by `doctor` and the TUI — the `LinkProblem` pattern one scope up.
- A registry row is an **entry point**, not necessarily a separate space. Rows with
  the same resolved durable planning id — a direct planning checkout and its
  implementation pointers — form one logical space. The grouping is derived from repo
  config, never persisted as parent/child registry state.
- The atlas is a dashboard-like screen (not an `entityTab`), cards rather than a
  list, one card/summary per planning identity, and registered entry points selectable
  within the card. It lands only when >1 logical space is registered.

## Sequencing

The cheap independently-useful parts first, the expensive commitment last:

1. **Home config** — location, load/save, env override, theme/pager precedence.
   Useful on its own, commits to nothing.
2. **Decide `space.id` identity** — durable minted id vs path-keyed. Pulled out as
   its own step because it gates four consumers and carries a safety property: the
   `--space` write guard rests on "naming a tree cannot resolve to the wrong one",
   which a path-keyed id cannot promise. See
   [decide-space.id-identity](../tasks/6g1m8mc8p46h-decide-space.id-identity-durable-minted-id-vs-path-keyed.md).
3. **Registry + CLI** — `space list|add|forget`, `--space`, best-effort fresh-`init`
   auto-registration, `doctor` section, and `status --all` are implemented. No TUI risk.
   Registry catalog, grouping, selection, mutation, completion, diagnosis, and the narrow
   already-initialized registration use case share `core.SpaceRegistryService`;
   planning-tree opening remains separate. `status --all` is the first live test of
   whether the atlas adds enough value.
4. **Re-decide — settled yes on 2026-08-22.** `--space` and `status --all` answer
   cross-space CLI questions, but they do not provide the atlas's defining interaction:
   navigate from the whole collection into one space without leaving the TUI.
5. **Atlas foundation and UI — shipped 2026-08-22.** The narrow reusable
   workspace-opening boundary in
   [establish-reusable-workspace-opening-for-atlas-navigation](../tasks/6g2jhr31f14p-establish-reusable-workspace-opening-for-atlas-navigation.md)
   is followed by the `spaceSession` refactor, cards, explicit ordering and entry paths,
   entry-point selection, and reversible navigation in
   [build-the-tui-atlas-and-navigate-into-spaces](../tasks/6g2jhr3g20ss-build-the-tui-atlas-and-navigate-into-spaces.md).

## Out of scope

- An ADR merely to restate the feature choice. Write one only if implementation settles
  a durable architectural rule that the existing architecture guide cannot carry.
- `space scan` (walking the filesystem for planning repos).
- Persisted per-space stats. Skeleton cards that fill in are honest; cached
  numbers lie.
- Non-local spaces (a served/remote planning repo) — that is ADR-0004's `serve`
  territory. Defining a space by planning identity rather than one local path leaves
  room for that later transport without deciding it now.
- Per-space theming. `accent` would be a card identity cue, not a theme override.
