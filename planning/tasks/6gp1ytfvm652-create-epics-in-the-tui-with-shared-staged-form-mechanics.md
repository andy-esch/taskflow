---
schema: 1
id: 6gp1ytfvm652
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Add page-local epic creation and extract staging mechanics against the existing task form without sharing entity policy.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-21"
depends_on: [6g4ym5y2dkp8]
updated_at: "2026-10-22"
started_at: "2026-10-21"
completed_at: "2026-10-22"
---
# Create epics in the TUI with shared staged form mechanics

## Objective

Create an epic from its own page using c or the palette, then use it as a task parent without a CLI round trip. This is the first bounded delivery slice of additional entity creation.

## Scope

Reuse the approved staged task interaction. Extract demonstrated common mechanics for focus, enum staging, dirty cancellation, input routing, bounded layout and origin-session state against both task and epic forms. Keep row definitions, NewEpicParams, explicit active/medium initial choices, typed EpicCreationReceipt policy and epic landing separate. Core still owns validation, numbering, scaffolding and persistence; no universal CreateEntity API or persisted-field schema becomes a form definition.

## Acceptance criteria

- [x] Epics-page c and New epic palette action open the same form from list or detail, including an empty list; contextual New task in this epic remains distinct.
- [x] Title, required description, optional tags, status and priority are staged; bounded selectors start at active and medium, and only explicit Create writes through Service.NewEpic in tea.Cmd.
- [x] Task and epic forms reuse actual mechanical staging, cancellation, widget routing and bounded layout without cloning the task controller or sharing parent/default/receipt policy.
- [x] Validation and read-only refusals preserve input; pending, uncertain, committed-cleanup and workspace-displaced outcomes cannot replay or retarget a create and retain recovery evidence.
- [x] Successful creation focuses the exact receipt epic identity across status/filter and watcher refresh; pathless adapters work without promising an editor.
- [x] Real key/paste flows, independent durable expectations, portable capabilities, stale widget results, enum cancellation, dirty cancellation and narrow-terminal modes are covered; build, race suite, lint and owning docs are updated.

## Out of scope

Audit/research/Thread forms, initial Thread membership, template selection and registry management. Additional entity creation remains the integration owner.

## Implementation checkpoint

- Epics-page `c` and the named palette action open the same staged form, including
  an empty list. The existing task-in-epic action remains separate. Initial status
  and priority are explicit active/medium choices, matching the CLI. Description
  is required by core; tags are optional.
- [Shared staging](../../internal/tui/staged_creation.go) and
  [presentation](../../internal/tui/staged_creation_view.go) now have task and epic
  consumers: focus, enum staging, dirty cancellation, input routing, bounded
  layout, origin displacement and exact landing mechanics. Each form retains its
  own widget array with edit-time text-buffer detachment, typed request/receipt
  and recovery policy.
  A small TUI-only draft interface hosts modal rendering/input without repeating
  kind switches in each routing path; it is not a core mutation interface.
- [Epic tests](../../internal/tui/epic_create_test.go) cover durable independent
  defaults/overrides, numbering, unchanged existing bytes/modes, empty lists,
  palette dispatch, read-only/pathless adapters, late origin results, no replay,
  watcher landing and runtime paste. Existing task regressions remain in place.
  The empty-space scenario also creates a task under the newly created epic
  through the contextual palette action without a CLI round trip.
- Quality lookback kept the displayed task default-to-empty mapping outside the
  host, added a kind guard to delayed widget results, preserved original recovery
  diagnostics across displacement, and refused preview receipts as inspection
  targets in both forms. Shared value-ownership/recovery regressions cover both.
- A compiler-valid Go-overlay mutation removed only the widget kind comparison
  in Model.update. `TestEpicCreateOriginAndWidgetBinding` failed specifically
  with “widget delivery crossed entity kinds”; the ordinary focused suite passed.
  This probe never edited production source. Running-program task/paste/epic
  flows also passed three repeated race runs.

Completed for PR handoff on 2026-10-22 after the Codex findings were fixed and all
criteria and validation gates passed. No new core port, persistence schema,
machine-output revision or CLI metadata is introduced. Audit/research and Thread
creation stay in their linked delivery tasks.

### Self-review cleanup and review checkpoint

- Delayed paste now re-checks visible modal ownership at delivery, as direct paste
  does. Both forms reject delivery behind a resize notice, help or configuration;
  restored visible input still works. Cursor/blink updates keep their own chain.
- Recovery pins destination, record identity, non-replay notice and actions before
  bounded diagnostics. A real committed epic with a long Unicode error retains
  those anchors at 48x20 and 46x18; full copyable evidence is unchanged.
- Replaced the overloaded header tuple and long positional render argument list
  with named UI presentation/layout values. Split runtime interaction from the
  narrow-mode matrix and corrected adjacent editor/palette routing comments.
- The old epic no-replay check discarded the returned command. It now executes
  commands and loads the exact real committed epic for inspection. A compiler-valid
  overlay replacing locked Enter with unlock/resubmit fails with `create calls=2`;
  removing the delayed-paste visibility route fails all six task/epic owner cases.
  Ordinary restored regressions and the three real-program flows pass three race
  repetitions. The overlays never changed production source.
- Existing editor/palette/command-bar terminal paste loss is separately scoped in
  [the paste-ownership task](6gp2qmse4prt-route-terminal-paste-to-active-tui-text-input-owners.md),
  added to the authoring Thread without an artificial dependency on this slice.

The independent [Codex review](../audits/6gp2rf2gyp1a-2026-10-21-tui-epic-creation-shared-staging-implementation-codex.md)
arrived on 2026-10-22; its isolated transfer receipt was confirmed before triage.
Both findings were accepted and fixed:

- M1: native Ctrl+V returns a private Bubbles message, not `tea.PasteMsg`. The
  delayed route now requires visible input ownership for every non-blink result,
  without reflecting on private types. Focus-issued blink ticks retain their chain.
  Native-provider regressions execute a stub `pbpaste` (never the user clipboard),
  including actual running-program resize delivery. These integration tests are
  Darwin-only; generic terminal/input ownership regressions run on every platform.
- L1: copying a widget array does not detach internal rune buffers. Shared input
  update now detaches text through the public API before edits while retaining
  cursor, focus, settings and existing errors; blink ticks avoid that copy.
  Non-empty insertion/deletion tests cover both titles and the task parent query.
  Owning comments distinguish retained draft text from animation bookkeeping.

Compiler-valid overlays restoring the old native-message split and omitting buffer
detachment fail the intended new regressions; ordinary code passes three repeated
race runs, including the native runtime flow and blink continuation. Full normal
gates pass again. The reviewer report remains evidence for its captured baseline;
owner dispositions record subsequent fixes, not a second independent verdict.

The owner cancelled this round's Antigravity review; its empty template was removed,
not treated as a completed review. The older task-creation Antigravity audit and
owner AC walkthrough remain untouched. No additional follow-up is needed for these
two bounded fixes; the older-input terminal-paste task remains independently queued.

Validation: `just build`, full `just test` race suite, `just lint`, planning/link
lint, audit lint and `git diff --check` pass. The build emitted the existing
nonfatal Go module stat-cache permission warning; no generated CLI pages changed.

### Interactive check

Run `just run ui`, switch to Epics and press `c` (or Ctrl+P, New epic). Try dirty
cancellation before writing; enter a title/description and optionally tags. Enter
on status/priority opens a bounded selector, Esc cancels the candidate, and only
the explicit Create action submits. Verify focus on the new epic, refresh, then
use Ctrl+P, New task in this epic to exercise the original task form against it.

## Related

- [Additional creation integration](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md).
- [Authoring interaction checkpoint](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md).
- [TUI architecture](../../docs/architecture/tui.md#staged-task-creation).
