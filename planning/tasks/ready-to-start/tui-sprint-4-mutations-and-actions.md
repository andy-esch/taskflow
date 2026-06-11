---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Lifecycle mutations from the TUI via core.Service with confirmation, reusing the async and reload plumbing
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-10"
---

# TUI sprint 4 mutations and actions

## Objective

Turn the browser into a doer — trigger lifecycle from the TUI, reusing the async
+ reload plumbing. (Design the action model at sprint start.) See
[[18-tui-bubble-tea-interactive-planning-browser]].

## Scope (to refine)

- [ ] **Action model:** decide lifecycle keys (e.g. on a task: a key → a small
      action menu / palette of valid transitions) vs a `:`-command verb surface.
      Confirmation for destructive (deprecate). Reuse the `:` infra from S2.
- [ ] Mutations go through `Service.Move`/`SetFields` as Cmds; on success, the
      S3 reload refreshes the list with the cursor preserved (folder-authoritative
      status means the moved task relocates correctly).
- [ ] Inline feedback line (success/error), reusing semantic colors.
- [ ] Reconsider **multi-select + bulk move** here — only if a concrete bulk
      need is real (the research flagged it as easy to over-build).
- [ ] Tests: an action Cmd calls the right `Service` method; failure surfaces an
      error without corrupting state.

## Acceptance

- [ ] Move/transition a task from the TUI with confirmation; the view reflects it
      live; errors are shown, not swallowed. Suite + lint green.

## Out of scope

- `task new` creation wizard in the TUI (its own follow-on if wanted).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
