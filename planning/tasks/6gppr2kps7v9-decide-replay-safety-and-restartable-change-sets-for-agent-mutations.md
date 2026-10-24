---
schema: 1
id: 6gppr2kps7v9
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'Design-first: decide whether non-idempotent mutations get replay identities after a lost acknowledgement, and keep or drop a restartable multi-operation change-set apply.'
effort: 1 day
tier: 3
priority: medium
autonomy_level: 3
tags: [agents, reliability, design]
created: "2026-10-23"
audit_sources: [planning/audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md]
updated_at: "2026-10-23"
---
# Decide replay safety and restartable change sets for agent mutations

## Objective

OCC and atomic renames make in-process retries safe, but an agent transport can lose the
acknowledgement after a committed append, and the retry duplicates durable content. A multi-command
closeout interrupted midway leaves state valid per file but semantically incomplete. Both findings
need decisions before code: whether a local-first CLI should persist any dedupe records, and
whether a change-set apply fits a git-agnostic tool at all.

## Acceptance criteria

- [ ] A recorded decision (research doc or ADR) on replay protection for appends and other non-idempotent mutations: operation keys with bounded receipts, a narrower if-absent content check, or explicitly none, with the trade-offs named.
- [ ] A recorded keep-or-drop decision on a restartable change-set `apply` (preflight, per-operation IDs, applied/no-op/failed receipt), weighing reuse of the Thread compose/apply plan machinery.
- [ ] Any accepted direction is split into scoped implementation tasks, and any rejection is recorded on the findings.

## Out of scope

- Implementing either mechanism in this task.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Audit [2026-07-24 AI-agent CLI ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md), findings H3 and M6
