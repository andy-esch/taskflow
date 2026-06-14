---
status: ready-to-start
epic: 17-pm-go-cli
description: 'Reserve a version key in init/task new scaffolds; move audit open-finding regexes into a testable domain.CountFindings (storage-layout review item #4)'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, domain, docs]
created: "2026-06-14"
---

# Scaffold schema-version key and domain-level audit finding counter

## Objective

Two small forward-compat items deferred from
[[put-storage-layout-knowledge-back-behind-the-port]] (its item #4). Both are
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

## Acceptance criteria

- [ ] New scaffolds carry the reserved version key; existing files still parse
      (the key is optional on read, surgical edits preserve it).
- [ ] `domain.CountFindings` exists with table tests (fenced-code exclusion,
      open-vs-total, the `open-ish`/`openness` guards); `auditstore.go` delegates
      to it and the parsing regexes no longer live in the store.
- [ ] `go build/test/vet` green; `gofmt` clean.

## Out of scope

- Migrating or rewriting existing planning files to add the version key — it's
  reserved going forward; back-fill is a separate decision.
- Changing what counts as a finding — this is a move + test, not a rule change.

## Related

- Epic [[17-pm-go-cli]]
- Item #4 from [[put-storage-layout-knowledge-back-behind-the-port]]
- Touches `internal/domain/`, `internal/store/auditstore.go`,
  `internal/store/create.go`, `internal/config/config.go`
