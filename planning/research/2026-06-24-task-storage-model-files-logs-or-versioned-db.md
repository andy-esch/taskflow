---
status: reference
created: "2026-06-24"
tags: [storage, data-model, git, merge, occ, epic-23]
---

# Task storage model — files, logs, or a versioned DB

A spike off two live threads: the **content-aware OCC** work and the question of
whether **status-as-directory** should exist. The pain that triggered it: git
branches that *move* a task (rename across status dirs) on one side and *edit its
body* on the other merge into **duplicate files in two dirs with different
content** — and flattening to one big `tasks/` bucket "fixes" that only by forcing
you to run the tool to find what's in-progress. This weighs storage models that
keep the **git-native, browsable, plain-text** feel without either failure mode,
including non-document and database-flavored options. No code here.

## The reframe: three independent axes

The instinct to reach for "a database" conflates things that are actually
separable. A task store has **three orthogonal choices**:

1. **Format** — how *one* entity serializes. Markdown+frontmatter today; could be
   TOML/JSON. **Mostly irrelevant to the pain** — switching md→json doesn't change
   merges or browsability in a way you care about.
2. **Layout** — what the *path/tree* encodes, i.e. the organizing key. Today it's
   **status**, which is *mutable* — so a status change is a **rename**, and renames
   are exactly what break git merges and force atomic-rename on remotes.
3. **Representation** — **current-state snapshot** (today: the file *is* the state)
   vs an **append-only event log** (state is folded from events).

Your two problems map cleanly:
- **Duplicate-on-merge** = axis 2 only (organizing by a *mutable* key).
- **"Huge messy bucket"** = axis 2 again (no organizing key at all).

Neither is a format or database problem. The fix is to **organize by something
that doesn't change**, and (optionally) **generate** the by-status view rather
than *storing* it as the tree. A key enabler: **`status:` already lives in
frontmatter today** — the directory is only the *authoritative* copy — so making
frontmatter the source of truth is a smaller move than it sounds.

## TL;DR

| | Recommendation |
| :-- | :-- |
| Stay git-native + browsable + plain files? | **Yes — you don't need a DB to fix this.** |
| The actual fix | Organize the tree by a **stable key** (identity or epic), keep `status` in frontmatter, and **generate** a browsable board for "what's in-progress." Kills the rename-duplicate class, avoids the flat bucket, and makes OCC trivial. |
| Want history/audit + best-possible merges? | An **append-only event log** is the most git-merge-friendly model — but you browse a *projection*, not the raw store. |
| Want real DB queries/integrity? | **SQLite as an uncommitted cache** (text stays canonical) gives query speed for free; **Dolt** ("git for data") is the only thing that's a *real* versioned DB, but it leaves plain-git. |

**Lean:** stay files; switch the organizing key from *status* to *stable
identity/epic*; add a generated board. Treat event-log and DB options as later
forks if history or scale demand them.

## Decision (2026-06-24)

Direction chosen: **stable-key layout + a generated board** (option A3). The two
problems — merge duplicates and the "huge bucket" — are both resolved by
organizing the tree by something *immutable* and **generating** the by-status view
instead of storing it as the tree.

The board is not just a file — it's the **materialized read model**: one
projection of planning state, surfaced through several adapters. That's the
convergence with the web companion (epic 19): the generated `BOARD.md`, the TUI's
board, a CLI `board`/`status`, and the web API's read endpoint are all the **same
`core` projection** rendered differently. Design rule that falls out of this:

> Define the projection **once in `core` as structured data** (not a pre-rendered
> string), and let each adapter render it — markdown, JSON, HTML, TUI. Writes stay
> funneled through `core.Service` (with the version-aware OCC).

So building the projection now (for the board) is a **down payment on the web
companion's read side** — the `serve` read endpoint becomes "expose the projection
as JSON," not new logic. `core.Summary()` is the embryonic version of this today;
the board generalizes it into the canonical read model.

Still open: the grouping sub-choice — **dir-per-epic (A3a)** vs **flat-by-slug +
board (A3b)** — and whether to add an immutable `id` distinct from the slug (so an
epic-move never breaks identity/cross-links). See open questions.

## Option group A — layout (keep one-file-per-task)

### A1. Status-as-directory (today)
`tasks/in-progress/foo.md`. ✅ `ls`-browsable, self-describing, drift-proof
("folder wins"). ❌ every status change is a rename → the merge-duplicate class,
non-atomic cross-dir moves, and the rename-aware path-CAS that complicates OCC and
remotes.

### A2. Flat + frontmatter status
`tasks/foo.md`, `status:` in frontmatter. ✅ stable path → clean 3-way merges, no
duplicates, trivial OCC, easy remote. ❌ the "huge bucket" — you must run the tool
(or grep) to see what's in-progress. *This is the one you balked at.*

### A3. Group by a **stable** key + generated board  ⭐
Organize by something that rarely/never changes, put `status` in frontmatter, and
**generate** the by-status view instead of storing it as the tree. Two sub-shapes:

- **A3a — by epic:** `tasks/18-tui/foo.md`. Epics are a natural partition (no
  giant bucket), browsable by area, and epic changes are *rare* (vs status changes
  every few hours), so renames almost never happen. A board (`BOARD.md` or
  `tskflwctl status`) covers the by-status view.
- **A3b — flat by identity + board:** `tasks/foo.md` (A2) **plus** a generated,
  committed-or-gitignored `BOARD.md`/`views/in-progress.md` that lists each
  status's tasks with links. You regain "open one file, see what's in-progress"
  without the tree encoding status. The board is *derived*, so it never causes a
  merge duplicate — a merge just regenerates it.

Both kill the rename-duplicate class and make OCC "edit one stable file if version
matches." A3a keeps filesystem browsing; A3b leans on the generated board. **The
generated-view idea is the crux** — it decouples "how it's stored" from "how you
browse it," which is exactly the database benefit you're sensing, achieved with a
text projection instead of a DB.

## Option group B — representation

### B1. Append-only event log (git-native, CRDT-flavored)
Each task is a sequence of events (`created`, `body-set`, `moved`, `field-set`)
appended to a per-task log (`tasks/foo.log.jsonl`) or a global log; current state
is **folded** from events. ✅ Appends from different branches **concatenate** —
git merges them almost conflict-free; you get full **history/audit/blame** for
free; a status change is an *event*, never a rename. ❌ The raw store isn't
hand-browsable — you need a projection/snapshot to read or edit current state, and
"fold order" needs a tiebreak (timestamp/lamport). Most powerful merges, least
direct readability. Good if audit trail is a goal; heavier model otherwise.

> A snapshot-per-task *plus* a log (event log as truth, snapshot as the readable
> projection) is the hybrid — it's basically A3 with history bolted on.

## Option group C — store paradigm

### C1. Plain files (today's family) — A1/A2/A3/B1 all live here
Zero new dependencies, works in any git repo, reviewable in PRs. The whole `core.Store`
port already assumes nothing beyond "a backend"; these are layout/representation
choices *within* the file paradigm.

### C2. SQLite as an **uncommitted cache** (text stays canonical)
Keep files as the committed source of truth; build a local, git-ignored SQLite
index for fast queries/joins/counts as the tree grows. ✅ Real query power without
giving up git or browsability; the DB is disposable (rebuild from files). ❌ Cache
invalidation; the DB doesn't help merges (the files still do). A pure performance
play, orthogonal to the layout fix — adopt only if scale bites.

### C3. Dolt — "git for data" (a real versioned SQL DB)
A SQL database with native branch/merge/diff at the **row/cell** level. ✅ The only
option that's genuinely *a database* **and** version-controlled with git-like
semantics — strongest query/integrity/merge-granularity story; conflicts are
per-cell, not per-file. ❌ It's a **separate system**, not "files in a normal git
repo": a heavyweight dependency, its own remote/clone/push model, no `cat`/`grep`
of plain text, and it abandons the "plain markdown anyone can read/PR" identity
that defines this project. The right answer if the project's center of gravity
ever shifts from "documents" to "data"; a large bet otherwise.

### C4. CRDT documents (Automerge/Yjs) — noted and rejected
True conflict-free merges, but the on-disk form is binary/opaque JSON — not
human-diffable or PR-reviewable. Fails the git-native-in-spirit test.

## Comparison

| | git-native / diffable | browse w/o tool | avoids huge bucket | no rename-on-status | OCC-natural | query power | hand-editable | cost |
| :-- | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
| A1 status-dir (today) | ✅ | ✅✅ | ✅ | ❌ | ❌ | ➖ | ✅ | — |
| A2 flat + frontmatter | ✅ | ❌ | ❌ | ✅ | ✅ | ➖ | ✅ | low |
| **A3 stable key + board** | ✅ | ✅ | ✅ | ✅ | ✅ | ➖ | ✅ | low-med |
| B1 event log | ✅ | ❌ (projection) | ✅ | ✅ | ✅✅ | ➖ | ❌ | med-high |
| C2 files + sqlite cache | ✅ | ✅ | ✅ | (per layout) | (per layout) | ✅ | ✅ | med |
| C3 Dolt | ❌ (not plain git) | ❌ | ✅ | ✅ | ✅✅ | ✅✅ | ❌ | high |

## The OCC through-line

Every option except A1 makes the content-OCC work **simpler**, because identity
stops moving:

- **A2 / A3:** a stable path → OCC is "read version, write iff version matches" on
  one file. The rename-aware path-CAS goes away; remote backends need only
  per-object preconditions (etag/generation), not atomic rename.
- **B1:** appends are near-commutative — OCC degrades to "append, fold, tiebreak,"
  with conflicts rare by construction.
- **C2/C3:** transactions/row-versions give OCC natively.

So whichever way you go, **decide the layout before locking OCC's shape** — A1 is
the only one that forces the harder, rename-aware OCC you'd then throw away.

## Recommendation

1. **Switch the organizing key from *status* (mutable) to a *stable* key, status
   in frontmatter** — option **A3**. This is the targeted fix: it removes the
   rename, so the merge-duplicate class disappears, OCC becomes trivial, and
   remote backends get easy — *without* a database and *without* abandoning plain
   files. Because `status:` is already in frontmatter, the data migration is mostly
   "stop treating the directory as authoritative."
2. **Add a generated board** for browsability (`tskflwctl board`/`status` writing a
   `BOARD.md`, or a `views/` tree) so you keep "open a file, see what's
   in-progress" that flattening (A2) would cost you. Generated ⇒ never a merge
   conflict.
3. **Pick the grouping:** A3a (dir-per-epic — keeps real filesystem browsing,
   epic-change is the only rare rename) vs A3b (flat + rely on the board). A3a fits
   the "I don't want a messy bucket" instinct best.
4. **Defer the DB:** add the **SQLite cache (C2)** only if query speed bites at
   scale; consider **Dolt (C3)** only if the project ever reconceives tasks as
   *data* rather than *documents*. The **event log (B1)** is the move if a real
   audit trail becomes a goal.

## Open questions

- **Browsability bar:** is `ls tasks/<status>/` a hard requirement, or is a
  generated `BOARD.md` + `grep` good enough? (Decides A3a vs A3b.)
- **Stable IDs:** identity is the *slug/filename* today (no UUID, no `rename`
  verb). If you group by epic, an epic change becomes a rename — do you want an
  **immutable `id`** separate from the human slug so moves/renames never break
  identity or cross-links?
- **Plain-git is sacred?** Is "just markdown in a normal git repo" a defining
  constraint (rules out Dolt), or a current convenience?
- **History:** is an audit trail (who/when/what) a real goal (→ event log), or is
  current-state-only fine (git history already covers most of it)?
- **Scale:** realistic ceiling on task count? (Decides whether the SQLite cache is
  ever worth it.)

## Related

- **Home: epic
  [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]** — this
  spike feeds its decisions (it's design-first; the open questions above are its).
- [[2026-06-24-remote-planning-repos-backends-and-sync]] — remote backends + the
  version-aware `Store` port (OCC). Same root cause: state encoded in the path.
- Epic [[23-point-an-impl-repo-at-an-external-planning-repo]].
- Today's invariant: `status == directory` (CLAUDE.md; `store/layout.go`,
  `Misfiled()`), which option A3 would relax (frontmatter becomes source of truth).
