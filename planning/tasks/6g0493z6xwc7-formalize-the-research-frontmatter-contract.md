---
schema: 1
id: 6g0493z6xwc7
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: 'Settle the three fields left open: supersedes/superseded_by, the vestigial status: reference on 18 legacy docs, and the empty description across the migrated corpus.'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [core, design]
created: "2026-08-14"
---

# Formalize the research frontmatter contract

## Objective

Research shipped with the **minimum viable** contract — `schema`, `id`, `created`
(required), `description`, `tags`. That was the right first cut, but three questions were
consciously left open. Settle them, and land the answers through the field registry
rather than as another hand-maintained per-kind list.

## The open questions

### 1. `supersedes:` / `superseded_by:` — how research obsoletes itself

This is the substantive one. Research has **no lifecycle** by decision (ADR-0001 draws
that line at ADRs: a research doc is a snapshot a later doc freely supersedes, not a
decision that freezes). But "freely supersedes" is currently unmodelled — there is no way
to say *this doc replaces that one*, so a reader can land on a stale doc with no signal
that a newer one exists.

The tension to resolve: a `supersedes:` pair is genuinely useful, but it is also **a
cross-reference and a soft lifecycle**, both of which the first pass deliberately
excluded. Options:

- **A — the pair, research→research only.** Obsolescence without status. Stays inside
  research (no epic/task coupling), so it doesn't reopen the no-cross-references
  decision, which was about *entity* references.
- **B — leave it to body prose.** A "superseded by [X]" line at the top costs nothing
  and stays unstructured, consistent with "provenance is a body concern".
- **C — a `status:` vocabulary** (`draft|reference|superseded`). Rejected once already —
  it reopens ADR-0001 directly. Named here only to keep the rejection on the record.

### 2. The vestigial `status: reference`

18 legacy docs carry it. It is **not** in the contract, not linted, and rides along as a
preserved unknown key. Decide: strip it in a cleanup pass, or formally declare it (which
is option C above and reopens ADR-0001). Leaving it indefinitely is also a choice — it is
currently harmless but it *looks* like a field the tool honors, which is misleading.

### 3. `description` is empty across the whole migrated corpus

Unrecoverable at migration time, so all 28 docs show `—` in `research list`. Options:
derive from each H1, hand-write them, or accept the gap. Wants
[research-mutation-verbs-research-set-edit-append](6g048wqgamte-research-mutation-verbs-research-set-edit-append.md) first, for a scriptable way to backfill.

## Acceptance criteria

- [ ] Each of the three questions above has a decision recorded (an ADR if it changes the
      ADR-0001 boundary; otherwise a note in this task + epic 28).
- [ ] Any new field is declared through the field registry (epic 26) rather than a
      hand-maintained per-kind list, and is lint-checked.
- [ ] `schema research` output reflects the settled contract.
- [ ] `schema_version` bumped if the `--json` shape changes.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md) — the registry any new field should ride
- [ADR-0001](../adrs/0001-adopt-adrs.md) — the research/ADR boundary this must not blur
