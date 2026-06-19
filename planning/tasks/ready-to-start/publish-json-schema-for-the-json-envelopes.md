---
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Generate Draft 2020-12 JSON Schema from the --json envelope structs, exposed via schema --json-schema so agents can validate output
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, json, dx]
created: "2026-06-19"
---
## Objective

We version every `--json` payload with one `schema_version`, but ship no machine
contract an agent can validate against. Generate a JSON Schema (Draft 2020-12) from
the envelope structs with [invopop/jsonschema](https://github.com/invopop/jsonschema)
— it reflects Go types and can pull doc-comments in as field descriptions — and
expose it so agents can validate our output and codegen typed clients. This is the
natural completion of the `schema` self-description surface (which already emits the
statuses/fields/exit-codes contract) and squarely on the agent-first line.

**Decision (2026-06-19):** expose via a **`tskflwctl schema --json-schema`
subcommand flag** (emit on demand), with **one schema using `$defs`** covering all
envelopes. A committed file is *not* required (the subcommand is the source of
truth); revisit if a static artifact is later wanted.

**Prerequisite refactor:** today's envelopes are built as *anonymous* inline
structs inside each `*JSON` render func (e.g. `TasksJSON`), which reflection can't
name. First extract them into named exported types in `render` (e.g.
`TasksEnvelope`, `MoveEnvelope`, …) and have the render funcs marshal those — a
worthwhile cleanup in its own right, and the bulk of this task's work.

## Design sketch

- Extract named envelope types; reflect them with invopop/jsonschema (Draft
  2020-12, `AddGoComments` for field descriptions) into one schema with `$defs`.
- `tskflwctl schema --json-schema` prints it; `schema --json-schema <kind>` could
  scope to one envelope later.

## Acceptance criteria

- [ ] Named envelope types extracted; the `*JSON` render funcs marshal them
      (output byte-identical to today — guard with the existing output tests).
- [ ] `tskflwctl schema --json-schema` emits a Draft 2020-12 schema with `$defs`
      and field descriptions from Go comments where present.
- [ ] The emitted schema actually validates real `--json` output for each command
      (a round-trip test).

## Out of scope

- Generating Go types FROM schema (we own the structs; schema is derived, not source).
- Changing any envelope shape — this is a derived artifact, not a redesign.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Extends the agent contract started by the `schema` command
  ([[schema-command-for-agent-self-discovery]]).
- Pairs with [[auto-generate-cli-reference-docs-with-a-ci-sync-check]].
