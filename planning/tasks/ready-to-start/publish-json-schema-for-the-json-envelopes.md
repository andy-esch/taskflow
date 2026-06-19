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

## Design sketch

- Reflect the envelope structs (TasksJSON/EpicsJSON/AuditsJSON/show/move/create/
  lint/schema payloads) into a schema, keyed by `schema_version`.
- Expose via `tskflwctl schema --json-schema` (machine) and/or commit
  `schema/v<MAJOR.MINOR>.json` so it's diffable and CI can guard it against the
  live structs (regenerate → fail on drift, mirroring the docs-gen guard).
- Decide one-schema-with-`$defs` vs per-envelope files (invopop `DoNotReference`
  / `Anonymous` knobs control this).

## Acceptance criteria

- [ ] A generator reflects the envelope structs → Draft 2020-12 schema with field
      descriptions sourced from Go comments where present.
- [ ] The schema is reachable as a committed file and/or a `schema` subcommand flag.
- [ ] CI regenerates and fails on drift, so the schema tracks the structs.
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
