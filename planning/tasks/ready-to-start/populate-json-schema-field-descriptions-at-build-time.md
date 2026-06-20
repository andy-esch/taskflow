---
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: emit per-field `description`s in `schema --json-schema` by feeding invopop AddGoComments at build time (binary lacks source at runtime)
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, schema, agents]
created: "2026-06-20"
---

# Populate JSON Schema field descriptions at build time

## Objective

`schema --json-schema` (shipped in [[publish-json-schema-for-the-json-envelopes]])
emits structure, types, and `required` for every `--json` envelope, but every
field's `description` is empty. invopop can fill descriptions from Go doc
comments via `Reflector.AddGoComments(pkg, path)` — but that reads the **source
tree at runtime**, which a shipped binary doesn't have, so wiring it directly
would yield empty descriptions exactly for the agents who'd consume the schema.

Generate the comment map at **build time** (it's the field docs that make the
schema self-explanatory): e.g. a `go:generate` step / small tool that runs
`AddGoComments` against the source and bakes the resulting map into an embedded
asset (`//go:embed`), which `JSONSchema()` loads. Keep it from silently going
stale — a CI check (like the existing docs-drift gate) should fail if the
embedded comments don't match the source.

## Acceptance criteria

- [ ] `schema --json-schema` emits non-empty `description`s for envelope fields,
      sourced from the Go doc comments on the envelope/shared structs.
- [ ] Works from the shipped binary (no source tree at runtime) — descriptions
      come from an embedded/generated asset, not a runtime `AddGoComments`.
- [ ] A CI check fails if the embedded descriptions drift from the source
      comments (mirrors the docs-gen drift gate).
- [ ] The round-trip schema test still passes (descriptions don't change
      structure/validation).

## Out of scope

- Changing the envelope shapes or `schema_version` (descriptions are additive).
- Per-field examples (the per-kind `schema <kind>` guidance already covers
  authoring examples).

## Related

- Epic [[20-cli-ux-and-ergonomics]] ·
  [[publish-json-schema-for-the-json-envelopes]] (where this was deferred).
