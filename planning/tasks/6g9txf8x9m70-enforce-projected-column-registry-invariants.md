---
schema: 1
id: 6g9txf8x9m70
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Make list-column registries reject selector collisions and canonical key/value drift without requiring DTO parity.
effort: 1 day
tier: 2
priority: high
autonomy_level: 4
tags: [cli, json, contract, testing]
created: "2026-09-13"
audit_sources: [planning/audits/6g9samx90yrd-2026-09-13-projected-json-and-cli-contract-fidelity-implementation-claude.md]
updated_at: "2026-09-13"
depends_on: [6g9s401jtqb7]
---
# Enforce projected-column registry invariants

## Objective

Make the shared list-column model reject contradictory registry declarations before they can ship a
header/value lie, selector collision, or projection drift. Preserve the intentionally curated list
surfaces and explicit compatibility aliases; this is an invariant pass, not DTO/column parity or a
reflection-driven registry redesign.

## Acceptance criteria

- [ ] Every task, research, audit, epic, and finding registry is checked for duplicate or colliding
      canonical selectors and aliases before selection/completion can silently shadow an entry.
- [ ] Contract-column construction makes the canonical selector and raw projector an inseparable
      pair, or an equivalent executable guard rejects one without the other.
- [ ] Table-driven fixtures compare canonical table/CSV and projected-JSON values with the matching
      full-envelope datum for present, absent, and zero values where a list column represents a wire
      field; documented synthetic and string-view exceptions remain explicit.
- [ ] Focused mutation probes demonstrate that selector collision, a missing raw projector, and a
      recreated display fallback under a canonical key each fail for the intended reason.
- [ ] Focused and full race tests, lint, generated-artifact checks, planning lint, and
      `git diff --check` pass.

## Scope

- Keep list registries curated for triage; do not require every DTO field to become a column.
- Prefer a small constructor/validation boundary and registry-wide tests over reflection or parallel
  sources of truth.
- Retain the schema-1.66 legacy-key and raw-value compatibility decisions established by
  `restore-projected-json-and-cli-contract-fidelity`.

## Related

- Task [Restore projected JSON and CLI contract fidelity](6g9s401jtqb7-restore-projected-json-and-cli-contract-fidelity.md)
- Audit [Projected JSON and CLI contract fidelity implementation — Claude](../audits/6g9samx90yrd-2026-09-13-projected-json-and-cli-contract-fidelity-implementation-claude.md), finding L4
- Thread [Harden CLI machine contracts](../threads/6g9txm0s3yyh-harden-cli-machine-contracts.md)
