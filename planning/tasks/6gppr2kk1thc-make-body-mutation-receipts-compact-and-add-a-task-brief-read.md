---
schema: 1
id: 6gppr2kk1thc
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'Stop echoing full bodies from body-mutation --json by default (opt-in instead), and add a task brief: objective, unmet criteria, blockers, latest progress entry.'
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [agents, json, cli]
created: "2026-10-23"
audit_sources: [planning/audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md]
updated_at: "2026-10-23"
depends_on: [6fpnn6zk157b]
---
# Make body-mutation receipts compact and add a task brief read

## Objective

Two token-economy gaps from agent dogfooding: body mutations echo the whole resulting document (a
ten-character append returned 4,271 bytes), and resuming a task needs several reads plus prose
interpretation. Both are compact read-model changes, not new stored state.

## Acceptance criteria

- [ ] Body-mutation `--json` receipts default to a compact shape (workspace, id, slug, operation, dry_run, changed, updated_at, post-write version) with an explicit opt-in for the full body, and the default change is classified per ADR-0008.
- [ ] A composed task brief read returns the objective, unmet criteria with their states, blockers and dependencies, and the latest dated progress entry, derived from the existing section and criterion parsers.
- [ ] Both have JSON contract tests and generated CLI reference updates.

## Out of scope

- Persisted summaries, and the progress-log write verb itself.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Audit [2026-07-24 AI-agent CLI ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md), findings M2 and L2
- The brief's latest-progress entry needs the canonical progress-section shape from [task log](6fpnn6zk157b-task-log-append-a-dated-progress-log-entry.md)
