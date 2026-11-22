---
schema: 1
id: 6gppr2j8g1h4
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: epic set --unset description writes a lint-invalid epic; refuse required-field unsets and give the unset path and NewEpic table negative tests.
effort: 1-2 hours
tier: 2
priority: high
autonomy_level: 3
tags: [core, validation, testing]
created: "2026-10-23"
audit_sources: [planning/audits/6g8zy840z6gj-2026-09-11-test-rigour.md, 2026-11-22-weekly-task-sweep]
updated_at: "2026-11-22"
audited: "2026-11-22"
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

Cross-referenced by audit 2026-11-17-adapter-hygiene: M1 (partial overlap, left open). M1 reports that NewEpic defaults nothing while prepareTaskCreation (internal/core/service_task.go:966-982) defaults the task analog and states that rule explicitly, so both adapters replicate epic status/priority. Verified: `epic new --priority ""` exits 11 where the equivalent `task new` exits 0. This task does not own that change, but its third AC is already rewriting TestService_NewEpic's failure table — the natural home for a blank-input defaulting test. Scoping call for a human; no status or scope changed here.

## Sweep verification (2026-11-22)

First automated weekly sweep for this task (no prior `audited:`). Every claim was
re-checked against `main` (`6abe9de`). `internal/` has not changed since
2026-10-25, so the code this task describes is the code it was written against.

**All three criteria are still unmet, and all three premises still hold:**

- **AC 1 — verified accurate (2026-11-22).** `prepareEpicFieldUpdates`
  (`internal/core/service_epic.go:335`) rejects an unset of `status` and of
  `updated_at` by name, then falls through to `clean[field] = val` for any other
  known epic field. `description` is a known epic field, so
  `epic set <epic> --unset description` still reaches the store and still writes
  the exact state `lint` reports. The write path/lint contradiction the Objective
  names is unchanged.
- **AC 3 — verified accurate (2026-11-22).** `TestService_NewEpic`
  (`internal/core/usecases_test.go:503`) still asserts only `err == nil` on its
  four-case failure table, so no case asserts `errors.Is(err, domain.ErrValidation)`.
  None of the four cases sets `Status` at all, so each may be tripping the status
  predicate rather than the one its comment names — exactly the defect AC 3
  describes. There is still no explicit invalid-status case.

**One clause of AC 2 is already satisfied — narrowing, not completion.** AC 2 ends
"`TestService_SetEpicFields_UnknownFieldNeedsForce` covers both branches".
That test (`internal/core/service_epic_test.go:942-951`) does now assert both
branches: rejection without `--force` and acceptance with it. What it does *not*
cover is the **unset** branch, which is what the rest of AC 2 asks for. Confirmed
by search: `domain.UnsetField` appears in no epic test anywhere in the tree — the
only occurrence in a test is `internal/core/selected_task_action_test.go:57`, on
the task path. So the epic unset path still has **zero** negative tests, and the
remaining AC 2 surface is the four negative cases plus the positive custom-field
unset with `--force`.

The box was **not** ticked: AC 2 is partially satisfied, not met, and which half
counts as done is the implementer's call rather than a sweep's.

**Cross-reference note re-verified.** The 2026-11-17 adapter-hygiene M1 citation
`internal/core/service_task.go:966-982` is accurate: `prepareTaskCreation` is
declared at `:965` and its priority/tier/autonomy/effort defaulting block with the
"Defaults for zero-valued fields live here" rationale runs `:966-981`. `NewEpic`
(`internal/core/service_epic.go:25`) still defaults nothing, so the asymmetry M1
reports is intact. No scoping call was made here.

## Progress Log

- 2026-11-22: automated weekly sweep — first audit; all three ACs re-verified unmet against `6abe9de`; AC 2's `UnknownFieldNeedsForce` clause is already satisfied while the epic unset path still has zero `UnsetField` tests; description tightened under 150 chars.
