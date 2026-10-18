---
schema: 1
id: 6gmrkq00348a
created: "2026-10-18"
description: Proposed task-creation and Atlas register/forget flows, selected-action gaps and owner decisions before implementation.
tags: [tui, authoring, ux, design]
---
# TUI authoring interaction checkpoint

**2026-10-18 · owner-confirmed creation direction; criterion authoring deferred.
Design intent, not shipped behavior.**

Owns the interaction proposal for [the design checkpoint](../tasks/6gmshzw7we0m-define-the-tui-authoring-and-workspace-action-contract.md)
in [the authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
The [clickable sketch](prototypes/6gmshzw7we0m-tui-authoring-checkpoint.html)
uses simulated data and performs no planning or registry writes. It compares a
direct staged form with a field-list form, and demonstrates Atlas register/forget.
This is a layout/click-through sketch, not an ANSI terminal renderer or evidence
that the final keyboard bindings, guarded writes or recovery paths are implemented.

## Owner decisions (2026-10-18)

| Decision | Outcome | Consequence |
| --- | --- | --- |
| Entity creation | Page-local direct staged form, opened with `c` | Tasks creates tasks, epics creates epics, and so on; not a universal creation wizard |
| Atlas registration removal | Change only the system-wide registry registration | No planning files/checkouts are removed; the already-open checkout need not close |
| Criteria v1 | Explicitly deferred by owner | Settle add/reword/remove versus state-only before the criteria task, not as a blocker for creation |
| Architecture | Preserve adapter-neutral interfaces and selected-action safety | UI conveniences cannot bypass application services, guarded persistence or explicit capabilities |

The owner reviewed the plain-text sketches because the embedded visualization was
unavailable on their device. The shared interaction defaults below guide the first
implementation; substantial new product choices still require feedback.

## Device-independent reference sketch

```text
New task                        Save to: taskflow / main
Title        [                                        ]
Epic         [choose existing epic                    v]
Tags         [                                        ]
Description  [                                        ]
> More options (core defaults)
Nothing saved until Create          [Cancel] [Create]
```

Each page opens its own kind-specific form via `c`; this task example is not a
global entity-type chooser. Atlas has a separate path/optional-label registration
form and an exact-entry forget confirmation:

```text
Forget registration: main · ~/git/taskflow
Only this machine-wide registration is removed.
Planning files, checkouts and other entries stay intact.
Current checkout remains open, marked not registered.
                         [Forget registration] [Cancel]
```

## Capability boundary and present reality

Native forms/actions should cover routine planning. $EDITOR owns long bodies and
authorized local-file repair. Existing configuration editing stays in its own
editor, with honest user/repository scope labels. Bulk manifests, migrations,
graph repair, exceptional force overrides, bootstrap/link setup and destructive
filesystem administration remain CLI-first. CLI automation is not being replaced.

Today, tasks/epics have immediate per-field editing and lifecycle menus; audits
have lifecycle actions; Threads are browsable but read-only and research has no
lifecycle. There is no TUI creation form or structured criterion/finding writer.
Atlas already selects entry points within grouped spaces, but exposes no registry
writes. This proposal extends those surfaces, not a claim of existing parity.

## Entry points and discoverability

- Reuse the searchable Ctrl+P palette. Add named contextual actions ahead of entity
  jumps for an empty query; keep fuzzy search across both. Do not replace it with
  another global action menu or overload the lifecycle-only `m` menu.
- `c` opens a new entity of the current page's kind in both list and detail:
  task on tasks, epic on epics, audit on audits, and so on. Preserve `n`/`N` search.
  “New task in this epic” is a separately named contextual action, not an overload
  of epic-page `c`. Atlas `c` is explicitly “Register existing entry point”, not
  initialize a new planning space. Advertise actions in footer/help as they ship.
- Palette rows name action and target: “New task in <epic>”, “Edit criteria for
  <task>”, “Forget registration <entry>”. Missing selection disables selected
  actions with a reason; empty lists still allow creation.
- Keep `e` field edits, `m` lifecycle and `E` external editor. New actions reuse
  the same declared action/key/help routes. No speculative key per capability.

## Creation: one record, one explicit submit

A bounded overlay keeps the originating space/checkout visible. Initial task form:
title, searchable epic, tags and optional description; collapsed More options
shows default priority/effort/tier/autonomy values and allows deliberate override.
Epics are selected by established identity, not a display slug. From epic detail,
preselect that epic and avoid asking again; do not silently replace it if it
disappears. An empty epic catalog explains the prerequisite and offers a route to
epic creation once available, never a nested surprise write.

Use actual core defaults and required inputs. Persisted FieldDoc.Required is not a
create-form schema. Form values are staged in memory: changing a field writes
nothing. Keep the current existing-record editor explicitly labelled “Enter saves
this field”; a create form says “Nothing saved until Create”.

In the selected direct form, Tab/Shift+Tab move focus; Enter in text advances or
accepts a picker, never unexpectedly creates. Enter on the explicit Create button
submits. In the field-list alternative, Enter opens/accepts a field and a separate
Create row submits. No terminal-dependent Ctrl+Enter requirement. Pickers use
arrows/jk only outside text entry; typing q/h/j/k remains normal text inside inputs.

Create in the core default lifecycle state; status changes are separate actions.
Use the default scaffold, then offer “Open in editor” after creation. No automatic
editor launch, template picker or create-and-start in the first slice.

## Cancel, submission, refresh and constrained terminals

| Situation | Proposed behavior |
| --- | --- |
| Esc in a picker/field sub-mode | Return to the form without persisting |
| Esc at a clean form | Close and restore origin selection |
| Esc at a dirty form / normal quit | Ask “Discard unsaved input?”; default is Keep editing |
| Submit pending | Show Saving in the origin space; disable repeat submit and ordinary workspace navigation; do not call it cancellable |
| Validation refusal | Keep all input, focus the relevant field when attributable; otherwise show a form-level core error |
| Definitively failed, uncommitted write | Keep input and explicit retry; no automatic mutation replay |
| Committed/partial or uncertain outcome | Show durable receipt/remedy; replace Create with inspect/recover actions, never retry as a new create |
| Successful write / no-op | Select exact resulting identity, reload coherent views, retain focus through watcher refresh |
| Selected source/child changed | Refuse stale intent; preserve draft for review/reselection rather than applying to a different item |

In v1 there is no cross-space draft store or disk autosave. Ordinary navigation
cannot switch spaces under a dirty form. Do not open a new authoring form while a
workspace-open request is pending. Service activation must invalidate any surviving
submission intent, never rebind it; preserve displaced typed input as a non-submittable
recovery copy rather than silently losing it. Already-issued commands retain their
originating Service; generation/session guards govern delivery, not rollback.
Force-quit/process termination is not a durable-draft guarantee.

At smaller widths the overlay uses available space, stacks labels and scrolls
fields while retaining target, primary action and errors. Extremely small terminals
show a resize notice; never truncate typed input as a mutation or hide where it
will save. Escape semantics and text-input key ownership remain unchanged.

## Atlas: register and forget, not initialize and delete

**Register existing entry point:** collect path and optional machine-local label.
Explain that this expects an existing configured planning entry point, not a
checkout to initialize. Submit through registry policy; show diagnosed canonical
entry/group context, duplicate/no-op or collision/refusal accurately. Registration
refreshes Atlas and selects the returned entry; it does not open/switch the active
workspace. A subsequent Enter is the explicit open action. Path input is resolved
by the application capability, not by View/Update or guessed TUI path policy.

**Forget registration:** choose the exact entry, including within a multi-entry
group. Confirmation names label, checkout, planning identity when known and whether
it is active; says “Planning files and checkouts will not be deleted”. Default is
Cancel. Forget only that registry row, never siblings or a whole group implicitly.
Allow forgetting broken entries without requiring the planning tree to load.

Active-entry behavior is to keep the open session/watcher, mark it
“not registered”, and retain the ordinary return-to-current-space route. Removing
the last registration leaves an empty Atlas with Register available; it need not
destroy the already-open workspace. This is registration removal, not a planning
lifecycle transition or an instruction to rewrite the workspace's repo config.

## Adapter-safe implementation gate

“Shared forms” means reusable TUI mechanics, not a universal CreateEntity API.
Per-kind core use cases retain semantic inputs, validation, defaults and receipts;
modal state, keys and Bubble Tea messages remain in the primary adapter.

- Keep CLI and TUI as peers over core. No Cobra, Bubble Tea, YAML, registry-file
  types or local-path inference in new portable use-case contracts.
- Reuse existing consumer-owned ports; add only bounded capabilities where a
  selected-target or persistence guarantee is genuinely missing. Do not make
  every adapter implement unrelated writes or editor/navigation capabilities.
- Select concrete persistence and mutation policy at binary composition. Preserve
  explicit source-set agreement, origin-Service binding and selected endpoint
  evidence; an ID-shaped reference string is not a substitute.
- Core owns authorization, validation, retries, graph effects and recovery. The
  TUI presents receipts, never implements a second policy or blindly retries
  committed/partial work. Optional local editor/watch capabilities stay explicit.
- Test new core seams with portable/pathless adapters as well as the local guarded
  adapter. Cover read-only refusal and absent capabilities, not just a mocked UI.
  Existing architectural checks and owner comments/guides change with code.

The current full TUI still needs explicit local workspace/Layout capabilities;
this checkpoint does not claim it is already a remote client. The requirement is
that new application semantics remain reusable by another primary/secondary
adapter without depending on this modal or filesystem representation.

## Implementation gaps, not assumed adapter-only work

| Surface | Verified foundation | Required design/implementation attention |
| --- | --- | --- |
| Create task / epic picker | NewTask and staged data can reuse existing widgets | Creation API currently takes an epic reference; ensure displayed epic intent cannot fall back to an impostor |
| Existing task/epic fields and lifecycle | Selected Service entry points already used by TUI | Preserve these routes; do not regress to fresh-reference helpers |
| Criteria / findings | Existing body transforms and disposition policy | Selected task/audit plus displayed child intent; stale index/code or body reorder must refuse rather than retarget, including retries |
| Task rename / membership / dependencies | Existing core guarded operations and receipts | Reference-oriented entry points need bounded selected-endpoint support before UI dispatch; no resolution-as-ID-string shortcut |
| Atlas registry | SpaceRegistryService.Add/Forget; current TUI injects overview/open services | Inject registry capability; conditional forget must bind the confirmed entry under persistence guard, not just recheck a reusable label in UI |

Selected-epic creation is now owned by
[its prerequisite task](../tasks/6gmy4c0x6e31-bind-task-creation-to-the-selected-epic-identity.md);
confirmed-entry removal by
[its prerequisite task](../tasks/6gmy4c15aj30-guard-forgetting-a-confirmed-space-registration-against-label-reuse.md).
Other feature tasks retain their bounded application extensions rather than imply
all core gaps are already solved.

Do not require every body edit to invalidate an unrelated item gratuitously; an
expected-body/child-intent comparison is an implementation choice, but must never
guess which repeated or changed criterion/finding the user meant. No new persistent
child IDs or frontmatter schema is approved by this checkpoint.

## Focused scenarios and delivery

Start with task creation and Atlas registration independently after this checkpoint.
Reuse staged mechanics for additional entity kinds; retain per-kind field
definitions. Keep simple metadata, rename, relationships and dispositions separately
reviewable. Existing configuration-scope and unreadable-record tickets are independent.
Finish with the disposable TUI-only walkthrough, not a claim of full CLI parity.

Implementation regressions must cover: cancel without writes; duplicate submit;
field/form refusal with input retained; same-ID rename versus missing/impostor
selection; edited/reordered child content; pending workspace opens; stale results;
watcher focus; committed cleanup/partial/unknown outcomes; narrow terminals;
grouped, broken and last Atlas entries; forgetting a reused label; no filesystem
deletion. Cover absent/pathless capabilities with explicit unavailability.

## Owning code and executable evidence

- [Fields](../../internal/tui/edit.go), [keys](../../internal/tui/keys.go),
  [palette and dispatch](../../internal/tui/command_dispatch.go) and
  [help](../../internal/tui/help.go) own current interaction mechanics.
- [Atlas](../../internal/tui/atlas.go), [session activation](../../internal/tui/session.go)
  and [registry use cases](../../internal/core/space_registry.go) own current navigation and registration policy.
- [Task create/criteria](../../internal/core/service_task.go),
  [finding writes](../../internal/core/finding.go),
  [Thread writes](../../internal/core/service_thread.go) and
  [dependency operations](../../internal/core/dependency_operations.go) expose the actual application seams.
- [Selected action](../../internal/tui/selected_task_action_test.go),
  [workspace activation](../../internal/tui/selected_task_session_test.go),
  [local editor](../../internal/tui/selected_local_path_test.go) and
  [Atlas](../../internal/tui/atlas_test.go) tests pin protections this work must extend.
- [TUI architecture](../../docs/architecture/tui.md) and
  [planning-data checklist](../../docs/architecture/planning-data.md#change-checklist)
  remain implemented-contract owners. Update them alongside implementation, not as
  if this proposal has shipped.

## Checkpoint evidence and remaining work

- Built the current CLI and started the checkpoint task/Thread through it.
- Browser-checked both layouts, dirty cancellation/Keep editing, explicit create
  feedback, registration feedback and active-entry forget preserving the sibling.
  Checked narrow rendering at 360px; restored browser sizing afterward.
- JavaScript syntax and planning/link lint pass. Production code is unchanged;
  no full Go race/lint run is claimed for this planning-only checkpoint.
- Owner approved page-local creation with `c` and registration-only Atlas removal.
  Criteria authoring remains explicitly deferred to its own task's design checkpoint;
  it does not block task creation or Atlas implementation.
- The interaction checkpoint is complete; implementation and its portable/local
  regression evidence remain pending in the capability tasks.
