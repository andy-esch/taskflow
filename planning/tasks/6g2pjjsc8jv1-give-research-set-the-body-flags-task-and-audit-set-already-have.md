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
updated_at: "2026-09-27"
audited: "2026-09-27"
audit_sources: [2026-09-27-weekly-task-sweep]
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

## Sweep verification (2026-09-27)

Automated weekly sweep. The **core ask is still valid and still unmet** — `research set`
accepts neither `--body` nor `--body-file` today (`internal/cli/research.go`; confirmed
from `research set --help`), so the scriptable body-rewrite gap described in
`## Objective` is real. No acceptance criterion was ticked. Two premises underneath it,
however, do not survive checking, and both touch the acceptance criteria — so they are
recorded here rather than edited.

### `audit set` does not exist, and never has

`## Objective` opens: *"`task set` and `audit set` both accept `--body` / `--body-file`;
`research set` accepts neither."* There is no `audit set` command. `audit`'s subcommands
are `append`, `close`, `defer`, `edit`, `finding`, `findings`, `info`, `lint`, `list`,
`new`, `path`, `reopen`, `show` — and `git log -S'newAuditSetCmd'` returns nothing, so it
was never removed either; the premise was wrong when written. CLAUDE.md's own command
roster agrees (it lists `audit new|list|show|findings|finding|lint|close|reopen|defer`).

What almost certainly got conflated: `audit` *does* have body flags, on
`audit append` (`internal/cli/audit.go:685-687`) and `audit finding new`
(`:347-351`) — just not on a `set` verb it doesn't have.

Two consequences, neither actioned:

- The real asymmetry is **one-to-one, not two-against-one**: `task set` is the *only*
  `set` verb in the CLI with body flags. That weakens the "every other entity has this"
  framing a little, though it does not weaken the two-faces-of-mutation argument, which
  stands on its own.
- `## Related` says *"Sibling implementations to mirror: `task set` and `audit set` in
  `internal/cli/`"*. Only the first half is followable. `internal/cli/task.go:636-673` is
  the implementation to mirror.

### Acceptance criterion 2 contradicts the sibling it asks to mirror

AC 2 requires the flags to *"compose with the frontmatter flags in one write, not two"*.
`task set` does the **opposite**, deliberately and with an error message: `task.go:638`
returns `ErrValidation` with *"--body/--body-file can't be combined with field flags — set
the body in its own call"*, and `--body`'s own help text reads *"(its own call — not
combined with field flags)"*.

So AC 2 and AC 1 (*"the same semantics ... as `task set`"*) cannot both be satisfied.
Whoever picks this up has to choose:

- **mirror `task set`** — reject the combination, and AC 2 is rewritten; or
- **keep AC 2** — `research set` becomes the first `set` verb where body and fields
  compose, which is a deliberate divergence from the sibling and arguably an argument for
  changing `task set` too.

I am **not** editing either criterion: which way this resolves is a contract decision about
the two-faces-of-mutation model, not a drift correction. `## Out of scope` already declines
the broader per-entity flag survey, but this specific contradiction is inside this task's
own criteria and has to be settled before AC 1 and AC 2 can both be signed off.

## Progress Log

- 2026-09-27: automated weekly sweep — core gap re-confirmed unmet; found the `audit set` premise false (that command has never existed; `audit append`/`audit finding` carry the body flags instead) and AC 2's "compose in one write" requirement to contradict AC 1's "mirror `task set`", which explicitly forbids that combination. Both left for a human to resolve.
