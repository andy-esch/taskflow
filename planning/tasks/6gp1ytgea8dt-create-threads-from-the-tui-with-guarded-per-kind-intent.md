---
schema: 1
id: 6gp1ytgea8dt
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Add staged Thread creation using typed graph receipts and explicit selected-member safety if initial membership is offered.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-21"
depends_on: [6gpb4y5yqmjp]
updated_at: "2026-10-22"
---
# Create Threads from the TUI with guarded per-kind intent

## Objective

Create an unstarted Thread from its own TUI page using shared staged-form mechanics and the actual guarded Thread creation contract.

## Acceptance criteria

- [ ] Threads-page c and New Thread palette action stage title, description, goal and appropriate optional metadata before one explicit submission through the originating Service.
- [ ] Typed Thread creation/graph receipts and failures retain committed/partial intent without UI retry loops, duplicate writes or invented task lifecycle semantics.
- [ ] If offering initial membership, add or consume an exact selected-member capability revalidated inside guarded creation; passing ID-looking strings through ordinary reference matching is insufficient.
- [ ] Resolve the initial-membership interaction at kickoff with the owner if needed; an empty unstarted Thread must not silently create dependencies or start work.
- [ ] Portable and real guarded-adapter tests cover identity/capability drift, refusals, committed/uncertain results, origin switches, watcher landing and keyboard/paste/narrow modes with normal gates and docs.

## Out of scope

Dependency editing, bulk compose/apply, graph layout and automatic task creation.

## Related

- [Additional creation integration](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md).
- [Thread creation use case](../../internal/core/service_thread.go).
- [TUI architecture](../../docs/architecture/tui.md#staged-task-creation).
