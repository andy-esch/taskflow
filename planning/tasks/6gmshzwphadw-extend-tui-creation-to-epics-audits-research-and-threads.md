---
schema: 1
id: 6gmshzwphadw
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Reuse the proven staged task-creation interaction to create the remaining four planning entity kinds without pretending their required inputs or mutation policies are identical.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gp1ytfvm652, 6gp1ytg50tev, 6gp1ytgea8dt]
updated_at: "2026-10-24"
started_at: "2026-10-24"
completed_at: "2026-10-24"
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
3. [Guarded Thread creation](6gp1ytgea8dt-create-threads-from-the-tui-with-guarded-per-kind-intent.md), with its own graph receipts and empty, unstarted membership; member selection and sequencing are separate workflows.

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

### Delivery checkpoint — 2026-10-24

The empty-Thread slice now supplies the fifth concrete form, with guarded
creation, independent durable/portable regressions and two closed reviews linked
from its task. It satisfies the empty/unstarted/no-task-effects criterion below.
Membership and sequencing remain separate design work. This integration task
stays pending for the combined creation walkthrough and remaining criteria; the
individual delivery slices do not substitute for that finish-line check.

### Integrated verification — 2026-10-24

Validated the merged creation implementation at `b560e20c` in one real
pseudo-terminal session (140×40), using page-local `c`, keyboard navigation and
bracketed terminal paste. A disposable tree contained an unrelated epic/task;
`TSKFLW_CONFIG_HOME` was isolated and initialization used `--no-register`.
All four creations were performed in the TUI, not through CLI creation verbs:

| Created record | Observed per-kind contract |
| --- | --- |
| Touring bike workshop | Auto-numbered epic; explicit active/medium choices, description and tags. |
| Touring release readiness | Open audit; blank date became 2026-10-24; zero parsed findings. |
| Touring tyre comparison 世界 | Research backdated to 2026-10-01, with description/tags and no lifecycle. |
| Release the touring bike | Empty, unstarted Thread; required description/goal, target date and tags; no task/dependency inputs. |

Each success selected its new record and loaded its correct scaffold. A missing
Thread goal produced an editable refusal with no Thread on disk; correcting the
retained draft created exactly one. A second draft survived Keep editing, then
Discard left the original Thread selected; refresh retained that selection.
Independent CLI JSON assertions checked exact counts, metadata and healthy Thread
projections after the walkthrough. The preexisting task, epic and repository
configuration retained their SHA-256 hashes and permissions.

The combined race-enabled creation regressions passed across all five consumers.
These cover duplicate/stale delivery, known refusals, committed/uncertain recovery,
watcher landing, portable capabilities, workspace-origin isolation, narrow modes
and native/delayed paste. See [shared host tests](../../internal/tui/creation_host_test.go),
[epic tests](../../internal/tui/epic_create_test.go),
[audit/research tests](../../internal/tui/document_create_test.go), and
[Thread tests](../../internal/tui/thread_create_test.go). Receipt fault injection
is regression evidence, not a claim that cleanup failures or workspace switches
were forced in the terminal walkthrough. Typed `New*Params`, defaults and receipt
policies remain per-kind; no persisted-field metadata became a form schema.

This check also corrected stale `ui --help`/generated reference text and incomplete
CLI-only empty-page advice for tasks/Threads. Focused tests pin the native empty-page
action and help boundary. Membership, sequencing and lifecycle authoring are not
silently included. The broader [TUI-only workflow closeout](6gmshzy2ppfb-prove-routine-tui-planning-without-cli-round-trips.md)
still owns owner feedback and the remaining authoring/workspace capabilities.

The planning dogfood pass also reproduced a preexisting
[criterion-section shadowing defect](6gpxh6ncaqpw-prevent-narrative-headings-from-hiding-task-criteria.md):
a narrative heading mentioning acceptance before the real section hides its
criteria from reads/writes and permits completion preview despite an unmet item;
lint does not diagnose it. The control's neutral heading restores 0/1 and the
completion refusal. This document uses a neutral verification heading; the parser
fix is tracked separately, in this Thread and ahead of criterion UI work.

Validation: `just build`, full `just test` (race), `just lint` and the combined
creation regression command passed. Focused native-empty-page and help tests
passed. `just docs` is repeatably idempotent; `just docs-check` passed after the
generated reference was committed. Its earlier failure was only the intended
uncommitted `tskflwctl_ui.md` update. No CLI wire shape or persisted format changed
in this closeout.

### Repeat the bounded smoke check

After `just build`, run in a subshell so the temporary home override cannot leak:

```sh
(
  creation_dir=$(mktemp -d "${TMPDIR:-/tmp}/taskflow-creation.XXXXXX")
  export TSKFLW_CONFIG_HOME="$creation_dir/home"
  ./bin/tskflwctl -C "$creation_dir" init --no-register
  ./bin/tskflwctl -C "$creation_dir" ui
)
```

Use `[ ]` to visit Epics, Audits, Research and Threads; `c` opens each form.
Create one of each, leave the audit date blank, backdate research, and deliberately
omit the Thread goal once before correcting it. Use Tab/Shift+Tab and explicit
Create; inspect loaded detail, not just the success flash. Open another draft to
exercise Keep/Discard, then refresh. The setup is CLI-assisted; creation is not.
Use the per-kind `list --json` commands against the printed tree path afterward
to verify counts/defaults. This is a bounded correctness check, not new human
usability sign-off or the full authoring Thread's end-to-end acceptance.

## Acceptance criteria

- [x] Create an epic, audit, research document and unstarted Thread from their respective TUI contexts through core services.
- [x] Each kind collects its actual creation inputs and defaults without treating persisted-field Required metadata as a generic creation form specification.
- [x] Thread creation produces an empty, unstarted record without selecting or
  creating tasks, changing dependencies or starting work; membership and
  sequencing belong to separately scoped interactions.
- [x] One explicit submission creates one record; duplicate submission, validation refusal, committed cleanup failure and watcher reload cannot create duplicates or steal selection.
- [x] Forms preserve input and origin-space binding, open/select the created record by stable identity, and pass focused tests, normal code gates and updated help/architecture guidance.
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
  selection is excluded from this creation slice by the owner's scope decision.
  Later membership curation needs an exact selected-endpoint contract; ordinary
  reference matching is not automatically that contract. A task receipt cannot
  model graph outcomes.

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
