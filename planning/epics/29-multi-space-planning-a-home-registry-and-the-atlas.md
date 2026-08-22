---
schema: 1
status: active
description: 'Multi-space planning: a home registry of planning repos, a --space handle for any command, and a cross-space TUI board (the atlas). Vocabulary locked; whether the board ships is decided at step 3.'
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
> **Still open:** the registration policy details, and — deliberately — **whether
> the TUI board is built at all** (step 3 below). The sketch is
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
- Registration is explicit through shipped `space add`; best-effort `init`
  auto-registration remains a planned follow-up (the `LinkBack` pattern). An
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
3. **Registry + CLI** — `space list|add|forget`, `--space`, `init`
   auto-register, `doctor` section, `status --all`. No TUI risk. Everything except
   init auto-registration is now implemented. Registry catalog, grouping, selection,
   mutation, completion, and diagnosis share `core.SpaceRegistryService`; planning-tree
   opening remains separate. `status --all` is the first live test of whether the atlas
   adds enough value.
4. **Re-decide.** Having lived on the CLI half, is the board still wanted? It is
   plausible that `--space` + `status --all` is most of the value.
5. Only if yes: the `Resolve() → Workspace` seam
   ([6fgcr2403sjn](../tasks/6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md),
   currently deferred), the `spaceSession` refactor, the atlas, the switcher.

## Out of scope

- **Any ADR** until step 4 — the vocabulary is settled, the overall model is not.
- The atlas's visual design — card layout, accent derivation, the cross-space
  rail. Sketched, not specified.
- `space scan` (walking the filesystem for planning repos).
- Persisted per-space stats. Skeleton cards that fill in are honest; cached
  numbers lie.
- Non-local spaces (a served/remote planning repo) — that is ADR-0004's `serve`
  territory. Defining a space by planning identity rather than one local path leaves
  room for that later transport without deciding it now.
- Per-space theming. `accent` would be a card identity cue, not a theme override.
