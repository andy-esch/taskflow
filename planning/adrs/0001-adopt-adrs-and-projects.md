---
status: proposed
date: "2026-06-20"
deciders: [andy-esch]
tags: [adr, planning-model, project]
supersedes: []
superseded_by: []
---

# ADR-0001: Adopt ADRs and Projects in tskflwctl planning

> **Bootstrapping note.** This is the first ADR, and it proposes the very concept
> of ADRs — so it is written by hand (there is no `tskflwctl adr` command yet) and
> it deliberately *follows the format it proposes*. Same move Nygard's ADR-0001
> ("record architecture decisions") makes. The format here is the one
> `tskflwctl adr new` will later scaffold; this file is its first instance and its
> living spec. Design rationale (best-practice + cross-tool research) lives in
> [[2026-06-20-adrs-and-projects-format-design]].

## Context and Problem Statement

tskflwctl is a generic, local-first work-management tool: its planning types are
knowledge-organization primitives that `tskflwctl` scaffolds for **any** repo. It
self-hosts its own planning under `planning/` and today has three kinds —
**tasks** (the unit of work; `status == directory`), **epics** (a durable thematic
*home*; flat, frontmatter status), and **audits** (point-in-time code reviews;
`status == directory`). Two gaps have shown up while dogfooding:

1. **No home for *why*.** Significant decisions — "split list output into
   `-o`/`-c`", "pickers use bubbles/list not huh.Select", "prompts are TTY-gated" —
   live scattered across task bodies, "Decided (date)" notes, and session history.
   There is no durable, discoverable record of *decisions and their rationale*, and
   no first-class way to mark one decision as superseding another.
2. **No cross-cutting grouping.** Epics are *areas* (e.g. "CLI UX & ergonomics") and
   a task belongs to exactly one. But a cohesive push to ship one large feature
   often pulls tasks from *several* epics at once. There is no way to say "these 14
   tasks, across 4 epics, are the same initiative."

Both gaps are generic — every project that uses tskflwctl has decisions worth
recording and features that cut across themes — so the fix is two new **generic
document types**, scaffolded by the tool, not bespoke files.

## Considered Options

- **A — Add ADR + Project as first-class types** (this decision). Two new kinds the
  tool scaffolds, lints, and cross-links, alongside task/epic/audit.
- **B — Keep using `research/` + task bodies for decisions, and a `project:` tag
  for grouping.** No new types. Rejected: decisions stay undiscoverable and
  unsupersedable; an unvalidated tag gives no rollup and drifts (the existing
  `project` stub even references a non-existent command).
- **C — ADRs only, defer Projects.** Rejected: the cross-epic grouping gap is real
  today (e.g. the TUI work spans four epics) and the two share one schema/lint/CLI
  pass, so splitting them doubles the integration cost for no benefit.

## Decision

Introduce two new first-class planning concepts, formatted per the generic
templates the tool will scaffold.

### ADR — Architecture Decision Record

A durable record of one significant decision: its **context**, the **decision**,
and its **consequences**. The *why* layer.

- **Format (Nygard core + MADR's one addition).** Required: Title, frontmatter
  `status`, Context and Problem Statement, Decision, Consequences. Scaffolded by
  default: Considered Options, Related. Optional behind flags (`--with-drivers`,
  `--with-confirmation`): Decision Drivers, Confirmation/Validation, per-option
  Pros & Cons.
- **Status is a frontmatter field, canonical and lint-validated** — *not* also a
  `## Status` section (avoids the two-places-drift the house style warns against).
- **Mostly immutable.** You don't rewrite an accepted ADR's decision; you write a
  *new* ADR that supersedes it. History is append-only and auditable.
- **Numbered** (`NNNN`, zero-padded) — the number is its identity. Monotonic, never
  reused.
- **Lifecycle:** `proposed → accepted → superseded | deprecated | rejected`.
  *Superseded* = replaced by a specific new ADR (bidirectional link); *deprecated* =
  retired with no replacement.
- **Spawns work, isn't work.** An accepted ADR can spawn an **epic**, a **project**,
  or **tasks**; those work items cite it back. ADRs are not on the task board.

### Project — a cross-cutting feature effort

A cohesive, goal-oriented collection of tasks that **builds one large feature** and
may **cut across epics**. (Validated against industry as Linear's "Project" model;
note the SAFe terminology clash — in SAFe "Epic" means what we call a Project.)

- **Time-bounded and completable** — a project has a clear "done" (the feature
  shipped, by a `target_date`), unlike an epic's open-ended theme.
- **Orthogonal to epics.** A task keeps its single `epic:` (its thematic home) and
  *optionally* gains `projects:` (zero-or-more initiatives it serves). One task →
  one epic, zero-or-more projects.
- **Membership lives on the task** (`projects: [...]`), the single source of truth;
  the project's member list is **computed**, never hand-maintained — same pattern as
  `epic show`.
- Has a **goal**, a **status**, an optional **health** (distinct from status), and a
  progress rollup of its tasks *across* epic boundaries.

### How the three layers relate

```
ADRs        the WHY      — decisions + rationale, durable history
  │ spawn
  ▼
Epics  +  Projects       the WHAT/HOW — work organization
(durable    (cross-cutting,
 thematic    completable
 home)       feature push)
  │ contain
  ▼
Tasks       the unit of work — one epic (home), 0..N projects, may cite an ADR
```

- **Epic vs Project:** an **Epic is a *place* work lives** (durable home, one per
  task, never "done"); a **Project is a *finish line* work moves toward**
  (cross-cutting, time-bounded, completes; zero-or-more per task). *If it can never
  be done, it's an epic; if it has a target date and a definition of done, it's a
  project.*
- **ADR vs Epic/Project:** an ADR is a *decision*; epics/projects are *work*. An ADR
  spawns either; work references the ADR. They never substitute for each other.

### Cross-linking (two tiers)

Information flow must be legible, and today's links are inconsistent (three
syntaxes for "related task"; only `epic:` is validated). This decision standardizes
**one form per purpose**:

- **Tier 1 — structured frontmatter, lint-validated, bidirectional integrity kept by
  the tool:** task `epic:` (existing), task `projects: []` (new), task `adrs: []`
  (optional, cites decisions); project `spawned_by:` and `epics: []`; ADR
  `supersedes:` / `superseded_by:`.
- **Tier 2 — `[[wikilink]]` in prose, one canonical form** (no `.md`; slug for
  tasks/projects/research, `NN-slug` for epics, `ADR-NNNN` for ADRs). `lint` learns
  to resolve these; the TUI's `f`-nav learns to follow them.
- **Relative paths only for implementation code** (`internal/...`, `../desirelines/...`),
  never for sibling planning docs. The older inline `[slug.md]` / mixed
  `related_tasks` styles are dropped.

## Data model & layout

| Kind | Location | Identity | Status model |
| :-- | :-- | :-- | :-- |
| ADR | `planning/adrs/NNNN-slug.md` (**flat**) | the number | `status:` **frontmatter field** |
| Project | `planning/projects/<slug>.md` (**flat**) | the slug | `status:` **frontmatter field** |
| Task (changed) | unchanged | slug | adds optional `projects: []` and `adrs: []` |

- **ADR layout** (decided): flat + frontmatter status. ADRs behave **like epics**
  (numbered, durable, flat, status in frontmatter) — not like board items that flow
  through `status` directories. Their identity is the immutable number; they change
  status without moving files.
- **Project layout** (decided): flat + frontmatter status, **like epics** — *not*
  `status == directory`. Projects are orthogonal groupings (a task can be in
  several), few in number, and closed conversationally ("what's left in X?" →
  "nothing" → set `status: complete`), so status-subdirectories add churn without
  value. This supersedes the earlier draft that proposed `projects/<status>/<slug>.md`.
- **ADR home** (decided): `planning/adrs/`, co-located with the other planning
  types, since ADRs are a knowledge-organization type *within* taskflow.
  Configurable via `.tskflwctl.toml` (`adr_dir`) for repos that prefer the impl
  tree's `docs/adr`.
- **Project frontmatter:** `status` ∈ {`unstarted`, `in-progress`, `complete`,
  `abandoned`}, `description` (≤150 chars), `goal`, nullable `created` /
  `target_date` / `started_at` / `ended_at`, `tags`, optional `health` ∈
  {`on-track`, `at-risk`, `off-track`}, optional `epics: []` and `spawned_by:`.

## CLI surface (sketch, for the implementation epic)

- `tskflwctl adr new "Title"` (auto-numbered) · `adr list|show` · `adr accept` ·
  `adr supersede <n> --by <m>` (writes both links + flips old status) · `adr
  deprecate|reject`.
- `tskflwctl project new "Title" --goal … [--target-date …]` · `project list|show`
  (cross-epic rollup) · `project set <slug> <field> <val>` (`status complete` stamps
  `ended_at`).
- `tskflwctl task new … --project <slug>` ; `task set --projects …` / `--adrs …`.
- `tskflwctl schema` learns `adr` and `project`; `lint` validates the new
  frontmatter, ADR numbering, and the Tier-1 cross-links + the `[[wikilink]]`
  resolver.

## Consequences

**Positive.** Decisions become discoverable, citable, and supersedable. Large
features get a cohesive cross-epic view without distorting the epic taxonomy.
Cross-linking gets *one* validated form per purpose, so information flow is legible
to humans, lint, and the TUI. The tool becomes a more complete planning system — an
agent can read *why*, not just *what*.

**Negative / cost.** Two more concepts to learn and three more command groups to
build, test, and document. The cross-link standardization means migrating existing
inconsistent links. Risk of over-modeling if projects/ADRs are used rarely —
mitigated by keeping both optional and the ADR default template short (~1–2 pages).

## Open questions

1. **Numbering allocation.** `adr new` writes files like everything else — how does
   it allocate the next `NNNN` race-free (advisory `flock`? scan-max-plus-one with
   exclusive create + retry)? Implementation detail for the epic.
2. **Amend vs. supersede.** Should an accepted ADR support a lightweight, append-only
   "Confirmation/changelog" note (status stays `accepted`) for minor extensions,
   distinct from full supersession? (Best practice recognizes both; lean: allow a
   dated note under `## Related`, reserve supersession for reversals.)
3. **`[[wikilink]]` resolution scope.** Validate at lint time only, or also rewrite
   to canonical form on `--fix`? Resolved entities also unlock TUI follow.

*Resolved since the draft:* ADR status = frontmatter field (was open); projects =
orthogonal, not nested (was open); project layout = flat + frontmatter (was
`status == directory`); ADR home = `planning/adrs/` (was undecided); project rollup
= filtered `done/total` over members, computed independently of epic rollups so the
orthogonal axes never share a denominator (no double-counting).

## Implementation (once accepted)

On acceptance, this ADR spawns an epic — *Planning model: projects & ADRs* — with
tasks roughly: (a) ADR support (`planning/adrs/`, `adr` command, schema, lint,
numbering); (b) Project support (`planning/projects/`, `projects:`/`adrs:` task
fields, `project` command, cross-epic rollup); (c) cross-link tier-1 validation +
`[[wikilink]]` resolver + TUI `f`-nav extension; (d) docs + `schema adr|project` +
generated CLI reference. Kept out of scope until this ADR is **accepted**.

## Related

- Design rationale & research synthesis: [[2026-06-20-adrs-and-projects-format-design]].
- Project model origin: [[2026-06-06-project-concept-cross-cutting-initiatives]].
- Command spec (the `adr` / `project` groups): [[2026-06-06-tskflwctl-command-spec]].
- Format lineage: Michael Nygard's ADRs + MADR (Markdown Any Decision Records).
