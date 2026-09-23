---
schema: 1
id: 6gcqz5b0j0dg
status: ready-to-start
epic: 26-frontmatter-schema-declared-validation-contract
description: Publish stable diagnostic kinds and explicit repair applicability so agents never route lint findings by prose.
effort: M
tier: 2
priority: medium
autonomy_level: 2
tags: [lint, diagnostics, json, agents]
created: "2026-09-22"
audit_sources: [planning/audits/6gc7jd9aq1q9-2026-09-21-arch-failure-and-recovery.md]
depends_on: [6fkkz41cax80, 6gcqz5aefjjf]
updated_at: "2026-09-22"
---

# Give lint diagnostics stable kinds and repair applicability

## Objective

Make repository lint a stable agent-facing diagnostic protocol instead of a collection of English
messages. Every reportable defect should carry a domain-owned kind and an explicit repair
applicability, while prose remains explanatory and free to improve without becoming routing logic.

## Design gate

Reconcile the vocabulary with the frontmatter-schema policy ADR and ADR-0007 before implementation.
The design must decide naming and compatibility rules for diagnostic kinds and the closed
applicability set (for example `fix`, `manual`, and `refused`) without implying that every defect is
safe to repair automatically.

## Scope

- Inventory every `FileProblem` and semantic lint `Issue` producer and assign each emitted defect a
  stable kind plus repair applicability.
- Put the vocabulary in the domain contract and derive or drift-test fix coverage from it; agents must
  not parse `message` to choose `lint --fix`, guarded graph repair, or manual intervention.
- Carry both fields through human diagnostics and the `lint --json` envelope under an explicit
  additive machine-contract revision.
- Build on adapter-neutral failed-record identities and the corrected unreadable-record collision
  rule so the final inventory covers both load and semantic diagnostics once.

## Acceptance criteria

- [ ] Every emitted lint defect has a non-empty stable kind from one closed, documented vocabulary.
- [ ] Every kind declares repair applicability, and a drift test fails when a new kind has neither a
      fix path nor an explicit manual/refused disposition.
- [ ] `lint --json` exposes kind and applicability additively with updated schema, fixtures, and
      compatibility notes; existing path, field, severity, and message evidence remains available.
- [ ] Human output stays explanatory and does not expose implementation-only enum spellings as a
      substitute for useful diagnostics.
- [ ] `lint --fix` and `task depend repair` routing can be selected from structured fields without
      matching English text.
- [ ] Filesystem and future pathless adapters produce the same diagnostic kinds for the same defect.

## Out of scope

- Expanding what `lint --fix` is authorized to modify.
- Folding guarded dependency repair into the generic fix path.
- Defining the entire frontmatter field registry or changing lint severity policy.
- Removing or freezing human-readable messages.

## Related

- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md)
- Audit [2026-09-21 architecture: failure and recovery](../audits/6gc7jd9aq1q9-2026-09-21-arch-failure-and-recovery.md), finding M2
- Policy gate [frontmatter-schema ADR questions](6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md)
- Diagnostic predecessor [adapter-neutral lint load diagnostics](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
- Rule predecessor [unreadable cross-kind ID collisions](6gcqz5aefjjf-lint-cross-kind-task-and-thread-id-collisions-on-unreadable-records.md)
