---
schema: 1
id: 6g21c0ga23bv
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: Supersession stays prose (every real instance is partial), but it is currently ad-hoc across three docs — give it one shape in the template and the conventions.
effort: Unknown
tier: 4
priority: low
autonomy_level: 3
tags: [planning, docs]
created: "2026-08-20"
---

# A prose `## Supersedes` convention for research docs

## Objective

Supersession was decided as **prose, not a field** (see
[formalize-the-research-frontmatter-contract](6g0493z6xwc7-formalize-the-research-frontmatter-contract.md)): every real instance in the corpus is
*partial*, and a `superseded_by:` pair can only assert whole-doc death. But right now the
prose form is ad-hoc — three docs express it three ways, in three different places. Give it
one shape.

## The pattern that already works

    Supersedes the over-reaching parts of
    `6fakbec02jvw-tui-ux-design-and-navigation-spec.md` (Projects-tab, multi-select).

Names the doc AND which parts. That specificity is the whole reason prose beats a field
here — and it is what the other two instances lack.

## Acceptance criteria

- [ ] An optional `## Supersedes` section in the research body template, with a one-line
      hint showing the shape: link the doc, then say which parts.
- [ ] `schema research`'s conventions mention it, so an agent authoring a doc discovers it.
- [ ] The three existing instances normalised into that section
      (`tui-design-decisions`, `fang-evaluation-spike`, and the intra-doc one in
      `go-cli-foundation-architecture` — which may be better left as inline prose, since it
      supersedes earlier calls WITHIN the same doc rather than another doc; judge it).
- [ ] Links are relative-path markdown so `lint --links` validates them for free.

## Out of scope

- Any frontmatter field for supersession — decided against, with evidence.
- Machine-queryable supersession. If it is ever wanted, the honest unit is a claim, not a
  doc, and that is a much larger design question.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Decision record: [formalize-the-research-frontmatter-contract](6g0493z6xwc7-formalize-the-research-frontmatter-contract.md)
