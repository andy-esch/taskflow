---
schema: 1
status: active
description: 'Rework the on-disk data model: a stable-key layout with status in frontmatter not the directory, a shared core read-model/projection, and content-aware OCC. Design-first epic; decisions made (ADRs 0003/0004); projection-shape is the last open piece.'
priority: medium
tags: [storage, architecture, core]
created: "2026-06-24"
---
# Data-model evolution — stable-key storage, shared read-model, content OCC

**Status: design-first. NOT ready to implement.** This epic collects a cluster of
intertwined decisions surfaced by two spikes; most of the load-bearing choices are
still **open** (see *Open questions*). The first task here is to *close those
decisions*, not to write storage code.

## Why this exists (the convergence)

Three threads turned out to share one root cause — *mutable state encoded in the
file path* — and one shared solution shape:

- **Merge duplicates.** `status == directory` makes a status change a *rename*; a
  branch that moves a task and a branch that edits its body merge into two files
  in two dirs (different content, neither flagged misfiled). Observed in practice.
- **Remote planning** (epic 23 phase 2) needs atomic cross-dir moves that object
  stores lack, plus a content-aware write guard.
- **Web companion** (epic 19) needs a read API — which is the same projection a
  generated board would render.

The shared answer being pressure-tested: organize the tree by a **stable key**
(status moves into frontmatter), expose state through **one `core`
read-model/projection** (board → TUI / CLI / web), and guard writes with
**content-aware optimistic concurrency**.

Spikes:
- `planning/research/2026-06-24-task-storage-model-files-logs-or-versioned-db.md`
- `planning/research/2026-06-24-remote-planning-repos-backends-and-sync.md`

## Direction so far (leans, NOT final)

- Stay **file-based + git-native** — not a DB, not Dolt, not an event log (for now).
- **Stable-key layout**; `status:` in frontmatter as source of truth (it is already
  dual-stored today — the directory is merely the authoritative copy).
- A **generated board** as the materialized read-model, defined once in `core` as
  **structured data**, rendered by every adapter (markdown / JSON / HTML / TUI).
- A **version-aware `Store` port** (read-version + `ifVersion`-on-write →
  `ErrConflict`) as the write-side foundation.

Starting positions to attack, not commitments.

## Decision 2026-06-30 — stable key committed (principle)

The **stable-key direction is committed**: organize entities by a stable key with
`status` in frontmatter, retiring `status == directory` as the *authority*. The
rationale that settled it — a stable key is the one choice that survives any later
backend (it is exactly the primary key a database would use), so it's safe to
commit independent of where storage ultimately lands.

Scope of the commitment is the *principle only* — identity is decoupled from
status. The key's **form** (a 12-char time-sortable id), the **grouping** (flat),
and the **filename + reference scheme** are now all decided in the sections below.
The merge pain a stable key does *not* fix — concurrent edits to the same file's
mutable fields across branches — is analyzed in the 2026-06-30 updates of both
spikes (content-vs-workflow-state; the fork between not-branching-planning and
not-making-git-canonical). Those remain undecided.

## Decision 2026-06-30 — identity is a 12-char time-sortable id

Identity form is decided: each entity carries a **12-char time-sortable id** as its
canonical identity — lowercase Crockford base32 (`0-9 a-z` minus look-alikes
`i l o u`), ~44 bits millisecond-timestamp + ~16 bits randomness, big-endian so a
lexical sort is chronological. ULID-shaped but short; **not** a truncated ULID (whose
prefix is all-timestamp, low-entropy) — a purpose-built 60-bit sortable id. It lives
**in** the filename next to the human slug (extending the epic `NN-<slug>` hybrid,
with `NN` replaced by the id). **References resolve on the id, not the slug**, so
re-titling breaks nothing. The exact filename ordering, link form, and reference
shape are in the *filename + reference scheme* section below (Scheme 2).

Why a random/time id over a monotonic integer: it needs **no allocation
coordination** — any branch, machine, or offline agent mints one with no central
counter — so identity is **robust to the still-open branching decision** (trunk-only
vs multi-branch vs server) rather than betting on central allocation the way an
integer would (cf. today's accepted `nextEpicNumber` race). Time-sortable (over a
purely random id) buys **self-dating ids** (creation time survives even if `created:`
is lost), **chronological `ls`**, and clean B-tree locality if the id ever becomes a
DB key. **Length 12** (≈60 bits) is collision-safe well past planning scale (cf.
PlanetScale's 12-char public-API ids); creation **collision-checks** the id and
regenerates on the astronomically-rare intra-millisecond clash (or generates
monotonically within a tick).

Follow-ons this implies (not yet scheduled): an `id` frontmatter field + a backfill
migration for existing tasks (epics/audits already carry a stable prefix); switching
stored `epic:` references to resolve on the id (carried as the readable stem, see
Scheme 2); converting body cross-links from `[[…]]` to standard relative-path
markdown; and a `rename` path that cascades inbound links.

## Decision 2026-06-30 — filename + reference scheme ("Scheme 2")

How identity shows up on disk and in links is decided — the weighing is in the
2026-06-30 link/lookup thread (storage spike + this epic). Chosen **Scheme 2**:
readable tree + GitHub-clickable links, at the cost of a tool-run link cascade on
the (rare) rename.

- **Filename: id-led `<id>-<slug>.md`.** The id lives in the filename so `id → file`
  resolves from a directory scan, never per-file reads (O(N) `readdir`, not O(N)
  parses), and a clean **prefix-glob** (`<id>*`) finds it. **Id-first, not
  slug-first**, so the filename sorts **chronologically** (the id is time-sortable) —
  letting a limited "N most recent" query sort cheap readdir names and read only the
  page, instead of read-all-then-sort. The cost: the slug sits *behind* 12 chars of id
  in `ls` / diffs / raw paths. That cost is bounded — **rendered links and body-link
  display text are unaffected**, so every human-*facing* surface stays readable, and
  chronological `ls` is itself a useful browse. (The pure-perf win is modest in
  practice — most list views read frontmatter for display anyway, and `serve` indexes
  in memory — so id-first is chosen as much for self-dating ids + chronological order
  as for speed.)
- **Frontmatter refs (`epic:`): the stem `<id>-<slug>`,** resolved on the **id
  prefix** (the slug travels for readability but isn't matched). If the slug drifts
  after a rename, resolution still works and `lint` refreshes the cosmetic suffix —
  the same shape as today's `epic: 20-cli-ux-…` resolving on `20`, id instead of `NN`.
- **Body cross-links: standard relative-path markdown** —
  `[display text](…/<id>-<slug>.md)`, **not `[[wikilinks]]`** (GitHub renders the
  former, not the latter). The display text carries the human meaning; the path is
  plumbing. Because the link embeds the real path, a rename breaks it — so the
  `rename` path **cascades inbound links** and `lint` flags danglers. (`[[…]]`, used
  in today's body template and research docs, is dropped for GitHub-honored links.)

**Why this and not the alternatives.** Lock-in is set by the *reference scheme*, not
the filename: Scheme 2 keeps the planning **navigable on GitHub without the tool**
(the project's git-native value) and only needs the tool to *rename safely* — the
same read-freely / mutate-with-the-tool posture the project already takes. The
pure-bare-id alternative (tool-only navigation) and the opaque-`<id>.md` alternative
(unreadable tree) both trade away more than the rename cascade costs, and file renames
are rare.

**Ordering note (supersedes an interim slug-led draft).** An earlier draft put the
slug first (`<slug>-<id>.md`) for raw-path readability; reversed to **id-first** so
the filename is chronologically sortable. Id-first only pays off *with* a time-sortable
id (else the order is arbitrary noise) — the two decisions are linked, which is why
they landed together.

## Open questions (the undecided core)

**Layout & identity**
- [x] Grouping — **decided 2026-06-30: flat** `tasks/<id>-<slug>.md` (no epic
      subdirs). Epic membership is mutable, so dir-per-epic would make an epic
      reassignment a *file move*; flat keeps it a one-field frontmatter edit and
      yields the simplest, most stable body link paths. (Status already left the path
      via stable-key; now epic does too — the tree is one flat `tasks/` dir, all
      grouping is frontmatter.)
- [x] **Immutable `id`** vs **slug-as-identity** — **decided 2026-06-30: a 12-char
      time-sortable id** (references resolve on the id; the slug is a cosmetic suffix
      — see the identity decision above). Removes the rename-is-an-identity-change
      problem outright; a random/time id (not a monotonic int) so identity doesn't
      presuppose the still-open branching model.
- [x] Fully **retire `status == directory`** as the *authority* (frontmatter is
      truth) — **decided 2026-06-30**. Open sub-detail: whether to still render a
      *derived* by-status view. This relaxes a stated non-negotiable invariant
      (the CLAUDE.md / ARCHITECTURE.md update lands with the implementation, not
      now).
- [x] **Filename + reference scheme** — **decided 2026-06-30 (Scheme 2):** id-led
      `<id>-<slug>.md` (filename sorts chronologically); frontmatter `epic:` carries
      the stem and resolves on the id prefix; body cross-links are standard
      relative-path markdown (GitHub-clickable), cascaded by `rename`. See the
      decision above.
- [x] **Audits** (bucket == directory) and **epics** — **decided 2026-06-30: same
      treatment.** All three entities get the 12-char id + id-led `<id>-<slug>.md`
      + Scheme 2 refs. For audits this also retires **bucket == directory** (bucket
      moves to frontmatter, mirroring status); epics are already flat + prefixed, so
      it's mostly swapping `NN` for the id. (The `Misfiled` replacement question below
      now covers audit bucket-drift too.)
- [x] What replaces **`Misfiled()`** drift-detection once frontmatter is truth? —
      **decided 2026-07-01.** The dir-vs-frontmatter cross-check is retired (no folder to
      disagree with); `lint` instead **validates the frontmatter status/bucket itself**
      (present + recognized), surfaced as a ⚠ in `list` / `show`. This **inherits the
      fail-open + flag model epics already use** (`IsKnownEpicStatus`): an unknown/missing
      status **still lists, flagged — never dropped, never a hard error**. `lint --fix`
      loses status-realignment (no second source to copy from), so a bad/missing status is
      **detect-and-report** — the human fixes it via `task set` / a lifecycle verb, no
      auto-guessing. Audit `bucket` gets identical treatment.

**Board / projection**
- [x] Committed `BOARD.md` vs git-ignored — **decided 2026-07-01: neither yet.** Build
      the **read-model/projection in `core`** + a **`tskflwctl board`** command that
      renders it **on demand** (always fresh, zero freshness machinery — same cost as
      `status` / `task list`). The **committed, GitHub-browsable `BOARD.md` is the
      nice-to-have, deferred to the serve era** — serve regenerates + commits it as part
      of its normal single-writer batched write, so freshness becomes free.
- [x] Freshness mechanism — **decided: none needed now** (on-demand can't go stale);
      **serve owns it later**. A **regen gate / auto-commit GHA is rejected** — a GHA that
      commits `BOARD.md` back to `main` is a *second automated writer* that fights
      ADR-0004's single-writer model + branch protection (plus loop/permission gotchas),
      and it is machinery serve would obsolete.
- [ ] **Still open — the projection's exact shape** (sections, grouping, sort, fields)
      and where it lives in `core` (generalize `Summary()`). The real remaining design
      work and the load-bearing part.

**Concurrency (OCC)**
- [x] Version token — **decided 2026-07-01: a whole-file content hash** (SHA-256).
      Backend-agnostic (same mechanism local / git-cache / object-store / server) and
      **correct for a "tool writes files, you commit later" model** — a git blob SHA
      tracks the *committed* state and misses uncommitted working-tree edits, whereas a
      content hash fingerprints the actual bytes on disk. ~Zero marginal cost (the
      read-modify-write already reads the file). `mtime+size` rejected: unreliable across
      machines/containers (clock skew, mtime not preserved) — a real risk given the cron
      agents. Mechanism: reads return the hash as `version`; writes take `ifVersion` →
      `ErrConflict` (exit 14) on mismatch, generalizing today's path-CAS.
- [x] Conflict UX — **decided 2026-07-01: plain retry** (surface `ErrConflict`; the
      caller re-reads, re-applies, retries). Ideal for the **cron AI-agent writers**,
      which re-run deterministically — no human merge UI needed. Refinement: the tool may
      do a small **internal bounded auto-retry** for the *scriptable* field-level
      mutations (`set` / `append` / `move`) so agents don't each reimplement it; the
      *human* whole-file `edit` path still surfaces the conflict (they edited a specific
      version and should see it). "Assist a merge" deferred — revisit only if humans hit
      conflicts often.
- **Note (standing, not an open decision):** OCC is a *runtime* guard — it does **not**
      fix the git-merge duplicates (born offline, when the tool isn't running); the
      *layout* change (ADR-0003) fixes those. Keep the two separate.

**Migration**
- [x] How do existing trees migrate (this repo + desirelines)? — **decided (ADR-0003
      §6): a one-time throwaway script**, not a `tskflwctl migrate` command (internal
      tool, small known set of repos).
- [x] Coexist vs hard cutover? — **hard cutover** (the tool reads only the new layout;
      no permanent dual-read path). Safety via git: run on a copy, verify, commit.
- [x] One big rename-churn commit — **accepted** (`git` is the undo). Residual impl
      detail: the exact field transform + slug derivation.

## Risks

- **Blast radius.** Relaxing `status == directory` ripples through the store,
  `layout.go`, `WatchPaths`, lint, completion, `schema`/agent guidance, CLAUDE.md +
  ARCHITECTURE.md + README, and essentially every test.
- **Data safety.** The on-disk format change is near-one-way and touches the tool's
  own dogfood data *and* desirelines — a bad migration corrupts planning data.
- **Agent contract.** Some agent guidance / muscle memory relies on `ls
  tasks/in-progress/`; a board doesn't help a script that globs directories. Audit
  what assumes today's layout before changing it.
- **Board churn vs browsability.** A committed board changes constantly (noise); a
  git-ignored one abandons the GitHub-browsable goal. No clean answer yet.
- **Conflation.** Tempting to assume OCC fixes the merge duplicates — it doesn't.
  Ship OCC alone and the duplicates persist.
- **Scope creep.** Tasks + audits + epics + remote + web all touch this; without a
  tight MVP it sprawls.
- **Reversibility.** The id-vs-slug choice is expensive to undo after cross-refs
  adopt it.

## Tentative phasing (a sequence, not a commitment)

1. **Decision consolidation** — resolve the open questions into a chosen design.
   This is the real first task; gate it before any storage code. *Drafted as
   [[0003-stable-key-id-addressed-storage]] (proposed) — the identity/layout half;
   board / OCC / migration still open below.*
2. **Read-model / projection + board** — independent of the storage change; ships a
   `board` command and proves the web read-model.
3. **Stable-key layout + version-aware OCC** — together (layout shapes OCC); needs a
   migration path.
4. **Payoffs (separate epics):** remote backends (epic 23 ph2), `serve` read
   endpoint (epic 19) — both ride on 2–3.

## Out of scope

- The remote backends themselves (epic 23 phase 2) and the web `serve` app (epic
  19) — this epic provides their *foundation* (projection + OCC + stable paths),
  not the surfaces.
- Event-log / CRDT / Dolt representations — considered in the spike, parked unless
  history or scale forces a revisit.
