---
schema: 1
id: 6g21brp13m8d
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: 18 docs carry an undeclared status in five spellings (4 provably stale) and 5 carry a date duplicating created; both look like fields the tool honours when it ignores them.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [planning, hygiene]
created: "2026-08-20"
---

# Strip the vestigial status and date keys from the research corpus

## Objective

Research has **no status** by decision (see
[formalize-the-research-frontmatter-contract](6g0493z6xwc7-formalize-the-research-frontmatter-contract.md)). 18 of 30 docs nonetheless
carry one, and 5 carry a `date:` that duplicates `created:`. Both are undeclared, unlinted,
and preserved only by the surgical-edit guarantee — so they look like fields the tool
honours when it ignores them entirely. Remove them.

## What is there

    18  status      reference ×12 · in-progress ×4 · proposed ×2 · proposal ×1 · unstarted ×1
     5  date        duplicate of created, same value

`proposal` and `proposed` are the same idea spelled two ways, which is itself evidence that
nothing validated these. The four `in-progress` values are provably stale: all four are the
2026-06-06 cohort, whose work (the Go CLI, the pm retirement, ADR-0002, the command spec)
shipped long ago. They tracked *the work the doc informed*, not the doc.

## The prose markers are a separate call

Nine docs also carry a `**Status**: Proposal` line in the BODY. Those are prose, not
frontmatter, and a reader can see they are dated commentary. Leave them: stripping body
prose is editorial, not schema hygiene. If they should go, that is its own pass.

## Acceptance criteria

- [ ] `status:` and `date:` removed from every research doc; no other key touched.
- [ ] The diff is deletion-only, verified by counting the diff by line type — a surgical
      write that reflows an unrelated value would be a silent regression (see the known
      block-scalar refold issue on epic 21).
- [ ] `research list`, `research show`, and `lint --links` clean afterward; the 30 docs
      still parse and every id/created pair is unchanged.
- [ ] Body prose `**Status**:` markers left alone, deliberately.

## Out of scope

- Declaring any status vocabulary — decided against.
- The 9 body prose markers.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Decision record: [formalize-the-research-frontmatter-contract](6g0493z6xwc7-formalize-the-research-frontmatter-contract.md)
