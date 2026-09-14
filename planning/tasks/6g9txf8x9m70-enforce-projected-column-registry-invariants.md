---
schema: 1
id: 6g9txf8x9m70
status: completed
epic: 20-cli-ux-and-ergonomics
description: Make list-column registries reject selector collisions and canonical key/value drift without requiring DTO parity.
effort: 1 day
tier: 2
priority: high
autonomy_level: 4
tags: [cli, json, contract, testing]
created: "2026-09-13"
audit_sources: [planning/audits/6g9samx90yrd-2026-09-13-projected-json-and-cli-contract-fidelity-implementation-claude.md, planning/audits/6g9z5wnzksth-2026-09-14-projected-column-registry-invariants-implementation-claude.md]
updated_at: "2026-09-14"
depends_on: [6g9s401jtqb7]
started_at: "2026-09-14"
completed_at: "2026-09-14"
---
# Enforce projected-column registry invariants

## Objective

Make the shared list-column model reject contradictory registry declarations before they can ship a
header/value lie, selector collision, or projection drift. Preserve the intentionally curated list
surfaces and explicit compatibility aliases; this is an invariant pass, not DTO/column parity or a
reflection-driven registry redesign.

## Acceptance criteria

- [x] Every task, research, audit, epic, and finding registry is checked for duplicate or colliding
      canonical selectors and aliases before selection/completion can silently shadow an entry.
- [x] Contract-column construction makes the canonical selector and raw projector an inseparable
      pair, or an equivalent executable guard rejects one without the other.
- [x] Table-driven fixtures compare canonical table/CSV and projected-JSON values with the matching
      full-envelope datum for present, absent, and zero values where a list column represents a wire
      field; documented synthetic and string-view exceptions remain explicit.
- [x] Focused mutation probes demonstrate that selector collision, a missing raw projector, and a
      recreated display fallback under a canonical key each fail for the intended reason.
- [x] Focused and full race tests, lint, generated-artifact checks, planning lint, and
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

## Implementation outcome (2026-09-14)

The shared `Column` model now carries a canonical selector and raw projector as one private projection value. Official task, epic, audit, research, and finding registries validate as they are constructed; `Specs` repeats the fail-fast check before help/completion is derived, while `SelectColumns` returns a diagnostic before building a selector map that could shadow an earlier entry.

Registry-wide fixtures compare every canonical table, CSV, and projected-JSON cell with the actual full JSON envelope for populated, absent, zero, and created-but-never-edited records. The only non-scalar exceptions are explicit and exercised: research tags remain a comma-joined string view, and finding `ref` remains the synthetic `audit:code` address.

Three restored mutation probes failed for the intended reasons: a duplicate epic `id` selector panicked during registry construction; a task contract column without its raw projector panicked at construction; and removing the canonical extractor switch made the full-wire fidelity test expose fabricated `updated_at` values for never-edited task and research records. The focused CLI/render/wire suites, full race suite, golangci-lint, module tidiness, independently regenerated CLI docs and schema comments, planning lint, and `git diff --check` all pass.

## Adversarial review closeout (2026-09-14)

Claude’s four findings were accepted and fixed. The fidelity harness now requires each canonical
list column to reach the real full-wire envelope in at least one fixture or declare an explicit
wire-derived exception; research tags and finding references are validated from encoded wire
fields, so coordinated mapper drift no longer passes. Invalid registries are exercised through
empty, unknown, and declared selections. Self-named contract columns intentionally remain
supported, but explicit canonical selection now swaps to the raw projector. A named test records
the pre-existing lint-invalid zero-tier divergence without changing schema 1.66 behavior.

The reviewer’s hostile mutations and additional local mutations fail for the intended reasons.
Final validation passes: focused CLI/render/wire tests, `go test -race ./...`, golangci-lint,
`go mod tidy -diff`, independently regenerated CLI docs and schema comments, planning lint, and
`git diff --check`. Audit
`2026-09-14-projected-column-registry-invariants-implementation-claude` is closed with M1 and
L1–L3 fixed.
