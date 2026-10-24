---
schema: 1
id: 6gppr2k482rm
status: ready-to-start
epic: 26-frontmatter-schema-declared-validation-contract
description: Accept or reject the proposed ADR-0008 amendment to annotate vocabulary-bearing fields with x-taskflow-vocabulary instead of enum, then implement and guard the outcome.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [schema, adr, wire]
created: "2026-10-23"
audit_sources: [planning/audits/6gjzzvdy5jnw-2026-10-12-arch-machine-contract.md]
updated_at: "2026-10-23"
---
# Decide how the generated JSON Schema expresses closed vocabularies

## Objective

ADR-0008 §3 makes closed-vocabulary changes the main trigger for a `NOT ADDITIVE` revision, and §4
makes the generated schema the validation artifact, yet the schema expresses none of the six
published vocabularies. An `enum` would break the tolerant reads `lint` depends on (ADR-0007 §4).
The audit proposes an `x-taskflow-vocabulary` annotation plus a stated rule for unknown values.
That amendment needs a human decision before implementation.

## Acceptance criteria

- [ ] The proposed ADR-0008 amendment is accepted, revised, or rejected, and recorded as an amendment rather than a rewrite.
- [ ] If accepted, every vocabulary-bearing property carries an annotation naming its `schema --json` registry, derived from the same domain owners, with no `enum` on tolerant-echo fields.
- [ ] If accepted, a guard test fails when a vocabulary-typed DTO field lacks the annotation, and `reader_expectation` states the unknown-value rule.
- [ ] Any revision is classified per ADR-0008, or the rejection reason is recorded on the finding.

## Out of scope

- The on-disk frontmatter field registry design, though it is the natural future source for these annotations.

## Related

- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md)
- Audit [2026-10-12 machine contract](../audits/6gjzzvdy5jnw-2026-10-12-arch-machine-contract.md), finding H1
- [ADR for frontmatter schema policy](6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md), which records the partial overlap
