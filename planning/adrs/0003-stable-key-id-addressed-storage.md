---
status: accepted
date: "2026-06-30"
deciders: [andy-esch]
tags: [adr, storage, data-model, identity]
supersedes: []
superseded_by: null
---

# ADR-0003: Stable-key, id-addressed, flat file storage

> ✔ **Accepted 2026-07-01 — finalized.** The decision sections below are frozen; add
> new information only under `## Amendments`. Reverse via a superseding ADR.

> Follows the ADR format established in [[0001-adopt-adrs]]. This is the
> **decision-consolidation** step for epic
> [[24-data-model-evolution-stable-key-storage-read-model-content-occ]] — it ratifies
> the *identity + layout* cluster the epic and its two spikes converged on over
> 2026-06-30, and explicitly bounds what it does **not** decide (branching/SSOT,
> `serve`, OCC token, board, migration — see *Out of scope*). Full rationale lives in
> [[2026-06-24-task-storage-model-files-logs-or-versioned-db]] and
> [[2026-06-24-remote-planning-repos-backends-and-sync]]; this ADR is the citable
> conclusion.

## Context and Problem Statement

Today every entity's **mutable workflow state is encoded in its path**: a task's
`status` *is* its directory (`tasks/<status>/<slug>.md`), an audit's bucket likewise.
That single fact causes three distinct pains, surfaced by two spikes:

- **Merge duplicates.** A status change is a *rename*; a branch that moves a task and
  a branch that edits its body merge into two files in two dirs (different content,
  neither flagged misfiled). Observed in practice.
- **Remote / atomic writes.** Object stores and remote backends lack the atomic
  cross-dir move the status-rename needs, and there is no content-aware write guard.
- **No stable address.** Identity is the slug/filename, which a re-title changes — so
  references, links, and (future) web URLs are brittle, and there is no `rename` verb.

Root cause: **mutable state (status/bucket) and identity (slug) both live in the
path.** The fix is to make the path carry only *stable* things. A key enabler:
`status:` is already dual-stored in frontmatter today (the directory is merely the
*authoritative* copy), so moving the source of truth into frontmatter is a smaller
move than it sounds.

This ADR decides the **on-disk storage model** — paradigm, organizing key, identity,
filename, and reference form. It does **not** decide how planning is *synced* across
machines or how a web app *writes* (the branching/SSOT questions, deferred below).

## Considered Options

**Storage paradigm.**

- **A — Plain files (git-native), chosen.** Zero new deps, diffable, PR-reviewable;
  the `core.Store` port already abstracts the backend.
- **B — SQLite-canonical / Dolt / event-log.** More query power or merge granularity,
  but abandons "plain markdown anyone can read/PR" (Dolt) or hand-browsability (event
  log), or adds a cache-invalidation surface (SQLite). Parked unless scale/history
  forces a revisit (the spike's B1/C2/C3).

**Organizing key.**

- **A1 — status-as-directory (today), rejected.** `ls`-browsable, but every status
  change is a rename → the merge-duplicate class and non-atomic moves.
- **A3 — flat, stable key + status in frontmatter, chosen.** Kills the rename class,
  makes OCC a one-file check, keeps a flat browsable tree.

**Identity form.**

- **slug-as-identity (today):** readable, but a re-title *is* an identity change.
- **monotonic integer:** readable + ordered, but needs a central allocator (the
  `nextEpicNumber` race) → presupposes the unresolved branching model.
- **full ULID (26 char):** stable + sortable but long/noisy.
- **12-char time-sortable id, chosen:** coordination-free, sortable, collision-safe at
  planning scale, less than half ULID's length.

**Filename + reference scheme.**

- **Scheme 1 (bare-id refs):** stable, but tool-only navigation (high lock-in).
- **Scheme 3 (opaque `<id>.md`):** stable + GitHub-clickable, but an unreadable tree.
- **Scheme 2 (id-led readable filename + standard-markdown links), chosen:** readable
  tree + GitHub-clickable links; cost is a tool-run link cascade on the (rare) rename.

## Decision

Adopt a **stable-key, id-addressed, flat, file-based** storage model for all three
entities (tasks, epics, audits).

### 1. Stay file-based + git-native

Plain markdown + frontmatter in a git repo remains canonical. Not a DB, not Dolt, not
an event log (parked, not rejected forever — see the spike).

### 2. Stable key — status/bucket move to frontmatter

The directory stops being the authority for mutable state. `status` (tasks) and
`bucket` (audits) become frontmatter source-of-truth; the tree is organized only by
*stable* things. The `status == directory` / `bucket == directory` invariant — a
stated non-negotiable in CLAUDE.md / ARCHITECTURE.md — is **retired** (those docs
update *with* the implementation, not before).

### 3. Identity — a 12-char time-sortable id

Each entity carries a stable **`id`** (new frontmatter field): **12 chars, lowercase
Crockford base32** (`0-9 a-z` minus look-alikes `i l o u`), **~44 bits ms-timestamp +
~16 bits randomness, big-endian** so a lexical sort is chronological. ULID-shaped but
short; **not** a truncated ULID (whose prefix is all-timestamp) — a purpose-built
60-bit sortable id. ~60 bits is collision-safe well past planning scale (cf.
PlanetScale's 12-char public-API ids). **Creation collision-checks** the id and
regenerates on the astronomically-rare intra-millisecond clash (or generates
monotonically within a tick). Chosen over a monotonic integer because it needs **no
central allocator**, so identity does not presuppose the still-open branching model.

### 4. Filename — id-led, flat

`<id>-<slug>.md`, in one flat directory per entity (`tasks/`, `epics/`, `audits/`); no
status, bucket, or epic subdirectories. **Id-led** (not slug-led) so the filename
sorts chronologically and a known id resolves by prefix-glob. The slug is a cosmetic,
renamable suffix. Flat because epic membership is mutable — dir-per-epic would make a
reassignment a *file move*; flat keeps it a one-field frontmatter edit.

### 5. References — Scheme 2

- **Frontmatter refs** (`epic:`): the **stem `<id>-<slug>`**, resolved on the **id
  prefix** (the slug travels for readability but isn't matched; `lint` refreshes a
  drifted slug). Same shape as today's `epic: 20-cli-ux-…` resolving on `20`.
- **Body cross-links:** **standard relative-path markdown** `[text](…/<id>-<slug>.md)`
  — GitHub-clickable — **not `[[wikilinks]]`** (GitHub renders the former, not the
  latter). A rename breaks the embedded path, so the `rename` verb **cascades inbound
  links** and `lint` flags danglers.

Lock-in is set by the *reference scheme*, not the filename: Scheme 2 keeps the planning
**navigable on GitHub without the tool**, needing the tool only to *rename safely* —
the same read-freely / mutate-with-the-tool posture the project already takes.

### 6. Migration — a one-time throwaway script (not a CLI command)

The format change requires migrating existing planning trees (this repo + desirelines).
Because this is an **internal** tool with a small, known set of repos, the migration is
a **one-time throwaway script**, *not* a permanent `tskflwctl migrate` command. It
assigns each file an `id`, renames it to `<id>-<slug>.md`, moves `status` / `bucket`
into frontmatter, and rewrites inbound references. Consequences of "throwaway":

- **Hard cutover, no coexistence.** The tool reads only the new layout — no permanent
  dual-layout read path to carry and later remove.
- **Safety via git, not built-in reversibility.** Run on a copy, verify, then commit the
  result as one churn commit; `git` is the undo. No reverse-migration code.
- **Run once per repo, then discarded** — it need not be general, configurable, or
  supported.

The exact field-by-field transform and slug derivation for existing files are
implementation detail, not fixed here.

## Data model & layout

| | Tasks | Epics | Audits |
| :-- | :-- | :-- | :-- |
| Location | `planning/tasks/<id>-<slug>.md` (flat) | `planning/epics/<id>-<slug>.md` (flat) | `planning/audits/<id>-<slug>.md` (flat) |
| Identity | `id` (12-char) | `id` (replaces `NN`) | `id` (replaces date prefix) |
| Mutable state | `status` (frontmatter) | `status` (already frontmatter) | `bucket` (frontmatter — was directory) |
| Refs in | `epic:` stem, body md-links | — | — |

All three get the same treatment; for audits this also **retires `bucket ==
directory`**. Epics are already flat + frontmatter-status, so for them it is mostly
swapping the `NN` prefix for the `id`.

## Out of scope (deferred — NOT decided here)

This ADR decides the *on-disk model only*. Related decisions live elsewhere.

**Decided since (elsewhere):**

- **Branching / single-source-of-truth model** and the **web-write landing model** —
  decided in [[0004-single-source-of-truth-serve-owns-git]] (serve-owns-git).
- **OCC version token + conflict UX** — decided in epic 24 (whole-file content hash +
  plain retry).

**Still open:**

- **The board / read-model** — committed vs git-ignored, freshness, shape — epic 24.
- **`Misfiled()` replacement** — what guards a hand-edited bad status/bucket once
  frontmatter is truth (now also covers audit bucket-drift) — epic 24.

**Deferred surface (a build, not a design call):**

- The **`serve` daemon + web app** — epic 19.

## Consequences

**Positive.**

- Kills the rename / merge-duplicate class; makes content-OCC a one-file check.
- Stable identity that survives re-titles, moves, and a future backend swap (it *is*
  the DB primary key); permanent web URLs.
- Coordination-free id generation — identical trunk-only, multi-branch, or
  server-backed, so it does not block the still-open branching decision.
- A readable, GitHub-navigable tree with clickable links and chronological `ls`.

**Negative / cost.**

- **Blast radius.** Retiring dir-as-authority ripples through the store, `layout.go`,
  `WatchPaths`, lint / `Misfiled`, completion, `schema` / agent guidance, CLAUDE.md +
  ARCHITECTURE.md + README, and essentially every test.
- **Near-one-way migration** over live data (this repo *and* desirelines) — the main
  data-safety risk. Handled by the throwaway script (§6): run on a copy, verify, commit;
  `git` is the undo, and the cutover is one accepted rename-churn commit.
- **Contract change.** A new `id` field + the `<id>-<slug>` filename change the
  `--json` envelopes and golden snapshots → a `schema_version` bump + docs/cli regen.
- **Id noise.** A 12-char id leads every raw filename / path / diff; mitigated (slug
  present, rendered links + link text unaffected, chronological `ls` is itself useful)
  but real. Lever if it bites: a shorter id (re-opens the id *form*).
- **Rename cascade.** Body links embed real paths, so `rename` must rewrite inbound
  links; external links (commit messages, bookmarks) still break — but that is true of
  moving any file in git.

## Amendments

<!-- Append-only, dated entries added AFTER this ADR is accepted. Format:
     ### 2026-07-01 — <what changed and why> -->

### 2026-07-02 — id split finalized at 43/17; migration backfill-timestamp policy

- **Bit split:** the "~44 bits time + ~16 bits random" in §3 is finalized as **43
  time + 17 random** — sortable through ~year 2248, with 2× the same-ms
  cross-process collision headroom. Length (12) and everything else in §3 unchanged.
- **Migration backfill timestamp** (the §6 "exact transform" detail, now decided):
  a backfilled id embeds the file's **`created:`** date; if absent, another
  frontmatter date; if neither exists, the migration **errors** so the operator adds
  the date key. Sub-day ordering uses a **random** low tail (no sequence state to
  track), with **dedupe-and-regenerate** on the rare same-day collision. Landed as
  the stateless `id.NewAt(unixMilli)`.

### 2026-07-04 — epics keep their `NN` prefix as their stable key (revises §3/§4/§5 for epics)

Grounding the flatten against the real `desirelines-planning` repo revised the "**all three**
entities get a 12-char id" decision **for epics only**. Epics carry no id today (0/17), are
`NN-<slug>.md`, and are referenced by ~644 tasks as `epic: NN-<slug>`. Since the `NN` prefix
is *already* a stable, human-meaningful key (epics aren't renamed to a different number) and
`epic:` refs resolve on it cleanly, minting 12-char ids for epics buys little and would force
rewriting ~644 refs. **Decision: the epic `NN` prefix IS the epic's stable key — epics keep
`NN-<slug>.md` and refs stay `epic: NN-<slug>`.** Tasks and audits are unchanged (12-char
id-led per §3/§4). So the flatten (§4) and migration (§6) apply to **tasks + audits only**;
epics are already flat and stable. The §-153 data-model table's "epics → `id` (replaces
`NN`)" row is superseded by this entry.

### 2026-07-04 — carveout contract: what the flat scan does with non-entity files

Grounding the flatten surfaced that a flat scanned bucket is a *pollution surface* —
`markdownCandidates` builds resolution candidates from filenames only, so any stray `.md`
(e.g. `audits/HOWTO-execute.md`) becomes a fuzzy-match candidate and a listing problem. This
amends §4 (id-led flat filenames) with the rule for **non-entity files inside a scanned
bucket**. Decided in
[[curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate]]; built as part
of [[flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory]].

- **Entity = a filename-shape test.** A file directly in a scanned bucket is an entity iff its
  name leads with a valid stable key — `<id>-` (tasks/audits, the 12-char id, `id.Valid()`) or
  `NN-` (epics). A *positive* signal, so a real entity that lost its frontmatter still fails
  loud (a broken entity), while a non-id-led file is simply not one.
- **Strays error, reusing the existing channel.** A non-entity `.md` in a bucket is a
  `FileProblem` ("not an entity — move to `meta/` or delete") on the same resilient-read channel
  that already carries malformed files (surfaced in list/lint/board, non-fatal, excluded from
  entity data and `--json`). No new severity tier and **no `schema_version` bump** for carveouts.
- **Resolution gates on the same shape.** `markdownCandidates` excludes non-id-led names from
  candidacy, so a stray can never win a fuzzy match or shadow a real entity, and `show <stray>`
  is a clean `ErrNotFound`. (Erroring on the listing side alone does not fix resolution
  pollution — this gate is the actual fix.)
- **`meta/` — the sanctioned home.** A single top-level `planning/meta/` (never scanned; the
  tool scans only `tasks/`/`epics/`/`audits/`) holds non-entity material, free internal
  structure (`meta/routines/`, `meta/HOWTO-execute.md`). Standardized — a hardcoded name, no
  config `ignore`/dotfile — to start. Ignore-only for now; a subtree like `meta/routines/` may
  be promoted to a scanned entity type later with zero file moves.
- **One carve.** A bucket-root `README.md` is silently ignored (GitHub renders it as the
  folder's landing page).

The *full* "what is a valid entity / valid frontmatter" contract is
[[26-frontmatter-schema-declared-validation-contract]]'s to formalize; this amendment fixes the
narrow layout-hygiene rule the flatten needs now.

## Related

- Home epic & the open-questions index:
  [[24-data-model-evolution-stable-key-storage-read-model-content-occ]].
- Storage-model rationale (paradigm, organizing key, the residue):
  [[2026-06-24-task-storage-model-files-logs-or-versioned-db]].
- Sync / concurrency & the branching/SSOT fork (deferred here):
  [[2026-06-24-remote-planning-repos-backends-and-sync]].
- ADR format this follows: [[0001-adopt-adrs]].
