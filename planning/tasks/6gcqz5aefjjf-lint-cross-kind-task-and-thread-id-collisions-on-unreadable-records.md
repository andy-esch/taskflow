---
schema: 1
id: 6gcqz5aefjjf
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Keep cross-kind identity collisions visible in lint when either task or Thread source is malformed.
effort: S
tier: 3
priority: medium
autonomy_level: 4
tags: [lint, diagnostics, identity]
created: "2026-09-22"
audit_sources: [planning/audits/6gch6xep2mh2-2026-09-22-correctness-and-errors.md]
depends_on: [6g5vm4efjcdv]
updated_at: "2026-09-22"
---

# Lint cross-kind task and Thread ID collisions on unreadable records

## Objective

Make the planning-space-wide task/Thread ID invariant fail closed in repository lint. When either
document is malformed, consume the stable identity already recovered by the read boundary instead of
silently dropping the cross-kind collision from the report.

## Scope

- Seed task and Thread identity sets from safe recovered IDs on unreadable records as well as decoded
  entities; do not infer identity from arbitrary prose or an invalid filename.
- Preserve the existing create-time guards and duplicate-within-kind checks while making the lint
  rule symmetric for readable/unreadable task and Thread pairs.
- Land after the adapter-neutral lint-load diagnostic shape so the rule consumes one portable
  identity contract rather than adding another filesystem-specific path dependency.

## Acceptance criteria

- [ ] A readable task colliding with an unreadable Thread remains reported, and the inverse case is
      covered by a regression test.
- [ ] Two unreadable records with the same safely recovered cross-kind ID are diagnosed when both
      read boundaries provide authoritative identity.
- [ ] Missing, malformed, or untrusted recovered IDs do not create false collision reports.
- [ ] Ordering and de-duplication remain deterministic in human and `--json` lint output.
- [ ] Creation guards and readable-record collision behavior remain unchanged.

## Out of scope

- Redesigning the lint diagnostic vocabulary or `lint --fix` applicability.
- Expanding stable-ID uniqueness beyond the currently decided task/Thread invariant.
- Repairing malformed records or changing guarded create behavior.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-22 correctness and errors](../audits/6gch6xep2mh2-2026-09-22-correctness-and-errors.md), finding M1
- Predecessor [adapter-neutral lint load diagnostics](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
