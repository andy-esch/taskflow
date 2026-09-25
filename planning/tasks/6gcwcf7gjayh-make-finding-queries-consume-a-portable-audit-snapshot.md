---
schema: 1
id: 6gcwcf7gjayh
status: completed
epic: 21-code-quality-architecture-hardening
description: Remove QueryFindings path-based audit rereads and consume one identity-bearing, adapter-neutral audit snapshot.
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, audits, diagnostics, ports]
created: "2026-09-23"
depends_on: [6g5vm4efjcdv]
updated_at: "2026-09-25"
started_at: "2026-09-24"
completed_at: "2026-09-24"
---

# Make finding queries consume a portable audit snapshot

## Objective

Remove the persistence-shaped `ListAudits` → `GetAuditByPath(a.Path)` round trip from
`Service.QueryFindings`. Core should consume one identity-bearing audit snapshot whose findings were
parsed from the same read, while the filesystem adapter remains free to obtain that snapshot however
it chooses.

## Scope

- Reuse or extract the body-aware audit read contract already used by lint and Summary rather than
  defining another path-keyed read.
- Preserve single-audit filtering semantics and the one-scan unfiltered query.
- Return adapter-neutral failed-record evidence with stable identity and optional location.
- Remove `GetAuditByPath` from the core audit port if no remaining application use case needs it.

## Acceptance criteria

- [x] An unfiltered finding query reads each audit source at most once and never calls a path-based
      core port.
- [x] A pathless audit adapter returns the same matching findings, audit IDs, slugs, and buckets as
      the filesystem adapter.
- [x] Unreadable audit records retain identity and optional location without being converted back to
      `domain.FileProblem`.
- [x] Filtering one audit preserves not-found and ambiguous-reference behavior.
- [x] Store and core tests prove no `GetAuditByPath` fallback remains in the query path.

## Out of scope

- Changing finding syntax, filter semantics, ordering, or mutation behavior.
- Redesigning every audit-list projection.
- Implementing pagination.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Adapter-neutral repository lint diagnostics](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
- [Historical read-by-path task](6fes83r01325-reshape-core.store-port-read-by-path-and-fixer-layout-split.md)

## Implementation notes

- Added the selector-aware `AuditSnapshotSource` port. An empty selector performs one resilient body-aware scan; a non-empty selector preserves ordinary exact/prefix/substring audit resolution and reads only the selected source. `QueryFindings` and audit lint consume the same parsed snapshot.
- Removed `GetAuditByPath` from `AuditStore` and the filesystem adapter. Pathless core tests verify finding attribution and identity-aware unreadable evidence without a broad `Store`; filesystem tests pin not-found and ambiguity behavior.
- Carried `LintLoadProblem` through human, projected JSON, and full findings output. Machine schema 1.74 adds portable audit identity/location fields while retaining the local `path` compatibility field.

Validation: `just test` (race-enabled full suite), `just lint`, planning lint, generated schema comments, and the complete machine-contract golden suite pass.

## Adversarial review closeout (2026-09-25)

Both implementation audits are closed with every finding fixed. Explicit narrow audit sources now win independently of option order; missing capability failures are fail-closed and precise; core and filesystem tests pin one snapshot call, zero aggregate-store rereads, selected-file isolation, and one open per source; and opaque non-path diagnostics are exercised through core, full JSON, projected JSON, wire mapping, and semantic schema validation. The hostile reread and location-to-path mutations fail the committed regression tests.
