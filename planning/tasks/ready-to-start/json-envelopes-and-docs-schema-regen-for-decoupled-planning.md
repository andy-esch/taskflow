---
schema: 1
status: ready-to-start
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Extend InitEnvelope (mode/planning_repo/tracked_repos) + doctor envelope; regen docs/cli via docgen to satisfy the docs-check gate.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, docs, json]
created: "2026-06-22"
updated_at: "2026-06-22"
---
# JSON envelopes + docs/schema regen

Close out the machine-readable surface and the docs-check gate.

## Scope

1. **`render.InitEnvelope`**: pointer mode has no `created` tree — add
   `role`/`mode` ("planning" | "pointer"), `planning_repo`, `tracked_repos`.
   Additive; keep `schema_version` in mind (agents parse this).
2. **`doctor` envelope**: structured findings for `tskflwctl doctor --json`.
3. **`schema` command**: document the new config keys (`planning_repo`,
   `tracked_repos`) if `schema` describes the config contract.
4. **docs-check gate**: regenerate committed CLI docs —
   `go run ./internal/tools/docgen -out docs/cli` — and commit, for the new
   `init` flags (`--planning-repo`, `--track`, `--no-link-back`) and the
   `doctor` command. CI fails if `docs/cli` drifts.

## Acceptance criteria

- [ ] `init --json` reflects pointer vs. scaffold mode + the new fields.
- [ ] `doctor --json` emits structured findings.
- [ ] `docs/cli` regenerated and committed; docs-check passes.
- [ ] Envelope test coverage in `render/envelopes_test.go`.
- [ ] Suite + lint green.

## Depends on

- `init` pointer mode (envelope shape) and `doctor` (its envelope).

## Related

- [[23-point-an-impl-repo-at-an-external-planning-repo]].
