---
schema: 1
id: 6gmshnmnew7p
status: in-progress
description: Make routine planning possible without CLI round trips, with native actions, editor-assisted prose, and safe Atlas registry management.
goal: Create and maintain planning work and registered entry points through the TUI, using the external editor for prose and the same guarded application services.
created: "2026-10-17"
tags: [tui, authoring, spaces, ux]
tasks: [6ffdv9g01q40, 6g321mybqvfz, 6g3cp32y9149, 6g4ym5y2dkp8, 6gkkzvr92waq, 6gmbfxy96tpw, 6gmshzw7we0m, 6gmshzwfkj62, 6gmshzwphadw, 6gmshzwxxce3, 6gmshzx55n1y, 6gmshzxchdzc, 6gmshzxksjcn, 6gmshzxvgagj, 6gmshzy2ppfb, 6gmy4c0x6e31, 6gmy4c15aj30, 6gp1ytfvm652, 6gp1ytg50tev, 6gp1ytgea8dt, 6gp2qmse4prt, 6gpb4y5yqmjp, 6gpxh6ncaqpw]
updated_at: "2026-10-24"
started_at: "2026-10-18"
---
# Enable routine planning and workspace management in the TUI

## Owner-aligned scope

Routine planning should be achievable from the TUI, with $EDITOR for longer prose.
The CLI remains available for every supported workflow and is the initial home for
automation, bulk manifests, migrations, graph repair, exceptional force overrides,
bootstrap/link setup and destructive filesystem administration.

Atlas v1 registers an existing entry point or forgets one registration. Forgetting
never deletes planning files or a checkout; a grouped space may have several entry
points. Initialization/linking is not implied by registration.

Owner confirmed page-local staged creation: `c` creates the current page's entity
kind; `n`/`N` remain search navigation. This is not a universal creation wizard.
Atlas removal changes only the machine-wide registration, not repo configuration
or planning lifecycle; an already-open checkout remains usable.

## Capability boundary

| Capability | Intended interface |
| --- | --- |
| Task and other entity creation | Native forms with explicit submission and core defaults |
| Routine metadata, rename and lifecycle | Native typed actions, not raw frontmatter editing |
| Criteria states and reasoned dispositions | Native structured selection; simple authoring scope deferred to the criteria task's design checkpoint |
| Thread membership and task dependencies | Native guarded actions with distinct semantics and impact feedback |
| Audit finding dispositions | Native structured actions; long finding prose stays editor-assisted |
| Register/forget existing space entry points | Atlas-native registry management; no deletion |
| Long bodies and malformed local records | External-editor handoff with explicit local capabilities |
| Preferences | Existing configuration editor with visible user/repository scope |
| Bulk operations, repair, bootstrap and destructive administration | CLI-first; not prerequisites for this Thread |

## Shared implementation contract

Primary adapters call core application services in tea.Cmd. No CLI subprocesses,
direct store/filesystem writes in Update/View, or parallel TUI validation policy.
Retain explicit service/source-set binding and adapter-established selected identities.
Do not assume passing an ID-shaped string to a reference-based method is a strict
selected-action contract; add a bounded use-case seam if needed.

Adapter-neutral does not mean a universal CreateEntity/form-schema API. Keep
per-kind application use cases and consumer-owned ports, explicit mutation policy
and optional local capabilities. New selected-action seams require portable
(including pathless/read-only cases) and real guarded-adapter evidence. UI types,
filesystem representation and registry-file details do not enter portable contracts.

Creation collects fields before a single explicit create. Feedback may validate
early, but does not persist partial records. Preserve entered values on errors,
distinguish committed cleanup failures from failed creates, and never remint/replay
after progress. Show the target space and return to the exact created/edited record
without watcher reloads stealing focus.

Actions must be discoverable in context and help. Shared form/action conventions
are reusable mechanics, not a universal descriptor-driven mutation framework.
Persisted-field requiredness is not necessarily required user input at creation.

## Delivery and finish line

Start with the bounded interaction contract, then task creation and Atlas registry
management as independent implementation branches. Each now has an explicit core
prerequisite: selected-epic creation before its task form, and confirmed-registration
removal before Atlas management. Their relative delivery order is preference, not
a graph prerequisite. Additional entity creation follows the first staged task form.
Other actions and existing editor/scope-clarity work remain independent where their
own requirements permit.

Reuse the existing tracked-criterion write/lint alignment task before criterion
actions; do not invent a TUI-only reason policy. The owner-deferred authoring choice
is still made at the criteria task's kickoff. The existing watcher/session fix is
parallel hardening and must finish before the final cross-workspace walkthrough,
not an invented blocker on every UI ticket.

Close after the scoped capability tickets and a disposable-space TUI-only routine
planning walkthrough pass. Require user feedback on usable forms/screens; do not
turn this into full CLI parity or require the separate navigation redesign first.

## Task map

| Slice | Task |
| --- | --- |
| Interaction decisions and feedback checkpoint | [Authoring/workspace action contract](../tasks/6gmshzw7we0m-define-the-tui-authoring-and-workspace-action-contract.md) |
| Selected-parent creation foundation | [Bind task creation to the selected epic](../tasks/6gmy4c0x6e31-bind-task-creation-to-the-selected-epic-identity.md) |
| First staged create | [Tasks and epic context](../tasks/6g4ym5y2dkp8-create-tasks-from-the-tui-and-epic-context.md) |
| Confirmed-entry removal foundation | [Guard forget against registry label reuse](../tasks/6gmy4c15aj30-guard-forgetting-a-confirmed-space-registration-against-label-reuse.md) |
| Atlas registry management | [Register/forget existing entry points](../tasks/6gmshzwfkj62-register-and-forget-existing-space-entry-points-from-atlas.md) |
| Further entity creation | [Epics, audits, research and Threads](../tasks/6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md) |
| First reuse slice | [Epic creation and shared staged mechanics](../tasks/6gp1ytfvm652-create-epics-in-the-tui-with-shared-staged-form-mechanics.md) |
| Creation foundation before more forms | [Host and interaction contracts](../tasks/6gpb4y5yqmjp-harden-the-shared-tui-creation-host-and-interaction-contracts.md) |
| Different input/default contracts | [Audit and research creation](../tasks/6gp1ytg50tev-create-audits-and-research-from-shared-tui-authoring-forms.md) |
| Graph-aware creation | [Guarded Thread creation](../tasks/6gp1ytgea8dt-create-threads-from-the-tui-with-guarded-per-kind-intent.md) |
| Checklist decisions | [Acceptance-criterion actions](../tasks/6g3cp32y9149-let-the-tui-flip-an-acceptance-criterion.md) |
| Existing checklist policy consistency | [Align tracked writes and lint](../tasks/6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md) |
| Routine task fields | [Remaining metadata pickers](../tasks/6gmshzwxxce3-expose-remaining-routine-task-metadata-through-tui-pickers.md) |
| Deliberate task rename | [Selected task and cascade recovery](../tasks/6gmshzx55n1y-rename-selected-tasks-from-the-tui-with-cascade-recovery.md) |
| Thread actions | [Membership and lifecycle](../tasks/6gmshzxchdzc-manage-thread-membership-and-lifecycle-from-the-tui.md) |
| Sequencing actions | [Guarded dependency editing](../tasks/6gmshzxksjcn-edit-task-dependencies-from-the-tui-with-guarded-graph-feedback.md) |
| Audit decisions | [Finding dispositions](../tasks/6gmshzxvgagj-manage-audit-finding-dispositions-from-the-tui.md) |
| Editor-assisted recovery | [Unreadable records](../tasks/6ffdv9g01q40-edit-unreadable-files-from-the-tui.md) |
| Existing preferences clarity | [Configuration scope](../tasks/6g321mybqvfz-stop-describing-the-configuration-editor-as-space-scoped-when-it-opens-globally.md) |
| Existing cross-space refresh hardening | [Bind watcher callbacks to their session](../tasks/6gmbfxy96tpw-bind-tui-watcher-health-and-reload-messages-to-their-originating-session.md) |
| Integrated finish line | [Disposable TUI-only walkthrough](../tasks/6gmshzy2ppfb-prove-routine-tui-planning-without-cli-round-trips.md) |

The graph records the foundations above and only actual delivery prerequisites.
Core capability work is not duplicated in the UI tasks: their relevant criteria
explicitly name the foundation tasks as tracked destinations, while UI closures,
selection and composition integration remain in the UI slices.

The final walkthrough follows the capabilities it exercises and watcher hardening.
Existing diagnostic-editor and configuration-scope tickets remain independent.
Shared tickets can stay in their existing Threads too; membership is not ownership
transfer or a new dependency.

## Neighboring work, not an automatic prerequisite

[Structured action-target consistency](../tasks/6g86g03zfj2f-define-consistent-action-targets-for-structured-tui-detail-selections.md)
belongs to navigation work. Initial writes must name their exact owner/endpoint;
do not silently mutate a parent Thread when a child task is highlighted. Add a
dependency on that target contract only if an implementation really needs its
unresolved structured-cursor behavior; explicit owning list/detail actions do not
require the entire navigation redesign.

## Design checkpoint and remaining questions

The [interaction checkpoint](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md)
records owner-confirmed creation/removal direction, cancellation/submission/session
defaults and the adapter-safe implementation gate. It describes design intent,
not shipped features.

- Settle criteria state-only v1 versus simple add/reword/remove authoring with the
  owner before that task; it does not block entity creation.
- Design bounded selected-child/endpoint capabilities in each owning implementation
  slice, preserving established architecture rather than bypassing it in the TUI.
- Ask about new substantive interaction choices or capability limitations found
  during implementation; no hidden expansion into a whole-TUI redesign.
