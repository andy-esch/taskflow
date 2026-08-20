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
updated_at: "2026-08-20"
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

### 3. ~~`description` is empty across the whole migrated corpus~~ — RESOLVED 2026-08-19

Backfilled by hand via `research set --description`, each written from the doc's H1 and
opening prose rather than paraphrasing its slug. `research list` now shows a real summary
for every doc (0 of 30 missing, including two authored since). Nothing left to decide here;
kept for the record because it was one of the three questions this task was opened on.

The original framing, for context: unrecoverable at migration time, so all 28 docs showed
`—` in `research list`. Options were to derive from each H1, hand-write them, or accept the
gap. It wanted
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

### 2026-08-19 — measured scope, post-v0.16.0

The description backfill shipped, so that half is done: **0 of 30** research docs have an
empty `description`. What remains is a single, countable defect.

**18 of 30 docs still carry a `status:` field that is not in the research contract**
(`schema research` declares `created` / `description` / `tags` only). The values are
incoherent, which is the point — they are fossils from before research had a contract:

| value | count |
| --- | --- |
| `reference` | 12 |
| `in-progress` | 4 |
| `proposed` | 2 |
| `unstarted` | 1 |
| `proposal` | 1 |

`in-progress` and `unstarted` are especially wrong: research **has no lifecycle** by
design (ADR-0001 — a later doc supersedes an earlier one; a decision needing a lifecycle
is an ADR). A doc claiming to be "in-progress" contradicts the entity's defining
property.

**`lint` does not flag any of them today** — unknown frontmatter keys are preserved and
tolerated. So this is invisible until someone reads a file.

The real decision this task has to make is therefore **not** "delete 18 fields", it is:

- Does `lint` learn to flag unknown keys for research (and if so, only research, or every
  entity)? That is a policy change with blast radius beyond this corpus — the surgical-
  preservation rule exists so hand-written fields survive.
- Or is this a one-shot cleanup with no lint change, accepting that the next stray key is
  equally invisible?
- If flagged: warn or error? `lint --fix`-able, or hand-edit only?

Note the overlap with epic 26 (declared frontmatter validation contract) — that epic is
where "which fields are legal per entity" is supposed to become declarative. Doing an
ad-hoc unknown-key check here could pre-empt or contradict it; check before building.

## Decisions (2026-08-20) — from an evidence review of the corpus

Questions 1 and 2 are **closed**. Both were decided from what the 30-doc corpus actually
does, not from first principles.

### Q1 `supersedes:` / `superseded_by:` — NO FIELD; prose convention instead

Option **B**. A structured pair would be lossy in every real instance. All supersession in
the corpus is **partial**, and none is whole-doc:

- `tui-design-decisions` — "Supersedes **the over-reaching parts of**
  tui-ux-design-and-navigation-spec (Projects-tab, multi-select)"
- `go-cli-foundation-architecture` — "Supersedes both the original 'goccy AST' and last
  round's 'canonical struct' calls" — **intra-document**, not doc-to-doc at all
- `fang-evaluation-spike` — "supersedes the man note in [a **task**]" — partial AND
  cross-kind

A `superseded_by:` pair asserts "this whole doc is dead", which is false in all three. The
underlying reason: a research doc is a **snapshot** — it stays a true record of what was
known then. What expires is whether you should still act on a given claim, which is
per-claim, not per-doc. Prose already carries that ("which parts", "which note"); a field
cannot. Conventionalized as an optional `## Supersedes` section — see the sibling task.

Corroborating: the one doc that WAS partially superseded still carries
`status: proposal`. Even the clear case went unmarked, so a field would not have been
maintained either.

### Q2 the vestigial `status: reference` — STRIP, do not declare

Research gets **no status vocabulary**. ADR-0001's reasoning holds, and is now backed by
evidence rather than assertion.

Authors did reach for a status — 18 of 30 docs carry one, in **five** values
(`reference` ×12, `in-progress` ×4, `proposed` ×2, `proposal` ×1, `unstarted` ×1 — note
`proposal` and `proposed` are the same idea spelled two ways), plus 9 docs with a second
prose `**Status**:` marker in the body. But the only lifecycle-shaped value in use is
**4/4 stale**: all four `in-progress` docs are the 2026-06-06 cohort, and the Go CLI
shipped, `pm` was retired, projects became ADR-0002, and the command spec was implemented.
That status tracked *the work the doc informed*, not the doc — and nothing updated it for
two and a half months. Declaring a status means declaring a field that rots plus a lint
rule that nags about it.

### What the demand actually was: discovery, not status

The reader question a status would answer is "what is the current guidance on X?" That is
already answered by `research list --tag X`, which is newest-first — no new field needed.
The gap is that **14 of 30 docs have no usable tags**. Backfilling them buys the currency
signal with existing machinery. Tracked as its own task.

Implementation is split across four sibling tasks (strip · tags · contract additions ·
prose convention) rather than done here; this task's job was the decisions.
