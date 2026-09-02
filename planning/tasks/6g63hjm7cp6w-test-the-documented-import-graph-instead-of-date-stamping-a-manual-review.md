---
schema: 1
id: 6g63hjm7cp6w
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Replace ARCHITECTURE.md's hand-reviewed import graph with a test that diffs it against go list
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, docs, go, dx]
created: "2026-09-02"
---
# Test the documented import graph instead of date-stamping a manual review

## Objective

`docs/ARCHITECTURE.md` carries an explicit package dependency graph (the
`domain -> id`, `core -> domain, id`, … block) under the note "Reviewed against
the production import graph on 2026-08-22". That is a manual, date-stamped review
of something `go list` can answer exactly, which means the doc is accurate only
until the next import lands and nobody re-runs it by hand.

`.golangci.yml` already makes the *stable* part of the direction executable, and
that seam works well. The documented graph is broader than the lint rules — it
includes the edges deliberately left unconstrained (the `cli` composition root,
`tui -> configui`) — so it cannot simply be deleted in favour of the linter. It
should instead be verified the same way: parse the block, diff it against the
real graph, fail on drift.

That turns the one-screen orientation doc from a claim into an assertion, and
removes the review date as a thing a human has to refresh.

## Acceptance criteria

- [ ] A test parses the dependency-graph block out of `docs/ARCHITECTURE.md` and compares it to the production import graph from `go list ./internal/...`
- [ ] Drift fails with a diff naming the added or removed edge, not just a boolean mismatch
- [ ] Test-only imports are excluded, matching the existing golangci exemption for UI integration tests that construct `store.FS`
- [ ] The "Reviewed against ... on <date>" line is replaced by a pointer to the test, so the doc no longer carries a staleness date
- [ ] The block's format is documented well enough (or the parser tolerant enough) that an editor cannot break the test with harmless prose edits

## Out of scope

- Extending `.golangci.yml` depguard rules to the currently-unconstrained edges — this task verifies what the doc claims, it does not change policy
- The adapter-edge classification table (the `cli -> configstore` dispositions); that is a judgement record, not a derivable fact
- Any similar treatment of the package-role table above the graph block

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
