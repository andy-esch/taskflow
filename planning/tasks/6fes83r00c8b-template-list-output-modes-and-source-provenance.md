---
schema: 1
status: ready-to-start
epic: 22-selectable-template-library
description: Add source (built-in vs repo-local) to the template --json envelope and route template list through the shared -o/-c/-q output grammar.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [api-contract, templates, dx]
created: "2026-06-22"
id: 6fes83r00c8b
---
# Template list output modes and source provenance

## Objective

The template `--json` envelopes shipped under schema_version 1.7 but omit the
`source` dimension (built-in vs repo-local, and which wins under override) the design
requires, and `template list` skips the `-o/-c/-q` output grammar every other list
command honors — so it isn't scriptable. Settle the contract shape before step 4
populates it dynamically.

## From the 2026-06-22 adversarial review of the template work (theme 3, ranks 5, 3)

- Add `Source string` (default `"built-in"`) to `render.TemplateInfo` (and the
  `KindSchema.Templates` entries), regenerating the goldens once, so step 4 just sets
  `"repo-local"` without reshaping the agreed envelope. Decide identity:
  `(kind,name)` vs `(kind,name,source)` — this drives `LookupTemplate`'s signature.
- Route `template list` through the shared `listMode`/`renderList` with a
  `TemplateColumns()` registry (kind, name, description, source), so `-o name|table|csv`,
  `-c`, and `-q` work like `task/epic/audit list`. Keep `--kind` as the entity filter.

## Acceptance criteria

- [ ] `template list -q | xargs`, `-o csv`, `-c name,source` work and match the other lists.
- [ ] The `--json` envelope carries `source`; goldens regenerated; `go build`/test/lint/docs-check green.

## Related

- Sequenced before/with step 4 of [[design-a-selectable-template-library]].
