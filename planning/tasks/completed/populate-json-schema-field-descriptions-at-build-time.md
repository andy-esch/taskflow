---
status: completed
epic: 20-cli-ux-and-ergonomics
description: emit per-field `description`s in `schema --json-schema` by feeding invopop AddGoComments at build time (binary lacks source at runtime)
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, schema, agents]
created: "2026-06-20"
started_at: "2026-06-21"
updated_at: "2026-06-21"
completed_at: "2026-06-21"
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

## Shipped (2026-06-21)

- **Generator** `internal/tools/schemacomments` runs `AddGoComments` over
  `internal/cli/render` + `internal/domain` (the only packages the schema
  references — so an unrelated comment edit elsewhere can't make the asset stale)
  and writes the sorted `CommentMap` to `internal/cli/render/schema_comments.json`.
- **`render.JSONSchema`** `//go:embed`s that file and sets `r.CommentMap` before
  `Reflect`, so the **shipped binary emits descriptions with no source at runtime**.
  Now 27/35 `$defs` carry descriptions (the rest are types whose Go decls have no
  doc comment).
- **Drift guard** `render.TestSchemaComments_NotStale` regenerates the map from the
  repo root (matching the generator's `gopath.Join(base, dir)` keys) and compares
  byte-for-byte — runs under the existing `go test ./...` in CI, no workflow change.
  Proven to fail on a stale comment.
- Schema golden regenerated; the round-trip schema test still passes.

## Acceptance criteria

- [x] `schema --json-schema` emits non-empty `description`s for envelope fields,
      sourced from the Go doc comments on the envelope/shared structs.
- [x] Works from the shipped binary (no source tree at runtime) — descriptions
      come from the embedded `schema_comments.json`, not a runtime `AddGoComments`.
- [x] A CI check fails if the embedded descriptions drift from the source comments
      (`TestSchemaComments_NotStale`, in the normal `go test` suite).
- [x] The round-trip schema test still passes (descriptions don't change
      structure/validation).

## Out of scope

- Changing the envelope shapes or `schema_version` (descriptions are additive).
- Per-field examples (the per-kind `schema <kind>` guidance already covers
  authoring examples).

## Related

- Epic [[20-cli-ux-and-ergonomics]] ·
  [[publish-json-schema-for-the-json-envelopes]] (where this was deferred).

## Correction (2026-06-21, adversarial review)

The 3rd review found the shipped version was **inert on payload fields**: invopop
`AddGoComments` skips UNEXPORTED types, but the schema $defs are the unexported
projections (taskJSON/findingJSON/…), so only exported-envelope TYPE descriptions
survived — the per-field guidance was missing. Fixed: field descriptions now come
from `jsonschema:"description=..."` struct tags (read for any type) on the entity
projections (taskJSON/findingJSON/auditJSON/epicMetaJSON); AddGoComments still
supplies envelope type descriptions. A guard test (TestJSONSchema_HasFieldDescriptions)
and a drift-floor were added. Follow-up (minor): the secondary projections
(MoveResult/CreatedItem/statusCountJSON/lintTaskJSON/ErrorItem) + envelope dry_run/body
fields still lack field descriptions — mechanical to add.
