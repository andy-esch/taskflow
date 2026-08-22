---
schema: 1
id: 6g0ajre026c6
created: "2026-08-15"
description: Explores a home-scoped registry of planning repos (spaces), a --space handle for any command, and a cross-space TUI board (the atlas) — plus the existing seams to reuse and the open forks.
tags: [config, cli, tui, multi-repo, design]
updated_at: "2026-08-22"
---

# Multi-space: home registry and the atlas

## Question

A machine can have several planning repos — `taskflow`, `desirelines-planning`,
dotfiles, whatever comes next — and `tskflwctl` has no notion of a second one.
`internal/config/config.go` says so in its own doc comment: *"One planning repo per
product; no cross-product registry."* That holds because discovery is **cwd-anchored**:
`Discover` walks up from where you stand, and everything downstream is built for the
one root it lands on.

So two questions have no answer today, and this doc explores what it would take to
give them one:

- **"What am I in the middle of, anywhere?"** The in-progress working set is per repo
  because the tool has no concept of a second repo. Answering it means N terminals or
  N `cd`s.
- **"Show me that other repo's planning, from here."** `-C` exists but needs the path,
  which means you must already know it. There is no handle.

Reference point: **k9s**, whose start screen lists available contexts — the same
situation, and a useful counter-example. That list is boilerplate-heavy text and hard
to scan; the bar here is to carry the TUI's existing visual language rather than print
a table of paths.

This is an exploration, not a design: names, file locations, schema, and screen shapes
were all provisional when written. Several have since been settled — the vocabulary, the
two-file config layout, the package boundary — and each is marked inline with a date. The
rest still stands as a lean, not a commitment.

## Findings

### Most of the mechanism already exists one scope down

Worth cataloguing before building anything:

- **Epic 23's cross-repo config**: `planning_repo` (impl → planning), `tracked_repos`
  (planning → impls), both resolved on **physical** paths.
- **`CheckLinks` → `LinkProblem{Repo, Message}` → `doctor`** (human + `--json` +
  nonzero exit) — *exactly* the "diagnose a broken cross-repo link kindly" pattern,
  already shipped.
- **Best-effort side-writes**: `LinkBack` writes into another repo's config and warns
  to stderr on failure rather than failing the command.
- **Surgical TOML editing**: `setTrackedReposInText` rewrites one array and preserves
  comments, key order, and unknown keys.
- **Dashboard widgets**: `miniBar`, `theme.Percent`, `rollupCounts`, `relDateCells`,
  `epicGlyph` — a cross-space card is mostly composition of these.
- **A command palette** (`ctrl+p`, fuzzy jump to anything) and the `:` command bar — a
  switcher probably needs no new keys at all.
- **[reusable-workspace-discovery-seam](../tasks/6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)**
  — deferred 2026-06-28 because the seam "only pays off once a second adapter (web)
  exists to reuse them." A TUI holding N repos would be a second consumer. **Not
  un-deferred** — noting the possible new trigger, nothing more.

The gap is genuinely a **third config scope**, above the repo. Epic 23's keys are
product-scoped by design (they link the repos of *one* product) and should not be
stretched: an impl repo has no business appearing in a peer product's config.

### The naming is constrained, not free

**"Project" is unavailable** — [ADR-0002](../adrs/0002-adopt-projects.md) defines a
Project as a cross-cutting initiative *inside* one planning repo, and a second, outer
meaning for the same word would be the worst available ambiguity.

The lean at the time — **settled 2026-08-18, refined 2026-08-20**: a **space** is one
durable planning identity; a registered direct or pointer checkout is an **entry point**
into it; the **atlas** is the whole set, and the name of the TUI screen over it. The split
is deliberate — the item noun turns up everywhere
(config keys, a `--space` flag, error strings, docs) so it wants to be short and plain,
while the collective turns up once, on a screen, and can carry the flavor. Also
considered and rejected: `orbit`/`constellation` (an orbit is a path, not a place),
`station`/`bridge` (heavy for "a directory"; `bridge` collides with the networking and
git senses), `space`/`universe` (grandiose for three repos). Full rationale and the
accepted cost — `space` sitting next to `core.Workspace` — are recorded in
[decide-the-multi-space-vocabulary](../tasks/6g1erb0p5893-decide-the-multi-space-vocabulary-blocks-slice-2.md).

Regardless of the skin, the **internal type is `Workspace`** (the resolved config +
store + service bundle for one root) — the vocabulary the deferred seam task already
uses.

### A home config wants to exist regardless of the registry

`$XDG_CONFIG_HOME/tskflwctl/`, falling back to `~/.config/tskflwctl/`. Probably **not**
Go's `os.UserConfigDir()` — it returns `~/Library/Application Support` on darwin,
wrong for a dotfile-friendly CLI whose users commit their config.

The location needs an env override (`TSKFLW_CONFIG_HOME`), and not as a nicety: the
suite is `t.TempDir()`-disciplined throughout and **nothing in CI can be allowed to
read or write a real `$HOME`**. That constraint should shape the API from the start
rather than get retrofitted.

Separately, `[theme]` and `[pager]` are documented in `config.go` as "local-terminal
concerns," yet can only be set in a *repo* config — so a preference about your own
terminal has to be repeated per project, and a shared planning repo ends up carrying
one contributor's taste. A home config gives them the tier they always wanted:

**flag > env > repo config > home config > built-in default.**

Notable because it is **independently valuable** — a small change to `themeName()`,
useful with or without any of the multi-space ideas, and a cheap way to prove out the
home-config plumbing before betting anything on it.

### Two files, not one

> **Settled 2026-08-18.** Written first as a single `config.toml` holding `[[space]]`
> alongside `[theme]`. That was actively misleading: `userconfig`'s decoder models only
> `[theme]` and `[pager]`, so `[[space]]` entries placed there are **silently
> discarded** (audit 2026-08-18-multi-space-config-foundation, L4).

```toml
# ~/.config/tskflwctl/config.toml — hand-edited; the tool READS this and never writes it
[theme]
name = "neon"

[pager]
enabled = true
```

```toml
# ~/.config/tskflwctl/spaces.toml — TOOL-OWNED; rewritten wholesale, comments not preserved
schema_version = 1

[[space]]
id     = "taskflow"                    # stable key; how you address it
path   = "~/git/andy-esch/taskflow"    # the repo dir, not the resolved planning root
label  = "tskflwctl"                   # optional display name
accent = "cyan"                        # optional; else derived from id
added  = "2026-08-15"
```

Splitting them is what makes writes tractable at all: TOML in this repo is
**decode-only** (`toml.DecodeFile` is the sole call; every write is hand-rolled text),
so a tool-owned file can be re-marshalled wholesale instead of needing a surgical
array-of-tables editor that preserves a human's comments.

Ideas embedded in that schema, each arguable:

- **Anchor `path` at the repo** (the `.tskflwctl.toml` dir), not the resolved root.
  Registration then means one thing in both config modes: a pointer repo resolves
  through its `planning_repo` to the real planning root via ordinary discovery, and
  epic 23 compatibility falls out for free.
- **`id` as a stable key**, defaulted from the dir basename —
  [ADR-0003](../adrs/0003-stable-key-id-addressed-storage.md)'s philosophy of
  addressing by a durable key rather than a path that moves. Collisions resolved at
  registration, never silently.
- **Dedup on physical paths**, reusing the helper `tracked_repos` already uses, so
  `../x`, an absolute path, and a symlinked checkout collapse to one entry.
- **Store `~` unexpanded** so the file stays portable and committable.

### The registry should stay advisory

The invariant most worth protecting: **nothing in the registry may change what
`Discover` resolves from a given cwd.** Same walk-up, same marker precedence, same
`.git` boundary. A machine with no home config behaves exactly as today, and deleting
the file costs convenience only — never data, never addressability.

The one opt-in entry point would be `--space <id>` (plus `TSKFLW_SPACE`): resolve an
id to a path, discover from *there* instead of the cwd. That also hands agents a
cross-repo handle, which may turn out to be the more valuable half.

> **Caveat noted 2026-08-18.** The invariant and `--space` are in tension — `--space`
> resolves *through* the registry and changes what gets discovered. The honest framing
> is narrower than "the registry cannot affect discovery": cwd-anchored discovery is
> unchanged, and the registry only adds an *explicit, opt-in* alternative entry point.
> Worth stating that way in any ADR rather than as a blanket invariant.

### Registration — "init registers, softly"

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

Three principles seem to separate "delightful" from "nagging":

- **A broken entry is a card with a diagnosis, never an error screen**, and never
  blocks the board or the switcher. Same philosophy `lint` already applies one scope
  down — a task with an unrecognized status is *listed and flagged*, never moved or
  dropped.
- **Never auto-forget.** Removal is always explicit.
- **One diagnosis function** behind both surfaces.

### The board — shape, not spec

Per `docs/ARCHITECTURE.md`'s own rule of thumb (*a browsable list ⇒ a new `entityTab`;
a read-only orientation screen ⇒ the dashboard pattern*), this looks like a second
dashboard-like screen rather than a tab — spaces aren't an entity inside a planning
repo.

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

Ideas worth keeping:

- **Cards, not a list** — the k9s differentiator. Mostly composition of existing
  dashboard widgets; the new parts are the card frame and the layout math.
- **A per-space accent** derived deterministically from the id (hash → a curated
  subset of `design.Palette`), overridable in config, so a space is recognizable by
  color before you read its name.
- **The cross-space in-progress rail may be the actual payload.** Cards orient; the
  rail answers the question that genuinely can't be asked today.
- **Land only when >1 space is registered** — with zero or one, `ui` opens today's
  Overview, so single-repo use is untouched.
- **Per-card async loading**: each logical space's `core.Summary` once as its own `tea.Cmd`,
  skeletons that fill in. Keeps the no-I/O-in-`Update`/`View` non-negotiable, keeps
  the screen fast with many spaces, and makes one slow or broken repo a slow or broken
  *card*.
- **The switcher rides `ctrl+p` / `:`** rather than claiming new keys on day one.

### Planning identity and entry points — refined 2026-08-20

Real registration exposed a case the first sketch collapsed: `desirelines-planning` is
the direct planning checkout, while `desirelines` and `desirelines-deploy` are separate
implementation repos whose configs point to it. Three registry paths do **not** represent
three plans, summaries, or atlas cards.

Modern configs already encode the complete relationship. The direct repo carries
`id = "…"`; pointers carry the same value as `planning_repo_id = "…"`; registry rows
record it as `verify_id`. Therefore:

- one durable planning id = one logical space, one summary, one atlas card;
- each registry row remains a locally labeled entry point, because `--space <label>`
  must select an exact checkout rather than an abstract identity;
- `direct` / `pointer` role is derived from repo discovery, never copied into
  `spaces.toml`; broken entries retain their intended group through `verify_id` and use
  `unknown` when their role cannot be inspected;
- human `space list` may use a tree with the direct checkout as the preferred anchor,
  but the indentation means “shares planning data,” not parent/child filesystem ownership;
- JSON stays flat and carries derived `planning_id` + `role`, while `status --all` and the
  atlas consume the shared grouped projection.

This is the same conceptual rule the worktree design was reaching for — **one plan, one
card** — but durable ids generalize it across independently registered repos. Worktree
enumeration remains live/derived and nests under the selected entry point; it does not add
registry rows.

### Worktrees on the board — refined 2026-08-19

A card is **one planning identity**, not one registered path. Where an entry point has several
worktrees, the card offers them as selectable *variants* rather than becoming several
cards. Settled while stress-testing, after the worktree/relative-path findings
([worktrees-and-relative-paths-in-cross-repo-planning](6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)).

**Worktree discovery remains scoped to worktrees only.** Arbitrary other copies of a repo
— second clones, rsynced directories — are not discovered automatically. Explicitly
registered copies that carry the same durable planning id can still join the identity
group; their local registry labels distinguish their checkouts.

**Grouping is free.** Every checkout of a repo derives the same **common gitdir**: read
`.git` (a dir for the base checkout, a file for a worktree), and for the file case strip
the `/worktrees/<name>` suffix from its `gitdir:`. Verified across a base checkout and two
worktrees — identical key.

This matters for sequencing: grouping needs **no committed repo id**, so the board does
NOT block on layer C. (Layer C still earns its keep for the wrong-tree guard, which is a
resolution-correctness concern, not a rendering one. Grouping across *clones* would have
required it — dropping clones from scope is what removed the dependency.)

**Discovery is free too.** Worktrees never register. They are enumerated live from a
registered checkout by reading `<common-gitdir>/worktrees/*/gitdir` — no git binary, no
subprocess. A stale worktree is self-evident: its path does not exist. `desirelines-planning`
carries exactly one such entry today (`/workspace/dl-sweepport`), which is the shape a
dead variant takes.

That answers the original worry directly: worktrees are **derived** from a registered
checkout, so nothing new needs registering and the registry gains no corpses.

#### Card shape

```
┌ desirelines (~/git/andy-esch/desirelines) ┐
│ epics                                     │
│    .....                                  │
├───────────────────────────────────────────┤
│ [3 worktrees found]                       │
└───────────────────────────────────────────┘
```

- **The path lives in the header.** It is the only thing distinguishing two variants that
  otherwise render identically.
- **The branch belongs there too.** Worktrees of one repo differ by branch, and their
  paths are often near-identical (`-wt-go-mod`, `-wt-tf-cleanup`). Note `workspace` does
  not report the branch today — that gap is now load-bearing for this header.
- **Always default to the base worktree**, with a visible marker for it, so "am I on the
  default?" is answerable at a glance rather than from memory.
- **`[N worktrees found]` is a discovery affordance, not N more cards.** One card per
  repo is the whole point.

#### Friction on a selected variant

Mutating through a *selected* non-base variant requires a warning + confirmation. Two
refinements that the naive version gets wrong:

- **It needs a headless counterpart.** `Gate.On()` opens only on a TTY, and off a TTY the
  gate is closed so agents never block — so a confirmation implemented purely as a prompt
  would silently no-op for exactly the caller most likely to get this wrong. Shape:
  confirm on a TTY, **refuse** off one, with an explicit flag to proceed.
- **Attach the friction to the SELECTION, not to worktree-ness.** The hazard is
  *forgetting you switched* on the board. When the cwd simply *is* a worktree, the user is
  deliberately there and no confirmation is warranted — and gating that case would put a
  prompt on every `task complete` in a taskflow feature branch, since this repo
  self-hosts its planning. A prompt that fires constantly is one you reflex through,
  which costs you the one time it matters.

#### Where this is most valuable

Chiefly the **self-hosted** case. A decoupled planning repo run trunk-only (ADR-0004)
rarely has meaningfully divergent worktrees; taskflow, with `taskflow_root = "./planning"`,
has a different planning tree on every branch. Local divergence is real and expected;
a hosted/served deployment is where it would show up most sharply.

### Costs, booked honestly

- **A real TUI refactor.** `Model` holds one service today; per-space state that
  survives a round trip needs something like a `spaceSession{svc, layout, tabs,
  dashboard, watcher}`, which means splitting `model.go` — already on epic 21's
  god-file list. This feature would be the forcing function. Biggest single cost.
- **The watcher.** One fsnotify watcher on one root today. Following the active space
  only (plus manual refresh and a slow tick for cards) is the cheap correct answer;
  watching all spaces is fd-heavy.
- **A new config scope to reason about.** Precedence grows a tier and "why is my theme
  X?" gains a place to look.
- **The tool writing outside the repo it was invoked in** — novel for a local-first
  CLI, even best-effort.
- **`$HOME` in the test matrix**, or tests become machine-dependent.
- **Path rot becomes a permanent, visible category of state** the tool has to keep
  explaining.

### Open forks

1. ~~**Naming.**~~ **Settled 2026-08-18: `space` / `atlas`** — see the naming section
   above. The accepted cost is `space`'s overlap with `core.Workspace`; if it ever
   grates, the item noun is a rename of one TOML key, one flag, and one command group,
   and the durable data (`id`, `path`) is unaffected.
2. **Is the cross-space board the point, or is `--space` the point?** Plausible that
   the flag plus a cross-space `status --all` delivers most of the value, and the board
   is a nice-to-have that costs a `model.go` refactor.
3. **Registry vs. `tracked_repos` overlap.** Both record "other repos." Kept separate
   here on the argument that they're different axes (one user's products vs. one
   product's repos) — but is that distinction obvious to anyone but its author?
4. ~~**One file or two?**~~ **Settled 2026-08-18: two** (see above). Residual: whether
   `spaces.toml` needs a `schema_version` migration path before it has any users.
5. **Per-space state on switch** — keep cursors/filters per space (nice, more state)
   or reset (simple, mildly annoying)?
6. **Is `moved` detection worth it at all**, or is "missing + repoint" honest enough?
   Guessing where a repo went is the kind of cleverness that misfires.
7. **What does this mean for `serve`/web (epic 19)?** A served planning repo isn't a
   local path, so `space` as "a local directory" may be the wrong long-run
   abstraction. Deliberately not answered here.

## Recommendation (as of 2026-08-15)

Proceed in slices ordered so the cheap, independently-useful parts come first and the
expensive commitment comes last — and **re-decide at step 3** rather than treating the
board as a foregone conclusion:

1. **Home config** — location, load/save, env override, `theme`/`pager` precedence.
   Useful on its own, commits to no multi-space vocabulary at all.
2. **Registry + CLI** — `space list|add|forget`, `--space`, `init` auto-register, a
   `doctor` section, `status --all`. Pure CLI, fully testable, no TUI risk.
3. **Re-decide.** Having lived on the CLI half, is the board still wanted?
4. Only if yes: the `Resolve() → Workspace` seam, the `spaceSession` refactor, the
   atlas, the switcher.

**What would change the call.** If step 2 lands and `--space` + `status --all` already
answers "what am I in the middle of, anywhere?", the board is decoration and steps 3–4
should be dropped — the `model.go` split is too large a cost for a nicer rendering of
an answer you already have. Conversely, if epic 19's `serve` work starts first, fork 7
becomes urgent: "a space is a local directory" would be the wrong abstraction to have
baked into a config schema by then.

> **Progress note, 2026-08-20.** Slice 2's two prerequisites are now *implemented*, not
> just decided. The vocabulary is locked (`space` / `atlas`), and identity shipped in epic
> 23: `init` mints a durable id into a planning repo's committed config, and a pointer that
> records `planning_repo_id` verifies it after resolving. So a registry entry's `verify_id`
> now references something real rather than a hypothetical. Also shipped: worktree-aware
> resolution, which means the atlas's "group by repo, select the checkout" design needs no
> committed id for grouping — the common gitdir does it for free.

> **Progress note, 2026-08-21.** The CLI experiment is now concrete: `status --all`
> returns one compact summary per durable planning identity and one space-badged
> in-progress rail. It reused the then-current `spacehealth.Group`, preferred a healthy
> direct checkout, fell back to a healthy pointer, and left dead entry-point diagnoses
> inside the affected group without failing the sweep. A smoke run over six registered
> entry points produced four logical summaries and five in-progress rows; the rail remains
> the strongest evidence for the feature, while the compact summaries were useful context.
>
> **Progress note, 2026-08-22.** The registry architecture is now consolidated behind
> `core.SpaceRegistryService`: catalog grouping, explicit selection, mutations,
> completion, and doctor share one core projection and consumer-owned port.
> `core.SpaceOverviewService` composes that catalog with a separate read-only planning-tree
> opener, so an atlas or served adapter can reuse it without importing home-registry
> storage. The August 21 behavior and wire output are unchanged.
>
> **Progress note, 2026-08-18.** Slice 1 shipped (commit `86e3e1d` + follow-ups):
> `internal/userconfig`, four-tier theme/pager precedence, and the test-isolation
> harness. Two decisions from the sketch were settled during it — two files rather than
> one, and a separate package so `internal/config` cannot import home-scope data (now
> enforced by a `depguard` rule). Reviewed adversarially in
> audit 2026-08-18-multi-space-config-foundation. The vocabulary was locked the same
> day (`space` / `atlas`), so slice 2 is unblocked.

## Related

- Cross-repo config this would sit above: epic
  [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)
- The epic this doc informs: epic
  [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
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
  [dashboard-extension-ideas](6fgcr2402att-dashboard-extension-ideas.md)
