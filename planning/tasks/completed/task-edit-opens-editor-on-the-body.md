---
status: completed
epic: 20-cli-ux-and-ergonomics
description: git-style task edit <slug> opening EDITOR on the whole file (parse-before-accept on save); human face of body-editing, with slug autocomplete
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, ux]
created: "2026-06-12"
updated_at: "2026-06-20"
started_at: "2026-06-20"
completed_at: "2026-06-20"
---
# `task edit` — open $EDITOR on the body

## Decided (2026-06-17)

De-drafted. Resolutions:
- **Human/agent faces stay split** — this owns the human face (`$EDITOR`); the
  agent face (`--body-file`/stdin, body replace/append) stays in
  [[agent-facing-cli-ergonomics-batch]]. Agents won't use `$EDITOR`.
- Edit the **whole file** (git-`commit`-style) with **parse-before-accept on
  save** — a frontmatter break or unknown key is caught and the editor reopened
  / file restored.
- **Slug auto-complete** is required (`task edit <TAB>`, like `task show`).

## Objective

Humans currently hand-edit task files to change a body — exactly what the
tool's validated/atomic write path exists to avoid. Close the gap git-style:

1. `task edit <slug>` opens the task file in `$EDITOR` (`$VISUAL` first,
   fallback vi), then on save re-validates the frontmatter (parse-before-
   accept: an edit that breaks the frontmatter gets the diagnose-style error
   and the file restored or the editor reopened — decide which).
2. Non-TTY / `--json` / missing $EDITOR → clear exit-11 error pointing at
   the flag-driven alternatives. Never blocks an agent.

## ⚠️ Conflicts to resolve before starting

- **[[agent-facing-cli-ergonomics-batch]] already owns the body-editing
  gap** (its remaining items: `--body-file`/stdin on `task new`, body
  replace/append through the tool). This draft is the *human face* of the
  same gap. Planning should either fold this into that task or split
  human/agent faces explicitly — defer to that task's framing; don't build
  both independently.
- Editing the WHOLE file vs body-only is a real design fork: whole-file is
  simple but bypasses the field registry (a hand-edit can write unknown
  keys the `--set` path now rejects); body-only is safer but needs temp-file
  plumbing. Needs a decision.

## Acceptance criteria

- [x] Planning conflicts above resolved (single owner for body editing);
      task de-drafted.
- [x] A frontmatter-breaking edit cannot land silently (re-validate on save).
- [x] No agent-reachable path blocks on an editor.

## Shipped (2026-06-20)

`task edit <slug>` — the human face of mutation, beside the agent-facing
`task set`.

- **`EditTask` lives behind the store port** (`store/edit.go`,
  `core.Store.EditTask`, `Service.EditTask`): it resolves the slug, hands the
  current file content to an `edit(current, prevErr)` callback, and accepts the
  result only if `parseTask` still parses it. The fs and the editor stay
  decoupled — the store orchestrates resolve/parse/atomic-write/loop; the cli's
  callback owns `$EDITOR`. So the whole parse-before-accept engine is unit-tested
  with fake callbacks (no TTY): valid→write, no-change→no-write, invalid→reopen,
  give-up→`ErrValidation` (original survives), editor-error→propagate.
- **Parse-before-accept = the task must still load** (`parseTask`: frontmatter
  parses + field types valid). A broken edit reopens the editor on the *broken*
  content with the error shown; re-saving it unchanged cancels (exit 11) and the
  original file is never touched — the invalid bytes never reach disk.
- **Decision — unknown keys are preserved, not rejected.** The "Decided" note
  said unknown keys should be caught, but rejecting them would (a) contradict the
  surgical-edit principle (preserve unknown fields/order) and (b) make any file
  with a pre-existing unknown key un-editable. `parseTask` (the same loader
  `task show` uses) tolerates them, so `edit` does too: the gate is "does it still
  load", which is what actually protects the file.
- **Editor resolution:** `$VISUAL` → `$EDITOR` → `vi`; the string is split on
  spaces so `EDITOR="code -w"` works. A missing/failing editor → exit 11.
- **Gate:** editing is interactive by nature, so it reuses the prompt `Gate`
  (TTY in+err, not `--json`, not `--no-input`). Closed → exit-11 pointing at
  `task set`. Agents never open an editor and never hang. Slug autocomplete via
  the shared `completeTaskSlugs`.
- **Bare `task edit` (no slug)** mirrors the transition verbs: `MaximumNArgs(1)`
  + `fillSelect`/`editOptions`, so on a TTY it opens the task picker and
  non-interactively it's exit-11 `specify a task to edit` — not a cobra
  `accepts 1 arg(s)` error (fixed post-merge after user feedback).
- Tests: store engine (6 cases) + cli gate (exit 11) + real exec glue
  (`editViaEditor` round-trips through a temp file via a fake editor script).
  Docs regenerated (`docs/cli/tskflwctl_task_edit.md`).

## Related

- Epic [[17-pm-go-cli]] · [[agent-facing-cli-ergonomics-batch]] (the agent face:
  `--body-file`/stdin, body replace/append — still open).