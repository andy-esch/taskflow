---
status: reference
created: "2026-08-15"
tags: [config, cli, tui, multi-repo, design]
---

# Many planning repos, one atlas — a home registry and a cross-space TUI board (sketch)

> **Nothing here is decided.** This is a directionally-correct sketch, not a design
> doc and not an ADR. Names, file locations, schema, and screen shapes are all
> provisional and expected to move. When (if) the shape settles, the durable parts
> get promoted into **ADR-0005** and the rest is deleted. Read every "leaning"
> below as *"this is where the thinking currently points,"* never as a commitment.

## The itch

A machine can have several planning repos — `taskflow`, `desirelines-planning`,
dotfiles, whatever comes next — and `tskflwctl` has no notion of a second one.
`internal/config/config.go` says so in its own doc comment:

> *"One planning repo per product; no cross-product registry."*

That holds because discovery is **cwd-anchored**: `Discover` walks up from where you
stand and everything downstream is built for the one root it lands on. Two questions
have no answer today:

- **"What am I in the middle of, anywhere?"** The in-progress working set is per
  repo because the tool has no concept of a second repo. Answering it means N
  terminals or N `cd`s.
- **"Show me that other repo's planning, from here."** `-C` exists but needs the
  path, which means you must already know it. There is no handle.

Reference point: **k9s**, whose start screen lists available contexts. Same
situation, and a useful counter-example — that list is boilerplate-heavy text and
hard to scan. The bar here is to carry through the TUI's existing visual language
rather than print a table of paths.

## What already exists (don't rebuild it)

Worth cataloguing, because most of the mechanism is present one scope down:

- **Epic 23's cross-repo config**: `planning_repo` (impl → planning),
  `tracked_repos` (planning → impls), both resolved on **physical** paths.
- **`CheckLinks` → `LinkProblem{Repo, Message}` → `doctor`** (human + `--json` +
  nonzero exit). This is *exactly* the "diagnose a broken cross-repo link kindly"
  pattern, already shipped.
- **Best-effort side-writes**: `LinkBack` writes into another repo's config and
  warns to stderr on failure rather than failing the command.
- **Surgical TOML editing**: `setTrackedReposInText` rewrites one array and preserves
  comments, key order, and unknown keys.
- **Dashboard widgets**: `miniBar`, `theme.Percent`, `rollupCounts`, `relDateCells`,
  `epicGlyph` — a cross-space card is mostly composition of these.
- **A command palette** (`ctrl+p`, fuzzy jump to anything) and the `:` command bar —
  a switcher probably needs no new keys at all.
- **`6fgcr2403sjn`** — *"reusable workspace discovery seam: lift init/doctor/fix off
  the cli … extract `Resolve() → Workspace`"*. Deferred 2026-06-28 because the seam
  "only pays off once a second adapter (web) exists to reuse them." A TUI holding N
  repos would be a second consumer. **Not un-deferred** — noting the possible new
  trigger, nothing more.

The gap is genuinely a **third config scope**, above the repo. Epic 23's keys are
product-scoped by design (they link the repos of *one* product) and should not be
stretched: an impl repo has no business appearing in a peer product's config.

## Naming — provisional

Constrained, not free: **"project" is unavailable.** ADR-0002 defines a Project as a
cross-cutting initiative *inside* one planning repo. A second, outer meaning for the
same word would be the worst available ambiguity.

Current lean, both replaceable:

- a **space** — one registered planning repo, the unit you switch between
- the **atlas** — the whole set, and the name of the TUI screen over it

The split is deliberate: the item noun turns up everywhere (config keys, a `--space`
flag, error strings, docs) so it wants to be short and plain; the collective turns up
once, on a screen, and can carry the flavor. Pairs also considered:
`orbit`/`constellation`, `station`/`bridge`, `space`/`universe`.

Regardless of the skin, the **internal type is `Workspace`** (the resolved
config + store + service bundle for one root) — the vocabulary
`6fgcr2403sjn` already uses.

## Where this is leaning

### A home config, whose first citizen is a registry

`$XDG_CONFIG_HOME/tskflwctl/config.toml`, falling back to
`~/.config/tskflwctl/config.toml`. Probably **not** Go's `os.UserConfigDir()` — it
returns `~/Library/Application Support` on darwin, wrong for a dotfile-friendly CLI
whose users commit their config.

The location almost certainly needs an env override (`TSKFLW_CONFIG_HOME`), and not
as a nicety: the suite is `t.TempDir()`-disciplined throughout and **nothing in CI
can be allowed to read or write a real `$HOME`**. This is the kind of constraint
that should shape the API from the start rather than get retrofitted.

### It probably wants to be the user config, not only a registry

`[theme]` and `[pager]` are documented in `config.go` as "local-terminal concerns,"
yet today they can only be set in a *repo* config — so a preference about your own
terminal has to be repeated per project, and a shared planning repo ends up carrying
one contributor's taste. A home config would give them the tier they always wanted:

**flag > env > repo config > home config > built-in default.**

Notable because it's **independently valuable**: a small change to `themeName()`,
useful with or without any of the multi-space ideas, and a cheap way to prove out
the home-config plumbing before betting anything on it.

### Sketch schema (expect this to move)

```toml
schema_version = 1

[theme]
name = "neon"

[[space]]
id     = "taskflow"                    # stable key; how you address it
path   = "~/git/andy-esch/taskflow"    # the repo dir, not the resolved planning root
label  = "tskflwctl"                   # optional display name
accent = "cyan"                        # optional; else derived from id
added  = "2026-08-15"
```

Ideas embedded in that, each arguable:

- **Anchor `path` at the repo** (the `.tskflwctl.toml` dir), not the resolved root.
  Registration then means one thing in both config modes: a pointer repo resolves
  through its `planning_repo` to the real planning root via ordinary discovery, and
  epic 23 compatibility falls out for free.
- **`id` as a stable key**, defaulted from the dir basename — ADR-0003's philosophy
  of addressing by a durable key rather than a path that moves. Collisions resolved
  at registration, never silently.
- **Dedup on physical paths**, reusing the helper `tracked_repos` already uses, so
  `../x`, an absolute path, and a symlinked checkout collapse to one entry.
- **Store `~` unexpanded** so the file stays portable and committable.
- **Atomic + surgical writes**, matching `setTrackedReposInText`'s discipline.

### The registry should stay advisory

The invariant that seems most worth protecting: **nothing in the registry may change
what `Discover` resolves from a given cwd.** Same walk-up, same marker precedence,
same `.git` boundary. A machine with no home config behaves exactly as today, and
deleting the file costs convenience only — never data, never addressability.

The one opt-in entry point would be `--space <id>` (plus `TSKFLW_SPACE`): resolve an
id to a path, discover from *there* instead of the cwd. That also happens to hand
agents a cross-repo handle, which may turn out to be the more valuable half.

### Registration — leaning toward "init registers, softly"

- **`init` appends a `[[space]]`** after a successful scaffold or pointer init, as one
  more `+` line in its existing output. **Best-effort, exactly like `LinkBack`**: an
  unwritable or missing `$HOME` warns to stderr and never fails the init. Opt out via
  `--no-register` / `TSKFLW_NO_REGISTER=1`, mirroring `--no-link-back`.
- **`space add [path]`** for repos that already exist, validated through discovery
  *before* anything is written — the "require + error, leave nothing behind" contract
  `InitPointer` already follows.
- **The atlas shows an unregistered current repo** dimmed and tagged
  `· not registered`, with a keypress to keep it. Discoverable without writing to
  `$HOME` behind your back.
- **Leaning against auto-registering any repo you run in.** A read-only command
  writing to `$HOME` is surprising, and throwaway clones and worktrees would pile up.
- **`space scan ~/git`** is tempting and probably a trap — it surfaces every stale
  checkout. If it ever lands it should *feed* the registry, dry-run first, through the
  existing `prompt.SelectOne` picker.

### Broken entries: diagnose, never auto-remove

Paths rot. The generalization that suggests itself is epic 23's `LinkProblem` widened
into a `SpaceProblem{ID, Path, Kind, Message, Remedy}`, resolved lazily per entry and
shared by **both** `doctor` (a registry section) and the TUI — so the two can't
disagree about what's wrong.

| state | detection | surfaced as | remedy offered |
| --- | --- | --- | --- |
| `ok` | discovery succeeds | full card | enter |
| `missing` | path gone | dimmed, "not found at ~/…" | forget · repoint |
| `moved` | gone, but a same-named planning repo nearby | dimmed, "moved to …?" | accept |
| `not-a-repo` | exists, no marker | dimmed, "no `.tskflwctl.toml` — init here?" | init · forget |
| `unreadable` | TOML / permission error | dimmed, verbatim parse error | `$EDITOR` · forget |
| `empty` | valid, zero entities | normal card, "no tasks yet" | enter anyway |

Three principles that seem to separate "delightful" from "nagging":

- **A broken entry is a card with a diagnosis, never an error screen**, and never
  blocks the board or the switcher. Same philosophy `lint` already applies one scope
  down — a task with an unrecognized status is *listed and flagged*, never moved or
  dropped.
- **Never auto-forget.** Removal is always explicit.
- **One diagnosis function** behind both surfaces.

### The board — shape, not spec

Per `docs/ARCHITECTURE.md`'s own rule of thumb (*a browsable list ⇒ a new
`entityTab`; a read-only orientation screen ⇒ the dashboard pattern*), this looks
like a second dashboard-like screen rather than a tab — spaces aren't an entity
inside a planning repo.

```
  ATLAS                                    4 spaces · 7 in progress

  ╭─ taskflow ───────────── 2h ─╮  ╭─ desirelines ────────── 3d ─╮
  │ ●●●●●●●○○○  62%  184/297    │  │ ●●●●○○○○○○  41%   38/92     │
  │ ▸ 3 active    ⚠ 2 attention │  │ ▸ 1 active                  │
  │ 12 epics · 4 open audits    │  │ 5 epics                     │
  ╰─────────────────────────────╯  ╰─────────────────────────────╯

  ╭─ dotfiles ─────────────────╮  ╭─ old-thing ─────── missing ─╮
  │ ●●●●●●●●●● 100%   18/18    │  │ not found at ~/git/old-thing│
  │ all clear                  │  │ x forget · p repoint        │
  ╰────────────────────────────╯  ╰─────────────────────────────╯

  ACROSS ALL SPACES · in progress
  ● taskflow    2h   route-tui-chrome-through-the-palette
  ● taskflow    5h   honor-c-columns-and-compact-output-for-json
  ● desirelines 3d   segment-matcher-tolerance
```

Ideas worth keeping from that sketch:

- **Cards, not a list** — the k9s differentiator. Mostly composition of existing
  dashboard widgets; the new parts are the card frame and the layout math.
- **A per-space accent** derived deterministically from the id (hash → a curated
  subset of `design.Palette`), overridable in config, so a space is recognizable by
  color before you read its name.
- **The cross-space in-progress rail may be the actual payload.** Cards orient; the
  rail answers the question that genuinely can't be asked today. Possibly co-equal
  with the cards rather than a garnish.
- **Land only when >1 space is registered** — with zero or one, `ui` opens today's
  Overview, so single-repo use is untouched.
- **Per-card async loading**: each space's `core.Summary` as its own `tea.Cmd`,
  skeletons that fill in. Keeps the no-I/O-in-`Update`/`View` non-negotiable, keeps
  the screen fast with many spaces, and makes one slow or broken repo a slow or
  broken *card*.
- **The switcher rides `ctrl+p` / `:`** rather than claiming new keys on day one.

## Known costs (booked honestly, not hand-waved)

- **A real TUI refactor.** `Model` holds one service today; per-space state that
  survives a round trip needs something like a `spaceSession{svc, layout, tabs,
  dashboard, watcher}`, which means splitting `model.go` — already on epic 21's
  god-file list. This feature would be the forcing function. Biggest single cost.
- **The watcher.** One fsnotify watcher on one root today. Following the active space
  only (plus manual refresh and a slow tick for cards) is the cheap correct answer;
  watching all spaces is fd-heavy.
- **A new config scope to reason about.** Precedence grows a tier and "why is my
  theme X?" gains a place to look.
- **The tool writing outside the repo it was invoked in** — novel for a local-first
  CLI, even best-effort.
- **`$HOME` in the test matrix**, or tests become machine-dependent.
- **Path rot becomes a permanent, visible category of state** the tool has to keep
  explaining.

## Open questions (the actual forks)

1. **Naming.** `space`/`atlas` is a lean, not a decision. Does `space` survive
   contact with `--space`, `[[space]]`, and a year of error messages? Is `atlas` still
   good on the tenth look, or does it read as decoration?
2. **Is the cross-space board the point, or is `--space` the point?** Plausible that
   the flag plus a cross-space `status --all` delivers most of the value, and the board
   is a nice-to-have that costs a `model.go` refactor. Worth deliberately testing by
   shipping the CLI half first and seeing what's still missing.
3. **Registry vs. `tracked_repos` overlap.** Both record "other repos." Kept separate
   here on the argument that they're different axes (one user's products vs. one
   product's repos) — but is that distinction obvious to anyone but its author?
4. **Does the home config want to be one file or two?** (`config.toml` with
   `[[space]]` vs. a separate `spaces.toml`.) One file couples registry churn to
   preference edits; two files means two locations to explain.
5. **Per-space state on switch** — keep cursors/filters per space (nice, more state)
   or reset (simple, mildly annoying)?
6. **Is `moved` detection worth it at all**, or is "missing + repoint" honest enough?
   Guessing where a repo went is the kind of cleverness that misfires.
7. **What does this mean for `serve`/web (epic 19)?** A served planning repo isn't a
   local path, so `space` as "a local directory" may be the wrong long-run
   abstraction. Deliberately not answered here.

## Possible sequencing (if it proceeds)

Ordered so the cheap, independently-useful parts come first and the expensive
commitment comes last:

1. Home config: location, load/save, env override, `theme`/`pager` precedence.
   **Useful on its own**, no multi-space commitment.
2. Registry model + `space list|add|forget` + `--space` + `init` auto-register +
   a `doctor` section + `status --all`. Pure CLI, fully testable, no TUI risk.
3. *Then decide whether the board is still wanted*, having lived on the CLI half.
4. Only if yes: the `Resolve() → Workspace` seam (`6fgcr2403sjn`), the
   `spaceSession` refactor, the atlas, the switcher.

## Related

- Cross-repo config this would sit above: epic
  [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)
- The word "space" exists because "project" is taken:
  [ADR-0002](../adrs/0002-adopt-projects.md)
- Stable-key identity, the philosophy `space.id` would follow:
  [ADR-0003](../adrs/0003-stable-key-id-addressed-storage.md)
- Where a non-local space would have to live:
  [ADR-0004](../adrs/0004-single-source-of-truth-serve-owns-git.md)
- The seam this would reuse:
  [reusable-workspace-discovery-seam](../tasks/6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)
- The TUI this would land in: epic
  [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Prior art for extending the landing screen:
  [2026-06-27-dashboard-extension-ideas](2026-06-27-dashboard-extension-ideas.md)
