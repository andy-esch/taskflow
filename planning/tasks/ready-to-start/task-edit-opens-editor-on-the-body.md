---
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: git-style task edit <slug> opening EDITOR on the whole file (parse-before-accept on save); human face of body-editing, with slug autocomplete
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, ux]
created: "2026-06-12"
updated_at: "2026-06-19"
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

## Acceptance criteria (draft)

- [ ] Planning conflicts above resolved (single owner for body editing);
      task de-drafted.
- [ ] A frontmatter-breaking edit cannot land silently (re-validate on save).
- [ ] No agent-reachable path blocks on an editor.

## Related

- Epic [[17-pm-go-cli]] · [[agent-facing-cli-ergonomics-batch]].