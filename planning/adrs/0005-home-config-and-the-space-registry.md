---
status: proposed
date: "2026-08-15"
deciders: [andy-esch]
tags: [adr, config, cli, tui, multi-repo]
supersedes: []
superseded_by: null
---

# ADR-0005: A home config and the space registry — many planning repos, one atlas

> **Implementation status (2026-08-20):** home config, the separate `spaces.toml`
> registry, `space list|add|forget`, and shared list/doctor health diagnosis are shipped.
> Global `--space`, `init` auto-registration, `status --all`, and any atlas UI remain
> planned. The ADR is still proposed because the final TUI commitment is deliberately
> deferred until the CLI half has been used.

> Follows the ADR format established in [0001-adopt-adrs](0001-adopt-adrs.md). Extends the
> cross-repo config model of epic
> [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)
> upward by one scope. Deliberately **does not** touch the on-disk entity model
> ([0003-stable-key-id-addressed-storage](0003-stable-key-id-addressed-storage.md)) or the
> single-source-of-truth question ([0004-single-source-of-truth-serve-owns-git](0004-single-source-of-truth-serve-owns-git.md)).

## Context and Problem Statement

`internal/config/config.go` states the current scope in its own doc comment:

> *"One planning repo per product; no cross-product registry."*

That has held because discovery is **cwd-anchored**: `Discover` walks up from where you
stand, and everything below it — the store, the service, both adapters — is built for the
one root it lands on. A machine with several planning repos (`taskflow`,
`desirelines-planning`, dotfiles, …) has no way to answer two questions:

- **"What am I in the middle of, anywhere?"** Today that is N terminals or N `cd`s. The
  in-progress working set is *per repo* precisely because the tool has no notion of a
  second one.
- **"Show me that other repo's planning without leaving this directory."** There is no
  `--somewhere-else` escape; you must `cd` or `-C`, and `-C` needs the path, which means
  you must already know it.

Epic 23 built real cross-repo machinery — `planning_repo` (impl → planning),
`tracked_repos` (planning → impls), `CheckLinks`/`doctor` — but it is deliberately
**product-scoped**: it links the repos of *one* product. It cannot express "these four
unrelated products are all mine," and it should not be stretched to: an impl repo has no
business appearing in a peer product's config.

So the gap is a **third scope**, above the repo: a user-level record of the planning repos
this machine's owner cares about, and a TUI surface over it.

The naming is load-bearing and constrained: **"project" is already taken** — ADR-0002
defines a Project as a cross-cutting initiative *inside* one planning repo. A second,
outer meaning for the same word would be the worst kind of ambiguity.

## Considered Options

- **A — Do nothing; use `-C`, shell aliases, and `tracked_repos`.** Free, and honest about
  the tool's local-first scope. But `tracked_repos` is the *wrong axis* (it links one
  product's repos, not a user's products), and aliases can't produce a cross-repo view —
  which is the actual ask. Rejected as the end state; it stays the fallback for anyone who
  never registers anything.
- **B — A home-level registry + a user config (chosen).** Two user-scoped TOML files: a
  hand-edited config for terminal concerns (`[theme]`, `[pager]`) and a tool-managed
  registry of known planning repos. Additive, inspectable, dotfile-committable, and a
  registry mutation cannot re-encode a person's presentation settings.
- **C — Elect one planning repo as the "hub" and keep the index there.** No new file
  location, no `$HOME` writes. Rejected: there is no natural hub, it makes one peer repo
  arbitrarily special, and cloning that repo onto a second machine would drag a
  machine-specific path list with it.
- **D — No registry; scan the filesystem for `.tskflwctl.toml` on demand.** Zero
  registration ceremony. Rejected as the *primary* mechanism: it is slow, it surfaces every
  stale checkout and worktree, and the result isn't stable enough to address by id. Kept as
  a possible later convenience (`space scan`) that **feeds** the registry rather than
  replacing it.

## Decision

Adopt a **home config** whose first citizen is a **space registry**.

### 1. Vocabulary: space (the one), atlas (the all)

- A **space** is one durable planning identity — the unit summarized once in cross-space
  views. A registered path is an **entry point** into it: either the direct planning
  checkout or a repo whose `planning_repo` points there.
- The **atlas** is the whole set, and the name of the TUI screen over it.
- **Not "project"** — ADR-0002 owns that word for cross-cutting initiatives inside a
  space. The two live at different scopes and must never share a noun.
- The **internal type stays `Workspace`** (`core.Workspace`: the resolved
  config + store + service bundle for one root), matching the vocabulary already used in
  the `Resolve() → Workspace` seam. `space` is the user-facing skin over it.

### 2. Location: XDG, explicitly not `os.UserConfigDir()`

Two files under `$XDG_CONFIG_HOME/tskflwctl/`, falling back to
`~/.config/tskflwctl/`:

- `config.toml` — hand-edited user presentation settings; the tool only reads it.
- `spaces.toml` — the tool-managed, still hand-inspectable space registry.

Go's `os.UserConfigDir()` returns `~/Library/Application Support` on darwin — wrong for a
dotfile-friendly CLI whose users commit their config. The path is **overridable by env**
(`TSKFLW_CONFIG_HOME`), which is not a nicety: the test suite is `t.TempDir()`-disciplined
throughout and **nothing in CI may read or write a real `$HOME`**.

### 3. It is the user config, not just a registry

`[theme]` and `[pager]` are documented in `config.go` as "local-terminal concerns," yet
today they can only be set in a *repo* config — so a preference about your terminal must be
repeated in every project, and a shared planning repo carries one contributor's taste.
The home config gives them their proper tier:

**flag > env > repo config > home config > built-in default.**

This is a small change to `themeName()` and the pager resolution, and it is worth shipping
on its own merits before any registry exists.

### 4. Schema

`config.toml` and `spaces.toml` are deliberately separate. The registry shape is:

```toml
schema_version = 1

[[space]]
id        = "taskflow"                  # local label; how this checkout is addressed
path      = "~/git/andy-esch/taskflow"  # repo dir, NOT the resolved planning root
verify_id = "6g1durable0aa"             # durable id of the resolved planning repo
label     = "tskflwctl"                 # optional display name
accent    = "cyan"                      # optional; else derived deterministically
added     = "2026-08-20"
```

- **`path` anchors at the repo (the `.tskflwctl.toml` dir), never the resolved root.**
  Registration is then *one* concept for both config modes: a pointer repo registered by
  path resolves through its `planning_repo` to the real planning root via ordinary
  discovery. Epic 23 compatibility falls out for free.
- **`id` is a local, shell-completable label**, defaulted from the checkout directory's
  basename. Label collisions are refused and require an explicit `--id`; the tool never
  invents a suffix. Multiple worktrees of one repo need distinct labels.
- **`verify_id` is the resolved planning repo's durable committed id.** Worktrees share
  it, so it cannot address a checkout; it instead detects a path that now resolves to the
  wrong tree. Path is the hint, `verify_id` the assertion.
- **Grouping is derived, never persisted.** Registry rows with the same `verify_id` form
  one logical space; discovery supplies `direct`/`pointer` role. The local `id` remains the
  address of one exact entry point. No parent, child, role, or group field is copied into
  `spaces.toml`, where it could drift from repo config.
- **Dedup on PHYSICAL paths** (`Abs` + `EvalSymlinks`), reusing the same helper
  `tracked_repos` already dedups with, so `../x`, an absolute path, and a symlinked
  checkout collapse to one entry.
- **`~` is stored unexpanded** so the file stays portable and committable.
- Writes are **atomic and surgical** — the same discipline as `setTrackedReposInText`:
  comments, key order, and unknown keys survive a write.

### 5. The registry is ADVISORY, never authoritative

**Nothing in the registry may change what `Discover` resolves from a given cwd.** Local
discovery is untouched: same walk-up, same `.tskflwctl.toml` precedence, same `.git`
boundary. A machine with no home config behaves **exactly** as today, and deleting the
file loses only convenience, never data or addressability.

The registry is planned to add one *opt-in* entry point: `--space <id>` (and
`TSKFLW_SPACE`) will resolve a label to a path, verify the target's durable id, and discover
from **there** instead of the cwd. Until that task ships, ordinary commands continue to
use cwd/`-C`; the shipped registry is inventory and health data only.

### 6. Registration: explicit now; `init` auto-registration planned

- **`space add [path]`** registers an existing repo, validated through discovery **before**
  anything is written — the "require + error, leave nothing behind" contract `InitPointer`
  already follows.
- **Today registration is explicit.** A separate planned task will make `init` append a
  `[[space]]` after a successful scaffold or pointer init. That side write will be
  best-effort like `LinkBack`, with `--no-register` / `TSKFLW_NO_REGISTER=1`; those flags
  do not exist until that task ships.
- **The atlas shows an unregistered current repo** dimmed, tagged `· not registered`, with
  a keypress to keep it. Discoverable without writing to `$HOME` behind your back.
- **Explicitly rejected: auto-registering any repo you happen to run in.** A read-only
  command must not write to `$HOME`, and throwaway clones and worktrees would accumulate.

### 7. Broken entries are diagnosed, never auto-removed

Paths rot: repos move, get deleted, get un-inited. Generalize epic 23's `LinkProblem` into a
`SpaceProblem{ID, Path, Kind, Message, Remedy}`, resolved lazily per entry and shared by
**both** `doctor` (a new registry section, human + `--json` + exit code) and the TUI.

| state | detection | surfaced as | remedy offered |
| --- | --- | --- | --- |
| `ok` | discovery succeeds | full card | enter |
| `missing` | path gone | dimmed, "not found at ~/…" | forget · repoint |
| `not-a-repo` | exists, no marker | dimmed, "no `.tskflwctl.toml` — init here?" | init · forget |
| `unreadable` | TOML / permission error | dimmed, verbatim parse error | `$EDITOR` · forget |
| `empty` | valid, zero entities | normal card, "no tasks yet" | enter anyway |
| `mismatch` | path resolves, but its durable id differs from `verify_id` | dimmed, expected/found ids | restore · forget/re-add |

No nearby-filesystem scan guesses where a missing repo moved. A truthful `missing` with an
explicit forget/re-add remedy is safer than accepting a same-named but unrelated checkout.

Three invariants make this delightful rather than nagging:

- **A broken entry is a card with a diagnosis, never an error screen**, and never blocks
  the atlas or the switcher. This is the philosophy `lint` already applies one scope down —
  a task with an unrecognized status is *listed and flagged*, never moved or dropped.
- **Never auto-forget.** Removal is always an explicit keypress or command.
- **One diagnosis function**, so the CLI and the TUI can never disagree about what is wrong
  with a space.

### 8. The atlas is a dashboard, not a tab

Per `docs/ARCHITECTURE.md`'s own rule of thumb — *a browsable list ⇒ a new `entityTab`; a
read-only orientation screen ⇒ the dashboard pattern* — the atlas is a second
dashboard-like screen: read-only, orients and routes, never mutates. It is **not** an
`entityTab`; spaces are not an entity inside a planning repo.

- It becomes the landing screen **only when more than one space is registered**. With zero
  or one, `ui` opens today's Overview — no regression for single-repo use.
- Each logical space's `core.Summary` loads **once as its own `tea.Cmd`**; cards render as skeletons and
  fill in as they land. This preserves the non-negotiable (no I/O in `Update`/`View`), keeps
  the screen fast with many spaces, and makes one slow or broken repo a slow or broken
  *card*.
- A card offers its registered direct/pointer entry points and later-discovered worktrees
  as checkout choices. Selecting one changes execution context, not the planning identity
  or the summary shown on the card.
- The **switcher** rides existing surfaces — the `ctrl+p` command palette and the `:`
  command bar — rather than claiming new keys on day one.

## Consequences

**Positive.**

- Answers the two questions that motivated this: cross-space "what's in progress" and
  "that repo, from here" (`--space`), the latter also a cross-repo handle for agents.
- **Purely additive.** Discovery is unchanged; no home config means today's behavior,
  exactly. Nothing becomes unaddressable if the file is lost.
- Fixes a standing wart for free: user-level `[theme]`/`[pager]` finally live at the tier
  their own doc comments describe.
- Reuses proven seams rather than inventing: surgical TOML writes, physical-path dedup,
  best-effort side-writes with a stderr warning, the `LinkProblem`/`doctor` reporting shape,
  and the dashboard widget set (`miniBar`, `Percent`, `rollupCounts`, `relDateCells`).

**Negative / cost.**

- **A new config scope to reason about.** Precedence grows a tier, and "why is my theme
  X?" now has one more place to look. Mitigated by keeping the registry advisory and
  `doctor` explanatory.
- **The tool now writes outside the repo it was invoked in.** Novel for a local-first CLI.
  Mitigated by: best-effort only, opt-out flags, an env-overridable location, and never on a
  read-only command.
- **`$HOME` in the test matrix.** Every path that touches the home config must be
  env-redirected; a leak would make tests machine-dependent.
- **A real TUI refactor.** `Model` holds one service today; per-space state that survives a
  round trip needs a `spaceSession` bundle, which means splitting `model.go` — already on
  epic 21's god-file list. This feature is the forcing function, and that is a genuine cost
  to book, not a freebie.
- **Path rot is now a visible, permanent category of state** the tool must keep explaining.

## Out of scope (deferred — NOT decided here)

- **The atlas's visual design** — card layout, accent derivation, the cross-space rail.
  Decided in the epic once the CLI half is in use.
- **`space scan`** (option D as a convenience) — feeds the registry if it ever lands.
- **Persisted per-space stats / caching.** Skeleton cards that fill in are honest; cached
  numbers lie. Revisit only if load time proves it necessary.
- **Non-local spaces** (a remote or served planning repo). That is ADR-0004's `serve`
  territory; the present entry points are local paths, while the logical identity model
  deliberately does not equate a space with one path.
- **Per-space theming.** `accent` is a card identity cue, not a theme override.

## Amendments

<!-- Append-only, dated entries added AFTER this ADR is accepted. Format:
     ### 2026-08-16 — <what changed and why> -->

_None yet (still `proposed`)._

## Related

- The cross-repo config foundation this extends by one scope: epic
  [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md).
- The word "project," reserved at a different scope:
  [0002-adopt-projects](0002-adopt-projects.md).
- Stable-key identity, the philosophy `space.id` follows:
  [0003-stable-key-id-addressed-storage](0003-stable-key-id-addressed-storage.md).
- Where a non-local space would have to go:
  [0004-single-source-of-truth-serve-owns-git](0004-single-source-of-truth-serve-owns-git.md).
- The TUI this lands in: epic
  [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md).
- The `model.go` split this forces: epic
  [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md).
- ADR format this follows: [0001-adopt-adrs](0001-adopt-adrs.md).
