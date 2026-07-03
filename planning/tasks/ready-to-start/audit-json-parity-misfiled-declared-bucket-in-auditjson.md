---
schema: 1
id: 6fjjx9czpd7z
status: ready-to-start
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: AuditJSON lacks misfiled/declared_bucket, so a --json consumer can't detect a misfiled audit (tasks expose them). Add them + bump schema_version — the audit analog of the 1.25 task contract.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [schema, wire]
created: "2026-07-03"
updated_at: "2026-07-03"
---
# Audit JSON parity: misfiled/declared_bucket in AuditJSON

## Objective

Phase A made audit `bucket` frontmatter-authoritative (like task `status`) and added an
audit-misfiled lint, but the JSON contract wasn't extended: `TaskJSON` carries `misfiled`
+ `declared_status` (so an agent detects drift from `task --json`), while `AuditJSON` has
neither. A `--json` consumer of audits can't tell a misfiled audit from a clean one.
Close the parity — the audit analog of the 1.25 task contract.

## Acceptance criteria

- [ ] `AuditJSON` gains `misfiled` (bool, omitempty) + `declared_bucket` (the stale
  mirror directory, omitempty) — the audit analog of `TaskJSON.misfiled`/`declared_status`.
- [ ] `ToAuditJSON` sets them when `a.Misfiled()` (declared_bucket = `a.FolderBucket`).
- [ ] `schema_version` bumped + changelog line.
- [ ] jsonschema descriptions mirror the task ones (bucket authoritative in frontmatter;
  declared_bucket = the stale mirror dir).
- [ ] Extend `TestJSONSchema_ValidatesRealOutput` with a misfiled audit so the populated
  shape validates against the schema (mirrors the misfiled-task case added in 1.25).
- [ ] Regen goldens.

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- [[flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory]] (Phase A) — where audit bucket became authoritative; this is the wire-contract half deferred from step 6.
