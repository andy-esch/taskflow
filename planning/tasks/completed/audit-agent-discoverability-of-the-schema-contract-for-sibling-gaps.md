---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: 'Review found epic_fields missing from schema --json + a brittle envelope count-guard (both fixed). Sweep for siblings: audit-field discoverability, other magic-count guards, missing no-drift tests.'
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-28"
started_at: "2026-06-26"
completed_at: "2026-06-28"
id: 6ffr4wc016wc
---
## Objective

The 2026-06-25 adversarial review found two agent-discoverability/contract gaps (epic_fields absent from `schema --json`; a literal-count envelope-validation guard that silently skipped envelopes) — both fixed. Sweep for siblings.

## Acceptance criteria
- [x] `schema --json` exposes every gate an agent needs (audit-settable fields, finding statuses), not just task/epic fields.
- [x] Grep for other brittle 'magic count' / literal-list guards that can silently drift.
- [x] Every authoring-field registry has a no-drift test (tasks + epics do; audits?).

## Notes
Hardening, not urgent — the review closed the worst cases.

**Status check 2026-06-28.** The two named findings (epic_fields missing from `schema --json`; the brittle envelope count-guard) are fixed — `schema audit` exists and `schema --json-schema` covers the envelopes. Open work is the sibling SWEEP only: audit-field discoverability in the JSON schema, other magic-count guards, and the missing no-drift tests.

## Completed 2026-06-28

**AC1 — finding-status discoverability.** Added `finding_statuses` to the `schema`
contract (`wire.SchemaContract` → `runSchemaContract` feeds `domain.FindingStatuses()`;
`SchemaHuman` prints it under "Audit buckets"). This was the real gate gap — an agent
writing a finding had no machine list of the legal status vocabulary (open · in-progress
· fixed · landed · deferred · superseded · wontfix). schema_version 1.20→1.21; goldens +
`schema_jsonschema.golden` regenerated; round-trip test case updated. (No `audit_fields`
added: audits have **no settable fields** — area/date are immutable identity, there is no
`audit set` — and audit *authoring* fields are already covered by `schema audit`, so a
top-level list would just duplicate it.)

**AC2 — magic-count / literal-list sweep.** Swept for the silent-drift class (literal
counts of enums/registries, hardcoded enum lists). The envelope count-guard was the only
registry *count* guard and is already registry-derived; the codebase otherwise iterates
registries (AllStatuses/SchemaKinds/Descriptors) rather than hardcoding counts. Found ONE
real drift risk: `render.findingStatusOrder` (the `audit show` lifecycle grouping) is a
hand-ordered literal of the finding vocab with no tie to the registry — a status added to
the registry but not there would silently misgroup. Pinned it with
`TestFindingStatusOrder_CoversRegistry` (same-SET guard; order stays hand-curated).

**AC3 — audit no-drift test.** Added `TestAuditAuthoringFieldsMatchStruct`
(domain/schema_test.go), mirroring the task/epic guards. Audits have no settable-field
map, so the registry it drifts against is the `Audit` struct's yaml tags: every
documented audit authoring field must be a real persisted field. (Type isn't compared to
`FieldType` — that's a task-field utility; the epic guard skips it for the same reason.)

build/vet/test/lint green; docs/cli + schema goldens regenerated.
