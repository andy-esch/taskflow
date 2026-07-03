---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Make schema + agent guidance lead with epic show and task list -o table -c for triage; --json only for full frontmatter.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [docs, agent]
created: "2026-06-22"
updated_at: "2026-06-22"
started_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r0137q
---
# Steer agents to the terse output path in `schema` + guidance

**Source.** Product feedback (2026-06-22) from an AI agent driving `tskflwctl`.
Its sharpest point: "the documented default is the costly one." `CLAUDE.md`
sends agents to `tskflwctl schema` "for agents," and repo guidance says "use
`tskflwctl` (with `--json`)" — so the *advertised* path is the heaviest one
(full-frontmatter `--json`), while the cheap triage path (`epic show <id>`,
`task list -o table -c slug,status,description`) is buried in `--help` examples.

**Unblocked now** — the terse commands it points to already exist; this is a
guidance change, shippable independently of the JSON-projection work.

## Scope

1. **`tskflwctl schema` (the agent-facing contract)**: lead with the terse
   triage recipe — "for triage use `epic show <id>` and
   `task list -o table -c …`; use `--json` only when you need full frontmatter."
2. **`CLAUDE.md` / agent guidance**: same reframe — the cheap path becomes the
   documented default; `--json` is positioned as the full-fidelity option, not
   the reflexive "machine-readable" choice.
3. (Pairs well with — but doesn't require — the JSON `-c`/compact work; once that
   lands, update the recipe to mention `--json -c …`.)

## Acceptance criteria

- [ ] `schema` output names the terse triage commands before `--json`.
- [ ] `CLAUDE.md` (and any agent-facing guidance) leads with the cheap path.
- [ ] No behavior change — guidance/docs only; docs-check still green.

## Related

- Companion to the JSON projection work (this epic).
- Epic [[20-cli-ux-and-ergonomics]].
