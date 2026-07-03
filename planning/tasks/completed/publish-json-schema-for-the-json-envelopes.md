---
status: completed
epic: 20-cli-ux-and-ergonomics
description: Generate Draft 2020-12 JSON Schema from the --json envelope structs, exposed via schema --json-schema so agents can validate output
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, json, dx]
created: "2026-06-19"
updated_at: "2026-06-20"
started_at: "2026-06-20"
completed_at: "2026-06-20"
id: 6fdtbb400n8x
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

- [x] Named envelope types extracted (`render/envelopes.go`, 16 incl. the error
      payload); the `*JSON` render funcs marshal them — output byte-identical
      (existing output tests green).
- [x] `tskflwctl schema --json-schema` emits a Draft 2020-12 schema with `$defs`
      (every envelope + shared types; `required` derived from non-`omitempty`).
- [x] The emitted schema validates real `--json` output (round-trip test over a
      representative spread, using santhosh-tekuri/jsonschema as a test dep).

## Shipped (2026-06-20)

`render/envelopes.go` holds the named envelope contract + `JSONSchema()`
(invopop/jsonschema); `schema --json-schema` prints it; `exit.go` uses the named
`ErrorEnvelope`. The round-trip test caught and fixed a latent bug:
`LintJSON`/`FixJSON` could emit `null` (not `[]`) for nil slices — now normalized,
matching the schema + the "empty not null" convention.

## Out of scope / deferred

- **Field descriptions** — deferred. `AddGoComments` needs the source at runtime
  (absent in a shipped binary), so it would yield empty descriptions for the very
  agents that consume the schema. The structure/types/required carry the value;
  descriptions via `jsonschema:"description=…"` struct tags are the runtime-safe
  follow-up.
- New deps: `invopop/jsonschema` (in the binary), `santhosh-tekuri/jsonschema/v6`
  (test-only validator).
- Generating Go types FROM schema; changing any envelope shape.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Extends the agent contract started by the `schema` command
  ([[schema-command-for-agent-self-discovery]]).
- Pairs with [[auto-generate-cli-reference-docs-with-a-ci-sync-check]].
