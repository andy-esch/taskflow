---
status: completed
epic: 17-pm-go-cli
description: 'Reserve a version key in init/task new scaffolds; move audit open-finding regexes into a testable domain.CountFindings (storage-layout review item #4)'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, domain, docs]
created: "2026-06-14"
updated_at: "2026-06-20"
started_at: "2026-06-20"
completed_at: "2026-06-20"
id: 6fc6vcr02j4x
---

# Scaffold schema-version key and domain-level audit finding counter

## Objective

Two small forward-compat items deferred from
[put-storage-layout-knowledge-back-behind-the-port](6fbj87001q7p-put-storage-layout-knowledge-back-behind-the-port.md) (its item #4). Both are
"cheap while in here" cleanups that were out of scope for the layout work but
worth doing on their own. Independent of each other — split if it reads cleaner.

1. **Scaffold version marker.** Task/epic/audit files carry no schema/version
   key today, so there's no in-file signal of which format wrote them. Reserve
   one in the `init` / `task new` / `epic new` scaffolds (a frontmatter key,
   e.g. `schema: 1`) so future format migrations have something to branch on.
   Decide: bump on every frontmatter shape change, or keep coarse.
2. **Domain-level finding counter.** The audit "what counts as an open finding"
   rule lives as regexes in the store (`auditstore.go` — `findingHeaderRe` /
   `openFindingRe` / the fenced-code stripping). Move it to a testable
   `domain.CountFindings(body) (total, open int)` so the invariant is unit-tested
   in `domain` and the store just calls it.

   > **Note (2026-06-17): item #2 is being subsumed by
   > [audit-finding-level-operations-query-write-lint-sync](6fd5r5c03v5y-audit-finding-level-operations-query-write-lint-sync.md).** That task needs
   > a full per-finding parser (`domain.ParseFindings`); the counts fall out of
   > it, so the count-move should be built there as part of `ParseFindings`
   > rather than a standalone `CountFindings`. Take item #2 there; item #1 (the
   > scaffold version marker) stays here as an independent cleanup — or close
   > this task once #1 lands.

## Acceptance criteria

- [x] New scaffolds carry the reserved version key; existing files still parse
      (the key is optional on read, surgical edits preserve it).
- [x] ~~`domain.CountFindings`~~ — **done as `domain.ParseFindings` +
      `CountOpenFindings`** (2026-06-17) in
      [audit-finding-level-operations-query-write-lint-sync](6fd5r5c03v5y-audit-finding-level-operations-query-write-lint-sync.md) item 1, exactly as
      this task's Note predicted: fence-aware, table-tested, `auditstore.go` now
      derives both counts from it and the regexes are gone from the store.
- [x] `go build/test/vet` green; `gofmt` clean.

## Shipped (2026-06-20) — item #1 (item #2 already landed via ParseFindings)

`domain.FileSchemaVersion = 1` (in `domain/layout.go`, beside the dir-layout
reservations like `ProjectsDir`) is stamped as the **first** frontmatter key,
`schema: 1`, by all three scaffolds (`taskFields`/`epicFields`/`auditFields` in
`store/create.go`). Decisions: **coarse** versioning (bump only on a breaking
shape change); key named **`schema`** (not `schema_version`, to stay distinct from
the `--json` output contract's `render.SchemaVersion`, which versions output);
**reserved, not a known field** — `parseTask` ignores it so existing files still
parse, every surgical edit preserves it (`yaml.Node` keeps unknown keys), it stays
out of `KnownTaskFieldNames` (so `--set` still needs `--force` and the `schema`
contract + goldens are unchanged), and it lints clean. `init` writes no
frontmatter, so it's untouched. Tests: store unit (all three kinds stamp it first;
a `schema:1` file loads and survives `SetFields` + a relocating `Move`) + e2e
create→lint→start→complete keeps it.

## Out of scope

- Migrating or rewriting existing planning files to add the version key — it's
  reserved going forward; back-fill is a separate decision.
- Changing what counts as a finding — this is a move + test, not a rule change.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md)
- Item #4 from [put-storage-layout-knowledge-back-behind-the-port](6fbj87001q7p-put-storage-layout-knowledge-back-behind-the-port.md)
- Touches `internal/domain/`, `internal/store/auditstore.go`,
  `internal/store/create.go`, `internal/config/config.go`
