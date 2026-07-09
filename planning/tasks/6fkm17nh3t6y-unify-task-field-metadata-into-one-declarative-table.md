---
schema: 1
id: 6fkm17nh3t6y
status: completed
epic: 26-frontmatter-schema-declared-validation-contract
description: Derive knownTaskFields + int/list/date maps from one declarative task-field table (not parallel hand-kept maps); add a Task-struct-tag sync test. Behavior-preserving groundwork for the field registry.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [schema, validation, core]
created: "2026-07-07"
started_at: "2026-07-07"
updated_at: "2026-07-07"
completed_at: "2026-07-07"
---
# Unify task field metadata into one declarative table

First implementation slice of epic 26's Direction #1 ("one declarative field
registry as the single source of truth"). Behavior-preserving; needs none of the
open policy decisions, so it can land ahead of the design ADR.

## Objective

Today the task field metadata is split across parallel, hand-kept maps that must
stay in lockstep by eye: `knownTaskFields` (fields.go), `intFields`/`listFields`
(fields.go), and `dateFields` (validate.go). Add a field and you must remember to
touch every one. Collapse them into a single declarative `taskFields` table
(name → YAML type) that all four maps derive from, so they cannot drift.

## Acceptance criteria

- [ ] `knownTaskFields`, `intFields`, `listFields`, `dateFields` are all derived
      from one `taskFields` table; no parallel hand-kept literals remain.
- [ ] Accessors (`KnownTaskField`, `IsIntField`, `IsListField`) and `FieldType`
      are unchanged in behavior; existing tests stay green.
- [ ] New sync test: every `domain.Task` frontmatter yaml tag except `id` is a
      known task field (closes the "keep in sync with the Task yaml tags" gap
      that was previously eyeballed).

## Out of scope

- The open policy questions (per-status strictness, severities, rollout) — those
  wait on the ADR.
- Folding LintTask's hand-written checks or the `--json` envelope schema into the
  registry — later slices.

## Related

- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md)
- Prior art: the `entity.go` descriptor registry, `schema.go` FieldType.
