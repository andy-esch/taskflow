---
schema: 1
id: 6fkkz41cax80
status: next-up
epic: 26-frontmatter-schema-declared-validation-contract
description: Design-first ADR closing epic 26's policy questions (strictness, unknown fields, schema location, severities, rollout) so one field registry drives lint, schema guidance, and the --json contract.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [validation, schema, adr]
created: "2026-07-07"
---
# ADR — declared frontmatter-schema contract: close the policy questions

## Objective

Epic 26 is design-first: before any validator is written, close the open policy
questions so a *single declared field registry* can generate lint rules, `schema
<entity>` authoring guidance, and a frontmatter JSON-schema from one source of
truth (replacing today's triplicated, drift-prone split). Output: an ADR (the
next number after 0003/0004) recording the decisions, plus this epic's open
questions resolved or explicitly deferred.

## Survey — fill this out

Andy fills this in; the answers become the ADR. For each question: tick a box
with `[x]` (options are mutually exclusive unless it says *pick any*), and use
**Notes** for the "why", caveats, or a different answer. `*(current lean)*` marks
where epic 26's body already leans — override freely. `*(load-bearing)*` marks
the five the ADR must decide (Q1, Q2, Q4, Q6, Q8); the other seven may be
decided *or* explicitly parked with a reason. "Park" = `[x] Park` + a one-line why.

---

### Q1 — Per-status strictness matrix  *(load-bearing)*

How strict should archived/deprecated tasks be? Today active tasks get field
nags; archived get only the universal checks (misfiled / missing-id / missing-
frontmatter).

- [ ] Per-field classification: **required-always / required-for-active / optional** *(current lean)*
- [ ] Keep today's binary split (active = full nags, archived = universal only)
- [ ] Fully strict — every known field required regardless of status

Sub-question — is `created` **required-always** (it seeds a stable id) while
`tier`/`effort` stay active-only? &nbsp; [ ] yes &nbsp; [ ] no

> Coupling: epic 24 (status-in-frontmatter) reshapes this axis — coordinate before finalizing.

**Notes:**

---

### Q2 — Unknown / custom fields  *(load-bearing)*

Preserve-and-ignore (today) vs a closed registry vs a namespaced escape hatch.
Surgical edits currently rely on unknown fields surviving untouched.

- [ ] Preserve-and-ignore unknown fields (current behavior)
- [ ] Closed registry — flag any unrecognized field as a problem
- [ ] Namespaced escape hatch — unknown allowed only under `x-*`, others flagged
- [ ] Closed registry **plus** `x-*` escape hatch

Sub-question — is `deprecated_date` (vs recognized `deprecated_at`) an *unknown
field* or a *misspelled known field* (→ handled as an alias, Q3)? &nbsp; [ ] unknown &nbsp; [ ] misspelled-known

**Notes:** *(draft 2026-07-10 — boxes left for you)* `deprecated_date` reads as a
**misspelled known field** → handled as an alias (Q3), not a genuine unknown
field. Leaning Q2 → *closed registry + `x-*` escape hatch*: `x-*` is the
sanctioned preserve-and-ignore zone (forward-compat + human scratch), so a bare
unrecognized key *outside* `x-*` is suspicious-by-default (probably a typo). That
split also gives Q6 its cut: unknown-in-`x-*` → at most `info`; bare unrecognized
key → `warn`; only structural/identity violations → `error`.

---

### Q3 — Aliases & field migration

How to treat legacy field names under `--fix`.

- [ ] Leave alone — never rewrite
- [ ] Warn-only (lint flags, `--fix` does not touch)
- [ ] Auto-migrate under `--fix` — the *one* sanctioned semantic rewrite, bounded to a fixed alias table
- [ ] Park

**Notes (if auto-migrate: what bounds it?):** *(leaning auto-migrate, decided in
chat 2026-07-10 — box left for you)* Bound to a **fixed alias table** (seed:
`deprecated_date → deprecated_at`). This is the *one* sanctioned semantic rewrite;
it lives in `--fix` (and a future `migrate`), never in a general repairer — so
contract C holds. Framing worth recording: entities already have "tag numbers"
(task/audit opaque `id`, epic `NN`) that decouple identity from the renameable
slug, which is why entity renames are safe. Frontmatter **keys have no such tag
layer**, so a key rename is breaking — the alias table *is* the field-level
equivalent of that entity-id decoupling.

---

### Q4 — Where the schema lives & its relation to `schema --json-schema`  *(load-bearing)*

Two sub-decisions.

**(a) Source of truth:**
- [ ] Go structs + tags (code-declared) *(current lean)*
- [ ] Checked-in YAML/JSON descriptor
- [ ] Go-declared, with a descriptor generated as a build artifact

**(b) Frontmatter schema vs the existing envelope JSON-schema:**
- [ ] Two clearly-named schemas, kept separate (envelope stays; a *frontmatter* JSON-schema joins it)
- [ ] Unify into one
- [ ] Frontmatter schema only for lint/authoring; do not emit a JSON-schema for it

**Notes:**

---

### Q5 — Types & value vocabularies

Which fields get constrained, and how. Fill the ones you have an opinion on.

| field | constraint |
|---|---|
| `priority` | [ ] closed enum → values: `__________` &nbsp; [ ] freeform |
| `tier` | [ ] int range `1`–`5` &nbsp; [ ] other: `______` |
| dates (`created`, `*_at`, …) | [ ] pattern `YYYY-MM-DD` (confirm) |
| `effort` | [ ] freeform, keep as-is &nbsp; [ ] constrain → how: `__________` |
| `tags` | [ ] freeform &nbsp; [ ] conventions: `__________` |

- [ ] Park the whole question (types come with the validator, not this ADR)

**Notes:**

---

### Q6 — Severity levels  *(load-bearing)*

Does lint grow beyond binary valid/invalid? Adding severities touches the
`--json` contract and exit-code mapping.

- [ ] Add `error | warn | info` *(current lean — unknown-field / alias / freeform-effort feel like warnings)*
- [ ] Keep binary (valid / invalid)

If adding severities — which land as **warn** (not hard failure)? *(pick any)*
- [ ] unknown field &nbsp; [ ] nonstandard alias &nbsp; [ ] freeform `effort` &nbsp; [ ] other: `______`

**Notes (exit-code / `--json` impact):**

---

### Q7 — Cross-field / referential rules

Are these part of "the schema" or a separate pass?

- [ ] Part of the declared schema
- [ ] A separate validation pass
- [ ] Park (out of scope for this ADR)

Which rules are in scope now? *(pick any)*
- [ ] `epic` exists &nbsp; [ ] pointer validity (`superseded_by` / `parent_task`) &nbsp; [ ] `related_tasks` resolvable &nbsp; [ ] `blocks` symmetry &nbsp; [ ] date ordering (`created ≤ completed_at`)

**Notes:**

---

### Q8 — Backward-compatibility / rollout  *(load-bearing)*

Turning on strict always-validation lights up real existing debt (we just saw
nine such files in desirelines-planning).

- [ ] Grandfather by date (files before cutoff exempt)
- [ ] Opt-in `lint --strict`; lenient default
- [ ] One-time `--fix`-then-enforce migration
- [ ] Strict-by-default immediately

**Notes:**

---

### Q9 — Entity coverage & sharing

- [ ] One engine, per-entity schemas over a shared core (tasks/epics/audits) *(current lean)*
- [ ] Separate schema per entity

Build on the existing `schema: 1` field as the version handle? &nbsp; [ ] yes &nbsp; [ ] no

**Notes:**

---

### Q10 — Schema versioning & evolution

How a repo declares which schema version it targets; how the schema evolves
without breaking older repos.

- [ ] Reuse the `schema:` frontmatter field as the version handle
- [ ] New dedicated field: `__________`
- [ ] Tied to the `tskflwctl` binary version (no per-repo declaration)
- [ ] Park for a later ADR

Relation to the envelope `schema_version`: distinct handles — `schema:` versions
the **on-disk file shape**, `schema_version` versions the **`--json` output
contract**; neither drives the other (already true in `layout.go`).

**Notes:** *(draft 2026-07-10)* **Park the migration framework** — YAGNI at
`schema: 1` (never bumped); the first breaking change is a one-off transform, not
a registry. Keep `schema:` **coarse** (bump only on a true reshape). Read-tolerance
is *already free*: the loader ignores `schema:` today, so old / `schema: 2` /
no-`schema:` files all read now — preserve that (on an unknown-higher version,
degrade + warn, never hard-fail).

> **Schema change rules (fold into the ADR body):** (1) add fields freely, never
> bump `schema:`; (2) rename = alias-table entry + bounded `--fix`, never in
> place; (3) a type/semantic change of an existing key is a *new* field (or, if
> unavoidable, a `schema:` bump with reader branching); (4) retired field names
> are tombstoned and never reused — the enforced epic-NN duplicate guard already
> *is* this rule, at the entity level. This change-discipline, not a "perfect"
> schema, is the futureproofing.

---

### Q11 — Authoring UX downstream

- [ ] Generate `schema task` output **and** `task new` scaffolding from the declared schema, so guidance can't drift from enforcement *(current lean)*
- [ ] Keep them hand-maintained for now
- [ ] Park

Should the loud "missing frontmatter" message point at generated valid-shape
content? &nbsp; [ ] yes &nbsp; [ ] no

**Notes:**

---

### Q12 — External planning repos (epic 23)

Can a downstream impl/planning repo extend or override the schema?

- [ ] Schema fixed by `tskflwctl` version — no override
- [ ] Downstream repos may extend (add custom fields / statuses)
- [ ] Park until epic 23 matures

**Notes:**

---

### C — Contract check (confirm, don't re-open)

The "**fail loud on invalid frontmatter, describe the valid shape, auto-fix only
safe/unambiguous cases**" contract is preserved; `--fix` does **not** become a
general repairer. &nbsp; [ ] confirmed

**Notes (any exception, e.g. the Q3 alias case):**

---

### D — Carveout amendment fold-in

Fold the settled ADR-0003 entity-carveout amendment (filename-shape classifier,
stray files as `FileProblem`s, the `meta/` folder) into this epic's "valid
entity" formalization as settled input. &nbsp; [ ] yes &nbsp; [ ] revisit

**Notes:**

---

### E — ADR number & logistics

Next number after 0003 / 0004 → **ADR-0005**. &nbsp; [ ] confirm &nbsp; [ ] other: `______`

**Anything else / global overrides:**

---

## Acceptance criteria

- [ ] ADR drafted covering the load-bearing questions: per-status strictness
      matrix (Q1), unknown/custom-field policy (Q2), where the schema lives and
      its relation to `schema --json-schema` (Q4), severity levels (Q6), and
      backward-compat/rollout (Q8).
- [ ] Each of epic 26's 12 open questions is either decided or explicitly parked
      with a reason.
- [ ] The "fail loud, don't over-fix" contract is preserved (non-goal: turning
      `--fix` into a general repairer).
- [ ] The ADR-0003 carveout amendment (what makes a file an entity) is folded in.

## Out of scope

- Implementing the field registry / validator (a follow-on task once decided).
- A runtime/plugin schema language — start code-declared and fixed.

## Related

- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md) — the 12 open questions live in its body.
- Prior art: `domain.LintTask`, `domain.MissingIDIssue`, `store.parseTask`'s loud
  missing-frontmatter failure, `schema --json-schema`.
