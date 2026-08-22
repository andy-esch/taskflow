---
schema: 1
id: 6g2pjjsc8jv1
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: Add --body/--body-file to research set so the agent mutation path is the same for research as for every other entity.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, research, consistency]
created: "2026-08-22"
updated_at: "2026-08-22"
---
# Give research set the body flags task and audit set already have

## Objective

`task set` and `audit set` both accept `--body` / `--body-file`; `research set` accepts
neither. All three have `append`. This is the "two faces of mutation" contract
(agent: field-level `set` plus scriptable body writes · human: `edit`) applied to
research with one face missing — an agent can add to a research body but cannot replace
one without dropping to `edit`, which needs `$EDITOR` and a human.

Found through use, not inspection: an agent authoring demo fixture content on
2026-08-22 reached for `research set --body-file`, got an unknown-flag error, and had to
route the whole body through `research new --body-file` at creation time instead. That
works only for a document being created; there is no scriptable way to rewrite an
existing one.

## Acceptance criteria

- [ ] `research set` accepts `--body` and `--body-file` (including `-` for stdin) with
  the same semantics, validation, and atomic single-write behavior as `task set`.
- [ ] The flags are mutually exclusive with each other in the same way `task set`
  enforces, and compose with the frontmatter flags in one write, not two.
- [ ] Body replacement is re-validated before it lands, so an invalid body fails without
  touching the file.
- [ ] `schema research` and the generated CLI docs describe the flags.
- [ ] A test covers set-body, set-body-file, stdin, and the invalid-body rejection —
  mirroring whatever `task set`'s body tests already assert rather than inventing a
  second shape.

## Out of scope

- Any change to `research new`, `append`, or `edit`.
- Auditing the rest of the CLI for other per-entity flag asymmetries. If this turns out
  to be one of several, that survey is its own task.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Sibling implementations to mirror: `task set` and `audit set` in `internal/cli/`.
