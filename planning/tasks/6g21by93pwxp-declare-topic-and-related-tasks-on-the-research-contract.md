---
schema: 1
id: 6g21by93pwxp
status: next-up
epic: 28-first-class-entities-new-planning-nouns
description: 'Adopt the two fields the corpus already invented: topic as curation axis, related_tasks as lint-validated linkage — reversing no-cross-references.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, contract]
created: "2026-08-20"
audited: "2026-10-04"
audit_sources: [2026-10-04-weekly-task-sweep]
updated_at: "2026-10-04"
---

# Declare topic and related_tasks on the research contract

## Objective

The corpus invented both fields because the contract lacked them. Adopt what authors
already write, rather than designing something new.

    5  topic           a one-line subject, distinct from description
    4  related_tasks   real slug lists pointing at the tasks a doc informs
    1  related_research / related_adrs   (too thin to declare)

## `related_tasks` — and the reversal it records

**This partially reverses the epic-28 "no cross-references, not even `epic:`" decision**,
which is currently stated in the README, `domain/research.go`'s doc comment, and the
research Descriptor conventions. That decision was reasoned from *"provenance is a body
concern"*. The corpus has since shown authors wanting it structured — four docs carry
`related_tasks` with real slugs:

    go-cli-foundation-architecture:
      - port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec.md
      - rethink-pm-command-hierarchy-pm-noun-verb-research-cli-best-practices.md

Two things make this the right shape where `epic:` was not:

- **A list, so many-to-many is free.** Research is genuinely many-to-many with epics (the
  color-palette doc informs a task in epic 20; the ADRs/projects doc informs two ADRs), so
  a singular `epic:` was the wrong shape. Note also that epics do NOT link down to tasks —
  a task carries `epic:` and the rollup is derived — so "linkage like epics have" would
  have meant research pointing up at one epic. `related_tasks` avoids that.
- **It points at tasks, not epics**, which is the relationship authors actually recorded.

Update the prose that states the old decision, and say WHY it changed, so the reversal is
legible rather than looking like drift.

## `topic` — the curation axis

Distinct from the other three handles: the slug is a filename, `description` is a summary
of findings, `topic` is the stable subject. E.g. *"Go CLI foundation — layout,
architecture, and patterns for the pm port"*. It is what makes the corpus groupable, and
pairs with the tag backfill.

## Acceptance criteria

- [ ] `topic` (string, optional) and `related_tasks` (list, optional) declared in the
      research Descriptor, so they flow to `schema research`, `research_fields`, and
      `research set` automatically — no second list.
- [ ] `related_tasks` entries are lint-validated as resolvable task references; a dangling
      one is a lint finding, not silent. Decide whether the value is a slug or an id-led
      filename and normalise the 4 existing docs to it.
- [ ] Both settable via `research set`; `SettableResearchFields()` picks them up
      automatically (they are not protected).
- [ ] `schema_version` bumped with a changelog entry; goldens regenerated.
- [ ] The "no cross-references" statements in README, `domain/research.go`, and the
      Descriptor conventions updated to describe the new shape and the reason for the change.
- [ ] `purpose` (×5) and `date` (×5) deliberately NOT declared — `purpose` overlaps
      `description` and the body's opening paragraph; `date` duplicates `created` and is
      stripped by the sibling task.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Decision record: [formalize-the-research-frontmatter-contract](6g0493z6xwc7-formalize-the-research-frontmatter-contract.md)
- Ordering note: this and the strip task both touch the 5 docs carrying `topic`/`date`;
  do the strip first so the contract change lands on a clean corpus.

## Sweep verification (2026-10-04)

Premise verified accurate and unstarted. The research `Descriptor`
(`internal/domain/entity.go`) still declares only `description` and `tags` as
authoring fields, and its conventions still carry the old line: *"research carries
no `epic:`/`tasks:` fields — link to entities from the body with ordinary
relative-path links instead."* Nothing in the contract has moved.

Corpus counts in the Objective re-measured against `planning/research/` (32 docs)
and all still exact: `topic` ×5, `related_tasks` ×4, `related_research` ×1,
`related_adrs` ×1, `purpose` ×5, `date` ×5, vestigial `status` ×18.

Two location corrections for whoever implements this:

- **The third prose site is not the README.** The "no cross-references" decision is
  stated in `internal/domain/research.go:16` and — unmentioned by this task —
  `internal/domain/lint.go:190` ("no cross-reference to …"). `README.md` carries no
  such statement; it defers to `schema research` (README:205, README:374-375). The
  last acceptance criterion names README as a site to update; the actual third site
  is `lint.go`. Left unticked and unedited — rescoping a criterion is a human call.
- **`related_tasks` is already a declared *task* field.** `internal/domain/fields.go:34`
  lists `{"related_tasks", "list"}` in `taskFields`, so the key already coerces as a
  list on the task entity. Declaring it on research is still net-new work, but the
  name is not free of precedent and the two must not be conflated.

2026-10-04: automated weekly sweep — contract verified unstarted, corpus counts
re-confirmed exact, third "no cross-references" prose site corrected from README to
`internal/domain/lint.go:190`; description tightened to 148 chars.
