---
schema: 1
id: 6gppr2j8g1h4
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: epic set --unset description writes a lint-invalid epic; reject required-field unsets and give the epic unset path and NewEpic failure table real negative tests.
effort: 1-2 hours
tier: 2
priority: high
autonomy_level: 3
tags: [core, validation, testing]
created: "2026-10-23"
audit_sources: [planning/audits/6g8zy840z6gj-2026-09-11-test-rigour.md]
updated_at: "2026-10-23"
---
# Refuse unsetting required epic fields and pin the epic unset path

## Objective

`SetEpicFields` lets `epic set --unset description` remove a field every epic requires, so the
write path creates exactly the state `lint` reports, contradicting "writes refuse to create it".
Refuse the unset and give the epic unset path the negative tests it has never had. On the same
surface, make `TestService_NewEpic`'s failure table assert the predicate each case names.

## Acceptance criteria

- [ ] `epic set <epic> --unset description` fails with `ErrValidation` and writes nothing; tasks keep their legal description-less unstarted state.
- [ ] The epic unset path has negative tests for `description`, `updated_at`, `status`, and an unknown field without `--force`, plus a positive custom-field unset with `--force`; `TestService_SetEpicFields_UnknownFieldNeedsForce` covers both branches.
- [ ] `TestService_NewEpic` cases assert `errors.Is(err, domain.ErrValidation)`, each case carries a valid status so it trips the predicate its comment names, and an explicit invalid-status case exists.

## Out of scope

- Declaring required fields once on the entity descriptor (epic 26).
- A full `epic set` / `task set` parity sweep beyond the required-field gap.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-11 test rigour](../audits/6g8zy840z6gj-2026-09-11-test-rigour.md), findings M1 and L2
- Precedent: [task-set follow-ups](6fbj87001d81-task-set-follow-ups-sentinels-unknown-keys-canonical-field-table.md)
