---
schema: 1
id: 6gp1ytg50tev
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Exercise the shared creation host with audit area/date and research title/date semantics, typed receipts and origin-bound recovery.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-21"
depends_on: [6gpb4y5yqmjp]
updated_at: "2026-10-22"
---
# Create audits and research from shared TUI authoring forms

## Objective

Create audits and research on their respective TUI pages using the shared staging host proven by tasks and epics, without transplanting task-parent or epic lifecycle assumptions.

## Acceptance criteria

- [ ] Page-local c and named palette actions expose audit and research creation using typed NewAuditParams and NewResearchParams through captured services in tea.Cmd.
- [ ] Audit collects area/date and keeps core open-bucket/default-date policy; research collects title/description/tags/created date with no lifecycle and leaves collision retry to core.
- [ ] Shared mechanics are tested across tasks, epics and these materially different forms; typed receipts, capability/default policy and exact destination selection remain per kind.
- [ ] Validation, absent/read-only/pathless capabilities, pending/unknown/committed outcomes, session drift, stale widget results and watcher landing retain input and never replay or retarget writes.
- [ ] Independent durable metadata/scaffold expectations, real keyboard/paste flows, narrow rendering and normal code/planning gates pass with owning docs updated.

## Out of scope

Thread graph/membership creation, universal creation API, bulk/template workflows and long Markdown editing.

## Related

- [Additional creation integration](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md).
- [TUI architecture](../../docs/architecture/tui.md#staged-task-creation).
