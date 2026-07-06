---
schema: 1
status: active
description: 'A single declared frontmatter schema (per-entity, per-status fields + types) that lint, `schema task` authoring guidance, and the `--json` contract all derive from — replacing today''s split between hand-written LintTask checks and the envelope JSON schema. Design-first; strictness and field-registry policy are the open pieces.'
priority: medium
tags: [validation, schema, frontmatter, core, cli]
created: "2026-07-03"
---
# Frontmatter schema — one declared validation contract

**Status: design-first. NOT ready to implement.** This epic collects the decisions
needed before any schema code is written. The first task here is to *close the
policy questions* below (probably an ADR), not to author a validator. Most
load-bearing choices are still **open**.

## Why this exists

Frontmatter validation is currently spread across three places that don't share a
source of truth:

- **`domain.LintTask`** — hand-written field checks (status/epic/tier/priority/
  effort/created/tags), applied **only to active tasks**. Archived tasks get just
  the universal checks (misfiled, missing-id, and — new — missing-frontmatter).
- **`schema task`** — human authoring guidance (a separately-maintained prose
  description of the same fields).
- **`schema --json-schema`** — a Draft 2020-12 schema, but for the **`--json`
  envelopes**, not for on-disk frontmatter.

So the "what is a valid task file" contract is triplicated and can drift, and the
archived-vs-active asymmetry means a completed task can carry near-empty (or
subtly wrong) frontmatter and sail through. This surfaced concretely while
cleaning `desirelines-planning`: nine archived entities with missing/dateless/
malformed frontmatter that the active-only lint never flagged (see the
`2026-07-03` cleanup — missing-frontmatter files, `deprecated_date` vs the
recognized `deprecated_at`, a dateless-backfill gap, and a slug collision).

**The guiding principle (already adopted):** *fail loudly on invalid frontmatter
and describe the valid shape; auto-fix only the safe, unambiguous cases.* Recent
work put down two markers along this line — the filename-date id fallback, and
`parseTask` now failing loudly (naming `schema task`) on a fence-less/malformed
file instead of parsing it as an empty task. A declared schema is the natural home
for "the valid shape."

## Direction so far (leans, NOT final)

- **One declarative field registry** as the single source of truth (Go, likely),
  from which lint rules, `schema <entity>` guidance, and a frontmatter JSON-schema
  are all **generated** — no third hand-maintained copy.
- Keep the **"fail loud, don't over-fix"** contract: the schema *describes* valid;
  `--fix` stays limited to safe repairs (quote-`:`, list normalization, id
  backfill from a real/inferred date). It does **not** synthesize frontmatter.
- Model **tasks, epics, and audits** through the same mechanism (they already share
  `Issue`/lint plumbing).

## Open questions (this is the work)

1. **Per-status strictness matrix.** How strict should archived/deprecated tasks
   be? Today they're deliberately spared the field nags ("no point demanding a
   priority on a dead task"). Do we want a *required-for-active / required-always /
   optional* classification per field — and is `created` (needed for a stable id)
   required-always while `tier`/`effort` stay active-only?
2. **Unknown/custom fields.** Preserve-and-ignore (current behavior, and what
   surgical edits rely on) vs a closed registry that flags strays vs a namespaced
   escape hatch (`x-*`)? The `deprecated_date` (nonstandard) vs `deprecated_at`
   (recognized) case is the motivating example — is that "unknown field" or
   "misspelled known field"?
3. **Aliases & field migration.** How do we treat legacy field names? Auto-migrate
   under `--fix`, warn-only, or leave alone? If migrate, is that the *one* case
   where `--fix` may rewrite semantic frontmatter, and how is it bounded?
4. **Where the schema lives & how it relates to `schema --json-schema`.** Go
   structs+tags, a checked-in YAML/JSON descriptor, or generated both ways? Does
   the existing envelope JSON-schema stay separate, or does a *frontmatter*
   JSON-schema join it (two schemas, clearly named)?
5. **Types & value vocabularies.** Which fields get closed enums (priority) vs
   ranges (tier 1–5) vs patterns (date `YYYY-MM-DD`) vs freeform (`effort` —
   "1 hour" / "3-5 days" today; do we constrain it)? Where do tag conventions fit?
6. **Severity levels.** Does lint grow `error | warn | info`? Several of these
   (unknown field, nonstandard alias, freeform-effort) feel like warnings, not hard
   failures — but the current model is binary. Adding severities touches the
   `--json` contract and exit-code mapping.
7. **Cross-field / referential rules.** Are these part of "the schema" or a
   separate pass: `epic` existence, `superseded_by`/`parent_task` pointer validity,
   `related_tasks` resolvability, `blocks` symmetry, date ordering
   (`created ≤ completed_at`)?
8. **Backward-compatibility / rollout.** Turning on strict always-validation lights
   up large existing debt across real repos (we just saw it). Grandfathering by
   date? Opt-in strictness flag? A `lint --strict` vs default? A one-time
   `--fix`-then-enforce migration?
9. **Entity coverage & sharing.** One schema engine for tasks/epics/audits, or
   per-entity schemas over a shared core? Epics already carry a `schema: 1` field —
   is that the version handle we build on?
10. **Schema versioning & evolution.** How does a repo declare which schema version
    it targets, and how do we evolve the schema without breaking older repos
    (relation to the envelope `schema_version`)?
11. **Authoring UX downstream.** Should `schema task` output and `task new`
    scaffolding both be *generated* from the declared schema (so guidance can't
    drift from enforcement)? Does the loud "missing frontmatter" message point at
    generated content?
12. **External planning repos (epic 23).** Can a downstream impl/planning repo
    extend or override the schema (custom fields/statuses), or is the schema fixed
    by `tskflwctl`'s version?

## Non-goals (for now)

- Rewriting `--fix` into a general frontmatter repairer. The contract stays: loud
  failure + described shape; safe repairs only.
- A runtime/plugin schema language. Start with a fixed, code-declared schema.

## Related

- Epic 24 (data-model evolution) — status-in-frontmatter would reshape the
  per-status strictness question (Q1); coordinate before committing.
- Epic 20 (CLI UX & ergonomics) — severity levels (Q6) and `schema`-output
  generation (Q11) are ergonomics-adjacent.
- Epic 23 (external planning repo) — schema extension/override (Q12).
- Settled input: the **carveout contract** — what makes a file an entity (a filename-shape
  classifier), stray files as `FileProblem`s, and the `meta/` folder — is decided in
  [curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate](../tasks/6fjvr03mr9zg-curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate.md) and recorded as
  an ADR-0003 amendment (2026-07-04). Fold it into this epic's "valid entity" formalization.
- Prior art in-repo: `domain.LintTask`, `domain.MissingIDIssue`,
  `store.parseTask`'s loud missing-frontmatter failure, `schema --json-schema`.
