---
schema: 1
id: 6fjvr03mr9zg
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'How should the tool handle non-entity files in its dirs (HOWTO, routines, task-review companions)? Convention vs config vs association. Design-first: propose options; Phase-B blocker.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, storage, config]
created: "2026-07-04"
updated_at: "2026-07-05"
completed_at: "2026-07-04"
---

# Curation carveouts: how should tskflwctl handle non-entity files in its dirs?

> **DESIGN-FIRST BRIEF (2026-07-04) — propose options, don't implement yet.** A fresh agent
> should read this end-to-end and return **design proposals**: how tskflwctl should tolerate/
> curate non-entity files in (and near) its scanned dirs — **standardized** (a convention)
> vs **configurable** (`.tskflwctl.toml`) vs a **richer association model** — with tradeoffs
> and a recommendation. The parenthetical the original title carried ("frontmatter-gate +
> .tskflwctlignore") is the *initial candidate only*, not a decision. The maintainer
> explicitly wants the standardized-vs-configurable question weighed here.

## The problem

Real planning repos keep **non-entity files in and around the tool's dirs** — an audit
cheat-sheet (`HOWTO-execute.md`), routine specs, READMEs, templates, and (soon) per-task
review companions. Today tskflwctl is a **purist** about `tasks/`/`epics/`/`audits/`; it has
no notion of "this file isn't mine, leave it alone." That bites once the layout flattens.

## Why it's forced now (Phase-B hard requirement)

Audits scan the **bucket subdirs** (`audits/open|closed|deferred/`), so `audits/HOWTO-
execute.md` at the audits *root* is invisible today. **Phase B flattens audits to one scanned
`audits/`** (`markdownCandidates(auditsDir, "")`), at which point that HOWTO becomes a
resolution candidate (`id=HOWTO-execute`) and a listing `FileProblem`. So *some* form of this
is a **flatten prerequisite** — at minimum the flat scan must skip non-entity files.

## Grounding (code facts — verify, don't trust)

- `markdownCandidates` (`store/resolve.go`) builds resolution candidates from **filenames
  only** (no frontmatter read) → any stray `.md` becomes a candidate and pollutes `show`.
- `scanDir` + `parseTask/Audit/Epic`: a no-frontmatter `.md` → `FileProblem` (listing noise),
  NOT a silent skip. A frontmatter'd-but-malformed file → also a `FileProblem` (a real error).
- **No ignore mechanism exists** (no `.tskflwctlignore`, no `ignore` config key) — confirmed.
- Config already supports `taskflow_root` (the [[isolate-desirelines-planning-entities-under-a-dedicated-planning-directory]] approach puts entities in a dedicated dir — the "out of scope entirely" answer for whole *directories*).

## The crux — these files are NOT all the same (a use-case spectrum)

The design must decide a mechanism **per category** — and whether one mechanism spans several:

1. **Pure utility docs** — `HOWTO-execute.md`, READMEs, templates. Want: **IGNORED** (invisible
   to `list`/`show`/`lint`), coexisting freely in an entity dir.
2. **Recurring routines** — `routines/*.md` specs (Claude Code fires them server-side; the spec
   produces an audit). Want: **maybe** ignored, **maybe** a first-class entity
   ([[spike-routines-as-a-first-class-entity-routine-audit-lineage]]), **maybe** a companion.
3. **Per-entity companion files** — epic 27's **Shape B**: a task's code-review findings stored
   as a **separate linked file** (`target_task: <slug>`); or a routine's execute-doc. Want:
   **ASSOCIATED with an entity** — tracked-and-linked, not ignored, not a top-level entity.
4. **Adjacent repo content** — `research/`, `archive/`, `incidents/` — handled by the
   **isolation** approach (out of the tool's scanned dirs entirely; `taskflow_root`).

## The open question (propose on this)

Weigh these candidate mechanisms (mix-and-match is fair):

- **A. Standardized convention — frontmatter-gate.** No-frontmatter `.md` = *not an entity* =
  silently ignored (list + resolve); frontmatter'd-but-malformed = a *broken entity*
  (`FileProblem`). Zero config; handles **category 1** cleanly. Can't express "this
  frontmatter'd file isn't an entity" or "ignore this subtree."
- **B. Configurable ignore — `.tskflwctlignore` (gitignore-style) or a config key
  `ignore = [...]`.** Explicit globs; handles frontmatter'd non-entities + subdirs. More
  surface; user-maintained; which file (`.tskflwctl.toml` vs a dotfile) is itself a call.
- **C. Companion / association convention.** A naming or frontmatter link that *attaches* a
  file to an entity (**category 3**) — e.g. `<id>.review.md`, or a `target_task:` pointer —
  tracked as a companion, not a top-level entity. **Intersects epic 27 directly** (its A/B
  storage fork).
- **D. Entity-type extensibility.** Promote some "non-entity" files (routines, reviews) to
  first-class entity *types* via the ADR-0003 entity-registry/`Descriptor` path — then they're
  not carveouts at all. **Where is the line between "ignore it" and "make it an entity"?**

## What the proposal should deliver

- A recommended mechanism (or combination) covering the spectrum, with tradeoffs.
- The **standardized-vs-configurable** call, with rationale (and *which* config surface).
- Where the line sits between **ignore / companion / first-class entity**.
- The **Phase-B minimum** (flat scan skips non-frontmatter files — the hard, ship-now part)
  vs the fuller design (can land incrementally).
- How it composes with the isolation approach (dirs) — carveouts are for *files inside* the
  scanned dirs; isolation is for whole *dirs* out of scope.

## Cross-references

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- **Hard prerequisite for** [[flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory]] (the flat-scan pollution).
- Use case: [[spike-routines-as-a-first-class-entity-routine-audit-lineage]] (category 2 — ignore vs first-class).
- **Coordinate with epic [[27-agent-code-review-on-tasks-structured-review-loop]]** — its Shape A
  (in-file `## Review` section) vs Shape B (a separate linked review file) *is* the category-3
  question; decide the carveout/companion model and 27's storage shape together.
- Complementary approach: [[isolate-desirelines-planning-entities-under-a-dedicated-planning-directory]] (category 4 — whole dirs out of scope).
- Epic [[26-frontmatter-schema-declared-validation-contract]] — a declared schema is where
  "what makes a file an entity" would be formalized; the frontmatter-gate leans on it.

## Decision (2026-07-04) — carveout model resolved (design-first outcome)

Resolved with the maintainer over a design session. Mechanism: a **filename-shape
classifier + one blessed `meta/` folder**, reusing the existing `FileProblem`
machinery rather than adding a warning tier.

### Model

For any `.md` directly inside a scanned entity dir (`tasks/`, `epics/`, `audits/`),
classify by **filename shape** (uniform once the flatten mints ids on every task/audit):

- id-/`NN`-led name **+ valid frontmatter** -> **entity** (listed, resolved).
- id-/`NN`-led name **+ missing/malformed frontmatter** -> **broken entity** -> hard
  `FileProblem` (today's behavior).
- any **other** name (not id-led) -> **stray non-entity** -> hard `FileProblem`
  ("not an entity - move to `meta/` or delete"), **except** a whitelisted `README.md`
  (GitHub renders it as the folder landing page), which is ignored silently.

Strays surface as errors through the **existing** problem channel (rendered in
`task list`, `audit list`, `lint`, and the TUI overview) and never pollute entity data
or `--json`. The scan stays resilient - a stray is collected, not fatal.

### Decisions

1. **Classify by filename, not frontmatter presence.** The id/`NN` prefix is the positive
   "meant to be an entity" signal - keeps fail-loud for a real entity that lost its
   frontmatter, while `HOWTO-execute.md` (even with stray frontmatter) is never mistaken
   for one. Rides the flatten's id-parsing predicate (`id.Valid()`), so it is nearly free.
2. **Strays error (reuse `FileProblem`), don't warn.** No new severity tier, no new warning
   type, **no `schema_version` bump** - the maintainer is fine erroring on strays. Drops the
   epic-26 Q6 severity coupling from the near-term path.
3. **Resolution must gate on filename shape too.** `markdownCandidates` is filename-only, so
   a stray would otherwise become a fuzzy-match candidate (`audit show howto` shadowing a
   real match). Exclude non-id-led names from candidacy -> `show <stray>` is a clean
   `ErrNotFound`. This is the hard Phase-B prerequisite: erroring on the listing side alone
   does NOT fix resolution pollution.
4. **One blessed folder: `meta/`.** Single, top-level (`planning/meta/`), free internal
   structure (`meta/routines/*.md`, `meta/HOWTO-execute.md`, ...). Hardcoded name to start;
   overrides/multiples only if pain appears. No config-ignore globs, no `.tskflwctlignore`.
5. **`README.md` exception.** A `README.md` at a bucket root is silently ignored (GitHub
   landing page). Everything else non-entity errors.
6. **`meta/` is ignore-only for now, promotable later.** e.g. `meta/routines/` is invisible
   today; the routines spike can later promote it to a scanned routine-entity path with zero
   file moves.

### Standardized vs configurable

**Standardized.** Classifier + `meta/` convention + `README.md` are zero-config conventions.
No `.tskflwctl.toml` `ignore` key and no dotfile - revisit only if a real frontmatter'd
non-entity case appears that the convention cannot express.

### Composition with isolation

Complementary granularities: `taskflow_root`/`planning_repo` evict whole *sibling dirs*
(`research/`, `archive/`, ...) out of scope; the classifier + `meta/` handle *files inside*
the scanned buckets. The `audits/HOWTO-execute.md` seam needs both - isolation lands it in
`planning/audits/`, and the migration sweeps it into `meta/`.

### Phase-B minimum vs incremental

- **Ship-now (unblocks the flatten):** the filename gate in `markdownCandidates` excludes
  strays from resolution + the parse path reports them as `FileProblem`s. Nearly free.
- **Incremental after:** bless `meta/` as a reserved constant; the `README.md` carve; the
  clearer stray-vs-malformed parse messages; the migration sweep of known loose files.

### Net implementation

Filename gate in `markdownCandidates` + `README.md` carve + a clearer parse-error message +
`meta/` as a constant + the migration sweep. No warning tier, no contract change.

### Deferred / coordinate

- **Epic 27 review shape (A vs B)** - decided *with* 27. Routing rule: a review you want
  **tracked** -> an entity (Shape B); throwaway notes -> `meta/`.
- **ADR-worthy:** "what makes a file an entity" (the filename classifier) + the stray-file
  contract belong in a short ADR, coordinated with epic
  [[26-frontmatter-schema-declared-validation-contract]].

## Epic analog implemented in Phase B (2026-07-05)

Beyond the decided design (README silently carved; any other non-id-led `.md` in a
scanned dir is a loud `FileProblem` → "move it to `meta/`"; non-`.md` files ignored;
`meta/` is the sanctioned home), the **epic** carveout shipped: `EpicNameIssue` flags a
non-`NN-<slug>` epic filename **fail-open** (still lists/resolves, mirroring
`StatusFellBack`) rather than dropping it. The one legacy non-NN epic was renamed
`00-taskflow-v1-core` first, so the real tree stays lint-clean.
