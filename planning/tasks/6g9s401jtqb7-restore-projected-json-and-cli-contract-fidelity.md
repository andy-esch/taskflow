---
schema: 1
id: 6g9s401jtqb7
status: completed
epic: 20-cli-ux-and-ergonomics
description: Resolve the contract audit's projected-column, schema-description, fang-gate, and compatibility-record drift without breaking legacy selectors.
effort: 1 day
tier: 2
priority: high
autonomy_level: 4
tags: [cli, json, contract, audit]
created: "2026-09-13"
audit_sources: [planning/audits/6g810b6ptmrp-2026-09-08-contract-and-compatibility.md, planning/audits/6g9samx90yrd-2026-09-13-projected-json-and-cli-contract-fidelity-implementation-claude.md, planning/audits/6g9sw0vjj48f-2026-09-13-projected-json-and-cli-contract-fidelity-implementation-antigravity.md, planning/audits/6g9sxk32b2er-2026-09-13-projected-json-and-cli-contract-remediation-design-antigravity.md]
updated_at: "2026-09-13"
started_at: "2026-09-13"
completed_at: "2026-09-13"
---

# Restore projected JSON and CLI contract fidelity

## Objective

Close the seven open findings from the contract-and-compatibility audit as one bounded machine-
contract repair. Keep the established table/CSV presentation stable, add canonical wire selectors,
make projected JSON use raw values without renaming legacy projected keys, and repair the adjacent
schema descriptions, machine-path gate, documentation, and compatibility record without widening
into a general command-schema redesign.

## Contract decisions

- The canonical selectors for task/research activity and audit open counts are `updated_at` and
  `open_findings`, matching the full JSON DTOs. Legacy `updated` and `open` selectors remain accepted.
- Default and explicitly legacy table/CSV selections retain their existing `updated` / `open`
  headers and display-oriented created-date fallback. Explicit canonical selectors use canonical
  headers and raw values. Projected JSON always uses raw source values; canonical selectors emit
  canonical keys while legacy aliases retain their requested keys for compatibility.
- A missing optional `updated_at` remains an empty projected string because projected values are
  string-valued; it must never be replaced with `created` on the JSON path.
- Wire descriptions point at the published `criterion_states` / `finding_statuses` registries rather
  than retyping those vocabularies in struct tags.

## Acceptance criteria

- [x] `task list --json -c updated_at` and `research list --json -c updated_at` expose raw
      `updated_at`; a never-edited record does not invent the `created` date.
- [x] `audit list --json -c open_findings` uses the same key as the full wire DTO.
- [x] Legacy `updated` / `open` selectors still resolve and retain their projected JSON keys;
      canonical selectors emit canonical keys, while default and legacy table/CSV presentation
      remains byte-compatible.
- [x] Column tests pin canonical selectors, compatibility aliases, raw projected values, and the
      invariant that the affected projection keys agree with their full-wire counterparts.
- [x] Every pflag truthy `--json=<value>` spelling bypasses fang on a TTY, while false and invalid
      values retain the ordinary cobra validation path.
- [x] Criterion/finding schema descriptions reference their published vocabularies and regression
      coverage prevents that guidance from silently drifting.
- [x] `AuditColumns` / `ResearchColumns` godoc is correctly attached and the schema-version changelog
      is monotonically ordered with an explicit append-at-the-bottom rule.
- [x] All seven audit findings are dispositioned through `audit finding`; focused tests, full suite,
      generated-artifact checks, and planning lint pass.

## Out of scope

- Changing typed full-JSON envelope shapes or persisted frontmatter.
- Renaming stable table/CSV headers or removing legacy column selectors.
- Making every column registry derive from wire DTO reflection; the focused projection invariants
  should provide evidence before a larger registry consolidation is considered.
- Expanding `schema` to describe the executable command surface; that remains owned by
  `make-the-command-safety-annotations-load-bearing`.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Audit [2026-09-08 contract and compatibility](../audits/6g810b6ptmrp-2026-09-08-contract-and-compatibility.md)
- Thread [Harden CLI machine contracts](../threads/6g9txm0s3yyh-harden-cli-machine-contracts.md)

## Implementation outcome (2026-09-13)

Schema 1.66 separates legacy display compatibility from canonical field selection: `updated_at` and `open_findings` are advertised wire selectors; legacy `updated`/`open` retain their selected headers and projected keys; JSON and explicitly canonical table/CSV selections use raw values. The fang gate and error writer share a last-value-wins argv scanner for every pflag boolean spelling, and alias-aware completion cannot suggest a canonical duplicate. Criterion/finding descriptions reference executable schema vocabularies, exported column godoc is repaired, and a source-level test keeps the schema changelog strictly ordered.

Validation: full `go test -race ./...`, golangci-lint, module tidiness, independently regenerated CLI docs/schema comments, planning lint, and `git diff --check` pass. All seven source-audit findings are fixed and the audit is closed. Independent implementation reviews drove canonical table/CSV value repair, legacy projected-key preservation, alias-aware completion, flag-order-independent JSON errors, and command-level never-edited task/research coverage. Changelog continuity is now future-major-aware. The broader registry-invariant concern is tracked by `enforce-projected-column-registry-invariants`; DTO/list parity remains intentionally rejected. Backlog mining also confirmed schema 1.32 had already satisfied `task-new-json-expose-the-minted-id-as-its-own-field`; that stale task was evidenced and completed instead of reimplemented.
