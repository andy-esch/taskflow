---
status: ready-to-start
epic: 17-pm-go-cli
description: 'DRAFT: git-style task edit <slug> opening EDITOR; the human face of the body-editing gap whose agent face (--body-file) is already queued'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, ux, draft]
created: "2026-06-12"
---
# `task edit` — open $EDITOR on the body

> 🚧 **DRAFT — not yet integrated into the overall plan.** Filed from the
> 2026-06-12 CLI-design discussion. Overlaps a queued task — see the
> conflict note; planning should decide where body-editing lives before
> either is started.

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