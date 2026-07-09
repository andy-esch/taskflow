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
