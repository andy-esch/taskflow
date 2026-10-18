---
schema: 1
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: The TUI surfaces malformed files only as a footer count (N unreadable); they are not selectable, so you cannot press E to open one in your editor and fix it. Make them reachable for editing.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui]
created: "2026-06-24"
id: 6ffdv9g01q40
audited: "2026-09-13"
updated_at: "2026-10-17"
---
# Edit unreadable files from the TUI

## Objective

Turn unreadable-record counts into an actionable diagnostic surface. A user should
be able to inspect an error and, when an authorized local repair capability is
available, open that exact record in $EDITOR and reload it after repair.

## Scope and implementation constraints

Offer a discoverable problems list or overlay without mixing invalid records into
ordinary typed entity rows. Include task, epic, audit and research load problems,
plus the separate unreadable Thread channel. Dashboard counts should route to the
owning diagnostic surface rather than imply the CLI is the only repair route.

The current portable diagnostic type is `core.LoadProblem`. Its opaque Location,
a filename, or an untrusted declared ID is not identity or permission to open a
path. Use an explicit local repair/path capability bound to the selected source
and active session; extend the application capability narrowly if malformed
records cannot safely use the existing editor handoff. Pathless adapters still
show actionable diagnostics without pretending a local editor is available.

Retain selection and generation checks across delayed replies, workspace changes,
explicit reloads and watcher events. Repaired records enter the normal list;
persisting errors remain inspectable. Duplicate or missing identities must not
open an unrelated file.

## Superseded assumptions

The original path-returning pseudo-row sketch and FileProblem/List* citations
predated portable reads. They are retired, not a recipe to bypass capability
authorization. This is an editor-assisted repair surface, not graph repair or raw
filesystem administration.

## Acceptance criteria

- [ ] Unreadable files (`problems`) are reachable in the TUI — e.g. a "problems"
      view/list or a footer affordance that jumps to them — without disturbing the
      normal browse flow.
- [ ] A selected unreadable record can open in the external editor only through
  an explicit authorized local repair/path capability; reloading promotes
  repaired records or retains actionable errors, without guessing authority from
  Location.
- [ ] The parse error is visible (detail pane or row) so you know what to fix
      before opening the editor.
- [ ] Normal code gates pass; update the README TUI entry point and
  docs/architecture/tui.md for new navigation or asynchronous behavior. Binding
  details stay in owning help sections rather than copied key inventories.
- [ ] Cover task, epic, audit, research and Thread diagnostics,
  duplicate/missing source identity, pathless adapters, delayed results and
  workspace activation without opening an unrelated file.

## Out of scope

A built-in text editor, reconstructing local authority from diagnostic strings,
graph-repair policy changes, or a new machine-output diagnostic format.

## Related

- [Authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [TUI asynchronous work and editor handoff](../../docs/architecture/tui.md).
- [Planning-data capability and identity contracts](../../docs/architecture/planning-data.md).
- [Portable read contracts](../../internal/core/store.go).
