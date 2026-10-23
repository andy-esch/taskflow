---
schema: 1
id: 6gmshzwphadw
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Reuse the proven staged task-creation interaction to create the remaining four planning entity kinds without pretending their required inputs or mutation policies are identical.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gp1ytfvm652, 6gp1ytg50tev, 6gp1ytgea8dt]
updated_at: "2026-10-23"
---
# Extend TUI creation to epics audits research and Threads

## Objective

Reuse the proven staged task-creation interaction to create the remaining four planning entity kinds without pretending their required inputs or mutation policies are identical.

## Scope

Add entity-specific form definitions and contextual create actions for epics, audits, research and Threads. Use existing New* use cases, their defaults and default body scaffolds; long bodies remain editor-assisted. Threads have a real tab and are explicitly included.

### Reviewable delivery slices (2026-10-21)

This task is the integration/finish-line owner, not one four-kind implementation PR:

1. [Epic creation and shared mechanics](6gp1ytfvm652-create-epics-in-the-tui-with-shared-staged-form-mechanics.md), reusing the task form without sharing entity policy.
2. [Audit and research forms](6gp1ytg50tev-create-audits-and-research-from-shared-tui-authoring-forms.md), exercising different input/default contracts.
3. [Guarded Thread creation](6gp1ytgea8dt-create-threads-from-the-tui-with-guarded-per-kind-intent.md), with its own graph receipts and selected-member design checkpoint.

Both later slices depend on the proven common host, not on each other. This task
depends on all three and verifies the combined criteria below before the final
TUI-only walkthrough. The parent estimate is historical, not a single-PR budget.

After the epic slice landed, the owner approved a bounded
[host and interaction-contract hardening step](6gpb4y5yqmjp-harden-the-shared-tui-creation-host-and-interaction-contracts.md)
before either remaining creation slice. It consolidates active-draft/result/session
wiring, finite widget testing and layout visibility bounds, not entity policy.

### Delivery checkpoint — 2026-10-23

Task, epic, audit and research forms now exercise the same host, staging and input
mechanics with separate typed requests/receipts and independent durable tests.
The audit/research slice completed both independent reviews with no findings;
its task links the evidence. This satisfies shared-reuse criterion 6 only.
Thread creation and the combined four-kind criteria remain outstanding; this
integration task stays pending until guarded Thread creation lands and the full
finish-line checks run.

## Acceptance criteria

- [ ] Create an epic, audit, research document and unstarted Thread from their respective TUI contexts through core services.
- [ ] Each kind collects its actual creation inputs and defaults without treating persisted-field Required metadata as a generic creation form specification.
- [ ] Thread initial-member selection, if offered, uses guarded creation and exact intended task identities; creating a Thread does not silently create dependencies or start it.
- [ ] One explicit submission creates one record; duplicate submission, validation refusal, committed cleanup failure and watcher reload cannot create duplicates or steal selection.
- [ ] Forms preserve input and origin-space binding, open/select the created record by stable identity, and pass focused tests, normal code gates and updated help/architecture guidance.
- [x] Shared staging, navigation and asynchronous result mechanics are exercised
  by the existing task form and at least two materially different entity forms,
  while typed requests, capabilities, defaults and receipt/recovery semantics
  remain per kind.

## Design attention

Reuse widgets and staging mechanics, not a universal CreateEntity port. Add a bounded core seam if a real interface contract is missing rather than promising this is adapter-only.

Start with a second real entity form and extract common TUI mechanics against it
and task creation under tests. Do not copy the whole task controller per kind or
turn its private row constants into a generic entity-field schema. Reuse the host
mechanics (focus, bounded layout, dirty confirmation, input/result routing and
pending/displaced intent handling); keep typed parameter conversion, capability
binding, receipt interpretation and destination selection per kind. Route page-local
creation and named palette actions through one per-kind opener as the registry grows.

Review the actual use cases at kickoff rather than transplant task defaults:

- `NewEpicParams` requires explicit status/priority and a description; its ID is
  allocated by the adapter, not minted as a task ID by the form.
- `NewAuditParams` collects area/date and creates an open audit, not title/epic/tags.
- `NewResearchParams` supports a created date tied to ID minting and has no lifecycle;
  its bounded collision handling stays in core, never a UI replay loop.
- `NewThreadParams` requires description/goal and guarded creation. Optional member
  selection needs an exact selected-endpoint contract; ordinary reference matching
  is not automatically that contract. A task receipt cannot model graph outcomes.

Regression expectations must be independent of shared preparation helpers. Prove
reuse with at least two materially different form definitions and the existing task
form, including absent/pathless capabilities and committed/uncertain outcomes. Split
delivery into reviewable per-kind slices if the combined implementation grows large.
Include live terminal paste as well as keys and delayed clipboard results when
extracting the host; the [first-form Codex review](../audits/6gnqmm1qp9mz-2026-10-20-tui-staged-task-creation-implementation-codex.md)
caught a missing runtime event class that direct input setters and key-only
program tests did not exercise.

## Out of scope

Bulk import, template-library design, long Markdown editing, implicit cross-entity creation chains and bootstrap.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning service_epic.go](../../internal/core/service_epic.go).
- [Owning service_audit.go](../../internal/core/service_audit.go).
- [Owning service_research.go](../../internal/core/service_research.go).
- [Owning service_thread.go](../../internal/core/service_thread.go).
