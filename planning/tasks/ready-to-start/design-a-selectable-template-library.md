---
schema: 1
status: ready-to-start
epic: 22-selectable-template-library
description: Generalize the single per-kind body template into a named, selectable library (built-in + repo-local), chosen via --template and agent-discoverable.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [feature, templates, dx]
created: "2026-06-22"
---
# Design a selectable template library

## Objective

Generalize the single per-kind body scaffold (now single-sourced on
`domain.Descriptor.BodyTemplate`, post-M1) into a **named, selectable template
library**: a growable, user-/agent-extensible reference of task/epic/audit (and
future-entity) templates, chosen per project need. M1 put one default template per
kind on the entity descriptor; this turns "one default" into "a set of named
templates" with built-in + repo-local resolution and a selection surface.

## Why

Different projects and different work want different scaffolds — a security audit
vs an architecture audit; a bug vs a feature vs a spike; a lightweight epic vs a
full charter. Today the scaffold is one-size-fits-all. A library lets a repo (or an
agent) pick the right shape, and grows into a shared reference of good templates.

## Sketch (for discussion — not locked)

**Layers / resolution (precedence high → low):**
- **Repo-local** — templates a project defines in its planning tree (e.g.
  `templates/<kind>/<name>.md`), tailoring scaffolds to its needs. Dogfooded:
  markdown + frontmatter like everything else, reusing the split/parse machinery.
- **Built-in** — the curated set embedded in the binary (`go:embed`), the growable
  "reference library" the tool ships. `default` == today's template.
- Repo-local overrides built-in by `<kind>/<name>`.

**Template format:** a markdown file with a small frontmatter header (`name`,
`description`, `kind`, maybe `frontmatter-defaults`/`placeholders`); body is the
scaffold.

**Selection (a flag twin, like the rest of the CLI):**
- `task new "X" --template bug`, `audit new sec --template security`,
  `epic new "Y" --template charter`. Omitted → `default`.
- `--template` is shell-completable (offer resolved names per kind).
- Off-TTY/agent: the flag is the only path (no prompt); a bad name fails exit 11
  listing the available names.

**Discovery (agent-facing):**
- `tskflwctl template list [--kind audit] [--json]` and `template show <kind> <name>`
  — list/inspect built-in + repo-local templates, with their source.
- `schema <kind>` references the available templates so an agent discovers them.
- Authoring face: `template new <kind> <name>` (+ `template edit`) scaffolds a
  repo-local template, mirroring the task/audit authoring surface.

**Architecture fit:** the template store is a new secondary-adapter concern (read
built-ins via embed + repo-local via the fs store) behind a port the cli/create
paths consume; the descriptor resolves a *named* template instead of a single
`BodyTemplate`. Core stays pure.

## Open questions

- Repo-local location: `templates/` at the planning root vs `.tskflwctl/templates/`?
- Do templates carry frontmatter defaults (e.g. a "bug" task pre-tags `bug`), or
  body-only to start?
- Placeholder model: keep per-kind Printf args, or named placeholders (`{{title}}`)
  so custom templates control their own fills?
- How `schema --json` advertises templates; precedence/versioning rules.

## Out of scope (for now)

- Remote/shared template registries (git-backed or URL library) — local built-in +
  repo-local first; remote is a later layer.
- Per-template validation beyond the existing frontmatter lint.

## Suggested first increment

1. Descriptor resolves a *named* template (built-in only to start; `default` ==
   today's template).
2. `--template` on `task/epic/audit new` + completion + off-TTY error.
3. `template list`/`show` read surface (`--json`).
4. Repo-local override via the store (embed built-ins; scan `templates/`).
5. Authoring surface (`template new`/`edit`) + docs.

## Related

- Builds directly on epic 21's M1 entity descriptor (`internal/domain/entity.go`,
  `BodyTemplate`) — the single-default seed this generalizes.
