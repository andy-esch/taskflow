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
depends_on: [6g6x7e2ef37r]
updated_at: "2026-10-03"
audit_sources: [2026-09-28-arch-hexagonal-boundaries]
---
# Test the documented import graph instead of date-stamping a manual review

## Objective

The current `docs/ARCHITECTURE.md` carries an explicit package dependency graph (the
`domain -> id`, `core -> domain, id`, … block), historically under a dated review note.
The composition extraction removes that stale date, but the map is still manual.
The architecture restructuring task will choose its
durable home before this task implements a checker. Wherever it lands, it is a manual
snapshot of something `go list` can answer exactly, which means the doc is accurate only
until the next import lands and nobody re-runs it by hand.

`.golangci.yml` already makes the *stable* part of the direction executable, and
that seam works well. The documented graph is broader than the lint rules — it
includes concrete composition in `internal/appwiring`, permitted utility edges, and
`tui -> configui` — so it cannot simply be deleted in favour of the linter. It
should instead be verified the same way: parse the block, diff it against the
real graph, fail on drift.

That turns the one-screen orientation doc from a claim into an assertion, and
removes the review date as a thing a human has to refresh.

## Acceptance criteria

- [ ] A test parses the dependency-graph block from the canonical focused architecture guide chosen
      by the restructuring task and compares it to the production import graph from
      `go list ./internal/...`.
- [ ] Drift fails with a diff naming the added or removed edge, not just a boolean mismatch
- [ ] Test-only imports are excluded, matching the existing golangci exemption for UI integration tests that construct `store.FS`
- [ ] The manual map points to its executable checker rather than a review date
  or pending task
- [ ] The block's format is documented well enough (or the parser tolerant enough) that an editor cannot break the test with harmless prose edits

## Out of scope

- Changing `.golangci.yml` dependency policy — this task verifies what the doc claims
- The adapter-edge classification table and named local exceptions; those are judgements, not derivable facts
- Any similar treatment of the package-role table above the graph block

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- Follows [Restructure the architecture documentation into focused guides](6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md)

The [composition-isolation task](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md)
now owns audit M1's default-deny policy and L1's recursive coverage. This task independently verifies
the descriptive graph, including the new `cli/ports` and `appwiring` roles. It remains complementary
to depguard and to the [proposed policy ADR](6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md).
