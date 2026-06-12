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
+ reload plumbing. See [[18-tui-bubble-tea-interactive-planning-browser]].

## Design (locked 2026-06-12) — dual surface

Mirror the S2a "discoverable affordance + `:` muscle-memory" pattern, for actions:

- **Action menu (discoverable):** a leader key (proposed `a` for *actions*; adjust
  if it collides) opens a small popup over the selected task listing the **valid
  lifecycle transitions** (computed from the current status → allowed targets, not
  a static list). Vim-select (`j`/`k`), `Enter` applies, `Esc` cancels. Reuse the
  `overlay()` compositor from the `?` help modal so it floats over the list.
- **`:` verbs (muscle memory):** extend the S2 command bar with lifecycle verbs —
  `:start` `:complete` `:defer` `:deprecate` `:promote` `:demote` — acting on the
  current selection (route through `dispatchCommand`).
- **Confirmation:** destructive transitions (**deprecate**) require an inline
  `y/n` confirm before applying; non-destructive moves apply immediately.
- **Execution + feedback:** mutations run as `tea.Cmd`s through `core.Service.Move`
  (→ a `movedMsg`/`actionErrMsg`). On success, fire the **S3 `reloadMsg`** path so
  the list refreshes and the moved task relocates (folder-authoritative status),
  cursor preserved by id (`markReload`). Show a transient inline line — `✔ moved
  <slug> → completed` / red error — reusing `theme` colors; never swallow errors.
- **Build on the hardened base:** the concurrent TUI rewrite added `tabMsg`
  routing + `loadGen` guards; the mutation→reload must respect those (the reload
  already does).

## Scope

- [ ] Action menu: valid-transition popup (overlay), vim-select, confirm-on-destructive.
- [ ] `:` lifecycle verbs on the selection, via the command bar.
- [ ] Mutations via `Service.Move` as Cmds → reload-on-success, cursor preserved.
- [ ] Inline success/error feedback line, semantic colors.
- [ ] Tests: an action Cmd calls `Service.Move` with the right status; a failed
      move surfaces an error without corrupting state or the cursor; the `:` verb
      path; the confirm gate blocks an unconfirmed deprecate.

## Deferred (not this sprint)

- **Multi-select + bulk move** — no concrete bulk need yet (research flagged it as
  easy to over-build); revisit when one appears.
- **Field edits** (`SetFields`: priority/tier/tags) from the TUI — a different
  interaction (needs text input, not a transition menu); its own follow-on.

## Acceptance

- [ ] Move/transition a task from the TUI with confirmation; the view reflects it
      live; errors are shown, not swallowed. Suite + lint green.

## Out of scope

- `task new` creation wizard in the TUI (its own follow-on if wanted).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
