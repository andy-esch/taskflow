---
schema: 1
id: 6g81npee5kv2
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Give task rename JSON success and dry-run output the same structured planning and cascade receipt available on failures.
effort: 0.5-1 day
tier: 3
priority: medium
autonomy_level: 3
tags: [rename, cli, json, wire]
created: "2026-09-08"
depends_on: [6g7wxs43g7nh]
updated_at: "2026-09-08"
started_at: "2026-09-08"
completed_at: "2026-09-08"
---
# Expose successful task rename receipts in JSON

## Objective

Emit a rename-specific success and dry-run envelope so agents can observe source and destination slugs, planned documents, and rewritten links without inferring them from human output or the generic task mutation shape.

## Acceptance criteria

- [x] Successful `task rename --json` reports the adapter-neutral rename receipt, including source/destination identity and planned/applied document and link counts.
- [x] Dry-run reports the same prospective plan while clearly remaining uncommitted and incomplete.
- [x] The success, dry-run, and existing post-commit failure receipts use consistent field meanings.
- [x] The additive wire contract is versioned, documented, schema-validated with non-default values, and covered by CLI goldens.
- [x] Human output remains concise and behavior-compatible.

## Out of scope

- Changing guarded rename ordering or recovery semantics.
- Adding task rename to the TUI or web adapter.

## Evidence

- Tracked from finding L2 in the [Claude task rename implementation audit](../audits/6g81f73jh6d9-2026-09-08-task-rename-snapshot-and-recovery-implementation-claude.md).
- Delivered by [PR #215](https://github.com/andy-esch/taskflow/pull/215).

## Implementation evidence (2026-09-08)

`task rename --json` now emits a dedicated, presentation-independent wire envelope for successful writes and dry-run previews. It preserves the prior `task`, `dry_run`, and `workspace` fields while adding stable source/destination identity, planned/applied document and link counts, and explicit durability state. The same mapper feeds post-commit failure recovery, so field meanings cannot drift between outcomes.

Wire schema 1.63 registers and documents the additive envelope. A byte-level dry-run golden includes a non-empty inbound-link plan; CLI integration tests compare preview and committed receipts, verify disk effects, reject failure-only remedy prose on success, and pin unchanged concise human output. Schema round-trip validation uses a fully populated committed receipt.

Validation: the full race-enabled test suite, golangci-lint, generated CLI-doc drift check, module tidy check, planning lint, schema-comment freshness, and JSON goldens pass.
