---
schema: 1
id: 6g3cp32y9149
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: The TUI renders the acceptance roll-up but cannot change a criterion — every flip drops to the CLI
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, domain, ux]
created: "2026-08-24"
updated_at: "2026-08-24"
---
## Objective

The TUI can now *show* a task's acceptance criteria — the detail header carries a roll-up
with the finding glyph vocabulary — but it cannot change one. `grep -rn "SetCriterionState"
internal/tui/` returns nothing. Every flip, deferral, or reword means quitting to the CLI,
which is the one thing the browser exists to avoid.

That asymmetry is new. Before the criterion vocabulary shipped, a criterion was a checkbox
in body prose and reading it in the rendered markdown was the whole story. Now it carries a
state and a reason, the header advertises them, and the surface that advertises a decision
is the natural place to make one.

## What exists to build on

- `core.Service.SetCriterionState` and `EditCriteria` are the write paths, both atomic and
  frontmatter-preserving; the TUI reads through `core.Service` as `tea.Cmd`s already.
- `domain.TallyCriteria` and `theme.CriterionState` give the counts and the marks.
- `criterionRollup` in `internal/tui/detail.go` renders the header line.
- The entity-action machinery (`internal/tui/entity.go`) already runs lifecycle transitions
  through the service and reports failures as an error flash — the completion gate's refusal
  surfaces that way today.

## Open questions

1. **Where does the interaction live?** A criterion is body content, and the detail pane
   renders markdown rather than a list of selectable rows. Options: a dedicated criteria
   sub-view, an inline overlay from the header roll-up, or a command-palette action taking
   an index. The third is cheapest and least like the rest of the TUI.
2. **How is a reason typed?** `deferred`, `wontfix`, `tracked`, and `n/a` all REQUIRE one, so
   a state change is not a single keypress — it needs a text input, which the config editor
   already has a pattern for.
3. **Does `--add`/`--remove`/`--replace` belong here at all?** Flipping a state is a
   decision made while reading; adding or rewording a criterion is authoring, which `task
   edit` and the CLI arguably serve better. Scoping this to STATE changes only would keep it
   small and is probably the right first cut.

## Acceptance criteria

- [ ] A decision is recorded on question 3 before any code — whether this covers state
      changes only, or criterion authoring as well.
- [ ] A criterion's state can be changed from the task detail view without leaving the TUI,
      including the states that require a reason.
- [ ] The write goes through `core.Service`, as a `tea.Cmd` — no store access and no I/O in
      `Update`/`View`.
- [ ] The roll-up in the detail header updates after the write, without a manual refresh.
- [ ] A refused write (a missing reason, an out-of-range index) surfaces as an error flash
      rather than a silent no-op, the way the `task complete` gate's refusal already does.
- [ ] The help legend covers whatever keys this adds.
- [ ] Suite + lint green; no new glyph or colour vocabulary — `theme.CriterionState` already
      owns the marks.

## Out of scope

- Editing criteria in the atlas or any cross-space view. A criterion belongs to one task in
  one workspace.
- A general body editor. `task edit` opens `$EDITOR` on the whole file and that remains the
  answer for prose.

## Related

Follows the criterion vocabulary (`6g31g9f8x4cv`) and the `task ac` write surface. M4 of
[2026-07-24-ai-agent-cli-ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md)
covers the CLI half of structure-aware writes; this is the browser half, which that finding
does not discuss.
