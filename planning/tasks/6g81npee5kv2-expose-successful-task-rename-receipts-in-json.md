---
schema: 1
id: 6g81npee5kv2
status: ready-to-start
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
---
# Expose successful task rename receipts in JSON

## Objective

Emit a rename-specific success and dry-run envelope so agents can observe source and destination slugs, planned documents, and rewritten links without inferring them from human output or the generic task mutation shape.

## Acceptance criteria

- [ ] Successful `task rename --json` reports the adapter-neutral rename receipt, including source/destination identity and planned/applied document and link counts.
- [ ] Dry-run reports the same prospective plan while clearly remaining uncommitted and incomplete.
- [ ] The success, dry-run, and existing post-commit failure receipts use consistent field meanings.
- [ ] The additive wire contract is versioned, documented, schema-validated with non-default values, and covered by CLI goldens.
- [ ] Human output remains concise and behavior-compatible.

## Out of scope

- Changing guarded rename ordering or recovery semantics.
- Adding task rename to the TUI or web adapter.

## Evidence

- Tracked from finding L2 in the [Claude task rename implementation audit](../audits/6g81f73jh6d9-2026-09-08-task-rename-snapshot-and-recovery-implementation-claude.md).
