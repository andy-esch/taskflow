---
schema: 1
id: 6ga3z57hqa70
status: completed
epic: 20-cli-ux-and-ergonomics
description: Expose durable task and audit IDs through token-cheap list projections without broadening into DTO parity.
effort: XS
tier: 2
priority: high
autonomy_level: 4
tags: [cli, json, contract, identity]
created: "2026-09-14"
started_at: "2026-09-14"
depends_on: [6g9txf8x9m70]
updated_at: "2026-09-14"
audit_sources: [planning/audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md]
completed_at: "2026-09-14"
---
## Objective

Expose the stable 12-character identity through the token-cheap task and audit list projection paths. Keep slugs as the human and `-q` handles, but let agents request the durable identifier directly instead of consuming full envelopes or issuing one info call per row.

## Acceptance criteria

- [x] `task list` and `audit list` advertise and accept an `id` column whose value matches the corresponding full JSON envelope.
- [x] `id` is appended after the existing columns so established fields retain their positions; the intentional trailing table/CSV addition, help, completion, and projected JSON behavior are pinned.
- [x] Default human output and `-q` remain unchanged, while explicit canonical selection preserves requested column order for task and audit lists.
- [x] The shared registry/full-wire fidelity harness exercises both new columns and fails if either extractor or envelope mapping drifts.
- [x] The global machine contract advances additively with an explicit changelog entry, regenerated goldens/docs/schema artifacts, and focused plus full validation.

## Scope

- Add only the durable `id` projection identified by the architecture audit; do not pursue general DTO/column parity.
- Preserve slug-first registry order because `renderList` deliberately uses the first column for `-q`.
- Do not add Thread list output modes or decide the broader machine-contract versioning policy here; those remain separate audit findings.

## Related

- Audit finding `2026-09-14-arch-machine-contract:M2`
- Task [Enforce projected-column registry invariants](6g9txf8x9m70-enforce-projected-column-registry-invariants.md)
- Thread [Harden CLI machine contracts](../threads/6g9txm0s3yyh-harden-cli-machine-contracts.md)

## Implementation outcome (2026-09-14)

Task and audit list registries now expose `id` as the final selectable column while retaining `slug` first for `-q` and `-o name`. Explicit table and projected-JSON selection respects caller order, and the default human views are unchanged. The stable table/CSV formats intentionally gain one trailing `id` field without moving their existing columns.

Schema 1.67 records the additive projection and machine-text change. CLI help and completion advertise the new selector; exact task/audit ID projections and both default CSV shapes have golden coverage. The shared registry/full-wire harness independently compares both IDs with the encoded envelopes. A hostile mutation making both extractors return their slugs failed across table, CSV, and projected JSON for each entity before the correct implementation was restored.

Validation passes: live task/audit projection probes, focused CLI/render/wire tests, `go test -race ./...`, golangci-lint, `go mod tidy -diff`, independently regenerated CLI docs and schema comments, planning lint, and `git diff --check`.
