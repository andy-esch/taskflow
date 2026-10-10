---
schema: 1
id: 6gbn4g1j40pj
status: completed
epic: 20-cli-ux-and-ergonomics
description: Add deterministic optional bounds, continuation, sorting, and filtering to planning-entity list queries without changing unbounded defaults.
effort: M
tier: 2
priority: high
autonomy_level: 3
tags: [cli, query, agents, json]
created: "2026-09-19"
updated_at: "2026-10-10"
started_at: "2026-10-10"
completed_at: "2026-10-10"
---

# Bound and page agent-facing list queries

## Objective

Let agents ask for a deterministic, resumable subset of planning entities instead of ingesting an
unbounded corpus. Define the query and continuation contract in shared application/core terms so a
filesystem adapter can scan locally today without locking future database or web adapters into CLI
or Markdown details.

## Scope

- Inventory task, epic, audit, research, Thread, and audit-finding list surfaces and define which
  common bounds apply truthfully to each.
- Add deterministic optional `--sort`, `--limit`, and continuation behavior plus bounded text and
  updated-since filtering where the underlying entity contract supports them.
- Return enough JSON metadata to resume without guessing: applied ordering, returned count, and an
  opaque or explicitly versioned next cursor. Preserve current unbounded behavior when no bound is
  requested.
- Keep CLI rendering downstream of a reusable query/result contract; do not claim filesystem reads
  avoid a full scan when only their returned projection is bounded.

## Acceptance criteria

- [x] A written contract defines stable ordering, tie-breaking, cursor invalidation, filter
      composition, limits, and behavior when the corpus changes between pages.
- [x] Supported list surfaces expose consistent optional bounds and reject unsupported or ambiguous
      combinations with structured validation errors.
- [x] JSON responses publish deterministic continuation metadata without making human/CSV output
      parse cursors from prose; legacy unbounded calls remain compatible.
- [x] The reusable query boundary is independent of filesystem paths, Cobra flags, and Markdown
      parsing, with adapter and CLI tests for empty, exact-boundary, final-page, mutation-between-
      pages, and invalid-cursor cases.
- [x] Contract revision/classification, schema/docs, token-volume evidence, full tests, lint,
      planning lint, and diff checks are recorded.

## Query contract

- Application boundary: `core.ListQuery`, `ListResult[T]`, and `ListPage`, consumed by
  `Service.QueryTasks/QueryEpics/QueryAudits/QueryResearch/QueryThreadViews/QueryFindingPage`.
  Existing `List*`/`QueryFindings` callers are unchanged. CLI binds flags; wire maps page
  metadata; no query code discovers paths, parses Markdown, or constructs persistence.
- No query options preserves legacy order/filter/JSON shape. Any nonzero option opts into
  deterministic query order. Supported sort fields are `id`, `name`, `created`, `updated`,
  optionally prefixed with `-`. Default: ID ascending, research creation descending.
  Ties always use canonical source ID ascending, then the complete value's SHA-256 digest;
  finding identity is `(audit source ID, finding code)`, not its mutable audit slug.
  Identical duplicate occurrences remain visible; pagination does not invent uniqueness.
- Existing entity filters compose by intersection with common filters. `--search` is a
  case-insensitive literal substring, trimmed/lowercased for the query, limited to 256 UTF-8
  bytes. It searches list metadata, not general bodies:

  | Surface | Search fields | Creation date |
  | --- | --- | --- |
  | Tasks | slug, derived title, description, tags | `created` |
  | Epics | ID/name, description, tags | `created` |
  | Audits | slug, area | `date` |
  | Research | slug, description, tags | `created` |
  | Threads | slug, description, goal, tags | `created` |
  | Findings | audit slug, code, title, component, file, resolution note | unavailable |

- Updated means the document's own `updated_at`, falling back to its creation date, not
  member-task activity. `--updated-since` is an inclusive ISO-date filter; undated/invalid
  values are excluded. Findings reject date sorting/filtering. This is not a change feed:
  hand edits and audit bucket transitions need not advance the stamp.
- Limits are 1–1000; omitted is unbounded. Cursor requires the same positive limit, normalized
  search/sort/date options, entity kind, and existing filters. Output format/column selection
  does not bind the cursor: a client can resume a typed page using projected JSON. Existing
  entity filters retain their supplied representation; repeat them rather than substituting
  an equivalent alias or reordered any-of list.
- Cursor version 1 is opaque base64url containing query/view digests and an offset. It holds
  no plaintext IDs, paths, bodies, authorization, source revision, or persisted source-set
  witness. Unsupported versions/oversized/malformed/foreign tokens refuse with validation
  (exit 11); tokens are not authenticated capabilities. Changing the continuation encoding
  or ordering/digest meaning requires a cursor-version change, not silently accepting old tokens.
- Every matching full loaded value (even fields omitted by `-c`) and the full diagnostic
  set binds continuation. Add/delete/rename/edit of a matching row, a changed rollup, or
  changed diagnostics refuses with conflict (14) and instructs restart. Nonmatching edits
  that leave the matching view and diagnostics unchanged do not invalidate it. Detection
  across independent reads is not snapshot isolation; ABA changes restored to the same
  view need not be detected. A source-set witness is not used as corpus identity.
- JSON gains optional `page` only for opt-in queries: `order`, `limit`, `returned`, `matched`,
  optional normalized `search`/`updated_since`, and `next_cursor` only when another page exists.
  Empty/final/exact-boundary pages have no cursor; empty lists stay `[]`. Typed and projected
  JSON share metadata. Human/name/table/CSV streams remain rows; stderr notes truncation and
  directs callers to JSON for continuation. All read diagnostics remain on every page and
  preserve partial-result exit behavior. An unhealthy empty Thread list retains graph health.
- Machine revision 1.85 is **ADDITIVE**: new optional query flags and `page`, no changed default
  fields, order, or scope. Bounded returned projections do not avoid a local corpus scan.

## Implementation and validation

All six CLI surfaces call the shared query use cases and expose the same optional flags.
Typed and projected JSON share continuation metadata; existing unbounded golden outputs
change only their declared machine revision (and the revision policy/schema descriptions).
Generated command docs and embedded schema comments are current.

Tests cover portable non-filesystem source IDs that disagree with declared IDs, deterministic
ties/duplicate occurrences, empty/exact/final pages, caller-slice preservation, date fallback,
query/limit/kind/filter mismatch, record/diagnostic changes, and cross-format continuation.
Real CLI/adapter tests exercise all six lists and all output modes, retain diagnostics on
continued partial reads, and preserve unhealthy graph context on an empty Thread query.

The 100-task fixture emits 19,836 bytes for full unbounded JSON versus 548 bytes for five
`id,slug` projected rows including continuation: **97.2% fewer bytes**. This is output-volume
evidence, not an LLM-token benchmark or a claim of cheaper filesystem reads.

Compiler-valid self-review mutations disabling changed-view refusal and substituting declared
research IDs for canonical source IDs each failed their intended regression test; restored
code passes. Full `just test` (race), `just lint`, `just build`, generated-doc/schema drift
checks, planning lint, and `git diff --check` pass.

Independent [Codex](../audits/6gje9v0x5g5d-2026-10-10-bounded-planning-queries-and-identity-recovery-implementation-codex.md)
and [Antigravity](../audits/6gje9v15sjjm-2026-10-10-bounded-planning-queries-and-identity-recovery-implementation-antigravity.md)
review briefs are prepared for the combined query/recovery PR. Implementation criteria are met;
reviews remain pending, not settled evidence.

## Out of scope

- Claiming keyset pagination avoids a filesystem scan in the local adapter.
- A general search index, ranking engine, remote API, or database migration.
- Changing body-mutation receipts; compact receipts remain a separate compatibility decision.

## Related

- Audit [AI-agent CLI ergonomics, M1](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md)
- Task [Make Thread list projectable](6gbn4g1eypzs-make-thread-list-projectable-like-other-planning-entity-lists.md)
- ADR [Monotonic JSON machine-contract revisions](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
